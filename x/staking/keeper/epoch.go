package keeper

import (
	"encoding/binary"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

func (k Keeper) GetEpoch(ctx sdk.Context, index uint64) (res sdk.Epoch, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetEpochKey(index))
	if b == nil {
		return res, false
	}
	err := k.cdc.UnmarshalBinaryLengthPrefixed(b, &res)
	if err != nil {
		panic(err)
	}
	return res, true
}

func (k Keeper) GetEpochByHeight(ctx sdk.Context, height uint64) sdk.Epoch {
	store := ctx.KVStore(k.storeKey)
	iterator := store.ReverseIterator(types.EpochByHeightKey, types.GetEpochByHeightKey(height+1))
	defer iterator.Close()
	if iterator.Valid() {
		index := binary.BigEndian.Uint64(iterator.Value())
		res, _ := k.GetEpoch(ctx, index)
		return res
	} else {
		return sdk.Epoch{MigrationFinished: true, Index: 1}
	}
}

func (k Keeper) GetCurrentEpoch(ctx sdk.Context) sdk.Epoch {
	return k.GetEpochByHeight(ctx, uint64(ctx.BlockHeight()))
}

func (k Keeper) StartNewEpoch(ctx sdk.Context, vals []sdk.CUAddress) sdk.Epoch {
	var lastIndex uint64
	migrationFinished := false
	if ctx.BlockHeight() == 0 {
		lastIndex = 0
		migrationFinished = true
	} else {
		currentEpoch := k.GetCurrentEpoch(ctx)
		currentEpoch.EndBlockNum = uint64(ctx.BlockHeight())
		k.SetEpoch(ctx, currentEpoch)
		lastIndex = currentEpoch.Index
	}

	newEpoch := sdk.NewEpoch(lastIndex+1, uint64(ctx.BlockHeight()+1), 0, vals, migrationFinished)
	k.SetEpoch(ctx, newEpoch)
	k.SetEpochByHeight(ctx, newEpoch)
	return newEpoch
}

func (k Keeper) SetEpoch(ctx sdk.Context, epoch sdk.Epoch) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(epoch)
	store.Set(types.GetEpochKey(epoch.Index), b)
}

func (k Keeper) SetEpochByHeight(ctx sdk.Context, epoch sdk.Epoch) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetEpochByHeightKey(epoch.StartBlockNum), sdk.Uint64ToBigEndian(epoch.Index))
}

func (k Keeper) IsMigrationFinished(ctx sdk.Context) bool {
	return k.GetCurrentEpoch(ctx).MigrationFinished
}

func (k Keeper) SetMigrationFinished(ctx sdk.Context) {
	epoch := k.GetCurrentEpoch(ctx)
	epoch.MigrationFinished = true
	k.SetEpoch(ctx, epoch)
}
