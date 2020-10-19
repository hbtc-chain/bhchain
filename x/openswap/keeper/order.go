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

func (k Keeper) CreateOrder(ctx sdk.Context, order *types.Order) {
	k.saveOrder(ctx, order)
	k.addUnfinishedOrder(ctx, order)
	k.marketManager.AddOrder(order)
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

func (k Keeper) GetAllOrders(baseSymbol, quoteSymbol sdk.Symbol) ([]*types.Order, []*types.Order) {
	return k.marketManager.GetAllOrders(baseSymbol, quoteSymbol)
}

func (k Keeper) MatchingOrders(ctx sdk.Context) {
	var (
		swapEvents  types.EventSwaps
		totalBurned = sdk.ZeroInt()
	)
	for _, market := range k.marketManager.GetMarkets() {
		pair := k.GetTradingPair(ctx, market.BaseSymbol(), market.QuoteSymbol())
		if pair == nil {
			continue
		}

		for {
			highestBuyOrder, lowestSellOrder := market.GetHighestBuyOrder(), market.GetLowestSellOrder()
			canMatchingSellOrder := lowestSellOrder != nil && calLimitSwapAmount(lowestSellOrder, pair).IsPositive()
			canMatchingBuyOrder := highestBuyOrder != nil && calLimitSwapAmount(highestBuyOrder, pair).IsPositive()

			if !canMatchingSellOrder && !canMatchingBuyOrder {
				break
			}

			var (
				matchedAmount  sdk.Int
				event          *types.EventSwap
				finishedOrders []*types.Order
				burned         sdk.Int
			)

			// matching sell order first
			if canMatchingSellOrder {
				sellOrderIter := market.SellOrderBook().Iterator()
				for sellOrderIter.Next() {
					order := sellOrderIter.Value()

					matchedAmount, burned, _, event, pair = k.limitSwap(ctx, order, pair)
					if matchedAmount.IsZero() {
						break
					}
					if order.LockedFund.IsZero() {
						finishedOrders = append(finishedOrders, order)
					}

					swapEvents = append(swapEvents, event)
					totalBurned = totalBurned.Add(burned)
				}
			}

			if canMatchingBuyOrder {
				buyOrderIter := market.BuyOrderBook().ReverseIterator()
				for buyOrderIter.Next() {
					order := buyOrderIter.Value()

					matchedAmount, burned, _, event, pair = k.limitSwap(ctx, order, pair)
					if matchedAmount.IsZero() {
						break
					}
					if order.LockedFund.IsZero() {
						finishedOrders = append(finishedOrders, order)
					}

					swapEvents = append(swapEvents, event)
					totalBurned = totalBurned.Add(burned)
				}
			}

			for _, order := range finishedOrders {
				k.finishOrderWithStatus(ctx, order, types.OrderStatusFilled)
			}
			finishedOrders = finishedOrders[:0]
		}

	}
	if len(swapEvents) > 0 {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSwap,
			sdk.NewAttribute(types.AttributeKeySwapResult, swapEvents.String()),
			sdk.NewAttribute(types.AttributeKeyBurned, totalBurned.String()),
		))
	}
}

func (k Keeper) ClearExpiredOrders(ctx sdk.Context) {
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
		cu := k.cuKeeper.GetCU(ctx, order.From)
		coins := k.getOrderLockedCoins(ctx, order)
		cu.SubCoinsHold(coins)
		cu.AddCoins(coins)
		k.cuKeeper.SetCU(ctx, cu)

		order.LockedFund = sdk.ZeroInt()

		for _, flow := range cu.GetBalanceFlows() {
			flows = append(flows, flow)
		}

	}

	k.saveOrder(ctx, order)
	k.delUnfinishedOrder(ctx, order)
	k.marketManager.DelOrder(order)
	return flows
}

func (k Keeper) getOrderLockedCoins(ctx sdk.Context, order *types.Order) sdk.Coins {
	symbol := order.BaseSymbol
	if order.Side == types.OrderSideBuy {
		symbol = order.QuoteSymbol
	}
	coin := sdk.NewCoin(symbol.String(), order.LockedFund)
	return sdk.NewCoins(coin)
}

func (k Keeper) saveOrder(ctx sdk.Context, order *types.Order) {
	bz := k.cdc.MustMarshalBinaryBare(order)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OrderKey(order.OrderID), bz)
}

func (k Keeper) addUnfinishedOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.UnfinishedOrderKey(order), []byte{})
}

func (k Keeper) delUnfinishedOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.UnfinishedOrderKey(order))
}
