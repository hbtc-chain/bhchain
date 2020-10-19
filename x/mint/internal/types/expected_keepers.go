package types // noalias

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	stakingtypes "github.com/hbtc-chain/bhchain/x/staking/types"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
)

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	StakingTokenSupply(ctx sdk.Context) sdk.Int
	BondedRatio(ctx sdk.Context) sdk.Dec
	TotalBondedTokens(ctx sdk.Context) sdk.Int
	GetLastValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.CUAddress

	// TODO remove with genesis 2-phases refactor https://github.com/hbtc-chain/bhchain/issues/2862
	SetModuleAccount(sdk.Context, exported.ModuleAccountI)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
}

type TokenKeeper interface {
	GetDecimals(ctx sdk.Context, symbol sdk.Symbol) uint64
}
