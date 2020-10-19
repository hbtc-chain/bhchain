package mapping

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/mapping/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	Params types.Params `json:"params"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params types.Params) GenesisState {
	return GenesisState{
		Params: params,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: types.DefaultParams(),
	}
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	keeper.SetParams(ctx, data.Params)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)
	return GenesisState{Params: params}
}

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return data.Params.Validate()
}
