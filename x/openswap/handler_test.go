package openswap

import (
	"fmt"
	"math/big"
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

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	tk  token.Keeper
	trk types.TransferKeeper
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
	ctx = ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())

	cuSp := subspace.NewSubspace(cdc, keyParams, tkeyParams, cutypes.DefaultParamspace)
	openswapSp := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	bkSp := subspace.NewSubspace(cdc, keyParams, tkeyParams, transfer.DefaultParamspace)
	tk := token.NewKeeper(tokenKey, cdc)
	ck := custodianunit.NewCUKeeper(cdc, cuKey, cuSp, cutypes.ProtoBaseCU)
	rk := receipt.NewKeeper(cdc)
	trk := transfer.NewBaseKeeper(cdc, keyTransfer, ck, nil, nil, nil, rk, nil, nil, bkSp, transfer.DefaultCodespace, nil)

	maccPerms := map[string][]string{
		mint.ModuleName:  {supply.Minter},
		types.ModuleName: {supply.Minter, supply.Burner},
	}
	sk := supply.NewKeeper(cdc, supplyKey, ck, trk, maccPerms)
	moduleCU := supply.NewEmptyModuleAccount(ModuleName, supply.Minter, supply.Burner)
	sk.SetModuleAccount(ctx, moduleCU)
	k := NewKeeper(cdc, openswapKey, &tk, rk, sk, trk, openswapSp)
	k.SetParams(ctx, types.DefaultParams())

	//init token info
	for _, tokenInfo := range token.TestTokenData {
		tk.SetToken(ctx, tokenInfo)
	}

	return &testInput{
		cdc: cdc,
		ctx: ctx,
		ck:  ck,
		tk:  tk,
		trk: trk,
		k:   k,
	}
}

func TestHandleMsgAddLiquidity(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	// test token not exists
	msg := types.NewMsgAddLiquidity(address, 0, "fakebtc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res := handleMsgAddLiquidity(ctx.WithMultiStore(ctx.MultiStore()), k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	msg = types.NewMsgAddLiquidity(address, 0, "usdt", "fakebtc", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx.WithMultiStore(ctx.MultiStore()), k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetIBCToken(ctx, "btc")
	btc.SendEnabled = false
	input.tk.SetToken(ctx.WithMultiStore(ctx.MultiStore()), btc)
	msg = types.NewMsgAddLiquidity(address, 0, "usdt", "btc", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx.WithMultiStore(ctx.MultiStore()), k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	msg = types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx.WithMultiStore(ctx.MultiStore()), k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.SendEnabled = true
	input.tk.SetToken(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 1)
	res = handleMsgAddLiquidity(ctx.WithMultiStore(ctx.MultiStore()), k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test liquidity too low
	msg = types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(200), sdk.NewInt(800), 999999999999)
	res = handleMsgAddLiquidity(ctx.WithMultiStore(ctx.MultiStore()), k, msg)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	assert.Contains(t, res.Log, "insufficient liquidity")

	// test insufficient funds
	msg = types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(800000000000), 999999999999)
	runCtx := ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgAddLiquidity(runCtx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)
	assert.Contains(t, res.Log, "balance not enough")

	// test success
	msg = types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.True(t, res.IsOK())

	coins := input.trk.GetAllBalance(ctx, address)
	btcRemain := coins.AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000), btcRemain)
	usdtRemain := coins.AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000), usdtRemain)

	expectedPair := &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      sdk.NewInt(20000),
		TokenBAmount:      sdk.NewInt(8000000),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, address, 0, "btc", "usdt"))

	msg = types.NewMsgAddLiquidity(address, 0, "usdt", "btc", sdk.NewInt(20000), sdk.NewInt(800), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg) // 实际注入 20000usdt, 50btc
	assert.True(t, res.IsOK())

	coins = input.trk.GetAllBalance(ctx, address)
	btcRemain = coins.AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000+50), btcRemain)
	usdtRemain = coins.AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000+20000), usdtRemain)

	expectedPair.TokenAAmount = expectedPair.TokenAAmount.AddRaw(50)
	expectedPair.TokenBAmount = expectedPair.TokenBAmount.AddRaw(20000)
	expectedPair.TotalLiquidity = sdk.NewInt(401000)
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(401000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, address, 0, "btc", "usdt"))
}

func TestHandleMsgRemoveLiquidity(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	// test unexistent trading pair
	msg := types.NewMsgRemoveLiquidity(address, 0, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res := handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "btc-usdt trading pair does not exist")

	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	// test token not exists
	msg = types.NewMsgRemoveLiquidity(address, 0, "fakebtc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	msg = types.NewMsgRemoveLiquidity(address, 0, "usdt", "fakebtc", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetIBCToken(ctx, "btc")
	btc.SendEnabled = false
	input.tk.SetToken(ctx, btc)
	msg = types.NewMsgRemoveLiquidity(address, 0, "usdt", "btc", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	msg = types.NewMsgRemoveLiquidity(address, 0, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.SendEnabled = true
	input.tk.SetToken(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgRemoveLiquidity(address, 0, "btc", "usdt", sdk.NewInt(30000), 1)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test liquidity too big
	msg = types.NewMsgRemoveLiquidity(address, 0, "btc", "usdt", sdk.NewInt(300000000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientFunds, res.Code)
	assert.Contains(t, res.Log, "insufficient liquidity, has 399000, need 300000000")

	// test success
	msg = types.NewMsgRemoveLiquidity(address, 0, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg) // 赎回 0.075*20000 btc  0.075*8000000 usdt
	assert.True(t, res.IsOK())

	coins := input.trk.GetAllBalance(ctx, address)
	btcRemain := coins.AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000-1500), btcRemain)
	usdtRemain := coins.AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000-600000), usdtRemain)

	expectedPair := &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      sdk.NewInt(18500),   // 20000 - 1500
		TokenBAmount:      sdk.NewInt(7400000), // 8000000 - 600000
		TotalLiquidity:    sdk.NewInt(370000),  // 400000 - 30000
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000-1000-30000), k.GetLiquidity(ctx, address, 0, "btc", "usdt"))

	// 赎回全部流动性
	msg = types.NewMsgRemoveLiquidity(address, 0, "usdt", "btc", sdk.NewInt(369000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg) // 赎回 18450 btc  7380000 usdt
	assert.True(t, res.IsOK())

	coins = input.trk.GetAllBalance(ctx, address)
	btcRemain = coins.AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(50), btcRemain)
	usdtRemain = coins.AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(20000), usdtRemain)

	expectedPair = &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      sdk.NewInt(50),
		TokenBAmount:      sdk.NewInt(20000),
		TotalLiquidity:    sdk.NewInt(1000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(0), k.GetLiquidity(ctx, address, 0, "btc", "usdt"))
}

