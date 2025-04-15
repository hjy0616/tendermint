package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/milestone/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// FetchMilestone은 단일 마일스톤을 조회합니다.
//
// ```shell
// curl 'localhost:26657/fetch_milestone?id=1'
// ```
//
// > 위 명령은 ID가 1인 마일스톤을 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// milestone, err := client.FetchMilestone(context.Background(), 1)
// ```
func FetchMilestone(ctx *rpctypes.Context, idStr string) (*coretypes.ResultMilestone, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 마일스톤 서비스로부터 마일스톤 조회
	milestone, err := env.MilestoneService.GetMilestone(context.Background(), id)
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultMilestone{
		Milestone: convertToRPCMilestone(milestone),
	}

	return result, nil
}

// FetchMilestones는 여러 마일스톤을 조회합니다.
//
// ```shell
// curl 'localhost:26657/fetch_milestones?from_id=1&to_id=10'
// ```
//
// > 위 명령은 ID 1부터 10까지의 마일스톤을 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// milestones, err := client.FetchMilestones(context.Background(), 1, 10)
// ```
func FetchMilestones(ctx *rpctypes.Context, fromIDstr, toIDstr string) (*coretypes.ResultMilestones, error) {
	fromID, err := strconv.ParseUint(fromIDstr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid from_id: %w", err)
	}

	toID, err := strconv.ParseUint(toIDstr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid to_id: %w", err)
	}

	// 마일스톤 서비스로부터 마일스톤 조회
	milestones, err := env.MilestoneService.GetMilestones(context.Background(), fromID, toID)
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultMilestones{}
	for _, milestone := range milestones {
		result.Milestones = append(result.Milestones, convertToRPCMilestone(&milestone))
	}

	return result, nil
}

// FetchMilestoneCount는 마일스톤 수를 조회합니다.
//
// ```shell
// curl 'localhost:26657/fetch_milestone_count'
// ```
//
// > 위 명령은 마일스톤 수를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// count, err := client.FetchMilestoneCount(context.Background())
// ```
func FetchMilestoneCount(ctx *rpctypes.Context) (*coretypes.ResultMilestoneCount, error) {
	// 마일스톤 서비스로부터 마일스톤 수 조회
	count, err := env.MilestoneService.GetMilestoneCount(context.Background())
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultMilestoneCount{
		Count: count,
	}

	return result, nil
}

// FetchNoAckMilestones는 확인되지 않은 마일스톤 목록을 조회합니다.
//
// ```shell
// curl 'localhost:26657/fetch_no_ack_milestones'
// ```
//
// > 위 명령은 확인되지 않은 마일스톤 목록을 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// milestones, err := client.FetchNoAckMilestones(context.Background())
// ```
func FetchNoAckMilestones(ctx *rpctypes.Context) (*coretypes.ResultNoAckMilestones, error) {
	// 마일스톤 서비스로부터 확인되지 않은 마일스톤 조회
	milestones, err := env.MilestoneService.GetNoAckMilestones(context.Background())
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultNoAckMilestones{}
	for _, milestone := range milestones {
		result.Milestones = append(result.Milestones, convertToRPCMilestone(&milestone))
	}

	return result, nil
}

// AckMilestone은 마일스톤을 확인 처리합니다.
//
// ```shell
// curl 'localhost:26657/ack_milestone' \
// --data-binary '{"jsonrpc":"2.0","id":"jsonrpc","method":"ack_milestone","params":{"id":"1"}}'
// ```
//
// > 위 명령은 ID가 1인 마일스톤을 확인 처리하고 성공 여부를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// success, err := client.AckMilestone(context.Background(), 1)
// ```
func AckMilestone(ctx *rpctypes.Context, idStr string) (*coretypes.ResultAckMilestone, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 마일스톤 확인 처리
	err = env.MilestoneService.AckMilestone(context.Background(), id)
	if err != nil {
		return nil, err
	}

	// 응답 구조체 생성
	result := &coretypes.ResultAckMilestone{
		Success: true,
	}

	return result, nil
}

// AddMilestone은 새로운 마일스톤을 추가합니다.
//
// ```shell
// curl 'localhost:26657/add_milestone' \
// --data-binary '{"jsonrpc":"2.0","id":"jsonrpc","method":"add_milestone","params":{"block_height":1000,"root_hash":"0x1234...","start_block":900,"end_block":1000,"proposer":"0x5678..."}}'
// ```
//
// > 위 명령은 새 마일스톤을 생성하고 성공 여부를 반환합니다.
//
// ```go
// client := rpchttp.New("tcp://0.0.0.0:26657")
// success, err := client.AddMilestone(context.Background(), blockHeight, rootHash, startBlock, endBlock, proposer)
// ```
func AddMilestone(ctx *rpctypes.Context, params map[string]interface{}) (*coretypes.ResultAddMilestone, error) {
	// 파라미터 파싱
	blockHeightFloat, ok := params["block_height"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid block_height parameter")
	}
	blockHeight := int64(blockHeightFloat)

	rootHashStr, ok := params["root_hash"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid root_hash parameter")
	}

	startBlockFloat, ok := params["start_block"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid start_block parameter")
	}
	startBlock := int64(startBlockFloat)

	endBlockFloat, ok := params["end_block"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid end_block parameter")
	}
	endBlock := int64(endBlockFloat)

	proposerStr, ok := params["proposer"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid proposer parameter")
	}

	// 마일스톤 ID는 서비스에서 자동 생성
	count, err := env.MilestoneService.GetMilestoneCount(context.Background())
	if err != nil {
		return nil, err
	}
	newID := count + 1

	// 바이트 데이터 파싱
	rootHash, err := bytes.HexBytes(rootHashStr).MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid root hash: %w", err)
	}

	proposer, err := bytes.HexBytes(proposerStr).MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid proposer: %w", err)
	}

	// 새 마일스톤 생성
	var rootHashBytes, proposerBytes bytes.HexBytes
	if err := json.Unmarshal(rootHash, &rootHashBytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal root hash: %w", err)
	}
	if err := json.Unmarshal(proposer, &proposerBytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proposer: %w", err)
	}

	milestone := &types.Milestone{
		ID:          newID,
		BlockHeight: blockHeight,
		RootHash:    rootHashBytes,
		StartBlock:  startBlock,
		EndBlock:    endBlock,
		Proposer:    proposerBytes,
		Timestamp:   time.Now().Unix(),
		Ack:         false,
	}

	// 마일스톤 추가
	if err := env.MilestoneService.AddMilestone(context.Background(), milestone); err != nil {
		return nil, err
	}

	return &coretypes.ResultAddMilestone{
		Success: true,
		ID:      newID,
	}, nil
}

// convertToRPCMilestone는 내부 마일스톤 구조체를 RPC 응답용 구조체로 변환합니다.
func convertToRPCMilestone(milestone *types.Milestone) coretypes.Milestone {
	return coretypes.Milestone{
		ID:          milestone.ID,
		BlockHeight: milestone.BlockHeight,
		RootHash:    milestone.RootHash.String(),
		StartBlock:  milestone.StartBlock,
		EndBlock:    milestone.EndBlock,
		Proposer:    milestone.Proposer.String(),
		Timestamp:   milestone.Timestamp,
		Ack:         milestone.Ack,
	}
}
