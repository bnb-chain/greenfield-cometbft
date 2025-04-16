package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto"
	ce "github.com/cometbft/cometbft/crypto/encoding"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

const (
	// BlsPubKeySize is the size of validator's relayer bls public key.
	BlsPubKeySize = 48

	// AddressSize is the size of validator's relayer address.
	AddressSize = 20
)

// Volatile state for each Validator
// NOTE: The ProposerPriority is not included in Validator.Hash();
// make sure to update that method if changes are made here
type Validator struct {
	Address     Address       `json:"address"`
	PubKey      crypto.PubKey `json:"pub_key"`
	VotingPower int64         `json:"voting_power"`

	ProposerPriority int64 `json:"proposer_priority"`

	BlsKey         []byte `json:"bls_key"`         // bls public key
	RelayerAddress []byte `json:"relayer_address"` // address of authorized relayer/operator
}

// NewValidator returns a new validator with the given pubkey and voting power.
func NewValidator(pubKey crypto.PubKey, votingPower int64) *Validator {
	return &Validator{
		Address:          pubKey.Address(),
		PubKey:           pubKey,
		VotingPower:      votingPower,
		ProposerPriority: 0,
	}
}

// ValidateBasic performs basic validation.
func (v *Validator) ValidateBasic() error {
	if v == nil {
		return errors.New("nil validator")
	}
	if v.PubKey == nil {
		return errors.New("validator does not have a public key")
	}

	if v.VotingPower < 0 {
		return errors.New("validator has negative voting power")
	}

	addr := v.PubKey.Address()
	if !bytes.Equal(v.Address, addr) {
		return fmt.Errorf("validator address is incorrectly derived from pubkey. Exp: %v, got %v", addr, v.Address)
	}

	if len(v.BlsKey) != 0 && len(v.BlsKey) != BlsPubKeySize {
		return fmt.Errorf("validator relayer bls key is the wrong size: %v", v.BlsKey)
	}
	if len(v.RelayerAddress) != 0 && len(v.RelayerAddress) != AddressSize {
		return fmt.Errorf("validator relayer address is the wrong size: %v", v.RelayerAddress)
	}

	return nil
}

// Creates a new copy of the validator so we can mutate ProposerPriority.
// Panics if the validator is nil.
func (v *Validator) Copy() *Validator {
	vCopy := *v
	return &vCopy
}

// Returns the one with higher ProposerPriority.
func (v *Validator) CompareProposerPriority(other *Validator) *Validator {
	if v == nil {
		return other
	}
	switch {
	case v.ProposerPriority > other.ProposerPriority:
		return v
	case v.ProposerPriority < other.ProposerPriority:
		return other
	default:
		result := bytes.Compare(v.Address, other.Address)
		switch {
		case result < 0:
			return v
		case result > 0:
			return other
		default:
			panic("Cannot compare identical validators")
		}
	}
}

// String returns a string representation of String.
//
// 1. address
// 2. public key
// 3. voting power
// 4. proposer priority
// 5. relayer address
func (v *Validator) String() string {
	if v == nil {
		return "nil-Validator"
	}
	return fmt.Sprintf("Validator{%v %v VP:%v A:%v R:%v}",
		v.Address,
		v.PubKey,
		v.VotingPower,
		v.ProposerPriority,
		v.RelayerAddress)
}

// ValidatorListString returns a prettified validator list for logging purposes.
func ValidatorListString(vals []*Validator) string {
	chunks := make([]string, len(vals))
	for i, val := range vals {
		chunks[i] = fmt.Sprintf("%s:%d", val.Address, val.VotingPower)
	}

	return strings.Join(chunks, ",")
}

// Bytes computes the unique encoding of a validator with a given voting power.
// These are the bytes that gets hashed in consensus. It excludes address
// as its redundant with the pubkey. This also excludes ProposerPriority
// which changes every round.
func (v *Validator) Bytes() []byte {
	pk, err := ce.PubKeyToProto(v.PubKey)
	if err != nil {
		panic(err)
	}

	pbv := cmtproto.SimpleValidator{
		PubKey:         &pk,
		VotingPower:    v.VotingPower,
		BlsKey:         v.BlsKey,
		RelayerAddress: v.RelayerAddress,
	}

	bz, err := pbv.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

// SetBlsKey will update the bls public key of relayer.
func (v *Validator) SetBlsKey(blsKey []byte) {
	v.BlsKey = blsKey
}

// SetRelayerAddress will update the relayer address of validator.
func (v *Validator) SetRelayerAddress(address []byte) {
	v.RelayerAddress = address
}

// ToProto converts Validator to protobuf
func (v *Validator) ToProto() (*cmtproto.Validator, error) {
	if v == nil {
		return nil, errors.New("nil validator")
	}

	pk, err := ce.PubKeyToProto(v.PubKey)
	if err != nil {
		return nil, err
	}

	vp := cmtproto.Validator{
		Address:          v.Address,
		PubKey:           pk,
		VotingPower:      v.VotingPower,
		ProposerPriority: v.ProposerPriority,
		BlsKey:           v.BlsKey,
		RelayerAddress:   v.RelayerAddress,
	}

	return &vp, nil
}

// FromProto sets a protobuf Validator to the given pointer.
// It returns an error if the public key is invalid.
func ValidatorFromProto(vp *cmtproto.Validator) (*Validator, error) {
	if vp == nil {
		return nil, errors.New("nil validator")
	}

	pk, err := ce.PubKeyFromProto(vp.PubKey)
	if err != nil {
		return nil, err
	}
	v := new(Validator)
	v.Address = vp.GetAddress()
	v.PubKey = pk
	v.VotingPower = vp.GetVotingPower()
	v.ProposerPriority = vp.GetProposerPriority()
	v.BlsKey = vp.GetBlsKey()
	v.RelayerAddress = vp.GetRelayerAddress()
	return v, nil
}

//----------------------------------------
// RandValidator

// RandValidator returns a randomized validator, useful for testing.
// UNSTABLE
func RandValidator(randPower bool, minPower int64) (*Validator, PrivValidator) {
	privVal := NewMockPV()
	votePower := minPower
	if randPower {
		votePower += int64(cmtrand.Uint32())
	}
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(fmt.Errorf("could not retrieve pubkey %w", err))
	}
	val := NewValidator(pubKey, votePower)
	return val, privVal
}
