package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	stakingexported "github.com/hbtc-chain/bhchain/x/staking/exported"
	"github.com/hbtc-chain/bhchain/x/staking/types"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type TokenKeeper interface {
	IsUtxoBased(ctx sdk.Context, symbol sdk.Symbol) bool

	IsSubToken(ctx sdk.Context, symbol sdk.Symbol) bool

	GetOpenFee(ctx sdk.Context, symbol sdk.Symbol) sdk.Int

	GetSysOpenFee(ctx sdk.Context, symbol sdk.Symbol) sdk.Int

	IsTokenSupported(ctx sdk.Context, symbol sdk.Symbol) bool

	GetMaxOpCUNumber(ctx sdk.Context, symbol sdk.Symbol) uint64

	GetChain(ctx sdk.Context, symbol sdk.Symbol) sdk.Symbol
}

type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) (validators []types.Validator)
	Validator(ctx sdk.Context, address sdk.ValAddress) stakingexported.ValidatorI
	SetEpoch(ctx sdk.Context, epoch sdk.Epoch)
	IsActiveKeyNode(ctx sdk.Context, addr sdk.CUAddress) (bool, int)
}

type ReceiptKeeper interface {
}

// SupplyKeeper defines the expected supply Keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
	GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.CUAddress
}
