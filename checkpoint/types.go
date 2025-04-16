package checkpoint

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/types"
)

// CheckpointStatus는 체크포인트의 상태를 나타냅니다.
type CheckpointStatus int

const (
	// StatusCreated는 체크포인트가 생성된 상태입니다.
	StatusCreated CheckpointStatus = iota

	// StatusVerified는 체크포인트가 검증된 상태입니다.
	StatusVerified

	// StatusFinalized는 체크포인트가 확정된 상태입니다.
	StatusFinalized
)

// String은 체크포인트 상태를 문자열로 반환합니다.
func (s CheckpointStatus) String() string {
	switch s {
	case StatusCreated:
		return "CREATED"
	case StatusVerified:
		return "VERIFIED"
	case StatusFinalized:
		return "FINALIZED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// Checkpoint는 특정 블록 높이에서의 체인 상태를 나타냅니다.
type Checkpoint struct {
	// 체크포인트의 블록 높이
	Height int64 `json:"height"`

	// 체크포인트 생성 시간
	CreatedAt time.Time `json:"created_at"`

	// 체크포인트 상태
	Status CheckpointStatus `json:"status"`

	// 체크포인트 블록의 해시
	BlockHash []byte `json:"block_hash"`

	// 상태 루트 해시
	AppHash []byte `json:"app_hash"`

	// 최종 블록 ID
	LastBlockID types.BlockID `json:"last_block_id"`

	// 체크포인트에 대한 검증자 서명 (선택적)
	Signatures []CheckpointSignature `json:"signatures,omitempty"`
}

// CheckpointSignature는 체크포인트에 대한 검증자 서명 정보입니다.
type CheckpointSignature struct {
	// 검증자 주소
	ValidatorAddress []byte `json:"validator_address"`

	// 서명 데이터
	Signature []byte `json:"signature"`

	// 서명 시간
	SignedAt time.Time `json:"signed_at"`
}

// NewCheckpoint는 새 체크포인트를 생성합니다.
func NewCheckpoint(height int64, blockHash, appHash []byte, lastBlockID types.BlockID) *Checkpoint {
	return &Checkpoint{
		Height:      height,
		CreatedAt:   time.Now().UTC(),
		Status:      StatusCreated,
		BlockHash:   blockHash,
		AppHash:     appHash,
		LastBlockID: lastBlockID,
		Signatures:  make([]CheckpointSignature, 0),
	}
}

// AddSignature는 체크포인트에 검증자 서명을 추가합니다.
func (c *Checkpoint) AddSignature(validatorAddress, signature []byte) {
	// 이미 존재하는 서명인지 확인
	for i, sig := range c.Signatures {
		if bytes.Equal(sig.ValidatorAddress, validatorAddress) {
			// 기존 서명 업데이트
			c.Signatures[i].Signature = signature
			c.Signatures[i].SignedAt = time.Now().UTC()
			return
		}
	}

	// 새 서명 추가
	c.Signatures = append(c.Signatures, CheckpointSignature{
		ValidatorAddress: validatorAddress,
		Signature:        signature,
		SignedAt:         time.Now().UTC(),
	})
}

// Verify는 체크포인트의 유효성을 검증합니다.
func (c *Checkpoint) Verify() error {
	// TODO: 체크포인트 검증 로직 구현
	// 1. 서명 검증
	// 2. 블록 해시 검증
	// 3. 기타 필요한 검증

	return nil
}

// Finalize는 체크포인트를 확정 상태로 변경합니다.
func (c *Checkpoint) Finalize() {
	c.Status = StatusFinalized
}
