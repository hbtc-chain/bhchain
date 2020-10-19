package keygen

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

func (k Keeper) afterNewEpoch(ctx sdk.Context, epoch sdk.Epoch) {
	k.delAllWaitAssignKeyGenOrderIDs(ctx)
	k.resetKeyGenOrders(ctx, epoch)
}

func (k Keeper) Hooks() hooks {
	return hooks{k}
}

type hooks struct {
	k Keeper
}

// nolint - unused hooks
func (h hooks) AfterValidatorBonded(sdk.Context, sdk.ConsAddress, sdk.ValAddress)               {}
func (h hooks) AfterValidatorRemoved(sdk.Context, sdk.ConsAddress, sdk.ValAddress)              {}
func (h hooks) AfterValidatorCreated(sdk.Context, sdk.ValAddress)                               {}
func (h hooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) {}
func (h hooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress)                         {}
func (h hooks) BeforeDelegationCreated(_ sdk.Context, _ sdk.CUAddress, _ sdk.ValAddress)        {}
func (h hooks) BeforeDelegationSharesModified(_ sdk.Context, _ sdk.CUAddress, _ sdk.ValAddress) {}
func (h hooks) BeforeDelegationRemoved(_ sdk.Context, _ sdk.CUAddress, _ sdk.ValAddress)        {}
func (h hooks) AfterDelegationModified(_ sdk.Context, _ sdk.CUAddress, _ sdk.ValAddress)        {}
func (h hooks) BeforeValidatorSlashed(_ sdk.Context, _ sdk.ValAddress, _ sdk.Dec)               {}
func (h hooks) AfterNewEpoch(ctx sdk.Context, epoch sdk.Epoch) {
	h.k.afterNewEpoch(ctx, epoch)
}
