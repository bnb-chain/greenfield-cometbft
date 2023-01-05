package votepool

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-kit/log/term"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
	blsCommon "github.com/prysmaticlabs/prysm/crypto/bls/common"
	"github.com/stretchr/testify/require"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
)

const testEventType = FromBscCrossChainEvent

// votepoolLogger is a TestingLogger which uses a different
// color for each validator ("validator" key must exist).
func votepoolLogger() log.Logger {
	return log.TestingLoggerWithColorFn(func(keyvals ...interface{}) term.FgBgColor {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] == "validator" {
				return term.FgBgColor{Fg: term.Color(uint8(keyvals[i+1].(int) + 1))}
			}
		}
		return term.FgBgColor{}
	})
}

func makeAndConnectReactors(config *cfg.Config, n int) ([]blsCommon.SecretKey, []*types.Validator, []*types.EventBus, []VotePool, []*Reactor) {
	pubKey1 := ed25519.GenPrivKey().PubKey()
	blsPrivKey1, _ := blst.RandKey()
	blsPubKey1 := blsPrivKey1.PublicKey().Marshal()
	val1 := &types.Validator{Address: pubKey1.Address(), PubKey: pubKey1, RelayerBlsKey: blsPubKey1, VotingPower: 10}

	pubKey2 := ed25519.GenPrivKey().PubKey()
	blsPrivKey2, _ := blst.RandKey()
	blsPubKey2 := blsPrivKey2.PublicKey().Marshal()
	val2 := &types.Validator{Address: pubKey2.Address(), PubKey: pubKey2, RelayerBlsKey: blsPubKey2, VotingPower: 10}

	pks := []blsCommon.SecretKey{
		blsPrivKey1, blsPrivKey2,
	}

	vals := []*types.Validator{
		val1, val2,
	}

	eventBuses := make([]*types.EventBus, n)
	votePools := make([]VotePool, n)
	reactors := make([]*Reactor, n)

	logger := votepoolLogger()
	for i := 0; i < n; i++ {
		eventBus := types.NewEventBus()
		err := eventBus.Start()
		if err != nil {
			panic(err)
		}

		votePool, err := NewVotePool(logger, vals, eventBus)
		if err != nil {
			panic(err)
		}

		eventBuses[i] = eventBus
		votePools[i] = votePool
		reactors[i] = NewReactor(votePool, eventBus)
		reactors[i].SetLogger(logger.With("validator", i))
	}

	p2p.MakeConnectedSwitches(config.P2P, n, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("VOTEPOOL", reactors[i])
		return s

	}, p2p.Connect2Switches)
	return pks, vals, eventBuses, votePools, reactors
}

func TestReactorBroadcastVotes(t *testing.T) {
	config := cfg.TestConfig()
	pks, vals, _, pools, reactors := makeAndConnectReactors(config, 2)

	secKey, _ := blst.SecretKeyFromBytes(pks[0].Marshal())
	eventHash1 := common.HexToHash("0xeefacfed87736ae1d8e8640f6fd7951862997782e5e79842557923e2779d5d5a").Bytes()
	sign1 := secKey.Sign(eventHash1).Marshal()
	vote1 := Vote{
		PubKey:    vals[0].RelayerBlsKey,
		Signature: sign1,
		EventType: testEventType,
		EventHash: eventHash1,
	}
	err := pools[0].AddVote(&vote1)
	require.NoError(t, err)

	waitVotesReceived(t, reactors, eventHash1)

	eventHash2 := common.HexToHash("0x7e19be15d0d524a1ca5e39be503d18584c23426920bdc23b159c37a2341913d0").Bytes()
	sign2 := secKey.Sign(eventHash2).Marshal()
	vote2 := Vote{
		PubKey:    vals[0].RelayerBlsKey,
		Signature: sign2,
		EventType: testEventType,
		EventHash: eventHash2,
	}
	err = pools[0].AddVote(&vote2)
	require.NoError(t, err)

	waitVotesReceived(t, reactors, eventHash2)
}

func waitVotesReceived(t *testing.T, reactors []*Reactor, eventHash []byte) {
	wg := new(sync.WaitGroup)
	for i, reactor := range reactors {
		wg.Add(1)
		go func(r *Reactor, reactorIndex int) {
			defer wg.Done()
			waitForVoteOnReactor(t, eventHash, r)
		}(reactor, i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	timer := time.After(20 * time.Second)
	select {
	case <-timer:
		t.Fatal("Timed out waiting for vote")
	case <-done:
	}
}

func waitForVoteOnReactor(t *testing.T, eventHash []byte, r *Reactor) {
	for {
		time.Sleep(time.Millisecond * 100)
		votes, _ := r.votePool.GetVotesByEventType(testEventType)
		found := false
		for _, vote := range votes {
			if bytes.Equal(eventHash, vote.EventHash) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
}
