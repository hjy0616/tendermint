package span

import (
	"errors"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// RegisterRPCMethods는 스팬 관리 관련 RPC 메서드를 등록합니다.
func RegisterRPCMethods(mgr *Manager, logger log.Logger) {
	// 스팬 관련 RPC 메서드 등록
	core.Routes["span"] = rpcserver.NewRPCFunc(GetSpanByID(mgr, logger), "id")
	core.Routes["span_current"] = rpcserver.NewRPCFunc(GetCurrentSpan(mgr, logger), "")
	core.Routes["span_by_block"] = rpcserver.NewRPCFunc(GetSpanByBlock(mgr, logger), "height")
	core.Routes["span_validators"] = rpcserver.NewRPCFunc(GetSpanValidators(mgr, logger), "id")

	logger.Info("Registered span RPC methods")
}

// SpanResponse는 스팬 RPC 응답 구조체입니다.
type SpanResponse struct {
	ID         uint64    `json:"span_id"`
	StartBlock uint64    `json:"start_block"`
	EndBlock   uint64    `json:"end_block"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time,omitempty"`
}

// GetSpanByID는 ID로 스팬을 조회하는 RPC 핸들러를 반환합니다.
func GetSpanByID(mgr *Manager, logger log.Logger) func(*rpctypes.Context, uint64) (*ctypes.ResultSpan, error) {
	return func(ctx *rpctypes.Context, id uint64) (*ctypes.ResultSpan, error) {
		span, err := mgr.GetSpanByID(id)
		if err != nil {
			return nil, err
		}

		// 스팬 상태 확인
		status := getSpanStatus(span)

		// 검증자 수 계산
		validatorCount := len(span.Validators.Validators)

		logger.Debug("Retrieved span by ID", "id", id, "status", status, "validators", validatorCount)

		return &ctypes.ResultSpan{
			SpanID:         span.ID,
			StartBlock:     span.StartBlock,
			EndBlock:       span.EndBlock,
			Status:         status,
			StartTime:      span.StartTime,
			EndTime:        span.EndTime,
			ValidatorCount: validatorCount,
		}, nil
	}
}

// GetCurrentSpan은 현재 활성 스팬을 반환하는 RPC 핸들러를 반환합니다.
func GetCurrentSpan(mgr *Manager, logger log.Logger) func(*rpctypes.Context) (*ctypes.ResultSpan, error) {
	return func(ctx *rpctypes.Context) (*ctypes.ResultSpan, error) {
		span := mgr.GetCurrentSpan()
		if span == nil {
			return nil, errors.New("no active span")
		}

		// 스팬 상태 확인
		status := getSpanStatus(span)

		// 검증자 수 계산
		validatorCount := len(span.Validators.Validators)

		logger.Debug("Retrieved current span", "id", span.ID, "status", status, "validators", validatorCount)

		return &ctypes.ResultSpan{
			SpanID:         span.ID,
			StartBlock:     span.StartBlock,
			EndBlock:       span.EndBlock,
			Status:         status,
			StartTime:      span.StartTime,
			EndTime:        span.EndTime,
			ValidatorCount: validatorCount,
		}, nil
	}
}

// GetSpanByBlock은 블록 높이로 스팬을 조회하는 RPC 핸들러를 반환합니다.
func GetSpanByBlock(mgr *Manager, logger log.Logger) func(*rpctypes.Context, int64) (*ctypes.ResultSpan, error) {
	return func(ctx *rpctypes.Context, height int64) (*ctypes.ResultSpan, error) {
		if height < 0 {
			return nil, fmt.Errorf("negative height %d", height)
		}

		span, err := mgr.GetSpanByBlockNumber(uint64(height))
		if err != nil {
			return nil, err
		}

		// 스팬 상태 확인
		status := getSpanStatus(span)

		// 검증자 수 계산
		validatorCount := len(span.Validators.Validators)

		logger.Debug("Retrieved span by block height", "height", height, "spanID", span.ID, "status", status)

		return &ctypes.ResultSpan{
			SpanID:         span.ID,
			StartBlock:     span.StartBlock,
			EndBlock:       span.EndBlock,
			Status:         status,
			StartTime:      span.StartTime,
			EndTime:        span.EndTime,
			ValidatorCount: validatorCount,
		}, nil
	}
}

// GetSpanValidators는 지정된 스팬의 검증자 세트를 반환하는 RPC 핸들러를 반환합니다.
func GetSpanValidators(mgr *Manager, logger log.Logger) func(*rpctypes.Context, uint64) (*ctypes.ResultValidators, error) {
	return func(ctx *rpctypes.Context, id uint64) (*ctypes.ResultValidators, error) {
		span, err := mgr.GetSpanByID(id)
		if err != nil {
			return nil, err
		}

		logger.Debug("Retrieved span validators", "spanID", id, "count", len(span.Validators.Validators))

		return &ctypes.ResultValidators{
			BlockHeight: int64(span.StartBlock),
			Validators:  span.Validators.Validators,
			Total:       len(span.Validators.Validators),
		}, nil
	}
}

// getSpanStatus는 스팬의 현재 상태를 문자열로 반환합니다.
func getSpanStatus(span *Span) string {
	now := time.Now()

	// 현재 시간이 스팬 시작 시간보다 이전인 경우
	if now.Before(span.StartTime) {
		return "pending"
	}

	// 종료 시간이 설정되었고 현재 시간이 종료 시간 이후인 경우
	if !span.EndTime.IsZero() && now.After(span.EndTime) {
		return "completed"
	}

	// 그 외의 경우는 활성 상태
	return "active"
}
