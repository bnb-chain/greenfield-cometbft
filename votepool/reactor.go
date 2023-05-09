package votepool

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	lru "github.com/hashicorp/golang-lru"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/cometbft/cometbft/proto/tendermint/votepool"
	"github.com/cometbft/cometbft/types"
)

const (
	// VotePoolChannel is the p2p channel used for sending and receiving votes in vote Pool.
	VotePoolChannel = byte(0x70)

	// Max number of kept vote histories from each peer, to avoiding broadcasting duplicated votes to a peer.
	maxVoteHistoryOfEachPeer = 256

	// After timeout current peer can broadcast votes to a remote peer even though the votes were received from it earlier.
	cacheTimeout = 3 * time.Second

	// Key for cache of votes from a peer.
	peerVoteCacheKey = "VotePoolReactor.voteCache"
)

var eventVotePoolAdded = types.QueryForEvent(eventBusVotePoolUpdates)

// Reactor will 1) subscribe votes from vote Pool and 2) broadcast votes to peers.
type Reactor struct {
	p2p.BaseReactor

	votePool VotePool
	eventBus *types.EventBus
}

// NewReactor returns a new Reactor with the given vote Pool.
func NewReactor(votePool VotePool, eventBus *types.EventBus) *Reactor {
	voteR := &Reactor{
		votePool: votePool,
		eventBus: eventBus,
	}
	voteR.BaseReactor = *p2p.NewBaseReactor("VotePoolReactor", voteR)
	return voteR
}

// OnStart implements Service.
func (voteR *Reactor) OnStart() error {
	if err := voteR.BaseReactor.OnStart(); err != nil {
		return err
	}
	if err := voteR.votePool.Start(); err != nil {
		return err
	}
	return nil
}

// OnStop implements Service.
func (voteR *Reactor) OnStop() {
	voteR.BaseReactor.OnStop()
	_ = voteR.votePool.Stop()
}

// SetLogger implements Service.
func (voteR *Reactor) SetLogger(l log.Logger) {
	voteR.Logger = l
}

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all local votes are forwarded to the remote peer.
func (voteR *Reactor) AddPeer(peer p2p.Peer) {
	cache, _ := lru.New(maxVoteHistoryOfEachPeer) // positive parameter will never return error
	peer.Set(peerVoteCacheKey, cache)
	go voteR.broadcastVotes(peer)
}

// RemovePeer implements Reactor.
func (voteR *Reactor) RemovePeer(peer p2p.Peer, reason interface{}) {
	if cache, ok := peer.Get(peerVoteCacheKey).(*lru.Cache); ok {
		cache.Purge()
	}

	peerID := peer.ID()
	err := voteR.eventBus.Unsubscribe(context.Background(), string(peerID), eventVotePoolAdded)
	if err != nil {
		voteR.Logger.Error("Cannot unsubscribe events", "peer", peerID, "event", eventVotePoolAdded)
	}
}

// GetChannels implements Reactor.
func (voteR *Reactor) GetChannels() []*conn.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  VotePoolChannel,
			Priority:            7,
			RecvMessageCapacity: 256, // size is bigger than Vote message
			MessageType:         &votepool.Message{},
		},
	}
}

// Receive implements Reactor.
func (voteR *Reactor) Receive(chID byte, peer p2p.Peer, msgBytes []byte) {
	msg := &votepool.Message{}
	err := proto.Unmarshal(msgBytes, msg)
	if err != nil {
		panic(err)
	}
	uw, err := msg.Unwrap()
	if err != nil {
		panic(err)
	}
	voteR.ReceiveEnvelope(p2p.Envelope{
		ChannelID: chID,
		Src:       peer,
		Message:   uw,
	})
}

func (voteR *Reactor) ReceiveEnvelope(e p2p.Envelope) {
	switch msg := e.Message.(type) {
	case *votepool.Vote:
		vote := NewVote(msg.PubKey, msg.Signature, uint8(msg.EventType), msg.EventHash)
		voteR.Logger.Debug("Receive vote", "vote", vote.Key(), "src", e.Src)
		if err := voteR.votePool.AddVote(vote); err != nil {
			voteR.Logger.Info("Could not add vote", "vote", vote.Key(), "err", err)
		} else {
			if cache, ok := e.Src.Get(peerVoteCacheKey).(*lru.Cache); ok {
				// keep track of votes from the remote peer, update timestamp
				cache.Add(vote.Key(), time.Now())
			}
		}
	default:
		voteR.Logger.Error("Unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		voteR.Switch.StopPeerForError(e.Src, fmt.Errorf("votepool cannot handle message of type: %T", e.Message))
		return
	}
}

// broadcastVotes routine will broadcast votes to peers.
func (voteR *Reactor) broadcastVotes(peer p2p.Peer) {
	if !voteR.IsRunning() || !peer.IsRunning() {
		return
	}
	sub, err := voteR.eventBus.Subscribe(context.Background(), string(peer.ID()), eventVotePoolAdded, eventBusSubscribeCap)
	if err != nil {
		voteR.Logger.Error("Cannot subscribe to vote update event", "err", err.Error())
		return
	}
	cache, ok := peer.Get(peerVoteCacheKey).(*lru.Cache)
	if !ok { // this should not happen
		voteR.Logger.Error(fmt.Sprintf("Peer %v has no cache state", peer))
		return
	}
	for {
		select {
		case voteData := <-sub.Out():
			vote := voteData.Data().(Vote)
			// send votes to remote peer, if
			// 1) it did not receive the vote from the remote peer, or,
			// 2) the vote is received earlier than `time.Now() - cacheTimeout`
			needToSend := true
			value, existed := cache.Get(vote.Key())
			if existed {
				needToSend = false
				t, ok := value.(time.Time)
				if ok && time.Now().After(t.Add(cacheTimeout)) {
					needToSend = true
				}
			}
			if needToSend {
				_ = peer.SendEnvelope(p2p.Envelope{
					ChannelID: VotePoolChannel,
					Message: &votepool.Vote{
						PubKey:    vote.PubKey,
						Signature: vote.Signature,
						EventType: uint32(vote.EventType),
						EventHash: vote.EventHash,
					},
				})
				voteR.Logger.Debug("Sent vote to", "peer", peer, "vote", vote.Key())
			}
		case <-sub.Cancelled():
			return
		case <-voteR.Quit():
			return
		case <-peer.Quit():
			return
		}
	}
}
