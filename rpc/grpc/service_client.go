package coregrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TendermintServiceClient는 TendermintService gRPC 서비스의 클라이언트입니다.
type TendermintServiceClient struct {
	conn    *grpc.ClientConn
	client  TendermintServiceClient
	timeout time.Duration
}

// NewTendermintServiceClient는 새로운 TendermintServiceClient를 생성합니다.
func NewTendermintServiceClient(addr string, timeout time.Duration) (*TendermintServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &TendermintServiceClient{
		conn:    conn,
		client:  NewTendermintServiceClient(conn),
		timeout: timeout,
	}, nil
}

// Close는 gRPC 연결을 닫습니다.
func (c *TendermintServiceClient) Close() error {
	return c.conn.Close()
}

// FetchCheckpoint는 체크포인트 데이터를 가져옵니다.
func (c *TendermintServiceClient) FetchCheckpoint(checkpointNumber int64) (*ResponseFetchCheckpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchCheckpoint{
		CheckpointNumber: checkpointNumber,
	}

	return c.client.FetchCheckpoint(ctx, req)
}

// FetchCheckpointCount는 체크포인트 수를 가져옵니다.
func (c *TendermintServiceClient) FetchCheckpointCount() (*ResponseFetchCheckpointCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchCheckpointCount{}

	return c.client.FetchCheckpointCount(ctx, req)
}

// FetchMilestone는 마일스톤 데이터를 가져옵니다.
func (c *TendermintServiceClient) FetchMilestone(milestoneID int64) (*ResponseFetchMilestone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchMilestone{
		MilestoneId: milestoneID,
	}

	return c.client.FetchMilestone(ctx, req)
}

// FetchMilestoneCount는 마일스톤 수를 가져옵니다.
func (c *TendermintServiceClient) FetchMilestoneCount() (*ResponseFetchMilestoneCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchMilestoneCount{}

	return c.client.FetchMilestoneCount(ctx, req)
}

// FetchNoAckMilestone는 승인되지 않은 마일스톤 ID 목록을 가져옵니다.
func (c *TendermintServiceClient) FetchNoAckMilestone() (*ResponseFetchNoAckMilestone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchNoAckMilestone{}

	return c.client.FetchNoAckMilestone(ctx, req)
}

// StateSyncEvents는 특정 기간 동안의 상태 동기화 이벤트를 가져옵니다.
func (c *TendermintServiceClient) StateSyncEvents(fromTime, toTime int64, recordType string) (*ResponseStateSyncEvents, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestStateSyncEvents{
		FromTime:   fromTime,
		ToTime:     toTime,
		RecordType: recordType,
	}

	return c.client.StateSyncEvents(ctx, req)
}

// CommitStates는 이벤트 ID 목록을 커밋합니다.
func (c *TendermintServiceClient) CommitStates(eventIDs []int64) (*ResponseCommitStates, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestCommitStates{
		EventIds: eventIDs,
	}

	return c.client.CommitStates(ctx, req)
}

// Span은 지정된 ID의 스팬 데이터를 가져옵니다.
func (c *TendermintServiceClient) Span(spanID int64) (*ResponseSpan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestSpan{
		SpanId: spanID,
	}

	return c.client.Span(ctx, req)
}

// BlockInfo는 지정된 높이의 블록 정보를 가져옵니다.
func (c *TendermintServiceClient) BlockInfo(height int64) (*ResponseBlockInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestBlockInfo{
		Height: height,
	}

	return c.client.BlockInfo(ctx, req)
}

// MempoolInfo는 메모리풀의 현재 상태 정보를 가져옵니다.
func (c *TendermintServiceClient) MempoolInfo() (*ResponseMempoolInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestMempoolInfo{}

	return c.client.MempoolInfo(ctx, req)
}
