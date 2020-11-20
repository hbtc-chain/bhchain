package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	supplyI "github.com/hbtc-chain/bhchain/x/supply/exported"
)

type TokenKeeper interface {
	GetToken(ctx sdk.Context, symbol sdk.Symbol) sdk.Token
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
	GetAllBalance(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins
	AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	AddCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	SubCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	SubCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	SubCoinHold(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	LockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)
	UnlockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)
}
