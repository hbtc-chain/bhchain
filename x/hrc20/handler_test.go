package hrc20

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleMsgNewTokenSuccess(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	tk := input.tk
	ck := input.ck
	dk := input.dk
	hk := input.hrc20k
	transferKeeper := input.transferKeeper
	supplyKeeper := input.supplyKeeper

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	assert.Nil(t, err)
	fromCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, fromCUAddr)
	fromCU.AddCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, types.DefaultIssueTokenFee.MulRaw(5))))
	ck.SetCU(ctx, fromCU)
	fromCU = ck.GetCU(ctx, fromCUAddr)
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(5), fromCU.GetCoins().AmountOf(sdk.NativeToken))

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	assert.Nil(t, err)
	toCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, toCUAddr)
	assert.Equal(t, sdk.Coins(nil), toCU.GetCoins())

	openFee := sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, types.DefaultIssueTokenFee))
	totalsupply, ok := sdk.NewIntFromString("10000000000000000000000")
	assert.True(t, ok)

	msg := types.NewMsgNewToken(fromCUAddr, toCUAddr, "bhd", 18, totalsupply)

	res := handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeOK, res.Code)
	assert.Equal(t, 2, len(res.Events))
	assert.Equal(t, types.EventTypeNewToken, res.Events[1].Type)

	//check tokenInfo
	ti := tk.GetTokenInfo(ctx, "bhd")
	assert.Equal(t, "bhd", ti.Symbol.String())
	assert.Equal(t, sdk.NativeToken, ti.Chain.String())
	assert.Equal(t, uint64(18), ti.Decimals)
	assert.Equal(t, totalsupply, ti.TotalSupply)
	assert.Equal(t, fromCUAddr.String(), ti.Issuer)
	assert.Equal(t, sdk.AccountBased, ti.TokenType)
	assert.True(t, ti.IsSendEnabled)
	assert.False(t, ti.IsWithdrawalEnabled)
	assert.False(t, ti.IsDepositEnabled)
	assert.Equal(t, sdk.ZeroInt(), ti.CollectThreshold)
	assert.Equal(t, sdk.ZeroInt(), ti.DepositThreshold)
	assert.Equal(t, sdk.ZeroInt(), ti.OpenFee)
	assert.Equal(t, sdk.ZeroInt(), ti.SysOpenFee)
	assert.Equal(t, sdk.ZeroDec(), ti.WithdrawalFeeRate)
	assert.Equal(t, sdk.ZeroInt(), ti.SysTransferAmount())
	assert.Equal(t, sdk.ZeroInt(), ti.OpCUSysTransferAmount())
	assert.Equal(t, sdk.ZeroInt(), ti.GasLimit)
	assert.Equal(t, sdk.ZeroInt(), ti.GasPrice)
	assert.Equal(t, uint64(0), ti.MaxOpCUNumber)

	//check coins
	toCU = ck.GetCU(ctx, toCUAddr)
	assert.Equal(t, totalsupply, toCU.GetCoins().AmountOf("bhd"))
	fromCU = ck.GetCU(ctx, fromCUAddr)
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(4), fromCU.GetCoins().AmountOf(sdk.NativeToken))

	//check supply
	supply := supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("bhd")
	assert.Equal(t, supply, ti.TotalSupply)

	//check feepool
	feePool := dk.GetFeePool(ctx)
	assert.Equal(t, openFee.AmountOf(sdk.NativeToken), feePool.CommunityPool.AmountOf(sdk.NativeToken).TruncateInt())

	//sendcoins back to fromCUAddr
	sendAmt := sdk.NewInt(2000000000)
	res, err = transferKeeper.SendCoins(ctx, toCUAddr, fromCUAddr, sdk.NewCoins(sdk.NewCoin("bhd", sendAmt)))
	toCU = ck.GetCU(ctx, toCUAddr)
	assert.Equal(t, totalsupply.Sub(sendAmt), toCU.GetCoins().AmountOf("bhd"))

	fromCU = ck.GetCU(ctx, fromCUAddr)
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(4), fromCU.GetCoins().AmountOf(sdk.NativeToken))
	assert.Equal(t, sendAmt, fromCU.GetCoins().AmountOf("bhd"))

}

func TestHandleMsgNewTokenFail(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	tk := input.tk
	ck := input.ck
	hk := input.hrc20k
	//transferKeeper := input.transferKeeper

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	assert.Nil(t, err)
	fromCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, fromCUAddr)
	fromCU.AddCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, types.DefaultIssueTokenFee.MulRaw(5))))
	ck.SetCU(ctx, fromCU)
	fromCU = ck.GetCU(ctx, fromCUAddr)
	assert.Equal(t, types.DefaultIssueTokenFee.MulRaw(5), fromCU.GetCoins().AmountOf(sdk.NativeToken))

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	assert.Nil(t, err)
	toCU := ck.GetOrNewCU(ctx, sdk.CUTypeUser, toCUAddr)
	assert.Equal(t, sdk.Coins(nil), toCU.GetCoins())

	ti := tk.GetTokenInfo(ctx, "btc")
	assert.NotNil(t, ti)

	ti = tk.GetTokenInfo(ctx, "eth")
	assert.NotNil(t, ti)

	ti = tk.GetTokenInfo(ctx, "usdt")
	assert.NotNil(t, ti)

	//token already exist
	msg := types.NewMsgNewToken(fromCUAddr, toCUAddr, "btc", 18, sdk.NewInt(1000000))
	res := handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)

	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "eth", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)

	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "usdt", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)

	//token is a reserved symbol
	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "eos", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, types.CodeSymbolReserved, res.Code)

	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "bsv", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, types.CodeSymbolReserved, res.Code)

	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "bch", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, types.CodeSymbolReserved, res.Code)

	//fromAccount does not exist
	nonExistAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	msg = types.NewMsgNewToken(nonExistAddr, toCUAddr, "bhd", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeInvalidAccount, res.Code)

	//insufficient openFee
	param := hk.GetParams(ctx)
	param.IssueTokenFee = types.DefaultIssueTokenFee.MulRaw(6)
	hk.SetParams(ctx, param)
	msg = types.NewMsgNewToken(fromCUAddr, toCUAddr, "bhd", 18, sdk.NewInt(1000000))
	res = handleMsgNewToken(ctx, hk, msg)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)

}
