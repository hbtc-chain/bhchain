package openswap

import (
	"testing"
	"time"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/mint"
	"github.com/hbtc-chain/bhchain/x/openswap/keeper"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/params/subspace"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/hbtc-chain/bhchain/x/transfer"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	tk  token.Keeper
	ck  custodianunit.CUKeeperI
	k   keeper.Keeper
}

func setupTestInput() *testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	custodianunit.RegisterCodec(cdc)
	token.RegisterCodec(cdc)
	receipt.RegisterCodec(cdc)
	types.RegisterCodec(cdc)

	cuKey := sdk.NewKVStoreKey("cuKey")
	keyTransfer := sdk.NewKVStoreKey(transfer.StoreKey)
	keyParams := sdk.NewKVStoreKey("subspace")
	tokenKey := sdk.NewKVStoreKey(token.ModuleName)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	openswapKey := sdk.NewKVStoreKey(ModuleName)
	receiptKey := sdk.NewKVStoreKey(receipt.StoreKey)
	supplyKey := sdk.NewKVStoreKey(supply.StoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(cuKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyTransfer, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(receiptKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(tokenKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(supplyKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(openswapKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	cuSp := subspace.NewSubspace(cdc, keyParams, tkeyParams, cutypes.DefaultParamspace)
	openswapSp := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	bkSp := subspace.NewSubspace(cdc, keyParams, tkeyParams, transfer.DefaultParamspace)
	tk := token.NewKeeper(tokenKey, cdc, subspace.NewSubspace(cdc, keyParams, tkeyParams, token.DefaultParamspace))
	ck := custodianunit.NewCUKeeper(cdc, cuKey, &tk, cuSp, cutypes.ProtoBaseCU)
	rk := receipt.NewKeeper(cdc)
	bk := transfer.NewBaseKeeper(cdc, keyTransfer, ck, nil, nil, rk, nil, nil, bkSp, transfer.DefaultCodespace, nil)

	maccPerms := map[string][]string{
		mint.ModuleName:  {supply.Minter},
		types.ModuleName: {supply.Minter, supply.Burner},
	}
	sk := supply.NewKeeper(cdc, supplyKey, ck, bk, maccPerms)
	moduleCU := supply.NewEmptyModuleAccount(ModuleName, supply.Minter, supply.Burner)
	sk.SetModuleAccount(ctx, moduleCU)
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.NativeDefiToken, sdk.NewInt(100000000000000000)))
	sk.SetSupply(ctx, supply.NewSupply(totalSupply))
	k := NewKeeper(cdc, openswapKey, &tk, ck, rk, sk, bk, openswapSp)
	k.SetParams(ctx, types.DefaultParams())

	//init token info
	for _, tokenInfo := range token.TestTokenData {
		token := token.NewTokenInfo(tokenInfo.Symbol, tokenInfo.Chain, tokenInfo.Issuer, tokenInfo.TokenType,
			tokenInfo.IsSendEnabled, tokenInfo.IsDepositEnabled, tokenInfo.IsWithdrawalEnabled, tokenInfo.Decimals,
			tokenInfo.TotalSupply, tokenInfo.CollectThreshold, tokenInfo.DepositThreshold, tokenInfo.OpenFee,
			tokenInfo.SysOpenFee, tokenInfo.WithdrawalFeeRate, tokenInfo.SysTransferNum, tokenInfo.OpCUSysTransferNum,
			tokenInfo.GasLimit, tokenInfo.GasPrice, tokenInfo.MaxOpCUNumber, tokenInfo.Confirmations, tokenInfo.IsNonceBased)
		tk.SetTokenInfo(ctx, token)
	}

	return &testInput{
		cdc: cdc,
		ctx: ctx,
		ck:  ck,
		tk:  tk,
		k:   k,
	}
}

func TestHandleMsgAddLiquidity(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	cu := input.ck.GetOrNewCU(ctx, sdk.CUTypeUser, address)
	originAmount := sdk.NewInt(100000000)
	cu.AddCoins(sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	input.ck.SetCU(ctx, cu)
	k := input.k

	// test token not exists
	msg := types.NewMsgAddLiquidity(address, "fakebtc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res := handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	msg = types.NewMsgAddLiquidity(address, "usdt", "fakebtc", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetTokenInfo(ctx, "btc")
	btc.IsSendEnabled = false
	input.tk.SetTokenInfo(ctx, btc)
	msg = types.NewMsgAddLiquidity(address, "usdt", "btc", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	msg = types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.IsSendEnabled = true
	input.tk.SetTokenInfo(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 1)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test liquidity too low
	msg = types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(200), sdk.NewInt(800), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	assert.Contains(t, res.Log, "insufficient liquidity")

	// test insufficient funds
	msg = types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(800000000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientFunds, res.Code)
	assert.Contains(t, res.Log, "insufficient funds")

	// test success
	msg = types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.True(t, res.IsOK())

	cu = input.ck.GetCU(ctx, address)
	btcRemain := cu.GetCoins().AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000), btcRemain)
	usdtRemain := cu.GetCoins().AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000), usdtRemain)

	expectedPair := &types.TradingPair{
		TokenA:         "btc",
		TokenB:         "usdt",
		TokenAAmount:   sdk.NewInt(20000),
		TokenBAmount:   sdk.NewInt(8000000),
		TotalLiquidity: sdk.NewInt(400000),
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, "btc", "usdt", address))

	msg = types.NewMsgAddLiquidity(address, "usdt", "btc", sdk.NewInt(20000), sdk.NewInt(800), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg) // 实际注入 320000usdt, 800btc
	assert.True(t, res.IsOK())

	cu = input.ck.GetCU(ctx, address)
	btcRemain = cu.GetCoins().AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000+800), btcRemain)
	usdtRemain = cu.GetCoins().AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000+320000), usdtRemain)

	expectedPair.TokenAAmount = expectedPair.TokenAAmount.AddRaw(800)
	expectedPair.TokenBAmount = expectedPair.TokenBAmount.AddRaw(320000)
	expectedPair.TotalLiquidity = sdk.NewInt(416000)
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(416000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, "btc", "usdt", address))
}