func TestHandleMsgSwapFail(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("usdt", originAmount)))
	k := input.k

	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	addMsg = types.NewMsgAddLiquidity(address, 0, "eth", "usdt", sdk.NewInt(40000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()
	// test token not exists
	msg := types.NewMsgSwapExactIn(0, buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "fakebtc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	msg2 := types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(1000000), sdk.NewInt(40000), []sdk.Symbol{"usdt", "fakebtc"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetIBCToken(ctx, "btc")
	btc.SendEnabled = false
	input.tk.SetToken(ctx, btc)
	msg = types.NewMsgSwapExactIn(0, buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(1000000), sdk.NewInt(40000), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.SendEnabled = true
	input.tk.SetToken(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgSwapExactIn(0, buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 1)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(1000000), sdk.NewInt(40000), []sdk.Symbol{"usdt", "btc"}, 1)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test insufficient funds
	msg = types.NewMsgSwapExactIn(0, buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"eth", "usdt"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)
	assert.Contains(t, res.Log, "balance not enough")

	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(1000000), sdk.NewInt(40000), []sdk.Symbol{"eth", "usdt"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)
	assert.Contains(t, res.Log, "balance not enough")

	// test no trading pair
	msg = types.NewMsgSwapExactIn(0, address, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"eth", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "eth-btc trading pair does not exist")

	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(1000000), sdk.NewInt(40000), []sdk.Symbol{"eth", "btc"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "eth-btc trading pair does not exist")

	// test less than minReturn
	msg = types.NewMsgSwapExactIn(0, buyer, referer, receiver, sdk.NewInt(40000), sdk.NewInt(100), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	assert.Contains(t, res.Log, "insufficient amount out, min: 100, got: 99")

	msg = types.NewMsgSwapExactIn(0, buyer, referer, receiver, sdk.NewInt(1000), sdk.NewInt(10000), []sdk.Symbol{"btc", "usdt", "eth"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	// 1000 btc -> 379864 usdt, 379864 usdt -> 1808 eth
	assert.Contains(t, res.Log, "insufficient amount out, min: 10000, got: 1808")

	// test excessive amount in
	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(10000), sdk.NewInt(40000), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	assert.Contains(t, res.Log, "excessive amount in, max: 40000, got: 8024073") // ceil(8000000/0.997) = 8024073

	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(10000), sdk.NewInt(400), []sdk.Symbol{"btc", "usdt", "eth"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeAmountError, res.Code)
	// 10000 eth needs 2674691 usdt, 2674691 usdt needs 10076 btc
	assert.Contains(t, res.Log, "excessive amount in, max: 400, got: 10076")

	// test insufficient reserve out
	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(20000), sdk.NewInt(40000000000), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "insufficient reserve out, have 20000, need 20000")

	msg2 = types.NewMsgSwapExactOut(0, buyer, referer, receiver, sdk.NewInt(20000), sdk.NewInt(40000000000), []sdk.Symbol{"btc", "usdt", "eth"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "insufficient reserve out, have 8000000, need 8024073") // ceil(8000000/0.997) = 8024073

	// test dex not exists
	msg = types.NewMsgSwapExactIn(1, buyer, referer, receiver, sdk.NewInt(40000), sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "dex id 1 not found")

	msg2 = types.NewMsgSwapExactOut(1, buyer, referer, receiver, sdk.NewInt(1000000), sdk.NewInt(40000), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactOut(ctx, k, msg2)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "dex id 1 not found")
}

func TestHandleMsgSwapExactInSuccess(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	btcUsdtAmountBtc := sdk.NewInt(20000)
	btcUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", btcUsdtAmountBtc, btcUsdtAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	ethUsdtAmountEth := sdk.NewInt(40000)
	ethUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg = types.NewMsgAddLiquidity(address, 0, "eth", "usdt", ethUsdtAmountEth, ethUsdtAmountUsdt, 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()

	// test success
	// 4usdt -> referer, 39880usdt 兑换 99btc -> receiver, 116usdt -> 资金池
	amtIn := sdk.NewInt(40000)
	msg := types.NewMsgSwapExactIn(0, buyer, referer, receiver, amtIn, sdk.NewInt(1), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	runCtx := ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgSwapExactIn(runCtx, k, msg)
	assert.True(t, res.IsOK())

	// buyer 减少 40000usdt
	coins := input.trk.GetAllBalance(runCtx, buyer)
	usdtRemain := coins.AmountOf("usdt")
	assert.Equal(t, originAmount.Sub(amtIn), usdtRemain)

	// referer 增加 4usdt
	coins = input.trk.GetAllBalance(runCtx, referer)
	usdtRemain = coins.AmountOf("usdt")
	expectedRefererBonus := amtIn.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererBonus, usdtRemain)

	feeRate := types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)
	feeAmt := amtIn.ToDec().Mul(feeRate).TruncateInt()
	realIn := amtIn.Sub(expectedRefererBonus).Sub(feeAmt)

	// receiver 增加 99btc
	coins = input.trk.GetAllBalance(runCtx, receiver)
	btcRemain := coins.AmountOf("btc")
	expectOut := realIn.Mul(btcUsdtAmountBtc).Quo(realIn.Add(btcUsdtAmountUsdt))
	assert.Equal(t, expectOut, btcRemain)

	expectedBtcUsdtPair := &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      btcUsdtAmountBtc.Sub(expectOut),
		TokenBAmount:      btcUsdtAmountUsdt.Add(realIn).Add(feeAmt),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(runCtx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "btc", "usdt"))

	// test success2
	msg = types.NewMsgSwapExactIn(0, buyer, referer, receiver, amtIn, sdk.ZeroInt(), []sdk.Symbol{"eth", "usdt", "btc"}, 999999999999)
	runCtx = ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgSwapExactIn(runCtx, k, msg)
	assert.True(t, res.IsOK())

	// buyer 减少 40000eth
	coins = input.trk.GetAllBalance(runCtx, buyer)
	ethRemain := coins.AmountOf("eth")
	assert.Equal(t, originAmount.Sub(amtIn), ethRemain)

	// referer 增加 20eth, 3991usdt
	coins = input.trk.GetAllBalance(runCtx, referer)
	ethRemain = coins.AmountOf("eth")
	expectedRefererEthBonus := amtIn.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererEthBonus, ethRemain)

	feeEthAmt := amtIn.ToDec().Mul(feeRate).TruncateInt()
	realIn = amtIn.Sub(expectedRefererEthBonus).Sub(feeEthAmt)
	expectUsdtOut := realIn.Mul(ethUsdtAmountUsdt).Quo(realIn.Add(ethUsdtAmountEth))

	usdtRemain = coins.AmountOf("usdt")
	expectedRefererUsdtBonus := expectUsdtOut.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererUsdtBonus, usdtRemain)

	coins = input.trk.GetAllBalance(runCtx, receiver)
	btcRemain = coins.AmountOf("btc")
	usdtFee := expectUsdtOut.ToDec().Mul(feeRate).TruncateInt()
	realUsdtIn := expectUsdtOut.Sub(expectedRefererUsdtBonus).Sub(usdtFee)
	expectBtcOut := realUsdtIn.Mul(btcUsdtAmountBtc).Quo(realUsdtIn.Add(btcUsdtAmountUsdt))
	assert.Equal(t, expectBtcOut, btcRemain, expectBtcOut.String())

	expectedBtcUsdtPair = &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      btcUsdtAmountBtc.Sub(expectBtcOut),
		TokenBAmount:      btcUsdtAmountUsdt.Add(realUsdtIn).Add(usdtFee),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(runCtx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "btc", "usdt"))

	expectedEthUsdtPair := &types.TradingPair{
		TokenA:            "eth",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      ethUsdtAmountEth.Add(realIn).Add(feeEthAmt),
		TokenBAmount:      ethUsdtAmountUsdt.Sub(expectUsdtOut),
		TotalLiquidity:    sdk.NewInt(565685),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedEthUsdtPair, k.GetTradingPair(runCtx, 0, "eth", "usdt"))
	assert.Equal(t, sdk.NewInt(565685).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "eth", "usdt"))
}

func TestHandleMsgSwapExactOutSuccess(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	btcUsdtAmountBtc := sdk.NewInt(20000)
	btcUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", btcUsdtAmountBtc, btcUsdtAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	ethUsdtAmountEth := sdk.NewInt(40000)
	ethUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg = types.NewMsgAddLiquidity(address, 0, "eth", "usdt", ethUsdtAmountEth, ethUsdtAmountUsdt, 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()

	amtOut := sdk.NewInt(100)
	msg := types.NewMsgSwapExactOut(0, buyer, referer, receiver, amtOut, sdk.NewInt(50000), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	runCtx := ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgSwapExactOut(runCtx, k, msg)
	assert.True(t, res.IsOK())

	// receiver 增加 100btc
	coins := input.trk.GetAllBalance(runCtx, receiver)
	btcRemain := coins.AmountOf("btc")
	assert.Equal(t, amtOut, btcRemain)

	// buyer 减少 40322usdt
	amtIn := amtOut.Mul(btcUsdtAmountUsdt).Quo(btcUsdtAmountBtc.Sub(amtOut))
	feeRate := types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate).Add(types.DefaultRefererTransactionBonusRate)
	cost := amtIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt().AddRaw(1)
	coins = input.trk.GetAllBalance(runCtx, buyer)
	usdtRemain := coins.AmountOf("usdt")
	assert.Equal(t, originAmount.Sub(cost), usdtRemain)

	coins = input.trk.GetAllBalance(runCtx, referer)
	usdtRemain = coins.AmountOf("usdt")
	expectedRefererBonus := cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, expectedRefererBonus, usdtRemain)

	lpRewardAmt := cost.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	realIn := cost.Sub(cost.ToDec().Mul(feeRate).TruncateInt())
	expectedBtcUsdtPair := &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      btcUsdtAmountBtc.Sub(amtOut),
		TokenBAmount:      btcUsdtAmountUsdt.Add(realIn).Add(lpRewardAmt),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(runCtx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "btc", "usdt"))

	// test success2
	receiver = sdk.NewCUAddress()
	amtOut = sdk.NewInt(5000)
	msg = types.NewMsgSwapExactOut(0, buyer, referer, receiver, amtOut, sdk.NewInt(50000), []sdk.Symbol{"eth", "usdt", "btc"}, 999999999999)
	runCtx = ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgSwapExactOut(runCtx, k, msg)
	assert.True(t, res.IsOK(), res.Log)

	// receiver 增加 100btc
	coins = input.trk.GetAllBalance(runCtx, receiver)
	btcRemain = coins.AmountOf("btc")
	assert.Equal(t, amtOut, btcRemain)

	usdtAmtIn := amtOut.Mul(btcUsdtAmountUsdt).Quo(btcUsdtAmountBtc.Sub(amtOut))
	usdtCost := usdtAmtIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt().AddRaw(1)
	ethAmtIn := usdtCost.Mul(ethUsdtAmountEth).Quo(ethUsdtAmountUsdt.Sub(usdtCost))
	ethCost := ethAmtIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt().AddRaw(1)

	ethReferBonus := ethCost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	ethLpReward := ethCost.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	ethAmtIn = ethCost.Sub(ethReferBonus).Sub(ethLpReward)
	usdtOut := ethAmtIn.Mul(ethUsdtAmountUsdt).Quo(ethAmtIn.Add(ethUsdtAmountEth))
	usdtReferBonus := usdtOut.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	usdtLpReward := usdtOut.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	usdtAmtIn = usdtOut.Sub(usdtReferBonus).Sub(usdtLpReward)

	// buyer 减少 ethCost eth
	coins = input.trk.GetAllBalance(runCtx, buyer)
	ethRemain := coins.AmountOf("eth")
	assert.Equal(t, originAmount.Sub(ethCost), ethRemain)

	// referer 增加 ethReferBonus eth, usdtReferBonus usdt
	coins = input.trk.GetAllBalance(runCtx, referer)
	assert.Equal(t, usdtReferBonus, coins.AmountOf("usdt"))
	assert.Equal(t, ethReferBonus, coins.AmountOf("eth"))

	expectedBtcUsdtPair = &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      btcUsdtAmountBtc.Sub(amtOut),
		TokenBAmount:      btcUsdtAmountUsdt.Add(usdtAmtIn).Add(usdtLpReward),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(runCtx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "btc", "usdt"))

	expectedEthUsdtPair := &types.TradingPair{
		TokenA:            "eth",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      ethUsdtAmountEth.Add(ethAmtIn).Add(ethLpReward),
		TokenBAmount:      ethUsdtAmountUsdt.Sub(usdtOut),
		TotalLiquidity:    sdk.NewInt(565685),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedEthUsdtPair, k.GetTradingPair(runCtx, 0, "eth", "usdt"))
	assert.Equal(t, sdk.NewInt(565685).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "eth", "usdt"))
}

