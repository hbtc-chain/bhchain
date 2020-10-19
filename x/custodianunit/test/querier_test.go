package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hbtc-chain/bhchain/x/custodianunit/types"

	sdk "github.com/hbtc-chain/bhchain/types"

	. "github.com/hbtc-chain/bhchain/x/custodianunit"
)

func Test_queryCU(t *testing.T) {
	input := setupTestInput()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", QuerierRoute, QueryCU),
		Data: []byte{},
	}

	res, err := QueryCUForTest(input.ctx, req, input.ak)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	req.Data = input.cdc.MustMarshalJSON(types.NewQueryCUParams([]byte("")))
	res, err = QueryCUForTest(input.ctx, req, input.ak)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryCUParams(addr))
	res, err = QueryCUForTest(input.ctx, req, input.ak)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	input.ak.SetCU(input.ctx, input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr))
	res, err = QueryCUForTest(input.ctx, req, input.ak)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	var CU CU
	err2 := input.cdc.UnmarshalJSON(res, &CU)
	assert.Nil(t, err2)
}

func Test_queryOpCU(t *testing.T) {
	input := setupTestInput()

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", QuerierRoute, types.QueryOpCU),
		Data: []byte{},
	}

	// have no op CU
	res, err := QueryOpCUForTest(input.ctx, req, input.ak)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	_, _, addr1 := types.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryOpCUParams(ethToken))

	// add  op CU1 of eth
	cu1 := input.ak.NewOpCUWithAddress(input.ctx, ethToken, addr1)
	input.ak.SetCU(input.ctx, cu1)
	res, err = QueryOpCUForTest(input.ctx, req, input.ak)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	var CUs []sdk.OpCUInfo
	err2 := input.cdc.UnmarshalJSON(res, &CUs)
	assert.Nil(t, err2)
	assert.Equal(t, 1, len(CUs))

	// disable CU1 to send eth tx
	cu1.SetEnableSendTx(false, ethToken, addr1.String())
	input.ak.SetCU(input.ctx, cu1)
	res, err = QueryOpCUForTest(input.ctx, req, input.ak)
	err2 = input.cdc.UnmarshalJSON(res, &CUs)
	assert.Nil(t, err2)
	assert.Equal(t, 1, len(CUs))
	// account based length of DepositList always 0
	assert.Equal(t, 0, len(CUs[0].DepositList))

	// add OP CU2 of eth
	_, _, addr2 := types.KeyTestPubAddr()
	cu2 := input.ak.NewOpCUWithAddress(input.ctx, ethToken, addr2)
	input.ak.SetCU(input.ctx, cu2)
	// query OP CU of eth,should got 2
	res, err = QueryOpCUForTest(input.ctx, req, input.ak)
	err2 = input.cdc.UnmarshalJSON(res, &CUs)
	assert.Nil(t, err2)
	assert.Equal(t, 2, len(CUs))

	// add OP CU3 of btc
	_, _, addr3 := types.KeyTestPubAddr()
	cu3 := input.ak.NewOpCUWithAddress(input.ctx, btcToken, addr3)
	input.ak.SetCU(input.ctx, cu3)
	// ad CU3 deposit List
	d1, _ := sdk.NewDepositItem("hash1", 0, sdk.NewInt(10), "", "memo", 0)
	d2, _ := sdk.NewDepositItem("hash2", 0, sdk.NewInt(10), "", "memo", 0)
	dls, _ := sdk.NewDepositList(d1, d2)
	input.ak.SetDepositList(input.ctx, btcToken, cu3.GetAddress(), dls)
	// query CU3 & check depositList
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryOpCUParams(btcToken))
	res, err = QueryOpCUForTest(input.ctx, req, input.ak)
	err2 = input.cdc.UnmarshalJSON(res, &CUs)
	assert.Nil(t, err2)
	assert.Equal(t, 1, len(CUs))
	assert.Equal(t, 2, len(CUs[0].DepositList))
	assert.EqualValues(t, d1.Amount.Add(d2.Amount), (*sdk.DepositList)(&CUs[0].DepositList).Sum())

	// add OP CU4 of erc20 tusdt
	_, _, addr4 := types.KeyTestPubAddr()
	cu4 := input.ak.NewOpCUWithAddress(input.ctx, usdtToken, addr4)
	// cu4 have 0.1 eth
	cu4.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(ethToken, sdk.NewIntWithDecimal(1, 17))))
	//cu4 have 100 tusdt
	cu4.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(usdtToken, sdk.NewIntWithDecimal(1, 20))))
	input.ak.SetCU(input.ctx, cu4)
	input.ak.SetDepositList(input.ctx, usdtToken, addr4, dls)
	// query CU4 & check depositList
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryOpCUParams(usdtToken))
	res, err = QueryOpCUForTest(input.ctx, req, input.ak)
	err2 = input.cdc.UnmarshalJSON(res, &CUs)
	assert.Nil(t, err2)
	assert.Equal(t, 1, len(CUs))
	// if not utxobased, have no DepositList
	assert.Equal(t, 0, len(CUs[0].DepositList))
	//? DepositList.Sum() == cu.GetAssetCoins().AmountOf(bhusdt)
	//assert.EqualValues(t, cu4.GetAssetCoins().AmountOf(usdtToken), (*sdk.DepositList)(&CUs[0].DepositList).Sum())
	// check mainnet token amount
	assert.EqualValues(t, cu4.GetAssetCoins().AmountOf(ethToken), CUs[0].MainNetAmount)

	// symbol ="" return opcu of all symbol
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryOpCUParams(""))
	res, err = QueryOpCUForTest(input.ctx, req, input.ak)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	err2 = input.cdc.UnmarshalJSON(res, &CUs)
	assert.Nil(t, err2)
	assert.Equal(t, 4, len(CUs))
}
