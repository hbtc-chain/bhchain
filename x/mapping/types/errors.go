package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidSwapAmount       sdk.CodeType = 2100
	CodeInvalidInitialIssuePool sdk.CodeType = 2101
	CodeInvalidIssuePool        sdk.CodeType = 2102
	CodeMappingNotFound         sdk.CodeType = 2103
	CodeUnmatchedDecimals       sdk.CodeType = 2104
	CodeMappingDisabled         sdk.CodeType = 2105
	CodeDuplicatedIssueSymbol   sdk.CodeType = 2106
	CodeTargetSymbolUsedAsIssue sdk.CodeType = 2107
	CodeIssueSymbolUsedAsTarget sdk.CodeType = 2108
	CodeUnmatchedTotalSupply    sdk.CodeType = 2109
)

func ErrInvalidSwapAmount(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidSwapAmount, format)
}

func ErrInvalidInitialIssuePool(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInitialIssuePool, format)
}

func ErrInvalidIssuePool(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidIssuePool, format)
}

func ErrMappingNotFound(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeMappingNotFound, format)
}

func ErrUnmatchedDecimals(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeUnmatchedDecimals, format)
}

func ErrMappingDisabled(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeMappingDisabled, format)
}

func ErrDuplicatedIssueSymbol(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeDuplicatedIssueSymbol, format)
}

func ErrTargetSymbolUsedAsIssue(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeTargetSymbolUsedAsIssue, format)
}

func ErrIssueSymbolUsedAsTarget(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeIssueSymbolUsedAsTarget, format)
}

func ErrUnmatchedTotalSupply(codespace sdk.CodespaceType, format string) sdk.Error {
	return sdk.NewError(codespace, CodeUnmatchedTotalSupply, format)
}
