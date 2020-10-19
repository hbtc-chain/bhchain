package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
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

	GetTokenInfo(ctx sdk.Context, symbol sdk.Symbol) *sdk.TokenInfo

	GetAllTokenInfo(ctx sdk.Context) []sdk.TokenInfo
}

type CUKeeper interface {
	GetOrNewCU(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CustodianUnit

	GetCU(context sdk.Context, addresses sdk.CUAddress) exported.CustodianUnit

	SetCU(ctx sdk.Context, cu exported.CustodianUnit)

	NewOpCUWithAddress(ctx sdk.Context, symbol string, addr sdk.CUAddress) exported.CustodianUnit

	GetOpCUs(ctx sdk.Context, symbol string) []exported.CustodianUnit

	SetExtAddresseWithCU(ctx sdk.Context, symbol, extAddress string, cuAddress sdk.CUAddress)

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
