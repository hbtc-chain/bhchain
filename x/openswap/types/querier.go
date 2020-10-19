package types

import sdk "github.com/hbtc-chain/bhchain/types"

// query endpoints supported by the upgrade Querier
const (
	QueryTradingPair       = "trading_pair"
	QueryAllTradingPair    = "all_trading_pair"
	QueryAddrLiquidity     = "addr_liquidity"
	QueryOrderbook         = "orderbook"
	QueryOrder             = "order"
	QueryUnfinishedOrder   = "unfinished_order"
	QueryUnclaimedEarnings = "unclaimed_earnings"
	QueryParameters        = "parameters"
)

type QueryTradingPairParams struct {
	TokenA sdk.Symbol
	TokenB sdk.Symbol
}

func NewQueryTradingPairParams(tokenA, tokenB sdk.Symbol) QueryTradingPairParams {
	return QueryTradingPairParams{
		TokenA: tokenA,
		TokenB: tokenB,
	}
}

type QueryAddrLiquidityParams struct {
	Addr sdk.CUAddress
}

func NewQueryAddrLiquidityParams(addr sdk.CUAddress) QueryAddrLiquidityParams {
	return QueryAddrLiquidityParams{
		Addr: addr,
	}
}

type QueryOrderbookParams struct {
	BaseSymbol  sdk.Symbol
	QuoteSymbol sdk.Symbol
	Merge       bool
}

func NewQueryOrderbookParams(baseSymbol, quoteSymbol sdk.Symbol, merge bool) QueryOrderbookParams {
	return QueryOrderbookParams{
		BaseSymbol:  baseSymbol,
		QuoteSymbol: quoteSymbol,
		Merge:       merge,
	}
}

type QueryOrderParams struct {
	OrderID string
}

func NewQueryOrderParams(orderID string) QueryOrderParams {
	return QueryOrderParams{OrderID: orderID}
}

type QueryUnfinishedOrderParams struct {
	BaseSymbol  sdk.Symbol
	QuoteSymbol sdk.Symbol
	Addr        sdk.CUAddress
}

func NewQueryUnfinishedOrderParams(baseSymbol, quoteSymbol sdk.Symbol, addr sdk.CUAddress) QueryUnfinishedOrderParams {
	return QueryUnfinishedOrderParams{
		BaseSymbol:  baseSymbol,
		QuoteSymbol: quoteSymbol,
		Addr:        addr,
	}
}

type QueryUnclaimedEarningParams struct {
	Addr sdk.CUAddress
}

func NewQueryUnclaimedEarningParams(addr sdk.CUAddress) QueryUnclaimedEarningParams {
	return QueryUnclaimedEarningParams{
		Addr: addr,
	}
}
