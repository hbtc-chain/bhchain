package types

import (
	"fmt"
	sdk "github.com/hbtc-chain/bhchain/types"
)

//========MsgTokenNew
type MsgNewToken struct {
	From        sdk.CUAddress `json:"from" yaml:"from"`
	To          sdk.CUAddress `json:"to" yaml:"to"`
	Symbol      sdk.Symbol    `json:"symbol" yaml:"symbol"`
	Decimals    uint64        `json:"decimals" yaml:"decimals"`
	TotalSupply sdk.Int       `json:"total_supply" yaml:"total_supply"`
}

//NewMsgNewToken is a constructor function for MsgTokenNew
func NewMsgNewToken(from, to sdk.CUAddress, symbol string, decimals uint64, totalSupply sdk.Int) MsgNewToken {
	return MsgNewToken{
		From:        from,
		To:          to,
		Symbol:      sdk.Symbol(symbol),
		Decimals:    decimals,
		TotalSupply: totalSupply,
	}
}

func (msg MsgNewToken) Route() string { return RouterKey }
func (msg MsgNewToken) Type() string  { return TypeMsgNewToken }

// ValidateBasic runs stateless checks on the message
func (msg MsgNewToken) ValidateBasic() sdk.Error {
	if msg.From.Empty() || !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", msg.From))
	}

	if msg.To.Empty() || !msg.To.IsValidAddr() {
		return sdk.ErrInvalidAddress(fmt.Sprintf("to address can not be empty or invalid:%v", msg.To))
	}

	if !msg.Symbol.IsValidTokenName() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("symbol %v is invalid", msg.Symbol))
	}

	if msg.Decimals > sdk.Precision {
		return sdk.ErrTooMuchPrecision(fmt.Sprintf("maximum:%v, provided:%v", sdk.Precision, msg.Decimals))
	}

	if !msg.TotalSupply.IsPositive() {
		return sdk.ErrInvalidAmount(fmt.Sprintf("totalSupply %v is not positive", msg.TotalSupply))
	}

	return nil
}

func (msg MsgNewToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgNewToken) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}
