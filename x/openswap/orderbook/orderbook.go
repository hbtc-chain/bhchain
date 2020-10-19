package orderbook

import (
	"strings"

	"github.com/hbtc-chain/bhchain/x/openswap/types"
	"github.com/emirpasic/gods/trees/redblacktree"
)

func compareOrderByPrice(l, r interface{}) int {
	orderA, orderB := l.(*types.Order), r.(*types.Order)
	if orderA.OrderID == orderB.OrderID {
		return 0
	}
	if orderA.LessThan(orderB) {
		return -1
	}
	return 1
}

func compareOrderByExpiredTime(l, r interface{}) int {
	orderA, orderB := l.(*types.Order), r.(*types.Order)
	if orderA.ExpiredTime == orderB.ExpiredTime {
		return strings.Compare(orderA.OrderID, orderB.OrderID)
	}
	return int(orderA.ExpiredTime - orderB.ExpiredTime)
}

type Orderbook struct {
	ordersByPrice       *redblacktree.Tree
	ordersByExpiredTime *redblacktree.Tree
	idToOrder           map[string]*types.Order
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		ordersByPrice:       redblacktree.NewWith(compareOrderByPrice),
		ordersByExpiredTime: redblacktree.NewWith(compareOrderByExpiredTime),
		idToOrder:           make(map[string]*types.Order),
	}
}

func (o *Orderbook) AddOrder(order *types.Order) {
	o.ordersByPrice.Put(order, nil)
	if order.ExpiredTime >= 0 {
		o.ordersByExpiredTime.Put(order, nil)
	}
	o.idToOrder[order.OrderID] = order
}

func (o *Orderbook) DelOrder(orderID string) {
	if order, exist := o.idToOrder[orderID]; exist {
		o.ordersByPrice.Remove(order)
		if order.ExpiredTime >= 0 {
			o.ordersByExpiredTime.Remove(order)
		}
		delete(o.idToOrder, orderID)
	}
}

func (o *Orderbook) GetExpiredOrder(currentTime int64) []*types.Order {
	var expired []*types.Order
	iter := o.ordersByExpiredTime.Iterator()
	for iter.Next() {
		order := iter.Key().(*types.Order)
		if order.ExpiredTime > currentTime {
			break
		}
		expired = append(expired, order)
	}
	return expired
}

func (o *Orderbook) Iterator() *OrderIterator {
	return NewOrderIterator(o.ordersByPrice)
}

func (o *Orderbook) ReverseIterator() *OrderReverseIterator {
	return NewOrderReverseIterator(o.ordersByPrice)
}
