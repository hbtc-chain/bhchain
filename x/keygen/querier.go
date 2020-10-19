package keygen

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/keygen/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryWaitAssignKeys:
			return queryWaitAssignKeys(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown order query endpoint")
		}
	}
}

func queryWaitAssignKeys(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	orderIDs := keeper.GetWaitAssignKeyGenOrderIDs(ctx)
	res, err := codec.MarshalJSONIndent(keeper.cdc, orderIDs)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}
