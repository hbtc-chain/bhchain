package types

import (
	"encoding/json"

	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	EventTypeAddLiquidity    = "add_liquidity"
	EventTypeRemoveLiquidity = "remove_liquidity"
	EventTypeSwap            = "swap"
	EventTypeCancelOrders    = "cancel_orders"
	EventTypeExpireOrders    = "expire_orders"
	EventTypeWithdrawEarning = "withdraw_earning"
	EventTypeMining          = "mining"

	AttributeKeyLiquidity  = "liquidity"
	AttributeKeySwapResult = "swap"
	AttributeKeyBurned     = "burned"
	AttributeKeyOrders     = "orders"
	AttributeKeyAmount     = "amount"
	AttributeKeyAddress    = "address"
)

type EventLiquidity struct {
	From                sdk.CUAddress `json:"from"`
	TokenA              sdk.Symbol    `json:"token_a"`
	TokenB              sdk.Symbol    `json:"token_b"`
	TokenAAmount        sdk.Int       `json:"token_a_amount"`
	TokenBAmount        sdk.Int       `json:"token_b_amount"`
	ChangedTokenAAmount sdk.Int       `json:"changed_a_amount"`
	ChangedTokenBAmount sdk.Int       `json:"changed_b_amount"`
}

func NewEventLiquidity(from sdk.CUAddress, tokenA, tokenB sdk.Symbol, tokenAAmount, tokenBAmount, changedAAmount, changedBAmount sdk.Int) *EventLiquidity {
	return &EventLiquidity{
		From:                from,
		TokenA:              tokenA,
		TokenB:              tokenB,
		TokenAAmount:        tokenAAmount,
		TokenBAmount:        tokenBAmount,
		ChangedTokenAAmount: changedAAmount,
		ChangedTokenBAmount: changedBAmount,
	}
}

func (e *EventLiquidity) String() string {
	bz, _ := json.Marshal(e)
	return string(bz)
}

type EventSwap struct {
	From         sdk.CUAddress `json:"from"`
	OrderID      string        `json:"order_id"`
	TokenA       sdk.Symbol    `json:"token_a"`
	TokenB       sdk.Symbol    `json:"token_b"`
	TokenAAmount sdk.Int       `json:"token_a_amount"`
	TokenBAmount sdk.Int       `json:"token_b_amount"`
	TokenIn      sdk.Symbol    `json:"token_in"`
	AmountIn     sdk.Int       `json:"amount_in"`
	AmountOut    sdk.Int       `json:"amount_out"`
}

func NewEventSwap(from sdk.CUAddress, orderID string, tokenA, tokenB, tokenIn sdk.Symbol,
	tokenAAmount, tokenBAmount, amountIn, amountOut sdk.Int) *EventSwap {
	return &EventSwap{
		From:         from,
		OrderID:      orderID,
		TokenA:       tokenA,
		TokenB:       tokenB,
		TokenAAmount: tokenAAmount,
		TokenBAmount: tokenBAmount,
		TokenIn:      tokenIn,
		AmountIn:     amountIn,
		AmountOut:    amountOut,
	}
}

type EventSwaps []*EventSwap

func (e *EventSwaps) String() string {
	bz, _ := json.Marshal(e)
	return string(bz)
}

type EventOrderStatusChanged struct {
	OrderIDs []string `json:"order_ids"`
}

func NewEventOrderStatusChanged(orderIDs []string) *EventOrderStatusChanged {
	return &EventOrderStatusChanged{
		OrderIDs: orderIDs,
	}
}

func (e *EventOrderStatusChanged) String() string {
	bz, _ := json.Marshal(e)
	return string(bz)
}
