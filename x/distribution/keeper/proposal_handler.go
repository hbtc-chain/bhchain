package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/distribution/types"
)

// HandleCommunityPoolSpendProposal is a handler for executing a passed community spend proposal
func HandleCommunityPoolSpendProposal(ctx sdk.Context, k Keeper, p types.CommunityPoolSpendProposal) sdk.Result {
	if k.blacklistedAddrs[p.Recipient.String()] {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is blacklisted from receiving external funds", p.Recipient)).Result()
	}

	err := k.DistributeFromFeePool(ctx, p.Amount, p.Recipient)
	if err != nil {
		return err.Result()
	}

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("transferred %s from the community pool to recipient %s", p.Amount, p.Recipient))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteCommunityPoolSpendProposal,
			sdk.NewAttribute(types.AttributeKeyRecipient, p.Recipient.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, p.Amount.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}
