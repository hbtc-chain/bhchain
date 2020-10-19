package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

func (k Keeper) ReachJailQueueLimit(ctx sdk.Context) bool {
	keyNodeCount := len(k.GetCurrentEpoch(ctx).KeyNodeSet)
	if keyNodeCount == 0 {
		return false
	}
	jailInfo := k.getJailQueueInfo(ctx)
	oneSixth := sdk.OneSixthCeil(uint16(keyNodeCount))
	return jailInfo.KeyNodeCount >= oneSixth
}

func (k Keeper) insertJailedQueue(ctx sdk.Context, val types.Validator) uint64 {
	jailInfo := k.increaseJailQueueInfo(ctx, val)

	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetJailedQueueKey(jailInfo.Index), val.GetOperator())
	return jailInfo.Index
}

func (k Keeper) deleteFromJailedQueue(ctx sdk.Context, val types.Validator) bool {
	store := ctx.KVStore(k.storeKey)

	key := types.GetJailedQueueKey(val.JailedIndex)
	got := store.Get(key)
	if got == nil {
		return false
	}

	k.decreaseJailQueueInfo(ctx, val)
	store.Delete(key)
	return true
}

func (k Keeper) jailedQueueIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.JailedQueueKey)
}

func (k Keeper) increaseJailQueueInfo(ctx sdk.Context, val types.Validator) types.JailedQueueInfo {
	jailInfo := k.getJailQueueInfo(ctx)
	jailInfo.Index++
	isKeyNode, _ := k.IsActiveKeyNode(ctx, sdk.CUAddress(val.OperatorAddress))
	if isKeyNode {
		jailInfo.KeyNodeCount++
	}
	k.setJailQueueInfo(ctx, jailInfo)
	return jailInfo
}

func (k Keeper) decreaseJailQueueInfo(ctx sdk.Context, val types.Validator) types.JailedQueueInfo {
	jailInfo := k.getJailQueueInfo(ctx)
	isKeyNode, _ := k.IsActiveKeyNode(ctx, sdk.CUAddress(val.OperatorAddress))
	if isKeyNode {
		jailInfo.KeyNodeCount--
	}
	k.setJailQueueInfo(ctx, jailInfo)
	return jailInfo
}

func (k Keeper) getJailQueueInfo(ctx sdk.Context) types.JailedQueueInfo {
	store := ctx.KVStore(k.storeKey)
	byteKey := types.JailedQueueInfoKey
	bytes := store.Get(byteKey)

	item := types.JailedQueueInfo{}
	if bytes != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bytes, &item)
	}
	return item
}

func (k Keeper) setJailQueueInfo(ctx sdk.Context, value types.JailedQueueInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(value)
	store.Set(types.JailedQueueInfoKey, bz)
}

func (k Keeper) jailQueuedValidatorNow(ctx sdk.Context, maxKeyNodes uint16) {
	iterator := k.jailedQueueIterator(ctx)
	defer iterator.Close()

	jailedVals := make([]types.Validator, 0) // buffer validator to be removed from jailedQueue
	var jailedKeyNodes int
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())
		isKeyNode, _ := k.IsActiveKeyNode(ctx, sdk.CUAddress(valAddr))
		if isKeyNode {
			if jailedKeyNodes == int(maxKeyNodes) {
				continue
			} else {
				jailedKeyNodes++
			}
		}

		validator := k.mustGetValidator(ctx, valAddr)
		k.DeleteValidatorByPowerIndex(ctx, validator)
		jailedVals = append(jailedVals, validator)
	}
	for _, val := range jailedVals {
		k.deleteFromJailedQueue(ctx, val)
		val.JailedIndex = 0
		k.SetValidator(ctx, val)
	}
}
