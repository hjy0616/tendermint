package milestone

import (
	"github.com/tendermint/tendermint/milestone/types"
)

// Manager는 마일스톤 관리 인터페이스를 정의합니다.
type Manager interface {
	// GetMilestone는 지정된 ID의 마일스톤을 반환합니다.
	GetMilestone(id uint64) (*types.Milestone, error)

	// GetMilestoneCount는 전체 마일스톤 수를 반환합니다.
	GetMilestoneCount() (uint64, error)

	// GetNoAckMilestones는 승인되지 않은 마일스톤 ID 목록을 반환합니다.
	GetNoAckMilestones() ([]uint64, error)
}
