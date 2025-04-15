package types

import (
	"time"

	"github.com/tendermint/tendermint/libs/bytes"
)

// Checkpoint는 특정 높이에서의 상태 스냅샷을 나타냅니다.
// 블록체인의 특정 지점에서 상태 동기화를 위한 참조 포인트로 사용됩니다.
type Checkpoint struct {
	Height    int64          `json:"height"`    // 체크포인트 높이
	AppHash   bytes.HexBytes `json:"app_hash"`  // 애플리케이션 상태 해시
	RootHash  bytes.HexBytes `json:"root_hash"` // 상태 루트 해시
	Timestamp time.Time      `json:"timestamp"` // 체크포인트 생성 시간

	ValidatorSetHash     bytes.HexBytes `json:"validator_set_hash"`      // 검증자 세트 해시
	NextValidatorSetHash bytes.HexBytes `json:"next_validator_set_hash"` // 다음 검증자 세트 해시

	// 추가 필드: 상태 동기화에 필요한 메타데이터
	ConsensusParams []byte `json:"consensus_params,omitempty"` // 합의 파라미터
	StateSyncData   []byte `json:"state_sync_data,omitempty"`  // 상태 동기화 데이터
}

// Store는 체크포인트를 저장하고 조회하는 기능을 정의합니다.
type Store interface {
	// SaveCheckpoint는 새 체크포인트를 저장합니다.
	SaveCheckpoint(checkpoint *Checkpoint) error

	// GetCheckpoint는 지정된 높이의 체크포인트를 조회합니다.
	GetCheckpoint(height int64) (*Checkpoint, error)

	// GetLatestCheckpoint는 가장 최근 체크포인트를 반환합니다.
	GetLatestCheckpoint() (*Checkpoint, error)

	// GetCheckpointCount는 저장된 체크포인트 수를 반환합니다.
	GetCheckpointCount() (int64, error)

	// PruneCheckpoints는 특정 높이 이전의 체크포인트를 제거합니다.
	PruneCheckpoints(heightThreshold int64) error
}

// CheckpointInterval은 체크포인트 생성 간격을 정의합니다.
// 기본값은 100 블록마다 체크포인트 생성입니다.
var CheckpointInterval int64 = 100

// IsCheckpointHeight는 주어진 높이가 체크포인트 생성 높이인지 확인합니다.
func IsCheckpointHeight(height int64) bool {
	return height%CheckpointInterval == 0 && height > 0
}

// SetCheckpointInterval은 체크포인트 간격을 설정합니다.
func SetCheckpointInterval(interval int64) {
	if interval > 0 {
		CheckpointInterval = interval
	}
}
