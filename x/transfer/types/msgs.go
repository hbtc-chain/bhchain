package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgConfirmedDeposit{}
	_ sdk.Msg = &MsgCollectWaitSign{}
	_ sdk.Msg = &MsgCollectSignFinish{}
	_ sdk.Msg = &MsgCollectFinish{}
	_ sdk.Msg = &MsgWithdrawal{}
	_ sdk.Msg = &MsgWithdrawalWaitSign{}
	_ sdk.Msg = &MsgWithdrawalSignFinish{}
	_ sdk.Msg = &MsgWithdrawalFinish{}
	_ sdk.Msg = &MsgSysTransferWaitSign{}
	_ sdk.Msg = &MsgSysTransferSignFinish{}
	_ sdk.Msg = &MsgSysTransferFinish{}
	_ sdk.Msg = &MsgSysTransfer{}
	_ sdk.Msg = &MsgOpcuAssetTransfer{}
	_ sdk.Msg = &MsgOpcuAssetTransferWaitSign{}
	_ sdk.Msg = &MsgOpcuAssetTransferSignFinish{}
	_ sdk.Msg = &MsgOpcuAssetTransferFinish{}
	_ sdk.Msg = &MsgSend{}
	_ sdk.Msg = &MsgMultiSend{}
	_ sdk.Msg = &MsgOrderRetry{}
	_ sdk.Msg = &MsgCancelWithdrawal{}
)

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	FromAddress sdk.CUAddress `json:"from_address" yaml:"from_address"`
	ToAddress   sdk.CUAddress `json:"to_address" yaml:"to_address"`
	Amount      sdk.Coins     `json:"amount" yaml:"amount"`
}

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(fromAddr, toAddr sdk.CUAddress, amount sdk.Coins) MsgSend {
	return MsgSend{FromAddress: fromAddr, ToAddress: toAddr, Amount: amount}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return "send" }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() sdk.Error {
	if !msg.FromAddress.IsValidAddr() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}

	if !msg.ToAddress.IsValidAddr() {
		return sdk.ErrInvalidAddress("invalid receipt address")
	}

	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("send amount is invalid: " + msg.Amount.String())
	}
	if !msg.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("send amount must be positive")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.FromAddress}

}

// MsgMultiSend - high level transaction of the coin module
type MsgMultiSend struct {
	Inputs    []Input  `json:"inputs" yaml:"inputs"`
	Outputs   []Output `json:"outputs" yaml:"outputs"`
	MaxHeight uint64   `json:"max_height" yaml:"max_height"`
}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in []Input, out []Output, maxHeight uint64) MsgMultiSend {
	return MsgMultiSend{Inputs: in, Outputs: out, MaxHeight: maxHeight}
}

// Route Implements Msg
func (msg MsgMultiSend) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMultiSend) Type() string { return "multisend" }

