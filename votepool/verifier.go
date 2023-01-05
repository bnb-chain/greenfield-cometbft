package votepool

import (
	"errors"

	"github.com/prysmaticlabs/prysm/crypto/bls/blst"

	"github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/types"
)

// Verifier will validate Votes by different policies.
type Verifier interface {
	Validate(vote *Vote) error
}

// FromValidatorVerifier will check whether the Vote is from a valid validator.
type FromValidatorVerifier struct {
	mtx        *sync.RWMutex
	validators map[string]*types.Validator
}

func NewFromValidatorVerifier() *FromValidatorVerifier {
	f := &FromValidatorVerifier{
		validators: make(map[string]*types.Validator),
		mtx:        &sync.RWMutex{},
	}
	return f
}

func (f *FromValidatorVerifier) initValidators(validators []*types.Validator) {
	for _, val := range validators {
		if len(val.RelayerBlsKey) > 0 {
			f.validators[string(val.RelayerBlsKey[:])] = val
		}
	}
}

func (f *FromValidatorVerifier) updateValidators(changes []*types.Validator) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	vals := make([]*types.Validator, 0)
	for _, val := range f.validators {
		vals = append(vals, val)
	}
	f.validators = make(map[string]*types.Validator)
	valSet := &types.ValidatorSet{Validators: vals}
	_ = valSet.UpdateWithChangeSet(changes) // use valSet's validators even if there are errors
	for _, val := range valSet.Validators {
		if len(val.RelayerBlsKey) > 0 {
			f.validators[string(val.RelayerBlsKey[:])] = val
		}
	}
}

func (f *FromValidatorVerifier) lenOfValidators() int {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	return len(f.validators)
}

// Validate implements Verifier.
func (f *FromValidatorVerifier) Validate(vote *Vote) error {
	f.mtx.RLock()
	defer f.mtx.RUnlock()

	if _, ok := f.validators[string(vote.PubKey[:])]; ok {
		return nil
	}
	return errors.New("vote is not from validators")
}

// BlsSignatureVerifier will check whether the Vote is correctly bls signed.
type BlsSignatureVerifier struct {
}

// Validate implements Verifier.
func (b *BlsSignatureVerifier) Validate(vote *Vote) error {
	valid := verifySignature(vote.EventHash, vote.PubKey, vote.Signature)
	if !valid {
		return errors.New("invalid signature")
	}
	return nil
}

func verifySignature(msg []byte, pubKey, sig []byte) bool {
	blsPubKey, err := blst.PublicKeyFromBytes(pubKey)
	if err != nil {
		return false
	}
	signature, err := blst.SignatureFromBytes(sig)
	if err != nil {
		return false
	}
	return signature.Verify(blsPubKey, msg)
}
