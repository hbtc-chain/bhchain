package openswap

import (
	"fmt"
	"strconv"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateDex:
			return handleMsgCreateDex(ctx, k, msg)
		case types.MsgEditDex:
			return handleMsgEditDex(ctx, k, msg)
		case types.MsgCreateTradingPair:
			return handleMsgCreateTradingPair(ctx, k, msg)
		case types.MsgEditTradingPair:
			return handleMsgEditTradingPair(ctx, k, msg)
		case types.MsgAddLiquidity:
			return handleMsgAddLiquidity(ctx, k, msg)
		case types.MsgRemoveLiquidity:
			return handleMsgRemoveLiquidity(ctx, k, msg)
		case types.MsgSwapExactIn:
			return handleMsgSwapExactIn(ctx, k, msg)
		case types.MsgSwapExactOut:
			return handleMsgSwapExactOut(ctx, k, msg)
		case types.MsgLimitSwap:
			return handleMsgLimitSwap(ctx, k, msg)
		case types.MsgCancelLimitSwap:
			return handleMsgCancelLimitSwap(ctx, k, msg)
		case types.MsgClaimEarning:
			return handleMsgClaimEarning(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized dex message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateDex(ctx sdk.Context, k Keeper, msg types.MsgCreateDex) sdk.Result {
	dex := &types.Dex{
		Name:           msg.Name,
		Owner:          msg.From,
		IncomeReceiver: msg.IncomeReceiver,
	}
	dex = k.SaveDex(ctx, dex)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeCreateDex,
			sdk.NewAttribute(types.AttributeKeyDexID, strconv.Itoa(int(dex.ID)))),
	})

	result := sdk.Result{}
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgEditDex(ctx sdk.Context, k Keeper, msg types.MsgEditDex) sdk.Result {
	dex := k.GetDex(ctx, msg.DexID)
	if dex == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("dex id %d not found", msg.DexID)).Result()
	}
	if !dex.Owner.Equals(msg.From) {
		return sdk.ErrInvalidTx(fmt.Sprintf("dex %d belongs to %s, not %s", msg.DexID, dex.Owner.String(), msg.From.String())).Result()
	}
	if msg.Name != "" {
		dex.Name = msg.Name
	}
	if msg.IncomeReceiver != nil {
		dex.IncomeReceiver = *msg.IncomeReceiver
	}

	k.SaveDex(ctx, dex)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeEditDex,
			sdk.NewAttribute(types.AttributeKeyDexID, strconv.Itoa(int(dex.ID)))),
	})

	result := sdk.Result{}
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgCreateTradingPair(ctx sdk.Context, k Keeper, msg types.MsgCreateTradingPair) sdk.Result {
	dex := k.GetDex(ctx, msg.DexID)
	if dex == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("dex id %d not found", msg.DexID)).Result()
	}
	if !dex.Owner.Equals(msg.From) {
		return sdk.ErrInvalidTx(fmt.Sprintf("dex %d belongs to %s, not %s", msg.DexID, dex.Owner.String(), msg.From.String())).Result()
	}
	tokenA, tokenB, result := k.SortTokens(ctx, msg.TokenA, msg.TokenB)
	if !result.IsOK() {
		return result
	}
	if k.GetTradingPair(ctx, msg.DexID, tokenA, tokenB) != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s-%s trading pair already exists in dex %d",
			tokenA, tokenB, msg.DexID)).Result()
	}

	if msg.IsPublic && msg.RefererRewardRate.LT(k.RefererTransactionBonusRate(ctx)) {
		return sdk.ErrInvalidTx(fmt.Sprintf("public pair's referer reward rate must be larger than %s",
			k.RefererTransactionBonusRate(ctx))).Result()
	}

	lpRewardRate := msg.LPRewardRate
	if msg.IsPublic {
		lpRewardRate = k.LpRewardRate(ctx)
	}
	if lpRewardRate.Add(msg.RefererRewardRate).Add(k.RepurchaseRate(ctx)).GT(k.MaxFeeRate(ctx)) {
		return sdk.ErrInvalidTx("sum of lp reward rate and referer reward rate is too large").Result()
	}
	pair := types.NewCustomTradingPair(msg.DexID, tokenA, tokenB, msg.IsPublic, msg.LPRewardRate, msg.RefererRewardRate)
	k.SaveTradingPair(ctx, pair)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeCreateTradingPair,
			sdk.NewAttribute(types.AttributeKeyDexID, strconv.Itoa(int(dex.ID))),
			sdk.NewAttribute(types.AttributeKeyTokenA, tokenA.String()),
			sdk.NewAttribute(types.AttributeKeyTokenB, tokenB.String()),
		),
	})

	result = sdk.Result{}
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgEditTradingPair(ctx sdk.Context, k Keeper, msg types.MsgEditTradingPair) sdk.Result {
	dex := k.GetDex(ctx, msg.DexID)
	if dex == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("dex id %d not found", msg.DexID)).Result()
	}
	if !dex.Owner.Equals(msg.From) {
		return sdk.ErrInvalidTx(fmt.Sprintf("dex %d belongs to %s, not %s", msg.DexID, dex.Owner.String(), msg.From.String())).Result()
	}
	tokenA, tokenB, result := k.SortTokens(ctx, msg.TokenA, msg.TokenB)
	if !result.IsOK() {
		return result
	}
	pair := k.GetTradingPair(ctx, msg.DexID, tokenA, tokenB)
	if pair == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s-%s trading pair does not exist in dex %d",
			tokenA, tokenB, msg.DexID)).Result()
	}

	if msg.IsPublic != nil {
		if !pair.IsPublic && pair.TotalLiquidity.IsPositive() {
			return sdk.ErrInvalidTx("cannot set pair public after adding liquidity").Result()
		}
		pair.IsPublic = *msg.IsPublic
	}
	if msg.LPRewardRate != nil {
		pair.LPRewardRate = *msg.LPRewardRate
	}
	if msg.RefererRewardRate != nil {
		pair.RefererRewardRate = *msg.RefererRewardRate
	}
	if pair.IsPublic && pair.RefererRewardRate.LT(k.RefererTransactionBonusRate(ctx)) {
		return sdk.ErrInvalidTx(fmt.Sprintf("public pair's referer reward rate must be larger than %s",
			k.RefererTransactionBonusRate(ctx))).Result()
	}
	lpRewardRate := pair.LPRewardRate
	if pair.IsPublic {
		lpRewardRate = k.LpRewardRate(ctx)
	}
	if lpRewardRate.Add(pair.RefererRewardRate).Add(k.RepurchaseRate(ctx)).GT(k.MaxFeeRate(ctx)) {
		return sdk.ErrInvalidTx("sum of lp reward rate and referer reward rate is too large").Result()
	}

	k.SaveTradingPair(ctx, pair)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeEditTradingPair,
			sdk.NewAttribute(types.AttributeKeyDexID, strconv.Itoa(int(dex.ID))),
			sdk.NewAttribute(types.AttributeKeyTokenA, tokenA.String()),
			sdk.NewAttribute(types.AttributeKeyTokenB, tokenB.String()),
		),
	})

	result = sdk.Result{}
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgAddLiquidity(ctx sdk.Context, k Keeper, msg types.MsgAddLiquidity) sdk.Result {
	tokenA, tokenB, result := k.SortTokens(ctx, msg.TokenA, msg.TokenB)
	if !result.IsOK() {
		return result
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	maxTokenAAmount, maxTokenBAmount := msg.MaxTokenAAmount, msg.MaxTokenBAmount
	if tokenA != msg.TokenA {
		maxTokenAAmount, maxTokenBAmount = maxTokenBAmount, maxTokenAAmount
	}
	return k.AddLiquidity(ctx, msg.From, msg.DexID, tokenA, tokenB, maxTokenAAmount, maxTokenBAmount)
}

func handleMsgRemoveLiquidity(ctx sdk.Context, k Keeper, msg types.MsgRemoveLiquidity) sdk.Result {
	tokenA, tokenB, result := k.SortTokens(ctx, msg.TokenA, msg.TokenB)
	if !result.IsOK() {
		return result
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	return k.RemoveLiquidity(ctx, msg.From, msg.DexID, tokenA, tokenB, msg.Liquidity)
}

func handleMsgSwapExactIn(ctx sdk.Context, k Keeper, msg types.MsgSwapExactIn) sdk.Result {
	for _, token := range msg.SwapPath {
		if _, result := k.CheckSymbol(ctx, token); !result.IsOK() {
			return result
		}
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}

	var referer sdk.CUAddress
	if msg.DexID != 0 {
		dex := k.GetDex(ctx, msg.DexID)
		if dex == nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("dex id %d not found", msg.DexID)).Result()
		}
		referer = dex.IncomeReceiver
	} else {
		referer = k.GetReferer(ctx, msg.From)
		if referer == nil {
			referer = msg.Referer
			k.BindReferer(ctx, msg.From, referer)
		}
	}
	return k.SwapExactIn(ctx, msg.DexID, msg.From, referer, msg.Receiver, msg.AmountIn, msg.MinAmountOut, msg.SwapPath)
}

