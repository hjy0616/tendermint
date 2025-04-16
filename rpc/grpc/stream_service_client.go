package coregrpc

// 이 파일은 프로토버프에서 생성된 gRPC 코드를 임시로 대체합니다.
// 실제 구현에서는 다음의 단계가 필요합니다:
//
// 1. tendermint/proto/tendermint/rpc/grpc/stream.proto 파일을 정의합니다.
// 2. protoc를 사용하여 Go 코드를 생성합니다:
//    protoc --go_out=. --go-grpc_out=. tendermint/proto/tendermint/rpc/grpc/stream.proto
// 3. 이 파일의 임시 타입 선언 부분은 모두 제거하고 생성된 코드를 임포트합니다.

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

//------------------------------------------------------------------------------
// 임시 gRPC 클라이언트 인터페이스 및 타입 선언
// 모든 타입 선언은 proto 파일을 컴파일하여 자동 생성된 코드로 대체되어야 합니다.
//------------------------------------------------------------------------------

// TendermintStreamServiceClient는 스트림 서비스의 클라이언트 인터페이스입니다.
type TendermintStreamServiceClient interface {
	SubscribeBlockEvents(ctx context.Context, in *RequestSubscribeBlockEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeBlockEventsClient, error)
	SubscribeStateEvents(ctx context.Context, in *RequestSubscribeStateEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeStateEventsClient, error)
	SubscribeValidatorEvents(ctx context.Context, in *RequestSubscribeValidatorEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeValidatorEventsClient, error)
	SubscribeMempoolEvents(ctx context.Context, in *RequestSubscribeMempoolEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeMempoolEventsClient, error)
}

// 스트림 인터페이스 정의
type TendermintStreamService_SubscribeBlockEventsClient interface {
	Recv() (*ResponseBlockEvent, error)
	grpc.ClientStream
}

type TendermintStreamService_SubscribeStateEventsClient interface {
	Recv() (*ResponseStateEvent, error)
	grpc.ClientStream
}

type TendermintStreamService_SubscribeValidatorEventsClient interface {
	Recv() (*ResponseValidatorEvent, error)
	grpc.ClientStream
}

type TendermintStreamService_SubscribeMempoolEventsClient interface {
	Recv() (*ResponseMempoolEvent, error)
	grpc.ClientStream
}

// 요청 타입 정의
type RequestSubscribeBlockEvents struct {
	StartHeight int64
}

type RequestSubscribeStateEvents struct {
	RecordType string
	FromTime   int64
}

type RequestSubscribeValidatorEvents struct {
	StartHeight int64
}

type RequestSubscribeMempoolEvents struct{}

// 응답 타입 정의
type ResponseBlockEvent struct {
	BlockInfo interface{} // 실제로는 *ResponseBlockInfo 타입
	Timestamp int64
}

type ResponseStateEvent struct {
	Records   []interface{} // 실제로는 []*EventRecord 타입
	Timestamp int64
}

type ValidatorEvent struct {
	Height            int64
	AddedValidators   []*types.Validator
	RemovedValidators []*types.Validator
	UpdatedValidators []*types.Validator
}

type ResponseValidatorEvent struct {
	Event     *ValidatorEvent
	Timestamp int64
}

type MempoolEvent_EventType int32

const (
	MempoolEvent_ADDED   MempoolEvent_EventType = 0
	MempoolEvent_REMOVED MempoolEvent_EventType = 1
	MempoolEvent_RECHECK MempoolEvent_EventType = 2
)

type MempoolEvent struct {
	Type      MempoolEvent_EventType
	TxHash    []byte
	TxData    []byte
	Timestamp int64
}

type ResponseMempoolEvent struct {
	Event     *MempoolEvent
	Timestamp int64
}

// 임시 클라이언트 구현 - 실제로는 자동 생성된 코드로 대체됩니다
func NewTendermintStreamServiceClient(cc *grpc.ClientConn) TendermintStreamServiceClient {
	return &dummyTendermintStreamServiceClient{cc}
}

type dummyTendermintStreamServiceClient struct {
	cc *grpc.ClientConn
}

func (c *dummyTendermintStreamServiceClient) SubscribeBlockEvents(ctx context.Context, in *RequestSubscribeBlockEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeBlockEventsClient, error) {
	return nil, io.EOF // 임시 응답
}

func (c *dummyTendermintStreamServiceClient) SubscribeStateEvents(ctx context.Context, in *RequestSubscribeStateEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeStateEventsClient, error) {
	return nil, io.EOF // 임시 응답
}

