package mapping

import (
	"github.com/hbtc-chain/bhchain/x/mapping/types"
	"github.com/hbtc-chain/bhchain/x/receipt"
)

type (
	MsgMappingSwap      = types.MsgMappingSwap
	MsgCreateDirectSwap = types.MsgCreateDirectSwap
	MsgCreateFreeSwap   = types.MsgCreateFreeSwap
	MsgSwapSymbol       = types.MsgSwapSymbol
	MsgCancelSwap       = types.MsgCancelSwap

	MappingInfo        = types.MappingInfo
	MappingBalanceFlow = receipt.MappingBalanceFlow

	FreeSwapInfo    = types.FreeSwapInfo
	DirectSwapInfo  = types.DirectSwapInfo
	SwapPool        = types.SwapPool
	FreeSwapOrder   = types.FreeSwapOrder
	DirectSwapOrder = types.DirectSwapOrder

	QueryMappingParams     = types.QueryMappingParams
	QueryMappingListParams = types.QueryMappingListParams
	QueryResMappingInfo    = types.QueryResMappingInfo
	QueryResMappingList    = types.QueryResMappingList

	QueryFreeSwapOrderParams       = types.QueryFreeSwapOrderParams
	QueryDirectSwapOrderParams     = types.QueryDirectSwapOrderParams
	QueryFreeSwapOrderListParams   = types.QueryFreeSwapOrderListParams
	QueryDirectSwapOrderListParams = types.QueryDirectSwapOrderListParams

	QueryFreeSwapOrderList   = types.QueryFreeSwapOrderList
	QueryDirectSwapOrderList = types.QueryDirectSwapOrderList
)

const (
	DefaultCodespace = types.DefaultCodespace

	ModuleName   = types.ModuleName
	RouterKey    = types.RouterKey
	StoreKey     = types.StoreKey
	QuerierRoute = types.QuerierRoute
	// DefaultParamsapce = types.DefaultParamspace

	QueryInfo           = types.QueryInfo
	QueryList           = types.QueryList
	QueryFreeSwapInfo   = types.QueryFreeSwapInfo
	QueryDirectSwapInfo = types.QueryDirectSwapInfo
	QueryFreeSwapList   = types.QueryFreeSwapList
	QueryDirectSwapList = types.QueryDirectSwapList
)

var (
	ModuleCdc         = types.ModuleCdc
	RegisterCodec     = types.RegisterCodec
	NewMsgMappingSwap = types.NewMsgMappingSwap
)
