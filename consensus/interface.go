package consensus

import (
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	"github.com/tendermint/tendermint/types"
)

// BlockInfo는 블록 정보를 나타냅니다.
type BlockInfo struct {
	Height  int64
	Hash    []byte
	AppHash []byte
}

// Consensus는 컨센서스 관련 인터페이스를 정의합니다.
type Consensus interface {
	// GetBlockStore는 블록 저장소 인스턴스를 반환합니다.
	GetBlockStore() store.BlockStore

	// GetState는 현재 상태를 반환합니다.
	GetState() state.State

	// GetLatestBlock은 최신 블록 정보를 반환합니다.
	GetLatestBlock() (*BlockInfo, error)

	// GetBlockByHeight는 특정 높이의 블록 정보를 반환합니다.
	GetBlockByHeight(height int64) (*BlockInfo, error)

	// GetLatestValidators는 최신 검증자 집합을 반환합니다.
	GetLatestValidators() (*types.ValidatorSet, error)
}