func handleMsgSwapExactOut(ctx sdk.Context, k Keeper, msg types.MsgSwapExactOut) sdk.Result {
	for _, token := range msg.SwapPath {
		if _, result := k.CheckSymbol(ctx, token); !result.IsOK() {
			return result
		}
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}

	var referer sdk.CUAddress
	if msg.DexID != 0 {
		dex := k.GetDex(ctx, msg.DexID)
		if dex == nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("dex id %d not found", msg.DexID)).Result()
		}
		referer = dex.IncomeReceiver
	} else {
		referer = k.GetReferer(ctx, msg.From)
		if referer == nil {
			referer = msg.Referer
			k.BindReferer(ctx, msg.From, referer)
		}
	}
	return k.SwapExactOut(ctx, msg.DexID, msg.From, referer, msg.Receiver, msg.AmountOut, msg.MaxAmountIn, msg.SwapPath)
}

func handleMsgLimitSwap(ctx sdk.Context, k Keeper, msg types.MsgLimitSwap) sdk.Result {
	tokenA, tokenB, result := k.SortTokens(ctx, msg.BaseSymbol, msg.QuoteSymbol)
	if !result.IsOK() {
		return result
	}
	if tokenA != msg.BaseSymbol || tokenB != msg.QuoteSymbol {
		return sdk.ErrInvalidSymbol("wrong symbol sequence").Result()
	}

	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	order := k.GetOrder(ctx, msg.OrderID)
	if order != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("order %s already exists", msg.OrderID)).Result()
	}

	var referer sdk.CUAddress
	if msg.DexID != 0 {
		dex := k.GetDex(ctx, msg.DexID)
		if dex == nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("dex id %d not found", msg.DexID)).Result()
		}
		referer = dex.IncomeReceiver
	} else {
		referer = k.GetReferer(ctx, msg.From)
		if referer == nil {
			referer = msg.Referer
			k.BindReferer(ctx, msg.From, referer)
		}
	}
	return k.LimitSwap(ctx, msg.DexID, msg.OrderID, msg.From, referer, msg.Receiver, msg.AmountIn, msg.Price,
		msg.BaseSymbol, msg.QuoteSymbol, msg.Side, msg.ExpiredAt)
}

func handleMsgCancelLimitSwap(ctx sdk.Context, k Keeper, msg types.MsgCancelLimitSwap) sdk.Result {
	return k.CancelOrders(ctx, msg.From, msg.OrderIDs)
}

func handleMsgClaimEarning(ctx sdk.Context, k Keeper, msg types.MsgClaimEarning) sdk.Result {
	tokenA, tokenB, result := k.SortTokens(ctx, msg.TokenA, msg.TokenB)
	if !result.IsOK() {
		return result
	}
	return k.ClaimEarning(ctx, msg.From, msg.DexID, tokenA, tokenB)
}
