package milestone

import (
	"errors"
	"sync"

	"github.com/tendermint/tendermint/milestone/types"
)

// ErrManagerMilestoneNotFound는 마일스톤 매니저에서 마일스톤을 찾을 수 없을 때 발생하는 오류입니다.
var ErrManagerMilestoneNotFound = errors.New("milestone manager: milestone not found")

// MilestoneManager는 마일스톤 관리 인터페이스의 구현체입니다.
type MilestoneManager struct {
	mtx          sync.RWMutex
	milestones   map[uint64]*types.Milestone
	noAckIDs     []uint64
	milestoneCnt uint64
}

// NewMilestoneManager는 새로운 마일스톤 관리자를 생성합니다.
func NewMilestoneManager() *MilestoneManager {
	return &MilestoneManager{
		milestones: make(map[uint64]*types.Milestone),
		noAckIDs:   make([]uint64, 0),
	}
}

// GetMilestone는 지정된 ID의 마일스톤을 반환합니다.
func (mm *MilestoneManager) GetMilestone(id uint64) (*types.Milestone, error) {
	mm.mtx.RLock()
	defer mm.mtx.RUnlock()

	milestone, exists := mm.milestones[id]
	if !exists {
		return nil, ErrManagerMilestoneNotFound
	}

	return milestone, nil
}

// GetMilestoneCount는 전체 마일스톤 수를 반환합니다.
func (mm *MilestoneManager) GetMilestoneCount() (uint64, error) {
	mm.mtx.RLock()
	defer mm.mtx.RUnlock()

	return mm.milestoneCnt, nil
}

// GetNoAckMilestones는 승인되지 않은 마일스톤 ID 목록을 반환합니다.
func (mm *MilestoneManager) GetNoAckMilestones() ([]uint64, error) {
	mm.mtx.RLock()
	defer mm.mtx.RUnlock()

	// 복사본 반환
	result := make([]uint64, len(mm.noAckIDs))
	copy(result, mm.noAckIDs)

	return result, nil
}

// AddMilestone은 새 마일스톤을 추가합니다.
func (mm *MilestoneManager) AddMilestone(id uint64, milestone *types.Milestone) error {
	mm.mtx.Lock()
	defer mm.mtx.Unlock()

	if _, exists := mm.milestones[id]; exists {
		return errors.New("milestone with specified ID already exists")
	}

	mm.milestones[id] = milestone
	mm.noAckIDs = append(mm.noAckIDs, id)
	mm.milestoneCnt++

	return nil
}

// AckMilestone은 마일스톤을 승인 처리합니다.
func (mm *MilestoneManager) AckMilestone(id uint64) error {
	mm.mtx.Lock()
	defer mm.mtx.Unlock()

	milestone, exists := mm.milestones[id]
	if !exists {
		return ErrManagerMilestoneNotFound
	}

	// 마일스톤 객체 업데이트
	milestone.Ack = true

	// noAckIDs에서 해당 ID 제거
	for i, noAckID := range mm.noAckIDs {
		if noAckID == id {
			mm.noAckIDs = append(mm.noAckIDs[:i], mm.noAckIDs[i+1:]...)
			break
		}
	}

	return nil
}
