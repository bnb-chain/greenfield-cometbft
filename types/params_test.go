package types

import (
	"bytes"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

var (
	valEd25519   = []string{ABCIPubKeyTypeEd25519}
	valSecp256k1 = []string{ABCIPubKeyTypeSecp256k1}
)

func TestConsensusParamsValidation(t *testing.T) {
	testCases := []struct {
		params ConsensusParams
		valid  bool
	}{
		// test block params
		0: {makeParams(2400, 1, 0, 2, 0, valEd25519), true},
		1: {makeParams(2400, 0, 0, 2, 0, valEd25519), false},
		2: {makeParams(2400, 47*1024*1024, 0, 2, 0, valEd25519), true},
		3: {makeParams(2400, 10, 0, 2, 0, valEd25519), true},
		4: {makeParams(2400, 100*1024*1024, 0, 2, 0, valEd25519), true},
		5: {makeParams(2400, 101*1024*1024, 0, 2, 0, valEd25519), false},
		6: {makeParams(2400, 1024*1024*1024, 0, 2, 0, valEd25519), false},
		// test evidence params
		7:  {makeParams(2400, 1, 0, 0, 0, valEd25519), false},
		8:  {makeParams(2400, 1, 0, 2, 2, valEd25519), false},
		9:  {makeParams(2400, 1000, 0, 2, 1, valEd25519), true},
		10: {makeParams(2400, 1, 0, -1, 0, valEd25519), false},
		// test no pubkey type provided
		11: {makeParams(2400, 1, 0, 2, 0, []string{}), false},
		// test invalid pubkey type provided
		12: {makeParams(2400, 1, 0, 2, 0, []string{"potatoes make good pubkeys"}), false},
	}
	for i, tc := range testCases {
		if tc.valid {
			assert.NoErrorf(t, tc.params.ValidateBasic(), "expected no error for valid params (#%d)", i)
		} else {
			assert.Errorf(t, tc.params.ValidateBasic(), "expected error for non valid params (#%d)", i)
		}
	}
}

func makeParams(
	blockTxs, blockBytes, blockGas int64,
	evidenceAge int64,
	maxEvidenceBytes int64,
	pubkeyTypes []string,
) ConsensusParams {
	return ConsensusParams{
		Block: BlockParams{
			MaxBytes: blockBytes,
			MaxGas:   blockGas,
			MaxTxs:   blockTxs,
		},
		Evidence: EvidenceParams{
			MaxAgeNumBlocks: evidenceAge,
			MaxAgeDuration:  time.Duration(evidenceAge),
			MaxBytes:        maxEvidenceBytes,
		},
		Validator: ValidatorParams{
			PubKeyTypes: pubkeyTypes,
		},
	}
}

func TestConsensusParamsHash(t *testing.T) {
	params := []ConsensusParams{
		makeParams(2400, 4, 2, 3, 1, valEd25519),
		makeParams(2400, 1, 4, 3, 1, valEd25519),
		makeParams(2400, 1, 2, 4, 1, valEd25519),
		makeParams(2400, 2, 5, 7, 1, valEd25519),
		makeParams(2400, 1, 7, 6, 1, valEd25519),
		makeParams(2400, 9, 5, 4, 1, valEd25519),
		makeParams(2400, 7, 8, 9, 1, valEd25519),
		makeParams(2400, 4, 6, 5, 1, valEd25519),
	}

	hashes := make([][]byte, len(params))
	for i := range params {
		hashes[i] = params[i].Hash()
	}

	// make sure there are no duplicates...
	// sort, then check in order for matches
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i], hashes[j]) < 0
	})
	for i := 0; i < len(hashes)-1; i++ {
		assert.NotEqual(t, hashes[i], hashes[i+1])
	}
}

func TestConsensusParamsUpdate(t *testing.T) {
	testCases := []struct {
		params        ConsensusParams
		updates       *cmtproto.ConsensusParams
		updatedParams ConsensusParams
	}{
		// empty updates
		{
			makeParams(2400, 1, 2, 3, 0, valEd25519),
			&cmtproto.ConsensusParams{},
			makeParams(2400, 1, 2, 3, 0, valEd25519),
		},
		// fine updates
		{
			makeParams(2400, 1, 2, 3, 0, valEd25519),
			&cmtproto.ConsensusParams{
				Block: &cmtproto.BlockParams{
					MaxBytes: 100,
					MaxGas:   200,
				},
				Evidence: &cmtproto.EvidenceParams{
					MaxAgeNumBlocks: 300,
					MaxAgeDuration:  time.Duration(300),
					MaxBytes:        50,
				},
				Validator: &cmtproto.ValidatorParams{
					PubKeyTypes: valSecp256k1,
				},
			},
			makeParams(2400, 100, 200, 300, 50, valSecp256k1),
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.updatedParams, tc.params.Update(tc.updates))
	}
}

func TestConsensusParamsUpdate_AppVersion(t *testing.T) {
	params := makeParams(2400, 1, 2, 3, 0, valEd25519)

	assert.EqualValues(t, 0, params.Version.App)

	updated := params.Update(
		&cmtproto.ConsensusParams{Version: &cmtproto.VersionParams{App: 1}})

	assert.EqualValues(t, 1, updated.Version.App)
}

func TestProto(t *testing.T) {
	params := []ConsensusParams{
		makeParams(2400, 4, 2, 3, 1, valEd25519),
		makeParams(2400, 1, 4, 3, 1, valEd25519),
		makeParams(2400, 1, 2, 4, 1, valEd25519),
		makeParams(2400, 2, 5, 7, 1, valEd25519),
		makeParams(2400, 1, 7, 6, 1, valEd25519),
		makeParams(2400, 9, 5, 4, 1, valEd25519),
		makeParams(2400, 7, 8, 9, 1, valEd25519),
		makeParams(2400, 4, 6, 5, 1, valEd25519),
	}

	for i := range params {
		pbParams := params[i].ToProto()

		oriParams := ConsensusParamsFromProto(pbParams)

		assert.Equal(t, params[i], oriParams)

	}
}
