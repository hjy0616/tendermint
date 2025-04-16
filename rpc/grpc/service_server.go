package coregrpc

import (
	"context"
	"fmt"
	"net"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/milestone"
	"github.com/tendermint/tendermint/clerk"
	"github.com/tendermint/tendermint/span"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/libs/log"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// TendermintServer는 TendermintService gRPC 서버 구현체입니다.
type TendermintServer struct {
	UnimplementedTendermintServiceServer
	logger       log.Logger
	checkpointMgr checkpoint.Manager
	milestoneMgr  milestone.Manager
	clerkMgr      clerk.Manager
	spanMgr       span.Manager
	consensus     *consensus.State
	mempool       mempool.Mempool
}

// NewTendermintServer는 새로운 TendermintServer를 생성합니다.
func NewTendermintServer(
	logger log.Logger,
	checkpointMgr checkpoint.Manager,
	milestoneMgr milestone.Manager,
	clerkMgr clerk.Manager,
	spanMgr span.Manager,
	consensus *consensus.State,
	mempool mempool.Mempool,
) *TendermintServer {
	return &TendermintServer{
		logger:       logger,
		checkpointMgr: checkpointMgr,
		milestoneMgr:  milestoneMgr,
		clerkMgr:      clerkMgr,
		spanMgr:       spanMgr,
		consensus:     consensus,
		mempool:       mempool,
	}
}

// StartGRPCServer는 gRPC 서버를 시작합니다.
func StartGRPCServer(
	server *TendermintServer,
	addr string,
) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer()
	RegisterTendermintServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			server.logger.Error("gRPC server failed", "err", err)
		}
	}()

	server.logger.Info("gRPC server started", "addr", addr)
	return grpcServer, nil
}

// FetchCheckpoint는 체크포인트 데이터를 가져옵니다.
func (s *TendermintServer) FetchCheckpoint(ctx context.Context, req *RequestFetchCheckpoint) (*ResponseFetchCheckpoint, error) {
	s.logger.Info("FetchCheckpoint called", "checkpoint_number", req.CheckpointNumber)
	
	checkpoint, err := s.checkpointMgr.GetCheckpoint(uint64(req.CheckpointNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}
	
	return &ResponseFetchCheckpoint{
		Data:       checkpoint.Data,
		RootHash:   checkpoint.RootHash,
		StartBlock: int64(checkpoint.StartBlock),
		EndBlock:   int64(checkpoint.EndBlock),
	}, nil
}

// FetchCheckpointCount는 체크포인트 수를 가져옵니다.
func (s *TendermintServer) FetchCheckpointCount(ctx context.Context, req *RequestFetchCheckpointCount) (*ResponseFetchCheckpointCount, error) {
	s.logger.Info("FetchCheckpointCount called")
	
	count, err := s.checkpointMgr.GetCheckpointCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint count: %w", err)
	}
	
	return &ResponseFetchCheckpointCount{
		Count: int64(count),
	}, nil
}

// FetchMilestone는 마일스톤 데이터를 가져옵니다.
func (s *TendermintServer) FetchMilestone(ctx context.Context, req *RequestFetchMilestone) (*ResponseFetchMilestone, error) {
	s.logger.Info("FetchMilestone called", "milestone_id", req.MilestoneId)
	
	milestone, err := s.milestoneMgr.GetMilestone(uint64(req.MilestoneId))
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}
	
	return &ResponseFetchMilestone{
		Data:       milestone.Data,
		RootHash:   milestone.RootHash,
		StartBlock: int64(milestone.StartBlock),
		EndBlock:   int64(milestone.EndBlock),
	}, nil
}

// FetchMilestoneCount는 마일스톤 수를 가져옵니다.
func (s *TendermintServer) FetchMilestoneCount(ctx context.Context, req *RequestFetchMilestoneCount) (*ResponseFetchMilestoneCount, error) {
	s.logger.Info("FetchMilestoneCount called")
	
	count, err := s.milestoneMgr.GetMilestoneCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone count: %w", err)
	}
	
	return &ResponseFetchMilestoneCount{
		Count: int64(count),
	}, nil
}

// FetchNoAckMilestone는 승인되지 않은 마일스톤 ID 목록을 가져옵니다.
func (s *TendermintServer) FetchNoAckMilestone(ctx context.Context, req *RequestFetchNoAckMilestone) (*ResponseFetchNoAckMilestone, error) {
	s.logger.Info("FetchNoAckMilestone called")
	
	ids, err := s.milestoneMgr.GetNoAckMilestones()
	if err != nil {
		return nil, fmt.Errorf("failed to get no-ack milestones: %w", err)
	}
	
	var milestoneIDs []int64
	for _, id := range ids {
		milestoneIDs = append(milestoneIDs, int64(id))
	}
	
	return &ResponseFetchNoAckMilestone{
		MilestoneIds: milestoneIDs,
	}, nil
}

