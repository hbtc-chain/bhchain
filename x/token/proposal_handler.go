package token

import (
	"errors"
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

func handleAddTokenProposal(ctx sdk.Context, keeper Keeper, proposal types.AddTokenProposal) sdk.Result {
	ctx.Logger().Info("handleAddTokenProposal", "proposal", proposal)

	tokenInfo := proposal.TokenInfo
	if tokenInfo.Chain == sdk.NativeToken {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("chain cannot be %s", tokenInfo.Chain)).Result()
	}

	if tokenInfo.Symbol != tokenInfo.Chain && tokenInfo.Symbol != types.CalSymbol(tokenInfo.Issuer, tokenInfo.Chain) {
		return sdk.ErrInvalidSymbol("invalid symbol").Result()
	}

	//symbol already exist
	if keeper.HasToken(ctx, tokenInfo.Symbol) {
		return sdk.ErrAlreadyExitSymbol(fmt.Sprintf("token symbol %s already exists", tokenInfo.Symbol)).Result()
	}

	//chain does not exist, if symbol != chain
	if tokenInfo.Symbol != tokenInfo.Chain {
		if !keeper.HasToken(ctx, proposal.TokenInfo.Chain) {
			return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s's chain %s does not exist", tokenInfo.Symbol, tokenInfo.Chain)).Result()
		}
	}

	err := keeper.CreateToken(ctx, tokenInfo)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteAddTokenProposal,
			sdk.NewAttribute(types.AttributeKeyToken, tokenInfo.Symbol.String()),
			sdk.NewAttribute(types.AttributeKeyTokeninfo, tokenInfo.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func processBaseTokenChangeParam(key, value string, ti *sdk.BaseToken, cdc *codec.Codec) error {
	switch key {
	case sdk.KeySendEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.SendEnabled = val
	default:
		return errors.New(fmt.Sprintf("Unkonwn parameter:%v for token %s", key, ti.Symbol))
	}

	return nil
}

func processIBCTokenChangeParam(key, value string, ti *sdk.IBCToken, cdc *codec.Codec) error {
	switch key {
	case sdk.KeySendEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.SendEnabled = val

	case sdk.KeyDepositEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.DepositEnabled = val

	case sdk.KeyWithdrawalEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.WithdrawalEnabled = val

	case sdk.KeyCollectThreshold:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.CollectThreshold = val

	case sdk.KeyDepositThreshold:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.DepositThreshold = val

	case sdk.KeyOpenFee:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.OpenFee = val

	case sdk.KeySysOpenFee:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.SysOpenFee = val

	case sdk.KeyWithdrawalFeeRate:
		val := sdk.ZeroDec()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}

		if val.LT(sdk.OneDec()) {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.WithdrawalFeeRate = val

	case sdk.KeyMaxOpCUNumber:
		val := uint64(0)
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.MaxOpCUNumber = val

	case sdk.KeySysTransferNum:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}

		ti.SysTransferNum = val

	case sdk.KeyOpCUSysTransferNum:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.OpCUSysTransferNum = val

	case sdk.KeyGasLimit:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.GasLimit = val
	case sdk.KeyConfirmations:
		var val uint64
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val == 0 {
			return types.ErrInvalidParameter(DefaultCodespace, key, value)
		}
		ti.Confirmations = val
	case sdk.KeyNeedCollectFee:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.NeedCollectFee = val

	default:
		return errors.New(fmt.Sprintf("Unkonwn parameter:%v for token %s", key, ti.Symbol))
	}

	return nil
}

func handleTokenParamsChangeProposal(ctx sdk.Context, keeper Keeper, proposal types.TokenParamsChangeProposal) sdk.Result {
	ctx.Logger().Info("handleTokenParamsChangeProposal", "proposal", proposal)

	ti := keeper.GetToken(ctx, sdk.Symbol(proposal.Symbol))
	if ti == nil {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s dose not exist", proposal.Symbol)).Result()
	}

	attr := []sdk.Attribute{}
	for _, pc := range proposal.Changes {
		var err error
		if ti.IsIBCToken() {
			err = processIBCTokenChangeParam(pc.Key, pc.Value, ti.(*sdk.IBCToken), keeper.cdc)
		} else {
			err = processBaseTokenChangeParam(pc.Key, pc.Value, ti.(*sdk.BaseToken), keeper.cdc)
		}
		if err != nil {
			return types.ErrInvalidParameter(types.DefaultCodespace, pc.Key, pc.Value).Result()
		}
		attr = append(attr, sdk.NewAttribute(types.AttributeKeyTokenParam, pc.Key), sdk.NewAttribute(types.AttributeKeyTokenParamValue, pc.Value))
	}

	keeper.SetToken(ctx, ti)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeExecuteTokenParamsChangeProposal, attr...),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func NewTokenProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Result {
		switch c := content.(type) {
		case types.AddTokenProposal:
			return handleAddTokenProposal(ctx, k, c)

		case types.TokenParamsChangeProposal:
			return handleTokenParamsChangeProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized token proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
