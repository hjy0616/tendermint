package coregrpc

import (
	"context"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// TendermintServiceClient는 TendermintService gRPC 서비스의 클라이언트입니다.
type TendermintServiceClient struct {
	conn      *grpc.ClientConn
	client    TendermintServiceClient_Client
	timeout   time.Duration
	addr      string
	logger    log.Logger
	reconnect bool
	mutex     sync.Mutex
}

// TendermintServiceClient_Client는 생성된 gRPC 클라이언트 인터페이스입니다.
type TendermintServiceClient_Client interface {
	FetchCheckpoint(ctx context.Context, in *RequestFetchCheckpoint, opts ...grpc.CallOption) (*ResponseFetchCheckpoint, error)
	FetchCheckpointCount(ctx context.Context, in *RequestFetchCheckpointCount, opts ...grpc.CallOption) (*ResponseFetchCheckpointCount, error)
	FetchMilestone(ctx context.Context, in *RequestFetchMilestone, opts ...grpc.CallOption) (*ResponseFetchMilestone, error)
	FetchMilestoneCount(ctx context.Context, in *RequestFetchMilestoneCount, opts ...grpc.CallOption) (*ResponseFetchMilestoneCount, error)
	FetchNoAckMilestone(ctx context.Context, in *RequestFetchNoAckMilestone, opts ...grpc.CallOption) (*ResponseFetchNoAckMilestone, error)
	StateSyncEvents(ctx context.Context, in *RequestStateSyncEvents, opts ...grpc.CallOption) (*ResponseStateSyncEvents, error)
	CommitStates(ctx context.Context, in *RequestCommitStates, opts ...grpc.CallOption) (*ResponseCommitStates, error)
	Span(ctx context.Context, in *RequestSpan, opts ...grpc.CallOption) (*ResponseSpan, error)
	BlockInfo(ctx context.Context, in *RequestBlockInfo, opts ...grpc.CallOption) (*ResponseBlockInfo, error)
	MempoolInfo(ctx context.Context, in *RequestMempoolInfo, opts ...grpc.CallOption) (*ResponseMempoolInfo, error)
}

// ClientOption은 클라이언트 설정을 위한 함수 타입입니다.
type ClientOption func(*TendermintServiceClient)

// WithLogger는 로거를 설정합니다.
func WithLogger(logger log.Logger) ClientOption {
	return func(c *TendermintServiceClient) {
		c.logger = logger
	}
}

// WithAutoReconnect는 자동 재연결 옵션을 설정합니다.
func WithAutoReconnect(reconnect bool) ClientOption {
	return func(c *TendermintServiceClient) {
		c.reconnect = reconnect
	}
}

// NewTendermintServiceClient는 새로운 TendermintServiceClient를 생성합니다.
func NewTendermintServiceClient(addr string, timeout time.Duration, opts ...ClientOption) (*TendermintServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := &TendermintServiceClient{
		conn:      conn,
		client:    NewTendermintServiceClient_Generated(conn),
		timeout:   timeout,
		addr:      addr,
		logger:    log.NewNopLogger(),
		reconnect: true,
	}

	// 옵션 적용
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// NewTendermintServiceClient_Generated는 gRPC가 생성한 클라이언트를 생성합니다.
func NewTendermintServiceClient_Generated(conn *grpc.ClientConn) TendermintServiceClient_Client {
	return NewTendermintServiceClient_Impl(conn)
}

// Close는 gRPC 연결을 닫습니다.
func (c *TendermintServiceClient) Close() error {
	return c.conn.Close()
}

// reconnectIfNeeded는 필요한 경우 gRPC 연결을 재연결합니다.
func (c *TendermintServiceClient) reconnectIfNeeded() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.reconnect {
		return nil
	}

	if c.conn.GetState() == connectivity.Ready {
		return nil
	}

	c.logger.Info("Reconnecting to gRPC server", "addr", c.addr)

	// 기존 연결 닫기
	if c.conn != nil {
		c.conn.Close()
	}

	// 새 연결 생성
	conn, err := grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.logger.Error("Failed to reconnect to gRPC server", "addr", c.addr, "error", err)
		return err
	}

	c.conn = conn
	c.client = NewTendermintServiceClient_Generated(conn)
	c.logger.Info("Successfully reconnected to gRPC server", "addr", c.addr)

	return nil
}

