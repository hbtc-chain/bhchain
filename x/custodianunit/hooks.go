package custodianunit

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

func (ck CUKeeper) afterNewEpoch(ctx sdk.Context, epoch sdk.Epoch) {
	ck.startMigrationForAllOpcus(ctx, epoch)
}

func (ck CUKeeper) Hooks() hooks {
	return hooks{ck}
}

type hooks struct {
	k CUKeeper
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
