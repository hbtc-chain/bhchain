package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

var (
	DefaultMinimumLiquidity            = sdk.NewInt(1000)
	DefaultLimitSwapMatchingGas        = sdk.NewUint(50000)
	DefaultMaxFeeRate                  = sdk.NewDecWithPrec(1, 1)  // 0.1
	DefaultLpRewardRate                = sdk.NewDecWithPrec(25, 4) // 0.0025
	DefaultRefererTransactionBonusRate = sdk.NewDecWithPrec(1, 4)  // 0.0001
	DefaultRepurchaseRate              = sdk.NewDecWithPrec(4, 4)  // 0.0004
	DefaultRefererMiningBonusRate      = sdk.NewDecWithPrec(1, 1)  // 0.1
	DefaultRepurchaseDuration          = int64(1000)
	DefaultRepurchaseToken             = sdk.NativeToken
	DefaultMiningWeights               = []*MiningWeight{}
	DefaultMiningPlans                 = []*MiningPlan{}
)

var (
	KeyMinimumLiquidity            = []byte("MinimumLiquidity")
	KeyLimitSwapMatchingGas        = []byte("LimitSwapMatchingGas")
	KeyMaxFeeRate                  = []byte("MaxFeeRate")
	KeyLpRewardRate                = []byte("LpRewardRate")
	KeyRepurchaseRate              = []byte("RepurchaseRate")
	KeyRefererTransactionBonusRate = []byte("RefererTransactionBonusRate")
	KeyRefererMiningBonusRate      = []byte("RefererMiningBonusRate")
	KeyRepurchaseDuration          = []byte("RepurchaseDuration")
	KeyRepurchaseToken             = []byte("RepurchaseToken")
	KeyMiningWeights               = []byte("MiningWeights")
	KeyMiningPlans                 = []byte("MiningPlans")
)

type MiningWeight struct {
	DexID  uint32     `json:"dex_id"`
	TokenA sdk.Symbol `json:"token_a"`
	TokenB sdk.Symbol `json:"token_b"`
	Weight sdk.Int    `json:"weight"`
}

func NewMiningWeight(dexID uint32, tokenA, tokenB sdk.Symbol, weight sdk.Int) *MiningWeight {
	return &MiningWeight{
		DexID:  dexID,
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
	LimitSwapMatchingGas        sdk.Uint        `json:"limit_swap_matching_gas"`
	MaxFeeRate                  sdk.Dec         `json:"max_fee_rate"`
	LpRewardRate                sdk.Dec         `json:"lp_reward_rate"`
	RepurchaseRate              sdk.Dec         `json:"repurchase_rate"`
	RefererTransactionBonusRate sdk.Dec         `json:"referer_transaction_bonus_rate"`
	RefererMiningBonusRate      sdk.Dec         `json:"referer_mining_bonus_rate"`
	RepurchaseDuration          int64           `json:"repurchase_duration"`
	MiningWeights               []*MiningWeight `json:"mining_weights"`
	MiningPlans                 []*MiningPlan   `json:"mining_plans"`
	RepurchaseToken             string          `json:"repurchase_token"`
}

// NewParams creates a new Params instance
func NewParams(minLiquidity sdk.Int, limitSwapMatchingGas sdk.Uint, maxFeeRate, lpRewardRate, repurchaseRate, refererTransactionBonusRate, refererMiningBonusRate sdk.Dec,
	repurchaseDuration int64, miningWeights []*MiningWeight, miningPlans []*MiningPlan, repurchaseToken string) Params {
	return Params{
		MinimumLiquidity:            minLiquidity,
		LimitSwapMatchingGas:        limitSwapMatchingGas,
		MaxFeeRate:                  maxFeeRate,
		LpRewardRate:                lpRewardRate,
		RepurchaseRate:              repurchaseRate,
		RefererTransactionBonusRate: refererTransactionBonusRate,
		RefererMiningBonusRate:      refererMiningBonusRate,
		RepurchaseDuration:          repurchaseDuration,
		MiningWeights:               miningWeights,
		MiningPlans:                 miningPlans,
		RepurchaseToken:             repurchaseToken,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMinimumLiquidity, &p.MinimumLiquidity},
		{KeyLimitSwapMatchingGas, &p.LimitSwapMatchingGas},
		{KeyMaxFeeRate, &p.MaxFeeRate},
		{KeyLpRewardRate, &p.LpRewardRate},
		{KeyRepurchaseRate, &p.RepurchaseRate},
		{KeyRefererTransactionBonusRate, &p.RefererTransactionBonusRate},
		{KeyRefererMiningBonusRate, &p.RefererMiningBonusRate},
		{KeyRepurchaseDuration, &p.RepurchaseDuration},
		{KeyMiningWeights, &p.MiningWeights},
		{KeyMiningPlans, &p.MiningPlans},
		{KeyRepurchaseToken, &p.RepurchaseToken},
	}
}

func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultMinimumLiquidity, DefaultLimitSwapMatchingGas, DefaultMaxFeeRate, DefaultLpRewardRate,
		DefaultRepurchaseRate, DefaultRefererTransactionBonusRate, DefaultRefererMiningBonusRate,
		DefaultRepurchaseDuration, DefaultMiningWeights, DefaultMiningPlans, DefaultRepurchaseToken)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  MinimumLiquidity: %s
  LimitSwapMatchingGas: %s
  MaxFeeRate: %s
  LpRewardRate: %s
  RepurchaseRate: %s
  RefererTransactionBonusRate: %s
  RefererMiningBonusRate: %s
  RepurchaseDuration: %d
  RepurchaseToken: %s
  MiningWeights: %v
  MiningPlans: %v`,
		p.MinimumLiquidity.String(), p.LimitSwapMatchingGas.String(), p.MaxFeeRate.String(),
		p.LpRewardRate.String(), p.RepurchaseRate.String(), p.RefererTransactionBonusRate.String(),
		p.RefererMiningBonusRate.String(), p.RepurchaseDuration, p.RepurchaseToken, p.MiningWeights, p.MiningPlans)
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
	if p.MaxFeeRate.IsNegative() || p.MaxFeeRate.GT(sdk.OneDec()) {
		return errors.New("max fee rate must be between 0 to 1")
	}
	if p.LpRewardRate.IsNegative() {
		return errors.New("fee rate cannot be negative")
	}
	if p.RepurchaseRate.IsNegative() {
		return errors.New("repurchase rate cannot be negative")
	}
	if p.RefererTransactionBonusRate.IsNegative() {
		return errors.New("referer transaction bonus rate cannot be negative")
	}
	if p.LpRewardRate.Add(p.RepurchaseRate).Add(p.RefererTransactionBonusRate).GT(p.MaxFeeRate) {
		return errors.New("sum of fee rate must be less than max fee rate")
	}
	if p.RepurchaseDuration <= 0 {
		return errors.New("repurchase duration should be positive")
	}
	if p.RefererMiningBonusRate.IsNegative() || p.RefererMiningBonusRate.GT(sdk.OneDec()) {
		return errors.New("referer mining bonus rate must be between 0 to 1")
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
