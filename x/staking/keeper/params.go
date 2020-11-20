package keeper

import (
	"time"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

// Default parameter namespace
const (
	DefaultParamspace = types.ModuleName
)

// ParamTable for staking module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyUnbondingTime, &res)
	return
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// MaxEntries - Maximum number of simultaneous unbonding
// delegations or redelegations (per pair/trio)
func (k Keeper) MaxEntries(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxEntries, &res)
	return
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyBondDenom, &res)
	return
}

func (k Keeper) MaxKeyNodes(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxKeyNodes, &res)
	return
}

func (k Keeper) MinValidatorDelegation(ctx sdk.Context) (res sdk.Int) {
	k.paramstore.Get(ctx, types.KeyMinValidatorDelegation, &res)
	return
}

func (k Keeper) MinKeyNodeDelegation(ctx sdk.Context) (res sdk.Int) {
	k.paramstore.Get(ctx, types.KeyMinKeyNodeDelegation, &res)
	return
}

func (k Keeper) MaxCandidateKeyNodeHeartbeatInterval(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxCandidateKeyNodeHeartbeatInterval, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.UnbondingTime(ctx),
		k.MaxValidators(ctx),
		k.MaxKeyNodes(ctx),
		k.MaxEntries(ctx),
		k.BondDenom(ctx),
		k.MinValidatorDelegation(ctx),
		k.MinKeyNodeDelegation(ctx),
		k.MaxCandidateKeyNodeHeartbeatInterval(ctx),
	)
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
