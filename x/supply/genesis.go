package supply

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	autypes "github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/supply/internal/types"
)

// InitGenesis sets supply information for genesis.
//
// CONTRACT: all types of accounts must have been already initialized/created
func InitGenesis(ctx sdk.Context, keeper Keeper, ak types.CUKeeper, tk types.TransferKeeper, data GenesisState) {
	// manually set the total supply based on accounts if not provided
	if data.Supply.Empty() {
		var totalSupply sdk.Coins
		ak.IterateCUs(ctx,
			func(acc autypes.CU) (stop bool) {
				totalSupply = totalSupply.Add(tk.GetAllBalance(ctx, acc.GetAddress()))
				return false
			},
		)

		data.Supply = totalSupply
	}

	keeper.SetSupply(ctx, types.NewSupply(data.Supply))
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(keeper.GetSupply(ctx).GetTotal())
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return types.NewSupply(data.Supply).ValidateBasic()
}
