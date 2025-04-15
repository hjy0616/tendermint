package core

import (
	rpc "github.com/tendermint/tendermint/rpc/jsonrpc/server"
)

// TODO: better system than "unsafe" prefix

// Routes is a map of available routes.
var Routes = map[string]*rpc.RPCFunc{
	// subscribe/unsubscribe are reserved for websocket events.
	"subscribe":       rpc.NewWSRPCFunc(Subscribe, "query"),
	"unsubscribe":     rpc.NewWSRPCFunc(Unsubscribe, "query"),
	"unsubscribe_all": rpc.NewWSRPCFunc(UnsubscribeAll, ""),

	// info API
	"health":               rpc.NewRPCFunc(Health, ""),
	"status":               rpc.NewRPCFunc(Status, ""),
	"net_info":             rpc.NewRPCFunc(NetInfo, ""),
	"blockchain":           rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight", rpc.Cacheable()),
	"genesis":              rpc.NewRPCFunc(Genesis, "", rpc.Cacheable()),
	"genesis_chunked":      rpc.NewRPCFunc(GenesisChunked, "chunk", rpc.Cacheable()),
	"block":                rpc.NewRPCFunc(Block, "height", rpc.Cacheable("height")),
	"block_by_hash":        rpc.NewRPCFunc(BlockByHash, "hash", rpc.Cacheable()),
	"block_results":        rpc.NewRPCFunc(BlockResults, "height", rpc.Cacheable("height")),
	"commit":               rpc.NewRPCFunc(Commit, "height", rpc.Cacheable("height")),
	"check_tx":             rpc.NewRPCFunc(CheckTx, "tx"),
	"tx":                   rpc.NewRPCFunc(Tx, "hash,prove", rpc.Cacheable()),
	"tx_search":            rpc.NewRPCFunc(TxSearch, "query,prove,page,per_page,order_by"),
	"block_search":         rpc.NewRPCFunc(BlockSearch, "query,page,per_page,order_by"),
	"validators":           rpc.NewRPCFunc(Validators, "height,page,per_page", rpc.Cacheable("height")),
	"dump_consensus_state": rpc.NewRPCFunc(DumpConsensusState, ""),
	"consensus_state":      rpc.NewRPCFunc(ConsensusState, ""),
	"consensus_params":     rpc.NewRPCFunc(ConsensusParams, "height", rpc.Cacheable("height")),
	"unconfirmed_txs":      rpc.NewRPCFunc(UnconfirmedTxs, "limit"),
	"num_unconfirmed_txs":  rpc.NewRPCFunc(NumUnconfirmedTxs, ""),

	// tx broadcast API
	"broadcast_tx_commit": rpc.NewRPCFunc(BroadcastTxCommit, "tx"),
	"broadcast_tx_sync":   rpc.NewRPCFunc(BroadcastTxSync, "tx"),
	"broadcast_tx_async":  rpc.NewRPCFunc(BroadcastTxAsync, "tx"),

	// abci API
	"abci_query": rpc.NewRPCFunc(ABCIQuery, "path,data,height,prove"),
	"abci_info":  rpc.NewRPCFunc(ABCIInfo, "", rpc.Cacheable()),

	// evidence API
	"broadcast_evidence": rpc.NewRPCFunc(BroadcastEvidence, "evidence"),

	// Clerk 시스템 관련 엔드포인트
	"state_sync_events":       rpc.NewRPCFunc(StateSyncEvents, "from_id,to_id"),
	"state_sync_event":        rpc.NewRPCFunc(StateSyncEvent, "id"),
	"commit_state_sync_event": rpc.NewRPCFunc(CommitStateSyncEvent, "params"),
	"clerk_sync_status":       rpc.NewRPCFunc(GetSyncStatus, ""),

	// Milestone 시스템 관련 엔드포인트
	"fetch_milestone":         rpc.NewRPCFunc(FetchMilestone, "id"),
	"fetch_milestones":        rpc.NewRPCFunc(FetchMilestones, "from_id,to_id"),
	"fetch_milestone_count":   rpc.NewRPCFunc(FetchMilestoneCount, ""),
	"fetch_no_ack_milestones": rpc.NewRPCFunc(FetchNoAckMilestones, ""),
	"ack_milestone":           rpc.NewRPCFunc(AckMilestone, "id"),
	"add_milestone":           rpc.NewRPCFunc(AddMilestone, "params"),
}

// AddUnsafeRoutes adds unsafe routes.
func AddUnsafeRoutes() {
	// control API
	Routes["dial_seeds"] = rpc.NewRPCFunc(UnsafeDialSeeds, "seeds")
	Routes["dial_peers"] = rpc.NewRPCFunc(UnsafeDialPeers, "peers,persistent,unconditional,private")
	Routes["unsafe_flush_mempool"] = rpc.NewRPCFunc(UnsafeFlushMempool, "")
}
