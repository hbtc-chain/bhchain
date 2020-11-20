package ibcasset

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/ibcasset/exported"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset/types"
)

// creates a querier for auth REST endpoints
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryCUAsset:
			return queryCUAsset(ctx, req, keeper)
		case types.QueryOpCUAstInfo:
			return queryOpCUAstInfo(ctx, req, keeper)
		case types.QueryPendingDepositList:
			return queryPendingDepositList(ctx, req, keeper)
		case types.QueryDepositExistance:
			return queryDepositExistance(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cu query endpoint")
		}
	}
}

func queryCUAsset(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	cuAst, sdkerr := getCUAst(ctx, req, keeper)
	if sdkerr != nil {
		return nil, sdkerr
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, cuAst)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryPendingDepositList(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryPendingDepositListParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	deposits := keeper.GetPendingDepositList(ctx, params.Address)
	bz, err := keeper.cdc.MarshalJSON(deposits)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil

}

func getCUAst(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (exported.CUIBCAsset, sdk.Error) {
	var params types.QueryCUAssetParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	cuAst := keeper.GetCUIBCAsset(ctx, params.Address)
	if cuAst == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("CU %s does not exist", params.Address))
	}
	return cuAst, nil
}

// queryOpCUAstInfo query opreation custodian units with depositList,in settle format
func queryOpCUAstInfo(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryOpCUAstInfoParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	opcuinfo := keeper.GetOpCUsAstInfo(ctx, params.Symbol)
	if opcuinfo == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("opreation CU of:%s does not found", params.Symbol))
	}
	bz, err := keeper.cdc.MarshalJSON(opcuinfo)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryDepositExistance(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.DepositExistanceParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	exist := keeper.IsDepositExist(ctx, params.Symbol, params.Address, params.TxHash, params.Index)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, exist)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil

}
