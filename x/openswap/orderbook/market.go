package orderbook

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

type Market struct {
	baseSymbol  sdk.Symbol
	quoteSymbol sdk.Symbol
	buyOrders   *Orderbook
	sellOrders  *Orderbook
}

func NewMarket(baseSymbol, quoteSymbol sdk.Symbol) *Market {
	e := &Market{
		baseSymbol:  baseSymbol,
		quoteSymbol: quoteSymbol,
		buyOrders:   NewOrderbook(),
		sellOrders:  NewOrderbook(),
	}
	return e
}

func (e *Market) BaseSymbol() sdk.Symbol {
	return e.baseSymbol
}

func (e *Market) QuoteSymbol() sdk.Symbol {
	return e.quoteSymbol
}

func (e *Market) SellOrderBook() *Orderbook {
	return e.sellOrders
}

func (e *Market) BuyOrderBook() *Orderbook {
	return e.buyOrders
}

func (e *Market) AddOrder(order *types.Order) {
	if order.Side == types.OrderSideBuy {
		e.buyOrders.AddOrder(order)
	} else {
		e.sellOrders.AddOrder(order)
	}
}

func (e *Market) DelOrder(order *types.Order) {
	if order.Side == types.OrderSideBuy {
		e.buyOrders.DelOrder(order.OrderID)
	} else {
		e.sellOrders.DelOrder(order.OrderID)
	}
}

func (e *Market) GetAllOrders() ([]*types.Order, []*types.Order) {
	var buyOrders, sellOrders []*types.Order
	buyOrderIter := e.buyOrders.ReverseIterator()
	for buyOrderIter.Next() {
		buyOrders = append(buyOrders, buyOrderIter.Value())
	}
	sellOrderIter := e.sellOrders.ReverseIterator()
	for sellOrderIter.Next() {
		sellOrders = append(sellOrders, sellOrderIter.Value())
	}
	return sellOrders, buyOrders
}

func (e *Market) GetHighestBuyOrder() *types.Order {
	buyOrderIter := e.buyOrders.ReverseIterator()
	if buyOrderIter.Next() {
		return buyOrderIter.Value()
	}
	return nil
}

func (e *Market) GetLowestSellOrder() *types.Order {
	sellOrderIter := e.sellOrders.Iterator()
	if sellOrderIter.Next() {
		return sellOrderIter.Value()
	}
	return nil
}

func (e *Market) GetExpiredOrders(ctx sdk.Context) []*types.Order {
	expiredSellOrders := e.sellOrders.GetExpiredOrder(ctx.BlockTime().Unix())
	expiredBuyOrders := e.buyOrders.GetExpiredOrder(ctx.BlockTime().Unix())
	return append(expiredSellOrders, expiredBuyOrders...)
}
