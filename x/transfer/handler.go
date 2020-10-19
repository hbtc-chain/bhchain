package transfer

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/keeper"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

// NewHandler returns a handler for "transfer" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, k, msg)

		case MsgMultiSend:
			return handleMsgMultiSend(ctx, k, msg)

		case MsgDeposit:
			return handleMsgDeposit(ctx, k, msg)

		case MsgConfirmedDeposit:
			return handleMsgConfirmedDeposit(ctx, k, msg)

		case MsgCollectWaitSign:
			return handleMsgCollectWaitSign(ctx, k, msg)

		case MsgCollectSignFinish:
			return handleMsgCollectSignFinish(ctx, k, msg)

		case MsgCollectFinish:
			return handleMsgCollectFinish(ctx, k, msg)

		case MsgWithdrawal:
			return handleMsgWithdrawal(ctx, k, msg)

		case MsgWithdrawalConfirm:
			return handleMsgWithdrawalConfirm(ctx, k, msg)

		case MsgWithdrawalWaitSign:
			return handleMsgWithdrawalWaitSign(ctx, k, msg)

		case MsgWithdrawalSignFinish:
			return handleMsgWithdrawalSignFinish(ctx, k, msg)

		case MsgWithdrawalFinish:
			return handleMsgWithdrawalFinish(ctx, k, msg)

		case MsgSysTransfer:
			return handleMsgSysTransfer(ctx, k, msg)

		case MsgSysTransferWaitSign:
			return handleMsgSysTransferWaitSign(ctx, k, msg)

		case MsgSysTransferSignFinish:
			return handleMsgSysTransferSignFinish(ctx, k, msg)

		case MsgSysTransferFinish:
			return handleMsgSysTransferFinish(ctx, k, msg)

		case MsgOpcuAssetTransfer:
			return handleMsgOpcuAssetTransfer(ctx, k, msg)

		case MsgOpcuAssetTransferWaitSign:
			return handleMsgOpcuAssetTransferWaitSign(ctx, k, msg)

		case MsgOpcuAssetTransferSignFinish:
			return handleMsgOpcuAssetTransferSignFinish(ctx, k, msg)

		case MsgOpcuAssetTransferFinish:
			return handleMsgOpcuAssetTransferFinish(ctx, k, msg)

		case MsgOrderRetry:
			return handleMsgOrderRetry(ctx, k, msg)

		case MsgCancelWithdrawal:
			return handleMsgCancelWithdrawal(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized bank message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k keeper.Keeper, msg types.MsgSend) sdk.Result {
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	if k.BlacklistedAddr(msg.ToAddress) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.ToAddress)).Result()
	}

	result, err := k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

// Handle MsgMultiSend.
func handleMsgMultiSend(ctx sdk.Context, k keeper.Keeper, msg types.MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	if msg.MaxHeight != 0 && uint64(ctx.BlockHeight()) > msg.MaxHeight {
		return sdk.ErrInvalidTx(fmt.Sprintf("height（%d) is higher than max height(%d)", ctx.BlockHeight(), msg.MaxHeight)).Result()
	}

	for _, out := range msg.Outputs {
		if k.BlacklistedAddr(out.Address) {
			return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", out.Address)).Result()
		}
	}

	result, err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgDeposit(ctx sdk.Context, k keeper.Keeper, msg MsgDeposit) sdk.Result {
	ctx.Logger().Info("handleMsgDeposit ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCU, toCU := msg.FromCU, msg.ToCU
	if k.BlacklistedAddr(fromCU) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to send transactions", msg.FromCU)).Result()
	}
	if k.BlacklistedAddr(toCU) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.ToCU)).Result()
	}

	result := k.Deposit(ctx, fromCU, toCU, msg.Symbol, msg.ToAddress, msg.Txhash, uint64(msg.Index), msg.Amount, msg.OrderID, msg.Memo)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
			sdk.NewAttribute(types.AttributeKeySender, fromCU.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, toCU.String()),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol.String()),
			sdk.NewAttribute(types.AttributeKeyHash, msg.Txhash),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyIndex, strconv.Itoa(int(msg.Index))),
			sdk.NewAttribute(types.AttributeKeyMemo, msg.Memo),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)

	return result
}

