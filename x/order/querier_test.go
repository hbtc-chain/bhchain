package order

//
//import (
//	"fmt"
//	sdk "github.com/hbtc-chain/bhchain/types"
//	"github.com/cosmos/cosmos-sdk/codec"
//	"github.com/stretchr/testify/assert"
//	abci "github.com/tendermint/tendermint/abci/types"
//	"testing"
//)
//
//func TestQueryOrder(t *testing.T) {
//	input := setupTestInput()
//
//	errQueryParamData := []byte("error ParamData bytes")
//	req := abci.RequestQuery{
//		Path: fmt.Sprintf("cu/%s/%s", QuerierRoute, QueryOrder),
//		Data: errQueryParamData,
//	}
//
//	res, err := queryOrder(input.ctx, req, input.ook)
//	assert.NotNil(t, err)
//	assert.Nil(t, res)
//
//	for _, av := range orderValues {
//		input.ook.SetOrder(input.ctx, av.in)
//	}
//
//	for _, tc := range cases {
//		queryParamData, err := codec.MarshalJSONIndent(input.ook.cdc, tc.queryParam)
//		req.Data = queryParamData
//		res, err = queryOrder(input.ctx, req, input.ook)
//		if tc.expError {
//			assert.NotNil(t, err, "case:%d", tc.caseNo)
//		} else {
//
//			assert.Nil(t, err, fmt.Sprintf("case:%d", tc.caseNo))
//			var ord Order
//			//var ord OrderKeyGen
//			if tc.resLength > 0 {
//				assert.NotNil(t, res)
//				err2 := input.cdc.UnmarshalJSON(res, &ord)
//				assert.Nil(t, err2)
//				assert.Equal(t, ord.GetCUAddress(), tc.queryParam.CuAddress, "case:%d", tc.caseNo)
//			} else {
//				err2 := input.cdc.UnmarshalJSON(res, &ord)
//				assert.Nil(t, err2)
//				assert.Nil(t, ord, "case:%d", tc.caseNo)
//
//			}
//		}
//	}
//}
//
//func TestQueryCUOrders(t *testing.T) {
//	input := setupTestInput()
//
//	errQueryParamData := []byte("error ParamData bytes")
//	req := abci.RequestQuery{
//		Path: fmt.Sprintf("cu/%s/%s", QuerierRoute, QueryCUOrders),
//		Data: errQueryParamData,
//	}
//
//	res, err := queryOrder(input.ctx, req, input.ook)
//	assert.NotNil(t, err)
//	assert.Nil(t, res)
//
//	for _, av := range orderValues {
//		input.ook.SetOrder(input.ctx, av.in)
//	}
//
//	for _, tc := range queryCUOrderscases {
//		queryParamData, err := codec.MarshalJSONIndent(input.ook.cdc, tc.queryParam)
//		req.Data = queryParamData
//		res, err = queryCUOrders(input.ctx, req, input.ook)
//		if tc.expError {
//			assert.NotNil(t, err, "case:%d", tc.caseNo)
//		} else {
//			assert.Nil(t, err)
//			var ords []Order
//			if tc.resLength > 0 {
//				assert.NotNil(t, res)
//				err2 := input.cdc.UnmarshalJSON(res, &ords)
//				assert.Nil(t, err2)
//				assert.Equal(t, ords[0].GetCUAddress(), tc.queryParam.CuAddress, "case:%d", tc.caseNo)
//				assert.Equal(t, tc.resLength, len(ords), "case:%d", tc.caseNo)
//			} else {
//				err2 := input.cdc.UnmarshalJSON(res, &ords)
//				assert.Nil(t, err2)
//				assert.Equal(t, tc.resLength, len(ords), "case:%d", tc.caseNo)
//			}
//		}
//	}
//}
//
//var cases = []struct {
//	caseNo     int
//	memo       string
//	queryParam QueryOrderParams
//	resLength  int
//	expError   bool
//}{
//	{0, "wrong cuAddress", QueryOrderParams{CuAddress: sdk.CUAddressFromBase58("NOTEXIST"), OrderID: 1}, 0, true},
//	{1, "wrong OrderID", QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"), OrderID: 0}, 0, true},
//	{2, "ok", QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"), OrderID: 2}, 1, false},
//	{3, "wrong OrderID", QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"), OrderID: 1}, 0, true},
//	{4, "ok", QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"), OrderID: 3}, 1, false},
//	{5, "wrong OrderID", QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"), OrderID: 3}, 0, true},
//	{6, "wrong cuAddress", QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE"), OrderID: 1}, 0, true},
//}
//
//var queryCUOrderscases = []struct {
//	caseNo     int
//	queryParam QueryOrderParams
//	resLength  int
//	expError   bool
//}{
//	{0, QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("NOTEXIST"), OrderID: 1}, 0, true},
//	{1, QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"), OrderID: 0}, 3, false},
//	{3, QueryOrderParams{CuAddress: sdk.StringToCustodianUnitAddress("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"), OrderID: 1}, 1, false},
//}
//
//var orderValues = []struct {
//	caseNo int
//	in     Order
//	exp    bool
//}{
//	{
//		caseNo: 1,
//		in: &OrderKeyGen{OrderBase: OrderBase{
//			ID:        1,
//			OrderType: OrderTypeKeyGen,
//			Symbol:    "btc",
//			CUAddress: sdk.StringToCustodianUnitAddress("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"),
//		},
//			KeyNodes:         []CustodianUnitAddress{sdk.StringToCustodianUnitAddress("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"), sdk.StringToCustodianUnitAddress("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE")},
//			SignThreshold:    2,
//			To:               sdk.StringToCustodianUnitAddress("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"),
//			MultiSignAddress: "",
//		},
//		exp: false,
//	},
//	{
//		caseNo: 2,
//		in: &OrderKeyGen{OrderBase: OrderBase{
//			ID:        2,
//			OrderType: OrderTypeKeyGen,
//			Symbol:    "eth",
//			CUAddress: sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"),
//		},
//			KeyNodes:         []CustodianUnitAddress{sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"), sdk.StringToCustodianUnitAddress("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE")},
//			SignThreshold:    2,
//			To:               sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"),
//			MultiSignAddress: "",
//		},
//		exp: false,
//	},
//	{
//		caseNo: 3,
//		in: &OrderKeyGen{OrderBase: OrderBase{
//			ID:        3,
//			OrderType: OrderTypeKeyGen,
//			Symbol:    "btc",
//			CUAddress: sdk.StringToCustodianUnitAddress("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"),
//		},
//			KeyNodes:         []CustodianUnitAddress{sdk.StringToCustodianUnitAddress("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"), sdk.StringToCustodianUnitAddress("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")},
//			SignThreshold:    2,
//			To:               sdk.StringToCustodianUnitAddress("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"),
//			MultiSignAddress: "",
//		},
//		exp: false,
//	},
//	{ //
//		caseNo: 4,
//		in: &OrderKeyGen{OrderBase: OrderBase{
//			ID:        0,
//			OrderType: OrderTypeKeyGen,
//			Symbol:    "eth",
//			CUAddress: sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"),
//		},
//			KeyNodes:         []CustodianUnitAddress{sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"), sdk.StringToCustodianUnitAddress("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE")},
//			SignThreshold:    2,
//			To:               sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"),
//			MultiSignAddress: "",
//		},
//		exp: false,
//	},
//	{
//		caseNo: 5,
//		in: &OrderTokenKeyGen{OrderBase: OrderBase{
//			ID:        5,
//			OrderType: OrderTypeTokenKeyGen,
//			Symbol:    "btc",
//			CUAddress: sdk.StringToCustodianUnitAddress("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"),
//		},
//			KeyNodes:         []CustodianUnitAddress{sdk.StringToCustodianUnitAddress("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"), sdk.StringToCustodianUnitAddress("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE")},
//			SignThreshold:    2,
//			MultiSignAddress: "",
//		},
//		exp: false,
//	},
//}
