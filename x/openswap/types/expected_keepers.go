package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	supplyI "github.com/hbtc-chain/bhchain/x/supply/exported"
)

type TokenKeeper interface {
	GetTokenInfo(ctx sdk.Context, symbol sdk.Symbol) *sdk.TokenInfo
}

type CUKeeper interface {
	GetCU(ctx sdk.Context, addr sdk.CUAddress) exported.CustodianUnit
	GetOrNewCU(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CustodianUnit
	SetCU(ctx sdk.Context, acc exported.CustodianUnit)
}

type ReceiptKeeper interface {
	NewReceipt(category sdk.CategoryType, flows []sdk.Flow) *sdk.Receipt
	SaveReceiptToResult(receipt *sdk.Receipt, result *sdk.Result) *sdk.Result
	GetReceiptFromResult(result *sdk.Result) (*sdk.Receipt, error)
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
	GetSupply(ctx sdk.Context) (supply supplyI.SupplyI)
}

type TransferKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
}
