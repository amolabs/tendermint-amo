package types

import (
	amino "github.com/tendermint/go-amino"
	"github.com/amolabs/tendermint-amo/types"
)

var cdc = amino.NewCodec()

func init() {
	types.RegisterBlockAmino(cdc)
}