// ValidateBasic Implements Msg.
func (msg MsgMultiSend) ValidateBasic() sdk.Error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs(DefaultCodespace).TraceSDK("")
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs(DefaultCodespace).TraceSDK("")
	}

	return ValidateInputsOutputs(msg.Inputs, msg.Outputs)
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []sdk.CUAddress {
	addrs := make([]sdk.CUAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

// Input models transaction input
type Input struct {
	Address sdk.CUAddress `json:"address" yaml:"address"`
	Coins   sdk.Coins     `json:"coins" yaml:"coins"`
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() sdk.Error {
	if len(in.Address) == 0 {
		return sdk.ErrInvalidAddress(in.Address.String())
	}
	if !in.Coins.IsValid() {
		return sdk.ErrInvalidCoins(in.Coins.String())
	}
	if !in.Coins.IsAllPositive() {
		return sdk.ErrInvalidCoins(in.Coins.String())
	}
	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
func NewInput(addr sdk.CUAddress, coins sdk.Coins) Input {
	return Input{
		Address: addr,
		Coins:   coins,
	}
}

// Output models transaction outputs
type Output struct {
	Address sdk.CUAddress `json:"address" yaml:"address"`
	Coins   sdk.Coins     `json:"coins" yaml:"coins"`
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() sdk.Error {
	if len(out.Address) == 0 {
		return sdk.ErrInvalidAddress(out.Address.String())
	}
	if !out.Coins.IsValid() {
		return sdk.ErrInvalidCoins(out.Coins.String())
	}
	if !out.Coins.IsAllPositive() {
		return sdk.ErrInvalidCoins(out.Coins.String())
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
func NewOutput(addr sdk.CUAddress, coins sdk.Coins) Output {
	return Output{
		Address: addr,
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that each respective input and output is
// valid and that the sum of inputs is equal to the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) sdk.Error {
	var totalIn, totalOut sdk.Coins

	for _, in := range inputs {
		if err := in.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalIn = totalIn.Add(in.Coins)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalOut = totalOut.Add(out.Coins)
	}

	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInputOutputMismatch(DefaultCodespace)
	}

	return nil
}

//________________________________
type MsgDeposit struct {
	FromCU    sdk.CUAddress `json:"from_cu"`
	ToCU      sdk.CUAddress `json:"to_cu"`
	ToAddress string        `json:"to_adddress"`
	Symbol    sdk.Symbol    `json:"symbol"`
	Amount    sdk.Int       `json:"amount"`
	Index     uint16        `json:"index"`
	Txhash    string        `json:"txhash"`
	OrderID   string        `json:"order_id"`
	Memo      string        `json:"memo"`
}

func NewMsgDeposit(fromCU, toCU sdk.CUAddress, symbol sdk.Symbol, toAddr, hash, orderID, memo string, amount sdk.Int, index uint16) MsgDeposit {
	return MsgDeposit{
		FromCU:    fromCU,
		ToCU:      toCU, //if ToCU is not a OP CU,FromCU must = ToCU
		ToAddress: toAddr,
		Symbol:    symbol,
		Amount:    amount,
		Index:     index,
		Txhash:    hash,
		OrderID:   orderID,
		Memo:      memo,
	}
}

//nolint
func (msg MsgDeposit) Route() string { return RouterKey }
func (msg MsgDeposit) Type() string  { return "deposit" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeposit) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	return []sdk.CUAddress{msg.FromCU}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidTx("Invalid order id")
	}
	if msg.FromCU == nil || !msg.FromCU.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU address: %s is invalid", msg.FromCU.String()))
	}
	if msg.ToCU == nil || !msg.ToCU.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("TO CU address: %s is invalid", msg.ToCU.String()))
	}
	if !msg.Symbol.IsValid() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("Invalid Symbol %s", msg.Symbol))
	}

	if msg.Txhash == "" {
		return sdk.ErrInvalidTx("Invalid tx hash")
	}
	if msg.ToAddress == "" {
		return ErrBadAddress(DefaultCodespace)
	}

	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidAmount("Amount is not positive")
	}
	return nil
}

func (msg MsgDeposit) IsSettleOnlyMsg() bool {
	return false
}

type MsgConfirmedDeposit struct {
	From            sdk.CUAddress `json:"from_cu"`
	ValidOrderIDs   []string      `json:"validorderids"`
	InvalidOrderIDs []string      `json:"invalidorderids"`
}

func NewMsgConfirmedDeposit(from sdk.CUAddress, validaOrderIDs, invalidOrderIDs []string) MsgConfirmedDeposit {
	return MsgConfirmedDeposit{
		From:            from,
		ValidOrderIDs:   validaOrderIDs,
		InvalidOrderIDs: invalidOrderIDs,
	}
}

//nolint
func (msg MsgConfirmedDeposit) Route() string { return RouterKey }
func (msg MsgConfirmedDeposit) Type() string  { return "confirmed_deposit" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgConfirmedDeposit) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	return []sdk.CUAddress{msg.From}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgConfirmedDeposit) GetSignBytes() []byte {
	// TODO: check whether it can be deleted
	if len(msg.ValidOrderIDs) <= 0 {
		msg.ValidOrderIDs = []string{}
	}
	if len(msg.InvalidOrderIDs) <= 0 {
		msg.InvalidOrderIDs = []string{}
	}
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgConfirmedDeposit) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	if sdk.IsIllegalOrderIDList(msg.ValidOrderIDs) {
		return sdk.ErrInvalidTx("Invalid order id list")
	}
	if sdk.IsIllegalOrderIDList(msg.InvalidOrderIDs) {
		return sdk.ErrInvalidTx("Invalid order id list")
	}
	if msg.From == nil || !msg.From.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU address: %s is invalid", msg.From.String()))
	}
	return nil
}