func TestHandleMsgSwapEdgeCase(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	btcUsdtAmountBtc := sdk.NewInt(20000)
	btcUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", btcUsdtAmountBtc, btcUsdtAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	btcEthAmountBtc := sdk.NewInt(40000)
	btcEthAmountEth := sdk.NewInt(8000000)
	addMsg = types.NewMsgAddLiquidity(address, 0, "btc", "eth", btcEthAmountBtc, btcEthAmountEth, 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()

	amtIn := sdk.NewInt(1)
	msg := types.NewMsgSwapExactIn(0, buyer, referer, receiver, amtIn, sdk.NewInt(0), []sdk.Symbol{"eth", "btc", "usdt"}, 999999999999)
	runCtx := ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgSwapExactIn(runCtx, k, msg)
	assert.True(t, res.IsOK())

	coins := input.trk.GetAllBalance(runCtx, receiver)
	assert.True(t, coins.IsZero())

	coins = input.trk.GetAllBalance(runCtx, referer)
	assert.True(t, coins.IsZero())

	coins = input.trk.GetAllBalance(runCtx, buyer)
	assert.Equal(t, originAmount.SubRaw(1), coins.AmountOf("eth"))
	assert.Equal(t, originAmount, coins.AmountOf("btc"))
	assert.Equal(t, originAmount, coins.AmountOf("usdt"))

	expectedBtcEthPair := &types.TradingPair{
		TokenA:            "eth",
		TokenB:            "btc",
		IsPublic:          true,
		TokenAAmount:      btcEthAmountEth.AddRaw(1),
		TokenBAmount:      btcEthAmountBtc,
		TotalLiquidity:    sdk.NewInt(565685),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcEthPair, k.GetTradingPair(runCtx, 0, "btc", "eth"))
	assert.Equal(t, sdk.NewInt(565685).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "eth", "btc"))

	expectedBtcUsdtPair := &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      btcUsdtAmountBtc,
		TokenBAmount:      btcUsdtAmountUsdt,
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(runCtx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "btc", "usdt"))

	amtOut := sdk.NewInt(1)
	msg2 := types.NewMsgSwapExactOut(0, buyer, referer, receiver, amtOut, sdk.NewInt(10000), []sdk.Symbol{"usdt", "btc", "eth"}, 999999999999)
	runCtx = ctx.WithMultiStore(ctx.MultiStore().CacheMultiStore())
	res = handleMsgSwapExactOut(runCtx, k, msg2)
	assert.True(t, res.IsOK())

	feeRate := types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate).Add(types.DefaultRefererTransactionBonusRate)
	btcAmtIn := amtOut.Mul(btcEthAmountBtc).Quo(btcEthAmountEth.Sub(amtOut))
	btcCost := btcAmtIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt().AddRaw(1)
	usdtAmtIn := btcCost.Mul(btcUsdtAmountUsdt).Quo(btcUsdtAmountBtc.Sub(btcCost))
	usdtCost := usdtAmtIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt().AddRaw(1)

	usdtReferBonus := usdtCost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	usdtLpReward := usdtCost.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	usdtAmtIn = usdtCost.Sub(usdtReferBonus).Sub(usdtLpReward)
	btcOut := usdtAmtIn.Mul(btcUsdtAmountBtc).Quo(usdtAmtIn.Add(btcUsdtAmountUsdt))
	btcReferBonus := btcOut.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	btcLpReward := btcOut.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	btcAmtIn = btcOut.Sub(btcReferBonus).Sub(btcLpReward)
	ethOut := btcAmtIn.Mul(btcEthAmountEth).Quo(btcAmtIn.Add(btcEthAmountBtc))

	coins = input.trk.GetAllBalance(runCtx, receiver)
	assert.Equal(t, ethOut, coins.AmountOf("eth"))
	// buyer 减少 usdtCost usdt
	coins = input.trk.GetAllBalance(runCtx, buyer)
	assert.Equal(t, originAmount.Sub(usdtCost), coins.AmountOf("usdt"))

	// referer 增加 btcReferBonus btc, usdtReferBonus usdt
	coins = input.trk.GetAllBalance(runCtx, referer)
	assert.True(t, usdtReferBonus.Equal(coins.AmountOf("usdt")))
	assert.True(t, btcReferBonus.Equal(coins.AmountOf("btc")))

	expectedBtcUsdtPair = &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      btcUsdtAmountBtc.Sub(btcOut),
		TokenBAmount:      btcUsdtAmountUsdt.Add(usdtAmtIn).Add(usdtLpReward),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcUsdtPair, k.GetTradingPair(runCtx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "btc", "usdt"))

	expectedBtcEthPair = &types.TradingPair{
		TokenA:            "eth",
		TokenB:            "btc",
		IsPublic:          true,
		TokenAAmount:      btcEthAmountEth.Sub(ethOut),
		TokenBAmount:      btcEthAmountBtc.Add(btcAmtIn).Add(btcLpReward),
		TotalLiquidity:    sdk.NewInt(565685),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedBtcEthPair, k.GetTradingPair(runCtx, 0, "btc", "eth"))
	assert.Equal(t, sdk.NewInt(565685).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(runCtx, address, 0, "eth", "btc"))
}

