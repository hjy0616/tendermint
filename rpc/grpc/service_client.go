package coregrpc

// 이 파일은 프로토버프에서 생성된 gRPC 클라이언트 코드를 임시로 대체합니다.
// 실제 구현에서는 다음의 단계가 필요합니다:
//
// 1. tendermint/proto/tendermint/rpc/grpc/service.proto 파일을 정의합니다.
// 2. protoc를 사용하여 Go 코드를 생성합니다:
//    protoc --go_out=. --go-grpc_out=. tendermint/proto/tendermint/rpc/grpc/service.proto
// 3. 이 파일의 임시 타입 선언 부분은 모두 제거하고 생성된 코드를 임포트합니다.

import (
	"context"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// TendermintServiceClient는 gRPC 서비스 클라이언트입니다.
type TendermintServiceClient struct {
	conn      *grpc.ClientConn
	client    interface{} // 실제로는 자동 생성된 클라이언트 타입이 사용됩니다
	timeout   time.Duration
	addr      string
	logger    log.Logger
	reconnect bool
	mutex     sync.Mutex
}

// ClientOption은 클라이언트 옵션 함수 타입입니다.
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
		client:    nil, // 실제로는 여기서 자동 생성된 클라이언트를 생성합니다
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
	// c.client = NewTendermintServiceClient_Generated(conn) // 실제로는 이 부분을 구현해야 합니다
	c.logger.Info("Successfully reconnected to gRPC server", "addr", c.addr)

	return nil
}

// 참고: 이 파일에는 실제 구현 코드가 필요합니다.
// 프로토버프를 사용하여 다음과 같은 메서드를 자동 생성하고 구현해야 합니다:
//
// - FetchCheckpoint
// - FetchCheckpointCount
// - FetchMilestone
// - FetchMilestoneCount
// - FetchNoAckMilestone
// - StateSyncEvents
// - CommitStates
// - Span
// - BlockInfo
// - MempoolInfo
//
// 이 메서드들은 각각 자동 생성된 요청/응답 타입을 사용해야 합니다.
// 현재는 컴파일 오류 방지를 위해 최소한의 구조만 제공합니다.

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
