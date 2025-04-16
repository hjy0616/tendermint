package clerk

import (
	"github.com/tendermint/tendermint/clerk/types"
)

// Manager는 클럭 서비스 관리 인터페이스를 정의합니다.
type Manager interface {
	// GetEventRecords는 지정된 기간 동안의 이벤트 기록을 반환합니다.
	GetEventRecords(fromTime, toTime int64, recordType string) ([]*types.EventRecord, error)

	// CommitEventRecords는 지정된 ID의 이벤트를 커밋하고 실패한 ID 목록을 반환합니다.
	CommitEventRecords(ids []uint64) ([]uint64, error)
}
