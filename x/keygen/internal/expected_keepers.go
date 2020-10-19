package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/staking/types"
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

type OrderKeeper interface {
	IsExist(ctx sdk.Context, uuid string) bool

	NewOrder(ctx sdk.Context, order sdk.Order) sdk.Order

	SetOrder(ctx sdk.Context, order sdk.Order)

	GetOrder(ctx sdk.Context, orderID string) sdk.Order

	NewOrderKeyGen(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
		keynodes []sdk.CUAddress, signThreshold uint64, to sdk.CUAddress, openFee sdk.Coins) *sdk.OrderKeyGen

	RemoveProcessOrder(ctx sdk.Context, orderType sdk.OrderType, orderID string)

	GetProcessOrderListByType(ctx sdk.Context, orderTypes ...sdk.OrderType) []string

	GetProcessOrderList(ctx sdk.Context) []string
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

type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) (validators []types.Validator)
	GetEpochByHeight(ctx sdk.Context, height uint64) sdk.Epoch
	GetCurrentEpoch(ctx sdk.Context) sdk.Epoch
}

type DistributionKeeper interface {
	AddToFeePool(ctx sdk.Context, coins sdk.DecCoins)
}
