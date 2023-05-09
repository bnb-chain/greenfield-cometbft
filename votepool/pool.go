package votepool

import (
	"context"
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/cometbft/cometbft/libs/service"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/types"
)

const (

	// The number of cached votes (i.e., keys) to quickly filter out when adding votes.
	cacheVoteSize = 1024

	// Vote will be assigned the expired at time when adding to the Pool.
	voteKeepAliveAfter = time.Second * 30

	// Votes in the Pool will be pruned periodically to remove useless ones.
	pruneVoteInterval = 3 * time.Second

	// Defines the channel size for event bus subscription.
	eventBusSubscribeCap = 1024

	// The event type of adding new votes to the Pool successfully.
	eventBusVotePoolUpdates = "votePoolUpdates"
)

// voteStore stores one type of votes.
type voteStore struct {
	mtx     *sync.RWMutex               // mutex for concurrency access of voteMap and others
	voteMap map[string]map[string]*Vote // map: eventHash -> pubKey -> Vote

	queue *VoteQueue // priority queue for prune votes
}

// newVoteStore creates a store to store votes.
func newVoteStore() *voteStore {
	s := &voteStore{
		mtx:     &sync.RWMutex{},
		voteMap: make(map[string]map[string]*Vote),
		queue:   NewVoteQueue(),
	}
	return s
}

// addVote will add a vote to the store.
// Be noted: no validation is conducted in this layer.
func (s *voteStore) addVote(vote *Vote) {
	eventHashStr := string(vote.EventHash[:])
	pubKeyStr := string(vote.PubKey[:])
	s.mtx.Lock()
	defer s.mtx.Unlock()

	subM, ok := s.voteMap[eventHashStr]
	if !ok {
		subM = make(map[string]*Vote)
		s.voteMap[eventHashStr] = subM
	}
	subM[pubKeyStr] = vote
	s.queue.Insert(vote)
}

// getVotesByEventHash will query events by event hash.
func (s *voteStore) getVotesByEventHash(eventHash []byte) []*Vote {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	votes := make([]*Vote, 0)
	if subM, ok := s.voteMap[string(eventHash[:])]; ok {
		for _, v := range subM {
			votes = append(votes, v)
		}
	}
	return votes
}

// getAllVotes will return all votes in the store.
func (s *voteStore) getAllVotes() []*Vote {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	votes := make([]*Vote, 0)
	for _, subM := range s.voteMap {
		for _, v := range subM {
			votes = append(votes, v)
		}
	}
	return votes
}

// flushVotes will clear all votes in the store.
func (s *voteStore) flushVotes() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.voteMap = make(map[string]map[string]*Vote)
	s.queue = NewVoteQueue()
}

// pruneVotes will prune votes which are expired and return the pruned votes' keys.
func (s *voteStore) pruneVotes() []string {
	keys := make([]string, 0)
	current := &Vote{expireAt: time.Now()}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if expires, err := s.queue.PopUntil(current); err == nil {
		for _, expire := range expires {
			keys = append(keys, expire.Key())
			delete(s.voteMap[string(expire.EventHash[:])], string(expire.PubKey[:]))
		}
	}
	return keys
}

// Pool implements VotePool to store different types of votes.
// Meanwhile, it will check the signature and source signer of a vote, only votes from validators will be saved.
type Pool struct {
	service.BaseService

	stores map[EventType]*voteStore // each event type will have a store
	ticker *time.Ticker             // prune ticker

	blsVerifier       *BlsSignatureVerifier  // verify a vote's signature
	validatorVerifier *FromValidatorVerifier // verify a vote is from a validator

	cache *lru.Cache // to cache recent added votes' keys

	eventBus *types.EventBus // to subscribe validator update events and publish new added vote events
}

