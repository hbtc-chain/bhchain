package orderbook

import (
	"github.com/hbtc-chain/bhchain/x/openswap/types"
	"github.com/emirpasic/gods/trees/redblacktree"
)

type OrderIterator struct {
	redblacktree.Iterator
}

func NewOrderIterator(t *redblacktree.Tree) *OrderIterator {
	return &OrderIterator{
		Iterator: t.Iterator(),
	}
}

func (it *OrderIterator) Value() *types.Order {
	return it.Iterator.Key().(*types.Order)
}

type OrderReverseIterator struct {
	redblacktree.Iterator
}

func NewOrderReverseIterator(t *redblacktree.Tree) *OrderReverseIterator {
	iter := t.Iterator()
	iter.End()
	return &OrderReverseIterator{
		Iterator: iter,
	}
}

func (it *OrderReverseIterator) Next() bool {
	return it.Iterator.Prev()
}

func (it *OrderReverseIterator) Value() *types.Order {
	return it.Iterator.Key().(*types.Order)
}
