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

	if proposal.TokenInfo.WithdrawalFeeRate.LTE(sdk.OneDec()) || proposal.TokenInfo.GasLimit.LTE(sdk.ZeroInt()) {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s parame error", proposal.TokenInfo.Symbol)).Result()

	}
	//symbol already exist
	if ti := keeper.GetTokenInfo(ctx, proposal.TokenInfo.Symbol); ti != nil {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s already exist", proposal.TokenInfo.Symbol)).Result()
	}

	//chain does not exist, if symbol != chain
	if proposal.TokenInfo.Symbol.String() != proposal.TokenInfo.Chain.String() {
		if ti := keeper.GetTokenInfo(ctx, proposal.TokenInfo.Chain); ti == nil {
			return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s's chain %s does not exist", proposal.TokenInfo.Symbol, proposal.TokenInfo.Chain)).Result()
		}
	}

	keeper.SetTokenInfo(ctx, &proposal.TokenInfo)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteAddTokenProposal,
			sdk.NewAttribute(types.AttributeKeyToken, proposal.TokenInfo.Symbol.String()),
			sdk.NewAttribute(types.AttributeKeyTokeninfo, proposal.TokenInfo.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func processChangeParam(key, value string, ti *sdk.TokenInfo, cdc *codec.Codec) error {
	switch key {
	case sdk.KeyIsSendEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.IsSendEnabled = val

	case sdk.KeyIsDepositEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.IsDepositEnabled = val

	case sdk.KeyIsWithdrawalEnabled:
		val := false
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		ti.IsWithdrawalEnabled = val

	case sdk.KeyCollectThreshold:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.CollectThreshold = val

	case sdk.KeyDepositThreshold:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.DepositThreshold = val

	case sdk.KeyOpenFee:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.OpenFee = val

	case sdk.KeySysOpenFee:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.SysOpenFee = val

	case sdk.KeyWithdrawalFeeRate:
		val := sdk.ZeroDec()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}

		if val.LT(sdk.OneDec()) {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
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
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}

		ti.SysTransferNum = val

	case sdk.KeyOpCUSysTransferNum:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.OpCUSysTransferNum = val

	case sdk.KeyGasLimit:
		val := sdk.ZeroInt()
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val.IsNegative() {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.GasLimit = val
	case sdk.KeyConfirmations:
		var val uint64
		err := cdc.UnmarshalJSON([]byte(value), &val)
		if err != nil {
			return err
		}
		if val == 0 {
			return types.ErrInvalidParameter(DefaultParamspace, key, value)
		}
		ti.Confirmations = val

	default:
		return errors.New(fmt.Sprintf("Unkonwn parameter:%v", key))
	}

	return nil
}

func handleTokenParamsChangeProposal(ctx sdk.Context, keeper Keeper, proposal types.TokenParamsChangeProposal) sdk.Result {
	ctx.Logger().Info("handleTokenParamsChangeProposal", "proposal", proposal)

	ti := keeper.GetTokenInfo(ctx, sdk.Symbol(proposal.Symbol))
	if ti == nil {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s dose not exist", proposal.Symbol)).Result()
	}

	attr := []sdk.Attribute{}
	for _, pc := range proposal.Changes {
		err := processChangeParam(pc.Key, pc.Value, ti, keeper.cdc)
		if err != nil {
			return types.ErrInvalidParameter(types.DefaultCodespace, pc.Key, pc.Value).Result()
		}
		attr = append(attr, sdk.NewAttribute(types.AttributeKeyTokenParam, pc.Key), sdk.NewAttribute(types.AttributeKeyTokenParamValue, pc.Value))
	}

	keeper.SetTokenInfo(ctx, ti)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeExecuteTokenParamsChangeProposal, attr...),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleDisableTokenProposal(ctx sdk.Context, keeper Keeper, proposal types.DisableTokenProposal) sdk.Result {
	ctx.Logger().Info("handleDisableTokenProposal", "proposal", proposal)

	ti := keeper.GetTokenInfo(ctx, sdk.Symbol(proposal.Symbol))
	if ti == nil {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("token %s does not exist", proposal.Symbol)).Result()
	}

	ti.IsSendEnabled = false
	ti.IsDepositEnabled = false
	ti.IsWithdrawalEnabled = false

	keeper.SetTokenInfo(ctx, ti)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteDisableTokenProposal,
			sdk.NewAttribute(types.AttributeKeyToken, proposal.Symbol),
		),
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

		case types.DisableTokenProposal:
			return handleDisableTokenProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized token proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
