package consensus

import (
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

// Manager는 합의 관리를 위한 인터페이스를 정의합니다.
type Manager interface {
	// GetState는 현재 상태를 반환합니다.
	GetState() state.State

	// GetBlockStore는 블록 저장소를 반환합니다.
	GetBlockStore() *state.BlockStore

	// GetValidators는 현재 검증자 목록을 반환합니다.
	GetValidators() (int64, []*types.Validator)

	// GetLastHeight는 마지막 블록 높이를 반환합니다.
	GetLastHeight() int64
}

// ConsensusManager는 합의 관리 인터페이스의 구현체입니다.
type ConsensusManager struct {
	logger     log.Logger
	state      state.State
	blockStore *state.BlockStore
}

// NewConsensusManager는 새로운 합의 관리자를 생성합니다.
func NewConsensusManager(logger log.Logger, state state.State, blockStore *state.BlockStore) *ConsensusManager {
	return &ConsensusManager{
		logger:     logger,
		state:      state,
		blockStore: blockStore,
	}
}

// GetState는 현재 상태를 반환합니다.
func (cm *ConsensusManager) GetState() state.State {
	return cm.state
}

// GetBlockStore는 블록 저장소를 반환합니다.
func (cm *ConsensusManager) GetBlockStore() *state.BlockStore {
	return cm.blockStore
}

// GetValidators는 현재 검증자 목록을 반환합니다.
func (cm *ConsensusManager) GetValidators() (int64, []*types.Validator) {
	state := cm.GetState()
	return state.LastBlockHeight, state.Validators.Validators
}

// GetLastHeight는 마지막 블록 높이를 반환합니다.
func (cm *ConsensusManager) GetLastHeight() int64 {
	state := cm.GetState()
	return state.LastBlockHeight
}
