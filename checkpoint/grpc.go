package checkpoint

import (
	"context"

	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer는 체크포인트 관련 GRPC 서비스 핸들러입니다.
type GRPCServer struct {
	manager Manager
	logger  log.Logger
}

// NewGRPCServer는 새로운 GRPCServer를 생성합니다.
func NewGRPCServer(manager Manager) *GRPCServer {
	return &GRPCServer{
		manager: manager,
		logger:  log.NewNopLogger(),
	}
}

// SetLogger는 로거를 설정합니다.
func (s *GRPCServer) SetLogger(logger log.Logger) {
	s.logger = logger.With("module", "checkpoint_grpc")
	s.manager.SetLogger(logger)
}

// GRPCCheckpointRequest는 체크포인트 조회 요청입니다.
type GRPCCheckpointRequest struct {
	CheckpointNumber int64 `json:"checkpoint_number"`
}

// GRPCCheckpointResponse는 체크포인트 조회 응답입니다.
type GRPCCheckpointResponse struct {
	Checkpoint *GRPCCheckpoint `json:"checkpoint"`
}

// GRPCCheckpoint는 GRPC용 체크포인트 데이터 구조체입니다.
type GRPCCheckpoint struct {
	Number     int64  `json:"number"`
	Data       []byte `json:"data"`
	RootHash   []byte `json:"root_hash"`
	StartBlock int64  `json:"start_block"`
	EndBlock   int64  `json:"end_block"`
}

// GRPCCheckpointCountRequest는 체크포인트 개수 조회 요청입니다.
type GRPCCheckpointCountRequest struct{}

// GRPCCheckpointCountResponse는 체크포인트 개수 조회 응답입니다.
type GRPCCheckpointCountResponse struct {
	Count int64 `json:"count"`
}

// FetchCheckpoint는 체크포인트를 조회합니다.
func (s *GRPCServer) FetchCheckpoint(ctx context.Context, req *GRPCCheckpointRequest) (*GRPCCheckpointResponse, error) {
	s.logger.Info("GRPC FetchCheckpoint called", "number", req.CheckpointNumber)

	checkpoint, err := s.manager.GetCheckpoint(uint64(req.CheckpointNumber))
	if err != nil {
		if err == ErrCheckpointNotFound {
			return nil, status.Errorf(codes.NotFound, "checkpoint not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get checkpoint: %v", err)
	}

	grpcCheckpoint := &GRPCCheckpoint{
		Number:     int64(checkpoint.Number),
		Data:       checkpoint.Data.StateRootHash, // 상태 루트 해시
		RootHash:   checkpoint.Data.BlockHash,     // 블록 해시
		StartBlock: int64(checkpoint.Data.Height), // 시작 블록 (단일 블록 체크포인트)
		EndBlock:   int64(checkpoint.Data.Height), // 종료 블록 (단일 블록 체크포인트)
	}

	return &GRPCCheckpointResponse{
		Checkpoint: grpcCheckpoint,
	}, nil
}

// FetchCheckpointCount는 체크포인트 개수를 조회합니다.
func (s *GRPCServer) FetchCheckpointCount(ctx context.Context, req *GRPCCheckpointCountRequest) (*GRPCCheckpointCountResponse, error) {
	s.logger.Info("GRPC FetchCheckpointCount called")

	count, err := s.manager.GetCheckpointCount()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get checkpoint count: %v", err)
	}

	return &GRPCCheckpointCountResponse{
		Count: int64(count),
	}, nil
}

// CreateCheckpointRequest는 체크포인트 생성 요청입니다.
type CreateCheckpointRequest struct {
	Height int64 `json:"height"`
}

// CreateCheckpointResponse는 체크포인트 생성 응답입니다.
type CreateCheckpointResponse struct {
	Checkpoint *GRPCCheckpoint `json:"checkpoint"`
}

// CreateCheckpoint는 새 체크포인트를 생성합니다.
func (s *GRPCServer) CreateCheckpoint(ctx context.Context, req *CreateCheckpointRequest) (*CreateCheckpointResponse, error) {
	s.logger.Info("GRPC CreateCheckpoint called", "height", req.Height)

	checkpoint, err := s.manager.CreateCheckpoint(ctx, uint64(req.Height))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create checkpoint: %v", err)
	}

	grpcCheckpoint := &GRPCCheckpoint{
		Number:     int64(checkpoint.Number),
		Data:       checkpoint.Data.StateRootHash, // 상태 루트 해시
		RootHash:   checkpoint.Data.BlockHash,     // 블록 해시
		StartBlock: int64(checkpoint.Data.Height), // 시작 블록 (단일 블록 체크포인트)
		EndBlock:   int64(checkpoint.Data.Height), // 종료 블록 (단일 블록 체크포인트)
	}

	return &CreateCheckpointResponse{
		Checkpoint: grpcCheckpoint,
	}, nil
}
