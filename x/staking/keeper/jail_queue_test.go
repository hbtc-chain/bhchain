package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

func TestJailQueue(t *testing.T) {

	ctx, _, keeper, _ := CreateTestInput(t, false, 0)

	n := 5
	validators := make([]types.Validator, 0)
	for i := 0; i < n; i++ {
		validator := types.NewValidator(sdk.ValAddress(addrVals[i+2]), PKs[i+2], types.Description{}, true)
		validator = validator.UpdateStatus(sdk.Bonded)
		tokens := sdk.TokensFromConsensusPower(int64(i + 500000))
		validator, _ = validator.AddTokensFromDel(tokens)
		keeper.SetValidatorByPowerIndex(ctx, validator)
		keeper.SetValidator(ctx, validator)

		fmt.Printf("validator record saved for address: %X\n", validator.OperatorAddress)
	}

	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	ctx = ctx.WithBlockHeight(ctx.BlockHeader().Height + 1)

	for i := 0; i < n; i++ {
		validator := keeper.mustGetValidator(ctx, sdk.ValAddress(addrVals[i+2]))
		keeper.jailValidator(ctx, validator)
		require.Equal(t, i >= 0, keeper.ReachJailQueueLimit(ctx), i)
		validators = append(validators, keeper.mustGetValidator(ctx, validator.OperatorAddress))
	}

	keeper.jailQueuedValidatorNow(ctx, 1)

	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, types.ValidatorsByPowerIndexKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())
		//fmt.Printf("current address: %X\n", valAddr)
		require.NotEqual(t, valAddr, validators[0].OperatorAddress)
		// require.NotEqual(t, valAddr, validators[1].OperatorAddress)
	}
	jailedQueueIterator := keeper.jailedQueueIterator(ctx)
	defer jailedQueueIterator.Close()
	for ; jailedQueueIterator.Valid(); jailedQueueIterator.Next() {
		valAddr := sdk.ValAddress(jailedQueueIterator.Value())
		// fmt.Printf("current address: %X\n", valAddr)
		require.NotEqual(t, valAddr, validators[0].OperatorAddress)
		// require.NotEqual(t, valAddr, validators[1].OperatorAddress)
	}

	keeper.unjailValidator(ctx, validators[0])
	require.True(t, found(ctx, keeper, validators[0].OperatorAddress))

	require.True(t, found(ctx, keeper, validators[2].OperatorAddress))
	keeper.unjailValidator(ctx, validators[2])
	require.False(t, foundInJailedQueue(ctx, keeper, validators[2].OperatorAddress))
}

func found(ctx sdk.Context, keeper Keeper, address sdk.ValAddress) bool {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, types.ValidatorsByPowerIndexKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())
		//fmt.Printf("current address: %X\n", valAddr)
		if valAddr.String() == address.String() {
			return true
		}
	}
	return false
}

func foundInJailedQueue(ctx sdk.Context, keeper Keeper, address sdk.ValAddress) bool {
	iterator := keeper.jailedQueueIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())
		//fmt.Printf("current address: %X\n", valAddr)
		if valAddr.String() == address.String() {
			return true
		}
	}
	return false
}
