package coregrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/clerk"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/milestone"
	"github.com/tendermint/tendermint/span"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	// 자동 생성된 proto 코드 임포트
)

// Proto에서 생성된 서버 인터페이스와 요청/응답 타입을 정의합니다.
// 실제 구현에서는 proto로 자동 생성된 파일에서 이런 타입들을 임포트해야 합니다.

// UnimplementedTendermintServiceServer는 모든 gRPC 메서드를 구현하는 기본 서버입니다.
type UnimplementedTendermintServiceServer struct{}

// TendermintServiceServer는 gRPC 서버 인터페이스입니다.
type TendermintServiceServer interface {
	FetchCheckpoint(context.Context, *RequestFetchCheckpoint) (*ResponseFetchCheckpoint, error)
	FetchCheckpointCount(context.Context, *RequestFetchCheckpointCount) (*ResponseFetchCheckpointCount, error)
	FetchMilestone(context.Context, *RequestFetchMilestone) (*ResponseFetchMilestone, error)
	FetchMilestoneCount(context.Context, *RequestFetchMilestoneCount) (*ResponseFetchMilestoneCount, error)
	FetchNoAckMilestone(context.Context, *RequestFetchNoAckMilestone) (*ResponseFetchNoAckMilestone, error)
	StateSyncEvents(context.Context, *RequestStateSyncEvents) (*ResponseStateSyncEvents, error)
	CommitStates(context.Context, *RequestCommitStates) (*ResponseCommitStates, error)
	Span(context.Context, *RequestSpan) (*ResponseSpan, error)
	BlockInfo(context.Context, *RequestBlockInfo) (*ResponseBlockInfo, error)
	MempoolInfo(context.Context, *RequestMempoolInfo) (*ResponseMempoolInfo, error)
}

// RegisterTendermintServiceServer는 gRPC 서버에 TendermintServiceServer를 등록합니다.
func RegisterTendermintServiceServer(s *grpc.Server, srv TendermintServiceServer) {
	s.RegisterService(&_TendermintService_serviceDesc, srv)
}

var _TendermintService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tendermint.rpc.grpc.TendermintService",
	HandlerType: (*TendermintServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchCheckpoint",
			Handler:    _TendermintService_FetchCheckpoint_Handler,
		},
		// 다른 메서드들은 필요에 따라 추가하세요
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tendermint/rpc/grpc/service.proto",
}

// Handler 메서드는 구현체와 연결하기 위한 것입니다.
func _TendermintService_FetchCheckpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestFetchCheckpoint)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TendermintServiceServer).FetchCheckpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tendermint.rpc.grpc.TendermintService/FetchCheckpoint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TendermintServiceServer).FetchCheckpoint(ctx, req.(*RequestFetchCheckpoint))
	}
	return interceptor(ctx, in, info, handler)
}

// 프로토버프에서 생성될 요청 및 응답 타입들을 정의합니다.
type RequestFetchCheckpoint struct {
	CheckpointNumber int64
}

type ResponseFetchCheckpoint struct {
	Checkpoint *Checkpoint
}

type RequestFetchCheckpointCount struct{}

type ResponseFetchCheckpointCount struct {
	Count int64
}

type RequestFetchMilestone struct {
	MilestoneId int64
}

type ResponseFetchMilestone struct {
	Milestone *Milestone
}

type RequestFetchMilestoneCount struct{}

type ResponseFetchMilestoneCount struct {
	Count int64
}

type RequestFetchNoAckMilestone struct{}

type ResponseFetchNoAckMilestone struct {
	MilestoneIds []int64
}

type EventRecord struct {
	Id         uint64
	Time       int64
	Data       []byte
	RecordType string
}

type RequestStateSyncEvents struct {
	FromTime   int64
	ToTime     int64
	RecordType string
}

type ResponseStateSyncEvents struct {
	Events []*Event
}

type RequestCommitStates struct {
	EventIds []int64
}

type ResponseCommitStates struct {
	FailedIds []int64
}

type RequestSpan struct {
	SpanId int64
}

type ResponseSpan struct {
	Span *Span
}

type RequestBlockInfo struct {
	Height int64
}

