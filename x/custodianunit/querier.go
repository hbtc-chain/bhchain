package custodianunit

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
)

// creates a querier for auth REST endpoints
func NewQuerier(keeper CUKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryCU:
			return queryCU(ctx, req, keeper)
		case types.QueryCUWithChainAddr:
			return queryCUWithChainAddr(ctx, req, keeper)
		case types.QueryMultiChainAddrInfo:
			return queryMultiChainAddrInfo(ctx, req, keeper)
		case types.QueryMinimumGasPrice:
			return queryMinimumGasPrice(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cu query endpoint")
		}
	}
}

func queryCU(ctx sdk.Context, req abci.RequestQuery, keeper CUKeeper) ([]byte, sdk.Error) {
	cu, sdkerr := getCU(ctx, req, keeper)
	if sdkerr != nil {
		return nil, sdkerr
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, cu)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryCUWithChainAddr(ctx sdk.Context, req abci.RequestQuery, keeper CUKeeper) ([]byte, sdk.Error) {
	var params types.QueryCUChainAddressParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	cuAddress, err := keeper.GetCUFromExtAddress(ctx, params.Chain, params.Address)
	if err != nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("chainAddr of:%s, %s does not found", params.Chain, params.Address))
	}

	CU := keeper.GetCU(ctx, cuAddress)
	if CU == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("CU %s does not exist", cuAddress))
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, CU)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryMultiChainAddrInfo(ctx sdk.Context, req abci.RequestQuery, keeper CUKeeper) ([]byte, sdk.Error) {
	var params types.MultiQueryChainAddrInfoParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	resChainCUInfos := []sdk.ChainCUInfo{}
	for _, item := range params.ChainInfos {
		cuAddress, err := keeper.GetCUFromExtAddress(ctx, item.Chain, item.Address)
		if err != nil {
			continue
		}

		cu := keeper.GetCU(ctx, cuAddress)
		if cu == nil {
			continue
		}

		info := sdk.ChainCUInfo{
			Chain:       item.Chain,
			Addr:        item.Address,
			CuAddress:   cu.GetAddress().String(),
			IsChainAddr: true,
			IsOPCU:      false,
		}

		if cu.GetCUType() == sdk.CUTypeOp {
			info.IsOPCU = true
		}

		resChainCUInfos = append(resChainCUInfos, info)
	}

	bz, err := keeper.cdc.MarshalJSON(resChainCUInfos)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func getCU(ctx sdk.Context, req abci.RequestQuery, keeper CUKeeper) (exported.CustodianUnit, sdk.Error) {
	var params types.QueryCUParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	CU := keeper.GetCU(ctx, params.Address)
	if CU == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("CU %s does not exist", params.Address))
	}
	return CU, nil
}

func queryMinimumGasPrice(ctx sdk.Context, req abci.RequestQuery, keeper CUKeeper) ([]byte, sdk.Error) {
	minGasPrices := ctx.MinGasPrices()
	bz, err := codec.MarshalJSONIndent(keeper.cdc, minGasPrices)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil

}
