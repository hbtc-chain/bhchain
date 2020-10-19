package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
)

func (k *Keeper) BehaviourWindow(ctx sdk.Context, behaviourName string) (res int64) {
	k.paramSubSpace.GetWithSubkey(ctx, types.KeyBehaviourWindow, []byte(behaviourName), &res)
	return
}

func (k *Keeper) MaxMisbehaviourCount(ctx sdk.Context, behaviourName string) (res int64) {
	k.paramSubSpace.GetWithSubkey(ctx, types.KeyMaxMisbehaviourCount, []byte(behaviourName), &res)
	return
}

func (k *Keeper) BehaviourSlashFraction(ctx sdk.Context, behaviourName string) (res sdk.Dec) {
	k.paramSubSpace.GetWithSubkey(ctx, types.KeyBehaviourSlashFraction, []byte(behaviourName), &res)
	return
}

func (k Keeper) GetBehaviourParams(ctx sdk.Context, behaviourName string) (params types.BehaviourParams) {
	k.paramSubSpace.GetParamSetWithSubkey(ctx, []byte(behaviourName), &params)
	return params
}

func (k Keeper) SetBehaviourParams(ctx sdk.Context, behaviourName string, params types.BehaviourParams) {
	k.paramSubSpace.SetParamSetWithSubkey(ctx, []byte(behaviourName), &params)
}