func TestHandleMsgRemoveLiquidity(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	cu := input.ck.GetOrNewCU(ctx, sdk.CUTypeUser, address)
	originAmount := sdk.NewInt(100000000)
	cu.AddCoins(sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	input.ck.SetCU(ctx, cu)
	k := input.k

	// test unexistent trading pair
	msg := types.NewMsgRemoveLiquidity(address, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res := handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "no trading pair of btc-usdt")

	addMsg := types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	// test token not exists
	msg = types.NewMsgRemoveLiquidity(address, "fakebtc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	msg = types.NewMsgRemoveLiquidity(address, "usdt", "fakebtc", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetTokenInfo(ctx, "btc")
	btc.IsSendEnabled = false
	input.tk.SetTokenInfo(ctx, btc)
	msg = types.NewMsgRemoveLiquidity(address, "usdt", "btc", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	msg = types.NewMsgRemoveLiquidity(address, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.IsSendEnabled = true
	input.tk.SetTokenInfo(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgRemoveLiquidity(address, "btc", "usdt", sdk.NewInt(30000), 1)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test liquidity too big
	msg = types.NewMsgRemoveLiquidity(address, "btc", "usdt", sdk.NewInt(300000000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientFunds, res.Code)
	assert.Contains(t, res.Log, "insufficient liquidity, has 399000, need 300000000")

	// test success
	msg = types.NewMsgRemoveLiquidity(address, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg) // 赎回 0.075*20000 btc  0.075*8000000 usdt
	assert.True(t, res.IsOK())

	cu = input.ck.GetCU(ctx, address)
	btcRemain := cu.GetCoins().AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000-1500), btcRemain)
	usdtRemain := cu.GetCoins().AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000-600000), usdtRemain)

	expectedPair := &types.TradingPair{
		TokenA:         "btc",
		TokenB:         "usdt",
		TokenAAmount:   sdk.NewInt(18500),   // 20000 - 1500
		TokenBAmount:   sdk.NewInt(7400000), // 8000000 - 600000
		TotalLiquidity: sdk.NewInt(370000),  // 400000 - 30000
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000-1000-30000), k.GetLiquidity(ctx, "btc", "usdt", address))

	// 赎回全部流动性
	msg = types.NewMsgRemoveLiquidity(address, "usdt", "btc", sdk.NewInt(369000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg) // 赎回 18450 btc  7380000 usdt
	assert.True(t, res.IsOK())

	cu = input.ck.GetCU(ctx, address)
	btcRemain = cu.GetCoins().AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(50), btcRemain)
	usdtRemain = cu.GetCoins().AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(20000), usdtRemain)

	expectedPair = &types.TradingPair{
		TokenA:         "btc",
		TokenB:         "usdt",
		TokenAAmount:   sdk.NewInt(50),
		TokenBAmount:   sdk.NewInt(20000),
		TotalLiquidity: sdk.NewInt(1000),
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(0), k.GetLiquidity(ctx, "btc", "usdt", address))
}

