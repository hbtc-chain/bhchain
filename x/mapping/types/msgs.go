package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

const RouterKey = ModuleName // this was defined in key.go file

const (
	TypeMsgMappingSwap      = "mapping_swap"
	TypeMsgCreateFreeSwap   = "free_swap"
	TypeMsgCreateDirectSwap = "direct_swap"
	TypeMsgSwapSymbol       = "swap_symbol"
	TypeMsgCancelSwap       = "cancel_swap"
)

var _ sdk.Msg = MsgMappingSwap{}

type MsgMappingSwap struct {
	From        sdk.CUAddress `json:"from"`
	IssueSymbol sdk.Symbol    `json:"issue_symbol"`
	Coins       sdk.Coins     `json:"coins"`
}

func NewMsgMappingSwap(from sdk.CUAddress, issueSymbol sdk.Symbol, amount sdk.Coins) MsgMappingSwap {
	return MsgMappingSwap{from, issueSymbol, amount}
}

func (msg MsgMappingSwap) Route() string { return RouterKey }

func (msg MsgMappingSwap) Type() string { return TypeMsgMappingSwap }

func (msg MsgMappingSwap) ValidateBasic() sdk.Error {
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid: %s", msg.From.String()))
	}
	if len(msg.IssueSymbol) <= 0 {
		return sdk.ErrInvalidSymbol("issue symbol should not be empty")
	}
	if !msg.IssueSymbol.IsValid() {
		return sdk.ErrInvalidSymbol("invalid issue symbol")
	}
	if msg.Coins.Len() != 1 {
		return ErrInvalidSwapAmount(DefaultCodespace, "swap amount should contain exactly 1 coin")
	}
	return nil
}

func (msg MsgMappingSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgMappingSwap) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

var _ sdk.Msg = MsgCreateDirectSwap{}

type MsgCreateDirectSwap struct {
	From     string         `json:"from"`
	OrderID  string         `json:"order_id"`
	SwapInfo DirectSwapInfo `json:"swap_info"`
}

func NewMsgCreateDirectSwap(from string, orderid string, swapInfo DirectSwapInfo) MsgCreateDirectSwap {
	return MsgCreateDirectSwap{from, orderid, swapInfo}
}

func (msg MsgCreateDirectSwap) Route() string { return RouterKey }

func (msg MsgCreateDirectSwap) Type() string { return TypeMsgCreateDirectSwap }

func (msg MsgCreateDirectSwap) ValidateBasic() sdk.Error {
	if msg.From == "" || !sdk.IsValidAddr(msg.From) || !sdk.IsValidAddr(msg.SwapInfo.ReceiveAddr) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v, %v", msg.From, msg.SwapInfo.ReceiveAddr))
	}

	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("invalid orderid:%v", msg.OrderID))
	}

	if !msg.SwapInfo.SrcSymbol.IsValid() || !msg.SwapInfo.TargetSymbol.IsValid() ||
		msg.SwapInfo.TargetSymbol.String() == msg.SwapInfo.SrcSymbol.String() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("invalid symbol:%v, %v", msg.SwapInfo.SrcSymbol, msg.SwapInfo.TargetSymbol))
	}

	if msg.SwapInfo.Amount.LTE(sdk.ZeroInt()) || msg.SwapInfo.SwapAmount.LTE(sdk.ZeroInt()) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("swap count invalid:%v, %v", msg.SwapInfo.Amount, msg.SwapInfo.SwapAmount))
	}

	return nil
}

func (msg MsgCreateDirectSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreateDirectSwap) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

var _ sdk.Msg = MsgCreateFreeSwap{}

type MsgCreateFreeSwap struct {
	From     string       `json:"from"`
	OrderID  string       `json:"order_id"`
	SwapInfo FreeSwapInfo `json:"swap_info"`
}

func NewMsgCreateFreeSwap(from string, orderid string, swapInfo FreeSwapInfo) MsgCreateFreeSwap {
	return MsgCreateFreeSwap{from, orderid, swapInfo}
}

