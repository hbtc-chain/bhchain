package types

type DepthBook struct {
	Height int64      `json:"h"`
	Time   int64      `json:"t"`
	Market string     `json:"m"`
	Bids   [][]string `json:"bids"`
	Asks   [][]string `json:"asks"`
}

func NewDepthBook(height, time int64, market string, buyOrders, sellOrders []*Order) *DepthBook {
	return &DepthBook{
		Height: height,
		Time:   time,
		Market: market,
		Bids:   mergeOrders(buyOrders),
		Asks:   mergeOrders(sellOrders),
	}
}

func mergeOrders(orders []*Order) [][]string {
	if len(orders) == 0 {
		return nil
	}
	var ret [][]string
	curOrder := orders[0]
	accQuantity := curOrder.RemainQuantity()
	for i := 1; i < len(orders); i++ {
		if curOrder.Price.Equal(orders[i].Price) {
			accQuantity = accQuantity.Add(orders[i].RemainQuantity())
		} else {
			ret = append(ret, []string{curOrder.Price.String(), accQuantity.String()})
			curOrder = orders[i]
			accQuantity = curOrder.RemainQuantity()
		}
	}
	ret = append(ret, []string{curOrder.Price.String(), accQuantity.String()})
	return ret
}
