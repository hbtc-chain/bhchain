package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	TypeMsgAddLiquidity    = "addliquidity"
	TypeMsgRemoveLiquidity = "removeliquidity"
	TypeMsgSwapExactIn     = "swapexactin"
	TypeMsgSwapExactOut    = "swapexactout"
	TypeMsgLimitSwap       = "limitswap"
	TypeMsgCancelLimitSwap = "cancellimitswap"
	TypeMsgClaimEarning    = "withdrawearning"
)

type MsgAddLiquidity struct {
	From            sdk.CUAddress `json:"from"`
	TokenA          sdk.Symbol    `json:"token_a"`
	TokenB          sdk.Symbol    `json:"token_b"`
	MinTokenAAmount sdk.Int       `json:"min_token_a_amount"`
	MinTokenBAmount sdk.Int       `json:"min_token_b_amount"`
	ExpiredAt       int64         `json:"expired_at"`
}

func NewMsgAddLiquidity(from sdk.CUAddress, tokenA, tokenB sdk.Symbol,
	minTokenAAmount, minTokenBAmount sdk.Int, expiredAt int64) MsgAddLiquidity {
	return MsgAddLiquidity{
		From:            from,
		TokenA:          tokenA,
		TokenB:          tokenB,
		MinTokenAAmount: minTokenAAmount,
		MinTokenBAmount: minTokenBAmount,
		ExpiredAt:       expiredAt,
	}
}

func (msg MsgAddLiquidity) Route() string {
	return RouterKey
}

func (msg MsgAddLiquidity) Type() string {
	return TypeMsgAddLiquidity
}

func (msg MsgAddLiquidity) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from address: %s is invalid", msg.From.String()))
	}
	if !msg.TokenA.IsValidTokenName() || !msg.TokenB.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid token")
	}
	if msg.TokenA == msg.TokenB {
		return sdk.ErrInvalidSymbol("same token")
	}
	if !msg.MinTokenAAmount.IsPositive() || !msg.MinTokenAAmount.IsPositive() {
		return sdk.ErrInvalidAmount("min token amount should be positive")
	}
	return nil
}

func (msg MsgAddLiquidity) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAddLiquidity) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgRemoveLiquidity struct {
	From      sdk.CUAddress `json:"from"`
	TokenA    sdk.Symbol    `json:"token_a"`
	TokenB    sdk.Symbol    `json:"token_b"`
	Liquidity sdk.Int       `json:"liquidity"`
	ExpiredAt int64         `json:"expired_at"`
}

func NewMsgRemoveLiquidity(from sdk.CUAddress, tokenA, tokenB sdk.Symbol, liquidity sdk.Int, expiredAt int64) MsgRemoveLiquidity {
	return MsgRemoveLiquidity{
		From:      from,
		TokenA:    tokenA,
		TokenB:    tokenB,
		Liquidity: liquidity,
		ExpiredAt: expiredAt,
	}
}

func (msg MsgRemoveLiquidity) Route() string {
	return RouterKey
}

func (msg MsgRemoveLiquidity) Type() string {
	return TypeMsgRemoveLiquidity
}

func (msg MsgRemoveLiquidity) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From address: %s is invalid", msg.From.String()))
	}
	if !msg.TokenA.IsValidTokenName() || !msg.TokenB.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid token")
	}
	if msg.TokenA == msg.TokenB {
		return sdk.ErrInvalidSymbol("same token")
	}
	if !msg.Liquidity.IsPositive() {
		return sdk.ErrInvalidAmount("liquidity should be positive")
	}
	return nil
}

func (msg MsgRemoveLiquidity) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRemoveLiquidity) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgSwapExactIn struct {
	From         sdk.CUAddress `json:"from"`
	Referer      sdk.CUAddress `json:"referer"`
	Receiver     sdk.CUAddress `json:"receiver"`
	AmountIn     sdk.Int       `json:"amount_in"`
	MinAmountOut sdk.Int       `json:"min_amount_out"`
	SwapPath     []sdk.Symbol  `json:"swap_path"`
	ExpiredAt    int64         `json:"expired_at"`
}

