package core

import (
	"context"
	"time"

	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/coretypes"
	rpc "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	tmstate "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/state/indexer"
	"github.com/tendermint/tendermint/types"
)

// TODO: better system than "unsafe" prefix

const (
	// see README
	defaultPerPage = 30
	maxPerPage     = 100

	// SubscribeTimeout is the maximum time we wait to subscribe for an event.
	// must be less than the server's write timeout (see rpcserver.DefaultConfig)
	SubscribeTimeout = 5 * time.Second
)

//----------------------------------------------
// These interfaces are used by RPC and must be thread safe

type Consensus interface {
	GetState() tmstate.State
	GetValidators() (int64, []*types.Validator)
	GetLastHeight() int64
	GetRoundStateJSON() ([]byte, error)
	GetRoundStateSimpleJSON() ([]byte, error)
}

type transport interface {
	Listeners() []string
	IsListening() bool
	NodeInfo() p2p.NodeInfo
}

type Environment struct {
	ProxyAppQuery   proxy.AppConnQuery
	ProxyAppMempool proxy.AppConnMempool

	StateStore       tmstate.Store
	BlockStore       tmstate.BlockStore
	EvidencePool     tmstate.EvidencePool
	ConsensusState   Consensus
	ConsensusReactor interface {
		WaitSync() bool
	}
	P2PPeers     p2p.Switch
	P2PTransport transport

	TxIndexer  indexer.TxIndexer
	BlockIndex indexer.BlockIndexer
	EventBus   types.BlockEventPublisher

	GenDoc   *types.GenesisDoc // cache the genesis structure
	LogLevel log.Option
	Logger   log.Logger
	Config   *config.RPCConfig
	ReadCtx  context.Context
	WriteCtx context.Context
}

//----------------------------------------------
// These routes are Endpoints to be used with JSONRPC via HTTP or WebSockets/Events

// Route is a function handler for a given rpc endpoint
type RPCFunc = coretypes.RPCFunc

// Route is used to define the functionality of an endpoint
type Route struct {
	Name     string
	Endpoint string
	Method   string
	Path     string
	Func     RPCFunc
}

// Routes is a collection of Route capabilities
var Routes = []Route{
	{
		"subscribe",
		"subscribe",
		"GET,POST",
		"",
		Subscribe,
	},
	{
		"unsubscribe",
		"unsubscribe",
		"GET,POST",
		"",
		Unsubscribe,
	},
	{
		"unsubscribe_all",
		"unsubscribe_all",
		"GET,POST",
		"",
		UnsubscribeAll,
	},
	{
		"health",
		"health",
		"GET,POST",
		"",
		Health,
	},
	{
		"status",
		"status",
		"GET,POST",
		"",
		Status,
	},
	{
		"net_info",
		"net_info",
		"GET,POST",
		"",
		NetInfo,
	},
	{
		"blockchain",
		"blockchain",
		"GET,POST",
		"",
		BlockchainInfo,
	},
	{
		"genesis",
		"genesis",
		"GET,POST",
		"",
		Genesis,
	},
	{
		"genesis_chunked",
		"genesis_chunked",
		"GET,POST",
		"",
		GenesisChunked,
	},
	{
		"block",
		"block",
		"GET,POST",
		"",
		Block,
	},
	{
		"block_by_hash",
		"block_by_hash",
		"GET,POST",
		"",
		BlockByHash,
	},
	{
		"block_results",
		"block_results",
		"GET,POST",
		"",
		BlockResults,
	},
	{
		"commit",
		"commit",
		"GET,POST",
		"",
		Commit,
	},
	{
		"check_tx",
		"check_tx",
		"GET,POST",
		"",
		CheckTx,
	},
	{
		"tx",
		"tx",
		"GET,POST",
		"",
		Tx,
	},
	{
		"tx_search",
		"tx_search",
		"GET,POST",
		"",
		TxSearch,
	},
	{
		"block_search",
		"block_search",
		"GET,POST",
		"",
		BlockSearch,
	},
	{
		"validators",
		"validators",
		"GET,POST",
		"",
		Validators,
	},
	{
		"dump_consensus_state",
		"dump_consensus_state",
		"GET,POST",
		"",
		DumpConsensusState,
	},
	{
		"consensus_state",
		"consensus_state",
		"GET,POST",
		"",
		ConsensusState,
	},
	{
		"consensus_params",
		"consensus_params",
		"GET,POST",
		"",
		ConsensusParams,
	},
	{
		"unconfirmed_txs",
		"unconfirmed_txs",
		"GET,POST",
		"",
		UnconfirmedTxs,
	},
	{
		"num_unconfirmed_txs",
		"num_unconfirmed_txs",
		"GET,POST",
		"",
		NumUnconfirmedTxs,
	},
	{
		"broadcast_tx_commit",
		"broadcast_tx_commit",
		"GET,POST",
		"",
		BroadcastTxCommit,
	},
	{
		"broadcast_tx_sync",
		"broadcast_tx_sync",
		"GET,POST",
		"",
		BroadcastTxSync,
	},
	{
		"broadcast_tx_async",
		"broadcast_tx_async",
		"GET,POST",
		"",
		BroadcastTxAsync,
	},
	{
		"abci_query",
		"abci_query",
		"GET,POST",
		"",
		ABCIQuery,
	},
	{
		"abci_info",
		"abci_info",
		"GET,POST",
		"",
		ABCIInfo,
	},
	// Evidence
	{
		"broadcast_evidence",
		"broadcast_evidence",
		"GET,POST",
		"",
		BroadcastEvidence,
	},
	// 체크포인트 API 경로 추가
	{
		"checkpoint",
		"checkpoint",
		"GET,POST",
		"",
		Checkpoint,
	},
	{
		"checkpoint_count",
		"checkpoint_count",
		"GET,POST",
		"",
		CheckpointCount,
	},
	{
		"create_checkpoint",
		"create_checkpoint",
		"POST",
		"",
		CreateCheckpoint,
	},
	{
		"prune_checkpoints",
		"prune_checkpoints",
		"POST",
		"",
		PruneCheckpoints,
	},
}

// // List returns list of available endpoints that can be used to query
// func List() []string {
// 	routes := make([]string, len(Routes))
// 	for i, route := range Routes {
// 		routes[i] = route.Name
// 	}
// 	return routes
// }

//-----------------------------------------------------------------------------
// Utils

// coreChainID is a global variable that is set through environment
var coreChainID string

// ChainID returns the ChainID.
func ChainID() string {
	return coreChainID
}

// SetChainID sets the ChainID.
func SetChainID(chainID string) {
	coreChainID = chainID
}

// AddUnsafeRoutes adds unsafe routes.
func AddUnsafeRoutes() {
	// control API
	Routes["dial_seeds"] = rpc.NewRPCFunc(UnsafeDialSeeds, "seeds")
	Routes["dial_peers"] = rpc.NewRPCFunc(UnsafeDialPeers, "peers,persistent,unconditional,private")
	Routes["unsafe_flush_mempool"] = rpc.NewRPCFunc(UnsafeFlushMempool, "")
}
