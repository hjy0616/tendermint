package coretypes

// 마일스톤 API 응답 타입

// Milestone은 단일 마일스톤 데이터를 나타냅니다.
type Milestone struct {
	ID          uint64 `json:"id"`
	BlockHeight int64  `json:"block_height"`
	RootHash    string `json:"root_hash"`
	StartBlock  int64  `json:"start_block"`
	EndBlock    int64  `json:"end_block"`
	Proposer    string `json:"proposer"`
	Timestamp   int64  `json:"timestamp"`
	Ack         bool   `json:"ack"`
}

// ResultMilestones는 여러 마일스톤 조회 결과를 나타냅니다.
type ResultMilestones struct {
	Milestones []Milestone `json:"milestones"`
}

// ResultMilestone은 단일 마일스톤 조회 결과를 나타냅니다.
type ResultMilestone struct {
	Milestone Milestone `json:"milestone"`
}

// ResultNoAckMilestones는 확인되지 않은 마일스톤 목록 조회 결과를 나타냅니다.
type ResultNoAckMilestones struct {
	Milestones []Milestone `json:"milestones"`
}

// ResultMilestoneCount는 마일스톤 수 조회 결과를 나타냅니다.
type ResultMilestoneCount struct {
	Count uint64 `json:"count"`
}

// ResultAckMilestone은 마일스톤 확인 처리 결과를 나타냅니다.
type ResultAckMilestone struct {
	Success bool `json:"success"`
}

// ResultAddMilestone은 마일스톤 추가 결과를 나타냅니다.
type ResultAddMilestone struct {
	Success bool   `json:"success"`
	ID      uint64 `json:"id"`
}