func NewMsgSwapExactIn(from, referer, receiver sdk.CUAddress, amountIn, minAmountOut sdk.Int,
	path []sdk.Symbol, expiredAt int64) MsgSwapExactIn {
	return MsgSwapExactIn{
		From:         from,
		Referer:      referer,
		Receiver:     receiver,
		AmountIn:     amountIn,
		MinAmountOut: minAmountOut,
		SwapPath:     path,
		ExpiredAt:    expiredAt,
	}
}

func (msg MsgSwapExactIn) Route() string {
	return RouterKey
}

func (msg MsgSwapExactIn) Type() string {
	return TypeMsgSwapExactIn
}

func (msg MsgSwapExactIn) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from address: %s is invalid", msg.From.String()))
	}
	if !msg.Referer.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("referer address: %s is invalid", msg.From.String()))
	}
	if !msg.Receiver.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("receiver address: %s is invalid", msg.From.String()))
	}
	if len(msg.SwapPath) < 2 {
		return sdk.ErrInvalidSymbol("length of path should be larger than 2")
	}
	for i := range msg.SwapPath {
		if !msg.SwapPath[i].IsValidTokenName() {
			return sdk.ErrInvalidSymbol("invalid symbol")
		}
		if i > 0 && msg.SwapPath[i] == msg.SwapPath[i-1] {
			return sdk.ErrInvalidSymbol("swap tokens are same")
		}
	}
	if !msg.AmountIn.IsPositive() || msg.MinAmountOut.IsNegative() {
		return sdk.ErrInvalidAmount("token amount should be positive")
	}
	return nil
}

func (msg MsgSwapExactIn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSwapExactIn) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgSwapExactOut struct {
	From        sdk.CUAddress `json:"from"`
	Referer     sdk.CUAddress `json:"referer"`
	Receiver    sdk.CUAddress `json:"receiver"`
	MaxAmountIn sdk.Int       `json:"max_amount_in"`
	AmountOut   sdk.Int       `json:"amount_out"`
	SwapPath    []sdk.Symbol  `json:"swap_path"`
	ExpiredAt   int64         `json:"expired_at"`
}

func NewMsgSwapExactOut(from, referer, receiver sdk.CUAddress, amountOut, maxAmountIn sdk.Int,
	path []sdk.Symbol, expiredAt int64) MsgSwapExactOut {
	return MsgSwapExactOut{
		From:        from,
		Referer:     referer,
		Receiver:    receiver,
		AmountOut:   amountOut,
		MaxAmountIn: maxAmountIn,
		SwapPath:    path,
		ExpiredAt:   expiredAt,
	}
}

func (msg MsgSwapExactOut) Route() string {
	return RouterKey
}

func (msg MsgSwapExactOut) Type() string {
	return TypeMsgSwapExactOut
}

func (msg MsgSwapExactOut) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from address: %s is invalid", msg.From.String()))
	}
	if !msg.Referer.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("referer address: %s is invalid", msg.From.String()))
	}
	if !msg.Receiver.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("receiver address: %s is invalid", msg.From.String()))
	}
	if len(msg.SwapPath) < 2 {
		return sdk.ErrInvalidSymbol("length of path should be larger than 2")
	}
	for i := range msg.SwapPath {
		if !msg.SwapPath[i].IsValidTokenName() {
			return sdk.ErrInvalidSymbol("invalid symbol")
		}
		if i > 0 && msg.SwapPath[i] == msg.SwapPath[i-1] {
			return sdk.ErrInvalidSymbol("swap tokens are same")
		}
	}
	if !msg.MaxAmountIn.IsPositive() || !msg.AmountOut.IsPositive() {
		return sdk.ErrInvalidAmount("token amount should be positive")
	}
	return nil
}

func (msg MsgSwapExactOut) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSwapExactOut) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgLimitSwap struct {
	From        sdk.CUAddress `json:"from"`
	OrderID     string        `json:"order_id"`
	Referer     sdk.CUAddress `json:"referer"`
	Receiver    sdk.CUAddress `json:"receiver"`
	AmountIn    sdk.Int       `json:"amount_in"`
	Price       sdk.Dec       `json:"price"`
	BaseSymbol  sdk.Symbol    `json:"base_symbol"`
	QuoteSymbol sdk.Symbol    `json:"quote_symbol"`
	Side        int           `json:"side"`
	ExpiredAt   int64         `json:"expired_at"`
}

