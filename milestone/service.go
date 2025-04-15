package milestone

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/milestone/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	// 기본 DB 키 접두사
	milestonePrefix       = "milestone_"
	lastMilestoneCountKey = "last_milestone_count"

	// 기본 설정 값
	defaultMaxQueueSize = 1000
	defaultSyncInterval = 60 * time.Second
	defaultBatchSize    = 100
)

var (
	// ErrMilestoneNotFound 마일스톤을 찾을 수 없을 때의 에러
	ErrMilestoneNotFound = errors.New("milestone not found")
	// ErrInvalidMilestoneID 유효하지 않은 마일스톤 ID
	ErrInvalidMilestoneID = errors.New("invalid milestone ID")
	// ErrQueueFull 큐가 가득 찼을 때의 에러
	ErrQueueFull = errors.New("milestone queue is full")
	// ErrSyncInProgress 동기화가 이미 진행 중일 때의 에러
	ErrSyncInProgress = errors.New("milestone synchronization already in progress")
)

// MilestoneService는 마일스톤 관리 시스템 인터페이스입니다.
type MilestoneService interface {
	// 마일스톤 관리
	AddMilestone(ctx context.Context, milestone *types.Milestone) error
	GetMilestone(ctx context.Context, id uint64) (*types.Milestone, error)
	GetMilestones(ctx context.Context, fromID, toID uint64) ([]types.Milestone, error)
	GetNoAckMilestones(ctx context.Context) ([]types.Milestone, error)
	AckMilestone(ctx context.Context, id uint64) error
	GetMilestoneCount(ctx context.Context) (uint64, error)

	// 마일스톤 라이프사이클 관리
	Start() error
	Stop() error
}

// milestoneServiceImpl은 MilestoneService 인터페이스의 구현체입니다.
type milestoneServiceImpl struct {
	db             dbm.DB
	logger         log.Logger
	milestoneQueue chan types.MilestoneWithTime
	maxQueueSize   int
	batchSize      int

	// 동시성 관리
	mu          sync.RWMutex
	isRunning   bool
	processorCh chan struct{}
}

// MilestoneServiceOption은 마일스톤 서비스 옵션 설정 함수입니다.
type MilestoneServiceOption func(*milestoneServiceImpl)

// WithMaxQueueSize는 마일스톤 큐 최대 크기를 설정합니다.
func WithMaxQueueSize(size int) MilestoneServiceOption {
	return func(m *milestoneServiceImpl) {
		if size > 0 {
			m.maxQueueSize = size
		}
	}
}

// WithBatchSize는 일괄 처리 크기를 설정합니다.
func WithBatchSize(size int) MilestoneServiceOption {
	return func(m *milestoneServiceImpl) {
		if size > 0 {
			m.batchSize = size
		}
	}
}

// WithLogger는 로거를 설정합니다.
func WithLogger(logger log.Logger) MilestoneServiceOption {
	return func(m *milestoneServiceImpl) {
		m.logger = logger
	}
}

// NewMilestoneService는 새로운 MilestoneService 인스턴스를 생성합니다.
func NewMilestoneService(db dbm.DB, options ...MilestoneServiceOption) MilestoneService {
	service := &milestoneServiceImpl{
		db:             db,
		logger:         log.NewNopLogger(),
		milestoneQueue: make(chan types.MilestoneWithTime, defaultMaxQueueSize),
		maxQueueSize:   defaultMaxQueueSize,
		batchSize:      defaultBatchSize,
		processorCh:    make(chan struct{}),
	}

	// 옵션 적용
	for _, option := range options {
		option(service)
	}

	return service
}

// Start는 마일스톤 서비스를 시작합니다.
func (m *milestoneServiceImpl) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return nil
	}

	m.logger.Info("Starting milestone service")
	m.isRunning = true

	// 마일스톤 처리 고루틴 시작
	go m.processMilestones()

	return nil
}

// Stop은 마일스톤 서비스를 중지합니다.
func (m *milestoneServiceImpl) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return nil
	}

	m.logger.Info("Stopping milestone service")

	// 마일스톤 처리 중지
	close(m.processorCh)

	m.isRunning = false
	return nil
}

// AddMilestone은 마일스톤을 추가합니다.
func (m *milestoneServiceImpl) AddMilestone(ctx context.Context, milestone *types.Milestone) error {
	if milestone.ID == 0 {
		return ErrInvalidMilestoneID
	}

	// 현재 큐 길이 확인
	if len(m.milestoneQueue) >= m.maxQueueSize {
		return ErrQueueFull
	}

	// 타임스탬프 추가
	milestoneWithTime := types.MilestoneWithTime{
		Milestone:  *milestone,
		RecordTime: time.Now(),
	}

	// 비동기로 마일스톤 큐에 추가
	select {
	case m.milestoneQueue <- milestoneWithTime:
		m.logger.Debug("Milestone added to queue", "id", milestone.ID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetMilestone은 ID로 마일스톤을 조회합니다.
func (m *milestoneServiceImpl) GetMilestone(ctx context.Context, id uint64) (*types.Milestone, error) {
	if id == 0 {
		return nil, ErrInvalidMilestoneID
	}

	key := []byte(milestonePrefix + string(id))
	data, err := m.db.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrMilestoneNotFound
	}

	var milestone types.Milestone
	if err := json.Unmarshal(data, &milestone); err != nil {
		return nil, err
	}

	return &milestone, nil
}

// GetMilestones는 ID 범위로 마일스톤 목록을 조회합니다.
func (m *milestoneServiceImpl) GetMilestones(ctx context.Context, fromID, toID uint64) ([]types.Milestone, error) {
	if fromID == 0 || toID == 0 || fromID > toID {
		return nil, ErrInvalidMilestoneID
	}

	var milestones []types.Milestone
	for id := fromID; id <= toID; id++ {
		key := []byte(milestonePrefix + string(id))
		data, err := m.db.Get(key)
		if err != nil {
			m.logger.Error("Failed to get milestone", "id", id, "error", err)
			continue
		}
		if data == nil {
			continue
		}

		var milestone types.Milestone
		if err := json.Unmarshal(data, &milestone); err != nil {
			m.logger.Error("Failed to unmarshal milestone", "id", id, "error", err)
			continue
		}

		milestones = append(milestones, milestone)
	}

	// ID 기준 정렬
	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].ID < milestones[j].ID
	})

	return milestones, nil
}

