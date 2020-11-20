package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func (k Keeper) GetOrder(ctx sdk.Context, orderID string) *types.Order {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OrderKey(orderID))
	if len(bz) == 0 {
		return nil
	}
	var order types.Order
	k.cdc.MustUnmarshalBinaryBare(bz, &order)
	return &order
}

func (k Keeper) CancelOrders(ctx sdk.Context, from sdk.CUAddress, orderIDs []string) sdk.Result {
	var flows []sdk.Flow
	for _, orderID := range orderIDs {
		order := k.GetOrder(ctx, orderID)
		if order == nil {
			return sdk.ErrNotFoundOrder(fmt.Sprintf("order %s not found", orderID)).Result()
		}
		if order.IsFinished() {
			return sdk.ErrInvalidTx(fmt.Sprintf("order %s has been finished, cannot be canceled", orderID)).Result()
		}
		if !order.From.Equals(from) {
			return sdk.ErrInvalidTx(fmt.Sprintf("no permission to cancel order %s", orderID)).Result()
		}
		balanceFlows := k.finishOrderWithStatus(ctx, order, types.OrderStatusCanceled)
		flows = append(flows, balanceFlows...)
	}
	k.addWaitToRemoveFromMatchingOrderID(ctx, orderIDs)
	result := sdk.Result{}
	if len(flows) > 0 {
		receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
		k.rk.SaveReceiptToResult(receipt, &result)
	}
	event := types.NewEventOrderStatusChanged(orderIDs)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCancelOrders,
		sdk.NewAttribute(types.AttributeKeyOrders, event.String()),
	))
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) ExpireOrder(ctx sdk.Context, order *types.Order) {
	k.finishOrderWithStatus(ctx, order, types.OrderStatusExpired)
}

func (k Keeper) InitMatchingManager(ctx sdk.Context) {
	k.marketManager.Init(ctx)
}

func (k Keeper) GetAllOrders(dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) ([]*types.Order, []*types.Order) {
	return k.marketManager.GetAllOrders(dexID, baseSymbol, quoteSymbol)
}

func (k Keeper) MatchingOrders(ctx sdk.Context) {
	var (
		swapEvents types.EventSwaps
	)
	for _, market := range k.marketManager.GetMarkets() {
		pair := k.GetTradingPair(ctx, market.DexID(), market.BaseSymbol(), market.QuoteSymbol())
		if pair == nil {
			continue
		}

		for {
			var (
				event            *types.EventSwap
				finishedOrders   []*types.Order
				priceSuitable    bool
				sellOrderMatched bool
				buyOrderMatched  bool
			)

			// matching sell order first
			sellOrderIter := market.SellOrderBook().Iterator()
			for sellOrderIter.Next() {
				order := sellOrderIter.Value()
				_, event, pair, priceSuitable = k.limitSwap(ctx, order, pair)
				if !priceSuitable {
					break
				}
				if event == nil || event.AmountIn.IsZero() {
					continue
				}

				sellOrderMatched = true
				if order.Status == types.OrderStatusFilled {
					finishedOrders = append(finishedOrders, order)
				}

				swapEvents = append(swapEvents, event)
			}

			buyOrderIter := market.BuyOrderBook().ReverseIterator()
			for buyOrderIter.Next() {
				order := buyOrderIter.Value()
				_, event, pair, priceSuitable = k.limitSwap(ctx, order, pair)
				if !priceSuitable {
					break
				}
				if event == nil || event.AmountIn.IsZero() {
					continue
				}

				buyOrderMatched = true
				if order.Status == types.OrderStatusFilled {
					finishedOrders = append(finishedOrders, order)
				}

				swapEvents = append(swapEvents, event)
			}

			if !sellOrderMatched && !buyOrderMatched {
				break
			}

			for _, order := range finishedOrders {
				k.delUnfinishedOrder(ctx, order)
				k.marketManager.DelOrder(order)
			}
			finishedOrders = finishedOrders[:0]
		}

	}
	if len(swapEvents) > 0 {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSwap,
			sdk.NewAttribute(types.AttributeKeySwapResult, swapEvents.String()),
		))
	}
}

func (k Keeper) UpdateOrdersInMatching(ctx sdk.Context) {
	k.clearExpiredOrders(ctx)
	k.insertOrderToMatching(ctx)
	k.removeOrderFromMatching(ctx)
}

