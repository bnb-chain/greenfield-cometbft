package votepool

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
	blsCommon "github.com/prysmaticlabs/prysm/crypto/bls/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"
)

func makeVotePool() (blsCommon.SecretKey, *types.Validator, blsCommon.SecretKey, *types.Validator, *types.EventBus, *Pool) {
	pubKey1 := ed25519.GenPrivKey().PubKey()
	blsPrivKey1, _ := blst.RandKey()
	blsPubKey1 := blsPrivKey1.PublicKey().Marshal()
	val1 := &types.Validator{Address: pubKey1.Address(), PubKey: pubKey1, BlsKey: blsPubKey1, VotingPower: 10}

	pubKey2 := ed25519.GenPrivKey().PubKey()
	blsPrivKey2, _ := blst.RandKey()
	blsPubKey2 := blsPrivKey2.PublicKey().Marshal()
	val2 := &types.Validator{Address: pubKey2.Address(), PubKey: pubKey2, BlsKey: blsPubKey2, VotingPower: 10}

	vals := []*types.Validator{
		val1, val2,
	}

	logger := log.TestingLogger()
	eventBus := types.NewEventBus()
	err := eventBus.Start()
	if err != nil {
		panic(err)
	}

	votePool, err := NewVotePool(logger, vals, eventBus)
	if err != nil {
		panic(err)
	}
	err = votePool.Start()
	if err != nil {
		panic(err)
	}

	return blsPrivKey1, val1, blsPrivKey2, val2, eventBus, votePool
}

func makeValidVotes(secKey blsCommon.SecretKey, val1 *types.Validator) (Vote, Vote, Vote) {
	eventHash1 := common.HexToHash("0xeefacfed87736ae1d8e8640f6fd7951862997782e5e79842557923e2779d5d5a").Bytes()
	sign1 := secKey.Sign(eventHash1).Marshal()
	vote1 := Vote{
		PubKey:    val1.BlsKey,
		Signature: sign1,
		EventType: FromBscCrossChainEvent,
		EventHash: eventHash1,
	}

	eventHash2 := common.HexToHash("0x7e19be15d0d524a1ca5e39be503d18584c23426920bdc23b159c37a2341913d0").Bytes()
	sign2 := secKey.Sign(eventHash2).Marshal()
	vote2 := Vote{
		PubKey:    val1.BlsKey,
		Signature: sign2,
		EventType: ToBscCrossChainEvent,
		EventHash: eventHash2,
	}

	eventHash3 := common.HexToHash("0xb941130c8d3508f642aba83db420f9cef6a6ebb7f869e3cef06f276bdcf205a9").Bytes()
	sign3 := secKey.Sign(eventHash3).Marshal()
	vote3 := Vote{
		PubKey:    val1.BlsKey,
		Signature: sign3,
		EventType: FromBscCrossChainEvent,
		EventHash: eventHash3,
	}
	return vote1, vote2, vote3
}

func TestPool_AddVote(t *testing.T) {
	pk1, val1, _, _, _, votePool := makeVotePool()

	eventHash := common.HexToHash("0xeefacfed87736ae1d8e8640f6fd7951862997782e5e79842557923e2779d5d5a").Bytes()
	secKey, _ := blst.SecretKeyFromBytes(pk1.Marshal())
	sign := secKey.Sign(eventHash).Marshal()

	anotherEventHash := common.HexToHash("0x7e19be15d0d524a1ca5e39be503d18584c23426920bdc23b159c37a2341913d0").Bytes()
	blsPrivKey, _ := blst.RandKey()
	blsPubKey := blsPrivKey.PublicKey().Marshal()
	blsSecKey, _ := blst.SecretKeyFromBytes(blsPrivKey.Marshal())
	anotherSign := blsSecKey.Sign(anotherEventHash).Marshal()

	testCases := []struct {
		vote Vote
		err  bool
		msg  string
	}{
		{
			vote: Vote{
				PubKey:    val1.BlsKey,
				Signature: sign,
				EventType: FromBscCrossChainEvent,
				EventHash: eventHash,
			},
			err: false,
			msg: "vote can be added",
		},
		{
			vote: Vote{
				PubKey:    val1.BlsKey,
				Signature: sign,
				EventType: FromBscCrossChainEvent,
				EventHash: eventHash,
			},
			err: false,
			msg: "vote can be re-added even it is not re-stored",
		},
		{
			vote: Vote{
				PubKey:    blsPubKey,
				Signature: anotherSign,
				EventType: FromBscCrossChainEvent,
				EventHash: anotherEventHash,
			},
			err: true,
			msg: "vote is not from validators",
		},
		{
			vote: Vote{
				PubKey:    val1.BlsKey,
				Signature: anotherSign,
				EventType: FromBscCrossChainEvent,
				EventHash: anotherEventHash,
			},
			err: true,
			msg: "invalid signature",
		},
	}

	for _, tc := range testCases {
		err := votePool.AddVote(&tc.vote)
		if tc.err {
			if assert.Error(t, err) {
				assert.Equal(t, tc.msg, err.Error())
			}
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestPool_QueryFlushVote(t *testing.T) {
	pk1, val1, _, _, _, votePool := makeVotePool()
	secKey, _ := blst.SecretKeyFromBytes(pk1.Marshal())

	vote1, vote2, vote3 := makeValidVotes(secKey, val1)

	err := votePool.AddVote(&vote1)
	require.NoError(t, err)
	err = votePool.AddVote(&vote2)
	require.NoError(t, err)
	err = votePool.AddVote(&vote3)
	require.NoError(t, err)

	result, err := votePool.GetVotesByEventType(FromBscCrossChainEvent)
	require.NoError(t, err)
	require.Equal(t, 2, len(result))
	result, err = votePool.GetVotesByEventType(ToBscCrossChainEvent)
	require.NoError(t, err)
	require.Equal(t, 1, len(result))

	result, err = votePool.GetVotesByEventTypeAndHash(vote1.EventType, vote1.EventHash)
	require.NoError(t, err)
	require.Equal(t, 1, len(result))
	require.Equal(t, vote1.EventHash, result[0].EventHash)

	// cannot find
	result, err = votePool.GetVotesByEventTypeAndHash(ToBscCrossChainEvent, vote1.EventHash)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))

	// flush
	votePool.FlushVotes()

	result, err = votePool.GetVotesByEventType(FromBscCrossChainEvent)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))
	result, err = votePool.GetVotesByEventType(ToBscCrossChainEvent)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))
}

