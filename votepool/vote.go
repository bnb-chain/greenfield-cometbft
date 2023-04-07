package votepool

import (
	"errors"
	"time"
)

const (
	// Length of Vote event hash
	eventHashLen = 32

	// Length of Vote public key
	pubKeyLen = 48

	// Length of Vote signature
	signatureLen = 96
)

// Vote stands for votes from differently relayers/validators, to agree/disagree on something (e.g., cross chain).
type Vote struct {
	PubKey    []byte    `json:"pub_key"`    //bls public key
	Signature []byte    `json:"signature"`  //bls signature
	EventType EventType `json:"event_type"` //event type of the vote
	EventHash []byte    `json:"event_hash"` //the data, here []byte is used, so that Vote will not care about the meaning of the data

	expireAt time.Time
}

func NewVote(pubKey, signature []byte, eventType uint8, eventHash []byte) *Vote {
	vote := Vote{
		PubKey:    pubKey,
		Signature: signature,
		EventType: EventType(eventType),
		EventHash: eventHash,
	}
	return &vote
}

// Key is used as an identifier of a vote, it is usually used as the key of map of cache.
func (v *Vote) Key() string {
	return string(v.EventHash[:]) + string(v.PubKey[:])
}

// ValidateBasic does basic validation of vote.
func (v *Vote) ValidateBasic() error {
	if len(v.EventHash) != eventHashLen {
		return errors.New("invalid event hash")
	}
	if v.EventType != ToBscCrossChainEvent &&
		v.EventType != FromBscCrossChainEvent &&
		v.EventType != DataAvailabilityChallengeEvent {
		return errors.New("invalid event type")
	}
	if len(v.PubKey) != pubKeyLen {
		return errors.New("invalid public key")
	}
	if len(v.Signature) != signatureLen {
		return errors.New("invalid signature")
	}
	return nil
}
