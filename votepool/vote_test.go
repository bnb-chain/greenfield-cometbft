package votepool

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v4/crypto/bls/blst"
	"github.com/stretchr/testify/assert"
)

func TestVote_ValidateBasic(t *testing.T) {
	privKey, _ := blst.RandKey()
	pubKey := privKey.PublicKey().Marshal()
	eventHash := common.HexToHash("0xeefacfed87736ae1d8e8640f6fd7951862997782e5e79842557923e2779d5d5a").Bytes()
	secKey, _ := blst.SecretKeyFromBytes(privKey.Marshal())
	sign := secKey.Sign(eventHash).Marshal()

	testCases := []struct {
		vote Vote
		err  bool
		msg  string
	}{
		{
			vote: Vote{
				PubKey:    pubKey,
				Signature: sign,
				EventType: FromBscCrossChainEvent,
				EventHash: eventHash,
				expireAt:  time.Time{},
			},
			err: false,
			msg: "valid vote",
		},
		{
			vote: Vote{
				PubKey:    pubKey,
				Signature: sign,
				EventType: 10,
				EventHash: eventHash,
				expireAt:  time.Time{},
			},
			err: true,
			msg: "invalid event type",
		},
		{
			vote: Vote{
				PubKey:    pubKey,
				Signature: sign,
				EventType: FromBscCrossChainEvent,
				EventHash: eventHash[0:12],
				expireAt:  time.Time{},
			},
			err: true,
			msg: "invalid event hash",
		},
		{
			vote: Vote{
				PubKey:    pubKey[0:10],
				Signature: sign,
				EventType: FromBscCrossChainEvent,
				EventHash: eventHash,
				expireAt:  time.Time{},
			},
			err: true,
			msg: "invalid public key",
		},
		{
			vote: Vote{
				PubKey:    pubKey,
				Signature: sign[0:48],
				EventType: FromBscCrossChainEvent,
				EventHash: eventHash,
				expireAt:  time.Time{},
			},
			err: true,
			msg: "invalid signature",
		},
	}

	for _, tc := range testCases {
		err := tc.vote.ValidateBasic()
		if tc.err {
			if assert.Error(t, err) {
				assert.Equal(t, tc.msg, err.Error())
			}
		} else {
			assert.NoError(t, err)
		}
	}
}
