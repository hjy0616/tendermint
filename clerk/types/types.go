package types

import (
	"time"

	"github.com/tendermint/tendermint/libs/bytes"
)

// EventRecord 구조체는 블록체인 상태 변경 이벤트를 나타냅니다.
type EventRecord struct {
	ID        uint64         `json:"id" yaml:"id"`
	Contract  bytes.HexBytes `json:"contract" yaml:"contract"`
	Data      bytes.HexBytes `json:"data" yaml:"data"`
	TxHash    bytes.HexBytes `json:"tx_hash" yaml:"tx_hash"`
	LogIndex  uint64         `json:"log_index" yaml:"log_index"`
	ChainID   string         `json:"chain_id" yaml:"chain_id"`
	Timestamp int64          `json:"timestamp" yaml:"timestamp"`
}

// EventRecordWithTime는 타임스탬프가 포함된 이벤트 레코드입니다.
type EventRecordWithTime struct {
	EventRecord
	RecordTime time.Time `json:"record_time" yaml:"record_time"`
}

// EventRecordList는 여러 이벤트 레코드를 나타냅니다.
type EventRecordList struct {
	Records []EventRecord `json:"records" yaml:"records"`
}

// EventRecordResponse는 API 응답 형식입니다.
type EventRecordResponse struct {
	Height string      `json:"height"`
	Result EventRecord `json:"result"`
}

// EventRecordListResponse는 다수의 이벤트 레코드에 대한 API 응답 형식입니다.
type EventRecordListResponse struct {
	Height string          `json:"height"`
	Result EventRecordList `json:"result"`
}

// EventSyncStatus는 이벤트 동기화 상태를 나타냅니다.
type EventSyncStatus struct {
	LastSyncedID uint64    `json:"last_synced_id" yaml:"last_synced_id"`
	LastSyncTime time.Time `json:"last_sync_time" yaml:"last_sync_time"`
	IsSyncing    bool      `json:"is_syncing" yaml:"is_syncing"`
}

// EventSyncStatusResponse는 이벤트 동기화 상태에 대한 API 응답 형식입니다.
type EventSyncStatusResponse struct {
	Height string          `json:"height"`
	Result EventSyncStatus `json:"result"`
}
