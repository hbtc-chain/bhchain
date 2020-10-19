package order

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/order/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryOrder:
			return queryOrder(ctx, req, keeper)
		case types.QueryProcessList:
			return queryProcessList(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown order query endpoint")
		}
	}
}

// nolint: unparam
func queryOrder(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryOrderParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	as := keeper.GetOrder(ctx, params.OrderID)
	if as == nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("has no order %v", params.OrderID))
	}
	res, err := codec.MarshalJSONIndent(keeper.cdc, as)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// nolint: unparam
func queryProcessList(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	orderIDs := keeper.GetProcessOrderList(ctx)
	res, err := codec.MarshalJSONIndent(keeper.cdc, orderIDs)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}
