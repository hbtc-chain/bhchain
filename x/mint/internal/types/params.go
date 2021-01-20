package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

// Parameter store keys
var (
	KeyMintDenom    = []byte("MintDenom")
	KeyInflation    = []byte("Inflation")
	KeyMintPerBlock = []byte("MintPerBlock")
)

// mint parameters
type Params struct {
	MintDenom    string  `json:"mint_denom" yaml:"mint_denom"`         // type of coin to mint
	Inflation    sdk.Dec `json:"inflation" yaml:"inflation"`           //target inflation, default=3%
	MintPerBlock sdk.Int `json:"mint_per_block" yaml:"mint_per_block"` //target inflation, default=3%
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(mintDenom string, inflation sdk.Dec, mintPerBlock sdk.Int) Params {
	return Params{
		MintDenom:    mintDenom,
		Inflation:    inflation,
		MintPerBlock: mintPerBlock,
	}
}

// default minting module parameters
func DefaultParams() Params {
	initMintPerBlock := sdk.NewIntWithDecimal(1, 17)
	return Params{
		MintDenom:    sdk.DefaultBondDenom,
		Inflation:    sdk.NewDecWithPrec(3, 2),
		MintPerBlock: initMintPerBlock,
	}
}

// validate params
func ValidateParams(params Params) error {
	if params.MintDenom == "" {
		return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	}

	if params.Inflation.IsNegative() {
		return fmt.Errorf("mint parameter Inflation can't be negative")
	}

	if !params.MintPerBlock.IsPositive() {
		return fmt.Errorf("mint parameter MintPerBlock can't be negative")
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:                %s
  Inflation:                 %s
  MintPerBlock:              %s
`,
		p.MintDenom, p.Inflation.String(), p.MintPerBlock.String())
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMintDenom, &p.MintDenom},
		{KeyInflation, &p.Inflation},
		{KeyMintPerBlock, &p.MintPerBlock},
	}
}
