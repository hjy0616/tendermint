package clerk

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/tendermint/tendermint/clerk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const (
	// 기본 DB 키 접두사
	eventRecordPrefix = "event_record_"
	syncStatusKey     = "event_sync_status"

	// 기본 설정 값
	defaultMaxQueueSize = 1000
	defaultSyncInterval = 60 * time.Second
	defaultBatchSize    = 100
)

var (
	// ErrEventNotFound 이벤트를 찾을 수 없을 때의 에러
	ErrEventNotFound = errors.New("event record not found")
	// ErrInvalidEventID 유효하지 않은 이벤트 ID
	ErrInvalidEventID = errors.New("invalid event record ID")
	// ErrQueueFull 큐가 가득 찼을 때의 에러
	ErrQueueFull = errors.New("event queue is full")
	// ErrSyncInProgress 동기화가 이미 진행 중일 때의 에러
	ErrSyncInProgress = errors.New("event synchronization already in progress")
)

// ClerkService는 이벤트 레코드 관리 시스템 인터페이스입니다.
type ClerkService interface {
	// 이벤트 레코드 관리
	AddEventRecord(ctx context.Context, record *types.EventRecord) error
	GetEventRecord(ctx context.Context, id uint64) (*types.EventRecord, error)
	GetEventRecords(ctx context.Context, fromID, toID uint64) ([]types.EventRecord, error)

	// 이벤트 동기화 관리
	SyncEvents(ctx context.Context) error
	GetSyncStatus(ctx context.Context) (*types.EventSyncStatus, error)

	// 시간 기반 이벤트 처리
	ScheduleEventProcessing(ctx context.Context, interval time.Duration) error
	StopEventProcessing() error

	// 라이프사이클 관리
	Start() error
	Stop() error
}

// clerkServiceImpl은 ClerkService 인터페이스의 구현체입니다.
type clerkServiceImpl struct {
	db           dbm.DB
	logger       log.Logger
	eventQueue   chan types.EventRecordWithTime
	maxQueueSize int
	syncInterval time.Duration
	batchSize    int

	// 동시성 관리
	mu          sync.RWMutex
	isRunning   bool
	isSyncing   bool
	processorCh chan struct{}
	syncCh      chan struct{}
}

// ClerkServiceOption은 클러크 서비스 옵션 설정 함수입니다.
type ClerkServiceOption func(*clerkServiceImpl)

// WithMaxQueueSize는 이벤트 큐 최대 크기를 설정합니다.
func WithMaxQueueSize(size int) ClerkServiceOption {
	return func(c *clerkServiceImpl) {
		if size > 0 {
			c.maxQueueSize = size
		}
	}
}

// WithSyncInterval은 동기화 간격을 설정합니다.
func WithSyncInterval(interval time.Duration) ClerkServiceOption {
	return func(c *clerkServiceImpl) {
		if interval > 0 {
			c.syncInterval = interval
		}
	}
}

// WithBatchSize는 일괄 처리 크기를 설정합니다.
func WithBatchSize(size int) ClerkServiceOption {
	return func(c *clerkServiceImpl) {
		if size > 0 {
			c.batchSize = size
		}
	}
}

// WithLogger는 로거를 설정합니다.
func WithLogger(logger log.Logger) ClerkServiceOption {
	return func(c *clerkServiceImpl) {
		c.logger = logger
	}
}

// NewClerkService는 새로운 ClerkService 인스턴스를 생성합니다.
func NewClerkService(db dbm.DB, options ...ClerkServiceOption) ClerkService {
	service := &clerkServiceImpl{
		db:           db,
		logger:       log.NewNopLogger(),
		eventQueue:   make(chan types.EventRecordWithTime, defaultMaxQueueSize),
		maxQueueSize: defaultMaxQueueSize,
		syncInterval: defaultSyncInterval,
		batchSize:    defaultBatchSize,
		processorCh:  make(chan struct{}),
		syncCh:       make(chan struct{}),
	}

	// 옵션 적용
	for _, option := range options {
		option(service)
	}

	return service
}

// Start는 클러크 서비스를 시작합니다.
func (c *clerkServiceImpl) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return nil
	}

	c.logger.Info("Starting clerk service")
	c.isRunning = true

	// 이벤트 처리 고루틴 시작
	go c.processEvents()

	return nil
}

// Stop은 클러크 서비스를 중지합니다.
func (c *clerkServiceImpl) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return nil
	}

	c.logger.Info("Stopping clerk service")

	// 이벤트 처리 중지
	close(c.processorCh)

	// 동기화 중지
	if c.isSyncing {
		close(c.syncCh)
	}

	c.isRunning = false
	return nil
}