// StateSyncEvents는 특정 기간 동안의 상태 동기화 이벤트를 가져옵니다.
func (s *TendermintServer) StateSyncEvents(ctx context.Context, req *RequestStateSyncEvents) (*ResponseStateSyncEvents, error) {
	s.logger.Info("StateSyncEvents called", "from", req.FromTime, "to", req.ToTime, "type", req.RecordType)
	
	events, err := s.clerkMgr.GetEventRecords(req.FromTime, req.ToTime, req.RecordType)
	if err != nil {
		return nil, fmt.Errorf("failed to get state sync events: %w", err)
	}
	
	var records []*EventRecord
	for _, event := range events {
		records = append(records, &EventRecord{
			Id:         uint64(event.ID),
			Time:       event.Time,
			Data:       event.Data,
			RecordType: event.RecordType,
		})
	}
	
	return &ResponseStateSyncEvents{
		Records: records,
	}, nil
}

// CommitStates는 이벤트 ID 목록을 커밋합니다.
func (s *TendermintServer) CommitStates(ctx context.Context, req *RequestCommitStates) (*ResponseCommitStates, error) {
	s.logger.Info("CommitStates called", "event_count", len(req.EventIds))
	
	var eventIDs []uint64
	for _, id := range req.EventIds {
		eventIDs = append(eventIDs, uint64(id))
	}
	
	failed, err := s.clerkMgr.CommitEventRecords(eventIDs)
	if err != nil {
		return &ResponseCommitStates{
			Success:        false,
			FailedEventIds: req.EventIds,
			ErrorMessage:   err.Error(),
		}, nil
	}
	
	var failedIDs []int64
	for _, id := range failed {
		failedIDs = append(failedIDs, int64(id))
	}
	
	return &ResponseCommitStates{
		Success:        len(failedIDs) == 0,
		FailedEventIds: failedIDs,
		ErrorMessage:   "",
	}, nil
}

// Span은 지정된 ID의 스팬 데이터를 가져옵니다.
func (s *TendermintServer) Span(ctx context.Context, req *RequestSpan) (*ResponseSpan, error) {
	s.logger.Info("Span called", "span_id", req.SpanId)
	
	span, err := s.spanMgr.GetSpan(uint64(req.SpanId))
	if err != nil {
		return nil, fmt.Errorf("failed to get span: %w", err)
	}
	
	var validators []*tendermint.types.Validator
	var powers []uint64
	
	for i, validator := range span.Validators {
		validators = append(validators, validator)
		powers = append(powers, span.ValidatorPowers[i])
	}
	
	return &ResponseSpan{
		SpanId:          req.SpanId,
		StartBlock:      int64(span.StartBlock),
		EndBlock:        int64(span.EndBlock),
		Validators:      validators,
		ValidatorPowers: powers,
	}, nil
}

// BlockInfo는 지정된 높이의 블록 정보를 가져옵니다.
func (s *TendermintServer) BlockInfo(ctx context.Context, req *RequestBlockInfo) (*ResponseBlockInfo, error) {
	s.logger.Info("BlockInfo called", "height", req.Height)
	
	block := s.consensus.GetBlockStore().LoadBlock(req.Height)
	if block == nil {
		return nil, fmt.Errorf("block at height %d not found", req.Height)
	}
	
	lastCommit := block.LastCommit
	
	var validators []string
	for _, validator := range s.consensus.GetValidators() {
		validators = append(validators, validator.Address.String())
	}
	
	return &ResponseBlockInfo{
		Block:           block,
		BlockId:         block.Hash().String(),
		AppHash:         fmt.Sprintf("%X", block.AppHash),
		LastCommitRound: int64(lastCommit.Round),
		Validators:      validators,
	}, nil
}

// MempoolInfo는 메모리풀의 현재 상태 정보를 가져옵니다.
func (s *TendermintServer) MempoolInfo(ctx context.Context, req *RequestMempoolInfo) (*ResponseMempoolInfo, error) {
	s.logger.Info("MempoolInfo called")
	
	txs := s.mempool.ReapMaxTxs(-1)
	
	return &ResponseMempoolInfo{
		Size:       int64(s.mempool.Size()),
		MaxSize:    int64(s.mempool.MaxSize()),
		PendingTxs: txs,
	}, nil
} 