package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	distrtypes "github.com/hbtc-chain/bhchain/x/distribution/types"
	"github.com/hbtc-chain/bhchain/x/params"
	paramstypes "github.com/hbtc-chain/bhchain/x/params/types"
)

var (
	minBaseProposerReward = sdk.NewDecWithPrec(5, 2)
	maxBaseProposerReward = sdk.NewDecWithPrec(15, 2)
	minKeyNodeReward      = sdk.NewDecWithPrec(5, 2)
	maxKeyNodeReward      = sdk.NewDecWithPrec(15, 2)
)

type ParameterChangeValidator struct {
}

func init() {
	params.RegisterParameterChangeValidator(&ParameterChangeValidator{})
}

func (p *ParameterChangeValidator) Validate(change paramstypes.ParamChange) error {
	if change.Subspace != DefaultParamspace {
		return nil
	}
	bz := []byte(change.Value)
	switch change.Key {
	case string(ParamStoreKeyBaseProposerReward):
		var value sdk.Dec
		err := distrtypes.ModuleCdc.UnmarshalJSON(bz, &value)
		if err != nil {
			return err
		}
		if value.GT(maxBaseProposerReward) || value.LT(minBaseProposerReward) {
			return fmt.Errorf("BaseProposerReward out of range")
		}

	case string(ParamStoreKeyKeyNodeReward):
		var value sdk.Dec
		err := distrtypes.ModuleCdc.UnmarshalJSON(bz, &value)
		if err != nil {
			return err
		}
		if value.GT(maxKeyNodeReward) || value.LT(minKeyNodeReward) {
			return fmt.Errorf("KeyNodeReward out of range")
		}
	default:
		return nil
	}
	return nil
}

// type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyCommunityTax, sdk.Dec{},
		ParamStoreKeyBaseProposerReward, sdk.Dec{},
		ParamStoreKeyBonusProposerReward, sdk.Dec{},
		ParamStoreKeyWithdrawAddrEnabled, false,
		ParamStoreKeyKeyNodeReward, sdk.Dec{},
	)
}

// returns the current CommunityTax rate from the global param store
// nolint: errcheck
func (k Keeper) GetCommunityTax(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyCommunityTax, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetCommunityTax(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyCommunityTax, &percent)
}

// returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetBaseProposerReward(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyBaseProposerReward, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetBaseProposerReward(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyBaseProposerReward, &percent)
}

// returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetBonusProposerReward(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyBonusProposerReward, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetBonusProposerReward(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyBonusProposerReward, &percent)
}

// returns the current WithdrawAddrEnabled
// nolint: errcheck
func (k Keeper) GetWithdrawAddrEnabled(ctx sdk.Context) bool {
	var enabled bool
	k.paramSpace.Get(ctx, ParamStoreKeyWithdrawAddrEnabled, &enabled)
	return enabled
}

// nolint: errcheck
func (k Keeper) SetWithdrawAddrEnabled(ctx sdk.Context, enabled bool) {
	k.paramSpace.Set(ctx, ParamStoreKeyWithdrawAddrEnabled, &enabled)
}

// nolint: errcheck
func (k Keeper) GetKeyNodeReward(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyKeyNodeReward, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetKeyNodeReward(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyKeyNodeReward, &percent)
}
