package genaccounts

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	authexported "github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/genaccounts/internal/types"
)

// InitGenesis initializes accounts and deliver genesis transactions
func InitGenesis(ctx sdk.Context, _ *codec.Codec, cuKeeper types.CUKeeper, genesisState GenesisState) {
	genesisState.Sanitize()

	// load the accounts
	for _, gacc := range genesisState {
		cu := gacc.ToCU()
		cuKeeper.SetCU(ctx, cu)

	}
}

// ExportGenesis exports genesis for all accounts
func ExportGenesis(ctx sdk.Context, cuKeeper types.CUKeeper) GenesisState {

	// iterate to get the cus
	cus := []GenesisCU{}
	cuKeeper.IterateCUs(ctx,
		func(acc authexported.CustodianUnit) (stop bool) {
			CU, err := NewGenesisCUI(acc)
			if err != nil {
				panic(err)
			}
			cus = append(cus, CU)
			return false
		},
	)

	return cus
}
