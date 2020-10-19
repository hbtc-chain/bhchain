package gov

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/gov/types"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)

		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)

		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)

		case MsgDaoVote:
			return handleMsgDaoVote(ctx, keeper, msg)

		case MsgCancelDaoVote:
			return handleMsgCancelDaoVote(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized gov message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {
	proposalToken := msg.Content.ProposalToken()
	if proposalToken == sdk.NativeToken {
		minInitDeposit := keeper.GetDepositParams(ctx).MinInitDeposit
		if msg.InitialDeposit.AmountOf(sdk.NativeToken).LT(minInitDeposit.AmountOf(sdk.NativeToken)) {
			errMsg := fmt.Sprintf("Init deposit %s Less than MiniInitDeposit %s", msg.InitialDeposit, minInitDeposit)
			return sdk.ErrInsufficientFunds(errMsg).Result()
		}
	} else {
		minInitDaoDeposit := keeper.GetDepositParams(ctx).MinDaoInitDeposit
		if msg.InitialDeposit.AmountOf(proposalToken).LT(minInitDaoDeposit.AmountOf(proposalToken)) {
			errMsg := fmt.Sprintf("Init deposit %s Less than MiniInitDeposit %s", msg.InitialDeposit, minInitDaoDeposit)
			return sdk.ErrInsufficientFunds(errMsg).Result()
		}
	}

	proposal, err := keeper.SubmitProposal(ctx, msg.Content, msg.VoteTime)
	if err != nil {
		return err.Result()
	}

	result, err, votingStarted := keeper.AddDeposit(ctx, proposal.ProposalID, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Proposer.String()),
		),
	)

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitProposal,
				sdk.NewAttribute(types.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", proposal.ProposalID)),
			),
		)
	}

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {
	result, err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositor, msg.Amount)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposalDeposit,
				sdk.NewAttribute(types.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", msg.ProposalID)),
			),
		)
	}
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {
	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}

}

func handleMsgDaoVote(ctx sdk.Context, keeper Keeper, msg MsgDaoVote) sdk.Result {
	result, err := keeper.AddDaoVote(ctx, msg.ProposalID, msg.Voter, msg.VoteAmount, msg.Option)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter.String()),
		),
	)
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result

}

func handleMsgCancelDaoVote(ctx sdk.Context, keeper Keeper, msg MsgCancelDaoVote) sdk.Result {
	vote, found := keeper.GetDaoVote(ctx, msg.ProposalID, msg.Voter)
	if !found {
		return sdk.ErrInvalidAddr(msg.Voter.String()).Result()
	}

	result, err := keeper.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, vote.Voter, vote.VoteAmount)
	if err != nil {
		return sdk.ErrInvalidAmount(vote.VoteAmount.String()).Result()
	}

	keeper.deleteVote(ctx, vote.ProposalID, vote.Voter)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter.String()),
		),
	)
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}
