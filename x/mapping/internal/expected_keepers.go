package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type TokenKeeper interface {
	GetToken(ctx sdk.Context, symbol sdk.Symbol) sdk.Token
}

type CUKeeper interface {
	GetOrNewCU(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CustodianUnit

	GetCU(context sdk.Context, addresses sdk.CUAddress) exported.CustodianUnit

	SetCU(ctx sdk.Context, cu exported.CustodianUnit)

	NewOpCUWithAddress(ctx sdk.Context, symbol string, addr sdk.CUAddress) exported.CustodianUnit

	GetOpCUs(ctx sdk.Context, symbol string) []exported.CustodianUnit

	SetExtAddressWithCU(ctx sdk.Context, symbol, extAddress string, cuAddress sdk.CUAddress)

	GetCUFromExtAddress(ctx sdk.Context, symbol, extAddress string) (sdk.CUAddress, error)
}

type ReceiptKeeper interface {
	// NewReceipt creates a new receipt with a list of flows
	NewReceipt(category sdk.CategoryType, flows []sdk.Flow) *sdk.Receipt

	// NewOrderFlow creates a new order flow
	NewOrderFlow(symbol sdk.Symbol, cuAddress sdk.CUAddress, orderID string, orderType sdk.OrderType,
		orderStatus sdk.OrderStatus) sdk.OrderFlow
	// NewBalanceFlow creates a new balance flow for an asset
	NewBalanceFlow(cuAddress sdk.CUAddress, symbol sdk.Symbol, orderID string, previousBalance,
		balanceChange, previousBalanceOnHold, balanceOnHoldChange sdk.Int) sdk.BalanceFlow

	// SaveReceiptToResult saves the receipt into a result.
	SaveReceiptToResult(receipt *sdk.Receipt, result *sdk.Result) *sdk.Result

	GetReceiptFromResult(result *sdk.Result) (*sdk.Receipt, error)
}

type TransferKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string) sdk.Int
	SendCoin(ctx sdk.Context, from, to sdk.CUAddress, amt sdk.Coin) (sdk.Result, []sdk.Flow, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	AddCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	SubCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	SubCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	SubCoinHold(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	LockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)
	UnlockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)
}
