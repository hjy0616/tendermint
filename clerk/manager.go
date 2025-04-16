package clerk

import (
	"errors"
	"sync"
	"time"

	"github.com/tendermint/tendermint/clerk/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	// ErrNoRecordsFound는 레코드를 찾을 수 없을 때의 오류입니다.
	ErrNoRecordsFound = errors.New("no event records found")
	// ErrInvalidTimeRange는 시간 범위가 잘못되었을 때의 오류입니다.
	ErrInvalidTimeRange = errors.New("invalid time range")
)

// ClerkManager는 클럭 서비스 관리 인터페이스의 구현체입니다.
type ClerkManager struct {
	mtx           sync.RWMutex
	records       map[uint64]*types.EventRecord
	recordsByTime map[int64][]uint64 // time -> record IDs
	logger        log.Logger
}

// NewClerkManager는 새 클럭 매니저를 생성합니다.
func NewClerkManager(logger log.Logger) *ClerkManager {
	return &ClerkManager{
		records:       make(map[uint64]*types.EventRecord),
		recordsByTime: make(map[int64][]uint64),
		logger:        logger,
	}
}

// GetEventRecords는 지정된 기간 동안의 이벤트 기록을 반환합니다.
func (cm *ClerkManager) GetEventRecords(fromTime, toTime int64, recordType string) ([]*types.EventRecord, error) {
	cm.mtx.RLock()
	defer cm.mtx.RUnlock()

	// 시간 범위 검증
	if fromTime > toTime {
		return nil, ErrInvalidTimeRange
	}

	// 지정된 시간 범위와 유형에 맞는 레코드 수집
	var result []*types.EventRecord
	for t := fromTime; t <= toTime; t++ {
		if recordIDs, exists := cm.recordsByTime[t]; exists {
			for _, id := range recordIDs {
				record := cm.records[id]
				// 레코드 타입이 지정되었으면 필터링 (ChainID를 recordType으로 사용)
				if recordType == "" || record.ChainID == recordType {
					result = append(result, record)
				}
			}
		}
	}

	if len(result) == 0 {
		return nil, ErrNoRecordsFound
	}

	return result, nil
}

// CommitEventRecords는 지정된 ID의 이벤트를 커밋하고 실패한 ID 목록을 반환합니다.
func (cm *ClerkManager) CommitEventRecords(ids []uint64) ([]uint64, error) {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	var failedIDs []uint64
	for _, id := range ids {
		// 여기서는 단순히 레코드가 존재하는지만 확인합니다.
		// 실제 구현에서는 커밋 로직이 필요합니다.
		if _, exists := cm.records[id]; !exists {
			failedIDs = append(failedIDs, id)
			cm.logger.Debug("Failed to commit event record", "id", id, "reason", "not found")
		}
	}

	return failedIDs, nil
}

// AddEventRecord는 새 이벤트 레코드를 추가합니다.
func (cm *ClerkManager) AddEventRecord(record *types.EventRecord) error {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	// ID가 이미 존재하는지 확인
	if _, exists := cm.records[record.ID]; exists {
		return errors.New("event record with the same ID already exists")
	}

	// 현재 시간 설정 (필요한 경우)
	if record.Timestamp == 0 {
		record.Timestamp = time.Now().Unix()
	}

	// 레코드 저장
	cm.records[record.ID] = record

	// 시간별 인덱스에 추가
	if _, exists := cm.recordsByTime[record.Timestamp]; !exists {
		cm.recordsByTime[record.Timestamp] = make([]uint64, 0)
	}
	cm.recordsByTime[record.Timestamp] = append(cm.recordsByTime[record.Timestamp], record.ID)

	cm.logger.Debug("Added new event record", "id", record.ID, "chain", record.ChainID, "time", record.Timestamp)

	return nil
}
