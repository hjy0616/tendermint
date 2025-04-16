package span

import (
	"sync"

	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

// DefaultEventHandler는 기본 스팬 이벤트 처리기를 구현합니다.
type DefaultEventHandler struct {
	mtx               sync.RWMutex
	logger            log.Logger
	state             *state.State
	consensusState    *consensus.State
	peerSwitch        *p2p.Switch
	validatorUpdaters []ValidatorUpdater
}

// ValidatorUpdater는 검증자 업데이트 관련 작업을 처리하는 인터페이스입니다.
type ValidatorUpdater interface {
	UpdateValidators(validators types.ValidatorSet) error
}

// NewDefaultEventHandler는 새 기본 이벤트 처리기를 생성합니다.
func NewDefaultEventHandler(
	logger log.Logger,
	state *state.State,
	consensusState *consensus.State,
	peerSwitch *p2p.Switch,
) *DefaultEventHandler {
	return &DefaultEventHandler{
		logger:            logger.With("module", "span_event_handler"),
		state:             state,
		consensusState:    consensusState,
		peerSwitch:        peerSwitch,
		validatorUpdaters: []ValidatorUpdater{},
	}
}

// AddValidatorUpdater는 검증자 업데이트 처리기를 추가합니다.
func (h *DefaultEventHandler) AddValidatorUpdater(updater ValidatorUpdater) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.validatorUpdaters = append(h.validatorUpdaters, updater)
}

// HandleSpanEvent는 스팬 이벤트를 처리합니다.
func (h *DefaultEventHandler) HandleSpanEvent(event *SpanEvent) {
	switch event.Type {
	case SpanEventTypeStart:
		h.handleSpanStart(event)
	case SpanEventTypeEnd:
		h.handleSpanEnd(event)
	case SpanEventTypeValidatorUpdate:
		h.handleValidatorUpdate(event)
	default:
		h.logger.Error("Unknown span event type", "type", event.Type)
	}
}

// handleSpanStart는 스팬 시작 이벤트를 처리합니다.
func (h *DefaultEventHandler) handleSpanStart(event *SpanEvent) {
	h.logger.Info("Handling span start event",
		"spanID", event.Span.ID,
		"startBlock", event.Span.StartBlock,
		"endBlock", event.Span.EndBlock,
		"timestamp", event.Timestamp)

	// 검증자 세트 업데이트
	if err := h.updateValidators(event.Span.Validators); err != nil {
		h.logger.Error("Failed to update validators on span start", "error", err)
	}

	// 관련 시스템에 스팬 시작 알림
	h.notifySystemsOfSpanStart(event.Span)
}

// handleSpanEnd는 스팬 종료 이벤트를 처리합니다.
func (h *DefaultEventHandler) handleSpanEnd(event *SpanEvent) {
	h.logger.Info("Handling span end event",
		"spanID", event.Span.ID,
		"startBlock", event.Span.StartBlock,
		"endBlock", event.Span.EndBlock,
		"timestamp", event.Timestamp)

	// 스팬 완료 메트릭 업데이트
	updateSpanCompletionMetrics(event.Span)

	// 관련 시스템에 스팬 종료 알림
	h.notifySystemsOfSpanEnd(event.Span)
}

// handleValidatorUpdate는 검증자 업데이트 이벤트를 처리합니다.
func (h *DefaultEventHandler) handleValidatorUpdate(event *SpanEvent) {
	h.logger.Info("Handling validator update event",
		"spanID", event.Span.ID,
		"validatorCount", len(event.Span.Validators.Validators),
		"timestamp", event.Timestamp)

	// 검증자 세트 업데이트
	if err := h.updateValidators(event.Span.Validators); err != nil {
		h.logger.Error("Failed to update validators", "error", err)
	}
}

// updateValidators는 검증자 세트를 업데이트합니다.
func (h *DefaultEventHandler) updateValidators(validators types.ValidatorSet) error {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	// 내부 상태 업데이트
	// 주의: 실제 구현에서는 적절한 상태 업데이트 메커니즘 사용 필요
	if h.state != nil {
		// state 객체의 검증자 세트 업데이트 로직 구현 필요
		h.logger.Info("Updating state validators", "count", len(validators.Validators))
	}

	// 등록된 모든 업데이터에 통지
	for _, updater := range h.validatorUpdaters {
		if err := updater.UpdateValidators(validators); err != nil {
			h.logger.Error("Validator updater failed", "error", err)
			// 오류를 기록하되 계속 진행
		}
	}

	return nil
}

// notifySystemsOfSpanStart는 시스템에 스팬 시작을 알립니다.
func (h *DefaultEventHandler) notifySystemsOfSpanStart(span *Span) {
	// 합의 상태 알림
	if h.consensusState != nil {
		// 실제 구현에서는 합의 상태 모듈에 스팬 시작 알림 메커니즘 구현 필요
		h.logger.Info("Notifying consensus state of span start")
	}

	// P2P 네트워크 알림
	if h.peerSwitch != nil {
		// 실제 구현에서는 P2P 네트워크에 스팬 시작 알림 메커니즘 구현 필요
		h.logger.Info("Notifying P2P network of span start", "peerCount", h.peerSwitch.Peers().Size())
	}
}

// notifySystemsOfSpanEnd는 시스템에 스팬 종료를 알립니다.
func (h *DefaultEventHandler) notifySystemsOfSpanEnd(span *Span) {
	// 합의 상태 알림
	if h.consensusState != nil {
		// 실제 구현에서는 합의 상태 모듈에 스팬 종료 알림 메커니즘 구현 필요
		h.logger.Info("Notifying consensus state of span end")
	}

	// P2P 네트워크 알림
	if h.peerSwitch != nil {
		// 실제 구현에서는 P2P 네트워크에 스팬 종료 알림 메커니즘 구현 필요
		h.logger.Info("Notifying P2P network of span end", "peerCount", h.peerSwitch.Peers().Size())
	}
}

// updateSpanCompletionMetrics는 스팬 완료 메트릭을 업데이트합니다.
func updateSpanCompletionMetrics(span *Span) {
	// TODO: 스팬 완료 메트릭 업데이트 구현
	// 블록 생성 통계, 검증자 참여도 등 집계
}
