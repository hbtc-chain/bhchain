package keeper

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/x/upgrade/types"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// NewQuerier creates a querier for upgrade cli and REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {

		case types.QueryCurrent:
			return queryCurrent(ctx, req, k)

		case types.QueryApplied:
			return queryApplied(ctx, req, k)

		default:
			return nil, sdk.ErrUnknownRequest("unknown upgrade query endpoint")
		}
	}
}

func queryCurrent(ctx sdk.Context, _ abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	plan, has := k.GetUpgradePlan(ctx)
	if !has {
		return nil, nil
	}

	res := k.cdc.MustMarshalJSON(&plan)
	return res, nil
}

func queryApplied(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryAppliedParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	applied := k.GetDoneHeight(ctx, params.Name)
	bz := k.cdc.MustMarshalJSON(applied)
	return bz, nil
}
