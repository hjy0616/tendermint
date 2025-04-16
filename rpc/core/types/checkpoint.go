package coretypes

import (
	"github.com/tendermint/tendermint/libs/bytes"
)

// ResultCheckpoint는 체크포인트 조회 결과를 담는 구조체입니다.
type ResultCheckpoint struct {
	Number       uint64         `json:"number"`
	Height       uint64         `json:"height"`
	BlockHash    bytes.HexBytes `json:"block_hash"`
	StateRoot    bytes.HexBytes `json:"state_root"`
	AppHash      bytes.HexBytes `json:"app_hash"`
	ValidatorSet bytes.HexBytes `json:"validator_set"`
	Timestamp    int64          `json:"timestamp"`
	Signature    bytes.HexBytes `json:"signature,omitempty"`
	Signer       bytes.HexBytes `json:"signer,omitempty"`
}

// ResultCheckpointCount는 체크포인트 개수 조회 결과를 담는 구조체입니다.
type ResultCheckpointCount struct {
	Count uint64 `json:"count"`
}
