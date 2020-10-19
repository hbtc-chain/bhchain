package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type MsgSynGasPrice struct {
	From     string               `json:"from"`
	Height   uint64               `json:"height"`
	GasPrice []sdk.TokensGasPrice `json:"gas_price"`
}

//NewMsgNewToken is a constructor function for MsgTokenNew
func NewMsgSynGasPrice(from string, height uint64, tokensgasprice []sdk.TokensGasPrice) MsgSynGasPrice {
	return MsgSynGasPrice{
		From:     from,
		Height:   height,
		GasPrice: tokensgasprice,
	}
}

func (msg MsgSynGasPrice) Route() string { return RouterKey }
func (msg MsgSynGasPrice) Type() string  { return TypeMsgSynGasPrice }

// ValidateBasic runs stateless checks on the message
func (msg MsgSynGasPrice) ValidateBasic() sdk.Error {
	_, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", msg.From))
	}
	if len(msg.GasPrice) == 0 {
		return sdk.ErrInvalidTx("empty gas price list")
	}

	return nil
}

func (msg MsgSynGasPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSynGasPrice) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

func (msg MsgSynGasPrice) IsSettleOnlyMsg() bool {
	return true
}
