package hrc20

import (
	"bytes"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strings"
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

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) []abci.ValidatorUpdate {
	k.SetParams(ctx, data.Params)
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	return GenesisState{Params: params}
}

// Checks whether 2 GenesisState structs are equivalent.
func (g GenesisState) Equal(g2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(g)
	b2 := ModuleCdc.MustMarshalBinaryBare(g2)
	return bytes.Equal(b1, b2)
}

// Returns if a GenesisState is empty or has data in it
func (g GenesisState) IsEmpty() bool {
	emptyGenState := GenesisState{}
	return g.Equal(emptyGenState)
}

func (g GenesisState) String() string {
	var b strings.Builder
	b.WriteString(g.Params.String())
	return b.String()
}
