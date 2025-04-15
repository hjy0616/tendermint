package checkpoint

import (
	"github.com/tendermint/tendermint/checkpoint/types"
	amino "github.com/tendermint/tendermint/libs/amino"
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterInterface((*types.Store)(nil), nil)
	cdc.RegisterConcrete(&types.Checkpoint{}, "tendermint/checkpoint", nil)
}
