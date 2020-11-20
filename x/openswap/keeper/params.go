package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = types.ModuleName
)

// ParamTable for staking module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

func (k Keeper) MinimumLiquidity(ctx sdk.Context) (res sdk.Int) {
	k.paramstore.Get(ctx, types.KeyMinimumLiquidity, &res)
	return
}

func (k Keeper) LimitSwapMatchingGas(ctx sdk.Context) (res sdk.Uint) {
	k.paramstore.Get(ctx, types.KeyLimitSwapMatchingGas, &res)
	return
}

func (k Keeper) MaxFeeRate(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyMaxFeeRate, &res)
	return
}

func (k Keeper) LpRewardRate(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyLpRewardRate, &res)
	return
}

func (k Keeper) RepurchaseRate(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyRepurchaseRate, &res)
	return
}

func (k Keeper) RefererTransactionBonusRate(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyRefererTransactionBonusRate, &res)
	return
}

func (k Keeper) RefererMiningBonusRate(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyRefererMiningBonusRate, &res)
	return
}

func (k Keeper) RepurchaseDuration(ctx sdk.Context) (res int64) {
	k.paramstore.Get(ctx, types.KeyRepurchaseDuration, &res)
	return
}

func (k Keeper) MiningWeights(ctx sdk.Context) (res []*types.MiningWeight) {
	k.paramstore.Get(ctx, types.KeyMiningWeights, &res)
	return
}

func (k Keeper) MiningPlans(ctx sdk.Context) (res []*types.MiningPlan) {
	k.paramstore.Get(ctx, types.KeyMiningPlans, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MinimumLiquidity(ctx),
		k.LimitSwapMatchingGas(ctx),
		k.MaxFeeRate(ctx),
		k.LpRewardRate(ctx),
		k.RepurchaseRate(ctx),
		k.RefererTransactionBonusRate(ctx),
		k.RefererMiningBonusRate(ctx),
		k.RepurchaseDuration(ctx),
		k.MiningWeights(ctx),
		k.MiningPlans(ctx),
	)
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
