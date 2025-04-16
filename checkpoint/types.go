// Package checkpoint 블록체인의 특정 높이에서 상태를 저장하고 검증하는 기능을 제공합니다.
package checkpoint

import (
	"context"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
)

// CheckpointData는 체크포인트의 기본 데이터 구조를 정의합니다.
type CheckpointData struct {
	Height        uint64         `json:"height"`          // 체크포인트 블록 높이
	BlockHash     []byte         `json:"block_hash"`      // 블록 해시
	StateRootHash []byte         `json:"state_root_hash"` // 상태 루트 해시
	AppHash       []byte         `json:"app_hash"`        // 애플리케이션 해시
	ValidatorHash []byte         `json:"validator_hash"`  // 검증자 세트 해시
	Timestamp     int64          `json:"timestamp"`       // 체크포인트 생성 시간 (Unix 타임스탬프)
	Signature     []byte         `json:"signature"`       // 체크포인트 서명
	SignerAddress crypto.Address `json:"signer_address"`  // 서명자 주소
}

// Checkpoint는 데이터를 포함한 체크포인트 전체 정보입니다.
type Checkpoint struct {
	Data        CheckpointData `json:"data"`         // 체크포인트 데이터
	Number      uint64         `json:"number"`       // 체크포인트 번호 (순차적으로 증가)
	BlockHeight uint64         `json:"block_height"` // 체크포인트가 생성된 블록 높이
}

// Manager는 체크포인트 관리 인터페이스를 정의합니다.
type Manager interface {
	// CreateCheckpoint는 지정된 블록 높이에서 새 체크포인트를 생성합니다.
	CreateCheckpoint(ctx context.Context, height uint64) (*Checkpoint, error)

	// GetCheckpoint는 지정된 체크포인트 번호로 체크포인트를 조회합니다.
	GetCheckpoint(number uint64) (*Checkpoint, error)

	// GetLatestCheckpoint는 가장 최근 체크포인트를 조회합니다.
	GetLatestCheckpoint() (*Checkpoint, error)

	// GetCheckpointCount는 전체 체크포인트 수를 반환합니다.
	GetCheckpointCount() (uint64, error)

	// VerifyCheckpoint는 체크포인트의 유효성을 검증합니다.
	VerifyCheckpoint(checkpoint *Checkpoint) error

	// SetLogger는 로거를 설정합니다.
	SetLogger(logger log.Logger)
}

// Config는 체크포인트 관리자 구성 옵션입니다.
type Config struct {
	// 체크포인트 생성 간격 (블록 수)
	CheckpointInterval uint64

	// 체크포인트 서명에 사용할 개인키
	PrivateKey crypto.PrivKey

	// 체크포인트 검증에 사용할 공개키
	PublicKey crypto.PubKey
}