func handleMsgConfirmedDeposit(ctx sdk.Context, k keeper.Keeper, msg MsgConfirmedDeposit) sdk.Result {
	ctx.Logger().Info("handleMsgConfirmedDeposit ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.ConfirmedDeposit(ctx, msg.From, msg.ValidOrderIDs, msg.InvalidOrderIDs)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDepositConfirm,
			sdk.NewAttribute(types.AttributeKeySender, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyValidOrderIDs, strings.Join(msg.ValidOrderIDs, ",")),
			sdk.NewAttribute(types.AttributeKeyInvalidOrderIDs, strings.Join(msg.InvalidOrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgCollectWaitSign(ctx sdk.Context, k keeper.Keeper, msg MsgCollectWaitSign) sdk.Result {
	ctx.Logger().Info("handleMsgCollectWaitSign ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	toCU, err := sdk.CUAddressFromBase58(msg.CollectToCU)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.CollectToCU)).Result()
	}

	if k.BlacklistedAddr(toCU) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.CollectToCU)).Result()
	}

	result := k.CollectWaitSign(ctx, toCU, msg.OrderIDs, msg.RawData)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCollectWaitSign,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
			sdk.NewAttribute(types.AttributeKeyRecipient, msg.CollectToCU),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)

	return result
}

func handleMsgCollectSignFinish(ctx sdk.Context, k keeper.Keeper, msg MsgCollectSignFinish) sdk.Result {
	ctx.Logger().Info("handleMsgCollectSignFinish ", "msg", msg, "signedTx", hex.EncodeToString(msg.SignedTx), "hash", msg.TxHash)

	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.CollectSignFinish(ctx, msg.OrderIDs, msg.SignedTx, msg.TxHash)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCollectSignFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)

	return result
}

func handleMsgCollectFinish(ctx sdk.Context, k keeper.Keeper, msg MsgCollectFinish) sdk.Result {
	ctx.Logger().Info("handleMsgCollectFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCU, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.Validator)).Result()
	}

	result := k.CollectFinish(ctx, fromCU, msg.OrderIDs, msg.CostFee)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCollectFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)

	return result
}

