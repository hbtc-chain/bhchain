package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

// query endpoints supported by the auth Querier
const (
	QueryCU                 = "CU"
	QueryPendingDepositList = "PendingDepositList"
	QueryOpCU               = "OpCU"
	QueryCUWithChainAddr    = "CUChainAddr"
	QueryMultiChainAddrInfo = "MultiChainAddrInfo"
	QueryDepositExistance   = "DepositExistance"
	QueryMinimumGasPrice    = "QueryMinimumGasPrice"
)

// QueryCUParams defines the params for querying cu.
type QueryCUParams struct {
	Address sdk.CUAddress
}

// NewQueryCUParams creates a new instance of QueryCUParams.
func NewQueryCUParams(addr sdk.CUAddress) QueryCUParams {
	return QueryCUParams{Address: addr}
}

// QueryCUParams defines the params for querying cu.
type QueryOpCUParams struct {
	Symbol string
}

// NewQueryCUParams creates a new instance of QueryCUParams.
func NewQueryOpCUParams(symbol string) QueryOpCUParams {
	return QueryOpCUParams{Symbol: symbol}
}

type QueryCUChainAddressParams struct {
	Chain   string
	Address string
}

func NewQueryCUChainAddressParams(chain string, address string) QueryCUChainAddressParams {
	return QueryCUChainAddressParams{Chain: chain, Address: address}
}

type MultiQueryChainAddrInfoParams struct {
	ChainInfos []QueryCUChainAddressParams
}

func NewMultiQueryChanAddrInfoParams(chaininfos []QueryCUChainAddressParams) MultiQueryChainAddrInfoParams {
	return MultiQueryChainAddrInfoParams{ChainInfos: chaininfos}
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