func (c *dummyTendermintStreamServiceClient) SubscribeValidatorEvents(ctx context.Context, in *RequestSubscribeValidatorEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeValidatorEventsClient, error) {
	return nil, io.EOF // 임시 응답
}

func (c *dummyTendermintStreamServiceClient) SubscribeMempoolEvents(ctx context.Context, in *RequestSubscribeMempoolEvents, opts ...grpc.CallOption) (TendermintStreamService_SubscribeMempoolEventsClient, error) {
	return nil, io.EOF // 임시 응답
}

//------------------------------------------------------------------------------
// 실제 스트림 클라이언트 구현
//------------------------------------------------------------------------------

// TendermintStreamClient는 TendermintStreamService gRPC 서비스의 클라이언트 래퍼입니다.
type TendermintStreamClient struct {
	conn      *grpc.ClientConn
	client    TendermintStreamServiceClient
	logger    log.Logger
	reconnect bool
	addr      string
	mutex     sync.Mutex
}

// NewTendermintStreamClient는 새로운 TendermintStreamClient를 생성합니다.
func NewTendermintStreamClient(addr string, logger log.Logger, reconnect bool) (*TendermintStreamClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &TendermintStreamClient{
		conn:      conn,
		client:    NewTendermintStreamServiceClient(conn),
		logger:    logger,
		reconnect: reconnect,
		addr:      addr,
	}, nil
}

// Close는 gRPC 연결을 닫습니다.
func (c *TendermintStreamClient) Close() error {
	return c.conn.Close()
}

// reconnectIfNeeded는 필요한 경우 gRPC 연결을 재연결합니다.
func (c *TendermintStreamClient) reconnectIfNeeded() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.reconnect {
		return nil
	}

	if c.conn.GetState() == connectivity.Ready {
		return nil
	}

	c.logger.Info("Reconnecting to gRPC stream server", "addr", c.addr)

	// 기존 연결 닫기
	if c.conn != nil {
		c.conn.Close()
	}

	// 새 연결 생성
	conn, err := grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.logger.Error("Failed to reconnect to gRPC stream server", "addr", c.addr, "error", err)
		return err
	}

	c.conn = conn
	c.client = NewTendermintStreamServiceClient(conn)
	c.logger.Info("Successfully reconnected to gRPC stream server", "addr", c.addr)

	return nil
}

// BlockEventHandler는 블록 이벤트 핸들러 함수 타입입니다.
type BlockEventHandler func(*ResponseBlockEvent) error

// StateEventHandler는 상태 이벤트 핸들러 함수 타입입니다.
type StateEventHandler func(*ResponseStateEvent) error

// ValidatorEventHandler는 검증자 이벤트 핸들러 함수 타입입니다.
type ValidatorEventHandler func(*ResponseValidatorEvent) error

// MempoolEventHandler는 메모리풀 이벤트 핸들러 함수 타입입니다.
type MempoolEventHandler func(*ResponseMempoolEvent) error

// StreamSubscription은 gRPC 스트림 구독을 관리합니다.
type StreamSubscription struct {
	cancel  context.CancelFunc
	done    chan struct{}
	errChan chan error
}

// Err은 구독 과정에서 발생한 오류를 반환합니다.
func (s *StreamSubscription) Err() <-chan error {
	return s.errChan
}

// Done은 구독이 완료되면 닫히는 채널을 반환합니다.
func (s *StreamSubscription) Done() <-chan struct{} {
	return s.done
}

// Close는 스트림 구독을 종료합니다.
func (s *StreamSubscription) Close() {
	s.cancel()
	<-s.done
}

// SubscribeBlockEvents는 블록 이벤트를 구독합니다.
func (c *TendermintStreamClient) SubscribeBlockEvents(
	startHeight int64,
	handler BlockEventHandler,
) (*StreamSubscription, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	errChan := make(chan error, 1)

	go func() {
		defer close(done)
		defer close(errChan)

		req := &RequestSubscribeBlockEvents{
			StartHeight: startHeight,
		}

		for {
			if ctx.Err() != nil {
				return
			}

			stream, err := c.client.SubscribeBlockEvents(ctx, req)
			if err != nil {
				c.logger.Error("Failed to subscribe to block events", "err", err)
				errChan <- err

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 5):
					// 재연결 시도
					if err := c.reconnectIfNeeded(); err != nil {
						continue
					}
				}
				continue
			}

			for {
				event, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					if ctx.Err() != nil {
						return
					}

					c.logger.Error("Error receiving block event", "err", err)
					errChan <- err
					break
				}

				if err := handler(event); err != nil {
					c.logger.Error("Error handling block event", "err", err)
					errChan <- err
				}
			}
		}
	}()

	return &StreamSubscription{
		cancel:  cancel,
		done:    done,
		errChan: errChan,
	}, nil
}

