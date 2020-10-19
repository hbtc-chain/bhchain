package gov

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	distrtype "github.com/hbtc-chain/bhchain/x/distribution/types"
	stakingexported "github.com/hbtc-chain/bhchain/x/staking/exported"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
)

// SupplyKeeper defines the supply Keeper for module accounts
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.CUAddress
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/hbtc-chain/bhchain/issues/2862
	SetModuleAccount(sdk.Context, supplyexported.ModuleAccountI)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
	GetSupply(ctx sdk.Context) supplyexported.SupplyI
}

// StakingKeeper expected staking keeper (Validator and Delegator sets)
type StakingKeeper interface {
	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	TotalBondedTokens(sdk.Context) sdk.Int // total bonded tokens within the validator set

	IterateDelegations(ctx sdk.Context, delegator sdk.CUAddress,
		fn func(index int64, delegation stakingexported.DelegationI) (stop bool))
}

type DistributionKeeper interface {
	AddToFeePool(ctx sdk.Context, coins sdk.DecCoins)
	SetFeePool(ctx sdk.Context, feePool distrtype.FeePool)
	GetFeePool(ctx sdk.Context) distrtype.FeePool
}
