package gov

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/gov/types"
)

// RegisterInvariants registers all governance invariants
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-cu", ModuleAccountInvariant(keeper))
}

// AllInvariants runs all invariants of the governance module
func AllInvariants(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return ModuleAccountInvariant(keeper)(ctx)
	}
}

// ModuleAccountInvariant checks that the module CU coins reflects the sum of
// deposit amounts held on store
func ModuleAccountInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var expectedDeposits sdk.Coins

		keeper.IterateAllDeposits(ctx, func(deposit types.Deposit) bool {
			expectedDeposits = expectedDeposits.Add(deposit.Amount)
			return false
		})

		macc := keeper.GetGovernanceAccount(ctx)
		have := keeper.tk.GetAllBalance(ctx, macc.GetAddress())
		broken := !have.IsEqual(expectedDeposits)

		return sdk.FormatInvariant(types.ModuleName, "deposits",
			fmt.Sprintf("\tgov ModuleAccount coins: %s\n\tsum of deposit amounts:  %s\n",
				have, expectedDeposits)), broken
	}
}
