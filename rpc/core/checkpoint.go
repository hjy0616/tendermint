package core

import (
	"fmt"
	"strconv"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/libs/bytes"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpc "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

var checkpointManager checkpoint.Manager

// SetCheckpointManager는 전역 체크포인트 매니저를 설정합니다.
func SetCheckpointManager(manager checkpoint.Manager) {
	checkpointManager = manager
}

// Checkpoint는 체크포인트를 조회하는 RPC 메서드입니다.
// URI: /checkpoint?number={number}
func Checkpoint(ctx *rpctypes.Context, numberStr string) (*ctypes.ResultCheckpoint, error) {
	if checkpointManager == nil {
		return nil, fmt.Errorf("checkpoint manager not initialized")
	}

	number, err := strconv.ParseUint(numberStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint number: %w", err)
	}

	cp, err := checkpointManager.GetCheckpoint(number)
	if err != nil {
		return nil, err
	}

	result := &ctypes.ResultCheckpoint{
		Number:       cp.Number,
		Height:       cp.Data.Height,
		BlockHash:    bytes.HexBytes(cp.Data.BlockHash),
		StateRoot:    bytes.HexBytes(cp.Data.StateRootHash),
		AppHash:      bytes.HexBytes(cp.Data.AppHash),
		ValidatorSet: bytes.HexBytes(cp.Data.ValidatorHash),
		Timestamp:    cp.Data.Timestamp,
	}

	if len(cp.Data.Signature) > 0 {
		result.Signature = bytes.HexBytes(cp.Data.Signature)
		result.Signer = bytes.HexBytes(cp.Data.SignerAddress)
	}

	return result, nil
}

// CheckpointLatest는 최신 체크포인트를 조회하는 RPC 메서드입니다.
// URI: /checkpoint/latest
func CheckpointLatest(ctx *rpctypes.Context) (*ctypes.ResultCheckpoint, error) {
	if checkpointManager == nil {
		return nil, fmt.Errorf("checkpoint manager not initialized")
	}

	cp, err := checkpointManager.GetLatestCheckpoint()
	if err != nil {
		return nil, err
	}

	result := &ctypes.ResultCheckpoint{
		Number:       cp.Number,
		Height:       cp.Data.Height,
		BlockHash:    bytes.HexBytes(cp.Data.BlockHash),
		StateRoot:    bytes.HexBytes(cp.Data.StateRootHash),
		AppHash:      bytes.HexBytes(cp.Data.AppHash),
		ValidatorSet: bytes.HexBytes(cp.Data.ValidatorHash),
		Timestamp:    cp.Data.Timestamp,
	}

	if len(cp.Data.Signature) > 0 {
		result.Signature = bytes.HexBytes(cp.Data.Signature)
		result.Signer = bytes.HexBytes(cp.Data.SignerAddress)
	}

	return result, nil
}

// CheckpointCount는 체크포인트 개수를 조회하는 RPC 메서드입니다.
// URI: /checkpoint/count
func CheckpointCount(ctx *rpctypes.Context) (*ctypes.ResultCheckpointCount, error) {
	if checkpointManager == nil {
		return nil, fmt.Errorf("checkpoint manager not initialized")
	}

	count, err := checkpointManager.GetCheckpointCount()
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultCheckpointCount{
		Count: count,
	}, nil
}

// CheckpointCreate는 새 체크포인트를 생성하는 RPC 메서드입니다.
// URI: /checkpoint/create?height={height}
func CheckpointCreate(ctx *rpctypes.Context, heightStr string) (*ctypes.ResultCheckpoint, error) {
	if checkpointManager == nil {
		return nil, fmt.Errorf("checkpoint manager not initialized")
	}

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid height: %w", err)
	}

	cp, err := checkpointManager.CreateCheckpoint(ctx.Context(), height)
	if err != nil {
		return nil, err
	}

	result := &ctypes.ResultCheckpoint{
		Number:       cp.Number,
		Height:       cp.Data.Height,
		BlockHash:    bytes.HexBytes(cp.Data.BlockHash),
		StateRoot:    bytes.HexBytes(cp.Data.StateRootHash),
		AppHash:      bytes.HexBytes(cp.Data.AppHash),
		ValidatorSet: bytes.HexBytes(cp.Data.ValidatorHash),
		Timestamp:    cp.Data.Timestamp,
	}

	if len(cp.Data.Signature) > 0 {
		result.Signature = bytes.HexBytes(cp.Data.Signature)
		result.Signer = bytes.HexBytes(cp.Data.SignerAddress)
	}

	return result, nil
}

// RegisterRPCRoutes는 체크포인트 RPC 라우트를 등록하는 함수입니다.
func RegisterRPCRoutes() {
	Routes["checkpoint"] = rpc.NewRPCFunc(Checkpoint, "number")
	Routes["checkpoint_latest"] = rpc.NewRPCFunc(CheckpointLatest, "")
	Routes["checkpoint_count"] = rpc.NewRPCFunc(CheckpointCount, "")
	Routes["checkpoint_create"] = rpc.NewRPCFunc(CheckpointCreate, "height")
}
