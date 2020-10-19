package keeper

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryTradingPair:
			return queryTradingPair(ctx, req, k)
		case types.QueryAllTradingPair:
			return queryAllTradingPair(ctx, req, k)
		case types.QueryAddrLiquidity:
			return queryAddrLiquidity(ctx, req, k)
		case types.QueryOrderbook:
			return queryOrderbook(ctx, req, k)
		case types.QueryOrder:
			return queryOrder(ctx, req, k)
		case types.QueryUnfinishedOrder:
			return queryUnfinishedOrder(ctx, req, k)
		case types.QueryUnclaimedEarnings:
			return queryUnclaimedEarnings(ctx, req, k)
		case types.QueryParameters:
			return queryParameters(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown dex query endpoint")
		}
	}
}

func queryTradingPair(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryTradingPairParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	tokenA, tokenB := params.TokenA, params.TokenB
	pair := k.GetTradingPair(ctx, tokenA, tokenB)
	if pair == nil {
		return nil, sdk.ErrInvalidTx(fmt.Sprintf("no trading pair of %s-%s", tokenA.String(), tokenB.String()))
	}
	bz, err := codec.MarshalJSONIndent(k.cdc, types.NewResTradingPair(pair))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryAllTradingPair(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	pairs := k.getAllTradingPairs(ctx)
	bz, err := codec.MarshalJSONIndent(k.cdc, types.NewResTradingPairs(pairs))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryAddrLiquidity(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryAddrLiquidityParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	liquidities := k.getAddrAllLiquidity(ctx, params.Addr)
	// filter zero liquidity
	var i int
	for _, liquidity := range liquidities {
		if liquidity.Liquidity.IsPositive() {
			liquidities[i] = liquidity
			i++
		}
	}
	liquidities = liquidities[:i]

	bz, err := codec.MarshalJSONIndent(k.cdc, liquidities)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryParameters(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryOrderbook(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryOrderbookParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	tokenA, tokenB := params.BaseSymbol, params.QuoteSymbol
	if tokenA > tokenB {
		tokenA, tokenB = tokenB, tokenA
	}
	pair := k.GetTradingPair(ctx, tokenA, tokenB)
	if pair == nil {
		return nil, sdk.ErrInvalidSymbol(fmt.Sprintf("%s-%s trading pair not found", params.BaseSymbol, params.QuoteSymbol))
	}

	sellOrders, buyOrders := k.GetAllOrders(tokenA, tokenB)
	var ret interface{}
	if params.Merge {
		ret = types.NewDepthBook(ctx.BlockHeight(), ctx.BlockTime().Unix(), fmt.Sprintf("%s-%s", tokenA, tokenB), buyOrders, sellOrders)
	} else {
		ret = map[string][]*types.Order{
			"buy":  buyOrders,
			"sell": sellOrders,
		}
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, ret)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryOrder(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryOrderParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	order := k.GetOrder(ctx, params.OrderID)
	if order == nil {
		return nil, sdk.ErrNotFoundOrder(fmt.Sprintf("order id %s not found", params.OrderID))
	}
	bz, err := codec.MarshalJSONIndent(k.cdc, types.NewResOrder(order))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryUnfinishedOrder(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryUnfinishedOrderParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	orders := k.GetAddrUnfinishedOrders(ctx, params.BaseSymbol, params.QuoteSymbol, params.Addr)
	bz, err := codec.MarshalJSONIndent(k.cdc, types.NewResOrders(orders))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryUnclaimedEarnings(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryUnclaimedEarningParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	liquidities := k.getAddrAllLiquidity(ctx, params.Addr)
	earnings := make([]*types.Earning, 0, len(liquidities))
	for _, liquidity := range liquidities {
		amount := k.CalculateEarning(ctx, params.Addr, liquidity.TokenA, liquidity.TokenB)
		earnings = append(earnings, types.NewEarning(liquidity.TokenA, liquidity.TokenB, amount))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, earnings)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
