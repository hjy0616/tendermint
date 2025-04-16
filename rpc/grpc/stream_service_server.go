package coregrpc

import (
	"time"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/clerk"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/span"
	"github.com/tendermint/tendermint/types"
)

// TendermintStreamServer는 TendermintStreamService gRPC 서버 구현체입니다.
type TendermintStreamServer struct {
	UnimplementedTendermintStreamServiceServer
	logger        log.Logger
	checkpointMgr checkpoint.Manager
	clerkMgr      clerk.Manager
	spanMgr       span.Manager
	consensus     *consensus.State
	mempool       mempool.Mempool
}

// NewTendermintStreamServer는 새로운 TendermintStreamServer를 생성합니다.
func NewTendermintStreamServer(
	logger log.Logger,
	checkpointMgr checkpoint.Manager,
	clerkMgr clerk.Manager,
	spanMgr span.Manager,
	consensus *consensus.State,
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

// SubscribeBlockEvents는 블록 이벤트를 구독합니다.
func (s *TendermintStreamServer) SubscribeBlockEvents(req *RequestSubscribeBlockEvents, stream TendermintStreamService_SubscribeBlockEventsServer) error {
	s.logger.Info("SubscribeBlockEvents called", "start_height", req.StartHeight)

	currentHeight := req.StartHeight
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("SubscribeBlockEvents stream closed")
			return nil
		case <-ticker.C:
			// 현재 블록 정보 가져오기
			block, err := s.consensus.GetBlockByHeight(currentHeight)
			if err != nil {
				// 아직 해당 높이의 블록이 없으면 계속 대기
				continue
			}

			// 블록 정보 변환
			blockInfo := &ResponseBlockInfo{
				Block:           block,
				BlockId:         block.Hash().String(),
				AppHash:         block.AppHash.String(),
				LastCommitRound: int64(block.LastCommit.Round),
			}

			// 블록 이벤트 전송
			event := &ResponseBlockEvent{
				BlockInfo: blockInfo,
				Timestamp: time.Now().Unix(),
			}

			if err := stream.Send(event); err != nil {
				s.logger.Error("Failed to send block event", "height", currentHeight, "err", err)
				return err
			}

			// 다음 블록으로 이동
			currentHeight++
		}
	}
}

// SubscribeStateEvents는 상태 이벤트를 구독합니다.
func (s *TendermintStreamServer) SubscribeStateEvents(req *RequestSubscribeStateEvents, stream TendermintStreamService_SubscribeStateEventsServer) error {
	s.logger.Info("SubscribeStateEvents called", "record_type", req.RecordType, "from_time", req.FromTime)

	lastTime := req.FromTime
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("SubscribeStateEvents stream closed")
			return nil
		case <-ticker.C:
			nowTime := time.Now().Unix()

			// 상태 이벤트 조회
			events, err := s.clerkMgr.GetEventRecords(lastTime, nowTime, req.RecordType)
			if err != nil {
				s.logger.Error("Failed to get state events", "err", err)
				continue
			}

			// 이벤트가 없으면 다음 틱까지 대기
			if len(events) == 0 {
				continue
			}

			// 이벤트 레코드 변환
			var records []*EventRecord
			for _, event := range events {
				records = append(records, &EventRecord{
					Id:         uint64(event.ID),
					Time:       event.Time,
					Data:       event.Data,
					RecordType: event.RecordType,
				})
			}

			// 상태 이벤트 전송
			event := &ResponseStateEvent{
				Records:   records,
				Timestamp: nowTime,
			}

			if err := stream.Send(event); err != nil {
				s.logger.Error("Failed to send state event", "err", err)
				return err
			}

			// 다음 타임스탬프 업데이트
			lastTime = nowTime
		}
	}
}

