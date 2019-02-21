package main

import (
	"fmt"
	"net/http"
	"os"

	amino "github.com/tendermint/go-amino"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/amolabs/tendermint-amo/libs/log"
	rpcserver "github.com/amolabs/tendermint-amo/rpc/lib/server"
)

var routes = map[string]*rpcserver.RPCFunc{
	"hello_world": rpcserver.NewRPCFunc(HelloWorld, "name,num"),
}

func HelloWorld(name string, num int) (Result, error) {
	return Result{fmt.Sprintf("hi %s %d", name, num)}, nil
}

type Result struct {
	Result string
}

func main() {
	mux := http.NewServeMux()
	cdc := amino.NewCodec()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	rpcserver.RegisterRPCFuncs(mux, routes, cdc, logger)
	listener, err := rpcserver.Listen("0.0.0.0:8008", rpcserver.Config{})
	if err != nil {
		cmn.Exit(err.Error())
	}
	go rpcserver.StartHTTPServer(listener, mux, logger)
	// Wait forever
	cmn.TrapSignal(func() {
	})

}
