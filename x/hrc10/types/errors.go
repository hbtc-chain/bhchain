package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace   sdk.CodespaceType = "hrc10"
	CodeSymbolReserved CodeType          = 103
)

func ErrSymbolReserved(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeSymbolReserved, msg)
}