func (msg MsgConfirmedDeposit) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgCollectWaitSign struct {
	OrderIDs    []string `json:"order_ids"`
	RawData     []byte   `json:"raw_data"`
	CollectToCU string   `json:"collect_to_cu"`
	Validator   string   `json:"validator"` //是谁在归集
}

func NewMsgCollectWaitSign(collectToCU, valAddr string, ids []string, rawdata []byte) MsgCollectWaitSign {
	msg := MsgCollectWaitSign{
		OrderIDs:    make([]string, len(ids)),
		RawData:     make([]byte, len(rawdata)),
		Validator:   valAddr,
		CollectToCU: collectToCU,
	}

	copy(msg.OrderIDs, ids)
	copy(msg.RawData, rawdata)
	return msg
}

//nolint
func (msg MsgCollectWaitSign) Route() string { return RouterKey }
func (msg MsgCollectWaitSign) Type() string  { return "collect_wait_sign" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCollectWaitSign) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCollectWaitSign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCollectWaitSign) ValidateBasic() sdk.Error {
	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.CollectToCU == "" {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}

	if len(msg.RawData) == 0 {
		return ErrNilRawData(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}
	return nil
}

func (msg MsgCollectWaitSign) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgCollectSignFinish struct {
	OrderIDs  []string `json:"order_ids"`
	SignedTx  []byte   `json:"signed_tx"`
	Validator string   `json:"validator"`
}

func NewMsgCollectSignFinish(valAddr string, ids []string, signedTx []byte) MsgCollectSignFinish {
	msg := MsgCollectSignFinish{
		OrderIDs:  make([]string, len(ids)),
		SignedTx:  make([]byte, len(signedTx)),
		Validator: valAddr,
	}

	copy(msg.OrderIDs, ids)
	copy(msg.SignedTx, signedTx)
	return msg
}

//nolint
func (msg MsgCollectSignFinish) Route() string { return RouterKey }
func (msg MsgCollectSignFinish) Type() string  { return "collect_sign_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCollectSignFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCollectSignFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCollectSignFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}

	if len(msg.SignedTx) == 0 {
		return ErrNilSignedTx(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}
	return nil
}

func (msg MsgCollectSignFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgCollectFinish struct {
	OrderIDs  []string `json:"order_ids"`
	CostFee   sdk.Int  `json:"cost_fee"`
	Validator string   `json:"validator"`
}

func NewMsgCollectFinish(valAddr string, ids []string, fee sdk.Int) MsgCollectFinish {
	msg := MsgCollectFinish{
		OrderIDs:  make([]string, len(ids)),
		Validator: valAddr,
		CostFee:   fee,
	}
	copy(msg.OrderIDs, ids)
	return msg
}

//nolint
func (msg MsgCollectFinish) Route() string { return RouterKey }
func (msg MsgCollectFinish) Type() string  { return "collect_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCollectFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCollectFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCollectFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}

	if msg.CostFee.IsNegative() {
		return ErrBadCostFee(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}
	return nil
}

func (msg MsgCollectFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgWithdrawal struct {
	FromCU             string  `json:"from_cu"`
	ToMultisignAddress string  `json:"to_multi_sign_address"`
	Symbol             string  `json:"symbol"`
	Amount             sdk.Int `json:"amount"`
	GasFee             sdk.Int `json:"gas_fee"`
	OrderID            string  `json:"order_id"`
}

func NewMsgWithdrawal(fromCU, toAddr, symbol, orderID string, amount, gasFee sdk.Int) MsgWithdrawal {
	return MsgWithdrawal{
		FromCU:             fromCU,
		ToMultisignAddress: toAddr,
		Symbol:             symbol,
		Amount:             amount,
		GasFee:             gasFee,
		OrderID:            orderID,
	}
}

//nolint
func (msg MsgWithdrawal) Route() string { return RouterKey }
func (msg MsgWithdrawal) Type() string  { return "withdrawal" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgWithdrawal) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWithdrawal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawal) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.ToMultisignAddress == "" {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.OrderID == "" {
		return ErrNilOrderID(DefaultCodespace)
	}

	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidAmount("amount is not positive")
	}

	if !msg.GasFee.IsPositive() {
		return sdk.ErrInvalidAmount("gasFee is not positive")
	}
	return nil
}

//________________________________
type MsgWithdrawalConfirm struct {
	FromCU  string `json:"from_cu"`
	OrderID string `json:"order_id"`
	Valid   bool   `json:"valid"`
}

func NewMsgWithdrawalConfirm(fromCU, orderID string, valid bool) MsgWithdrawalConfirm {
	return MsgWithdrawalConfirm{
		FromCU:  fromCU,
		OrderID: orderID,
		Valid:   valid,
	}
}

//nolint
func (msg MsgWithdrawalConfirm) Route() string { return RouterKey }
func (msg MsgWithdrawalConfirm) Type() string  { return "withdrawal_confirm" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgWithdrawalConfirm) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWithdrawalConfirm) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawalConfirm) ValidateBasic() sdk.Error {
	_, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}
	if msg.OrderID == "" {
		return ErrNilOrderID(DefaultCodespace)
	}

	return nil
}

