/*
 * *******************************************************************
 * @项目名称: token
 * @文件名称: querier.go
 * @Date: 2019/06/06
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

package token

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryToken:
			return queryToken(ctx, req, keeper)
		case types.QuerySymbols:
			return querySymbols(ctx, keeper)
		case types.QueryDecimal:
			return queryDecimal(ctx, req, keeper)
		case types.QueryTokens:
			return queryTokens(ctx, keeper)
		case types.QueryParameters:
			return queryParams(ctx, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nameservice query endpoint")
		}
	}
}

//=====
func queryToken(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var ti QueryTokenInfo
	if err := keeper.cdc.UnmarshalJSON(req.Data, &ti); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	symbol := sdk.Symbol(ti.Symbol)
	token := keeper.GetTokenInfo(ctx, symbol)
	if token == nil {
		return nil, sdk.ErrUnknownRequest("Non-exits symbol")
	}

	res := tokenToResToken(*token)
	bz, err := keeper.cdc.MarshalJSON(res)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryTokens(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	tokens := keeper.GetAllTokenInfo(ctx)
	bz, err := keeper.cdc.MarshalJSON(tokens)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

//=====
func querySymbols(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	var symbols QueryResSymbols
	symbols = keeper.GetSymbols(ctx)
	bz, err := keeper.cdc.MarshalJSON(symbols)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryDecimal(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryDecimals
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	symbol := sdk.Symbol(params.Symbol)
	dec := keeper.GetDecimals(ctx, symbol)

	res := QueryResDecimals{
		Decimals: dec,
	}

	bz, err := keeper.cdc.MarshalJSON(res)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)
	res, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}
	return res, nil
}

func tokenToResToken(token sdk.TokenInfo) (resToken types.QueryResToken) {
	resToken = types.QueryResToken{
		Issuer:              token.Issuer,
		Chain:               token.Chain.String(),
		TokenType:           uint64(token.TokenType),
		IsSendEnabled:       token.IsSendEnabled,
		IsDepositEnabled:    token.IsDepositEnabled,
		IsWithdrawalEnabled: token.IsWithdrawalEnabled,
		Decimals:            token.Decimals,
		TotalSupply:         sdk.NewIntFromBigInt(token.TotalSupply.BigInt()),
		CollectThreshold:    sdk.NewIntFromBigInt(token.CollectThreshold.BigInt()),
		OpenFee:             sdk.NewIntFromBigInt(token.OpenFee.BigInt()),
		SysOpenFee:          sdk.NewIntFromBigInt(token.SysOpenFee.BigInt()),
		WithdrawalFeeRate:   token.WithdrawalFeeRate,
		Symbol:              token.Symbol,
		DepositThreshold:    token.DepositThreshold,
		MaxOpCUNumber:       token.MaxOpCUNumber,
		SysTransferNum:      token.SysTransferNum,
		OpCUSysTransferNum:  token.OpCUSysTransferNum,
		GasLimit:            token.GasLimit,
		GasPrice:            token.GasPrice,
		Confirmations:       token.Confirmations,
		IsNonceBased:        token.IsNonceBased,
	}
	return
}
