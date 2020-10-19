package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

// QueryBalanceParams defines the params for querying an CU balance.
type QueryBalanceParams struct {
	Address sdk.CUAddress
}

// NewQueryBalanceParams creates a new instance of QueryBalanceParams.
func NewQueryBalanceParams(addr sdk.CUAddress) QueryBalanceParams {
	return QueryBalanceParams{Address: addr}
}
