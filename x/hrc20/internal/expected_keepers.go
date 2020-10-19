package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	supplyexport "github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/token"
)

type TokenKeeper interface {
	SetTokenInfo(ctx sdk.Context, tokenInfo *sdk.TokenInfo)
	GetTokenInfo(ctx sdk.Context, symbol sdk.Symbol) *sdk.TokenInfo
	GetParams(ctx sdk.Context) (params token.Params)
}

type CustodianUnitKeeper interface {
	GetOrNewCU(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CustodianUnit
	SetCU(ctx sdk.Context, cu exported.CustodianUnit)
	GetCU(context sdk.Context, addresses sdk.CUAddress) exported.CustodianUnit
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