func TestHandleMsgCreateDex(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.k

	assert.Nil(t, k.GetDex(ctx, 1))

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	msg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res := handleMsgCreateDex(ctx, k, msg)
	assert.True(t, res.IsOK())
	assert.Len(t, res.Events, 1)
	assert.Equal(t, types.EventTypeCreateDex, res.Events[0].Type)
	assert.Len(t, res.Events[0].Attributes, 1)
	assert.Equal(t, []byte(types.AttributeKeyDexID), res.Events[0].Attributes[0].Key)
	assert.Equal(t, []byte("1"), res.Events[0].Attributes[0].Value)
	expectedDex := &types.Dex{
		ID:             1,
		Owner:          dexOwner,
		Name:           dexName,
		IncomeReceiver: incomeReceiver,
	}
	assert.Equal(t, expectedDex, k.GetDex(ctx, 1))

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	res = handleMsgCreateDex(ctx, k, msg)
	assert.True(t, res.IsOK())
	assert.Len(t, res.Events, 1)
	assert.Equal(t, types.EventTypeCreateDex, res.Events[0].Type)
	assert.Len(t, res.Events[0].Attributes, 1)
	assert.Equal(t, []byte(types.AttributeKeyDexID), res.Events[0].Attributes[0].Key)
	assert.Equal(t, []byte("2"), res.Events[0].Attributes[0].Value)
	expectedDex = &types.Dex{
		ID:             2,
		Owner:          dexOwner,
		Name:           dexName,
		IncomeReceiver: incomeReceiver,
	}
	assert.Equal(t, expectedDex, k.GetDex(ctx, 2))

	assert.Nil(t, k.GetDex(ctx, 3))
}

func TestHandleMsgEditDex(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.k

	// test dex not exists
	msg := types.NewMsgEditDex(sdk.NewCUAddress(), 1, "hbtc", nil)
	res := handleMsgEditDex(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "dex id 1 not found")

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	createMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createMsg)
	assert.True(t, res.IsOK())

	// test not owner
	newCU := sdk.NewCUAddress()
	msg = types.NewMsgEditDex(newCU, 1, "hbtc", nil)
	res = handleMsgEditDex(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, fmt.Sprintf("dex 1 belongs to %s, not %s", dexOwner.String(), newCU.String()))

	// test success
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgEditDex(dexOwner, 1, "hbtc", &newCU)
	res = handleMsgEditDex(ctx, k, msg)
	assert.True(t, res.IsOK())
	assert.Len(t, res.Events, 1)
	assert.Equal(t, types.EventTypeEditDex, res.Events[0].Type)
	assert.Len(t, res.Events[0].Attributes, 1)
	assert.Equal(t, []byte(types.AttributeKeyDexID), res.Events[0].Attributes[0].Key)
	assert.Equal(t, []byte("1"), res.Events[0].Attributes[0].Value)
	expectedDex := &types.Dex{
		ID:             1,
		Owner:          dexOwner,
		Name:           "hbtc",
		IncomeReceiver: newCU,
	}
	assert.Equal(t, expectedDex, k.GetDex(ctx, 1))
}

func TestHandleMsgCreateTradingPair(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.k

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	// test dex not exists
	msg := types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", true, lpRewardRate, refererRewardRate)
	res := handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "dex id 1 not found")

	createMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createMsg)
	assert.True(t, res.IsOK())

	// test not owner
	newCU := sdk.NewCUAddress()
	msg = types.NewMsgCreateTradingPair(newCU, 1, "btc", "eth", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, fmt.Sprintf("dex 1 belongs to %s, not %s", dexOwner.String(), newCU.String()))

	// test token not exists
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "fakebtc", "eth", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "fakeeth", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakeeth does not exist")

	// test token not sendable
	btc := input.tk.GetIBCToken(ctx, "btc")
	btc.SendEnabled = false
	input.tk.SetToken(ctx.WithMultiStore(ctx.MultiStore()), btc)
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "eth", "btc", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.SendEnabled = true
	input.tk.SetToken(ctx, btc)
	// test referer reward too small
	smallRefererRate := types.DefaultRefererTransactionBonusRate.Sub(sdk.SmallestDec())
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", true, lpRewardRate, smallRefererRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "public pair's referer reward rate must be larger than")

	// test sum of fee rate too large
	bigRefererRate := types.DefaultMaxFeeRate.Sub(types.DefaultRepurchaseRate).Sub(types.DefaultLpRewardRate).Add(sdk.SmallestDec())
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", true, sdk.ZeroDec(), bigRefererRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "sum of lp reward rate and referer reward rate is too large")

	bigRefererRate = types.DefaultMaxFeeRate.Sub(types.DefaultRepurchaseRate).Sub(lpRewardRate).Add(sdk.SmallestDec())
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", false, lpRewardRate, bigRefererRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "sum of lp reward rate and referer reward rate is too large")

	// test success
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.True(t, res.IsOK())
	assert.Len(t, res.Events, 1)
	assert.Equal(t, types.EventTypeCreateTradingPair, res.Events[0].Type)
	assert.Len(t, res.Events[0].Attributes, 3)
	assert.Equal(t, []byte(types.AttributeKeyDexID), res.Events[0].Attributes[0].Key)
	assert.Equal(t, []byte("1"), res.Events[0].Attributes[0].Value)
	assert.Equal(t, []byte(types.AttributeKeyTokenA), res.Events[0].Attributes[1].Key)
	assert.Equal(t, []byte("eth"), res.Events[0].Attributes[1].Value)
	assert.Equal(t, []byte(types.AttributeKeyTokenB), res.Events[0].Attributes[2].Key)
	assert.Equal(t, []byte("btc"), res.Events[0].Attributes[2].Value)
	expectedTradingPair := &types.TradingPair{
		DexID:             1,
		TokenA:            "eth",
		TokenB:            "btc",
		TokenAAmount:      sdk.ZeroInt(),
		TokenBAmount:      sdk.ZeroInt(),
		TotalLiquidity:    sdk.ZeroInt(),
		IsPublic:          true,
		LPRewardRate:      lpRewardRate,
		RefererRewardRate: refererRewardRate,
	}
	assert.Equal(t, expectedTradingPair, k.GetTradingPair(ctx, 1, "btc", "eth"))

	// test trading pair already exists
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", false, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "trading pair already exists in dex 1")

	originAmount := sdk.NewInt(100000000)
	lpAddr := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, lpAddr, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))

	// test create public trading pair after public pair is created
	addMsg := types.NewMsgAddLiquidity(lpAddr, 0, "btc", "usdt", sdk.NewInt(1000), sdk.NewInt(10000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "usdt", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.True(t, res.IsOK())
	expectedTradingPair.TokenA = "btc"
	expectedTradingPair.TokenB = "usdt"
	assert.Equal(t, expectedTradingPair, k.GetTradingPair(ctx, 1, "btc", "usdt"))

	// test create private trading pair after public pair is created
	addMsg = types.NewMsgAddLiquidity(lpAddr, 0, "eth", "usdt", sdk.NewInt(1000), sdk.NewInt(10000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	msg = types.NewMsgCreateTradingPair(dexOwner, 1, "eth", "usdt", false, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, msg)
	assert.True(t, res.IsOK())
	expectedTradingPair.TokenA = "eth"
	expectedTradingPair.IsPublic = false
	assert.Equal(t, expectedTradingPair, k.GetTradingPair(ctx, 1, "eth", "usdt"))
}

func TestHandleMsgEditTradingPair(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.k

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	// test dex not exists
	msg := types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", nil, nil, nil)
	res := handleMsgEditTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "dex id 1 not found")

	createDexMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createDexMsg)
	assert.True(t, res.IsOK())

	// test trading pair not exists
	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", nil, nil, nil)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "trading pair does not exist in dex 1")

	createPairMsg := types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "eth", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	// test not owner
	newCU := sdk.NewCUAddress()
	msg = types.NewMsgEditTradingPair(newCU, 1, "btc", "eth", nil, nil, nil)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, fmt.Sprintf("dex 1 belongs to %s, not %s", dexOwner.String(), newCU.String()))

	// test referer reward too small
	smallRefererRate := types.DefaultRefererTransactionBonusRate.Sub(sdk.SmallestDec())
	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", nil, nil, &smallRefererRate)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "public pair's referer reward rate must be larger than")

	// test sum of fee rate too large
	bigRefererRate := types.DefaultMaxFeeRate.Sub(types.DefaultRepurchaseRate).Sub(types.DefaultLpRewardRate).Add(sdk.SmallestDec())
	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", nil, nil, &bigRefererRate)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "sum of lp reward rate and referer reward rate is too large")

	// test success
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	f := false
	newLpRewardRate := lpRewardRate.Add(sdk.SmallestDec())
	newRefererRate := refererRewardRate.Add(sdk.SmallestDec())
	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", &f, &newLpRewardRate, &newRefererRate)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.True(t, res.IsOK())
	assert.Len(t, res.Events, 1)
	assert.Equal(t, types.EventTypeEditTradingPair, res.Events[0].Type)
	assert.Len(t, res.Events[0].Attributes, 3)
	assert.Equal(t, []byte(types.AttributeKeyDexID), res.Events[0].Attributes[0].Key)
	assert.Equal(t, []byte("1"), res.Events[0].Attributes[0].Value)
	assert.Equal(t, []byte(types.AttributeKeyTokenA), res.Events[0].Attributes[1].Key)
	assert.Equal(t, []byte("eth"), res.Events[0].Attributes[1].Value)
	assert.Equal(t, []byte(types.AttributeKeyTokenB), res.Events[0].Attributes[2].Key)
	assert.Equal(t, []byte("btc"), res.Events[0].Attributes[2].Value)
	expectedTradingPair := &types.TradingPair{
		DexID:             1,
		TokenA:            "eth",
		TokenB:            "btc",
		TokenAAmount:      sdk.ZeroInt(),
		TokenBAmount:      sdk.ZeroInt(),
		TotalLiquidity:    sdk.ZeroInt(),
		IsPublic:          false,
		LPRewardRate:      newLpRewardRate,
		RefererRewardRate: newRefererRate,
	}
	assert.Equal(t, expectedTradingPair, k.GetTradingPair(ctx, 1, "btc", "eth"))

	tr := true
	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", &tr, nil, nil)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.True(t, res.IsOK())

	originAmount := sdk.NewInt(100000000)
	lpAddr := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, lpAddr, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))

	// test set public pair to private after liquidity is provided
	addMsg := types.NewMsgAddLiquidity(lpAddr, 1, "btc", "eth", sdk.NewInt(1000), sdk.NewInt(10000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", &f, nil, nil)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.True(t, res.IsOK())

	// test set private pair to public after liquidity is provided
	addMsg = types.NewMsgAddLiquidity(lpAddr, 1, "btc", "eth", sdk.NewInt(1000), sdk.NewInt(10000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	msg = types.NewMsgEditTradingPair(dexOwner, 1, "btc", "eth", &tr, nil, nil)
	res = handleMsgEditTradingPair(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "cannot set pair public after adding liquidity")
}

func TestHandleMsgAddLiquidityInCustomDex(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	createDexMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res := handleMsgCreateDex(ctx, k, createDexMsg)
	assert.True(t, res.IsOK())

	// test trading pair not exists
	msg := types.NewMsgAddLiquidity(address, 1, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "trading pair does not exist in dex 1")

	createPairMsg := types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "usdt", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	// test success in public pair
	msg = types.NewMsgAddLiquidity(address, 1, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.True(t, res.IsOK())

	coins := input.trk.GetAllBalance(ctx, address)
	btcRemain := coins.AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(20000), btcRemain)
	usdtRemain := coins.AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(8000000), usdtRemain)

	expectedPair := &types.TradingPair{
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      sdk.NewInt(20000),
		TokenBAmount:      sdk.NewInt(8000000),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      sdk.ZeroDec(),
		RefererRewardRate: sdk.ZeroDec(),
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 0, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, address, 0, "btc", "usdt"))

	expectedPair = &types.TradingPair{
		DexID:             1,
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          true,
		TokenAAmount:      sdk.ZeroInt(),
		TokenBAmount:      sdk.ZeroInt(),
		TotalLiquidity:    sdk.ZeroInt(),
		LPRewardRate:      lpRewardRate,
		RefererRewardRate: refererRewardRate,
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 1, "btc", "usdt"))
	assert.Equal(t, sdk.ZeroInt(), k.GetLiquidity(ctx, address, 1, "btc", "usdt"))

	// test success in private pair
	createPairMsg = types.NewMsgCreateTradingPair(dexOwner, 1, "eth", "usdt", false, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	msg = types.NewMsgAddLiquidity(address, 1, "eth", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, msg)
	assert.True(t, res.IsOK())

	coins = input.trk.GetAllBalance(ctx, address)
	assert.Equal(t, originAmount.SubRaw(20000), coins.AmountOf("eth"))
	assert.Equal(t, originAmount.SubRaw(16000000), coins.AmountOf("usdt"))

	assert.Nil(t, k.GetTradingPair(ctx, 0, "eth", "usdt"))
	assert.Equal(t, sdk.ZeroInt(), k.GetLiquidity(ctx, address, 0, "eth", "usdt"))

	expectedPair = &types.TradingPair{
		DexID:             1,
		TokenA:            "eth",
		TokenB:            "usdt",
		IsPublic:          false,
		TokenAAmount:      sdk.NewInt(20000),
		TokenBAmount:      sdk.NewInt(8000000),
		TotalLiquidity:    sdk.NewInt(400000),
		LPRewardRate:      lpRewardRate,
		RefererRewardRate: refererRewardRate,
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 1, "eth", "usdt"))
	assert.Equal(t, sdk.NewInt(400000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, address, 1, "eth", "usdt"))
}

