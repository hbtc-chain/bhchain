package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/distribution/types"
)

// DistributeFromFeePool distributes funds from the distribution module CU to
// a receiver address while updating the community pool
func (k Keeper) DistributeFromFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.CUAddress) sdk.Error {
	feePool := k.GetFeePool(ctx)

	// NOTE the community pool isn't a module CU, however its coins
	// are held in the distribution module CU. Thus the community pool
	// must be reduced separately from the SendCoinsFromModuleToAccount call
	newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoins(amount))
	if negative {
		return types.ErrBadDistribution(k.codespace)
	}
	feePool.CommunityPool = newPool

	_, err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}

	k.SetFeePool(ctx, feePool)
	return nil
}
