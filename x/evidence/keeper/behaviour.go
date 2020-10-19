package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
)

func (k Keeper) HandleBehaviour(ctx sdk.Context, behaviourKey string, validator sdk.ValAddress, height uint64, normal bool) {
	// check if validator is not unboned
	// fetch behavior info
	validatorBehavior := k.GetValidatorBehaviour(ctx, behaviourKey, validator)

	// this is a relative index, so it counts blocks the validator *should* have signed
	// will use the 0-value default signing info if not present, except for start height
	index := validatorBehavior.IndexOffset % k.BehaviourWindow(ctx, behaviourKey)
	validatorBehavior.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := k.GetValidatorBehaviourBitArray(ctx, behaviourKey, validator, index)
	misbehaved := !normal
	switch {
	case !previous && misbehaved:
		// Array value has changed from not misbehaved to misbehaved, increment counter
		k.SetValidatorBehaviourBitArray(ctx, behaviourKey, validator, index, true)
		validatorBehavior.MisbehaviourCounter++
	case previous && !misbehaved:
		// Array value has changed from misbehaved to not misbehaved, decrement counter
		k.SetValidatorBehaviourBitArray(ctx, behaviourKey, validator, index, false)
		validatorBehavior.MisbehaviourCounter--
	default:
		// Array value at this index has not changed, no need to update counter
	}

	// get max missed
	if validatorBehavior.MisbehaviourCounter > k.MaxMisbehaviourCount(ctx, behaviourKey) {
		k.stakingKeeper.SlashByOperator(ctx, validator, int64(height), k.BehaviourSlashFraction(ctx, behaviourKey))
		k.stakingKeeper.JailByOperator(ctx, validator)
		// We need to reset the counter & array so that the validator won't be immediately slashed for downtime upon rebonding.
		validatorBehavior.MisbehaviourCounter = 0
		validatorBehavior.IndexOffset = 0
		k.clearValidatorBehaviourBitArray(ctx, behaviourKey, validator)
	}

	k.SetValidatorBehaviour(ctx, behaviourKey, validator, validatorBehavior)
}

func (k Keeper) GetValidatorBehaviour(ctx sdk.Context, behaviourName string, address sdk.ValAddress) (validatorBehaviour types.ValidatorBehaviour) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorBehaviourKey(behaviourName, address))
	if bz == nil {
		return types.ValidatorBehaviour{}
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &validatorBehaviour)
	return
}

func (k Keeper) SetValidatorBehaviour(ctx sdk.Context, behaviourName string, address sdk.ValAddress, behavior types.ValidatorBehaviour) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&behavior)
	key := types.GetValidatorBehaviourKey(behaviourName, address)
	if bz == nil {
		store.Delete(key)
	} else {
		store.Set(key, bz)
	}
}

func (k Keeper) GetValidatorBehaviourBitArray(ctx sdk.Context, behaviourName string, address sdk.ValAddress, index int64) (mis bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorBehaviourBitArrayKey(behaviourName, address, index))
	if bz == nil {
		// lazy: treat empty key as not missed
		mis = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &mis)
	return
}

func (k *Keeper) SetValidatorBehaviourBitArray(ctx sdk.Context, behaviourName string, address sdk.ValAddress, index int64, mis bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(mis)
	store.Set(types.GetValidatorBehaviourBitArrayKey(behaviourName, address, index), bz)
}

func (k Keeper) clearValidatorBehaviourBitArray(ctx sdk.Context, behaviourName string, address sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetValidatorBehaviourBitArrayPrefixKey(behaviourName, address))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
