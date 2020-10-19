package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

const (
	RepurchaseRoutingCoin = sdk.NativeToken
)

var (
	DefaultMinimumLiquidity            = sdk.NewInt(1000)
	DefaultFeeRate                     = sdk.NewDecWithPrec(225, 5) // 0.00225
	DefaultRefererTransactionBonusRate = sdk.NewDecWithPrec(25, 5)  // 0.00025
	DefaultRepurchaseRate              = sdk.NewDecWithPrec(5, 4)   // 0.0005
	DefaultRefererMiningBonusRate      = sdk.NewDecWithPrec(1, 1)   // 0.1
	DefaultMiningWeights               = []*MiningWeight{NewMiningWeight("hbc", "test", sdk.OneInt())}
	DefaultMiningPlans                 = []*MiningPlan{
		NewMiningPlan(1, sdk.NewInt(3000000000)),      // 30 perblock
		NewMiningPlan(650001, sdk.NewInt(1500000000)), // 15 perblock
		NewMiningPlan(1000001, sdk.NewInt(750000000)), // 7.5 perblock
		NewMiningPlan(1700001, sdk.NewInt(375000000)), // 3.75 perblock
		NewMiningPlan(3500001, sdk.NewInt(300000000)), // 3 perblock
	}
)

var (
	KeyMinimumLiquidity            = []byte("MinimumLiquidity")
	KeyFeeRate                     = []byte("FeeRate")
	KeyRepurchaseRate              = []byte("RepurchaseRate")
	KeyRefererTransactionBonusRate = []byte("RefererTransactionBonusRate")
	KeyRefererMiningBonusRate      = []byte("RefererMiningBonusRate")
	KeyMiningWeights               = []byte("MiningWeights")
	KeyMiningPlans                 = []byte("MiningPlans")
)

type MiningWeight struct {
	TokenA sdk.Symbol `json:"token_a"`
	TokenB sdk.Symbol `json:"token_b"`
	Weight sdk.Int    `json:"weight"`
}

func NewMiningWeight(tokenA, tokenB sdk.Symbol, weight sdk.Int) *MiningWeight {
	return &MiningWeight{
		TokenA: tokenA,
		TokenB: tokenB,
		Weight: weight,
	}
}

type MiningPlan struct {
	StartHeight    uint64  `json:"start_height"`
	MiningPerBlock sdk.Int `json:"mining_per_block"`
}

func NewMiningPlan(startHeight uint64, miningPerBlock sdk.Int) *MiningPlan {
	return &MiningPlan{
		StartHeight:    startHeight,
		MiningPerBlock: miningPerBlock,
	}
}

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	MinimumLiquidity            sdk.Int         `json:"minimum_liquidity"`
	FeeRate                     sdk.Dec         `json:"fee_rate"`
	RepurchaseRate              sdk.Dec         `json:"repurchase_rate"`
	RefererTransactionBonusRate sdk.Dec         `json:"referer_transaction_bonus_rate"`
	RefererMiningBonusRate      sdk.Dec         `json:"referer_mining_bonus_rate"`
	MiningWeights               []*MiningWeight `json:"mining_weights"`
	MiningPlans                 []*MiningPlan   `json:"mining_plans"`
}

// NewParams creates a new Params instance
func NewParams(minLiquidity sdk.Int, feeRate, repurchaseRate, refererTransactionBonusRate, refererMiningBonusRate sdk.Dec,
	miningWeights []*MiningWeight, miningPlans []*MiningPlan) Params {
	return Params{
		MinimumLiquidity:            minLiquidity,
		FeeRate:                     feeRate,
		RepurchaseRate:              repurchaseRate,
		RefererTransactionBonusRate: refererTransactionBonusRate,
		RefererMiningBonusRate:      refererMiningBonusRate,
		MiningWeights:               miningWeights,
		MiningPlans:                 miningPlans,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMinimumLiquidity, &p.MinimumLiquidity},
		{KeyFeeRate, &p.FeeRate},
		{KeyRepurchaseRate, &p.RepurchaseRate},
		{KeyRefererTransactionBonusRate, &p.RefererTransactionBonusRate},
		{KeyRefererMiningBonusRate, &p.RefererMiningBonusRate},
		{KeyMiningWeights, &p.MiningWeights},
		{KeyMiningPlans, &p.MiningPlans},
	}
}

func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultMinimumLiquidity, DefaultFeeRate, DefaultRepurchaseRate, DefaultRefererTransactionBonusRate,
		DefaultRefererMiningBonusRate, DefaultMiningWeights, DefaultMiningPlans)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  MinimumLiquidity: %s
  FeeRate: %s
  RepurchaseRate: %s
  RefererTransactionBonusRate: %s
  RefererMiningBonusRate: %s
  MiningWeights: %v
  MiningPlans: %v`,
		p.MinimumLiquidity.String(), p.FeeRate.String(), p.RepurchaseRate.String(), p.RefererTransactionBonusRate.String(),
		p.RefererMiningBonusRate.String(), p.MiningWeights, p.MiningPlans)
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
	if !p.MinimumLiquidity.IsPositive() {
		return errors.New("minimun liquidity should be positive")
	}
	if p.FeeRate.IsNegative() {
		return errors.New("fee rate cannot be negative")
	}
	if p.RepurchaseRate.IsNegative() {
		return errors.New("repurchase rate cannot be negative")
	}
	if p.RefererTransactionBonusRate.IsNegative() {
		return errors.New("referer transaction bonus rate cannot be negative")
	}
	if p.FeeRate.Add(p.RefererTransactionBonusRate).GT(sdk.OneDec()) {
		return errors.New("sum of fee rate and referer transaction bonus rate must be less than 1")
	}
	if p.RefererMiningBonusRate.IsNegative() || p.RefererMiningBonusRate.GT(sdk.OneDec()) {
		return errors.New("referer mining bonus rate must be between 0 to 1")
	}
	if len(p.MiningWeights) == 0 {
		return errors.New("empty mining weights")
	}
	exists := make(map[string]bool)
	for _, w := range p.MiningWeights {
		if w.TokenA == w.TokenB {
			return errors.New("tokenA and tokenB can be the same")
		}
		tokenA, tokenB := w.TokenA, w.TokenB
		if tokenA > tokenB {
			tokenA, tokenB = tokenB, tokenA
		}
		pair := fmt.Sprintf("%s-%s", tokenA, tokenB)
		if exists[pair] {
			return fmt.Errorf("%s is duplicated", pair)
		}
		exists[pair] = true
		if !w.Weight.IsPositive() {
			return errors.New("weight should be positive")
		}
	}
	if len(p.MiningPlans) == 0 {
		return errors.New("empty mining plans")
	}
	for i, w := range p.MiningPlans {
		if !w.MiningPerBlock.IsPositive() {
			return errors.New("mining perblock should be positive")
		}
		if i > 0 && w.StartHeight <= p.MiningPlans[i-1].StartHeight {
			return errors.New("start height should be ascending")
		}
	}

	return nil
}
