package types

import sdk "github.com/hbtc-chain/bhchain/types"

// query endpoints supported by the upgrade Querier
const (
	QueryDex               = "dex"
	QueryAllDex            = "all-dex"
	QueryTradingPair       = "trading_pair"
	QueryAllTradingPair    = "all_trading_pair"
	QueryAddrLiquidity     = "addr_liquidity"
	QueryOrderbook         = "orderbook"
	QueryOrder             = "order"
	QueryUnfinishedOrder   = "unfinished_order"
	QueryUnclaimedEarnings = "unclaimed_earnings"
	QueryRepurchaseFunds   = "repurchase_funds"
	QueryParameters        = "parameters"
)

type QueryDexParams struct {
	DexID uint32
}

func NewQueryDexParams(dexID uint32) QueryDexParams {
	return QueryDexParams{
		DexID: dexID,
	}
}

type QueryTradingPairParams struct {
	DexID  uint32
	TokenA sdk.Symbol
	TokenB sdk.Symbol
}

func NewQueryTradingPairParams(dexID uint32, tokenA, tokenB sdk.Symbol) QueryTradingPairParams {
	return QueryTradingPairParams{
		DexID:  dexID,
		TokenA: tokenA,
		TokenB: tokenB,
	}
}

type QueryAllTradingPairParams struct {
	DexID *uint32
}

func NewQueryAllTradingPairParams(dexID *uint32) QueryAllTradingPairParams {
	return QueryAllTradingPairParams{
		DexID: dexID,
	}
}

type QueryAddrLiquidityParams struct {
	Addr  sdk.CUAddress
	DexID *uint32
}

func NewQueryAddrLiquidityParams(addr sdk.CUAddress, dexID *uint32) QueryAddrLiquidityParams {
	return QueryAddrLiquidityParams{
		Addr:  addr,
		DexID: dexID,
	}
}

type QueryOrderbookParams struct {
	DexID       uint32
	BaseSymbol  sdk.Symbol
	QuoteSymbol sdk.Symbol
	Merge       bool
}

func NewQueryOrderbookParams(dexID uint32, baseSymbol, quoteSymbol sdk.Symbol, merge bool) QueryOrderbookParams {
	return QueryOrderbookParams{
		DexID:       dexID,
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
	Addr        sdk.CUAddress
	DexID       uint32
	BaseSymbol  sdk.Symbol
	QuoteSymbol sdk.Symbol
}

func NewQueryUnfinishedOrderParams(addr sdk.CUAddress, dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) QueryUnfinishedOrderParams {
	return QueryUnfinishedOrderParams{
		Addr:        addr,
		DexID:       dexID,
		BaseSymbol:  baseSymbol,
		QuoteSymbol: quoteSymbol,
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