func TestHandleMsgRemoveLiquidityInCustomDex(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(10000), sdk.NewInt(4000000), 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	createDexMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createDexMsg)
	assert.True(t, res.IsOK())

	// test trading pair not exists
	msg := types.NewMsgRemoveLiquidity(address, 1, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "trading pair does not exist in dex 1")

	createPairMsg := types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "usdt", false, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	addMsg = types.NewMsgAddLiquidity(address, 1, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	coins := input.trk.GetAllBalance(ctx, address)
	btcRemain := coins.AmountOf("btc")
	assert.Equal(t, originAmount.SubRaw(30000), btcRemain)
	usdtRemain := coins.AmountOf("usdt")
	assert.Equal(t, originAmount.SubRaw(12000000), usdtRemain)

	msg = types.NewMsgRemoveLiquidity(address, 1, "btc", "usdt", sdk.NewInt(30000), 999999999999)
	res = handleMsgRemoveLiquidity(ctx, k, msg) // 赎回 0.075*20000 btc  0.075*8000000 usdt
	assert.True(t, res.IsOK())

	coins = input.trk.GetAllBalance(ctx, address)
	assert.Equal(t, btcRemain.AddRaw(1500), coins.AmountOf("btc"))
	assert.Equal(t, usdtRemain.AddRaw(600000), coins.AmountOf("usdt"))

	expectedPair := &types.TradingPair{
		DexID:             1,
		TokenA:            "btc",
		TokenB:            "usdt",
		IsPublic:          false,
		TokenAAmount:      sdk.NewInt(18500),   // 20000 - 1500
		TokenBAmount:      sdk.NewInt(7400000), // 8000000 - 600000
		TotalLiquidity:    sdk.NewInt(370000),  // 400000 - 30000
		LPRewardRate:      lpRewardRate,
		RefererRewardRate: refererRewardRate,
	}
	assert.Equal(t, expectedPair, k.GetTradingPair(ctx, 1, "btc", "usdt"))
	assert.Equal(t, sdk.NewInt(400000-30000).Sub(types.DefaultMinimumLiquidity), k.GetLiquidity(ctx, address, 1, "btc", "usdt"))
}

func TestSwapRefererReward(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(10000), sdk.NewInt(4000000), 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))

	amtIn := sdk.NewInt(40000)
	// referer reward to self
	msg := types.NewMsgSwapExactIn(0, buyer, buyer, buyer, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())
	coins := input.trk.GetAllBalance(ctx, buyer)
	refererReward := amtIn.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt()
	assert.Equal(t, originAmount.Sub(amtIn).Add(refererReward), coins.AmountOf("usdt"))

	// referer reward to other
	referer := sdk.NewCUAddress()
	beforeBalance := input.trk.GetAllBalance(ctx, buyer).AmountOf("usdt")
	msg = types.NewMsgSwapExactIn(0, buyer, referer, buyer, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())
	coins = input.trk.GetAllBalance(ctx, referer)
	assert.Equal(t, refererReward, coins.AmountOf("usdt"))
	coins = input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, beforeBalance.Sub(amtIn), coins.AmountOf("usdt"))

	// referer reward to other
	beforeBalance = input.trk.GetAllBalance(ctx, buyer).AmountOf("usdt")
	beforeRefererBalance := input.trk.GetAllBalance(ctx, referer).AmountOf("usdt")
	msg = types.NewMsgSwapExactIn(0, buyer, buyer, buyer, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())
	coins = input.trk.GetAllBalance(ctx, referer)
	assert.Equal(t, beforeRefererBalance.Add(refererReward), coins.AmountOf("usdt"))
	coins = input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, beforeBalance.Sub(amtIn), coins.AmountOf("usdt"))

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	createDexMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createDexMsg)
	assert.True(t, res.IsOK())

	createPairMsg := types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "usdt", false, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	addMsg = types.NewMsgAddLiquidity(address, 1, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(4000000), 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	// referer reward to dex income receiver
	beforeBalance = input.trk.GetAllBalance(ctx, buyer).AmountOf("usdt")
	msg = types.NewMsgSwapExactIn(1, buyer, buyer, buyer, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())
	refererReward = amtIn.ToDec().Mul(refererRewardRate).TruncateInt()
	coins = input.trk.GetAllBalance(ctx, incomeReceiver)
	assert.Equal(t, refererReward, coins.AmountOf("usdt"))
	coins = input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, beforeBalance.Sub(amtIn), coins.AmountOf("usdt"))
}