// SubscribeStateEvents는 상태 이벤트를 구독합니다.
func (c *TendermintStreamClient) SubscribeStateEvents(
	recordType string,
	fromTime int64,
	handler StateEventHandler,
) (*StreamSubscription, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	errChan := make(chan error, 1)

	go func() {
		defer close(done)
		defer close(errChan)

		req := &RequestSubscribeStateEvents{
			RecordType: recordType,
			FromTime:   fromTime,
		}

		for {
			if ctx.Err() != nil {
				return
			}

			stream, err := c.client.SubscribeStateEvents(ctx, req)
			if err != nil {
				c.logger.Error("Failed to subscribe to state events", "err", err)
				errChan <- err

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 5):
					// 재연결 시도
					if err := c.reconnectIfNeeded(); err != nil {
						continue
					}
				}
				continue
			}

			for {
				event, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					if ctx.Err() != nil {
						return
					}

					c.logger.Error("Error receiving state event", "err", err)
					errChan <- err
					break
				}

				if err := handler(event); err != nil {
					c.logger.Error("Error handling state event", "err", err)
					errChan <- err
				}

				// FromTime 업데이트
				if event.Timestamp > req.FromTime {
					req.FromTime = event.Timestamp
				}
			}
		}
	}()

	return &StreamSubscription{
		cancel:  cancel,
		done:    done,
		errChan: errChan,
	}, nil
}

// SubscribeValidatorEvents는 검증자 이벤트를 구독합니다.
func (c *TendermintStreamClient) SubscribeValidatorEvents(
	startHeight int64,
	handler ValidatorEventHandler,
) (*StreamSubscription, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	errChan := make(chan error, 1)

	go func() {
		defer close(done)
		defer close(errChan)

		req := &RequestSubscribeValidatorEvents{
			StartHeight: startHeight,
		}

		for {
			if ctx.Err() != nil {
				return
			}

			stream, err := c.client.SubscribeValidatorEvents(ctx, req)
			if err != nil {
				c.logger.Error("Failed to subscribe to validator events", "err", err)
				errChan <- err

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 5):
					// 재연결 시도
					if err := c.reconnectIfNeeded(); err != nil {
						continue
					}
				}
				continue
			}

			for {
				event, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					if ctx.Err() != nil {
						return
					}

					c.logger.Error("Error receiving validator event", "err", err)
					errChan <- err
					break
				}

				if err := handler(event); err != nil {
					c.logger.Error("Error handling validator event", "err", err)
					errChan <- err
				}

				// StartHeight 업데이트
				if event.Event != nil && event.Event.Height > req.StartHeight {
					req.StartHeight = event.Event.Height + 1
				}
			}
		}
	}()

	return &StreamSubscription{
		cancel:  cancel,
		done:    done,
		errChan: errChan,
	}, nil
}

// SubscribeMempoolEvents는 메모리풀 이벤트를 구독합니다.
func (c *TendermintStreamClient) SubscribeMempoolEvents(
	handler MempoolEventHandler,
) (*StreamSubscription, error) {
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	errChan := make(chan error, 1)

	go func() {
		defer close(done)
		defer close(errChan)

		req := &RequestSubscribeMempoolEvents{}

		for {
			if ctx.Err() != nil {
				return
			}

			stream, err := c.client.SubscribeMempoolEvents(ctx, req)
			if err != nil {
				c.logger.Error("Failed to subscribe to mempool events", "err", err)
				errChan <- err

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 5):
					// 재연결 시도
					if err := c.reconnectIfNeeded(); err != nil {
						continue
					}
				}
				continue
			}

			for {
				event, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					if ctx.Err() != nil {
						return
					}

					c.logger.Error("Error receiving mempool event", "err", err)
					errChan <- err
					break
				}

				if err := handler(event); err != nil {
					c.logger.Error("Error handling mempool event", "err", err)
					errChan <- err
				}
			}
		}
	}()

	return &StreamSubscription{
		cancel:  cancel,
		done:    done,
		errChan: errChan,
	}, nil
}
