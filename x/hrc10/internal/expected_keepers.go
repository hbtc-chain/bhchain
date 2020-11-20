package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	supplyexport "github.com/hbtc-chain/bhchain/x/supply/exported"
)

type TokenKeeper interface {
	CreateToken(ctx sdk.Context, tokenInfo sdk.Token) error
	HasToken(ctx sdk.Context, symbol sdk.Symbol) bool
}

type DistributionKeeper interface {
	AddToFeePool(ctx sdk.Context, coins sdk.DecCoins)
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.CUAddress

	// TODO remove with genesis 2-phases refactor https://github.com/hbtc-chain/bhchain/issues/2862
	SetModuleAccount(sdk.Context, supplyexport.ModuleAccountI)

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
}

type ReceiptKeeper interface {
	NewReceipt(category sdk.CategoryType, flows []sdk.Flow) *sdk.Receipt
	SaveReceiptToResult(receipt *sdk.Receipt, result *sdk.Result) *sdk.Result
	GetReceiptFromResult(result *sdk.Result) (*sdk.Receipt, error)
}

type TransferKeeper interface {
	AddCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	SubCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
}