type ResponseBlockInfo struct {
	Block *BlockInfo
}

type RequestMempoolInfo struct{}

type ResponseMempoolInfo struct {
	TxCount   int32
	TotalSize int64
	TxsInfo   []*TxInfo
}

// pb는 프로토콜 버퍼로 생성된 코드의 별칭입니다.
// 실제 프로젝트에서는 올바른 import 경로로 대체해야 합니다.
// 예: "github.com/tendermint/tendermint/rpc/grpcgen"
type pb struct{}

// 프로토콜 버퍼로 생성된 타입 정의
type Checkpoint struct {
	Number     int64
	Data       []byte
	RootHash   []byte
	StartBlock int64
	EndBlock   int64
}

type Milestone struct {
	Id         int64
	Data       []byte
	RootHash   []byte
	StartBlock int64
	EndBlock   int64
	Ack        bool
}

type Event struct {
	Id         int64
	Time       int64
	Data       []byte
	RecordType string
}

type Span struct {
	Id        int64
	Data      []byte
	StartTime int64
	EndTime   int64
}

type BlockInfo struct {
	Height  int64
	Hash    []byte
	Data    []byte
	AppHash []byte
	Time    int64
	NumTxs  int32
}

type TxInfo struct {
	Hash      []byte
	Size      int32
	GasWanted int64
	GasUsed   int64
}

// TendermintServer는 TendermintService gRPC 서버 구현체입니다.
type TendermintServer struct {
	UnimplementedTendermintServiceServer
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

// 서버 시작 함수는 한 번만 정의되어야 합니다
// StartGRPCServer 함수가 client_server.go에 이미 정의되어 있으므로
// 혼동을 피하기 위해 이 함수 이름을 변경합니다.
func StartTendermintGRPCServer(
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

// FetchCheckpoint는 체크포인트 데이터를 조회합니다.
func (s *TendermintServer) FetchCheckpoint(ctx context.Context, req *RequestFetchCheckpoint) (*ResponseFetchCheckpoint, error) {
	s.logger.Info("FetchCheckpoint called", "number", req.CheckpointNumber)

	checkpoint, err := s.checkpointMgr.GetCheckpoint(uint64(req.CheckpointNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %v", err)
	}

	return &ResponseFetchCheckpoint{
		Checkpoint: &Checkpoint{
			Number:     req.CheckpointNumber,
			Data:       checkpoint.StateRootHash,
			RootHash:   checkpoint.BlockHash,
			StartBlock: int64(checkpoint.Height),
			EndBlock:   int64(checkpoint.Height),
		},
	}, nil
}

// FetchCheckpointCount는 체크포인트 수를 조회합니다.
func (s *TendermintServer) FetchCheckpointCount(ctx context.Context, req *RequestFetchCheckpointCount) (*ResponseFetchCheckpointCount, error) {
	s.logger.Info("FetchCheckpointCount called")

	count, err := s.checkpointMgr.GetCheckpointCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint count: %v", err)
	}

	return &ResponseFetchCheckpointCount{
		Count: int64(count),
	}, nil
}

// FetchMilestone는 마일스톤 데이터를 조회합니다.
func (s *TendermintServer) FetchMilestone(ctx context.Context, req *RequestFetchMilestone) (*ResponseFetchMilestone, error) {
	s.logger.Info("FetchMilestone called", "id", req.MilestoneId)

	milestone, err := s.milestoneMgr.GetMilestone(uint64(req.MilestoneId))
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %v", err)
	}

	return &ResponseFetchMilestone{
		Milestone: &Milestone{
			Id:         req.MilestoneId,
			Data:       []byte{},
			RootHash:   []byte{},
			StartBlock: int64(milestone.StartBlock),
			EndBlock:   int64(milestone.EndBlock),
			Ack:        milestone.Ack,
		},
	}, nil
}

// FetchMilestoneCount는 마일스톤 수를 조회합니다.
func (s *TendermintServer) FetchMilestoneCount(ctx context.Context, req *RequestFetchMilestoneCount) (*ResponseFetchMilestoneCount, error) {
	s.logger.Info("FetchMilestoneCount called")

	count, err := s.milestoneMgr.GetMilestoneCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone count: %v", err)
	}

	return &ResponseFetchMilestoneCount{
		Count: int64(count),
	}, nil
}