var _ sdk.SettleMsg = MsgWithdrawalConfirm{}

func (msg MsgWithdrawalConfirm) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgWithdrawalWaitSign struct {
	OpCU       string   `json:"opcu"`
	OrderIDs   []string `json:"order_ids"`
	SignHashes [][]byte `json:"sign_hashes"`
	RawData    []byte   `json:"raw_data"`
	Validator  string   `json:"validator"`
}

func NewMsgWithdrawalWaitSign(opCUAddr, valAddr string, ids []string, signHashes [][]byte, rawdata []byte) MsgWithdrawalWaitSign {
	msg := MsgWithdrawalWaitSign{
		OpCU:       opCUAddr,
		OrderIDs:   make([]string, len(ids)),
		SignHashes: signHashes,
		RawData:    make([]byte, len(rawdata)),
		Validator:  valAddr,
	}

	copy(msg.OrderIDs, ids)
	copy(msg.RawData, rawdata)
	return msg
}

//nolint
func (msg MsgWithdrawalWaitSign) Route() string { return RouterKey }
func (msg MsgWithdrawalWaitSign) Type() string  { return "withdrawal_wait_sign" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgWithdrawalWaitSign) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWithdrawalWaitSign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawalWaitSign) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.CUAddressFromBase58(msg.OpCU)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	_, err = sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}

	if len(msg.RawData) == 0 {
		return ErrNilRawData(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}

	return nil
}

func (msg MsgWithdrawalWaitSign) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgWithdrawalSignFinish struct {
	OrderIDs  []string `json:"order_ids"`
	SignedTx  []byte   `json:"signed_tx"`
	Validator string   `json:"validator"`
}

func NewMsgWithdrawalSignFinish(valAddr string, ids []string, signedTx []byte) MsgWithdrawalSignFinish {
	msg := MsgWithdrawalSignFinish{
		OrderIDs:  make([]string, len(ids)),
		SignedTx:  make([]byte, len(signedTx)),
		Validator: valAddr,
	}

	copy(msg.OrderIDs, ids)
	copy(msg.SignedTx, signedTx)
	return msg
}

