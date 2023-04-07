package votepool

import (
	"container/heap"
	"errors"
)

// VoteQueue represents a priority queue for Votes. The expiredAt field of a Vote will be used as priority.
type VoteQueue struct {
	itemHeap *itemHeap
}

// NewVoteQueue initializes an empty priority queue.
func NewVoteQueue() *VoteQueue {
	h := newItemHeap()
	return &VoteQueue{
		itemHeap: &h,
	}
}

// Len returns the number of elements in the queue.
func (p *VoteQueue) Len() int {
	return p.itemHeap.Len()
}

// Insert inserts a new Vote into the queue.
func (p *VoteQueue) Insert(vote *Vote) {
	newItem := &voteItem{
		vote: vote,
	}
	heap.Push(p.itemHeap, newItem)
}

// Pop removes the Vote with the highest priority from the queue and returns it.
// In case of an empty queue, an error is returned.
func (p *VoteQueue) Pop() (*Vote, error) {
	if len(*p.itemHeap) == 0 {
		return nil, errors.New("empty queue")
	}

	item := heap.Pop(p.itemHeap).(*voteItem)
	return item.vote, nil
}

// PopUntil removes the Votes, which have higher priority than the passed one.
// In case of an empty queue, an error is returned.
func (p *VoteQueue) PopUntil(vote *Vote) ([]*Vote, error) {
	if len(*p.itemHeap) == 0 {
		return nil, errors.New("empty queue")
	}

	votes := make([]*Vote, 0)
	for p.Len() > 0 {
		top := heap.Pop(p.itemHeap).(*voteItem)
		if top.vote.expireAt.After(vote.expireAt) {
			heap.Push(p.itemHeap, top)
			break
		}
		votes = append(votes, top.vote)
	}
	return votes, nil
}

type itemHeap []*voteItem

func newItemHeap() itemHeap {
	return make([]*voteItem, 0)
}

type voteItem struct {
	vote  *Vote
	index int
}

func (ih *itemHeap) Len() int {
	return len(*ih)
}

func (ih *itemHeap) Less(i, j int) bool {
	return (*ih)[i].vote.expireAt.Before((*ih)[j].vote.expireAt)
}

func (ih *itemHeap) Swap(i, j int) {
	index := ih.Len() - 1
	if i > index || j > index || i == j {
		return
	}
	(*ih)[i], (*ih)[j] = (*ih)[j], (*ih)[i]
	(*ih)[i].index = i
	(*ih)[j].index = j
}

func (ih *itemHeap) Push(x interface{}) {
	it := x.(*voteItem)
	it.index = len(*ih)
	*ih = append(*ih, it)
}

func (ih *itemHeap) Pop() interface{} {
	old := *ih
	item := old[len(old)-1]
	*ih = old[0 : len(old)-1]
	return item
}
