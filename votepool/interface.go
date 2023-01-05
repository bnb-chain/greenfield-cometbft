package votepool

import (
	"github.com/tendermint/tendermint/libs/service"
)

// VotePool is used for pooling cross chain, challenge votes from different validators/relayers.
// Votes in the VotePool will be pruned based on Vote's expired time.
type VotePool interface {
	service.Service

	// AddVote will add a vote to the Pool. Different validations can be conducted before adding.
	AddVote(vote *Vote) error

	// GetVotesByEventTypeAndHash will filter votes by event hash and event type.
	GetVotesByEventTypeAndHash(eventType EventType, eventHash []byte) ([]*Vote, error)

	// GetVotesByEventType will filter votes by event type.
	GetVotesByEventType(eventType EventType) ([]*Vote, error)

	// FlushVotes will clear all votes in the Pool, no matter what types of events.
	FlushVotes()
}
