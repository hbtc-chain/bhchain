package orderbook

import (
	"math/rand"
	"strconv"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrderbook(t *testing.T) {
	o := NewOrderbook()
	baseOrders := getAscendingOrders(0)
	randIndex := rand.Perm(len(baseOrders))
	for _, i := range randIndex {
		o.AddOrder(baseOrders[i])
	}

	var iter iterInterface = o.Iterator()
	orders := getOrdersFromIterator(iter)
	assertOrderList(t, baseOrders, orders)

	iter = o.ReverseIterator()
	orders = getOrdersFromIterator(iter)
	assertOrderList(t, baseOrders, reverseOrderList(orders))

	expiredOrders := o.GetExpiredOrder(3)
	assertOrderList(t, baseOrders[:3], expiredOrders)

	o.DelOrder(baseOrders[0].OrderID)
	o.DelOrder(baseOrders[4].OrderID)

	iter = o.Iterator()
	orders = getOrdersFromIterator(iter)
	assertOrderList(t, baseOrders[1:4], orders)

	iter = o.ReverseIterator()
	orders = getOrdersFromIterator(iter)
	assertOrderList(t, baseOrders[1:4], reverseOrderList(orders))
}

func getRandomOrder() *types.Order {
	return &types.Order{
		OrderID:     uuid.NewV4().String(),
		Price:       sdk.NewDec(rand.Int63n(10000) + 1),
		AmountIn:    sdk.NewInt(rand.Int63n(10) + 1),
		BaseSymbol:  "btc",
		QuoteSymbol: "usdt",
		Side:        byte(rand.Intn(2)),
	}
}

func getAscendingOrders(side byte) []*types.Order {
	var orders []*types.Order
	for i := 1; i <= 5; i++ {
		order := &types.Order{
			OrderID:     strconv.Itoa(int(side)) + ":" + strconv.Itoa(i),
			Price:       sdk.NewDec(int64(i)),
			BaseSymbol:  "btc",
			QuoteSymbol: "usdt",
			Side:        side,
			ExpiredTime: int64(i),
		}
		orders = append(orders, order)
	}
	return orders
}

type iterInterface interface {
	Next() bool
	Value() *types.Order
}

func getOrdersFromIterator(iter iterInterface) []*types.Order {
	var orders []*types.Order
	for iter.Next() {
		order := iter.Value()
		orders = append(orders, order)
	}
	return orders
}

func reverseOrderList(orders []*types.Order) []*types.Order {
	for i := 0; i < len(orders)/2; i++ {
		j := len(orders) - 1 - i
		orders[i], orders[j] = orders[j], orders[i]
	}
	return orders
}

func assertOrderList(t *testing.T, expected, got []*types.Order) {
	assert.Equal(t, len(expected), len(got))
	for i := range expected {
		assert.Equal(t, expected[i].OrderID, got[i].OrderID)
		assert.True(t, expected[i].Price.Equal(got[i].Price))
		assert.Equal(t, expected[i].BaseSymbol, got[i].BaseSymbol)
		assert.Equal(t, expected[i].QuoteSymbol, got[i].QuoteSymbol)
		assert.Equal(t, expected[i].ExpiredTime, got[i].ExpiredTime)
	}
}

func BenchmarkAddOrder(b *testing.B) {
	o := NewOrderbook()
	var orders []*types.Order
	for i := 0; i < b.N; i++ {
		orders = append(orders, getRandomOrder())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.AddOrder(orders[i])
	}
}

func BenchmarkDelOrder(b *testing.B) {
	o := NewOrderbook()
	var orders []*types.Order
	for i := 0; i < b.N; i++ {
		order := getRandomOrder()
		orders = append(orders, order)
		o.AddOrder(order)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.DelOrder(orders[i].OrderID)
	}
}

func BenchmarkIterOrder(b *testing.B) {
	o := NewOrderbook()
	for i := 0; i < b.N; i++ {
		o.AddOrder(getRandomOrder())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := o.Iterator()
		for iter.Next() {
			_ = iter.Value()
		}
	}
}
