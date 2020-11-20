package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	stakingexported "github.com/hbtc-chain/bhchain/x/staking/exported"
	"github.com/hbtc-chain/bhchain/x/staking/types"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/transfer keeper.

type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) (validators []types.Validator)
	Validator(ctx sdk.Context, address sdk.ValAddress) stakingexported.ValidatorI
	SetEpoch(ctx sdk.Context, epoch sdk.Epoch)
	IsActiveKeyNode(ctx sdk.Context, addr sdk.CUAddress) (bool, int)
}

type TransferKeeper interface {
	SendCoins(ctx sdk.Context, from, to sdk.CUAddress, amt sdk.Coins) (sdk.Result, []sdk.Flow, sdk.Error)
}

// SupplyKeeper defines the expected supply Keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
	GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.CUAddress
}
