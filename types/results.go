package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/merkle"
)

// ABCIResults wraps the begin block result, deliver tx results, and end block result to return a proof.
type ABCIResults struct {
	ResponseBeginBlock *abci.ResponseBeginBlock
	ResponseDeliverTxs []*abci.ResponseDeliverTx
	ResponseEndBlock   *abci.ResponseEndBlock
}

// NewResults strips non-deterministic fields from ResponseBeginBlock/ResponseDeliverTx/ResponseEndBlock responses
// and returns ABCIResults.
func NewResults(beginBlock *abci.ResponseBeginBlock, deliverTxs []*abci.ResponseDeliverTx, endBlock *abci.ResponseEndBlock) ABCIResults {
	res := make([]*abci.ResponseDeliverTx, len(deliverTxs))
	for i, d := range deliverTxs {
		res[i] = deterministicResponseDeliverTx(d)
	}
	return ABCIResults{
		ResponseBeginBlock: deterministicResponseBeginBlock(beginBlock),
		ResponseDeliverTxs: res,
		ResponseEndBlock:   deterministicResponseEndBlock(endBlock),
	}
}

// Hash returns a merkle hash of all results.
func (a ABCIResults) Hash() []byte {
	return merkle.HashFromByteSlices(a.toByteSlices())
}

// ProveResult returns a merkle proof of one result from the set
func (a ABCIResults) ProveResult(i int) merkle.Proof {
	_, proofs := merkle.ProofsFromByteSlices(a.toByteSlices())
	return *proofs[i]
}

func (a ABCIResults) toByteSlices() [][]byte {
	l := len(a.ResponseDeliverTxs)
	bzs := make([][]byte, l+2)

	bz, err := a.ResponseBeginBlock.Marshal()
	if err != nil {
		panic(err)
	}
	bzs[0] = bz

	for i := 0; i < l; i++ {
		bz, err = a.ResponseDeliverTxs[i].Marshal()
		if err != nil {
			panic(err)
		}
		bzs[i+1] = bz
	}

	bz, err = a.ResponseEndBlock.Marshal()
	if err != nil {
		panic(err)
	}
	bzs[l+1] = bz

	return bzs
}

// deterministicResponseDeliverTx strips non-deterministic fields from
// ResponseDeliverTx and returns another ResponseDeliverTx.
func deterministicResponseDeliverTx(response *abci.ResponseDeliverTx) *abci.ResponseDeliverTx {
	return &abci.ResponseDeliverTx{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}

// deterministicResponseDeliverTx strips non-deterministic fields from
// ResponseBeginBlock and returns another ResponseBeginBlock.
func deterministicResponseBeginBlock(response *abci.ResponseBeginBlock) *abci.ResponseBeginBlock {
	return &abci.ResponseBeginBlock{
		ExtraData: response.ExtraData,
	}
}

// deterministicResponseEndBlock strips non-deterministic fields from
// ResponseEndBlock and returns another ResponseEndBlock.
func deterministicResponseEndBlock(response *abci.ResponseEndBlock) *abci.ResponseEndBlock {
	return &abci.ResponseEndBlock{
		ExtraData: response.ExtraData,
	}
}
