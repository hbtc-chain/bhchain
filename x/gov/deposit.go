package gov

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/gov/types"
)

// GetDeposit gets the deposit of a specific depositor on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.CUAddress) (deposit Deposit, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.DepositKey(proposalID, depositorAddr))
	if bz == nil {
		return deposit, false
	}

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
	return deposit, true
}

func (keeper Keeper) setDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.CUAddress, deposit Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	store.Set(types.DepositKey(proposalID, depositorAddr), bz)
}

// AddDeposit adds or updates a deposit of a specific depositor on a specific proposal
// Activates voting period when appropriate
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.CUAddress, depositAmount sdk.Coins) (sdk.Result, sdk.Error, bool) {
	// Checks to see if proposal exists
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return sdk.Result{}, ErrUnknownProposal(keeper.codespace, proposalID), false
	}

	// Check if proposal is still depositable
	if (proposal.Status != StatusDepositPeriod) && (proposal.Status != StatusVotingPeriod) {
		return sdk.Result{}, ErrAlreadyFinishedProposal(keeper.codespace, proposalID), false
	}

	// update the governance module's CU coins pool
	result, err := keeper.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount)
	if err != nil {
		return result, err, false
	}

	// Update proposal
	proposal.TotalDeposit = proposal.TotalDeposit.Add(depositAmount)
	keeper.SetProposal(ctx, proposal)

	needDeposit := keeper.GetDepositParams(ctx).MinDaoDeposit
	if proposal.ProposalToken() == sdk.NativeToken {
		needDeposit = keeper.GetDepositParams(ctx).MinDeposit
	}
	// Check if deposit has provided sufficient total funds to transition the proposal into the voting period
	activatedVotingPeriod := false
	token := proposal.ProposalToken()
	if proposal.Status == StatusDepositPeriod && proposal.TotalDeposit.AmountOf(token).GTE(needDeposit.AmountOf(token)) {
		keeper.activateVotingPeriod(ctx, proposal)
		activatedVotingPeriod = true
	}

	// Add or update deposit object
	deposit, found := keeper.GetDeposit(ctx, proposalID, depositorAddr)
	if found {
		deposit.Amount = deposit.Amount.Add(depositAmount)
	} else {
		deposit = NewDeposit(proposalID, depositorAddr, depositAmount)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	keeper.setDeposit(ctx, proposalID, depositorAddr, deposit)
	return result, nil, activatedVotingPeriod
}

// GetAllDeposits returns all the deposits from the store
func (keeper Keeper) GetAllDeposits(ctx sdk.Context) (deposits Deposits) {
	keeper.IterateAllDeposits(ctx, func(deposit Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}

// GetDeposits returns all the deposits from a proposal
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID uint64) (deposits Deposits) {
	keeper.IterateDeposits(ctx, proposalID, func(deposit Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}

// GetDepositsIterator gets all the deposits on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetDepositsIterator(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.DepositsKey(proposalID))
}

// RefundDeposits refunds and deletes all the deposits on a specific proposal
func (keeper Keeper) RefundDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateDeposits(ctx, proposalID, func(deposit types.Deposit) bool {
		_, err := keeper.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, deposit.Depositor, deposit.Amount)
		if err != nil {
			panic(err)
		}

		store.Delete(DepositKey(proposalID, deposit.Depositor))
		return false
	})
}

//obseleted by Keep@10/26
// DeleteDeposits deletes all the deposits on a specific proposal without refunding them
//func (keeper Keeper) DeleteDeposits(ctx sdk.Context, proposalID uint64) {
//	store := ctx.KVStore(keeper.storeKey)
//
//	keeper.IterateDeposits(ctx, proposalID, func(deposit types.Deposit) bool {
//		err := keeper.supplyKeeper.BurnCoins(ctx, types.ModuleName, deposit.Amount)
//		if err != nil {
//			panic(err)
//		}
//
//		store.Delete(DepositKey(proposalID, deposit.Depositor))
//		return false
//	})
//}

// TransferDepositsToCommunityPool transfer all the deposits on a specific proposal to community pool
func (keeper Keeper) TransferDepositsToCommunityPool(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateDeposits(ctx, proposalID, func(deposit types.Deposit) bool {
		keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(deposit.Amount))
		store.Delete(DepositKey(proposalID, deposit.Depositor))
		return false
	})
}
