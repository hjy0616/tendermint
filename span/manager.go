package span

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

// Manager는 스팬 관리 시스템의 핵심 구성 요소로 스팬의 생성, 관리, 전환 등을 담당합니다.
type Manager struct {
	mtx           sync.RWMutex
	currentSpan   *Span
	spans         map[uint64]*Span  // spanID -> Span
	blockSpans    map[uint64]uint64 // blockNumber -> spanID
	eventHandlers []EventHandler
	logger        log.Logger
}

// EventHandler는 스팬 이벤트 핸들러 인터페이스입니다.
type EventHandler interface {
	HandleSpanEvent(event *SpanEvent)
}

// EventHandlerFunc는 EventHandler 인터페이스를 함수로 구현합니다.
type EventHandlerFunc func(event *SpanEvent)

// HandleSpanEvent는 이벤트 핸들러 함수를 실행합니다.
func (f EventHandlerFunc) HandleSpanEvent(event *SpanEvent) {
	f(event)
}

// NewManager는 새 스팬 관리자를 생성합니다.
func NewManager(logger log.Logger) *Manager {
	return &Manager{
		spans:         make(map[uint64]*Span),
		blockSpans:    make(map[uint64]uint64),
		eventHandlers: []EventHandler{},
		logger:        logger,
	}
}

// GetCurrentSpan은 현재 활성화된 스팬을 반환합니다.
func (m *Manager) GetCurrentSpan() *Span {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return m.currentSpan
}

// GetSpanByID는 ID로 스팬을 조회합니다.
func (m *Manager) GetSpanByID(spanID uint64) (*Span, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	span, ok := m.spans[spanID]
	if !ok {
		return nil, fmt.Errorf("span with ID %d not found", spanID)
	}
	return span, nil
}

// GetSpanByBlockNumber는 블록 번호에 해당하는 스팬을 조회합니다.
func (m *Manager) GetSpanByBlockNumber(blockNumber uint64) (*Span, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	spanID, ok := m.blockSpans[blockNumber]
	if !ok {
		// 정확한 매핑이 없는 경우 블록 번호보다 작거나 같은 최대 스팬을 찾습니다.
		for bn, id := range m.blockSpans {
			if bn <= blockNumber {
				span := m.spans[id]
				if span.IsActive(blockNumber) {
					return span, nil
				}
			}
		}
		return nil, fmt.Errorf("no span found for block number %d", blockNumber)
	}

	return m.spans[spanID], nil
}

// CreateSpan은 새 스팬을 생성합니다.
func (m *Manager) CreateSpan(id, startBlock, endBlock uint64, validators types.ValidatorSet) (*Span, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// 이미 존재하는 스팬인지 확인
	if _, exists := m.spans[id]; exists {
		return nil, fmt.Errorf("span with ID %d already exists", id)
	}

	// 블록 번호 범위 충돌 검사
	for _, span := range m.spans {
		// 기존 스팬과 새 스팬의 블록 범위가 겹치는지 확인
		if startBlock <= span.EndBlock && endBlock >= span.StartBlock {
			return nil, fmt.Errorf("span block range conflict with existing span %d", span.ID)
		}
	}

	span := NewSpan(id, startBlock, endBlock, validators, time.Now())
	m.spans[id] = span

	// 블록 번호 매핑 추가
	for bn := startBlock; bn <= endBlock; bn++ {
		m.blockSpans[bn] = id
	}

	m.logger.Info("Created new span", "id", id, "startBlock", startBlock, "endBlock", endBlock)

	return span, nil
}

// ActivateSpan은 스팬을 활성화 상태로 전환합니다.
func (m *Manager) ActivateSpan(spanID uint64) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	span, ok := m.spans[spanID]
	if !ok {
		return fmt.Errorf("span with ID %d not found", spanID)
	}

	// 이전 활성 스팬이 있는 경우 완료 이벤트 발생
	if m.currentSpan != nil {
		m.currentSpan.CompleteSpan(time.Now())
		m.emitSpanEvent(SpanEventTypeEnd, m.currentSpan)
	}

	m.currentSpan = span
	m.emitSpanEvent(SpanEventTypeStart, span)

	m.logger.Info("Activated span", "id", spanID)

	return nil
}

// AddEventHandler는 이벤트 핸들러를 등록합니다.
func (m *Manager) AddEventHandler(handler EventHandler) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.eventHandlers = append(m.eventHandlers, handler)
}

// emitSpanEvent는 스팬 이벤트를 발생시키고 등록된 핸들러에 전달합니다.
func (m *Manager) emitSpanEvent(eventType SpanEventType, span *Span) {
	event := NewSpanEvent(eventType, span)

	// 비동기적으로 모든 핸들러에게 이벤트 전달
	for _, handler := range m.eventHandlers {
		go handler.HandleSpanEvent(event)
	}

	m.logger.Info("Emitted span event",
		"type", eventType.String(),
		"spanID", span.ID,
		"timestamp", event.Timestamp)
}

// UpdateValidators는 현재 스팬의 검증자 세트를 업데이트합니다.
func (m *Manager) UpdateValidators(validators types.ValidatorSet) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.currentSpan == nil {
		return fmt.Errorf("no active span")
	}

	// 검증자 세트 업데이트
	m.currentSpan.Validators = validators

	// 검증자 업데이트 이벤트 발생
	m.emitSpanEvent(SpanEventTypeValidatorUpdate, m.currentSpan)

	m.logger.Info("Updated validators for current span", "spanID", m.currentSpan.ID)

	return nil
}

// Export는 현재 스팬 상태를 JSON으로 내보냅니다.
func (m *Manager) Export() ([]byte, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	state := struct {
		CurrentSpan *Span            `json:"current_span"`
		Spans       map[uint64]*Span `json:"spans"`
	}{
		CurrentSpan: m.currentSpan,
		Spans:       m.spans,
	}

	return json.Marshal(state)
}

// Import는 JSON에서 스팬 상태를 가져옵니다.
func (m *Manager) Import(data []byte) error {
	state := struct {
		CurrentSpan *Span            `json:"current_span"`
		Spans       map[uint64]*Span `json:"spans"`
	}{}

	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.currentSpan = state.CurrentSpan
	m.spans = state.Spans

	// 블록 매핑 재구성
	m.blockSpans = make(map[uint64]uint64)
	for _, span := range m.spans {
		for bn := span.StartBlock; bn <= span.EndBlock; bn++ {
			m.blockSpans[bn] = span.ID
		}
	}

	return nil
}