//nolint
func (msg MsgWithdrawalSignFinish) Route() string { return RouterKey }
func (msg MsgWithdrawalSignFinish) Type() string  { return "withdrawal_sign_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgWithdrawalSignFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWithdrawalSignFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawalSignFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}

	if len(msg.SignedTx) == 0 {
		return ErrNilSignedTx(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}
	return nil
}

func (msg MsgWithdrawalSignFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgWithdrawalFinish struct {
	OrderIDs  []string `json:"order_ids"`
	CostFee   sdk.Int  `json:"cost_fee"`
	Validator string   `json:"validator"`
	Valid     bool     `json:"valid"`
}

func NewMsgWithdrawalFinish(valAddr string, ids []string, fee sdk.Int, valid bool) MsgWithdrawalFinish {
	msg := MsgWithdrawalFinish{
		OrderIDs:  make([]string, len(ids)),
		Validator: valAddr,
		CostFee:   fee,
		Valid:     valid,
	}
	copy(msg.OrderIDs, ids)
	return msg
}

//nolint
func (msg MsgWithdrawalFinish) Route() string { return RouterKey }
func (msg MsgWithdrawalFinish) Type() string  { return "withdrawal_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgWithdrawalFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWithdrawalFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawalFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}

	if msg.CostFee.IsNegative() {
		return ErrBadCostFee(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}
	return nil
}

func (msg MsgWithdrawalFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgSysTransfer struct {
	FromCU    sdk.CUAddress `json:"from_cu"`
	ToCU      sdk.CUAddress `json:"to_cu"`
	ToAddress string        `json:"to_address"`
	Symbol    string        `json:"symbol"`
	OrderID   string        `json:"order_id"`
	Validator string        `json:"validator"`
}

func NewMsgSysTransfer(fromCU, toCU sdk.CUAddress, toAddr, symbol, orderID, valAddr string) MsgSysTransfer {
	return MsgSysTransfer{
		FromCU:    fromCU,
		ToCU:      toCU,
		ToAddress: toAddr,
		Symbol:    symbol,
		OrderID:   orderID,
		Validator: valAddr,
	}
}

//nolint
func (msg MsgSysTransfer) Route() string { return RouterKey }
func (msg MsgSysTransfer) Type() string  { return "sys_transfer" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgSysTransfer) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgSysTransfer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSysTransfer) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.ToAddress == "" {
		return ErrBadAddress(DefaultCodespace)
	}

	if sdk.IsIllegalOrderID(msg.OrderID) {
		return sdk.ErrInvalidTx("Invalid order id")
	}
	if msg.FromCU == nil || !msg.FromCU.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU address: %s is invalid", msg.FromCU.String()))
	}
	if msg.ToCU == nil || !msg.ToCU.IsValidAddr() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("TO CU address: %s is invalid", msg.ToCU.String()))
	}
	return nil
}

func (msg MsgSysTransfer) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgSysTransferWaitSign struct {
	OrderID   string `json:"order_id"`
	SignHash  []byte `json:"sign_hash"`
	RawData   []byte `json:"raw_data"`
	Validator string `json:"validator"`
}

func NewMsgSysTransferWaitSign(valAddr string, orderid string, signHash []byte, rawdata []byte) MsgSysTransferWaitSign {
	msg := MsgSysTransferWaitSign{
		OrderID:   orderid,
		SignHash:  signHash,
		RawData:   make([]byte, len(rawdata)),
		Validator: valAddr,
	}

	copy(msg.RawData, rawdata)
	return msg
}

//nolint
func (msg MsgSysTransferWaitSign) Route() string { return RouterKey }
func (msg MsgSysTransferWaitSign) Type() string  { return "sys_transfer_wait_sign" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgSysTransferWaitSign) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgSysTransferWaitSign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSysTransferWaitSign) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if len(msg.RawData) == 0 {
		return ErrNilRawData(DefaultCodespace)
	}

	return nil
}

func (msg MsgSysTransferWaitSign) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgSysTransferSignFinish struct {
	OrderID   string `json:"order_id"`
	SignedTx  []byte `json:"signed_tx"`
	Validator string `json:"validator"`
}

func NewMsgSysTransferSignFinish(valAddr, orderid string, signedTx []byte) MsgSysTransferSignFinish {
	msg := MsgSysTransferSignFinish{
		OrderID:   orderid,
		SignedTx:  make([]byte, len(signedTx)),
		Validator: valAddr,
	}

	copy(msg.SignedTx, signedTx)
	return msg
}

//nolint
func (msg MsgSysTransferSignFinish) Route() string { return RouterKey }
func (msg MsgSysTransferSignFinish) Type() string  { return "sys_transfer_sign_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgSysTransferSignFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgSysTransferSignFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSysTransferSignFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if len(msg.SignedTx) == 0 {
		return ErrNilSignedTx(DefaultCodespace)
	}

	return nil
}

func (msg MsgSysTransferSignFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgSysTransferFinish struct {
	OrderID   string  `json:"order_id"`
	CostFee   sdk.Int `json:"cost_fee"`
	Validator string  `json:"validator"`
}

func NewMsgSysTransferFinish(valAddr, orderid string, fee sdk.Int) MsgSysTransferFinish {
	msg := MsgSysTransferFinish{
		OrderID:   orderid,
		Validator: valAddr,
		CostFee:   fee,
	}

	return msg
}

//nolint
func (msg MsgSysTransferFinish) Route() string { return RouterKey }
func (msg MsgSysTransferFinish) Type() string  { return "sys_transfer_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgSysTransferFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgSysTransferFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSysTransferFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.CostFee.IsNegative() {
		return ErrBadCostFee(DefaultCodespace)
	}

	return nil
}

