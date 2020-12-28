package types

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

// Staking params default values
const (
	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	// TODO: Justify our choice of default here.
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 3

	// Default maximum number of bonded validators
	DefaultMaxValidators uint16 = 35

	// Default maximum entries in a UBD/RED pair
	DefaultMaxEntries uint16 = 7

	DefaultMaxKeyNodes uint16 = 15

	DefaultMaxCandidateKeyNodeHeartbeatInterval uint64 = 100
)

var (
	//Default min validator delegation, 100K HBC, HBC's precesion is 1^10-18.
	DefaultMinValidatorDelegation = sdk.NewIntWithDecimal(10, 22)

	// default min keynode delegation, 500K HBC
	DefaultMinKeyNodeDelegation = sdk.NewIntWithDecimal(50, 22)
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime                        = []byte("UnbondingTime")
	KeyMaxValidators                        = []byte("MaxValidators")
	KeyMaxEntries                           = []byte("KeyMaxEntries")
	KeyBondDenom                            = []byte("BondDenom")
	KeyMaxKeyNodes                          = []byte("MaxKeyNodes")
	KeyMinValidatorDelegation               = []byte("MinValidatorDelegation")
	KeyMinKeyNodeDelegation                 = []byte("MinKeyNodeDelegation")
	KeyMaxCandidateKeyNodeHeartbeatInterval = []byte("MaxCandidateKeyNodeHeartbeatInterval")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	UnbondingTime time.Duration `json:"unbonding_time" yaml:"unbonding_time"` // time duration of unbonding
	MaxValidators uint16        `json:"max_validators" yaml:"max_validators"` // maximum number of validators (max uint16 = 65535)
	MaxEntries    uint16        `json:"max_entries" yaml:"max_entries"`       // max entries for either unbonding delegation or redelegation (per pair/trio)
	// note: we need to be a bit careful about potential overflow here, since this is user-determined
	BondDenom                            string  `json:"bond_denom" yaml:"bond_denom"`       // bondable coin denomination
	MaxKeyNodes                          uint16  `json:"max_key_nodes" yaml:"max_key_nodes"` // maximum number of keynodes
	MinValidatorDelegation               sdk.Int `json:"min_validator_delegation" yaml:"min_validator_delegation"`
	MinKeyNodeDelegation                 sdk.Int `json:"min_key_node_delegation" yaml:"min_key_node_delegation"`
	MaxCandidateKeyNodeHeartbeatInterval uint64  `json:"max_candidate_key_node_heartbeat_interval" yaml:"max_candidate_key_node_heartbeat_interval"`
}

// NewParams creates a new Params instance
func NewParams(unbondingTime time.Duration, maxValidators, maxKeyNodes, maxEntries uint16, bondDenom string,
	minValidatorDelegation, minKeyNodeDelegation sdk.Int, maxCandidateKeyNodeHeartbeatInterval uint64) Params {
	return Params{
		UnbondingTime:                        unbondingTime,
		MaxValidators:                        maxValidators,
		MaxKeyNodes:                          maxKeyNodes,
		MaxEntries:                           maxEntries,
		BondDenom:                            bondDenom,
		MinValidatorDelegation:               minValidatorDelegation,
		MinKeyNodeDelegation:                 minKeyNodeDelegation,
		MaxCandidateKeyNodeHeartbeatInterval: maxCandidateKeyNodeHeartbeatInterval,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyUnbondingTime, &p.UnbondingTime},
		{KeyMaxValidators, &p.MaxValidators},
		{KeyMaxEntries, &p.MaxEntries},
		{KeyBondDenom, &p.BondDenom},
		{KeyMaxKeyNodes, &p.MaxKeyNodes},
		{KeyMinValidatorDelegation, &p.MinValidatorDelegation},
		{KeyMinKeyNodeDelegation, &p.MinKeyNodeDelegation},
		{KeyMaxCandidateKeyNodeHeartbeatInterval, &p.MaxCandidateKeyNodeHeartbeatInterval},
	}
}

// Equal returns a boolean determining if two Param types are identical.
// TODO: This is slower than comparing struct fields directly
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultUnbondingTime, DefaultMaxValidators, DefaultMaxKeyNodes, DefaultMaxEntries, sdk.DefaultBondDenom,
		DefaultMinValidatorDelegation, DefaultMinKeyNodeDelegation, DefaultMaxCandidateKeyNodeHeartbeatInterval)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  UnbondingTime: %s
  MaxValidators: %d
  MaxKeyNodes: %d
  MaxEntries: %d
  MaxCandidateKeyNodeHeartbeatInterval: %d
  BondDenom: %s
  MinValidatorDelegation: %s
  MinKeyNodeDelegation: %s`,
		p.UnbondingTime, p.MaxValidators, p.MaxKeyNodes, p.MaxEntries, p.MaxCandidateKeyNodeHeartbeatInterval,
		p.BondDenom, p.MinValidatorDelegation.String(), p.MinKeyNodeDelegation.String())
}

// unmarshal the current staking params value from store key or panic
func MustUnmarshalParams(cdc *codec.Codec, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}
	return params
}

// unmarshal the current staking params value from store key
func UnmarshalParams(cdc *codec.Codec, value []byte) (params Params, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &params)
	if err != nil {
		return
	}
	return
}

// validate a set of params
func (p Params) Validate() error {
	if p.BondDenom == "" {
		return fmt.Errorf("staking parameter BondDenom can't be an empty string")
	}
	if p.MaxValidators == 0 {
		return fmt.Errorf("staking parameter MaxValidators must be a positive integer")
	}
	if p.MaxKeyNodes == 0 {
		return fmt.Errorf("staking parameter MaxKeyNodes must be a positive integer")
	}
	if p.MinValidatorDelegation.IsNegative() {
		return errors.New("min validator delegation cannot be negative")
	}
	if p.MinKeyNodeDelegation.IsNegative() {
		return errors.New("min validator delegation cannot be negative")
	}
	return nil
}
