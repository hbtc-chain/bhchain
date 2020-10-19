package receipt

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

type (
	Symbol             = sdk.Symbol
	CUAddress          = sdk.CUAddress
	Result             = sdk.Result
	Int                = sdk.Int
	CategoryType       = sdk.CategoryType
	Flow               = sdk.Flow
	OrderFlow          = sdk.OrderFlow
	BalanceFlow        = sdk.BalanceFlow
	DepositFlow        = sdk.DepositFlow
	OrderRetryFlow     = sdk.OrderRetryFlow
	MappingBalanceFlow = sdk.MappingBalanceFlow
)

const (
	CategoryTypeTransfer   = sdk.CategoryTypeTransfer
	CategoryTypeKeyGen     = sdk.CategoryTypeKeyGen
	CategoryTypeWithdrawal = sdk.CategoryTypeWithdrawal
	CategoryTypeDeposit    = sdk.CategoryTypeDeposit
)