// SubscribeValidatorEvents는 검증자 이벤트를 구독합니다.
func (s *TendermintStreamServer) SubscribeValidatorEvents(req *RequestSubscribeValidatorEvents, stream TendermintStreamService_SubscribeValidatorEventsServer) error {
	s.logger.Info("SubscribeValidatorEvents called", "start_height", req.StartHeight)

	currentHeight := req.StartHeight
	previousValidators := make(map[string]types.Validator)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("SubscribeValidatorEvents stream closed")
			return nil
		case <-ticker.C:
			// 현재 검증자 세트 가져오기
			vals, err := s.consensus.GetValidatorSet(currentHeight)
			if err != nil {
				// 아직 해당 높이의 검증자 세트가 없으면 계속 대기
				continue
			}

			currentValidators := make(map[string]types.Validator)
			for _, val := range vals.Validators {
				addrStr := val.Address.String()
				currentValidators[addrStr] = *val
			}

			// 검증자 변경 사항 계산
			var added, removed, updated []types.Validator

			// 추가된 검증자
			for addrStr, val := range currentValidators {
				if _, exists := previousValidators[addrStr]; !exists {
					added = append(added, val)
				}
			}

			// 제거된 검증자
			for addrStr, val := range previousValidators {
				if _, exists := currentValidators[addrStr]; !exists {
					removed = append(removed, val)
				}
			}

			// 업데이트된 검증자 (투표력 변경 등)
			for addrStr, currentVal := range currentValidators {
				if previousVal, exists := previousValidators[addrStr]; exists {
					if !currentVal.VotingPower.Equal(previousVal.VotingPower) {
						updated = append(updated, currentVal)
					}
				}
			}

			// 변경 사항이 있는 경우에만 이벤트 전송
			if len(added) > 0 || len(removed) > 0 || len(updated) > 0 {
				// Validator 객체를 proto 형식으로 변환
				var addedValidators, removedValidators, updatedValidators []*types.Validator

				for _, val := range added {
					valCopy := val
					addedValidators = append(addedValidators, &valCopy)
				}

				for _, val := range removed {
					valCopy := val
					removedValidators = append(removedValidators, &valCopy)
				}

				for _, val := range updated {
					valCopy := val
					updatedValidators = append(updatedValidators, &valCopy)
				}

				// 검증자 이벤트 생성
				event := &ResponseValidatorEvent{
					Event: &ValidatorEvent{
						Height:            currentHeight,
						AddedValidators:   addedValidators,
						RemovedValidators: removedValidators,
						UpdatedValidators: updatedValidators,
					},
					Timestamp: time.Now().Unix(),
				}

				if err := stream.Send(event); err != nil {
					s.logger.Error("Failed to send validator event", "height", currentHeight, "err", err)
					return err
				}
			}

			// 다음 높이로 이동 및 현재 검증자 세트 저장
			previousValidators = currentValidators
			currentHeight++
		}
	}
}

// SubscribeMempoolEvents는 메모리풀 이벤트를 구독합니다.
func (s *TendermintStreamServer) SubscribeMempoolEvents(req *RequestSubscribeMempoolEvents, stream TendermintStreamService_SubscribeMempoolEventsServer) error {
	s.logger.Info("SubscribeMempoolEvents called")

	// 메모리풀 이벤트 채널 구독
	eventCh := make(chan mempool.MempoolTxEvent, 100)
	s.mempool.SubscribeForTxEvents(eventCh)
	defer s.mempool.UnsubscribeForTxEvents(eventCh)

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("SubscribeMempoolEvents stream closed")
			return nil
		case mempoolEvent := <-eventCh:
			// 이벤트 타입 변환
			var eventType MempoolEvent_EventType
			switch mempoolEvent.Type {
			case mempool.TxAdded:
				eventType = MempoolEvent_ADDED
			case mempool.TxRemoved:
				eventType = MempoolEvent_REMOVED
			case mempool.TxRecheck:
				eventType = MempoolEvent_RECHECK
			default:
				continue // 알 수 없는 이벤트 타입은 무시
			}

			// 메모리풀 이벤트 생성
			event := &ResponseMempoolEvent{
				Event: &MempoolEvent{
					Type:      eventType,
					TxHash:    mempoolEvent.Tx.Hash(),
					TxData:    mempoolEvent.Tx,
					Timestamp: time.Now().Unix(),
				},
				Timestamp: time.Now().Unix(),
			}

			if err := stream.Send(event); err != nil {
				s.logger.Error("Failed to send mempool event", "err", err)
				return err
			}
		}
	}
}