func (msg MsgSysTransferFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgOpcuAssetTransfer struct {
	FromCU        string             `json:"from_cu"`
	OpCU          string             `json:"opcu"`
	ToAddr        string             `json:"to_addr"`
	Symbol        string             `json:"symbol"`
	TransferItems []sdk.TransferItem `json:"transfer_items,omitempty"`
	OrderID       string             `json:"order_id"`
}

func NewMsgOpcuAssetTransfer(fromCU, opCU, toAddr, symbol, orderID string, items []sdk.TransferItem) MsgOpcuAssetTransfer {
	msg := MsgOpcuAssetTransfer{
		FromCU:        fromCU,
		OpCU:          opCU,
		ToAddr:        toAddr,
		Symbol:        symbol,
		TransferItems: make([]sdk.TransferItem, len(items)),
		OrderID:       orderID,
	}

	copy(msg.TransferItems, items)
	return msg
}

//nolint
func (msg MsgOpcuAssetTransfer) Route() string { return RouterKey }
func (msg MsgOpcuAssetTransfer) Type() string  { return "opcuasset_transfer" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgOpcuAssetTransfer) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgOpcuAssetTransfer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgOpcuAssetTransfer) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	_, err = sdk.CUAddressFromBase58(msg.OpCU)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.ToAddr == "" {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.OrderID == "" {
		return ErrNilOrderID(DefaultCodespace)
	}

	if len(msg.TransferItems) == 0 {
		return sdk.ErrInvalidTx("transfer items are empty")
	}

	return nil
}

func (msg MsgOpcuAssetTransfer) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgOpcuAssetTransferWaitSign struct {
	OrderID    string   `json:"order_id"`
	SignHashes [][]byte `json:"sign_hashes"`
	RawData    []byte   `json:"raw_data"`
	Validator  string   `json:"validator"`
}

func NewMsgOpcuAssetTransferWaitSign(valAddr, id string, signHashes [][]byte, rawdata []byte) MsgOpcuAssetTransferWaitSign {
	msg := MsgOpcuAssetTransferWaitSign{
		OrderID:    id,
		SignHashes: signHashes,
		RawData:    make([]byte, len(rawdata)),
		Validator:  valAddr,
	}

	copy(msg.RawData, rawdata)
	return msg
}

//nolint
func (msg MsgOpcuAssetTransferWaitSign) Route() string { return RouterKey }
func (msg MsgOpcuAssetTransferWaitSign) Type() string  { return "opcuasset_transfer_wait_sign" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgOpcuAssetTransferWaitSign) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgOpcuAssetTransferWaitSign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgOpcuAssetTransferWaitSign) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if len(msg.RawData) == 0 {
		return ErrNilRawData(DefaultCodespace)
	}

	return nil
}

func (msg MsgOpcuAssetTransferWaitSign) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgOpcuAssetTransferSignFinish struct {
	OrderID   string `json:"order_id"`
	SignedTx  []byte `json:"signed_tx"`
	Validator string `json:"validator"`
}

func NewMsgOpcuAssetTransferSignFinish(valAddr, id string, signedTx []byte) MsgOpcuAssetTransferSignFinish {
	msg := MsgOpcuAssetTransferSignFinish{
		OrderID:   id,
		SignedTx:  make([]byte, len(signedTx)),
		Validator: valAddr,
	}

	copy(msg.SignedTx, signedTx)
	return msg
}

//nolint
func (msg MsgOpcuAssetTransferSignFinish) Route() string { return RouterKey }
func (msg MsgOpcuAssetTransferSignFinish) Type() string  { return "opcuasset_transfer_sign_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgOpcuAssetTransferSignFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgOpcuAssetTransferSignFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgOpcuAssetTransferSignFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if len(msg.SignedTx) == 0 {
		return ErrNilSignedTx(DefaultCodespace)
	}

	return nil
}

func (msg MsgOpcuAssetTransferSignFinish) IsSettleOnlyMsg() bool {
	return true
}