func NewMsgLimitSwap(orderID string, from, referer, receiver sdk.CUAddress, amountIn sdk.Int, price sdk.Dec,
	baseSymbol, quoteSymbol sdk.Symbol, side int, expiredAt int64) MsgLimitSwap {
	return MsgLimitSwap{
		From:        from,
		OrderID:     orderID,
		Referer:     referer,
		Receiver:    receiver,
		AmountIn:    amountIn,
		Price:       price,
		BaseSymbol:  baseSymbol,
		QuoteSymbol: quoteSymbol,
		Side:        side,
		ExpiredAt:   expiredAt,
	}
}

func (msg MsgLimitSwap) Route() string {
	return RouterKey
}

func (msg MsgLimitSwap) Type() string {
	return TypeMsgLimitSwap
}

func (msg MsgLimitSwap) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from address: %s is invalid", msg.From.String()))
	}
	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("Order id %s is invalid", msg.OrderID))
	}
	if !msg.Referer.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("referer address: %s is invalid", msg.From.String()))
	}
	if !msg.Receiver.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("receiver address: %s is invalid", msg.From.String()))
	}
	if msg.BaseSymbol >= msg.QuoteSymbol {
		return sdk.ErrInvalidSymbol("wrong symbol sequence")
	}
	if !msg.BaseSymbol.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid base symbol")
	}
	if !msg.QuoteSymbol.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid quote symbol")
	}
	if !msg.AmountIn.IsPositive() {
		return sdk.ErrInvalidAmount("token amount should be positive")
	}
	if !msg.Price.IsPositive() {
		return sdk.ErrInvalidAmount("price should be positive")
	}
	if msg.Side != OrderSideBuy && msg.Side != OrderSideSell {
		return sdk.ErrInvalidTx("invalid order side")
	}
	return nil
}

func (msg MsgLimitSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgLimitSwap) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgCancelLimitSwap struct {
	From     sdk.CUAddress `json:"from"`
	OrderIDs []string      `json:"order_ids"`
}

func NewMsgCancelLimitSwap(from sdk.CUAddress, orderIDs []string) MsgCancelLimitSwap {
	return MsgCancelLimitSwap{
		From:     from,
		OrderIDs: orderIDs,
	}
}

func (msg MsgCancelLimitSwap) Route() string {
	return RouterKey
}

func (msg MsgCancelLimitSwap) Type() string {
	return TypeMsgCancelLimitSwap
}

func (msg MsgCancelLimitSwap) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from address: %s is invalid", msg.From.String()))
	}
	if len(msg.OrderIDs) == 0 {
		return sdk.ErrInvalidTx("empty order id list")
	}
	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return sdk.ErrInvalidTx("invalid order id list")
	}
	return nil
}

func (msg MsgCancelLimitSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCancelLimitSwap) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgClaimEarning struct {
	From   sdk.CUAddress `json:"from"`
	TokenA sdk.Symbol    `json:"token_a"`
	TokenB sdk.Symbol    `json:"token_b"`
}

func NewMsgClaimEarning(from sdk.CUAddress, tokenA, tokenB sdk.Symbol) MsgClaimEarning {
	return MsgClaimEarning{
		From:   from,
		TokenA: tokenA,
		TokenB: tokenB,
	}
}

func (msg MsgClaimEarning) Route() string {
	return RouterKey
}

func (msg MsgClaimEarning) Type() string {
	return TypeMsgClaimEarning
}

func (msg MsgClaimEarning) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from address: %s is invalid", msg.From.String()))
	}
	if !msg.TokenA.IsValidTokenName() || !msg.TokenB.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid token")
	}
	if msg.TokenA == msg.TokenB {
		return sdk.ErrInvalidSymbol("same token")
	}

	return nil
}

func (msg MsgClaimEarning) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgClaimEarning) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}
