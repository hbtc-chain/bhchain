package types

import (
	"errors"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type TradingPair struct {
	TokenA         sdk.Symbol `json:"token_a"`
	TokenB         sdk.Symbol `json:"token_b"`
	TokenAAmount   sdk.Int    `json:"token_a_amount"`
	TokenBAmount   sdk.Int    `json:"token_b_amount"`
	TotalLiquidity sdk.Int    `json:"total_liquidity"`
}

func NewTradingPair(tokenA, tokenB sdk.Symbol, initialLiquidity sdk.Int) *TradingPair {
	return &TradingPair{
		TokenA:         tokenA,
		TokenB:         tokenB,
		TokenAAmount:   sdk.ZeroInt(),
		TokenBAmount:   sdk.ZeroInt(),
		TotalLiquidity: initialLiquidity,
	}
}

func (t *TradingPair) Validate() error {
	if t.TokenA >= t.TokenB {
		return errors.New("wrong symbol sequence")
	}
	if !t.TokenA.IsValidTokenName() || !t.TokenB.IsValidTokenName() {
		return errors.New("invalid symbol")
	}
	return nil
}

func (t *TradingPair) Price() sdk.Dec {
	return t.TokenBAmount.ToDec().Quo(t.TokenAAmount.ToDec())
}

type ResTradingPair struct {
	TokenA         sdk.Symbol `json:"token_a"`
	TokenB         sdk.Symbol `json:"token_b"`
	TokenAAmount   sdk.Int    `json:"token_a_amount"`
	TokenBAmount   sdk.Int    `json:"token_b_amount"`
	TotalLiquidity sdk.Int    `json:"total_liquidity"`
}

func NewResTradingPair(pair *TradingPair) *ResTradingPair {
	return &ResTradingPair{
		TokenA:         pair.TokenA,
		TokenB:         pair.TokenB,
		TokenAAmount:   pair.TokenAAmount,
		TokenBAmount:   pair.TokenBAmount,
		TotalLiquidity: pair.TotalLiquidity,
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
