package hrc10

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc10/types"
)

func TestHandleMsgNewTokenSuccess(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	tk := input.tk
	ck := input.ck
	dk := input.dk
	hk := input.hrc10k
	transferKeeper := input.transferKeeper
	//supplyKeeper := input.supplyKeeper

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	assert.Nil(t, err)
	fromCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, fromCUAddr)
	transferKeeper.AddCoins(ctx, fromCU.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, types.DefaultIssueTokenFee.MulRaw(5))))

	ck.SetCU(ctx, fromCU)
	fromCU = ck.GetCU(ctx, fromCUAddr)
	balances := transferKeeper.GetAllBalance(ctx, fromCU.GetAddress())
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(5), balances.AmountOf(sdk.NativeToken))

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	assert.Nil(t, err)
	toCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, toCUAddr)
	ck.SetCU(ctx, toCU)
	balances = transferKeeper.GetAllBalance(ctx, toCU.GetAddress())
	assert.Equal(t, sdk.Coins(nil), balances)

	openFee := sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, types.DefaultIssueTokenFee))
	totalsupply, ok := sdk.NewIntFromString("10000000000000000000000")
	assert.True(t, ok)

	msg := types.NewMsgNewToken(fromCUAddr, toCUAddr, "bhd", 18, totalsupply)

	res := handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeOK, res.Code)
	assert.Equal(t, 1, len(res.Events))
	assert.Equal(t, types.EventTypeNewToken, res.Events[0].Type)

	symbol := getSymbolFromNewTokenRes(t, res)
	//check tokenInfo
	ti := tk.GetToken(ctx, sdk.Symbol(symbol))
	assert.Equal(t, symbol, ti.GetSymbol().String())
	assert.Equal(t, sdk.NativeToken, ti.GetChain().String())
	assert.Equal(t, uint64(18), ti.GetDecimals())
	assert.Equal(t, totalsupply, ti.GetTotalSupply())
	assert.Equal(t, fromCUAddr.String(), ti.GetIssuer())
	assert.False(t, ti.IsIBCToken())

	//check coins
	toCU = ck.GetCU(ctx, toCUAddr)
	balances = transferKeeper.GetAllBalance(ctx, toCU.GetAddress())
	assert.Equal(t, totalsupply.String(), balances.AmountOf(symbol).String())
	fromCU = ck.GetCU(ctx, fromCUAddr)
	balances = transferKeeper.GetAllBalance(ctx, fromCU.GetAddress())
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(4), balances.AmountOf(sdk.NativeToken))

	//check supply

	//supply := supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(symbol)

	//assert.Equal(t, supply.String(), ti.GetTotalSupply().String())

	//check feepool
	feePool := dk.GetFeePool(ctx)
	assert.Equal(t, openFee.AmountOf(sdk.NativeToken), feePool.CommunityPool.AmountOf(sdk.NativeToken).TruncateInt())

	//sendcoins back to fromCUAddr
	sendAmt := sdk.NewInt(2000000000)
	res, _, err = transferKeeper.SendCoins(ctx, toCUAddr, fromCUAddr, sdk.NewCoins(sdk.NewCoin(symbol, sendAmt)))
	toCU = ck.GetCU(ctx, toCUAddr)
	balances = transferKeeper.GetAllBalance(ctx, toCU.GetAddress())
	assert.Equal(t, totalsupply.Sub(sendAmt), balances.AmountOf(symbol))

	fromCU = ck.GetCU(ctx, fromCUAddr)
	balances = transferKeeper.GetAllBalance(ctx, fromCU.GetAddress())
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(4), balances.AmountOf(sdk.NativeToken))
	assert.Equal(t, sendAmt, balances.AmountOf(symbol))

}

func TestHandleMsgNewTokenFail(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	tk := input.tk
	ck := input.ck
	hk := input.hrc10k
	transferKeeper := input.transferKeeper

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	assert.Nil(t, err)
	fromCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, fromCUAddr)
	transferKeeper.AddCoins(ctx, fromCU.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, types.DefaultIssueTokenFee.MulRaw(5))))
	ck.SetCU(ctx, fromCU)
	fromCU = ck.GetCU(ctx, fromCUAddr)
	balances := transferKeeper.GetAllBalance(ctx, fromCU.GetAddress())
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(5), balances.AmountOf(sdk.NativeToken))

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	assert.Nil(t, err)
	toCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, toCUAddr)
	ck.SetCU(ctx, toCU)
	balances = transferKeeper.GetAllBalance(ctx, toCU.GetAddress())
	assert.Equal(t, sdk.Coins(nil), balances)

	ti := tk.GetToken(ctx, "btc")
	assert.NotNil(t, ti)

	ti = tk.GetToken(ctx, "eth")
	assert.NotNil(t, ti)

	ti = tk.GetToken(ctx, "usdt")
	assert.NotNil(t, ti)


	//token already exist
	//msg := types.NewMsgNewToken(fromCUAddr, toCUAddr, "btc", 18, sdk.NewInt(1000000))
	//res := handleMsgNewToken(ctx, hk, msg)
	//assert.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)
	//
	//msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "eth", 18, sdk.NewInt(1000000))
	//res = handleMsgNewToken(ctx, hk, msg)
	//assert.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)
	//
	//msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "usdt", 18, sdk.NewInt(1000000))
	//res = handleMsgNewToken(ctx, hk, msg)
	//assert.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)

	//token is a reserved symbol
	//msg := types.NewMsgNewToken(fromCUAddr, toCUAddr, "eos", 18, sdk.NewInt(1000000))
	//res := handleMsgNewToken(ctx, hk, msg)
	//assert.Equal(t, types.CodeSymbolReserved, res.Code)
	//
	//msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "bsv", 18, sdk.NewInt(1000000))
	//res = handleMsgNewToken(ctx, hk, msg)
	//assert.Equal(t, types.CodeSymbolReserved, res.Code)
	//
	//msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "bch", 18, sdk.NewInt(1000000))
	//res = handleMsgNewToken(ctx, hk, msg)
	//assert.Equal(t, types.CodeSymbolReserved, res.Code)

	//fromAccount does not exist
	nonExistAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	msg := types.NewMsgNewToken(nonExistAddr, toCUAddr, "bhd", 18, sdk.NewInt(1000000))
	res := handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)

	//insufficient openFee
	param := hk.GetParams(ctx)
	param.IssueTokenFee = types.DefaultIssueTokenFee.MulRaw(6)
	hk.SetParams(ctx, param)
	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "bhd", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)

}
