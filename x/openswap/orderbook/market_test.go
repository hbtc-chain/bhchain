package orderbook

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

var (
	defaultTestFeeRate = sdk.NewDecWithPrec(2, 3) // 0.002
)

func TestMarketGetOrderbook(t *testing.T) {
	market := NewMarket("btc", "usdt")

	buyOrders := getAscendingOrders(0)
	sellOrders := getAscendingOrders(1)
	randomIndex := rand.Perm(len(buyOrders))
	for _, i := range randomIndex {
		market.AddOrder(buyOrders[i])
		market.AddOrder(sellOrders[i])
	}

	sell, buy := market.GetAllOrders()
	assertOrderList(t, sellOrders, reverseOrderList(sell))
	assertOrderList(t, buyOrders, reverseOrderList(buy))

	assert.Equal(t, buyOrders[len(buyOrders)-1].Price, market.GetHighestBuyOrder().Price)
	assert.Equal(t, sellOrders[0].Price, market.GetLowestSellOrder().Price)

	ctx := sdk.Context{}.WithBlockTime(time.Unix(3, 0))
	orders := market.GetExpiredOrders(ctx)
	assertOrderList(t, append(sellOrders[:3:3], buyOrders[:3]...), orders)

	market.DelOrder(buyOrders[0])
	market.DelOrder(sellOrders[len(sellOrders)-1])
	sell, buy = market.GetAllOrders()
	assertOrderList(t, buyOrders[1:], reverseOrderList(buy))
	assertOrderList(t, sellOrders[:len(sellOrders)-1], reverseOrderList(sell))
}
