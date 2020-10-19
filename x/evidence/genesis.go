package evidence

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for key, behaviourParams := range data.BehaviourParams {
		k.SetBehaviourParams(ctx, key, behaviourParams)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	params := map[string]types.BehaviourParams{}
	for _, key := range types.AllBehaviourKeys {
		params[key] = k.GetBehaviourParams(ctx, key)
	}
	return NewGenesisState(params)
}
