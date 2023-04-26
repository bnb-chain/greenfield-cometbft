package core

import (
	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// EthQuery handle EVM json-rpc request.
// This is exclusively designed for wallet connections to sign EIP712 typed messages.
// It does not offer any other unnecessary EVM RPC APIs, such as eth_call.
func EthQuery(
	ctx *rpctypes.Context,
	request []byte,
) (*ctypes.ResultEthQuery, error) {
	resEthQuery, err := env.ProxyAppEthQuery.EthQuerySync(abci.RequestEthQuery{
		Request: request,
	})
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultEthQuery{Response: *resEthQuery}, nil
}
