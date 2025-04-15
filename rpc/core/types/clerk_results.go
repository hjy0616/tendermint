package coretypes

// 상태 동기화 이벤트 API 응답 타입

// StateSyncEvent는 단일 상태 동기화 이벤트를 나타냅니다.
type StateSyncEvent struct {
	ID        uint64 `json:"id"`
	Contract  string `json:"contract"`
	Data      string `json:"data"`
	TxHash    string `json:"tx_hash"`
	LogIndex  uint64 `json:"log_index"`
	ChainID   string `json:"chain_id"`
	Timestamp int64  `json:"timestamp"`
}

// ResultStateSyncEvents는 여러 상태 동기화 이벤트 조회 결과를 나타냅니다.
type ResultStateSyncEvents struct {
	Events []StateSyncEvent `json:"events"`
}

// ResultStateSyncEvent는 단일 상태 동기화 이벤트 조회 결과를 나타냅니다.
type ResultStateSyncEvent struct {
	Event StateSyncEvent `json:"event"`
}

// ResultCommitStateSyncEvent는 상태 동기화 이벤트 커밋 결과를 나타냅니다.
type ResultCommitStateSyncEvent struct {
	Success bool   `json:"success"`
	ID      uint64 `json:"id"`
}

// ResultClerkSyncStatus는 클러크 동기화 상태 조회 결과를 나타냅니다.
type ResultClerkSyncStatus struct {
	LastSyncedID uint64 `json:"last_synced_id"`
	LastSyncTime int64  `json:"last_sync_time"`
	IsSyncing    bool   `json:"is_syncing"`
}