// AddEventRecord는 이벤트 레코드를 추가합니다.
func (c *clerkServiceImpl) AddEventRecord(ctx context.Context, record *types.EventRecord) error {
	if record.ID == 0 {
		return ErrInvalidEventID
	}

	// 현재 큐 길이 확인
	if len(c.eventQueue) >= c.maxQueueSize {
		return ErrQueueFull
	}

	// 타임스탬프 추가
	eventWithTime := types.EventRecordWithTime{
		EventRecord: *record,
		RecordTime:  time.Now(),
	}

	// 비동기로 이벤트 큐에 추가
	select {
	case c.eventQueue <- eventWithTime:
		c.logger.Debug("Event record added to queue", "id", record.ID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetEventRecord는 지정된 ID의 이벤트 레코드를 조회합니다.
func (c *clerkServiceImpl) GetEventRecord(ctx context.Context, id uint64) (*types.EventRecord, error) {
	if id == 0 {
		return nil, ErrInvalidEventID
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := []byte(eventRecordPrefix + string(id))
	data, err := c.db.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrEventNotFound
	}

	var record types.EventRecord
	err = json.Unmarshal(data, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// GetEventRecords는 지정된 범위의 이벤트 레코드를 조회합니다.
func (c *clerkServiceImpl) GetEventRecords(ctx context.Context, fromID, toID uint64) ([]types.EventRecord, error) {
	if fromID == 0 || toID == 0 || fromID > toID {
		return nil, ErrInvalidEventID
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var records []types.EventRecord

	// 지정된 범위의 이벤트 레코드 조회
	for id := fromID; id <= toID; id++ {
		key := []byte(eventRecordPrefix + string(id))
		data, err := c.db.Get(key)
		if err != nil {
			c.logger.Error("Failed to get event record", "id", id, "error", err)
			continue
		}
		if data == nil {
			continue
		}

		var record types.EventRecord
		err = json.Unmarshal(data, &record)
		if err != nil {
			c.logger.Error("Failed to unmarshal event record", "id", id, "error", err)
			continue
		}

		records = append(records, record)
	}

	// ID 기준 정렬
	sort.Slice(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})

	return records, nil
}

// SyncEvents는 이벤트 동기화를 수행합니다.
func (c *clerkServiceImpl) SyncEvents(ctx context.Context) error {
	c.mu.Lock()

	// 이미 동기화 중인지 확인
	if c.isSyncing {
		c.mu.Unlock()
		return ErrSyncInProgress
	}

	c.isSyncing = true
	c.syncCh = make(chan struct{})
	c.mu.Unlock()

	c.logger.Info("Starting event synchronization")

	// 동기화 완료 후 상태 업데이트를 위한 defer 함수
	defer func() {
		c.mu.Lock()
		c.isSyncing = false
		c.mu.Unlock()
		c.logger.Info("Event synchronization completed")
	}()

	// 현재 동기화 상태 조회
	status, err := c.GetSyncStatus(ctx)
	if err != nil {
		status = &types.EventSyncStatus{
			LastSyncedID: 0,
			LastSyncTime: time.Time{},
			IsSyncing:    true,
		}
	}

	// 동기화 상태 업데이트
	status.IsSyncing = true
	if err := c.updateSyncStatus(status); err != nil {
		return err
	}

	// 큐에 있는 모든 이벤트 처리
	processedCount := 0
	batch := c.db.NewBatch()
	defer batch.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.syncCh:
			return nil
		case eventWithTime, ok := <-c.eventQueue:
			if !ok {
				// 큐가 닫힘
				break
			}

			key := []byte(eventRecordPrefix + string(eventWithTime.ID))
			data, err := json.Marshal(eventWithTime.EventRecord)
			if err != nil {
				c.logger.Error("Failed to marshal event record", "id", eventWithTime.ID, "error", err)
				continue
			}

			batch.Set(key, data)
			processedCount++

			// 마지막으로 처리된 이벤트 ID 업데이트
			if eventWithTime.ID > status.LastSyncedID {
				status.LastSyncedID = eventWithTime.ID
			}

			// 배치 크기에 도달하면 커밋
			if processedCount >= c.batchSize {
				if err := batch.Write(); err != nil {
					c.logger.Error("Failed to write event batch", "error", err)
					return err
				}
				batch.Close()
				batch = c.db.NewBatch()
				processedCount = 0
			}
		default:
			// 큐가 비었거나 더 이상 이벤트가 없음
			if processedCount > 0 {
				if err := batch.Write(); err != nil {
					c.logger.Error("Failed to write final event batch", "error", err)
					return err
				}
			}

			// 동기화 상태 업데이트
			status.IsSyncing = false
			status.LastSyncTime = time.Now()
			if err := c.updateSyncStatus(status); err != nil {
				return err
			}

			return nil
		}
	}
}

// GetSyncStatus는 현재 동기화 상태를 조회합니다.
func (c *clerkServiceImpl) GetSyncStatus(ctx context.Context) (*types.EventSyncStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := c.db.Get([]byte(syncStatusKey))
	if err != nil {
		return nil, err
	}

	if data == nil {
		return &types.EventSyncStatus{
			LastSyncedID: 0,
			LastSyncTime: time.Time{},
			IsSyncing:    false,
		}, nil
	}

	var status types.EventSyncStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	// 현재 동기화 상태 반영
	status.IsSyncing = c.isSyncing

	return &status, nil
}

// updateSyncStatus는 동기화 상태를 업데이트합니다.
func (c *clerkServiceImpl) updateSyncStatus(status *types.EventSyncStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	c.db.Set([]byte(syncStatusKey), data)
	return nil
}

// ScheduleEventProcessing은 정기적인 이벤트 처리를 예약합니다.
func (c *clerkServiceImpl) ScheduleEventProcessing(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = c.syncInterval
	}

	c.mu.Lock()
	if !c.isRunning {
		c.mu.Unlock()
		return errors.New("clerk service is not running")
	}
	c.mu.Unlock()

	c.logger.Info("Scheduling event processing", "interval", interval)

	// 주기적 동기화 고루틴 시작
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.SyncEvents(ctx); err != nil && err != ErrSyncInProgress {
					c.logger.Error("Scheduled event sync failed", "error", err)
				}
			case <-c.processorCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// StopEventProcessing은 예약된 이벤트 처리를 중지합니다.
func (c *clerkServiceImpl) StopEventProcessing() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return nil
	}

	c.logger.Info("Stopping scheduled event processing")

	// 프로세서 채널이 닫히지 않았다면 닫기
	select {
	case <-c.processorCh:
		// 이미 닫힘
	default:
		close(c.processorCh)
	}

	return nil
}

// processEvents는 이벤트 큐에서 이벤트를 처리하는 고루틴을 실행합니다.
func (c *clerkServiceImpl) processEvents() {
	c.logger.Info("Event processor started")

	buffer := make([]types.EventRecordWithTime, 0, c.batchSize)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.processorCh:
			// 서비스 중지 시 남은 이벤트 처리
			if len(buffer) > 0 {
				c.processEventBatch(buffer)
			}
			c.logger.Info("Event processor stopped")
			return
		case event, ok := <-c.eventQueue:
			if !ok {
				// 채널이 닫힘
				return
			}

			buffer = append(buffer, event)

			// 버퍼가 가득 차면 즉시 처리
			if len(buffer) >= c.batchSize {
				c.processEventBatch(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			// 주기적으로 버퍼 처리
			if len(buffer) > 0 {
				c.processEventBatch(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

// processEventBatch는 이벤트 배치를 처리합니다.
func (c *clerkServiceImpl) processEventBatch(events []types.EventRecordWithTime) {
	if len(events) == 0 {
		return
	}

	c.logger.Debug("Processing event batch", "count", len(events))

	batch := c.db.NewBatch()
	defer batch.Close()

	var lastID uint64

	for _, event := range events {
		key := []byte(eventRecordPrefix + string(event.ID))
		data, err := json.Marshal(event.EventRecord)
		if err != nil {
			c.logger.Error("Failed to marshal event", "id", event.ID, "error", err)
			continue
		}

		batch.Set(key, data)

		if event.ID > lastID {
			lastID = event.ID
		}
	}

	// 배치 커밋
	if err := batch.Write(); err != nil {
		c.logger.Error("Failed to write event batch", "error", err)
		return
	}

	// 동기화 상태 업데이트
	status, err := c.GetSyncStatus(context.Background())
	if err != nil {
		c.logger.Error("Failed to get sync status", "error", err)
		return
	}

	if lastID > status.LastSyncedID {
		status.LastSyncedID = lastID
		status.LastSyncTime = time.Now()

		if err := c.updateSyncStatus(status); err != nil {
			c.logger.Error("Failed to update sync status", "error", err)
		}
	}

	c.logger.Debug("Event batch processed successfully", "count", len(events), "lastID", lastID)
}
