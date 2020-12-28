package types

import (
	"bytes"
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

var (
	DefaultNewMappingFee = sdk.NewInt(10)

	MaxDescLength = 1000
)

var (
	KeyNewMappingFee = []byte("NewMappingFee")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	NewMappingFee sdk.Int `json:"new_mapping_fee"`
}

// NewParams creates a new Params instance
func NewParams(newMappingFee sdk.Int) Params {
	return Params{
		NewMappingFee: newMappingFee,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyNewMappingFee, &p.NewMappingFee},
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
	return NewParams(DefaultNewMappingFee)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  NewMappingFee: %s`, p.NewMappingFee.String())
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
	if p.NewMappingFee.IsNegative() {
		return fmt.Errorf("NewMappingFee cannot be negative")
	}
	return nil
}
