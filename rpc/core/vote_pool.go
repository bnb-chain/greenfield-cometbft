package core

import (
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/votepool"
)

func BroadcastVote(ctx *rpctypes.Context, vote votepool.Vote) (*ctypes.ResultBroadcastVote, error) {
	err := env.VotePool.AddVote(&vote)
	return &ctypes.ResultBroadcastVote{}, err
}

func QueryVote(ctx *rpctypes.Context, eventType int, eventHash []byte) (*ctypes.ResultQueryVote, error) {
	var votes []*votepool.Vote
	var err error
	if len(eventHash) == 0 {
		votes, err = env.VotePool.GetVotesByEventType(votepool.EventType(eventType))
	} else {
		votes, err = env.VotePool.GetVotesByEventTypeAndHash(votepool.EventType(eventType), eventHash)
	}

	return &ctypes.ResultQueryVote{Votes: votes}, err
}

func UnsafeFlushVotePool(ctx *rpctypes.Context) (*ctypes.ResultFlushVote, error) {
	env.VotePool.FlushVotes()
	return &ctypes.ResultFlushVote{}, nil
}
