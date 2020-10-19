package openswap

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
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

func handleMsgAddLiquidity(ctx sdk.Context, k Keeper, msg types.MsgAddLiquidity) sdk.Result {
	if result := k.CheckSymbol(ctx, msg.TokenA); !result.IsOK() {
		return result
	}
	if result := k.CheckSymbol(ctx, msg.TokenB); !result.IsOK() {
		return result
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	return k.AddLiquidity(ctx, msg.From, msg.TokenA, msg.TokenB, msg.MinTokenAAmount, msg.MinTokenBAmount)
}

func handleMsgRemoveLiquidity(ctx sdk.Context, k Keeper, msg types.MsgRemoveLiquidity) sdk.Result {
	if result := k.CheckSymbol(ctx, msg.TokenA); !result.IsOK() {
		return result
	}
	if result := k.CheckSymbol(ctx, msg.TokenB); !result.IsOK() {
		return result
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	return k.RemoveLiquidity(ctx, msg.From, msg.TokenA, msg.TokenB, msg.Liquidity)
}

func handleMsgSwapExactIn(ctx sdk.Context, k Keeper, msg types.MsgSwapExactIn) sdk.Result {
	for _, token := range msg.SwapPath {
		if result := k.CheckSymbol(ctx, token); !result.IsOK() {
			return result
		}
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	referer := k.GetReferer(ctx, msg.From)
	if referer == nil {
		referer = msg.Referer
		k.BindReferer(ctx, msg.From, referer)
	}
	return k.SwapExactIn(ctx, msg.From, referer, msg.Receiver, msg.AmountIn, msg.MinAmountOut, msg.SwapPath)
}

func handleMsgSwapExactOut(ctx sdk.Context, k Keeper, msg types.MsgSwapExactOut) sdk.Result {
	for _, token := range msg.SwapPath {
		if result := k.CheckSymbol(ctx, token); !result.IsOK() {
			return result
		}
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	referer := k.GetReferer(ctx, msg.From)
	if referer == nil {
		referer = msg.Referer
		k.BindReferer(ctx, msg.From, referer)
	}
	return k.SwapExactOut(ctx, msg.From, referer, msg.Receiver, msg.AmountOut, msg.MaxAmountIn, msg.SwapPath)
}

func handleMsgLimitSwap(ctx sdk.Context, k Keeper, msg types.MsgLimitSwap) sdk.Result {
	if result := k.CheckSymbol(ctx, msg.BaseSymbol); !result.IsOK() {
		return result
	}
	if result := k.CheckSymbol(ctx, msg.QuoteSymbol); !result.IsOK() {
		return result
	}
	if msg.ExpiredAt > 0 && ctx.BlockTime().Unix() >= msg.ExpiredAt {
		return sdk.ErrInvalidTx("expired tx").Result()
	}
	order := k.GetOrder(ctx, msg.OrderID)
	if order != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("order %s already exists", msg.OrderID)).Result()
	}
	referer := k.GetReferer(ctx, msg.From)
	if referer == nil {
		referer = msg.Referer
		k.BindReferer(ctx, msg.From, referer)
	}
	return k.LimitSwap(ctx, msg.OrderID, msg.From, referer, msg.Receiver, msg.AmountIn, msg.Price,
		msg.BaseSymbol, msg.QuoteSymbol, msg.Side, msg.ExpiredAt)
}

func handleMsgCancelLimitSwap(ctx sdk.Context, k Keeper, msg types.MsgCancelLimitSwap) sdk.Result {
	return k.CancelOrders(ctx, msg.From, msg.OrderIDs)
}

func handleMsgClaimEarning(ctx sdk.Context, k Keeper, msg types.MsgClaimEarning) sdk.Result {
	return k.ClaimEarning(ctx, msg.From, msg.TokenA, msg.TokenB)
}