func (msg MsgCreateFreeSwap) Route() string { return RouterKey }

func (msg MsgCreateFreeSwap) Type() string { return TypeMsgCreateFreeSwap }

func (msg MsgCreateFreeSwap) ValidateBasic() sdk.Error {
	if msg.From == "" || !sdk.IsValidAddr(msg.From) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", msg.From))
	}

	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("invalid orderid:%v", msg.OrderID))
	}

	if !msg.SwapInfo.SrcSymbol.IsValid() || !msg.SwapInfo.TargetSymbol.IsValid() ||
		msg.SwapInfo.TargetSymbol.String() == msg.SwapInfo.SrcSymbol.String() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("invalid symbol:%v, %v", msg.SwapInfo.SrcSymbol, msg.SwapInfo.TargetSymbol))
	}

	if msg.SwapInfo.TotalAmount.LTE(sdk.ZeroInt()) || msg.SwapInfo.MinSwapAmount.LTE(sdk.ZeroInt()) ||
		msg.SwapInfo.MaxSwapAmount.LTE(sdk.ZeroInt()) || msg.SwapInfo.MaxSwapAmount.LT(msg.SwapInfo.MinSwapAmount) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("swap count invalid:%v, %v", msg.SwapInfo.MinSwapAmount, msg.SwapInfo.MaxSwapAmount))
	}

	return nil
}

func (msg MsgCreateFreeSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreateFreeSwap) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

var _ sdk.Msg = MsgSwapSymbol{}

type MsgSwapSymbol struct {
	From       string  `json:"from"`
	DstOrderID string  `json:"dst_order_id"`
	SwapAmount sdk.Int `json:"swap_amount"`
	SwapType   int     `json:"swap_type"`
}

func NewMsgSwapSymbol(from string, orderid string, amount sdk.Int, swapType int) MsgSwapSymbol {
	return MsgSwapSymbol{from, orderid, amount, swapType}
}

func (msg MsgSwapSymbol) Route() string { return RouterKey }

func (msg MsgSwapSymbol) Type() string { return TypeMsgSwapSymbol }

func (msg MsgSwapSymbol) ValidateBasic() sdk.Error {
	if msg.From == "" || !sdk.IsValidAddr(msg.From) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", msg.From))
	}

	if msg.SwapType != SwapTypeFree && msg.SwapType != SwapTypeDirect {
		return sdk.ErrInvalidAddress(fmt.Sprintf("err swap type:%v", msg.SwapType))
	}

	if sdk.IsIllegalOrderID(msg.DstOrderID) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("invalid orderid:%v", msg.DstOrderID))
	}

	if msg.SwapAmount.LTE(sdk.ZeroInt()) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("swap count invalid:%v", msg.SwapAmount))
	}

	return nil
}

func (msg MsgSwapSymbol) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSwapSymbol) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

var _ sdk.Msg = MsgCancelSwap{}

type MsgCancelSwap struct {
	From     string `json:"from"`
	SwapType int    `json:"swap_type"`
	OrderID  string `json:"order_id"`
}

func NewMsgCancelSwap(from string, orderid string, swapType int) MsgCancelSwap {
	return MsgCancelSwap{from, swapType, orderid}
}

func (msg MsgCancelSwap) Route() string { return RouterKey }

func (msg MsgCancelSwap) Type() string { return TypeMsgCancelSwap }

func (msg MsgCancelSwap) ValidateBasic() sdk.Error {
	if msg.From == "" || !sdk.IsValidAddr(msg.From) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", msg.From))
	}

	if msg.SwapType != SwapTypeFree && msg.SwapType != SwapTypeDirect {
		return sdk.ErrInvalidAddress(fmt.Sprintf("err swap type:%v", msg.SwapType))
	}

	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("invalid orderid:%v", msg.OrderID))
	}

	return nil
}

func (msg MsgCancelSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCancelSwap) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}