func TestPool_PruneVote(t *testing.T) {
	pk1, val1, _, _, _, votePool := makeVotePool()
	secKey, _ := blst.SecretKeyFromBytes(pk1.Marshal())

	vote1, vote2, _ := makeValidVotes(secKey, val1)

	err := votePool.AddVote(&vote1)
	require.NoError(t, err)
	err = votePool.AddVote(&vote2)
	require.NoError(t, err)

	time.Sleep(voteKeepAliveAfter)
	time.Sleep(pruneVoteInterval)

	result, err := votePool.GetVotesByEventType(FromBscCrossChainEvent)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))
	result, err = votePool.GetVotesByEventType(ToBscCrossChainEvent)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))
}

func TestPool_SubscribeNewVoteEvent(t *testing.T) {
	pk1, val1, _, _, eventBus, votePool := makeVotePool()
	secKey, _ := blst.SecretKeyFromBytes(pk1.Marshal())

	vote1, _, _ := makeValidVotes(secKey, val1)

	sub, err := eventBus.Subscribe(context.Background(), "VotePoolUpdateSubscriber", eventVotePoolAdded, eventBusSubscribeCap)
	require.NoError(t, err)

	err = votePool.AddVote(&vote1)
	require.NoError(t, err)

	select {
	case msg := <-sub.Out():
		event, ok := msg.Data().(Vote)
		require.True(t, ok, "Expected event of type Vote, got %T", msg.Data())
		require.Equal(t, vote1.EventHash, event.EventHash)
	case <-sub.Cancelled():
		t.Fatalf("sub was cancelled (reason: %v)", sub.Err())
	case <-time.After(1 * time.Second):
		t.Fatal("Did not receive EventVotePoolUpdates within 1 sec.")
	}
}

func TestPool_ValidatorSetUpdate(t *testing.T) {
	pk1, val1, _, _, eventBus, votePool := makeVotePool()
	secKey, _ := blst.SecretKeyFromBytes(pk1.Marshal())

	// remove validator 1
	removeVal := types.Validator{PubKey: val1.PubKey, Address: val1.Address, VotingPower: 0}
	validatorUpdates := []*types.Validator{
		&removeVal,
	}
	validatorUpdateEvents := types.EventDataValidatorSetUpdates{
		ValidatorUpdates: validatorUpdates,
	}

	err := eventBus.PublishEventValidatorSetUpdates(validatorUpdateEvents)
	require.NoError(t, err)
	// resend the same validator updates should be fine
	for votePool.validatorVerifier.lenOfValidators() == 2 {
		err = eventBus.PublishEventValidatorSetUpdates(validatorUpdateEvents)
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond)
	}

	vote1, _, _ := makeValidVotes(secKey, val1)
	err = votePool.AddVote(&vote1)
	require.Error(t, err, "vote is not from validators")

	// add validator 1
	addVal := types.Validator{PubKey: val1.PubKey, Address: val1.Address, BlsKey: val1.BlsKey, VotingPower: 10}
	validatorUpdates = []*types.Validator{
		&addVal,
	}
	validatorUpdateEvents = types.EventDataValidatorSetUpdates{
		ValidatorUpdates: validatorUpdates,
	}

	err = eventBus.PublishEventValidatorSetUpdates(validatorUpdateEvents)
	require.NoError(t, err)
	// even resend the same validator updates should be fine
	for votePool.validatorVerifier.lenOfValidators() == 1 {
		err = eventBus.PublishEventValidatorSetUpdates(validatorUpdateEvents)
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond)
	}

	err = votePool.AddVote(&vote1)
	require.NoError(t, err)
}
