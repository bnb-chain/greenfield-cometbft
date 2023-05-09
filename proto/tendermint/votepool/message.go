package votepool

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/p2p"
)

var _ p2p.Wrapper = &Vote{}
var _ p2p.Unwrapper = &Message{}

// Wrap implements the p2p Wrapper interface and wraps a votepool message.
func (m *Vote) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_Vote{Vote: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped votepool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_Vote:
		return m.GetVote(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
