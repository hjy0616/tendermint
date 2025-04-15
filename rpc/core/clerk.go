package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/tendermint/tendermint/clerk/types"
	"github.com/tendermint/tendermint/libs/bytes"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// StateSyncEvents는 블록체인의 상태 동기화 이벤트를 조회합니다.
// Tendermint Core를 사용한 ABCI 애플리케이션이 상태 동기화를 관리하기 위해 사용됩니다.
//
// ```shell
// curl 'localhost:26657/state_sync_events?from_id=1&to_id=10'
// ```
//
// > 위 명령은 ID 1부터 10까지의 상태 동기화 이벤트를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// events, err := client.StateSyncEvents(context.Background(), 1, 10)
// ```
func StateSyncEvents(ctx *rpctypes.Context, fromIDstr, toIDstr string) (*coretypes.ResultStateSyncEvents, error) {
	fromID, err := strconv.ParseUint(fromIDstr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid from_id: %w", err)
	}

	toID, err := strconv.ParseUint(toIDstr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid to_id: %w", err)
	}

	// 클러크 서비스로부터 이벤트 레코드 조회
	eventRecords, err := env.ClerkService.GetEventRecords(context.Background(), fromID, toID)
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultStateSyncEvents{}
	for _, record := range eventRecords {
		result.Events = append(result.Events, coretypes.StateSyncEvent{
			ID:        record.ID,
			Contract:  record.Contract.String(),
			Data:      record.Data.String(),
			TxHash:    record.TxHash.String(),
			LogIndex:  record.LogIndex,
			ChainID:   record.ChainID,
			Timestamp: record.Timestamp,
		})
	}

	return result, nil
}

// StateSyncEvent는 단일 상태 동기화 이벤트를 조회합니다.
//
// ```shell
// curl 'localhost:26657/state_sync_event?id=1'
// ```
//
// > 위 명령은 ID가 1인 상태 동기화 이벤트를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// event, err := client.StateSyncEvent(context.Background(), 1)
// ```
func StateSyncEvent(ctx *rpctypes.Context, idStr string) (*coretypes.ResultStateSyncEvent, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 클러크 서비스로부터 단일 이벤트 레코드 조회
	record, err := env.ClerkService.GetEventRecord(context.Background(), id)
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultStateSyncEvent{
		Event: coretypes.StateSyncEvent{
			ID:        record.ID,
			Contract:  record.Contract.String(),
			Data:      record.Data.String(),
			TxHash:    record.TxHash.String(),
			LogIndex:  record.LogIndex,
			ChainID:   record.ChainID,
			Timestamp: record.Timestamp,
		},
	}

	return result, nil
}

// CommitStateSyncEvent는 새로운 상태 동기화 이벤트를 커밋합니다.
//
// ```shell
// curl 'localhost:26657/commit_state_sync_event' \
// --data-binary '{"jsonrpc":"2.0","id":"jsonrpc","method":"commit_state_sync_event","params":{"contract":"0x1234...","data":"0xabcd...","tx_hash":"0x5678...","log_index":1,"chain_id":"eirene"}}'
// ```
//
// > 위 명령은 새 상태 동기화 이벤트를 생성하고 성공 여부를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// success, err := client.CommitStateSyncEvent(context.Background(), contract, data, txHash, logIndex, chainID)
// ```
func CommitStateSyncEvent(ctx *rpctypes.Context, params map[string]interface{}) (*coretypes.ResultCommitStateSyncEvent, error) {
	// 파라미터 파싱
	contractStr, ok := params["contract"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid contract parameter")
	}

	dataStr, ok := params["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid data parameter")
	}

	txHashStr, ok := params["tx_hash"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid tx_hash parameter")
	}

	logIndexFloat, ok := params["log_index"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid log_index parameter")
	}
	logIndex := uint64(logIndexFloat)

	chainID, ok := params["chain_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid chain_id parameter")
	}

	// 이벤트 ID는 서비스에서 자동 생성
	lastStatus, err := env.ClerkService.GetSyncStatus(context.Background())
	if err != nil {
		return nil, err
	}
	newID := lastStatus.LastSyncedID + 1

	// 바이트 데이터 파싱
	contract, err := bytes.HexBytes(contractStr).MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	data, err := bytes.HexBytes(dataStr).MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid data: %w", err)
	}

	txHash, err := bytes.HexBytes(txHashStr).MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid tx hash: %w", err)
	}

	// 새 이벤트 레코드 생성
	var contractBytes, dataBytes, txHashBytes bytes.HexBytes
	if err := json.Unmarshal(contract, &contractBytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract: %w", err)
	}
	if err := json.Unmarshal(data, &dataBytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	if err := json.Unmarshal(txHash, &txHashBytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tx hash: %w", err)
	}

	record := &types.EventRecord{
		ID:        newID,
		Contract:  contractBytes,
		Data:      dataBytes,
		TxHash:    txHashBytes,
		LogIndex:  logIndex,
		ChainID:   chainID,
		Timestamp: time.Now().Unix(),
	}

	// 이벤트 레코드 추가
	if err := env.ClerkService.AddEventRecord(context.Background(), record); err != nil {
		return nil, err
	}

	return &coretypes.ResultCommitStateSyncEvent{
		Success: true,
		ID:      newID,
	}, nil
}

// GetSyncStatus는 상태 동기화 상태를 조회합니다.
//
// ```shell
// curl 'localhost:26657/clerk_sync_status'
// ```
//
// > 위 명령은 현재 상태 동기화 상태를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// status, err := client.GetSyncStatus(context.Background())
// ```
func GetSyncStatus(ctx *rpctypes.Context) (*coretypes.ResultClerkSyncStatus, error) {
	// 클러크 서비스로부터 동기화 상태 조회
	status, err := env.ClerkService.GetSyncStatus(context.Background())
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultClerkSyncStatus{
		LastSyncedID: status.LastSyncedID,
		LastSyncTime: status.LastSyncTime.Unix(),
		IsSyncing:    status.IsSyncing,
	}

	return result, nil
}
