package types

import (
	"errors"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type TradingPair struct {
	DexID             uint32     `json:"dex_id"`
	TokenA            sdk.Symbol `json:"token_a"`
	TokenB            sdk.Symbol `json:"token_b"`
	TokenAAmount      sdk.Int    `json:"token_a_amount"`
	TokenBAmount      sdk.Int    `json:"token_b_amount"`
	TotalLiquidity    sdk.Int    `json:"total_liquidity"`
	IsPublic          bool       `json:"is_public"`
	LPRewardRate      sdk.Dec    `json:"lp_reward_rate"`
	RefererRewardRate sdk.Dec    `json:"referer_reward_rate"`
}

func NewDefaultTradingPair(tokenA, tokenB sdk.Symbol, initialLiquidity sdk.Int) *TradingPair {
	return &TradingPair{
		DexID:             0,
		TokenA:            tokenA,
		TokenB:            tokenB,
		TokenAAmount:      sdk.ZeroInt(),
		TokenBAmount:      sdk.ZeroInt(),
		TotalLiquidity:    initialLiquidity,
		IsPublic:          true,
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
}

func NewCustomTradingPair(dexID uint32, tokenA, tokenB sdk.Symbol, isPublic bool, lpRewardRate, refererRewardRate sdk.Dec) *TradingPair {
	return &TradingPair{
		DexID:             dexID,
		TokenA:            tokenA,
		TokenB:            tokenB,
		TokenAAmount:      sdk.ZeroInt(),
		TokenBAmount:      sdk.ZeroInt(),
		TotalLiquidity:    sdk.ZeroInt(),
		IsPublic:          isPublic,
		LPRewardRate:      lpRewardRate,
		RefererRewardRate: refererRewardRate,
	}
}

func (t *TradingPair) Validate() error {
	if t.TokenA >= t.TokenB {
		return errors.New("wrong symbol sequence")
	}
	if !t.TokenA.IsValid() || !t.TokenB.IsValid() {
		return errors.New("invalid symbol")
	}
	return nil
}

func (t *TradingPair) Price() sdk.Dec {
	if t.TokenAAmount.IsZero() {
		return sdk.ZeroDec()
	}
	return t.TokenBAmount.ToDec().Quo(t.TokenAAmount.ToDec())
}

type ResTradingPair struct {
	DexID             uint32     `json:"dex_id"`
	TokenA            sdk.Symbol `json:"token_a"`
	TokenB            sdk.Symbol `json:"token_b"`
	TokenAAmount      sdk.Int    `json:"token_a_amount"`
	TokenBAmount      sdk.Int    `json:"token_b_amount"`
	TotalLiquidity    sdk.Int    `json:"total_liquidity"`
	IsPublic          bool       `json:"is_public"`
	LPRewardRate      sdk.Dec    `json:"lp_reward_rate"`
	RefererRewardRate sdk.Dec    `json:"referer_reward_rate"`
}

func NewResTradingPair(pair *TradingPair) *ResTradingPair {
	return &ResTradingPair{
		DexID:             pair.DexID,
		TokenA:            pair.TokenA,
		TokenB:            pair.TokenB,
		TokenAAmount:      pair.TokenAAmount,
		TokenBAmount:      pair.TokenBAmount,
		TotalLiquidity:    pair.TotalLiquidity,
		IsPublic:          pair.IsPublic,
		LPRewardRate:      pair.LPRewardRate,
		RefererRewardRate: pair.RefererRewardRate,
	}
}

func NewResTradingPairs(pairs []*TradingPair) []*ResTradingPair {
	ret := make([]*ResTradingPair, len(pairs))
	for i := range pairs {
		ret[i] = NewResTradingPair(pairs[i])
	}
	return ret
}

type AddrLiquidity struct {
	*TradingPair   `json:"trading_pair"`
	Liquidity      sdk.Int `json:"liquidity"`
	LiquidityShare sdk.Dec `json:"liquidity_share"`
}

func NewAddrLiquidity(pair *TradingPair, liquidity sdk.Int) *AddrLiquidity {
	ret := &AddrLiquidity{
		TradingPair: pair,
		Liquidity:   liquidity,
	}
	ret.LiquidityShare = sdk.NewDecFromInt(ret.Liquidity).Quo(sdk.NewDecFromInt(ret.TotalLiquidity))
	return ret
}
