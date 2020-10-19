package types

import sdk "github.com/hbtc-chain/bhchain/types"

type Earning struct {
	TokenA sdk.Symbol `json:"token_a"`
	TokenB sdk.Symbol `json:"token_b"`
	Amount sdk.Int    `json:"amount"`
}

func NewEarning(tokenA, tokenB sdk.Symbol, amount sdk.Int) *Earning {
	return &Earning{
		TokenA: tokenA,
		TokenB: tokenB,
		Amount: amount,
	}
}
