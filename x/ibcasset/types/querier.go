package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	QueryPendingDepositList = "PendingDepositList"
	QueryCUAsset            = "CUAsset"
	QueryOpCUAstInfo        = "OpCUAstInfo"
	QueryDepositExistance   = "DepositExistance"
)

// QueryCUParams defines the params for querying cu.
type QueryCUAssetParams struct {
	Address sdk.CUAddress
}

// NewQueryCUParams creates a new instance of QueryCUParams.
func NewQueryCUAssetParams(addr sdk.CUAddress) QueryCUAssetParams {
	return QueryCUAssetParams{Address: addr}
}

type QueryCUChainAddressParams struct {
	Chain   string
	Address string
}

// QueryCUParams defines the params for querying cu.
type QueryOpCUAstInfoParams struct {
	Symbol string
}

type DepositExistanceParams struct {
	Symbol  string
	Address sdk.CUAddress
	TxHash  string
	Index   uint64
}

func NewDepositExistanceParams(symbol string, addr sdk.CUAddress, txHash string, index uint64) *DepositExistanceParams {
	return &DepositExistanceParams{
		Symbol:  symbol,
		Address: addr,
		TxHash:  txHash,
		Index:   index,
	}
}

type QueryPendingDepositListParams struct {
	Address sdk.CUAddress
}

func NewQueryPendingDepositListParams(addr sdk.CUAddress) *QueryPendingDepositListParams {
	return &QueryPendingDepositListParams{
		Address: addr,
	}
}
