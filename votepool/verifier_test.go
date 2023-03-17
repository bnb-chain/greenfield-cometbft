package votepool

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"
)

func TestVoteFromValidatorVerifier(t *testing.T) {
	pubKey1 := ed25519.GenPrivKey().PubKey()
	blsPrivKey1, _ := blst.RandKey()
	blsPubKey1 := blsPrivKey1.PublicKey().Marshal()
	val1 := &types.Validator{Address: pubKey1.Address(), PubKey: pubKey1, BlsKey: blsPubKey1, VotingPower: 10}

	pubKey2 := ed25519.GenPrivKey().PubKey()
	blsPrivKey2, _ := blst.RandKey()
	blsPubKey2 := blsPrivKey2.PublicKey().Marshal()
	val2 := &types.Validator{Address: pubKey2.Address(), PubKey: pubKey2, BlsKey: blsPubKey2, VotingPower: 10}

	vals := make([]*types.Validator, 0)
	vals = append(vals, val1)
	vals = append(vals, val2)

	verifier := NewFromValidatorVerifier()
	verifier.initValidators(vals)

	voteFromVal1 := Vote{PubKey: blsPubKey1}
	err := verifier.Validate(&voteFromVal1)
	require.NoError(t, err)

	blsPrivKey, _ := blst.RandKey()
	blsPubKey := blsPrivKey.PublicKey().Marshal()
	voteFromOthers := Vote{PubKey: blsPubKey}
	err = verifier.Validate(&voteFromOthers)
	require.Error(t, err)
}

func TestVoteFromValidatorVerifier_UpdateValidators(t *testing.T) {
	pubKey1 := ed25519.GenPrivKey().PubKey()
	blsPrivKey1, _ := blst.RandKey()
	blsPubKey1 := blsPrivKey1.PublicKey().Marshal()
	val1 := &types.Validator{Address: pubKey1.Address(), PubKey: pubKey1, BlsKey: blsPubKey1, VotingPower: 10}

	pubKey2 := ed25519.GenPrivKey().PubKey()
	blsPrivKey2, _ := blst.RandKey()
	blsPubKey2 := blsPrivKey2.PublicKey().Marshal()
	val2 := &types.Validator{Address: pubKey2.Address(), PubKey: pubKey2, BlsKey: blsPubKey2, VotingPower: 10}

	vals := make([]*types.Validator, 0)
	vals = append(vals, val1)
	vals = append(vals, val2)

	verifier := NewFromValidatorVerifier()
	verifier.initValidators(vals)

	//remove validator
	removeVal := &types.Validator{PubKey: pubKey1, Address: pubKey1.Address(), VotingPower: 0}
	verifier.updateValidators([]*types.Validator{removeVal})

	require.Equal(t, 1, len(verifier.validators))

	//add validator
	pubKey3 := ed25519.GenPrivKey().PubKey()
	blsPrivKey3, _ := blst.RandKey()
	blsPubKey3 := blsPrivKey3.PublicKey().Marshal()

	addVal := &types.Validator{PubKey: pubKey3, Address: pubKey3.Address(), BlsKey: blsPubKey3, VotingPower: 10}
	verifier.updateValidators([]*types.Validator{addVal})

	require.Equal(t, 2, len(verifier.validators))
}

func TestVoteBlsVerifier(t *testing.T) {
	privKey, _ := blst.RandKey()
	pubKey := privKey.PublicKey().Marshal()
	eventHash := common.HexToHash("0xeefacfed87736ae1d8e8640f6fd7951862997782e5e79842557923e2779d5d5a").Bytes()
	secKey, _ := blst.SecretKeyFromBytes(privKey.Marshal())
	sign := secKey.Sign(eventHash).Marshal()

	verifier := &BlsSignatureVerifier{}

	vote1 := Vote{
		PubKey:    pubKey,
		Signature: sign,
		EventType: 0,
		EventHash: eventHash,
		expireAt:  time.Time{},
	}
	err := verifier.Validate(&vote1)
	require.NoError(t, err)

	vote2 := Vote{
		PubKey:    pubKey,
		Signature: sign,
		EventType: 0,
		EventHash: common.HexToHash("0xb3989c2ba4b4b91b35162c137c154848f7261e16ce3f6d8c88f64cf06b737a3c").Bytes(),
		expireAt:  time.Time{},
	}
	err = verifier.Validate(&vote2)
	require.Error(t, err)
}
