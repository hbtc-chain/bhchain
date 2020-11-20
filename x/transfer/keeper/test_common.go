package keeper

import (
	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset/exported"
)

func (keeper BaseKeeper) CheckCollectOrders(ctx sdk.Context, orderIDs []string, orderStatus sdk.OrderStatus, depositStatus sdk.DepositItemStatus) (sdk.Coins, *sdk.IBCToken, []*sdk.UtxoIn, []*sdk.OrderCollect, []*sdk.DepositItem, sdk.Error) {
	return keeper.checkCollectOrders(ctx, orderIDs, orderStatus, depositStatus)
}

func (keeper BaseKeeper) CheckWithdrawalOrders(ctx sdk.Context, orderIDs []string, orderStatus sdk.OrderStatus) (*sdk.IBCToken, []*sdk.OrderWithdrawal, sdk.Error) {
	return keeper.checkWithdrawalOrders(ctx, orderIDs, orderStatus)
}

func (keeper BaseKeeper) CheckDecodedUtxoTransaction(ctx sdk.Context, chain, symbol string, opCUAddr sdk.CUAddress, orderIDs []*sdk.OrderWithdrawal, tx *chainnode.ExtUtxoTransaction, fromAddr string) (sdk.Int, sdk.Error) {
	return keeper.checkDecodedUtxoTransaction(ctx, symbol, opCUAddr, orderIDs, tx, fromAddr)
}

func (keeper BaseKeeper) CheckWithdrawalOpCU(ctx sdk.Context, opCU exported.CUIBCAsset, chain, symbol string, sendable bool, fromAddr string) sdk.Error {
	return keeper.checkWithdrawalOpCU(opCU, chain, symbol, sendable, fromAddr)
}

func (keeper BaseKeeper) CheckSysTransferOrders(ctx sdk.Context, order *sdk.OrderSysTransfer, orderStatus sdk.OrderStatus) (*sdk.IBCToken, sdk.Error) {
	return keeper.checkSysTransferOrder(ctx, order, orderStatus)
}
