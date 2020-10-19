package types

import (
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// MappingInfo struct for mapping
type MappingInfo struct {
	IssueSymbol  sdk.Symbol `json:"issue_symbol"`
	TargetSymbol sdk.Symbol `json:"target_symbol"`
	TotalSupply  sdk.Int    `json:"total_supply"`
	IssuePool    sdk.Int    `json:"issue_pool"`
	Enabled      bool       `json:"enabled"`
}

func (m MappingInfo) String() string {
	return fmt.Sprintf(`
	IssuerSymbol:%v
	TargetSymbol:%v
	TotalSupply:%v
	IssuePool:%v
	Enabled:%v
	`, m.IssueSymbol, m.TargetSymbol, m.TotalSupply, m.IssuePool, m.Enabled)
}

type QueryMappingParams struct {
	IssueSymbol string `json:"issue_symbol"`
}

type QueryMappingListParams struct {
	Page, Limit int
}

func NewQueryMappingListParams(page, limit int) QueryMappingListParams {
	return QueryMappingListParams{page, limit}
}

type QueryResMappingInfo struct {
	IssueSymbol  sdk.Symbol `json:"issue_symbol"`
	TargetSymbol sdk.Symbol `json:"target_symbol"`
	TotalSupply  sdk.Int    `json:"total_supply"`
	IssuePool    sdk.Int    `json:"issue_pool"`
	Enabled      bool       `json:"enabled"`
}

func (m QueryResMappingInfo) String() string {
	return fmt.Sprintf(`
	IssuerSymbol:%v
	TargetSymbol:%v
	TotalSupply:%v
	IssuePool:%v
	Enabled:%v
	`, m.IssueSymbol, m.TargetSymbol, m.TotalSupply, m.IssuePool, m.Enabled)
}

func MappingInfoToQueryRes(mi *MappingInfo) QueryResMappingInfo {
	return QueryResMappingInfo{
		IssueSymbol:  mi.IssueSymbol,
		TargetSymbol: mi.TargetSymbol,
		TotalSupply:  mi.TotalSupply,
		IssuePool:    mi.IssuePool,
		Enabled:      mi.Enabled,
	}
}

type QueryResMappingList []QueryResMappingInfo

func (m QueryResMappingList) String() (out string) {
	for _, item := range m {
		out += item.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// query endpoints supported by the mapping module
const (
	QueryInfo           = "info"
	QueryList           = "list"
	QueryFreeSwapInfo   = "freeswapinfo"
	QueryDirectSwapInfo = "directswapinfo"
	QueryFreeSwapList   = "freeswaplist"
	QueryDirectSwapList = "directswaplist"
)

type QueryFreeSwapOrderParams struct {
	OrderID string `json:"order_id"`
}

type QueryFreeSwapOrderListParams struct {
	Page, Limit int
}

func NewQueryFreeSwapInfoListParams(page, limit int) QueryFreeSwapOrderListParams {
	return QueryFreeSwapOrderListParams{page, limit}
}

type QueryDirectSwapOrderParams struct {
	OrderID string `json:"order_id"`
}

type QueryDirectSwapOrderListParams struct {
	Page, Limit int
}

func NewQueryDirectSwapInfoListParams(page, limit int) QueryDirectSwapOrderListParams {
	return QueryDirectSwapOrderListParams{page, limit}
}

type FreeSwapInfo struct {
	SrcSymbol     sdk.Symbol `json:"src_symbol"`
	TargetSymbol  sdk.Symbol `json:"target_symbol"`
	TotalAmount   sdk.Int    `json:"total_amount"`
	MaxSwapAmount sdk.Int    `json:"max_swap_amount"`
	MinSwapAmount sdk.Int    `json:"min_swap_amount"`
	SwapPrice     sdk.Int    `json:"swap_price"`
	ExpiredTime   int64      `json:"expired_time"`
	Desc          string     `json:"desc"`
}

type DirectSwapInfo struct {
	SrcSymbol    sdk.Symbol `json:"src_symbol"`
	TargetSymbol sdk.Symbol `json:"target_symbol"`
	Amount       sdk.Int    `json:"amount"`
	SwapAmount   sdk.Int    `json:"swap_amount"`
	ExpiredTime  int64      `json:"expired_time"`
	ReceiveAddr  string     `json:"receieve_addr"`
	Desc         string     `json:"desc"`
}

func (s FreeSwapInfo) String() string {
	return fmt.Sprintf(`
	SrcSymbol:%v
	TargetSymbol:%v
	TotalAmount:%v
	MaxSwapAmount:%v
	MinSwapAmount:%v
    SwapPrice:%v,
    ExpiredTime:%v,
    Decs:%v,
	`, s.SrcSymbol, s.TargetSymbol, s.TotalAmount, s.MaxSwapAmount, s.MinSwapAmount, s.SwapPrice, s.ExpiredTime, s.Desc)
}

func (s DirectSwapInfo) String() string {
	return fmt.Sprintf(`
	SrcSymbol:%v
	TargetSymbol:%v
	Amount:%v
	SwapAmount:%v
    ExpiredTime:%v, 
    ReciveAddr:%v,
    Decs:%v,
	`, s.SrcSymbol, s.TargetSymbol, s.Amount, s.SwapAmount, s.ExpiredTime, s.ReceiveAddr, s.Desc)
}

type SwapPool struct {
	SwapCoins sdk.Coins `json:"swap_coins"`
}

const (
	SwapTypeFree   = 0x0
	SwapTypeDirect = 0x1
)

type FreeSwapOrder struct {
	OrderId      string        `json:"order_id"`
	Owner        sdk.CUAddress `json:"owner"`
	SwapInfo     FreeSwapInfo  `json:"swap_info"`
	RemainAmount sdk.Int       `json:"remain_amount"`
}

type DirectSwapOrder struct {
	OrderId  string         `json:"order_id"`
	Owner    sdk.CUAddress  `json:"owner"`
	SwapInfo DirectSwapInfo `json:"swap_info"`
}

func (o FreeSwapOrder) String() string {
	return fmt.Sprintf(`
	OrderID:%v
	Owner:%v
	SwapInfo:%v
	RemainAmount:%v
	`, o.OrderId, o.Owner, o.SwapInfo.String(), o.RemainAmount)
}

func (o DirectSwapOrder) String() string {
	return fmt.Sprintf(`
	OrderID:%v
	Owner:%v
	SwapInfo:%v
	`, o.OrderId, o.Owner, o.SwapInfo.String())
}

type QueryFreeSwapOrderList []FreeSwapOrder

func (o QueryFreeSwapOrderList) String() (out string) {
	for _, item := range o {
		out += item.String() + "\n"
	}
	return strings.TrimSpace(out)
}

type QueryDirectSwapOrderList []DirectSwapOrder

func (o QueryDirectSwapOrderList) String() (out string) {
	for _, item := range o {
		out += item.String() + "\n"
	}
	return strings.TrimSpace(out)
}
