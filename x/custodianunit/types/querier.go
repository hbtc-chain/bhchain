package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

// query endpoints supported by the auth Querier
const (
	QueryCU                 = "CU"
	QueryCUWithChainAddr    = "CUChainAddr"
	QueryMultiChainAddrInfo = "MultiChainAddrInfo"
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