func handleMsgWithdrawal(ctx sdk.Context, k keeper.Keeper, msg MsgWithdrawal) sdk.Result {
	ctx.Logger().Info("handleMsgWithdrawal ", "msg", msg)

	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCUAddr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.FromCU)).Result()
	}

	if k.BlacklistedAddr(fromCUAddr) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.FromCU)).Result()
	}

	result := k.Withdrawal(ctx, fromCUAddr, msg.ToMultisignAddress, msg.OrderID, msg.Symbol, msg.Amount, msg.GasFee)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawal,
			sdk.NewAttribute(types.AttributeKeySender, msg.FromCU),
			sdk.NewAttribute(types.AttributeKeyRecipient, msg.ToMultisignAddress),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgWithdrawalConfirm(ctx sdk.Context, k keeper.Keeper, msg MsgWithdrawalConfirm) sdk.Result {
	ctx.Logger().Info("handleMsgWithdrawalConfirm", "msg", msg)

	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCUAddr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.FromCU)).Result()
	}

	if k.BlacklistedAddr(fromCUAddr) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.FromCU)).Result()
	}

	result := k.WithdrawalConfirm(ctx, fromCUAddr, msg.OrderID, msg.Valid)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawalConfirm,
			sdk.NewAttribute(types.AttributeKeySender, msg.FromCU),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgWithdrawalWaitSign(ctx sdk.Context, k keeper.Keeper, msg MsgWithdrawalWaitSign) sdk.Result {
	ctx.Logger().Info("handleMsgWithdrawalWaitSign ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	opCU, _ := sdk.CUAddressFromBase58(msg.OpCU)
	result := k.WithdrawalWaitSign(ctx, opCU, msg.OrderIDs, msg.SignHashes, msg.RawData)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawalWaitSign,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgWithdrawalSignFinish(ctx sdk.Context, k keeper.Keeper, msg MsgWithdrawalSignFinish) sdk.Result {
	ctx.Logger().Info("handleMsgWithdrawalSignFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.WithdrawalSignFinish(ctx, msg.OrderIDs, msg.SignedTx, msg.TxHash)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawalSignFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgWithdrawalFinish(ctx sdk.Context, k keeper.Keeper, msg MsgWithdrawalFinish) sdk.Result {
	ctx.Logger().Info("handleMsgWithdrawalFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCU, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.Validator)).Result()
	}

	result := k.WithdrawalFinish(ctx, fromCU, msg.OrderIDs, msg.CostFee, msg.Valid)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawalFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgSysTransfer(ctx sdk.Context, k keeper.Keeper, msg MsgSysTransfer) sdk.Result {
	ctx.Logger().Info("handleMsgSysTransfer ", "msg", msg)

	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	if k.BlacklistedAddr(msg.FromCU) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.FromCU.String())).Result()
	}

	if k.BlacklistedAddr(msg.ToCU) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.ToCU.String())).Result()
	}

	result := k.SysTransfer(ctx, msg.FromCU, msg.ToCU, msg.ToAddress, msg.OrderID, msg.Symbol)
	if result.Code != sdk.CodeOK {
		return result
	}

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgSysTransferWaitSign(ctx sdk.Context, k keeper.Keeper, msg MsgSysTransferWaitSign) sdk.Result {
	ctx.Logger().Info("handleMsgSysTransferWaitSign ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.SysTransferWaitSign(ctx, msg.OrderID, msg.SignHash, msg.RawData)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSysTransferWaitSign,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgSysTransferSignFinish(ctx sdk.Context, k keeper.Keeper, msg MsgSysTransferSignFinish) sdk.Result {
	ctx.Logger().Info("handleSysTransferSignFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.SysTransferSignFinish(ctx, msg.OrderID, msg.SignedTx)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSysTransferSignFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgSysTransferFinish(ctx sdk.Context, k keeper.Keeper, msg MsgSysTransferFinish) sdk.Result {
	ctx.Logger().Info("handleSysTransferFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCU, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.Validator)).Result()
	}

	result := k.SysTransferFinish(ctx, fromCU, msg.OrderID, msg.CostFee)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSysTransferFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgOpcuAssetTransfer(ctx sdk.Context, k keeper.Keeper, msg MsgOpcuAssetTransfer) sdk.Result {
	ctx.Logger().Info("handleMsgOpcuAssetTransfer ", "msg", msg)

	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCUAddr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.FromCU)).Result()
	}

	opCUAddr, err := sdk.CUAddressFromBase58(msg.OpCU)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.FromCU)).Result()
	}

	if k.BlacklistedAddr(fromCUAddr) {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.FromCU)).Result()
	}

	result := k.OpcuAssetTransfer(ctx, opCUAddr, msg.ToAddr, msg.OrderID, msg.Symbol, msg.TransferItems)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOpcuTransfer,
			sdk.NewAttribute(types.AttributeKeySender, msg.FromCU),
			sdk.NewAttribute(types.AttributeKeyRecipient, msg.ToAddr),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgOpcuAssetTransferWaitSign(ctx sdk.Context, k keeper.Keeper, msg MsgOpcuAssetTransferWaitSign) sdk.Result {
	ctx.Logger().Info("handleMsgOpcuTransferWaitSign ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.OpcuAssetTransferWaitSign(ctx, msg.OrderID, msg.SignHashes, msg.RawData)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOpcuTransferWaitSign,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgOpcuAssetTransferSignFinish(ctx sdk.Context, k keeper.Keeper, msg MsgOpcuAssetTransferSignFinish) sdk.Result {
	ctx.Logger().Info("handleMsgOpcuTransferSignFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	result := k.OpcuAssetTransferSignFinish(ctx, msg.OrderID, msg.SignedTx, msg.TxHash)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOpcuTransferSignFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgOpcuAssetTransferFinish(ctx sdk.Context, k keeper.Keeper, msg MsgOpcuAssetTransferFinish) sdk.Result {
	ctx.Logger().Info("handleMsgOpcuTransferFinish ", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCU, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.Validator)).Result()
	}

	result := k.OpcuAssetTransferFinish(ctx, fromCU, msg.OrderID, msg.CostFee)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOpcuTransferFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgOrderRetry(ctx sdk.Context, k keeper.Keeper, msg MsgOrderRetry) sdk.Result {
	ctx.Logger().Info("handleMsgOrderRetry", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCU, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.From)).Result()
	}

	result := k.OrderRetry(ctx, fromCU, msg.OrderIDs, msg.RetryTimes, msg.Evidences)
	if result.Code != sdk.CodeOK {
		return result
	}

	//Add events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOrderRetry,
			sdk.NewAttribute(types.AttributeKeySender, msg.From),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgCancelWithdrawal(ctx sdk.Context, k keeper.Keeper, msg MsgCancelWithdrawal) sdk.Result {
	ctx.Logger().Info("handleMsgCancelWithdrawal", "msg", msg)
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	fromCUAddr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid to CU:%v", msg.FromCU)).Result()
	}

	result := k.CancelWithdrawal(ctx, fromCUAddr, msg.OrderID)
	if result.Code != sdk.CodeOK {
		return result
	}

	//Add events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelWithdrawal,
			sdk.NewAttribute(types.AttributeKeySender, msg.FromCU),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}
