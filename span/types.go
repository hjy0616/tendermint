package span

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/types"
)

// Span 구조체는 일정 기간의 블록 범위와 검증자 세트 정보를 나타냅니다.
type Span struct {
	ID         uint64             `json:"span_id" yaml:"span_id"`
	StartBlock uint64             `json:"start_block" yaml:"start_block"`
	EndBlock   uint64             `json:"end_block" yaml:"end_block"`
	Validators types.ValidatorSet `json:"validators" yaml:"validators"`
	StartTime  time.Time          `json:"start_time" yaml:"start_time"`
	EndTime    time.Time          `json:"end_time,omitempty" yaml:"end_time,omitempty"`
}

// SpanStatus는 스팬의 현재 상태를 나타냅니다.
type SpanStatus int

const (
	// SpanStatusPending은 시작되지 않은 스팬을 나타냅니다.
	SpanStatusPending SpanStatus = iota
	// SpanStatusActive는 현재 활성화된 스팬을 나타냅니다.
	SpanStatusActive
	// SpanStatusCompleted는 완료된 스팬을 나타냅니다.
	SpanStatusCompleted
)

// String은 SpanStatus를 문자열로 변환합니다.
func (s SpanStatus) String() string {
	switch s {
	case SpanStatusPending:
		return "pending"
	case SpanStatusActive:
		return "active"
	case SpanStatusCompleted:
		return "completed"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// SpanEvent는 스팬 관련 이벤트를 나타냅니다.
type SpanEvent struct {
	Type      SpanEventType `json:"type" yaml:"type"`
	Span      *Span         `json:"span" yaml:"span"`
	Timestamp time.Time     `json:"timestamp" yaml:"timestamp"`
}

// SpanEventType은 스팬 이벤트의 유형을 나타냅니다.
type SpanEventType int

const (
	// SpanEventTypeStart는 스팬 시작 이벤트를 나타냅니다.
	SpanEventTypeStart SpanEventType = iota
	// SpanEventTypeEnd는 스팬 종료 이벤트를 나타냅니다.
	SpanEventTypeEnd
	// SpanEventTypeValidatorUpdate는 검증자 세트 업데이트 이벤트를 나타냅니다.
	SpanEventTypeValidatorUpdate
)

// String은 SpanEventType을 문자열로 변환합니다.
func (t SpanEventType) String() string {
	switch t {
	case SpanEventTypeStart:
		return "start"
	case SpanEventTypeEnd:
		return "end"
	case SpanEventTypeValidatorUpdate:
		return "validator_update"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// Validator는 단일 검증자 정보를 나타냅니다.
type Validator struct {
	ID               uint64         `json:"id" yaml:"id"`
	Address          crypto.Address `json:"address" yaml:"address"`
	PubKey           crypto.PubKey  `json:"pub_key" yaml:"pub_key"`
	VotingPower      int64          `json:"voting_power" yaml:"voting_power"`
	ProposerPriority int64          `json:"proposer_priority" yaml:"proposer_priority"`
}

// NewSpan은 새로운 스팬을 생성합니다.
func NewSpan(id, startBlock, endBlock uint64, validators types.ValidatorSet, startTime time.Time) *Span {
	return &Span{
		ID:         id,
		StartBlock: startBlock,
		EndBlock:   endBlock,
		Validators: validators,
		StartTime:  startTime,
	}
}

// CompleteSpan은 스팬을 완료 상태로 표시합니다.
func (s *Span) CompleteSpan(endTime time.Time) {
	s.EndTime = endTime
}

// IsActive는 주어진 블록 번호가 현재 스팬 내에 있는지 확인합니다.
func (s *Span) IsActive(blockNumber uint64) bool {
	return blockNumber >= s.StartBlock && blockNumber <= s.EndBlock
}

// NewSpanEvent는 새로운 스팬 이벤트를 생성합니다.
func NewSpanEvent(eventType SpanEventType, span *Span) *SpanEvent {
	return &SpanEvent{
		Type:      eventType,
		Span:      span,
		Timestamp: time.Now(),
	}
}
