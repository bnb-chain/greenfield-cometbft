package mempool

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

// TxInfo are parameters that get passed when attempting to add a tx to the
// mempool.
type TxInfo struct {
	// SenderID is the internal peer ID used in the mempool to identify the
	// sender, storing two bytes with each transaction instead of 20 bytes for
	// the types.NodeID.
	SenderID uint16

	// SenderP2PID is the actual p2p.ID of the sender, used e.g. for logging.
	SenderP2PID p2p.ID
}

// CheckTxRequest is a request to CheckTx.
type CheckTxRequest struct {
	Tx     types.Tx
	CB     func(*abci.Response)
	TxInfo TxInfo
	Err    chan error
}
