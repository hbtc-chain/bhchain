package internal

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	ibcexported "github.com/hbtc-chain/bhchain/x/ibcasset/exported"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type TokenKeeper interface {
	GetIBCToken(ctx sdk.Context, symbol sdk.Symbol) *sdk.IBCToken
}

type OrderKeeper interface {
	IsExist(ctx sdk.Context, uuid string) bool

	NewOrder(ctx sdk.Context, order sdk.Order) sdk.Order

	SetOrder(ctx sdk.Context, order sdk.Order)

	GetOrder(ctx sdk.Context, orderID string) sdk.Order

	NewOrderKeyGen(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
		keynodes []sdk.CUAddress, signThreshold uint64, to sdk.CUAddress, openFee sdk.Coin) *sdk.OrderKeyGen

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

	SetExtAddressWithCU(ctx sdk.Context, symbol, extAddress string, cuAddress sdk.CUAddress)

	GetCUFromExtAddress(ctx sdk.Context, symbol, extAddress string) (sdk.CUAddress, error)
}

type IBCAssetKeeper interface {
	GetCUIBCAsset(context sdk.Context, addresses sdk.CUAddress) ibcexported.CUIBCAsset
	NewCUIBCAssetWithAddress(ctx sdk.Context, cuType sdk.CUType, cuaddr sdk.CUAddress) ibcexported.CUIBCAsset
	SetCUIBCAsset(ctx sdk.Context, cuAst ibcexported.CUIBCAsset)
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
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool)
	GetEpochByHeight(ctx sdk.Context, height uint64) sdk.Epoch
	GetCurrentEpoch(ctx sdk.Context) sdk.Epoch
}

type DistributionKeeper interface {
	AddToFeePool(ctx sdk.Context, coins sdk.DecCoins)
}

type TransferKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string) sdk.Int
	SubCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	SubCoinHold(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error)
	LockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)
}
