package types

import sdk "github.com/hbtc-chain/bhchain/types"

const (
	// query balance path
	QueryBalance    = "balance"
	QueryAllBalance = "balances"
)

type QueryBalanceParams struct {
	Addr   sdk.CUAddress
	Symbol string
}

func NewQueryBalanceParams(addr sdk.CUAddress, symbol string) QueryBalanceParams {
	return QueryBalanceParams{
		Addr:   addr,
		Symbol: symbol,
	}
}

type QueryAllBalanceParams struct {
	Addr sdk.CUAddress
}

func NewQueryAllBalanceParams(addr sdk.CUAddress) QueryAllBalanceParams {
	return QueryAllBalanceParams{
		Addr: addr,
	}
}

type ResBalance struct {
	Available sdk.Coin `json:"available"`
	Locked    sdk.Coin `json:"locked"`
}

type ResAllBalance struct {
	Available sdk.Coins `json:"available"`
	Locked    sdk.Coins `json:"locked"`
}
