package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	cTypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
)

const (
	TypeMsgKeyGen              = "keygen"
	TypeMsgKeyGenWaitSign      = "keygenwaitsign"
	TypeMsgKeyGenFinish        = "keygenfinish"
	TypeMsgPreKeyGen           = "prekeygen"
	TypeMsgOpcuMigrationKeyGen = "opcumigrationkeygen"
)

type MsgKeyGen struct {
	OrderID string        `json:"order_id"` // client type in
	Symbol  sdk.Symbol    `json:"symbol"`
	From    sdk.CUAddress `json:"from"`
	To      sdk.CUAddress `json:"to"`
}

func NewMsgKeyGen(orderID string, symbol sdk.Symbol, from, to sdk.CUAddress) MsgKeyGen {
	return MsgKeyGen{
		OrderID: orderID,
		Symbol:  symbol,
		From:    from,
		To:      to,
	}
}

func (msg MsgKeyGen) Route() string {
	return RouterKey
}

func (msg MsgKeyGen) Type() string {
	return TypeMsgKeyGen
}

func (msg MsgKeyGen) ValidateBasic() sdk.Error {
	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidTx("OrderID is invalid")
	}
	if msg.Symbol.String() == "" || !msg.Symbol.IsValid() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("Invalid Symbol %s", msg.Symbol))
	}
	if msg.Symbol.String() == sdk.NativeToken {
		return sdk.ErrInvalidTx("No need to generate address for native token")
	}
	if msg.From == nil {
		return sdk.ErrInvalidTx("Message's from cu address is nil")
	}

	if msg.To == nil {
		return sdk.ErrInvalidTx("Message's to from cu address is nil")
	}
	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU address: %s is invalid", msg.From.String()))
	}
	if !msg.To.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("TO CU address: %s is invalid", msg.To.String()))
	}

	return nil
}

func (msg MsgKeyGen) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgKeyGen) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

type MsgKeyGenWaitSign struct {
	From     sdk.CUAddress         `json:"from"`
	OrderID  string                `json:"order_id"`
	PubKey   []byte                `json:"pubkey"`
	KeyNodes []sdk.CUAddress       `json:"key_nodes"`
	KeySigs  []cTypes.StdSignature `json:"key_sigs"`
	Epoch    uint64                `json:"epoch"`
}

func NewMsgKeyGenWaitSign(from sdk.CUAddress, orderId string, pubKey []byte, keyNodes []sdk.CUAddress, keysigs []cTypes.StdSignature, epoch uint64) MsgKeyGenWaitSign {
	return MsgKeyGenWaitSign{
		From:     from,
		OrderID:  orderId,
		PubKey:   pubKey,
		KeyNodes: keyNodes,
		KeySigs:  keysigs,
		Epoch:    epoch,
	}
}

func (msg MsgKeyGenWaitSign) Route() string { return RouterKey }
func (msg MsgKeyGenWaitSign) Type() string  { return TypeMsgKeyGenWaitSign }

// ValidateBasic runs stateless checks on the message
func (msg MsgKeyGenWaitSign) ValidateBasic() sdk.Error {
	if msg.From == nil || !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU address: %s is invalid", msg.From.String()))
	}

	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidOrder("OrderID is invalid")
	}
	if len(msg.PubKey) == 0 {
		return sdk.ErrInvalidPubKey("PubKey's length is zero")
	}

	if len(msg.KeyNodes) == 0 {
		return sdk.ErrInvalidAddress("No proposers")
	}

	for _, keyNode := range msg.KeyNodes {
		if keyNode.Empty() || !keyNode.IsValidAddr() {
			return sdk.ErrInvalidAddress(fmt.Sprintf("keyNode address can not be empty or invalid:%v", keyNode.String()))
		}
	}

	if len(msg.KeySigs) == 0 {
		return sdk.ErrInvalidTx("KeySigs is empty")
	}

	if msg.Epoch == 0 {
		return sdk.ErrInvalidTx("Invalid epoch")
	}

	return nil
}

func (msg MsgKeyGenWaitSign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgKeyGenWaitSign) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

func (msg MsgKeyGenWaitSign) IsSettleOnlyMsg() bool {
	return true
}

type MsgPreKeyGen struct {
	OrderIDs []string      `json:"order_ids"`
	From     sdk.CUAddress `json:"from"`
}

func NewMsgPreKeyGen(orderIDs []string, from sdk.CUAddress) *MsgPreKeyGen {
	return &MsgPreKeyGen{
		OrderIDs: orderIDs,
		From:     from,
	}
}

func (msg MsgPreKeyGen) Route() string {
	return RouterKey
}

