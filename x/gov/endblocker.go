package gov

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/gov/types"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, keeper Keeper) {
	logger := keeper.Logger(ctx)

	// delete inactive proposal from store and its deposits
	keeper.IterateInactiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal Proposal) bool {
		keeper.DeleteProposal(ctx, proposal.ProposalID)
		keeper.TransferDepositsToCommunityPool(ctx, proposal.ProposalID)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeInactiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ProposalID)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, types.AttributeValueProposalDropped),
			),
		)

		logger.Info(
			fmt.Sprintf("proposal %d (%s) didn't meet minimum deposit of %s (had only %s); deleted",
				proposal.ProposalID,
				proposal.GetTitle(),
				keeper.GetDepositParams(ctx).MinDeposit,
				proposal.TotalDeposit,
			),
		)
		return false
	})

	// fetch active proposals whose voting periods have ended (are passed the block time)
	keeper.IterateActiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal Proposal) bool {
		var tagValue, logMsg string
		var passes, burnDeposits bool
		var tallyResults TallyResult

		if proposal.ProposalToken() == sdk.NativeToken {
			passes, burnDeposits, tallyResults = tally(ctx, keeper, proposal)
		} else {
			passes, burnDeposits, tallyResults = daotally(ctx, keeper, proposal)
		}

		if burnDeposits {
			keeper.TransferDepositsToCommunityPool(ctx, proposal.ProposalID)
		} else {
			keeper.RefundDeposits(ctx, proposal.ProposalID)
		}

		if passes {
			handler := keeper.router.GetRoute(proposal.ProposalRoute())
			cacheCtx, writeCache := ctx.CacheContext()

			// The proposal handler may execute state mutating logic depending
			// on the proposal content. If the handler fails, no state mutation
			// is written and the error message is logged.
			res := handler(cacheCtx, proposal.Content)
			if res.Code == sdk.CodeOK {
				proposal.Status = StatusPassed
				tagValue = types.AttributeValueProposalPassed
				logMsg = "passed"

				// write state to the underlying multi-store
				writeCache()
			} else {
				proposal.Status = StatusFailed
				tagValue = types.AttributeValueProposalFailed
				logMsg = fmt.Sprintf("passed, but failed on execution: %s", res.Log)
			}
		} else {
			proposal.Status = StatusRejected
			tagValue = types.AttributeValueProposalRejected
			logMsg = "rejected"
		}

		proposal.FinalTallyResult = tallyResults

		keeper.SetProposal(ctx, proposal)
		keeper.RemoveFromActiveProposalQueue(ctx, proposal.ProposalID, proposal.VotingEndTime)

		logger.Info(
			fmt.Sprintf(
				"proposal %d (%s) tallied; result: %s",
				proposal.ProposalID, proposal.GetTitle(), logMsg,
			),
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeActiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ProposalID)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, tagValue),
			),
		)
		return false
	})
}
