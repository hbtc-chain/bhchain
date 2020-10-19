package hrc20

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgNewToken:
			return handleMsgNewToken(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized hrc20 Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgNewToken(ctx sdk.Context, keeper Keeper, msg types.MsgNewToken) sdk.Result {
	ctx.Logger().Info("handleMsgNewToken", "msg", msg)

	if ti := keeper.tk.GetTokenInfo(ctx, msg.Symbol); ti != nil {
		return sdk.ErrAlreadyExitSymbol(fmt.Sprintf("token %s already exist", msg.Symbol)).Result()
	}
	reserved := keeper.tk.GetParams(ctx).ReservedSymbols
	for _, r := range reserved {
		if r == msg.Symbol.String() {
			return types.ErrSymbolReserved(DefaultCodespace, fmt.Sprintf("%v already reserved", msg.Symbol)).Result()
		}
	}

	from := keeper.ck.GetCU(ctx, msg.From)
	if from == nil {
		return sdk.ErrInvalidAccount(fmt.Sprintf("%v does not exist", msg.From)).Result()
	}

	issueFee := sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, keeper.GetParams(ctx).IssueTokenFee))
	if from.GetCoins().AmountOf(sdk.NativeToken).LT(keeper.GetParams(ctx).IssueTokenFee) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, have:%v", from.GetCoins(), issueFee)).Result()
	}

	totalSupply := msg.TotalSupply
	ti := sdk.TokenInfo{
		Symbol:              msg.Symbol,
		Chain:               sdk.NativeToken,
		Issuer:              msg.From.String(),
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            msg.Decimals,
		TotalSupply:         totalSupply,
		CollectThreshold:    sdk.ZeroInt(),
		DepositThreshold:    sdk.ZeroInt(),
		OpenFee:             sdk.ZeroInt(),
		SysOpenFee:          sdk.ZeroInt(),
		WithdrawalFeeRate:   sdk.ZeroDec(),
		MaxOpCUNumber:       0,
		SysTransferNum:      sdk.ZeroInt(),
		OpCUSysTransferNum:  sdk.ZeroInt(),
		GasLimit:            sdk.ZeroInt(),
		GasPrice:            sdk.ZeroInt(),
		Confirmations:       0,
		IsNonceBased:        true,
	}

	// transfer openFee to communityPool
	from.SubCoins(issueFee)
	keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(issueFee))
	keeper.ck.SetCU(ctx, from)
	flows := make([]sdk.Flow, 0, len(from.GetBalanceFlows()))
	for _, balanceFlow := range from.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}

	//set tokeninfo
	keeper.tk.SetTokenInfo(ctx, &ti)

	//minted newCoins
	mintedCoins := sdk.NewCoins(sdk.NewCoin(msg.Symbol.String(), totalSupply))
	err := keeper.sk.MintCoins(ctx, types.ModuleName, mintedCoins)
	if err != nil {
		return err.Result()
	}
	//allocate minted coins to receipt address
	toCU := keeper.ck.GetOrNewCU(ctx, sdk.CUTypeUser, msg.To)
	result, err := keeper.sk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toCU.GetAddress(), mintedCoins)
	if err != nil {
		return err.Result()
	}
	receipt, _ := keeper.rk.GetReceiptFromResult(&result)
	receipt.Category = sdk.CategoryTypeHrc20
	receipt.Flows = append(receipt.Flows, flows...)

	keeper.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewToken,
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, msg.To.String()),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, totalSupply.String()),
			sdk.NewAttribute(types.AttributeKeyIssueFee, issueFee.String()),
		),
	)
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}
