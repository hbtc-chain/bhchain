package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidInput         CodeType     = 101
	CodeSendDisabled         sdk.CodeType = 102
	CodeInvalidInputsOutputs sdk.CodeType = 13

	CodeInvalidAddress CodeType = sdk.CodeInvalidAddress
)

// ErrNoInputs is an error
func ErrNoInputs(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInputsOutputs, "no inputs to send transaction")
}

// ErrNoOutputs is an error
func ErrNoOutputs(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInputsOutputs, "no outputs to send transaction")
}

// ErrInputOutputMismatch is an error
func ErrInputOutputMismatch(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInputsOutputs, "sum inputs != sum outputs")
}

// ErrSendDisabled is an error
func ErrSendDisabled(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSendDisabled, "send transactions are currently disabled")
}

func ErrBadAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAddress, "address is invalid")
}

func ErrNilOrderID(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "orderid is nil")
}

func ErrNilRawData(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "raw data is nil")
}

func ErrNilSignedTx(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "sign tx is nil")
}

func ErrNilSignHash(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "sign hash is nil")
}

func ErrBadCostFee(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "sign tx is nil")
}

func ErrInvalidParam(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "invalid tx input param")
}
