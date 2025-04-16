package coretypes

import (
	"time"
)

// 스팬 API 응답 타입

// ResultSpan은 스팬 조회 결과를 나타냅니다.
type ResultSpan struct {
	SpanID         uint64    `json:"span_id"`
	StartBlock     uint64    `json:"start_block"`
	EndBlock       uint64    `json:"end_block"`
	Status         string    `json:"status"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time,omitempty"`
	ValidatorCount int       `json:"validator_count"`
}

// ResultSpans는 여러 스팬 조회 결과를 나타냅니다.
type ResultSpans struct {
	Spans []ResultSpan `json:"spans"`
}

// ResultCreateSpan은 스팬 생성 결과를 나타냅니다.
type ResultCreateSpan struct {
	Success bool   `json:"success"`
	SpanID  uint64 `json:"span_id"`
	Error   string `json:"error,omitempty"`
}
