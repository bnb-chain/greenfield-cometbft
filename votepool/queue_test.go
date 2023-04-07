package votepool

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func makeVotes() (*Vote, *Vote, *Vote) {
	now := time.Now()

	vote1 := &Vote{
		EventHash: common.HexToHash("0xf7ded74d86f9cf164e6e1a1f2d5fb2429140bb6d701a39bd2c416d36c57100e5").Bytes(),
		expireAt:  now.Add(1 * time.Hour),
	}
	vote2 := &Vote{
		EventHash: common.HexToHash("0xd6ab0606cfd6b517656a9b60dc127069a2fc27146946a872d3e18c87164fe2ba").Bytes(),
		expireAt:  now.Add(-1 * time.Hour),
	}
	vote3 := &Vote{
		EventHash: common.HexToHash("0xb9947bea2b0c2dc936df248397638462769601e4d5c0b48975731b48c206507e").Bytes(),
		expireAt:  now.Add(2 * time.Hour),
	}
	return vote1, vote2, vote3
}

func TestVoteQueuePop(t *testing.T) {
	vote1, vote2, vote3 := makeVotes()

	q := NewVoteQueue()
	q.Insert(vote1)
	q.Insert(vote2)
	q.Insert(vote3)

	pop, err := q.Pop()
	require.NoError(t, err)
	require.Equal(t, vote2.EventHash, pop.EventHash)

	pop, err = q.Pop()
	require.NoError(t, err)
	require.Equal(t, vote1.EventHash, pop.EventHash)

	pop, err = q.Pop()
	require.NoError(t, err)
	require.Equal(t, vote3.EventHash, pop.EventHash)

	_, err = q.Pop()
	require.Error(t, err)
}

func TestVoteQueuePopUntil(t *testing.T) {
	vote1, vote2, vote3 := makeVotes()

	q := NewVoteQueue()
	q.Insert(vote1)
	q.Insert(vote2)
	q.Insert(vote3)

	current := &Vote{expireAt: time.Now()}
	expires, err := q.PopUntil(current)
	require.NoError(t, err)
	require.Equal(t, 1, len(expires))
	require.Equal(t, vote2.EventHash, expires[0].EventHash)
}