//________________________________
type MsgOpcuAssetTransferFinish struct {
	OrderID   string  `json:"order_id"`
	CostFee   sdk.Int `json:"cost_fee"`
	Validator string  `json:"validator"`
}

func NewMsgOpcuAssetTransferFinish(valAddr string, id string, fee sdk.Int) MsgOpcuAssetTransferFinish {
	msg := MsgOpcuAssetTransferFinish{
		OrderID:   id,
		Validator: valAddr,
		CostFee:   fee,
	}

	return msg
}

//nolint
func (msg MsgOpcuAssetTransferFinish) Route() string { return RouterKey }
func (msg MsgOpcuAssetTransferFinish) Type() string  { return "opcuasset_transfer_finish" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgOpcuAssetTransferFinish) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgOpcuAssetTransferFinish) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgOpcuAssetTransferFinish) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid

	_, err := sdk.CUAddressFromBase58(msg.Validator)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if msg.CostFee.IsNegative() {
		return ErrBadCostFee(DefaultCodespace)
	}

	return nil
}

func (msg MsgOpcuAssetTransferFinish) IsSettleOnlyMsg() bool {
	return true
}

type EvidenceValidator struct {
	EvidenceType int    `json:"evidence_type"`
	Validator    string `json:"validator"`
}

//________________________________
type MsgOrderRetry struct {
	OrderIDs   []string            `json:"order_ids"`
	RetryTimes uint32              `json:"retry_times"`
	Evidences  []EvidenceValidator `json:"evidences,omitempty"`
	From       string              `json:"from"`
}

func NewMsgOrderRetry(from string, ids []string, retrytimes uint32, evidences []EvidenceValidator) MsgOrderRetry {
	msg := MsgOrderRetry{
		OrderIDs:   make([]string, len(ids)),
		RetryTimes: retrytimes,
		Evidences:  make([]EvidenceValidator, len(evidences)),
		From:       from,
	}
	copy(msg.OrderIDs, ids)
	copy(msg.Evidences, evidences)
	return msg
}

//nolint
func (msg MsgOrderRetry) Route() string { return RouterKey }
func (msg MsgOrderRetry) Type() string  { return "order_retry" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgOrderRetry) GetSigners() []sdk.CUAddress {
	// delegator is first signer so delegator pays fees
	addr, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgOrderRetry) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgOrderRetry) ValidateBasic() sdk.Error {
	_, err := sdk.CUAddressFromBase58(msg.From)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}

	if len(msg.OrderIDs) == 0 {
		return ErrNilOrderID(DefaultCodespace)
	}
	if sdk.IsIllegalOrderIDList(msg.OrderIDs) {
		return ErrNilOrderID(DefaultCodespace)
	}
	// evidence 中 validator 不可重复
	validators := make(map[string]bool)
	for _, evidence := range msg.Evidences {
		if validators[evidence.Validator] {
			return sdk.NewError(DefaultCodespace, sdk.CodeInvalidTx, "duplicated validators")
		}
		validators[evidence.Validator] = true
	}
	return nil
}

func (msg MsgOrderRetry) IsSettleOnlyMsg() bool {
	return true
}

type MsgCancelWithdrawal struct {
	FromCU  string `json:"from_cu"`
	OrderID string `json:"order_id"`
}

func NewMsgCancelWithdrawal(fromCU string, orderID string) MsgCancelWithdrawal {
	msg := MsgCancelWithdrawal{
		FromCU:  fromCU,
		OrderID: orderID,
	}

	return msg
}

//nolint
func (msg MsgCancelWithdrawal) Route() string { return RouterKey }
func (msg MsgCancelWithdrawal) Type() string  { return "cancel_withdrawal" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCancelWithdrawal) GetSigners() []sdk.CUAddress {
	addr, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return []sdk.CUAddress{}
	}
	return []sdk.CUAddress{addr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCancelWithdrawal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCancelWithdrawal) ValidateBasic() sdk.Error {
	_, err := sdk.CUAddressFromBase58(msg.FromCU)
	if err != nil {
		return ErrBadAddress(DefaultCodespace)
	}
	if sdk.IsIllegalOrderID(msg.OrderID) {
		return ErrNilOrderID(DefaultCodespace)
	}

	return nil
}
