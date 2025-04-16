package coregrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/clerk"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/milestone"
	pb "github.com/tendermint/tendermint/rpc/grpc"
	"github.com/tendermint/tendermint/span"
	"google.golang.org/grpc"
)

// TendermintServer는 TendermintService gRPC 서버 구현체입니다.
type TendermintServer struct {
	pb.UnimplementedTendermintServiceServer
	logger        log.Logger
	checkpointMgr checkpoint.Manager
	milestoneMgr  milestone.Manager
	clerkMgr      clerk.Manager
	spanMgr       span.Manager
	consensus     consensus.Consensus
	mempool       mempool.Mempool
}

// NewTendermintServer는 새로운 TendermintServer를 생성합니다.
func NewTendermintServer(
	logger log.Logger,
	checkpointMgr checkpoint.Manager,
	milestoneMgr milestone.Manager,
	clerkMgr clerk.Manager,
	spanMgr span.Manager,
	consensus consensus.Consensus,
	mempool mempool.Mempool,
) *TendermintServer {
	return &TendermintServer{
		logger:        logger,
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
	grpcServer := grpc.NewServer()
	pb.RegisterTendermintServiceServer(grpcServer, server)

	return grpcServer, nil
}

// FetchCheckpoint는 체크포인트 데이터를 조회합니다.
func (s *TendermintServer) FetchCheckpoint(ctx context.Context, req *pb.RequestFetchCheckpoint) (*pb.ResponseFetchCheckpoint, error) {
	s.logger.Info("FetchCheckpoint called", "number", req.CheckpointNumber)

	checkpoint, err := s.checkpointMgr.GetCheckpoint(uint64(req.CheckpointNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %v", err)
	}

	// 체크포인트의 실제 필드에 맞게 조정
	return &pb.ResponseFetchCheckpoint{
		Checkpoint: &pb.Checkpoint{
			Number:     req.CheckpointNumber,
			Data:       checkpoint.StateRootHash,
			RootHash:   checkpoint.BlockHash,
			StartBlock: int64(checkpoint.Height),
			EndBlock:   int64(checkpoint.Height),
		},
	}, nil
}

// FetchCheckpointCount는 체크포인트 수를 조회합니다.
func (s *TendermintServer) FetchCheckpointCount(ctx context.Context, req *pb.RequestFetchCheckpointCount) (*pb.ResponseFetchCheckpointCount, error) {
	s.logger.Info("FetchCheckpointCount called")

	count, err := s.checkpointMgr.GetCheckpointCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint count: %v", err)
	}

	return &pb.ResponseFetchCheckpointCount{
		Count: int64(count),
	}, nil
}

// FetchMilestone는 마일스톤 데이터를 조회합니다.
func (s *TendermintServer) FetchMilestone(ctx context.Context, req *pb.RequestFetchMilestone) (*pb.ResponseFetchMilestone, error) {
	s.logger.Info("FetchMilestone called", "id", req.MilestoneId)

	// 임시 구현: 고정 데이터 반환
	return &pb.ResponseFetchMilestone{
		Milestone: &pb.Milestone{
			Id:         req.MilestoneId,
			Data:       []byte("milestone_data"),
			RootHash:   []byte("root_hash"),
			StartBlock: 1000,
			EndBlock:   2000,
			Ack:        false,
		},
	}, nil
}

// FetchMilestoneCount는 마일스톤 수를 조회합니다.
func (s *TendermintServer) FetchMilestoneCount(ctx context.Context, req *pb.RequestFetchMilestoneCount) (*pb.ResponseFetchMilestoneCount, error) {
	s.logger.Info("FetchMilestoneCount called")

	count, err := s.milestoneMgr.GetMilestoneCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone count: %v", err)
	}

	return &pb.ResponseFetchMilestoneCount{
		Count: int64(count),
	}, nil
}

// FetchNoAckMilestone는 승인되지 않은 마일스톤 ID 목록을 조회합니다.
func (s *TendermintServer) FetchNoAckMilestone(ctx context.Context, req *pb.RequestFetchNoAckMilestone) (*pb.ResponseFetchNoAckMilestone, error) {
	s.logger.Info("FetchNoAckMilestone called")

	ids, err := s.milestoneMgr.GetNoAckMilestones()
	if err != nil {
		return nil, fmt.Errorf("failed to get no ack milestones: %v", err)
	}

	// uint64 배열을 int64 배열로 변환
	int64Ids := make([]int64, len(ids))
	for i, id := range ids {
		int64Ids[i] = int64(id)
	}

	return &pb.ResponseFetchNoAckMilestone{
		MilestoneIds: int64Ids,
	}, nil
}

