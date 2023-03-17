package votepool

// EventType defines the types for voting.
type EventType uint8

const (
	// ToBscCrossChainEvent defines the type of cross chain events from the current chain to BSC.
	ToBscCrossChainEvent EventType = 1

	// FromBscCrossChainEvent defines the type of cross chain events from BSC to the current chain.
	FromBscCrossChainEvent EventType = 2

	// DataAvailabilityChallengeEvent defines the type of events for data availability challenges.
	DataAvailabilityChallengeEvent EventType = 3
)