func (msg MsgPreKeyGen) Type() string {
	return TypeMsgPreKeyGen
}

func (msg MsgPreKeyGen) ValidateBasic() sdk.Error {
	if len(msg.OrderIDs) == 0 {
		return sdk.ErrInvalidTx("Empty order id list")
	}
	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return sdk.ErrInvalidTx("Invalid order id list")
	}
	if msg.From == nil || !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU address: %s is invalid", msg.From.String()))
	}
	return nil
}

func (msg MsgPreKeyGen) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgPreKeyGen) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

func (msg MsgPreKeyGen) IsSettleOnlyMsg() bool {
	return true
}

type MsgKeyGenFinish struct {
	OrderID   string        `json:"order_ids"`
	Signature []byte        `json:"signature"`
	Validator sdk.CUAddress `json:"validator"`
}

func NewMsgKeyGenFinish(orderID string, signature []byte, validator sdk.CUAddress) *MsgKeyGenFinish {
	return &MsgKeyGenFinish{
		OrderID:   orderID,
		Validator: validator,
		Signature: signature,
	}
}

func (msg MsgKeyGenFinish) Route() string { return RouterKey }
func (msg MsgKeyGenFinish) Type() string  { return TypeMsgKeyGenFinish }

// ValidateBasic runs stateless checks on the message
func (msg MsgKeyGenFinish) ValidateBasic() sdk.Error {
	if msg.Validator == nil || !msg.Validator.IsValidAddr() {
		return sdk.ErrInvalidTx("Message's from cu address is invalid")
	}

	return nil
}

func (msg MsgKeyGenFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgKeyGenFinish) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.Validator}
}

func (msg MsgKeyGenFinish) IsSettleOnlyMsg() bool {
	return true
}

type MsgOpcuMigrationKeyGen struct {
	OrderIDs []string      `json:"order_ids"`
	From     sdk.CUAddress `json:"from"`
}

func NewMsgOpcuMigrationKeyGen(orderIDs []string, from sdk.CUAddress) *MsgOpcuMigrationKeyGen {
	return &MsgOpcuMigrationKeyGen{
		OrderIDs: orderIDs,
		From:     from,
	}
}

func (msg MsgOpcuMigrationKeyGen) Route() string { return RouterKey }
func (msg MsgOpcuMigrationKeyGen) Type() string  { return TypeMsgOpcuMigrationKeyGen }

// ValidateBasic runs stateless checks on the message
func (msg MsgOpcuMigrationKeyGen) ValidateBasic() sdk.Error {
	if len(msg.OrderIDs) == 0 {
		return sdk.ErrInvalidTx("Empty order id list")
	}
	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return sdk.ErrInvalidTx("Invalid order id list")
	}
	if msg.From == nil || !msg.From.IsValidAddr() {
		return sdk.ErrInvalidTx("Message's from cu address is invalid")
	}

	return nil
}

func (msg MsgOpcuMigrationKeyGen) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgOpcuMigrationKeyGen) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.From}
}

func (msg MsgOpcuMigrationKeyGen) IsSettleOnlyMsg() bool {
	return true
}

const (
	TypeMsgNewOpCU = "new_op_cu"
)

//========MsgNewOpCU
type MsgNewOpCU struct {
	Symbol      string        `json:"symbol"`
	From        sdk.CUAddress `json:"from"`
	OpCUAddress sdk.CUAddress `json:"opcu_address"`
}

//NewMsgKeyGen is a constructor function for MsgNewOpCU
func NewMsgNewOpCU(symbol string, opCUAddress, from sdk.CUAddress) MsgNewOpCU {
	return MsgNewOpCU{
		Symbol:      symbol,
		OpCUAddress: opCUAddress,
		From:        from,
	}
}

func (msg MsgNewOpCU) Route() string { return RouterKey }
func (msg MsgNewOpCU) Type() string  { return TypeMsgNewOpCU }

// ValidateBasic runs stateless checks on the message
func (msg MsgNewOpCU) ValidateBasic() sdk.Error {
	if msg.Symbol == "" {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("Invalid Symbol %v", msg.Symbol))
	}

	if msg.Symbol == sdk.NativeToken {
		return sdk.ErrInvalidTx(fmt.Sprintf("No need to generate operation CU for native token"))
	}

	if !msg.OpCUAddress.IsValidAddr() {
		return sdk.ErrInvalidAddress(fmt.Sprintf("invalid operation CU address :%v", msg.OpCUAddress))
	}

	if !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", msg.From))

	}
	return nil
}

func (msg MsgNewOpCU) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgNewOpCU) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{sdk.CUAddress(msg.From)}
}
