package mapping

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/mapping/types"
)

// NewHandler returns a handler for token type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgMappingSwap:
			return handleMsgMappingSwap(ctx, keeper, msg)
		case MsgCreateFreeSwap:
			return handleMsgCreateFreeSwap(ctx, keeper, msg)
		case MsgCreateDirectSwap:
			return handleMsgCreateDirectSwap(ctx, keeper, msg)
		case MsgSwapSymbol:
			return handleMsgSwapSymbol(ctx, keeper, msg)
		case MsgCancelSwap:
			return handleMsgCancelSwap(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized token Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgMappingSwap(ctx sdk.Context, keeper Keeper, msg types.MsgMappingSwap) sdk.Result {
	ctx.Logger().Info("handleMsgMappingSwap", "msg", msg)

	mappingInfo := keeper.GetMappingInfo(ctx, msg.IssueSymbol)
	if mappingInfo == nil {
		return types.ErrMappingNotFound(DefaultCodespace, "mapping for issuer symbol is not found").Result()
	}
	if !mappingInfo.Enabled {
		return types.ErrInvalidSwapAmount(DefaultCodespace, "swap is disabled").Result()
	}

	issueTokenInfo := keeper.tk.GetToken(ctx, mappingInfo.IssueSymbol)
	if issueTokenInfo == nil {
		return sdk.ErrInvalidSymbol("issuer symbol does not exist").Result()
	}
	if !issueTokenInfo.IsSendEnabled() {
		return sdk.ErrInvalidSymbol("issuer symbol does not allow send").Result()
	}
	targetTokenInfo := keeper.tk.GetToken(ctx, mappingInfo.TargetSymbol)
	if targetTokenInfo == nil {
		return sdk.ErrInvalidSymbol("target symbol does not exist").Result()
	}
	if !targetTokenInfo.IsSendEnabled() {
		return sdk.ErrInvalidSymbol("target symbol does not allow send").Result()
	}

	issueChangeAmount := sdk.NewInt(0)
	var gotCoin, costCoin sdk.Coin
	if msg.Coins.AmountOf(mappingInfo.IssueSymbol.String()).IsPositive() {
		// swap from issue symbol
		needAmount := msg.Coins.AmountOf(mappingInfo.IssueSymbol.String())
		gotCoin = sdk.NewCoin(mappingInfo.TargetSymbol.String(), needAmount)
		costCoin = sdk.NewCoin(mappingInfo.IssueSymbol.String(), needAmount)
		issueChangeAmount = needAmount.Neg()
	} else if msg.Coins.AmountOf(mappingInfo.TargetSymbol.String()).IsPositive() {
		// swap from target symbol
		issueChangeAmount = msg.Coins.AmountOf(mappingInfo.TargetSymbol.String())
		gotCoin = sdk.NewCoin(mappingInfo.IssueSymbol.String(), issueChangeAmount)
		costCoin = sdk.NewCoin(mappingInfo.TargetSymbol.String(), issueChangeAmount)
	} else {
		return sdk.ErrInvalidCoins("coins do not match mapping").Result()
	}

	oldIssuePool := mappingInfo.IssuePool
	mappingInfo.IssuePool = mappingInfo.IssuePool.Sub(issueChangeAmount)
	if !mappingInfo.IssuePool.IsPositive() || mappingInfo.IssuePool.GT(mappingInfo.TotalSupply) {
		return types.ErrInvalidSwapAmount(DefaultCodespace, "invalid swap amount").Result()
	}
	keeper.SetMappingInfo(ctx, mappingInfo)

	var flows []sdk.Flow
	_, flow, err := keeper.trk.SubCoin(ctx, msg.From, costCoin)
	if err != nil {
		return err.Result()
	}
	flows = append(flows, flow)

	_, flow, err = keeper.trk.AddCoin(ctx, msg.From, gotCoin)
	if err != nil {
		return err.Result()
	}
	flows = append(flows, flow)

	flows = append(flows, MappingBalanceFlow{
		mappingInfo.IssueSymbol,
		oldIssuePool,
		issueChangeAmount.Neg(),
	})

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeMapping, flows)
	res := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &res)
	return res
}

func handleMsgCreateDirectSwap(ctx sdk.Context, keeper Keeper, msg types.MsgCreateDirectSwap) sdk.Result {
	ctx.Logger().Info("handleMsgCreateDirectSwap", "msg", msg)
	fromCUAddr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid from CU:%v", msg.From)).Result()
	}
	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount("from CU does not exist").Result()
	}

	if keeper.IsSwapOrderExist(ctx, msg.OrderID, types.SwapTypeDirect) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("free swap order exitst(%v)", msg.OrderID)).Result()
	}

	return keeper.CreateDirectSwapOrder(ctx, fromCUAddr, msg.SwapInfo, msg.OrderID)
}

func handleMsgCreateFreeSwap(ctx sdk.Context, keeper Keeper, msg types.MsgCreateFreeSwap) sdk.Result {
	ctx.Logger().Info("handleMsgCreateFreeSwap", "msg", msg)
	fromCUAddr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid from CU:%v", msg.From)).Result()
	}
	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount("from CU does not exist").Result()
	}

	if keeper.IsSwapOrderExist(ctx, msg.OrderID, types.SwapTypeFree) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("free swap order exitst(%v)", msg.OrderID)).Result()
	}

	return keeper.CreateFreeSwapOrder(ctx, fromCUAddr, msg.SwapInfo, msg.OrderID)
}

func handleMsgSwapSymbol(ctx sdk.Context, keeper Keeper, msg types.MsgSwapSymbol) sdk.Result {
	ctx.Logger().Info("handleMsgSwapSymbol", "msg", msg)
	fromCUAddr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid from CU:%v", msg.From)).Result()
	}
	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount("from CU does not exist").Result()
	}

	return keeper.SwapSymbol(ctx, fromCUAddr, msg.SwapType, msg.DstOrderID, msg.SwapAmount)
}

func handleMsgCancelSwap(ctx sdk.Context, keeper Keeper, msg types.MsgCancelSwap) sdk.Result {
	ctx.Logger().Info("handleMsgCancelSwap", "msg", msg)
	fromCUAddr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid from CU:%v", msg.From)).Result()
	}
	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount("from CU does not exist").Result()
	}

	return keeper.CancelSwap(ctx, fromCUAddr, msg.SwapType, msg.OrderID)
}
