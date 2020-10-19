package evidence

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
)

// NewHandler creates an sdk.Handler for all the evidence type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case exported.MsgSubmitEvidence:
			return handleMsgSubmitEvidence(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized evidence message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitEvidence(ctx sdk.Context, k Keeper, msg exported.MsgSubmitEvidence) sdk.Result {
	evidence := msg.GetEvidence()
	if err := k.SubmitEvidence(ctx, evidence); err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSubmitter().String()),
		),
	)

	return sdk.Result{
		Data:   evidence.Hash(),
		Events: ctx.EventManager().Events(),
	}
}
