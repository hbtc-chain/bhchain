package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	OrderStatusNew             = 0x00
	OrderStatusPartiallyFilled = 0x01
	OrderStatusFilled          = 0x02
	OrderStatusCanceled        = 0x03
	OrderStatusExpired         = 0x04
)

const (
	OrderSideBuy  = 0x0
	OrderSideSell = 0x1
)

type FeeRate struct {
	LPRewardRate      sdk.Dec `json:"lp_reward_rate"`
	RepurchaseRate    sdk.Dec `json:"repurchase_rate"`
	RefererRewardRate sdk.Dec `json:"referer_reward_rate"`
}

func NewFeeRate(lpRewardRate, repurchaseRate, refererRewardRate sdk.Dec) *FeeRate {
	return &FeeRate{
		LPRewardRate:      lpRewardRate,
		RepurchaseRate:    repurchaseRate,
		RefererRewardRate: refererRewardRate,
	}
}

func (f *FeeRate) TotalFeeRate() sdk.Dec {
	return f.LPRewardRate.Add(f.RefererRewardRate).Add(f.RepurchaseRate)
}

type Order struct {
	OrderID      string        `json:"order_id"`
	DexID        uint32        `json:"dex_id"`
	From         sdk.CUAddress `json:"from"`
	Referer      sdk.CUAddress `json:"referer"`
	Receiver     sdk.CUAddress `json:"receiver"`
	CreatedTime  int64         `json:"created_time"`
	ExpiredTime  int64         `json:"expired_time"`
	FinishedTime int64         `json:"finished_time"`
	Status       byte          `json:"status"`
	Side         byte          `json:"side"`
	BaseSymbol   sdk.Symbol    `json:"base_symbol"`
	QuoteSymbol  sdk.Symbol    `json:"quote_symbol"`
	Price        sdk.Dec       `json:"price"`
	AmountIn     sdk.Int       `json:"amount_int"`
	LockedFund   sdk.Int       `json:"locked_fund"`
	FeeRate      *FeeRate      `json:"fee_rate"`
}

func (o *Order) IsFinished() bool {
	return o.Status == OrderStatusFilled || o.Status == OrderStatusCanceled || o.Status == OrderStatusExpired
}

func (o *Order) RemainQuantity() sdk.Int {
	if o.Side == OrderSideSell {
		return o.LockedFund
	}
	return o.LockedFund.ToDec().Quo(o.Price).TruncateInt()
}

func (o *Order) LessThan(order *Order) bool {
	if o.Price.Equal(order.Price) {
		if o.CreatedTime == order.CreatedTime {
			return o.OrderID < order.OrderID
		}
		return o.CreatedTime < order.CreatedTime
	}
	return o.Price.LT(order.Price)
}

type OrderByCreatedTime []*Order

func (o OrderByCreatedTime) Len() int {
	return len(o)
}

func (o OrderByCreatedTime) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o OrderByCreatedTime) Less(i, j int) bool {
	return o[i].CreatedTime < o[j].CreatedTime
}

type ResOrder struct {
	OrderID        string        `json:"order_id"`
	From           sdk.CUAddress `json:"from"`
	Referer        sdk.CUAddress `json:"referer"`
	Receiver       sdk.CUAddress `json:"receiver"`
	CreatedTime    int64         `json:"created_time"`
	ExpiredTime    int64         `json:"expired_time"`
	FinishedTime   int64         `json:"finished_time"`
	Status         byte          `json:"status"`
	Side           byte          `json:"side"`
	BaseSymbol     sdk.Symbol    `json:"base_symbol"`
	QuoteSymbol    sdk.Symbol    `json:"quote_symbol"`
	Price          sdk.Dec       `json:"price"`
	AmountIn       sdk.Int       `json:"amount_int"`
	LockedFund     sdk.Int       `json:"locked_fund"`
	RemainQuantity sdk.Int       `json:"remain_quantity"`
	DexID          uint32        `json:"dex_id"`
	FeeRate        *FeeRate      `json:"fee_rate"`
}

func NewResOrder(order *Order) *ResOrder {
	return &ResOrder{
		OrderID:        order.OrderID,
		From:           order.From,
		Referer:        order.Referer,
		Receiver:       order.Receiver,
		CreatedTime:    order.CreatedTime,
		ExpiredTime:    order.ExpiredTime,
		FinishedTime:   order.FinishedTime,
		Status:         order.Status,
		Side:           order.Side,
		BaseSymbol:     order.BaseSymbol,
		QuoteSymbol:    order.QuoteSymbol,
		Price:          order.Price,
		AmountIn:       order.AmountIn,
		LockedFund:     order.LockedFund,
		RemainQuantity: order.RemainQuantity(),
		DexID:          order.DexID,
		FeeRate:        order.FeeRate,
	}
}

func NewResOrders(orders []*Order) []*ResOrder {
	ret := make([]*ResOrder, len(orders))
	for i := range orders {
		ret[i] = NewResOrder(orders[i])
	}
	return ret
}