// FetchNoAckMilestone는 승인되지 않은 마일스톤 ID 목록을 조회합니다.
func (s *TendermintServer) FetchNoAckMilestone(ctx context.Context, req *RequestFetchNoAckMilestone) (*ResponseFetchNoAckMilestone, error) {
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

	return &ResponseFetchNoAckMilestone{
		MilestoneIds: int64Ids,
	}, nil
}

// StateSyncEvents는 특정 기간 동안의 상태 동기화 이벤트를 조회합니다.
func (s *TendermintServer) StateSyncEvents(ctx context.Context, req *RequestStateSyncEvents) (*ResponseStateSyncEvents, error) {
	s.logger.Info("StateSyncEvents called", "from", req.FromTime, "to", req.ToTime, "record_type", req.RecordType)

	records, err := s.clerkMgr.GetEventRecords(req.FromTime, req.ToTime, req.RecordType)
	if err != nil {
		return nil, fmt.Errorf("failed to get event records: %v", err)
	}

	// 이벤트 레코드 변환
	events := make([]*Event, len(records))
	for i, record := range records {
		events[i] = &Event{
			Id:         int64(record.ID),
			Time:       req.FromTime,
			Data:       record.Data,
			RecordType: req.RecordType,
		}
	}

	return &ResponseStateSyncEvents{
		Events: events,
	}, nil
}

// CommitStates는 이벤트 ID 목록을 커밋합니다.
func (s *TendermintServer) CommitStates(ctx context.Context, req *RequestCommitStates) (*ResponseCommitStates, error) {
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

	return &ResponseCommitStates{
		FailedIds: int64FailedIds,
	}, nil
}

// Span은 지정된 ID의 스팬 데이터를 조회합니다.
func (s *TendermintServer) Span(ctx context.Context, req *RequestSpan) (*ResponseSpan, error) {
	s.logger.Info("Span called", "id", req.SpanId)

	spanData, err := s.spanMgr.GetSpan(uint64(req.SpanId))
	if err != nil {
		return nil, fmt.Errorf("failed to get span: %v", err)
	}

	return &ResponseSpan{
		Span: &Span{
			Id:        req.SpanId,
			Data:      []byte("span_data"), // 실제 span.Span에는 Data 필드가 없으므로 임시 값 사용
			StartTime: int64(spanData.StartBlock),
			EndTime:   int64(spanData.EndBlock),
		},
	}, nil
}

// BlockInfo는 지정된 높이의 블록 정보를 조회합니다.
func (s *TendermintServer) BlockInfo(ctx context.Context, req *RequestBlockInfo) (*ResponseBlockInfo, error) {
	s.logger.Info("BlockInfo called", "height", req.Height)

	// 더미 구현 - 실제 구현에서는 의존성 및 타입에 맞게 수정해야 함
	blockInfo := &BlockInfo{
		Height:  req.Height,
		Hash:    []byte("block_hash_placeholder"),
		Data:    []byte("block_data_placeholder"),
		AppHash: []byte("app_hash_placeholder"),
		Time:    time.Now().Unix(),
		NumTxs:  0,
	}

	return &ResponseBlockInfo{
		Block: blockInfo,
	}, nil
}

// MempoolInfo는 mempool의 현재 상태 정보를 조회합니다.
func (s *TendermintServer) MempoolInfo(ctx context.Context, req *RequestMempoolInfo) (*ResponseMempoolInfo, error) {
	s.logger.Info("MempoolInfo called")

	// 더미 구현 - 실제 구현에서는 의존성 및 타입에 맞게 수정해야 함
	txInfo := &TxInfo{
		Hash:      []byte("tx_hash_placeholder"),
		Size:      100,
		GasWanted: 1000,
		GasUsed:   800,
	}

	return &ResponseMempoolInfo{
		TxCount:   1,
		TotalSize: 100,
		TxsInfo:   []*TxInfo{txInfo},
	}, nil
}
