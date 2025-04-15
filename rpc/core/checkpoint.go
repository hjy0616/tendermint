package core

import (
	"fmt"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/checkpoint/types"
	"github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// Checkpoint는 특정 높이의 체크포인트를 조회합니다.
// 높이를 지정하지 않으면 가장 최근 체크포인트를 반환합니다.
// ```shell
// curl -X GET "http://localhost:26657/checkpoint?height=1000"
// ```
func Checkpoint(ctx *rpctypes.Context, heightPtr *int64) (*types.Checkpoint, error) {
	if !checkpoint.IsInitialized() {
		return nil, fmt.Errorf("checkpoint module not initialized")
	}

	// 높이가 지정되지 않은 경우 최신 체크포인트 반환
	if heightPtr == nil {
		latest, err := checkpoint.GetLatestCheckpoint()
		if err != nil {
			return nil, fmt.Errorf("failed to get latest checkpoint: %w", err)
		}
		return latest, nil
	}

	height := *heightPtr
	if height <= 0 {
		return nil, fmt.Errorf("height must be greater than 0, got %d", height)
	}

	// 특정 높이의 체크포인트 조회
	cp, err := checkpoint.GetCheckpoint(height)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint at height %d: %w", height, err)
	}

	return cp, nil
}

// CheckpointCount는 저장된 체크포인트 수를 반환합니다.
// ```shell
// curl -X GET "http://localhost:26657/checkpoint_count"
// ```
func CheckpointCount(ctx *rpctypes.Context) (int64, error) {
	if !checkpoint.IsInitialized() {
		return 0, fmt.Errorf("checkpoint module not initialized")
	}

	count, err := checkpoint.GetCheckpointCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get checkpoint count: %w", err)
	}

	return count, nil
}

// CreateCheckpoint는 현재 블록에서 체크포인트를 수동으로 생성합니다.
// 일반적으로 체크포인트는 자동으로 생성되지만, 특별한 상황에서 수동 생성이 필요할 수 있습니다.
// 참고: 이 메서드는 관리자 전용이며, 필요한 경우 액세스 제한을 추가해야 합니다.
// ```shell
// curl -X POST "http://localhost:26657/create_checkpoint" -d '{"height": 1000, "app_hash": "hash", "root_hash": "hash"}'
// ```
func CreateCheckpoint(
	ctx *rpctypes.Context,
	height int64,
	appHash []byte,
	rootHash []byte,
	validatorSetHash []byte,
	nextValidatorSetHash []byte,
) (*types.Checkpoint, error) {
	if !checkpoint.IsInitialized() {
		return nil, fmt.Errorf("checkpoint module not initialized")
	}

	if height <= 0 {
		return nil, fmt.Errorf("height must be greater than 0, got %d", height)
	}

	// 체크포인트 생성
	cp, err := checkpoint.CreateCheckpoint(height, appHash, rootHash, validatorSetHash, nextValidatorSetHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	if cp == nil {
		return nil, fmt.Errorf("not a checkpoint height: %d", height)
	}

	return cp, nil
}

// PruneCheckpoints는 특정 높이 이전의 체크포인트를 제거합니다.
// 참고: 이 메서드는 관리자 전용이며, 필요한 경우 액세스 제한을 추가해야 합니다.
// ```shell
// curl -X POST "http://localhost:26657/prune_checkpoints" -d '{"height": 1000}'
// ```
func PruneCheckpoints(ctx *rpctypes.Context, height int64) (int64, error) {
	if !checkpoint.IsInitialized() {
		return 0, fmt.Errorf("checkpoint module not initialized")
	}

	if height <= 0 {
		return 0, fmt.Errorf("height must be greater than 0, got %d", height)
	}

	count, err := checkpoint.GetCheckpointCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get checkpoint count: %w", err)
	}

	if err := checkpoint.PruneCheckpoints(height); err != nil {
		return 0, fmt.Errorf("failed to prune checkpoints: %w", err)
	}

	newCount, err := checkpoint.GetCheckpointCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get updated checkpoint count: %w", err)
	}

	return count - newCount, nil
}