// StateSyncEvents는 특정 기간 동안의 상태 동기화 이벤트를 조회합니다.
func (s *TendermintServer) StateSyncEvents(ctx context.Context, req *pb.RequestStateSyncEvents) (*pb.ResponseStateSyncEvents, error) {
	s.logger.Info("StateSyncEvents called", "from", req.FromTime, "to", req.ToTime, "record_type", req.RecordType)

	records, err := s.clerkMgr.GetEventRecords(req.FromTime, req.ToTime, req.RecordType)
	if err != nil {
		return nil, fmt.Errorf("failed to get event records: %v", err)
	}

	// 이벤트 레코드 변환
	events := make([]*pb.Event, len(records))
	for i, record := range records {
		events[i] = &pb.Event{
			Id:         int64(record.ID),
			Time:       record.Timestamp,
			Data:       record.Data,
			RecordType: record.ChainID,
		}
	}

	return &pb.ResponseStateSyncEvents{
		Events: events,
	}, nil
}

// CommitStates는 이벤트 ID 목록을 커밋합니다.
func (s *TendermintServer) CommitStates(ctx context.Context, req *pb.RequestCommitStates) (*pb.ResponseCommitStates, error) {
	s.logger.Info("CommitStates called", "event_count", len(req.EventIds))

	// int64 배열을 uint64 배열로 변환
	uint64Ids := make([]uint64, len(req.EventIds))
	for i, id := range req.EventIds {
		uint64Ids[i] = uint64(id)
	}

	failedIds, err := s.clerkMgr.CommitEventRecords(uint64Ids)
	if err != nil {
		return nil, fmt.Errorf("failed to commit event records: %v", err)
	}

	// uint64 배열을 int64 배열로 변환
	int64FailedIds := make([]int64, len(failedIds))
	for i, id := range failedIds {
		int64FailedIds[i] = int64(id)
	}

	return &pb.ResponseCommitStates{
		FailedIds: int64FailedIds,
	}, nil
}

// Span은 지정된 ID의 스팬 데이터를 조회합니다.
func (s *TendermintServer) Span(ctx context.Context, req *pb.RequestSpan) (*pb.ResponseSpan, error) {
	s.logger.Info("Span called", "id", req.SpanId)

	// 임시 구현: 고정 데이터 반환
	return &pb.ResponseSpan{
		Span: &pb.Span{
			Id:        req.SpanId,
			Data:      []byte("span_data"),
			StartTime: 1000,
			EndTime:   2000,
		},
	}, nil
}

// BlockInfo는 지정된 높이의 블록 정보를 조회합니다.
func (s *TendermintServer) BlockInfo(ctx context.Context, req *pb.RequestBlockInfo) (*pb.ResponseBlockInfo, error) {
	s.logger.Info("BlockInfo called", "height", req.Height)

	// 임시 구현: 고정 데이터 반환
	return &pb.ResponseBlockInfo{
		Block: &pb.BlockInfo{
			Height:  req.Height,
			Hash:    []byte("block_hash"),
			Data:    []byte("block_data"),
			AppHash: []byte("app_hash"),
			Time:    time.Now().Unix(),
			NumTxs:  0,
		},
	}, nil
}

// MempoolInfo는 메모리풀의 현재 상태 정보를 조회합니다.
func (s *TendermintServer) MempoolInfo(ctx context.Context, req *pb.RequestMempoolInfo) (*pb.ResponseMempoolInfo, error) {
	s.logger.Info("MempoolInfo called")

	if s.mempool == nil {
		return nil, fmt.Errorf("mempool not available")
	}

	txs := s.mempool.ReapMaxTxs(-1)
	txInfos := make([]*pb.TxInfo, len(txs))

	var totalSize int64
	for i, tx := range txs {
		size := int32(len(tx))
		txInfos[i] = &pb.TxInfo{
			Hash: tx.Hash(),
			Size: size,
			// 실제 gas 정보는 메모리풀에서 가져와야 합니다
			GasWanted: 0,
			GasUsed:   0,
		}
		totalSize += int64(size)
	}

	return &pb.ResponseMempoolInfo{
		TxCount:   int32(len(txs)),
		TotalSize: totalSize,
		TxsInfo:   txInfos,
	}, nil
}
