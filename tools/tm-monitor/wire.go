package main

import (
	amino "github.com/tendermint/go-amino"
	ctypes "github.com/amolabs/tendermint-amo/rpc/core/types"
)

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}