func TestSwapLpReward(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	lpAddr := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, lpAddr, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	addAmountBtc := sdk.NewInt(10000)
	addAmountUsdt := sdk.NewInt(4000000)
	addMsg := types.NewMsgAddLiquidity(lpAddr, 0, "btc", "usdt", addAmountBtc, addAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	product := addAmountBtc.Mul(addAmountUsdt)
	totalLiquidity := sdk.NewIntFromBigInt(big.NewInt(0).Sqrt(product.BigInt()))
	liquidityLp1 := totalLiquidity.Sub(types.DefaultMinimumLiquidity)

	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))

	amtIn := sdk.NewInt(4000000)
	msg := types.NewMsgSwapExactIn(0, buyer, buyer, buyer, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "btc"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())

	lpReward := amtIn.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	fees := amtIn.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		amtIn.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt()).Add(
		amtIn.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt())
	realIn := amtIn.Sub(fees)
	btcOut := realIn.Mul(addAmountBtc).Quo(realIn.Add(addAmountUsdt))

	btcUsdtAmountBtc := addAmountBtc.Sub(btcOut)
	btcUsdtAmountUsdt := addAmountUsdt.Add(realIn).Add(lpReward)

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	createDexMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createDexMsg)
	assert.True(t, res.IsOK())

	createPairMsg := types.NewMsgCreateTradingPair(dexOwner, 1, "btc", "usdt", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	lpAddr2 := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, lpAddr2, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	addAmountBtc2 := sdk.NewInt(10000)
	addAmountUsdt2 := sdk.NewInt(4000000)
	addMsg = types.NewMsgAddLiquidity(lpAddr2, 1, "btc", "usdt", addAmountBtc2, addAmountUsdt2, 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	needUsdt := addAmountBtc2.Mul(btcUsdtAmountUsdt).Quo(btcUsdtAmountBtc)
	needBtc := addAmountBtc2
	if needUsdt.GT(addAmountUsdt2) {
		needUsdt = addAmountUsdt2
		needBtc = addAmountUsdt2.Mul(btcUsdtAmountBtc).Quo(btcUsdtAmountUsdt)
	}
	liquidityLp2 := sdk.MinInt(needUsdt.Mul(totalLiquidity).Quo(btcUsdtAmountUsdt), needBtc.Mul(totalLiquidity).Quo(btcUsdtAmountBtc))
	btcUsdtAmountBtc = btcUsdtAmountBtc.Add(needBtc)
	btcUsdtAmountUsdt = btcUsdtAmountUsdt.Add(needUsdt)
	totalLiquidity = totalLiquidity.Add(liquidityLp2)

	btcAmtIn := sdk.NewInt(5000)
	msg = types.NewMsgSwapExactIn(1, buyer, buyer, buyer, btcAmtIn, sdk.ZeroInt(), []sdk.Symbol{"btc", "usdt"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())
	btcLpReward := btcAmtIn.ToDec().Mul(types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	fees = btcAmtIn.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		btcAmtIn.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt()).Add(
		btcAmtIn.ToDec().Mul(refererRewardRate).TruncateInt())
	realIn = btcAmtIn.Sub(fees)
	usdtOut := realIn.Mul(btcUsdtAmountUsdt).Quo(realIn.Add(btcUsdtAmountBtc))

	btcUsdtAmountBtc = btcUsdtAmountBtc.Add(realIn).Add(btcLpReward)
	btcUsdtAmountUsdt = btcUsdtAmountUsdt.Sub(usdtOut)

	pair := k.GetTradingPair(ctx, 0, "btc", "usdt")
	assert.Equal(t, totalLiquidity, pair.TotalLiquidity)
	assert.Equal(t, btcUsdtAmountBtc, pair.TokenAAmount)
	assert.Equal(t, btcUsdtAmountUsdt, pair.TokenBAmount)
	assert.Equal(t, liquidityLp1, k.GetLiquidity(ctx, lpAddr, 0, "btc", "usdt"))
	assert.Equal(t, liquidityLp2, k.GetLiquidity(ctx, lpAddr2, 0, "btc", "usdt"))

	removeMsg := types.NewMsgRemoveLiquidity(lpAddr, 1, "btc", "usdt", liquidityLp1, 0)
	res = handleMsgRemoveLiquidity(ctx, k, removeMsg)
	assert.True(t, res.IsOK())
	btcReturn := liquidityLp1.Mul(btcUsdtAmountBtc).Quo(totalLiquidity)
	usdtReturn := liquidityLp1.Mul(btcUsdtAmountUsdt).Quo(totalLiquidity)
	coins := input.trk.GetAllBalance(ctx, lpAddr)
	assert.Equal(t, originAmount.Sub(addAmountBtc).Add(btcReturn), coins.AmountOf("btc"))
	assert.Equal(t, originAmount.Sub(addAmountUsdt).Add(usdtReturn), coins.AmountOf("usdt"))

	removeMsg = types.NewMsgRemoveLiquidity(lpAddr2, 0, "btc", "usdt", liquidityLp2, 0)
	res = handleMsgRemoveLiquidity(ctx, k, removeMsg)
	assert.True(t, res.IsOK())
	btcReturn = liquidityLp2.Mul(btcUsdtAmountBtc).Quo(totalLiquidity)
	usdtReturn = liquidityLp2.Mul(btcUsdtAmountUsdt).Quo(totalLiquidity)
	coins = input.trk.GetAllBalance(ctx, lpAddr2)
	assert.Equal(t, originAmount.Sub(needBtc).Add(btcReturn), coins.AmountOf("btc"))
	assert.Equal(t, originAmount.Sub(needUsdt).Add(usdtReturn), coins.AmountOf("usdt"))

	// test lp reward in private trading pair
	createPairMsg = types.NewMsgCreateTradingPair(dexOwner, 1, "eth", "usdt", false, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	lpAddr3 := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, lpAddr3, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	addAmountEth := sdk.NewInt(10000)
	addAmountUsdt = sdk.NewInt(4000000)
	addMsg = types.NewMsgAddLiquidity(lpAddr3, 1, "eth", "usdt", addAmountEth, addAmountUsdt, 999999999999)
	res = handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())
	product = addAmountEth.Mul(addAmountUsdt)
	totalLiquidity = sdk.NewIntFromBigInt(big.NewInt(0).Sqrt(product.BigInt()))
	liquidityLp3 := totalLiquidity.Sub(types.DefaultMinimumLiquidity)

	amtIn = sdk.NewInt(4000000)
	msg = types.NewMsgSwapExactIn(1, buyer, buyer, buyer, amtIn, sdk.ZeroInt(), []sdk.Symbol{"usdt", "eth"}, 999999999999)
	res = handleMsgSwapExactIn(ctx, k, msg)
	assert.True(t, res.IsOK())

	lpReward = amtIn.ToDec().Mul(lpRewardRate.Add(types.DefaultRepurchaseRate)).TruncateInt()
	fees = amtIn.ToDec().Mul(lpRewardRate).TruncateInt().Add(
		amtIn.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt()).Add(
		amtIn.ToDec().Mul(refererRewardRate).TruncateInt())
	realIn = amtIn.Sub(fees)
	ethOut := realIn.Mul(addAmountEth).Quo(realIn.Add(addAmountUsdt))

	btcUsdtAmountEth := addAmountEth.Sub(ethOut)
	btcUsdtAmountUsdt = addAmountUsdt.Add(realIn).Add(lpReward)

	pair = k.GetTradingPair(ctx, 1, "eth", "usdt")
	assert.Equal(t, totalLiquidity, pair.TotalLiquidity)
	assert.Equal(t, btcUsdtAmountEth, pair.TokenAAmount)
	assert.Equal(t, btcUsdtAmountUsdt, pair.TokenBAmount)
	assert.Equal(t, liquidityLp3, k.GetLiquidity(ctx, lpAddr3, 1, "eth", "usdt"))

	removeMsg = types.NewMsgRemoveLiquidity(lpAddr3, 1, "eth", "usdt", liquidityLp3, 0)
	res = handleMsgRemoveLiquidity(ctx, k, removeMsg)
	assert.True(t, res.IsOK())
	ethReturn := liquidityLp3.Mul(btcUsdtAmountEth).Quo(totalLiquidity)
	usdtReturn = liquidityLp3.Mul(btcUsdtAmountUsdt).Quo(totalLiquidity)
	coins = input.trk.GetAllBalance(ctx, lpAddr3)
	assert.Equal(t, originAmount.Sub(addAmountEth).Add(ethReturn), coins.AmountOf("eth"))
	assert.Equal(t, originAmount.Sub(addAmountUsdt).Add(usdtReturn), coins.AmountOf("usdt"))
}

func TestHandleMsgLimitSwapFail(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("usdt", originAmount)))
	k := input.k

	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", sdk.NewInt(20000), sdk.NewInt(8000000), 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()
	// test token not exists
	msg := types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"fakebtc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token fakebtc does not exist")

	// test token not sendable
	btc := input.tk.GetIBCToken(ctx, "btc")
	btc.SendEnabled = false
	input.tk.SetToken(ctx, btc)
	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeUnsupportToken, res.Code)
	assert.Contains(t, res.Log, "token btc is not enable to send")

	btc.SendEnabled = true
	input.tk.SetToken(ctx, btc)
	ctx = ctx.WithBlockTime(time.Unix(1000, 0))
	// test expired tx
	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"btc", "usdt", 0, 1)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "expired tx")

	// test insufficient funds
	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, originAmount.AddRaw(1), sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInsufficientCoins, res.Code)
	assert.Contains(t, res.Log, "balance not enough")

	// test order amount too small
	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, sdk.NewInt(2), sdk.NewDecWithPrec(101, 2),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "limit order amount is too small")

	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, sdk.NewInt(2), sdk.NewDecWithPrec(99, 2),
		"btc", "usdt", 1, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "limit order amount is too small")

	// test dex not exists
	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 1, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "dex id 1 not found")

	dexOwner := sdk.NewCUAddress()
	dexName := "test"
	incomeReceiver := sdk.NewCUAddress()
	lpRewardRate := sdk.NewDecWithPrec(1, 2)
	refererRewardRate := sdk.NewDecWithPrec(2, 2)

	createDexMsg := types.NewMsgCreateDex(dexOwner, dexName, incomeReceiver)
	res = handleMsgCreateDex(ctx, k, createDexMsg)
	assert.True(t, res.IsOK())

	// test no trading pair
	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 0, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"eth", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "eth-usdt trading pair does not exist in dex 0")

	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 1, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "btc-usdt trading pair does not exist in dex 1")

	// test liquidity not enough
	createPairMsg := types.NewMsgCreateTradingPair(dexOwner, 1, "eth", "usdt", true, lpRewardRate, refererRewardRate)
	res = handleMsgCreateTradingPair(ctx, k, createPairMsg)
	assert.True(t, res.IsOK())

	msg = types.NewMsgLimitSwap(uuid.NewV4().String(), 1, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"eth", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "eth-usdt trading pair does not have enough liquidity")
}