func (k Keeper) clearExpiredOrders(ctx sdk.Context) {
	var orderIDs []string
	for _, order := range k.marketManager.GetExpiredOrders(ctx) {
		k.ExpireOrder(ctx, order)
		orderIDs = append(orderIDs, order.OrderID)
	}
	if len(orderIDs) > 0 {
		event := types.NewEventOrderStatusChanged(orderIDs)
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeExpireOrders,
			sdk.NewAttribute(types.AttributeKeyOrders, event.String()),
		))
	}
}

func (k Keeper) insertOrderToMatching(ctx sdk.Context) {
	orderIDs := k.getWaitToInsertMatchingOrderIDs(ctx)
	if len(orderIDs) == 0 {
		return
	}
	for _, orderID := range orderIDs {
		order := k.GetOrder(ctx, orderID)
		if !order.IsFinished() {
			k.marketManager.AddOrder(order)
		}
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.WaitToInsertMatchingKey)
}

func (k Keeper) removeOrderFromMatching(ctx sdk.Context) {
	orderIDs := k.getWaitToRemoveFromMatchingOrderIDs(ctx)
	if len(orderIDs) == 0 {
		return
	}
	for _, orderID := range orderIDs {
		order := k.GetOrder(ctx, orderID)
		k.marketManager.DelOrder(order)
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.WaitToRemoveFromMatchingKey)
}

func (k Keeper) IteratorAllUnfinishedOrder(ctx sdk.Context, f func(*types.Order)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.UnfinishedOrderKeyPrefix)
	for ; iter.Valid(); iter.Next() {
		orderID := types.GetOrderIDFromUnfinishedOrderKey(iter.Key())
		order := k.GetOrder(ctx, orderID)
		f(order)
	}
	iter.Close()
}

func (k Keeper) finishOrderWithStatus(ctx sdk.Context, order *types.Order, status byte) []sdk.Flow {
	order.FinishedTime = ctx.BlockTime().Unix()
	order.Status = status
	flows := make([]sdk.Flow, 0)
	if order.LockedFund.IsPositive() {
		coin := k.getOrderLockedCoin(ctx, order)
		flows, _ = k.tk.UnlockCoin(ctx, order.From, coin)
		order.LockedFund = sdk.ZeroInt()
	}

	k.saveOrder(ctx, order)
	k.delUnfinishedOrder(ctx, order)
	return flows
}

func (k Keeper) getOrderLockedCoin(ctx sdk.Context, order *types.Order) sdk.Coin {
	symbol := order.BaseSymbol
	if order.Side == types.OrderSideBuy {
		symbol = order.QuoteSymbol
	}
	return sdk.NewCoin(symbol.String(), order.LockedFund)
}

func (k Keeper) saveOrder(ctx sdk.Context, order *types.Order) {
	bz := k.cdc.MustMarshalBinaryBare(order)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OrderKey(order.OrderID), bz)
}

func (k Keeper) addUnfinishedOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.UnfinishedOrderKey(order), []byte{})
	k.addWaitToInsertMatchingOrderID(ctx, order.OrderID)
}

func (k Keeper) delUnfinishedOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.UnfinishedOrderKey(order))
}

func (k Keeper) getWaitToInsertMatchingOrderIDs(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.WaitToInsertMatchingKey)
	orderIDs := make([]string, 0)
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryBare(bz, &orderIDs)
	}
	return orderIDs
}

func (k Keeper) addWaitToInsertMatchingOrderID(ctx sdk.Context, orderID string) {
	orderIDs := k.getWaitToInsertMatchingOrderIDs(ctx)
	orderIDs = append(orderIDs, orderID)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(orderIDs)
	store.Set(types.WaitToInsertMatchingKey, bz)
}

func (k Keeper) getWaitToRemoveFromMatchingOrderIDs(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.WaitToRemoveFromMatchingKey)
	orderIDs := make([]string, 0)
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryBare(bz, &orderIDs)
	}
	return orderIDs
}

func (k Keeper) addWaitToRemoveFromMatchingOrderID(ctx sdk.Context, ids []string) {
	orderIDs := k.getWaitToRemoveFromMatchingOrderIDs(ctx)
	orderIDs = append(orderIDs, ids...)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(orderIDs)
	store.Set(types.WaitToRemoveFromMatchingKey, bz)
}
