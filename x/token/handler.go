package token

import (
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {

		case types.MsgSynGasPrice:
			return handleMsgSynGasPrice(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized token Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSynGasPrice(ctx sdk.Context, keeper Keeper, msg types.MsgSynGasPrice) sdk.Result {
	ctx.Logger().Info("handleMsgSynGasFee", "msg", msg)
	updatedGasPrice, result := keeper.SynGasPrice(ctx, msg.From, msg.Height, msg.GasPrice)
	if result.Code != sdk.CodeOK {
		return result
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSyncGasPrice,
			sdk.NewAttribute(types.AttributeKeyFrom, msg.From),
			sdk.NewAttribute(types.AttributeKeyHeight, sdk.NewInt(int64(msg.Height)).String()),
			sdk.NewAttribute(types.AttributeKeyGasPrice, gasPriceString(updatedGasPrice)),
		),
	)

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func gasPriceString(gasPrices []sdk.TokensGasPrice) string {
	var b strings.Builder

	for _, gs := range gasPrices {
		b.WriteString(gs.String())
		b.WriteString(",")
	}
	return b.String()
}
