package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/distribution/types"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/staking/exported"
)

// AllocateTokens handles distribution of the collected fees
func (k Keeper) AllocateTokens(
	ctx sdk.Context, sumPreviousPrecommitPower, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, previousVotes []abci.VoteInfo,
) {

	logger := k.Logger(ctx)

	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.transferKeeper.GetAllBalance(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoins(feesCollectedInt)

	// transfer collected fees to the distribution module CustodianUnit
	_, err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		panic(err)
	}

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/hbtc-chain/bhchain/issues/2906#issuecomment-441867634
	feePool := k.GetFeePool(ctx)
	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected)
		k.SetFeePool(ctx, feePool)
		return
	}

	remaining := feesCollected
	// calculate previous core nodes reward
	keyNodes := k.stakingKeeper.GetCurrentEpoch(ctx).KeyNodeSet
	keyNodeReward := k.GetKeyNodeReward(ctx)
	var availableKeyNodes []staking.ValidatorI
	if len(keyNodes) > 0 && keyNodeReward.IsPositive() {
		for _, nodeAddress := range keyNodes {
			valAddress := sdk.ValAddress(nodeAddress)
			validator := k.stakingKeeper.Validator(ctx, valAddress)
			if validator == nil {
				logger.Error("key node validator not exist", "valAddress", valAddress)
				continue
			}
			if !validator.IsJailed() {
				availableKeyNodes = append(availableKeyNodes, validator)
			}
		}

		if len(availableKeyNodes) > 0 {
			totalKeyNodeReward := feesCollected.MulDecTruncate(keyNodeReward)
			perKeyNodeReward := totalKeyNodeReward.QuoDec(sdk.NewDec(int64(len(availableKeyNodes))))
			for _, validator := range availableKeyNodes {
				k.AllocateTokensToValidator(ctx, validator, perKeyNodeReward)
				remaining = remaining.Sub(perKeyNodeReward)
			}
		}
	}

	// calculate previous proposer reward
	baseProposerReward := k.GetBaseProposerReward(ctx)

	proposerReward := feesCollected.MulDecTruncate(baseProposerReward)
	// pay previous proposer

	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, previousProposer)

	if proposerValidator != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposerReward,
				sdk.NewAttribute(sdk.AttributeKeyAmount, proposerReward.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, proposerValidator.GetOperator().String()),
			),
		)

		k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
		remaining = remaining.Sub(proposerReward)
	} else {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
		logger.Error(fmt.Sprintf(
			"WARNING: Attempt to allocate proposer rewards to unknown proposer %s. "+
				"This should happen only if the proposer unbonded completely within a single block, "+
				"which generally should not happen except in exceptional circumstances (or fuzz testing). "+
				"We recommend you investigate immediately.",
			previousProposer.String()))
	}

	// allocate tokens proportionally to voting power
	votersReward := remaining
	// TODO consider parallelizing later, ref https://github.com/hbtc-chain/bhchain/pull/3099#discussion_r246276376
	for _, vote := range previousVotes {
		if !vote.SignedLastBlock {
			continue
		}
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)
		if validator == nil {
			logger.Error("key node validator not exist", "consAddress", vote.Validator.Address)
			continue
		}

		// TODO consider microslashing for missing votes.
		// ref https://github.com/hbtc-chain/bhchain/issues/2525#issuecomment-430838701
		powerFraction := sdk.NewDec(vote.Validator.Power).QuoTruncate(sdk.NewDec(sumPreviousPrecommitPower))
		reward := votersReward.MulDecTruncate(powerFraction)
		k.AllocateTokensToValidator(ctx, validator, reward)
		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining)
	k.SetFeePool(ctx, feePool)
}

func (k Keeper) AllocateTokensToValidatorWithoutShare(ctx sdk.Context, val exported.ValidatorI, tokens sdk.DecCoins) {
	commission := tokens
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)

	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(commission)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	outstanding = outstanding.Add(tokens)
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}

// AllocateTokensToValidator allocate tokens to a particular validator, splitting according to commission
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val exported.ValidatorI, tokens sdk.DecCoins) {
	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)

	// update current commission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(commission)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	currentRewards.Rewards = currentRewards.Rewards.Add(shared)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)

	// update outstanding rewards
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, tokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	outstanding = outstanding.Add(tokens)
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}