// NewVotePool creates a Pool. The initial validators should be supplied.
func NewVotePool(logger log.Logger, validators []*types.Validator, eventBus *types.EventBus) *Pool {
	eventTypes := []EventType{ToBscCrossChainEvent, FromBscCrossChainEvent, DataAvailabilityChallengeEvent}

	ticker := time.NewTicker(pruneVoteInterval)
	stores := make(map[EventType]*voteStore, len(eventTypes))
	for _, et := range eventTypes {
		store := newVoteStore()
		stores[et] = store
	}

	cache, _ := lru.New(cacheVoteSize) // positive parameter will never return error

	// set the initial validators
	validatorVerifier := NewFromValidatorVerifier()
	validatorVerifier.initValidators(validators)
	votePool := &Pool{
		stores:            stores,
		ticker:            ticker,
		cache:             cache,
		eventBus:          eventBus,
		blsVerifier:       &BlsSignatureVerifier{},
		validatorVerifier: validatorVerifier,
	}
	votePool.BaseService = *service.NewBaseService(logger, "VotePool", votePool)

	return votePool
}

// OnStart implements Service.
func (p *Pool) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}
	go p.validatorUpdateRoutine()
	go p.pruneVoteRoutine()
	return nil
}

// OnStop implements Service.
func (p *Pool) OnStop() {
	p.BaseService.OnStop()
	p.ticker.Stop()
}

// AddVote implements VotePool.
func (p *Pool) AddVote(vote *Vote) error {
	err := vote.ValidateBasic()
	if err != nil {
		return err
	}
	store, ok := p.stores[vote.EventType]
	if !ok {
		return errors.New("unsupported event type")
	}

	if ok = p.cache.Contains(vote.Key()); ok {
		return nil
	}

	if err = p.validatorVerifier.Validate(vote); err != nil {
		return err
	}
	if err = p.blsVerifier.Validate(vote); err != nil {
		return err
	}

	vote.expireAt = time.Now().Add(voteKeepAliveAfter)
	store.addVote(vote)

	if err = p.eventBus.Publish(eventBusVotePoolUpdates, *vote); err != nil {
		p.Logger.Error("Cannot publish vote pool event", "err", err.Error())
	}
	p.cache.Add(vote.Key(), struct{}{})
	return nil
}

// GetVotesByEventTypeAndHash implements VotePool.
func (p *Pool) GetVotesByEventTypeAndHash(eventType EventType, eventHash []byte) ([]*Vote, error) {
	store, ok := p.stores[eventType]
	if !ok {
		return nil, errors.New("unsupported event type")
	}
	return store.getVotesByEventHash(eventHash), nil
}

// GetVotesByEventType implements VotePool.
func (p *Pool) GetVotesByEventType(eventType EventType) ([]*Vote, error) {
	store, ok := p.stores[eventType]
	if !ok {
		return nil, errors.New("unsupported event type")
	}
	return store.getAllVotes(), nil
}

// FlushVotes implements VotePool.
func (p *Pool) FlushVotes() {
	for _, store := range p.stores {
		store.flushVotes()
	}
	p.cache.Purge()
}

// validatorUpdateRoutine will sync validator updates.
func (p *Pool) validatorUpdateRoutine() {
	if !p.IsRunning() || !p.eventBus.IsRunning() {
		return
	}
	sub, err := p.eventBus.Subscribe(context.Background(), "VotePoolService", types.EventQueryValidatorSetUpdates, eventBusSubscribeCap)
	if err != nil {
		p.Logger.Error("Cannot subscribe to validator set update event", "err", err.Error())
		return
	}
	for {
		select {
		case validatorData := <-sub.Out():
			changes := validatorData.Data().(types.EventDataValidatorSetUpdates)
			p.validatorVerifier.updateValidators(changes.ValidatorUpdates)
			p.Logger.Info("Validators updated", "changes", changes.ValidatorUpdates)
		case <-sub.Cancelled():
			return
		case <-p.Quit():
			return
		}
	}
}

// pruneVoteRoutine will prune votes at the given intervals.
func (p *Pool) pruneVoteRoutine() {
	for range p.ticker.C {
		for _, s := range p.stores {
			keys := s.pruneVotes()
			for _, key := range keys {
				p.cache.Remove(key)
			}
		}
	}
}