func TestHandleMsgLimitSwapExactInSuccess(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	btcUsdtAmountBtc := sdk.NewInt(20000)
	btcUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", btcUsdtAmountBtc, btcUsdtAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	receiver := sdk.NewCUAddress()
	referer := sdk.NewCUAddress()
	orderID := uuid.NewV4().String()
	amtIn := sdk.NewInt(40000)
	msg := types.NewMsgLimitSwap(orderID, 0, buyer, referer, receiver, amtIn, sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.True(t, res.IsOK(), res.Log)

	expectOrder := &types.Order{
		OrderID:     orderID,
		CreatedTime: ctx.BlockTime().Unix(),
		ExpiredTime: 999999999999,
		From:        buyer,
		Referer:     referer,
		Receiver:    receiver,
		Price:       sdk.OneDec(),
		Side:        0,
		BaseSymbol:  "btc",
		QuoteSymbol: "usdt",
		AmountIn:    amtIn,
		LockedFund:  amtIn,
		FeeRate:     types.NewFeeRate(types.DefaultLpRewardRate, types.DefaultRepurchaseRate, types.DefaultRefererTransactionBonusRate),
	}
	assert.Equal(t, expectOrder, k.GetOrder(ctx, orderID))
	coins := input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, originAmount.Sub(amtIn), coins.AmountOf("usdt"))

	// test order already exists
	msg = types.NewMsgLimitSwap(orderID, 0, buyer, referer, receiver, sdk.NewInt(40000), sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, fmt.Sprintf("order %s already exists", orderID))

	// test partially filled
	// buy order
	orderID = uuid.NewV4().String()
	amtIn = sdk.NewInt(4000000)
	price := sdk.NewDec(500)
	beforeBuyerBalances := input.trk.GetAllBalance(ctx, buyer)
	msg = types.NewMsgLimitSwap(orderID, 0, buyer, referer, receiver, amtIn, price, "btc", "usdt", types.OrderSideBuy, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.True(t, res.IsOK(), res.Log)
	realIn := price.Mul(btcUsdtAmountBtc.ToDec()).TruncateInt().Sub(btcUsdtAmountUsdt)
	feeRate := types.DefaultLpRewardRate.Add(types.DefaultRepurchaseRate).Add(types.DefaultRefererTransactionBonusRate)
	cost := realIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt()
	lpReward := cost.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		cost.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt())
	fees := cost.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		cost.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt()).Add(
		cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt())
	realIn = cost.Sub(fees)
	amtOut := realIn.Mul(btcUsdtAmountBtc).Quo(realIn.Add(btcUsdtAmountUsdt))
	dealPrice := realIn.ToDec().Quo(amtOut.ToDec())
	assert.True(t, dealPrice.Sub(price).Quo(price).Abs().LTE(sdk.NewDecWithPrec(1, 3)))

	expectOrder = &types.Order{
		OrderID:     orderID,
		CreatedTime: ctx.BlockTime().Unix(),
		ExpiredTime: 999999999999,
		From:        buyer,
		Referer:     referer,
		Receiver:    receiver,
		Price:       price,
		Side:        types.OrderSideBuy,
		BaseSymbol:  "btc",
		QuoteSymbol: "usdt",
		AmountIn:    amtIn,
		LockedFund:  amtIn.Sub(cost),
		FeeRate:     types.NewFeeRate(types.DefaultLpRewardRate, types.DefaultRepurchaseRate, types.DefaultRefererTransactionBonusRate),
		Status:      types.OrderStatusPartiallyFilled,
	}
	assert.Equal(t, expectOrder, k.GetOrder(ctx, orderID))
	afterBuyerBalances := input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, amtIn, beforeBuyerBalances.AmountOf("usdt").Sub(afterBuyerBalances.AmountOf("usdt")))
	assert.Equal(t, cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt(), input.trk.GetAllBalance(ctx, referer).AmountOf("usdt"))
	assert.Equal(t, amtOut, input.trk.GetAllBalance(ctx, receiver).AmountOf("btc"))

	btcUsdtAmountBtc = btcUsdtAmountBtc.Sub(amtOut)
	btcUsdtAmountUsdt = btcUsdtAmountUsdt.Add(realIn).Add(lpReward)
	pair := k.GetTradingPair(ctx, 0, "btc", "usdt")
	assert.Equal(t, btcUsdtAmountBtc, pair.TokenAAmount)
	assert.Equal(t, btcUsdtAmountUsdt, pair.TokenBAmount)

	// sell order
	orderID = uuid.NewV4().String()
	amtIn = sdk.NewInt(20000)
	price = sdk.NewDec(400)
	beforeBuyerBalances = afterBuyerBalances
	msg = types.NewMsgLimitSwap(orderID, 0, buyer, referer, receiver, amtIn, price, "btc", "usdt", types.OrderSideSell, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.True(t, res.IsOK(), res.Log)
	realIn = btcUsdtAmountUsdt.ToDec().Quo(price).TruncateInt().Sub(btcUsdtAmountBtc)
	cost = realIn.ToDec().Quo(sdk.OneDec().Sub(feeRate)).TruncateInt()
	lpReward = cost.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		cost.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt())
	fees = cost.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		cost.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt()).Add(
		cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt())
	realIn = cost.Sub(fees)
	amtOut = realIn.Mul(btcUsdtAmountUsdt).Quo(realIn.Add(btcUsdtAmountBtc))
	dealPrice = amtOut.ToDec().Quo(realIn.ToDec())
	assert.True(t, dealPrice.Sub(price).Quo(price).Abs().LTE(sdk.NewDecWithPrec(1, 3)))

	expectOrder = &types.Order{
		OrderID:     orderID,
		CreatedTime: ctx.BlockTime().Unix(),
		ExpiredTime: 999999999999,
		From:        buyer,
		Referer:     referer,
		Receiver:    receiver,
		Price:       price,
		Side:        types.OrderSideSell,
		BaseSymbol:  "btc",
		QuoteSymbol: "usdt",
		AmountIn:    amtIn,
		LockedFund:  amtIn.Sub(cost),
		FeeRate:     types.NewFeeRate(types.DefaultLpRewardRate, types.DefaultRepurchaseRate, types.DefaultRefererTransactionBonusRate),
		Status:      types.OrderStatusPartiallyFilled,
	}
	assert.Equal(t, expectOrder, k.GetOrder(ctx, orderID))
	afterBuyerBalances = input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, amtIn, beforeBuyerBalances.AmountOf("btc").Sub(afterBuyerBalances.AmountOf("btc")))
	assert.True(t, cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt().Equal(input.trk.GetAllBalance(ctx, referer).AmountOf("btc")))
	assert.Equal(t, amtOut, input.trk.GetAllBalance(ctx, receiver).AmountOf("usdt"))

	btcUsdtAmountBtc = btcUsdtAmountBtc.Add(realIn).Add(lpReward)
	btcUsdtAmountUsdt = btcUsdtAmountUsdt.Sub(amtOut)
	pair = k.GetTradingPair(ctx, 0, "btc", "usdt")
	assert.Equal(t, btcUsdtAmountBtc, pair.TokenAAmount)
	assert.Equal(t, btcUsdtAmountUsdt, pair.TokenBAmount)

	// test full filled
	orderID = uuid.NewV4().String()
	amtIn = sdk.NewInt(4000000)
	price = sdk.NewDec(10000)
	beforeBuyerBalances = afterBuyerBalances
	beforeRefererBalances := input.trk.GetAllBalance(ctx, referer)
	beforeReceiverBalances := input.trk.GetAllBalance(ctx, receiver)
	msg = types.NewMsgLimitSwap(orderID, 0, buyer, referer, receiver, amtIn, price, "btc", "usdt", types.OrderSideBuy, 999999999999)
	res = handleMsgLimitSwap(ctx, k, msg)
	assert.True(t, res.IsOK(), res.Log)
	cost = amtIn
	lpReward = cost.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		cost.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt())
	fees = cost.ToDec().Mul(types.DefaultLpRewardRate).TruncateInt().Add(
		cost.ToDec().Mul(types.DefaultRepurchaseRate).TruncateInt()).Add(
		cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt())
	realIn = cost.Sub(fees)
	amtOut = realIn.Mul(btcUsdtAmountBtc).Quo(realIn.Add(btcUsdtAmountUsdt))
	dealPrice = realIn.ToDec().Quo(amtOut.ToDec())
	assert.True(t, dealPrice.LT(price))

	expectOrder = &types.Order{
		OrderID:      orderID,
		CreatedTime:  ctx.BlockTime().Unix(),
		ExpiredTime:  999999999999,
		FinishedTime: ctx.BlockTime().Unix(),
		From:         buyer,
		Referer:      referer,
		Receiver:     receiver,
		Price:        price,
		Side:         types.OrderSideBuy,
		BaseSymbol:   "btc",
		QuoteSymbol:  "usdt",
		AmountIn:     amtIn,
		LockedFund:   sdk.ZeroInt(),
		FeeRate:      types.NewFeeRate(types.DefaultLpRewardRate, types.DefaultRepurchaseRate, types.DefaultRefererTransactionBonusRate),
		Status:       types.OrderStatusFilled,
	}
	assert.Equal(t, expectOrder, k.GetOrder(ctx, orderID))
	afterBuyerBalances = input.trk.GetAllBalance(ctx, buyer)
	assert.Equal(t, amtIn, beforeBuyerBalances.AmountOf("usdt").Sub(afterBuyerBalances.AmountOf("usdt")))
	assert.Equal(t, cost.ToDec().Mul(types.DefaultRefererTransactionBonusRate).TruncateInt(),
		input.trk.GetAllBalance(ctx, referer).AmountOf("usdt").Sub(beforeRefererBalances.AmountOf("usdt")))
	assert.Equal(t, amtOut, input.trk.GetAllBalance(ctx, receiver).AmountOf("btc").Sub(beforeReceiverBalances.AmountOf("btc")))

	btcUsdtAmountBtc = btcUsdtAmountBtc.Sub(amtOut)
	btcUsdtAmountUsdt = btcUsdtAmountUsdt.Add(realIn).Add(lpReward)
	pair = k.GetTradingPair(ctx, 0, "btc", "usdt")
	assert.Equal(t, btcUsdtAmountBtc, pair.TokenAAmount)
	assert.Equal(t, btcUsdtAmountUsdt, pair.TokenBAmount)
}

