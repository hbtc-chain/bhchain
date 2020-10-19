package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = "mint"
	CodeInvalidInput CodeType          = 103
)

func ErrInvalidProposalPrice(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "invalid inflation parameters update proposal price")
}
func ErrInvalidProposalInflation(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "invalid inflation parameters update proposal inflation")
}
func ErrInvalidProposalNodeCostPerMonth(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "invalid inflation parameters update proposal node cost per month")
}