// FetchCheckpoint는 체크포인트 데이터를 가져옵니다.
func (c *TendermintServiceClient) FetchCheckpoint(checkpointNumber int64) (*ResponseFetchCheckpoint, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchCheckpoint{
		CheckpointNumber: checkpointNumber,
	}

	return c.client.FetchCheckpoint(ctx, req)
}

// FetchCheckpointCount는 체크포인트 수를 가져옵니다.
func (c *TendermintServiceClient) FetchCheckpointCount() (*ResponseFetchCheckpointCount, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchCheckpointCount{}

	return c.client.FetchCheckpointCount(ctx, req)
}

// FetchMilestone는 마일스톤 데이터를 가져옵니다.
func (c *TendermintServiceClient) FetchMilestone(milestoneID int64) (*ResponseFetchMilestone, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchMilestone{
		MilestoneId: milestoneID,
	}

	return c.client.FetchMilestone(ctx, req)
}

// FetchMilestoneCount는 마일스톤 수를 가져옵니다.
func (c *TendermintServiceClient) FetchMilestoneCount() (*ResponseFetchMilestoneCount, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchMilestoneCount{}

	return c.client.FetchMilestoneCount(ctx, req)
}

// FetchNoAckMilestone는 승인되지 않은 마일스톤 ID 목록을 가져옵니다.
func (c *TendermintServiceClient) FetchNoAckMilestone() (*ResponseFetchNoAckMilestone, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestFetchNoAckMilestone{}

	return c.client.FetchNoAckMilestone(ctx, req)
}

// StateSyncEvents는 특정 기간 동안의 상태 동기화 이벤트를 가져옵니다.
func (c *TendermintServiceClient) StateSyncEvents(fromTime, toTime int64, recordType string) (*ResponseStateSyncEvents, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

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
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestCommitStates{
		EventIds: eventIDs,
	}

	return c.client.CommitStates(ctx, req)
}

// Span은 지정된 ID의 스팬 데이터를 가져옵니다.
func (c *TendermintServiceClient) Span(spanID int64) (*ResponseSpan, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestSpan{
		SpanId: spanID,
	}

	return c.client.Span(ctx, req)
}

// BlockInfo는 지정된 높이의 블록 정보를 가져옵니다.
func (c *TendermintServiceClient) BlockInfo(height int64) (*ResponseBlockInfo, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestBlockInfo{
		Height: height,
	}

	return c.client.BlockInfo(ctx, req)
}

// MempoolInfo는 메모리풀의 현재 상태 정보를 가져옵니다.
func (c *TendermintServiceClient) MempoolInfo() (*ResponseMempoolInfo, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &RequestMempoolInfo{}

	return c.client.MempoolInfo(ctx, req)
}

// EventSubscription은 이벤트 구독을 관리합니다.
type EventSubscription struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// Close는 이벤트 구독을 종료합니다.
func (s *EventSubscription) Close() {
	s.cancel()
	<-s.done
}

// SubscribeStateSyncEvents는 상태 동기화 이벤트를 구독합니다.
func (c *TendermintServiceClient) SubscribeStateSyncEvents(
	recordType string,
	interval time.Duration,
	callback func(*ResponseStateSyncEvents, error),
) (*EventSubscription, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)

		lastTime := time.Now().Unix()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				nowTime := time.Now().Unix()
				events, err := c.StateSyncEvents(lastTime, nowTime, recordType)
				callback(events, err)
				if err == nil {
					lastTime = nowTime
				}
			}
		}
	}()

	return &EventSubscription{
		cancel: cancel,
		done:   done,
	}, nil
}

// SubscribeBlockInfo는 새 블록 정보를 구독합니다.
func (c *TendermintServiceClient) SubscribeBlockInfo(
	startHeight int64,
	interval time.Duration,
	callback func(*ResponseBlockInfo, error),
) (*EventSubscription, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)

		currentHeight := startHeight
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				info, err := c.BlockInfo(currentHeight)
				callback(info, err)
				if err == nil && info != nil && info.Block != nil {
					currentHeight++
				}
			}
		}
	}()

	return &EventSubscription{
		cancel: cancel,
		done:   done,
	}, nil
}
