package openswap

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/keeper"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

type GenesisState struct {
	Params types.Params `json:"params"`
}

func NewGenesisState(params types.Params) GenesisState {
	return GenesisState{
		Params: params,
	}
}

func ValidateGenesis(data GenesisState) error {
	return data.Params.Validate()
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: types.DefaultParams(),
	}
}

func InitGenesis(ctx sdk.Context, k keeper.Keeper, data GenesisState) []abci.ValidatorUpdate {
	k.SetParams(ctx, data.Params)
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) GenesisState {
	params := k.GetParams(ctx)
	return GenesisState{Params: params}
}
