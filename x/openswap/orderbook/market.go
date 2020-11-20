package orderbook

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

type Market struct {
	dexID       uint32
	baseSymbol  sdk.Symbol
	quoteSymbol sdk.Symbol
	buyOrders   *Orderbook
	sellOrders  *Orderbook
}

func NewMarket(dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) *Market {
	e := &Market{
		dexID:       dexID,
		baseSymbol:  baseSymbol,
		quoteSymbol: quoteSymbol,
		buyOrders:   NewOrderbook(),
		sellOrders:  NewOrderbook(),
	}
	return e
}

func (e *Market) DexID() uint32 {
	return e.dexID
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

func (e *Market) GetExpiredOrders(ctx sdk.Context) []*types.Order {
	expiredSellOrders := e.sellOrders.GetExpiredOrder(ctx.BlockTime().Unix())
	expiredBuyOrders := e.buyOrders.GetExpiredOrder(ctx.BlockTime().Unix())
	return append(expiredSellOrders, expiredBuyOrders...)
}
