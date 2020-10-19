package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hbtc-chain/bhchain/client"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/supply/internal/types"
)

// NewQuerier creates a querier for supply REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {

		case types.QueryTotalSupply:
			return queryTotalSupply(ctx, req, k)

		case types.QuerySupplyOf:
			return querySupplyOf(ctx, req, k)

		case types.QueryBurned:
			return queryBurned(ctx, req, k)

		case types.QueryBurnedOf:
			return queryBurnedOf(ctx, req, k)

		default:
			return nil, sdk.ErrUnknownRequest("unknown supply query endpoint")
		}
	}
}

func queryTotalSupply(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryTotalSupplyParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	totalSupply := k.GetSupply(ctx).GetTotal()

	start, end := client.Paginate(len(totalSupply), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		totalSupply = sdk.Coins{}
	} else {
		totalSupply = totalSupply[start:end]
	}

	res, err := totalSupply.MarshalJSON()
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func querySupplyOf(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QuerySupplyOfParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	supply := k.GetSupply(ctx).GetTotal().AmountOf(params.Denom)

	res, err := supply.MarshalJSON()
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func queryBurned(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryTotalSupplyParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	burned := k.GetSupply(ctx).GetBurned()

	start, end := client.Paginate(len(burned), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		burned = sdk.Coins{}
	} else {
		burned = burned[start:end]
	}

	res, err := burned.MarshalJSON()
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func queryBurnedOf(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QuerySupplyOfParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	burned := k.GetSupply(ctx).GetBurned().AmountOf(params.Denom)

	res, err := burned.MarshalJSON()
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}
