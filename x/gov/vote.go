package gov

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/gov/types"
)

// AddVote Adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress, option VoteOption) sdk.Error {
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if proposal.Status != StatusVotingPeriod {
		return ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if !ValidVoteOption(option) {
		return ErrInvalidVote(keeper.codespace, option)
	}

	vote := NewVote(proposalID, voterAddr, option)
	keeper.setVote(ctx, proposalID, voterAddr, vote)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyOption, option.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return nil
}

func (keeper Keeper) AddDaoVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress, voteAmount sdk.Coins, option VoteOption) (sdk.Result, sdk.Error) {
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return sdk.Result{}, ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if proposal.Status != StatusVotingPeriod {
		return sdk.Result{}, ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if voteAmount[0].Denom != proposal.ProposalToken() {
		return sdk.Result{}, sdk.ErrInvalidSymbol(voteAmount[0].Denom)
	}

	if !ValidVoteOption(option) {
		return sdk.Result{}, ErrInvalidVote(keeper.codespace, option)
	}

	vote, found := keeper.GetDaoVote(ctx, proposalID, voterAddr)
	if !found {
		vote = NewDaoVote(proposalID, voterAddr, voteAmount, option)
	} else {
		if vote.Option != option {
			return sdk.Result{}, sdk.NewError(keeper.codespace, CodeInvalidVote, fmt.Sprintf("has already vote %s, cannot vote %s", vote.Option.String(), option.String()))
		}
		vote.VoteAmount = vote.VoteAmount.Add(voteAmount)
	}

	result, err := keeper.supplyKeeper.SendCoinsFromAccountToModule(ctx, voterAddr, types.ModuleName, voteAmount)
	if err != nil {
		return sdk.Result{}, sdk.ErrInvalidAmount(voteAmount.String())
	}

	keeper.setDaoVote(ctx, proposalID, voterAddr, vote)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalDaoVote,
			sdk.NewAttribute(types.AttributeKeyVoteAmount, voteAmount.String()),
			sdk.NewAttribute(types.AttributeKeyOption, option.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return result, nil
}

// GetAllVotes returns all the votes from the store
func (keeper Keeper) GetAllVotes(ctx sdk.Context) (votes Votes) {
	keeper.IterateAllVotes(ctx, func(vote Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVotes returns all the votes from a proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID uint64) (votes Votes) {
	keeper.IterateVotes(ctx, proposalID, func(vote Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVote gets the vote from an address on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress) (vote Vote, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.VoteKey(proposalID, voterAddr))
	if bz == nil {
		return vote, false
	}

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &vote)
	return vote, true
}

func (keeper Keeper) setVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress, vote Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(types.VoteKey(proposalID, voterAddr), bz)
}

// GetVotesIterator gets all the votes on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetVotesIterator(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.VotesKey(proposalID))
}

func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.VoteKey(proposalID, voterAddr))
}

// GetAllVotes returns all the votes from the store
func (keeper Keeper) GetAllDaoVotes(ctx sdk.Context) (votes DaoVotes) {
	keeper.IterateAllDaoVotes(ctx, func(vote DaoVote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVotes returns all the votes from a proposal
func (keeper Keeper) GetDaoVotes(ctx sdk.Context, proposalID uint64) (votes DaoVotes) {
	keeper.IterateDaoVotes(ctx, proposalID, func(vote DaoVote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVote gets the vote from an address on a specific proposal
func (keeper Keeper) GetDaoVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress) (vote DaoVote, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.DaoVoteKey(proposalID, voterAddr))
	if bz == nil {
		return vote, false
	}

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &vote)
	return vote, true
}

func (keeper Keeper) setDaoVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress, vote DaoVote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(types.DaoVoteKey(proposalID, voterAddr), bz)
}

// GetVotesIterator gets all the votes on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetDaoVotesIterator(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.DaoVotesKey(proposalID))
}

func (keeper Keeper) deleteDaoVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.CUAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.DaoVoteKey(proposalID, voterAddr))
}
