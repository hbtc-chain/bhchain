package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
)

// Keeper of the evidence store
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	router        types.Router
	paramSubSpace types.ParamSubspace
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a evidence keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace types.ParamSubspace, stakingKeeper types.StakingKeeper) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		paramSubSpace: paramSpace.WithKeyTable(types.ParamKeyTable()),
		stakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Vote(ctx sdk.Context, voteID string, voter sdk.CUAddress, vote exported.Vote, height uint64) (bool, bool, []*exported.VoteItem) {
	return k.VoteWithCustomBox(ctx, voteID, voter, vote, height, types.NewVoteBox)
}

func (k Keeper) VoteWithCustomBox(ctx sdk.Context, voteID string, voter sdk.CUAddress, vote exported.Vote, height uint64, newVoteBox exported.NewVoteBox) (bool, bool, []*exported.VoteItem) {
	voteBox := k.getVoteBox(ctx, voteID)
	if voteBox == nil {
		curEpoch := k.stakingKeeper.GetCurrentEpoch(ctx)
		voteBox = newVoteBox(sdk.Majority23(len(curEpoch.KeyNodeSet)))
	}
	firstConfirmed := voteBox.AddVote(voter, vote)
	if firstConfirmed {
		k.setConfirmedVote(ctx, voteID, height)
	}
	k.setVoteBox(ctx, voteID, voteBox)
	return firstConfirmed, voteBox.HasConfirmed(), voteBox.ValidVotes()
}

func (k Keeper) RecordMisbehaviourVoter(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	currentHeight := uint64(ctx.BlockHeight())
	if currentHeight <= types.BehaviourCountDelayBlock {
		return
	}

	iterator := store.ReverseIterator(types.ConfirmedVoteKey, types.GetConfirmedVoteKeyPrefix(currentHeight-types.BehaviourCountDelayBlock+1))
	toDeleted := make([][]byte, 0)
	for ; iterator.Valid(); iterator.Next() {
		height, voteID := types.DecodeConfirmedVoteKey(iterator.Key())
		epoch := k.stakingKeeper.GetEpochByHeight(ctx, height)
		voteBox := k.getVoteBox(ctx, voteID)
		validVotes := voteBox.ValidVotes()
		for _, validator := range epoch.KeyNodeSet {
			var found bool
			for _, item := range validVotes {
				if item.Voter.Equals(validator) {
					found = true
				}
			}
			k.HandleBehaviour(ctx, types.VoteBehaviourKey, sdk.ValAddress(validator), height, found)
		}
		toDeleted = append(toDeleted, iterator.Key())
	}
	iterator.Close()
	// delete after iterating finished
	for _, key := range toDeleted {
		store.Delete(key)
	}

}

func (k Keeper) getVoteBox(ctx sdk.Context, voteID string) exported.VoteBox {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetVoteBoxKey(voteID))
	if len(bz) == 0 {
		return nil
	}
	var voteBox exported.VoteBox
	k.cdc.MustUnmarshalBinaryBare(bz, &voteBox)
	return voteBox
}

func (k Keeper) setVoteBox(ctx sdk.Context, voteID string, voteBox exported.VoteBox) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(voteBox)
	store.Set(types.GetVoteBoxKey(voteID), bz)
}

func (k Keeper) setConfirmedVote(ctx sdk.Context, voteID string, height uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(append(types.GetConfirmedVoteKeyPrefix(height), voteID...), []byte{})
}
