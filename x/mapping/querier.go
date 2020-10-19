package mapping

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/mapping/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryInfo:
			return queryMapping(ctx, req, keeper)
		case QueryList:
			return queryMappingList(ctx, req, keeper)
		case QueryDirectSwapInfo:
			return queryDirectSwapInfo(ctx, req, keeper)
		case QueryFreeSwapInfo:
			return queryFreeSwapInfo(ctx, req, keeper)
		case QueryDirectSwapList:
			return queryDirectSwapOrderList(ctx, req, keeper)
		case QueryFreeSwapList:
			return queryFreeSwapOrderList(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown mapping query endpoint")
		}
	}
}

//=====
func queryMapping(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryMappingParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	symbol := sdk.Symbol(params.IssueSymbol)
	res := keeper.GetMappingInfo(ctx, symbol)
	if res == nil {
		return nil, sdk.ErrUnknownRequest("Non-exits issue symbol")
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.MappingInfoToQueryRes(res))
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

//=====
func queryMappingList(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryMappingListParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	res := keeper.GetIssueSymbols(ctx)

	start, end := client.Paginate(len(res), params.Page, params.Limit, 100)

	var mappingList QueryResMappingList

	if start >= 0 && end >= 0 {
		for _, symbol := range res[start:end] {
			mappingInfo := keeper.GetMappingInfo(ctx, symbol)
			if mappingInfo == nil {
				panic("could not get mapping info")
			}
			mappingList = append(mappingList, types.MappingInfoToQueryRes(mappingInfo))
		}
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, mappingList)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryFreeSwapInfo(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryFreeSwapOrderParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	res := keeper.GetFreeSwapOrder(ctx, params.OrderID)
	if res == nil {
		return nil, sdk.ErrUnknownRequest("Non-exits issue symbol")
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, res)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryFreeSwapOrderList(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryFreeSwapOrderListParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	res := keeper.GetFreeSwapOrders(ctx)
	start, end := client.Paginate(len(res), params.Page, params.Limit, 100)

	var freeSwapOrderList QueryFreeSwapOrderList

	if start >= 0 && end >= 0 {
		for _, order := range res[start:end] {
			freeSwapOrderList = append(freeSwapOrderList, order)
		}
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, freeSwapOrderList)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryDirectSwapInfo(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryDirectSwapOrderParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	res := keeper.GetDirectSwapOrder(ctx, params.OrderID)
	if res == nil {
		return nil, sdk.ErrUnknownRequest("Non-exits issue symbol")
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, res)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryDirectSwapOrderList(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryDirectSwapOrderListParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	res := keeper.GetDirectSwapOrders(ctx)
	start, end := client.Paginate(len(res), params.Page, params.Limit, 100)

	var directSwapOrderList QueryDirectSwapOrderList

	if start >= 0 && end >= 0 {
		for _, order := range res[start:end] {
			directSwapOrderList = append(directSwapOrderList, order)
		}
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, directSwapOrderList)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}