// GetNoAckMilestones는 확인되지 않은 마일스톤 목록을 조회합니다.
func (m *milestoneServiceImpl) GetNoAckMilestones(ctx context.Context) ([]types.Milestone, error) {
	count, err := m.GetMilestoneCount(ctx)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return []types.Milestone{}, nil
	}

	// 모든 마일스톤 조회
	milestones, err := m.GetMilestones(ctx, 1, count)
	if err != nil {
		return nil, err
	}

	// 확인되지 않은 마일스톤만 필터링
	var noAckMilestones []types.Milestone
	for _, milestone := range milestones {
		if !milestone.Ack {
			noAckMilestones = append(noAckMilestones, milestone)
		}
	}

	return noAckMilestones, nil
}

// AckMilestone은 마일스톤을 확인 처리합니다.
func (m *milestoneServiceImpl) AckMilestone(ctx context.Context, id uint64) error {
	if id == 0 {
		return ErrInvalidMilestoneID
	}

	// 마일스톤 조회
	milestone, err := m.GetMilestone(ctx, id)
	if err != nil {
		return err
	}

	// 이미 확인된 마일스톤인 경우
	if milestone.Ack {
		return nil
	}

	// 확인 처리
	milestone.Ack = true

	// 저장
	data, err := json.Marshal(milestone)
	if err != nil {
		return err
	}

	key := []byte(milestonePrefix + string(id))
	return m.db.Set(key, data)
}

// GetMilestoneCount는 마일스톤 수를 조회합니다.
func (m *milestoneServiceImpl) GetMilestoneCount(ctx context.Context) (uint64, error) {
	data, err := m.db.Get([]byte(lastMilestoneCountKey))
	if err != nil {
		return 0, err
	}

	if data == nil {
		return 0, nil
	}

	var count uint64
	if err := json.Unmarshal(data, &count); err != nil {
		return 0, err
	}

	return count, nil
}

// processMilestones는 마일스톤 큐에서 마일스톤을 처리합니다.
func (m *milestoneServiceImpl) processMilestones() {
	batch := make([]types.MilestoneWithTime, 0, m.batchSize)
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-m.processorCh:
			return
		case milestone := <-m.milestoneQueue:
			// 배치에 마일스톤 추가
			batch = append(batch, milestone)

			// 배치 크기에 도달하면 처리
			if len(batch) >= m.batchSize {
				m.processMilestoneBatch(batch)
				batch = batch[:0] // 배치 초기화
				timer.Reset(time.Second)
			}
		case <-timer.C:
			// 타이머 만료 시 배치 처리
			if len(batch) > 0 {
				m.processMilestoneBatch(batch)
				batch = batch[:0] // 배치 초기화
			}
			timer.Reset(time.Second)
		}
	}
}

// processMilestoneBatch는 마일스톤 배치를 처리합니다.
func (m *milestoneServiceImpl) processMilestoneBatch(milestones []types.MilestoneWithTime) {
	m.logger.Debug("Processing milestone batch", "count", len(milestones))

	// ID 기준 정렬
	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].ID < milestones[j].ID
	})

	// 최대 ID 추적
	var maxID uint64

	// 배치 트랜잭션 시작
	batch := m.db.NewBatch()
	defer batch.Close()

	for _, milestone := range milestones {
		// 마일스톤 저장
		data, err := json.Marshal(milestone.Milestone)
		if err != nil {
			m.logger.Error("Failed to marshal milestone", "id", milestone.ID, "error", err)
			continue
		}

		key := []byte(milestonePrefix + string(milestone.ID))
		if err := batch.Set(key, data); err != nil {
			m.logger.Error("Failed to set milestone in batch", "id", milestone.ID, "error", err)
			continue
		}

		// 최대 ID 업데이트
		if milestone.ID > maxID {
			maxID = milestone.ID
		}
	}

	// 마일스톤 수 업데이트
	if maxID > 0 {
		countData, err := json.Marshal(maxID)
		if err != nil {
			m.logger.Error("Failed to marshal milestone count", "error", err)
		} else {
			if err := batch.Set([]byte(lastMilestoneCountKey), countData); err != nil {
				m.logger.Error("Failed to set milestone count in batch", "error", err)
			}
		}
	}

	// 배치 커밋
	if err := batch.Write(); err != nil {
		m.logger.Error("Failed to write milestone batch", "error", err)
		return
	}

	m.logger.Info("Successfully processed milestone batch", "count", len(milestones), "maxID", maxID)
}
