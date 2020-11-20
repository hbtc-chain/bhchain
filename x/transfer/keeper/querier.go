package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

// NewQuerier returns a new sdk.Keeper instance.
func NewQuerier(k BaseKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryBalance:
			return queryBalance(ctx, req, k)
		case types.QueryAllBalance:
			return queryAllBalance(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown bank query endpoint")
		}
	}
}

// queryBalance fetch an CU's balance for the supplied height.
// Height and CU address are passed as first and second path components respectively.
func queryBalance(ctx sdk.Context, req abci.RequestQuery, k BaseKeeper) ([]byte, sdk.Error) {

	var r types.QueryBalanceParams
	if err := k.cdc.UnmarshalJSON(req.Data, &r); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	balance := types.ResBalance{
		Available: sdk.NewCoin(r.Symbol, k.GetBalance(ctx, r.Addr, r.Symbol)),
		Locked:    sdk.NewCoin(r.Symbol, k.GetHoldBalance(ctx, r.Addr, r.Symbol)),
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, balance)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryAllBalance(ctx sdk.Context, req abci.RequestQuery, k BaseKeeper) ([]byte, sdk.Error) {

	var r types.QueryAllBalanceParams
	if err := k.cdc.UnmarshalJSON(req.Data, &r); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	balance := types.ResAllBalance{
		Available: k.GetAllBalance(ctx, r.Addr),
		Locked:    k.GetAllHoldBalance(ctx, r.Addr),
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, balance)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}
