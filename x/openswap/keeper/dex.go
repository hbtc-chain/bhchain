package keeper

import (
	"encoding/binary"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func (k Keeper) SaveDex(ctx sdk.Context, dex *types.Dex) *types.Dex {
	if dex.ID == 0 {
		dex.ID = k.incDexID(ctx)
	}
	bz := k.cdc.MustMarshalBinaryBare(dex)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.DexKey(dex.ID), bz)
	return dex
}

func (k Keeper) GetDex(ctx sdk.Context, id uint32) *types.Dex {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DexKey(id))
	if len(bz) == 0 {
		return nil
	}
	var dex types.Dex
	k.cdc.MustUnmarshalBinaryBare(bz, &dex)
	return &dex
}

func (k Keeper) GetAllDex(ctx sdk.Context) []*types.Dex {
	var ret []*types.Dex
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DexKeyPrefix)
	for ; iter.Valid(); iter.Next() {
		var dex types.Dex
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &dex)
		ret = append(ret, &dex)
	}
	return ret
}

func (k Keeper) incDexID(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DexIDKey)
	var id uint32
	if len(bz) != 0 {
		id = binary.BigEndian.Uint32(bz)
	}
	id++
	k.SetDexID(ctx, id)
	return id
}

func (k Keeper) SetDexID(ctx sdk.Context, dexID uint32) {
	store := ctx.KVStore(k.storeKey)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, dexID)
	store.Set(types.DexIDKey, buf)
}

func (k Keeper) SaveTradingPair(ctx sdk.Context, pair *types.TradingPair) {
	bz := k.cdc.MustMarshalBinaryBare(pair)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.TradingPairKey(pair.DexID, pair.TokenA, pair.TokenB), bz)
}

func (k Keeper) GetTradingPair(ctx sdk.Context, dexID uint32, tokenA, tokenB sdk.Symbol) *types.TradingPair {
	tokenA, tokenB = k.SortToken(tokenA, tokenB)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TradingPairKey(dexID, tokenA, tokenB))
	if len(bz) == 0 {
		return nil
	}
	var pair types.TradingPair
	k.cdc.MustUnmarshalBinaryBare(bz, &pair)
	return &pair
}

func (k Keeper) GetAllTradingPairs(ctx sdk.Context, dexID *uint32) []*types.TradingPair {
	prefix := types.TradingPairKeyPrefix
	if dexID != nil {
		prefix = types.TradingPairKeyPrefixWithDexID(*dexID)
	}
	var ret []*types.TradingPair
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	for ; iter.Valid(); iter.Next() {
		var pair types.TradingPair
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &pair)
		ret = append(ret, &pair)
	}
	return ret
}