func TestHandleMsgCancelLimit(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	address := sdk.NewCUAddress()
	originAmount := sdk.NewInt(100000000)
	input.trk.AddCoins(ctx, address, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	buyer := sdk.NewCUAddress()
	input.trk.AddCoins(ctx, buyer, sdk.NewCoins(sdk.NewCoin("btc", originAmount), sdk.NewCoin("eth", originAmount),
		sdk.NewCoin("usdt", originAmount)))
	k := input.k

	btcUsdtAmountBtc := sdk.NewInt(20000)
	btcUsdtAmountUsdt := sdk.NewInt(8000000)
	addMsg := types.NewMsgAddLiquidity(address, 0, "btc", "usdt", btcUsdtAmountBtc, btcUsdtAmountUsdt, 999999999999)
	res := handleMsgAddLiquidity(ctx, k, addMsg)
	assert.True(t, res.IsOK())

	orderID := uuid.NewV4().String()
	// test order not exists
	msg := types.NewMsgCancelLimitSwap(buyer, []string{orderID})
	res = handleMsgCancelLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeNotFoundOrder, res.Code)
	assert.Contains(t, res.Log, fmt.Sprintf("order %s not found", orderID))

	amtIn := sdk.NewInt(40000)
	limitMsg := types.NewMsgLimitSwap(orderID, 0, buyer, buyer, buyer, amtIn, sdk.OneDec(),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, limitMsg)
	assert.True(t, res.IsOK(), res.Log)

	// test not order owner
	msg = types.NewMsgCancelLimitSwap(sdk.NewCUAddress(), []string{orderID})
	res = handleMsgCancelLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, "no permission to cancel order")

	// test success
	beforeTime := ctx.BlockTime().Unix()
	now := time.Now()
	ctx = ctx.WithBlockTime(now)
	beforeBalances := input.trk.GetAllBalance(ctx, buyer)
	msg = types.NewMsgCancelLimitSwap(buyer, []string{orderID})
	res = handleMsgCancelLimitSwap(ctx, k, msg)
	assert.True(t, res.IsOK())
	expectOrder := &types.Order{
		OrderID:      orderID,
		CreatedTime:  beforeTime,
		ExpiredTime:  999999999999,
		FinishedTime: now.Unix(),
		From:         buyer,
		Referer:      buyer,
		Receiver:     buyer,
		Price:        sdk.OneDec(),
		Side:         types.OrderSideBuy,
		BaseSymbol:   "btc",
		QuoteSymbol:  "usdt",
		AmountIn:     amtIn,
		LockedFund:   sdk.ZeroInt(),
		FeeRate:      types.NewFeeRate(types.DefaultLpRewardRate, types.DefaultRepurchaseRate, types.DefaultRefererTransactionBonusRate),
		Status:       types.OrderStatusCanceled,
	}
	assert.Equal(t, expectOrder, k.GetOrder(ctx, orderID))
	assert.Equal(t, amtIn, input.trk.GetAllBalance(ctx, buyer).AmountOf("usdt").Sub(beforeBalances.AmountOf("usdt")))

	// test order has finished
	orderID2 := uuid.NewV4().String()
	limitMsg = types.NewMsgLimitSwap(orderID2, 0, buyer, buyer, buyer, sdk.NewInt(10000), sdk.NewDec(1000),
		"btc", "usdt", 0, 999999999999)
	res = handleMsgLimitSwap(ctx, k, limitMsg)
	assert.True(t, res.IsOK(), res.Log)
	msg = types.NewMsgCancelLimitSwap(buyer, []string{orderID2})
	res = handleMsgCancelLimitSwap(ctx, k, msg)
	assert.Equal(t, sdk.CodeInvalidTx, res.Code)
	assert.Contains(t, res.Log, fmt.Sprintf("order %s has been finished, cannot be canceled", orderID2))

}
