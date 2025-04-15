package types

import (
	"time"

	"github.com/tendermint/tendermint/libs/bytes"
)

// Milestone은 블록체인의 중요 지점을 나타내는 구조체입니다.
type Milestone struct {
	ID          uint64         `json:"id" yaml:"id"`
	BlockHeight int64          `json:"block_height" yaml:"block_height"`
	RootHash    bytes.HexBytes `json:"root_hash" yaml:"root_hash"`
	StartBlock  int64          `json:"start_block" yaml:"start_block"`
	EndBlock    int64          `json:"end_block" yaml:"end_block"`
	Proposer    bytes.HexBytes `json:"proposer" yaml:"proposer"`
	Timestamp   int64          `json:"timestamp" yaml:"timestamp"`
	Ack         bool           `json:"ack" yaml:"ack"`
}

// MilestoneWithTime은 타임스탬프가 포함된 마일스톤입니다.
type MilestoneWithTime struct {
	Milestone
	RecordTime time.Time `json:"record_time" yaml:"record_time"`
}

// MilestoneList는 여러 마일스톤을 나타냅니다.
type MilestoneList struct {
	Milestones []Milestone `json:"milestones" yaml:"milestones"`
}

// MilestoneResponse는 API 응답 형식입니다.
type MilestoneResponse struct {
	Height string    `json:"height"`
	Result Milestone `json:"result"`
}

// MilestoneListResponse는 다수의 마일스톤에 대한 API 응답 형식입니다.
type MilestoneListResponse struct {
	Height string        `json:"height"`
	Result MilestoneList `json:"result"`
}

// MilestoneCount는 마일스톤 수를 나타냅니다.
type MilestoneCount struct {
	Count uint64 `json:"count" yaml:"count"`
}

// MilestoneCountResponse는 마일스톤 수에 대한 API 응답 형식입니다.
type MilestoneCountResponse struct {
	Height string         `json:"height"`
	Result MilestoneCount `json:"result"`
}

// NoAckMilestoneList는 확인되지 않은 마일스톤 목록을 나타냅니다.
type NoAckMilestoneList struct {
	Milestones []Milestone `json:"milestones" yaml:"milestones"`
}

// NoAckMilestoneListResponse는 확인되지 않은 마일스톤 목록에 대한 API 응답 형식입니다.
type NoAckMilestoneListResponse struct {
	Height string             `json:"height"`
	Result NoAckMilestoneList `json:"result"`
}