func TestHandleMsgSwapFail(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	cu := input.ck.GetOrNewCU(ctx, sdk.CUTypeUser, address)
	originAmount := sdk.NewInt(100000000)
	cu.AddCoins(sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	input.ck.SetCU(ctx, cu)

	buyer := sdk.NewCUAddress()
	cu = input.ck.GetOrNewCU(ctx, sdk.CUTypeUser, buyer)
	cu.AddCoins(sdk.NewCoins(sdk.NewCoin("usdt", originAmount)))
	input.ck.SetCU(ctx, cu)

	k := input.k

	addMsg := types.NewMsgAddLiquidity(address, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	addMsg = types.NewMsgAddLiquidity(address, "eth", "usdt", sdk.NewInt(40000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()
	// test token not exists
	msg := types.NewMsgSwapExactIn(buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "fakebtc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetTokenInfo(ctx, "btc")
	btc.IsSendEnabled = false
	input.tk.SetTokenInfo(ctx, btc)
	msg = types.NewMsgSwapExactIn(buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.IsSendEnabled = true
	input.tk.SetTokenInfo(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgSwapExactIn(buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 1)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test insufficient funds
	msg = types.NewMsgSwapExactIn(buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"eth", "usdt"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientFunds, res.Code)
	assert.Contains(t, res.Log, "insufficient")

	// test no trading pair
	msg = types.NewMsgSwapExactIn(address, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"eth", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "not found")

	// test less than minReturn
	msg = types.NewMsgSwapExactIn(buyer, referer, receiver, sdk.NewInt(40000), sdk.NewInt(100), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	assert.Contains(t, res.Log, "insufficient amount out, min: 100, got: 99")
}

func TestHandleMsgSwapSuccess(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	cu := input.ck.GetOrNewCU(ctx, sdk.CUTypeUser, address)
	originAmount := sdk.NewInt(100000000)
	cu.AddCoins(sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	input.ck.SetCU(ctx, cu)

	buyer := sdk.NewCUAddress()
	cu = input.ck.GetOrNewCU(ctx, sdk.CUTypeUser, buyer)
	cu.AddCoins(sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	input.ck.SetCU(ctx, cu)

	k := input.k

	btcUsdtAmountBtc := sdk.NewInt(20000)
	btcUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg := types.NewMsgAddLiquidity(address, "btc", "usdt", btcUsdtAmountBtc, btcUsdtAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	ethUsdtAmountEth := sdk.NewInt(40000)
	ethUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg = types.NewMsgAddLiquidity(address, "eth", "usdt", ethUsdtAmountEth, ethUsdtAmountUsdt, 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()

	// test success
	// 20usdt -> referer, 39880usdt 兑换 99btc -> receiver, 100usdt -> 资金池
	amtIn := sdk.NewInt(40000)
	msg := types.NewMsgSwapExactIn(buyer, referer, receiver, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())

	// buyer 减少 40000usdt
	cu = input.ck.GetCU(ctx, buyer)
	usdtRemain := cu.GetCoins().AmountOf("usdt")
	assert.Equal(t, originAmount.Sub(amtIn), usdtRemain)

	// referer 增加 20usdt
	cu = input.ck.GetCU(ctx, referer)
	usdtRemain = cu.GetCoins().AmountOf("usdt")
	expectedRefererBonus := amtIn.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererBonus, usdtRemain)

	feeRate := types.DefaultFeeRate.Add(types.DefaultRepurchaseRate)
	feeAmt := amtIn.ToDec().Mul(feeRate).TruncateInt()
	realIn := amtIn.Sub(expectedRefererBonus).Sub(feeAmt)

	// receiver 增加 99btc
	cu = input.ck.GetCU(ctx, receiver)
	btcRemain := cu.GetCoins().AmountOf("btc")
	expectOut := realIn.Mul(btcUsdtAmountBtc).Quo(realIn.Add(btcUsdtAmountUsdt))
	assert.Equal(t, expectOut, btcRemain)

	expectedBtcUsdtPair := &types.TradingPair{
		TokenA:         "btc",
		TokenB:         "usdt",
		TokenAAmount:   btcUsdtAmountBtc.Sub(expectOut),
		TokenBAmount:   btcUsdtAmountUsdt.Add(realIn).Add(feeAmt),
		TotalLiquidity: sdk.NewInt(400000),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(ctx, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, "btc", "usdt", address))

	// test success2
	// 20eth -> referer, 39840eth 兑换 3991983usdt, 120eth -> 资金池
	// 3991usdt -> referer, 3976017usdt 兑换 6585btc -> receiver, 11975usdt -> 资金池
	receiver = sdk.NewCUAddress()
	amtIn = sdk.NewInt(40000)
	msg = types.NewMsgSwapExactIn(buyer, referer, receiver, amtIn, sdk.ZeroInt(), []sdk.Symbol{"eth", "usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())

	// buyer 减少 40000eth
	cu = input.ck.GetCU(ctx, buyer)
	ethRemain := cu.GetCoins().AmountOf("eth")
	assert.Equal(t, originAmount.Sub(amtIn), ethRemain)

	// referer 增加 20eth, 3991usdt
	cu = input.ck.GetCU(ctx, referer)
	ethRemain = cu.GetCoins().AmountOf("eth")
	expectedRefererEthBonus := amtIn.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererEthBonus, ethRemain)

	feeEthAmt := amtIn.ToDec().Mul(feeRate).TruncateInt()
	realIn = amtIn.Sub(expectedRefererEthBonus).Sub(feeEthAmt)
	expectUsdtOut := realIn.Mul(ethUsdtAmountUsdt).Quo(realIn.Add(ethUsdtAmountEth))

	usdtRemain = cu.GetCoins().AmountOf("usdt")
	expectedRefererUsdtBonus := expectUsdtOut.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererUsdtBonus.Add(expectedRefererBonus), usdtRemain)

	// receiver 增加 6585btc
	cu = input.ck.GetCU(ctx, receiver)
	btcRemain = cu.GetCoins().AmountOf("btc")
	usdtFee := expectUsdtOut.ToDec().Mul(feeRate).TruncateInt()
	realUsdtIn := expectUsdtOut.Sub(expectedRefererUsdtBonus).Sub(usdtFee)
	expectBtcOut := realUsdtIn.Mul(expectedBtcUsdtPair.TokenAAmount).Quo(realUsdtIn.Add(expectedBtcUsdtPair.TokenBAmount))
	assert.Equal(t, expectBtcOut, btcRemain)

	expectedBtcUsdtPair = &types.TradingPair{
		TokenA:         "btc",
		TokenB:         "usdt",
		TokenAAmount:   expectedBtcUsdtPair.TokenAAmount.Sub(expectBtcOut),
		TokenBAmount:   expectedBtcUsdtPair.TokenBAmount.Add(realUsdtIn).Add(usdtFee),
		TotalLiquidity: sdk.NewInt(400000),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(ctx, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, "btc", "usdt", address))

	expectedEthUsdtPair := &types.TradingPair{
		TokenA:         "eth",
		TokenB:         "usdt",
		TokenAAmount:   ethUsdtAmountEth.Add(realIn).Add(feeEthAmt),
		TokenBAmount:   ethUsdtAmountUsdt.Sub(expectUsdtOut),
		TotalLiquidity: sdk.NewInt(565685),
	}
	assert.Equal(t, expectedEthUsdtPair, k.GetTradingPair(ctx, "eth", "usdt"))
	assert.Equal(t, sdk.NewInt(565685).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, "eth", "usdt", address))
}
