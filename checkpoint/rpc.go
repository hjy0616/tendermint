package checkpoint

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/tendermint/tendermint/libs/log"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// RPCHandler는 체크포인트 관련 RPC 요청을 처리하는 핸들러입니다.
type RPCHandler struct {
	manager Manager
	logger  log.Logger
}

// NewRPCHandler는 새로운 RPCHandler를 생성합니다.
func NewRPCHandler(manager Manager) *RPCHandler {
	return &RPCHandler{
		manager: manager,
		logger:  log.NewNopLogger(),
	}
}

// SetLogger는 로거를 설정합니다.
func (h *RPCHandler) SetLogger(logger log.Logger) {
	h.logger = logger
	h.manager.SetLogger(logger)
}

// CheckpointResult는 체크포인트 RPC 응답 구조체입니다.
type CheckpointResult struct {
	Height       uint64 `json:"height"`
	Number       uint64 `json:"number"`
	BlockHash    string `json:"block_hash"`
	StateRoot    string `json:"state_root"`
	AppHash      string `json:"app_hash"`
	ValidatorSet string `json:"validator_set"`
	Timestamp    int64  `json:"timestamp"`
	Signature    string `json:"signature,omitempty"`
	Signer       string `json:"signer,omitempty"`
}

// CheckpointCountResult는 체크포인트 개수 RPC 응답 구조체입니다.
type CheckpointCountResult struct {
	Count uint64 `json:"count"`
}

// FetchCheckpoint는 특정 번호의 체크포인트를 조회합니다.
// URI: /checkpoint?number={number}
func (h *RPCHandler) FetchCheckpoint(ctx *rpctypes.Context, numberParam string) (*CheckpointResult, error) {
	number, err := strconv.ParseUint(numberParam, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint number: %w", err)
	}

	h.logger.Debug("RPC FetchCheckpoint called", "number", number)

	checkpoint, err := h.manager.GetCheckpoint(number)
	if err != nil {
		if errors.Is(err, ErrCheckpointNotFound) {
			return nil, rpctypes.RPCError{
				Code:    http.StatusNotFound,
				Message: "checkpoint not found",
				Data:    fmt.Sprintf("checkpoint number %d not found", number),
			}
		}
		return nil, err
	}

	result := &CheckpointResult{
		Height:       checkpoint.Data.Height,
		Number:       checkpoint.Number,
		BlockHash:    fmt.Sprintf("%X", checkpoint.Data.BlockHash),
		StateRoot:    fmt.Sprintf("%X", checkpoint.Data.StateRootHash),
		AppHash:      fmt.Sprintf("%X", checkpoint.Data.AppHash),
		ValidatorSet: fmt.Sprintf("%X", checkpoint.Data.ValidatorHash),
		Timestamp:    checkpoint.Data.Timestamp,
	}

	if len(checkpoint.Data.Signature) > 0 {
		result.Signature = fmt.Sprintf("%X", checkpoint.Data.Signature)
		result.Signer = fmt.Sprintf("%X", checkpoint.Data.SignerAddress)
	}

	return result, nil
}

// FetchLatestCheckpoint는 가장 최근 체크포인트를 조회합니다.
// URI: /checkpoint/latest
func (h *RPCHandler) FetchLatestCheckpoint(ctx *rpctypes.Context) (*CheckpointResult, error) {
	h.logger.Debug("RPC FetchLatestCheckpoint called")

	checkpoint, err := h.manager.GetLatestCheckpoint()
	if err != nil {
		if errors.Is(err, ErrCheckpointNotFound) {
			return nil, rpctypes.RPCError{
				Code:    http.StatusNotFound,
				Message: "no checkpoints available",
			}
		}
		return nil, err
	}

	result := &CheckpointResult{
		Height:       checkpoint.Data.Height,
		Number:       checkpoint.Number,
		BlockHash:    fmt.Sprintf("%X", checkpoint.Data.BlockHash),
		StateRoot:    fmt.Sprintf("%X", checkpoint.Data.StateRootHash),
		AppHash:      fmt.Sprintf("%X", checkpoint.Data.AppHash),
		ValidatorSet: fmt.Sprintf("%X", checkpoint.Data.ValidatorHash),
		Timestamp:    checkpoint.Data.Timestamp,
	}

	if len(checkpoint.Data.Signature) > 0 {
		result.Signature = fmt.Sprintf("%X", checkpoint.Data.Signature)
		result.Signer = fmt.Sprintf("%X", checkpoint.Data.SignerAddress)
	}

	return result, nil
}

// FetchCheckpointCount는 체크포인트 개수를 조회합니다.
// URI: /checkpoint/count
func (h *RPCHandler) FetchCheckpointCount(ctx *rpctypes.Context) (*CheckpointCountResult, error) {
	h.logger.Debug("RPC FetchCheckpointCount called")

	count, err := h.manager.GetCheckpointCount()
	if err != nil {
		return nil, err
	}

	return &CheckpointCountResult{
		Count: count,
	}, nil
}

// CreateCheckpoint는 새 체크포인트를 생성합니다.
// URI: /checkpoint/create?height={height}
func (h *RPCHandler) CreateCheckpoint(ctx *rpctypes.Context, heightParam string) (*CheckpointResult, error) {
	height, err := strconv.ParseUint(heightParam, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid height: %w", err)
	}

	h.logger.Debug("RPC CreateCheckpoint called", "height", height)

	checkpoint, err := h.manager.CreateCheckpoint(context.Background(), height)
	if err != nil {
		return nil, err
	}

	result := &CheckpointResult{
		Height:       checkpoint.Data.Height,
		Number:       checkpoint.Number,
		BlockHash:    fmt.Sprintf("%X", checkpoint.Data.BlockHash),
		StateRoot:    fmt.Sprintf("%X", checkpoint.Data.StateRootHash),
		AppHash:      fmt.Sprintf("%X", checkpoint.Data.AppHash),
		ValidatorSet: fmt.Sprintf("%X", checkpoint.Data.ValidatorHash),
		Timestamp:    checkpoint.Data.Timestamp,
	}

	if len(checkpoint.Data.Signature) > 0 {
		result.Signature = fmt.Sprintf("%X", checkpoint.Data.Signature)
		result.Signer = fmt.Sprintf("%X", checkpoint.Data.SignerAddress)
	}

	return result, nil
}
