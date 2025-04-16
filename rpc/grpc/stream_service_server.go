package coregrpc

import (
	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/clerk"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/span"
)

// TendermintStreamServer는 TendermintStreamService gRPC 서버 구현체입니다.
type TendermintStreamServer struct {
	logger        log.Logger
	checkpointMgr checkpoint.Manager
	clerkMgr      clerk.Manager
	spanMgr       span.Manager
	consensus     consensus.Consensus
	mempool       mempool.Mempool
}

// NewTendermintStreamServer는 새로운 TendermintStreamServer를 생성합니다.
func NewTendermintStreamServer(
	logger log.Logger,
	checkpointMgr checkpoint.Manager,
	clerkMgr clerk.Manager,
	spanMgr span.Manager,
	consensus consensus.Consensus,
	mempool mempool.Mempool,
) *TendermintStreamServer {
	return &TendermintStreamServer{
		logger:        logger,
		checkpointMgr: checkpointMgr,
		clerkMgr:      clerkMgr,
		spanMgr:       spanMgr,
		consensus:     consensus,
		mempool:       mempool,
	}
}

// 참고: 이 파일은 프로토버프에서 자동 생성된 코드로 대체되어야 합니다.
// 여기서는 기본 구조만 제공하며, 실제 구현에서는 다음 단계를 따라야 합니다:
//
// 1. proto 파일에서 서비스 및 메시지 정의:
//    - TendermintStreamService 서비스
//    - 관련 요청/응답 메시지 타입
//
// 2. protoc 컴파일러로 Go 코드 생성:
//    - 서비스 인터페이스 및 메시지 타입 생성
//
// 3. 생성된 인터페이스 구현:
//    - SubscribeBlockEvents
//    - SubscribeStateEvents
//    - SubscribeValidatorEvents
//    - SubscribeMempoolEvents
//
// 4. 스트리밍 메서드에서는 각 이벤트를 처리하고 클라이언트에게 전송하는 로직 구현
