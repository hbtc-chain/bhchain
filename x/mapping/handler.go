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

	fromCUAddr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddr(fmt.Sprintf("invalid from CU:%v", msg.From)).Result()
	}
	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount("from CU does not exist").Result()
	}

	mappingInfo := keeper.GetMappingInfo(ctx, msg.IssueSymbol)
	if mappingInfo == nil {
		return types.ErrMappingNotFound(DefaultCodespace, "mapping for issuer symbol is not found").Result()
	}
	if !mappingInfo.Enabled {
		return types.ErrInvalidSwapAmount(DefaultCodespace, "swap is disabled").Result()
	}

	if !keeper.tk.IsTokenSupported(ctx, mappingInfo.IssueSymbol) {
		return sdk.ErrUnSupportToken(mappingInfo.IssueSymbol.String()).Result()
	}
	issueTokenInfo := keeper.tk.GetTokenInfo(ctx, mappingInfo.IssueSymbol)
	if issueTokenInfo == nil {
		return sdk.ErrInvalidSymbol("issuer symbol does not exist").Result()
	}
	if !issueTokenInfo.IsSendEnabled {
		return sdk.ErrInvalidSymbol("issuer symbol does not allow send").Result()
	}
	if !keeper.tk.IsTokenSupported(ctx, mappingInfo.TargetSymbol) {
		return sdk.ErrUnSupportToken(mappingInfo.TargetSymbol.String()).Result()
	}
	targetTokenInfo := keeper.tk.GetTokenInfo(ctx, mappingInfo.TargetSymbol)
	if targetTokenInfo == nil {
		return sdk.ErrInvalidSymbol("target symbol does not exist").Result()
	}
	if !targetTokenInfo.IsSendEnabled {
		return sdk.ErrInvalidSymbol("target symbol does not allow send").Result()
	}

	issueChangeAmount := sdk.NewInt(0)
	newCoins := sdk.NewCoins()
	needCoins := sdk.NewCoins()
	have := fromCU.GetCoins()
	if msg.Coins.AmountOf(issueTokenInfo.Symbol.String()).IsPositive() {
		// swap from issue symbol
		needAmount := msg.Coins.AmountOf(issueTokenInfo.Symbol.String())
		if have.AmountOf(issueTokenInfo.Symbol.String()).LT(needAmount) {
			return sdk.ErrInsufficientCoins("from CU does not have sufficient coins").Result()
		}
		newCoins = sdk.NewCoins(sdk.NewCoin(targetTokenInfo.Symbol.String(), needAmount))
		needCoins = sdk.NewCoins(sdk.NewCoin(issueTokenInfo.Symbol.String(), needAmount))
		issueChangeAmount = needAmount.Neg()
	} else if msg.Coins.AmountOf(targetTokenInfo.Symbol.String()).IsPositive() {
		// swap from target symbol
		issueChangeAmount = msg.Coins.AmountOf(targetTokenInfo.Symbol.String())
		if have.AmountOf(targetTokenInfo.Symbol.String()).LT(issueChangeAmount) {
			return sdk.ErrInsufficientCoins("from CU does not have sufficient coins").Result()
		}
		newCoins = sdk.NewCoins(sdk.NewCoin(issueTokenInfo.Symbol.String(), issueChangeAmount))
		needCoins = sdk.NewCoins(sdk.NewCoin(targetTokenInfo.Symbol.String(), issueChangeAmount))
	} else {
		return sdk.ErrInvalidCoins("coins do not match mapping").Result()
	}

	oldIssuePool := mappingInfo.IssuePool
	mappingInfo.IssuePool = mappingInfo.IssuePool.Sub(issueChangeAmount)
	if !mappingInfo.IssuePool.IsPositive() || mappingInfo.IssuePool.GT(mappingInfo.TotalSupply) {
		return types.ErrInvalidSwapAmount(DefaultCodespace, "invalid swap amount").Result()
	}

	fromCU.ResetBalanceFlows()
	fromCU.SubCoins(needCoins)
	fromCU.AddCoins(newCoins)

	keeper.SetMappingInfo(ctx, mappingInfo)
	keeper.ck.SetCU(ctx, fromCU)

	var flows []sdk.Flow
	for _, balanceFlow := range fromCU.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}
	flows = append(flows, MappingBalanceFlow{
		mappingInfo.IssueSymbol,
		oldIssuePool,
		issueChangeAmount.Neg(),
	})

	fromCU.ResetBalanceFlows()

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
