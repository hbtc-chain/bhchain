package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/supply/internal/types"
)

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an CUAddress
func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string,
	recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		err := sdk.ErrUnknownAddress(fmt.Sprintf("module CU %s does not exist", senderModule))
		return err.Result(), err
	}

	return k.tk.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another
func (k Keeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error) {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		err := sdk.ErrUnknownAddress(fmt.Sprintf("module CU %s does not exist", senderModule))
		return err.Result(), err
	}

	// create the CU if it doesn't yet exist
	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(fmt.Sprintf("module CU %s isn't able to be created", recipientModule))
	}

	return k.tk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromAccountToModule transfers coins from an CUAddress to a ModuleAccount
func (k Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress,
	recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error) {

	// create the CU if it doesn't yet exist
	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(fmt.Sprintf("module CU %s isn't able to be created", recipientModule))
	}

	return k.tk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// DelegateCoinsFromAccountToModule delegates coins and transfers
// them from a delegator CU to a module CU
func (k Keeper) DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress,
	recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error) {

	// create the CU if it doesn't yet exist
	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(fmt.Sprintf("module CU %s isn't able to be created", recipientModule))
	}

	if !recipientAcc.HasPermission(types.Staking) {
		panic(fmt.Sprintf("module CU %s does not have permissions to receive delegated coins", recipientModule))
	}
	return k.tk.DelegateCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// UndelegateCoinsFromModuleToAccount undelegates the unbonding coins and transfers
// them from a module CU to the delegator CU
func (k Keeper) UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string,
	recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {

	acc := k.GetModuleAccount(ctx, senderModule)
	if acc == nil {
		err := sdk.ErrUnknownAddress(fmt.Sprintf("module CU %s does not exist", senderModule))
		return err.Result(), err
	}

	if !acc.HasPermission(types.Staking) {
		panic(fmt.Sprintf("module CU %s does not have permissions to undelegate coins", senderModule))
	}

	return k.tk.UndelegateCoins(ctx, acc.GetAddress(), recipientAddr, amt)
}

// MintCoins creates new coins from thin air and adds it to the module CU.
// Panics if the name maps to a non-minter module CU or if the amount is invalid.
func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error {

	// create the CU if it doesn't yet exist
	acc := k.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module CU %s does not exist", moduleName))
	}

	if !acc.HasPermission(types.Minter) {
		panic(fmt.Sprintf("module CU %s does not have permissions to mint tokens", moduleName))
	}

	_, _, err := k.tk.AddCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		panic(err)
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply = supply.Inflate(amt)

	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("minted %s from %s module CU", amt.String(), moduleName))

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module CU.
// Panics if the name maps to a non-burner module CU or if the amount is invalid.
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error {

	// create the CU if it doesn't yet exist
	acc := k.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module CU %s does not exist", moduleName))
	}

	if !acc.HasPermission(types.Burner) {
		panic(fmt.Sprintf("module CU %s does not have permissions to burn tokens", moduleName))
	}

	_, _, err := k.tk.SubtractCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		panic(err)
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply = supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("burned %s from %s module CU", amt.String(), moduleName))

	return nil
}
