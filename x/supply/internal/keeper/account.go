package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/supply/internal/types"
)

// GetModuleAddress returns an address based on the module name
func (k Keeper) GetModuleAddress(moduleName string) sdk.CUAddress {
	permAddr, ok := k.permAddrs[moduleName]
	if !ok {
		return nil
	}
	return permAddr.GetAddress()
}

// GetModuleAddressAndPermissions returns an address and permissions based on the module name
func (k Keeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.CUAddress, permissions []string) {
	permAddr, ok := k.permAddrs[moduleName]
	if !ok {
		return addr, permissions
	}
	return permAddr.GetAddress(), permAddr.GetPermissions()
}

// GetModuleAccountAndPermissions gets the module CustodianUnit from the auth CustodianUnit store and its
// registered permissions
func (k Keeper) GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (exported.ModuleAccountI, []string) {
	addr, perms := k.GetModuleAddressAndPermissions(moduleName)
	if addr == nil {
		return nil, []string{}
	}

	acc := k.ck.GetCU(ctx, addr)
	if acc != nil {
		macc, ok := acc.(exported.ModuleAccountI)
		if !ok {
			panic("CustodianUnit is not a module CustodianUnit")
		}
		return macc, perms
	}

	// create a new module CustodianUnit
	macc := types.NewEmptyModuleAccount(moduleName, perms...)
	maccI := (k.ck.NewCU(ctx, macc)).(exported.ModuleAccountI) // set the CustodianUnit number

	k.SetModuleAccount(ctx, maccI)

	return maccI, perms
}

// GetModuleAccount gets the module CustodianUnit from the auth CustodianUnit store
func (k Keeper) GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI {
	acc, _ := k.GetModuleAccountAndPermissions(ctx, moduleName)
	return acc
}

// SetModuleAccount sets the module CustodianUnit to the auth CustodianUnit store
func (k Keeper) SetModuleAccount(ctx sdk.Context, macc exported.ModuleAccountI) {
	k.ck.SetCU(ctx, macc)
}
