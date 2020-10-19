package types

import (
	"fmt"
	sdk "github.com/hbtc-chain/bhchain/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace  sdk.CodespaceType = "token"
	CodeInvalidInput  CodeType          = 103
	CodeEmptyData     CodeType          = 104
	CodeDuplicatedKey CodeType          = 105
)

func ErrInvalidProposalTokenInfo(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "invalid tokenInfo for add token proposal")
}

// ErrEmptyKey returns an error for when an empty key is given.
func ErrEmptyKey(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter key is empty")
}

// ErrEmptyValue returns an error for when an empty key is given.
func ErrEmptyValue(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter value is empty")
}

func ErrInvalidParameter(codespace sdk.CodespaceType, key, value string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, fmt.Sprintf("key:%v, value:%v", key, value))
}

func ErrDuplicatedKey(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeDuplicatedKey, "parameter key is duplicated")
}
