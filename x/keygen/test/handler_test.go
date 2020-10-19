package test

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/chainnode"
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/distribution"
	"github.com/hbtc-chain/bhchain/x/keygen"
	"github.com/hbtc-chain/bhchain/x/keygen/types"
	"github.com/hbtc-chain/bhchain/x/mint"
	"github.com/hbtc-chain/bhchain/x/order"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/params/subspace"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/staking"
	stakingtypes "github.com/hbtc-chain/bhchain/x/staking/types"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

var (
	ethToken       = "eth"
	btcToken       = "btc"
	usdtToken      = "usdt"
	keygenFromAddr = sdk.NewCUAddress()
	validatorPriv1 = ed25519.GenPrivKey()
	validatorPriv2 = ed25519.GenPrivKey()
	validatorAddr1 = sdk.CUAddressFromPubKey(validatorPriv1.PubKey())
	validatorAddr2 = sdk.CUAddressFromPubKey(validatorPriv2.PubKey())
)

func TestHandleMsgNewOpCU(t *testing.T) {
	input := SetupTestInput()
	ctx := input.Ctx
	cuKeeper := input.Ck.(custodianunit.CUKeeper)
	keygenkeeper := input.Kk
	from := sdk.NewCUAddress()
	opcuAddress := sdk.NewCUAddress()

	// token not support
	msg := types.NewMsgNewOpCU("NOSupport", opcuAddress, from)
	res := keygen.HandleMsgNewOpCUForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// OP CU address has been used
	cuKeeper.SetCU(ctx, cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, opcuAddress))
	msg = types.NewMsgNewOpCU(ethToken, opcuAddress, from)
	res = keygen.HandleMsgNewOpCUForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// too many OP CU
	for i := 0; uint64(i) < cuKeeper.GetTokenKeeper(ctx).GetMaxOpCUNumber(ctx, sdk.Symbol(ethToken)); i++ {
		msg = types.NewMsgNewOpCU(ethToken, sdk.NewCUAddress(), from)
		keygen.HandleMsgNewOpCUForTest(ctx, keygenkeeper, msg)
		assert.False(t, res.IsOK())
	}
	msg = types.NewMsgNewOpCU(ethToken, sdk.NewCUAddress(), from)
	res = keygen.HandleMsgNewOpCUForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// from not validator
	msg = types.NewMsgNewOpCU(btcToken, sdk.NewCUAddress(), from)
	res = keygen.HandleMsgNewOpCUForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// should ok
	msg = types.NewMsgNewOpCU(btcToken, sdk.NewCUAddress(), validatorAddr1)
	res = keygen.HandleMsgNewOpCUForTest(ctx, keygenkeeper, msg)
	assert.True(t, res.IsOK())
}

func TestHandleMsgKeyGen(t *testing.T) {
	input := SetupTestInput()
	ctx := input.Ctx
	cuKeeper := input.Ck.(custodianunit.CUKeeper)
	keygenkeeper := input.Kk

	pubkey := ed25519.GenPrivKey().PubKey()
	toAddr := sdk.CUAddress(pubkey.Address())
	toOpCUAddr := sdk.NewCUAddress()
	fromCU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, keygenFromAddr)
	toCU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, toAddr)
	validator := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, validatorAddr1)
	toOpCU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeOp, toOpCUAddr)
	toOpCU.AddAsset(ethToken, "", 0)
	toNoExist := sdk.NewCUAddress()
	ethSymbol := sdk.Symbol(ethToken)
	cuKeeper.SetCU(ctx, fromCU)
	cuKeeper.SetCU(ctx, toCU)
	cuKeeper.SetCU(ctx, toOpCU)
	cuKeeper.SetCU(ctx, validator)
	input.ChainNode.On("SupportChain", ethToken).Return(true)
	input.ChainNode.On("SupportChain", "NOSupport").Return(false)

	tcs := []struct {
		OrderID string
		Symbol  sdk.Symbol
		From    sdk.CUAddress
		To      sdk.CUAddress
		Ok      bool
	}{
		{Ok: false, From: keygenFromAddr, To: toAddr, OrderID: uuid.NewV4().String(), Symbol: "NOSupport"},
		{Ok: false, From: keygenFromAddr, To: toAddr, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		{Ok: true, From: keygenFromAddr, To: toAddr, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		{Ok: true, From: keygenFromAddr, To: toNoExist, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		{Ok: true, From: keygenFromAddr, To: keygenFromAddr, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		{Ok: false, From: keygenFromAddr, To: toOpCUAddr, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		{Ok: true, From: validatorAddr1, To: toOpCUAddr, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		// 7 pubkey exist
		{Ok: true, From: validatorAddr1, To: toAddr, OrderID: uuid.NewV4().String(), Symbol: ethSymbol},
		{Ok: true, From: validatorAddr1, To: toAddr, OrderID: uuid.NewV4().String(), Symbol: sdk.Symbol(usdtToken)},
		{Ok: true, From: validatorAddr1, To: toAddr, OrderID: uuid.NewV4().String(), Symbol: sdk.Symbol(usdtToken)},
	}

	caseNo := 0 // symbol not support
	msg := types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	res := keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)

	caseNo = 1 // keygenFromAddr cu insufficient gas
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)

	caseNo = 2 // should ok
	fromCU.AddCoins(sdk.Coins{sdk.NewCoin(sdk.NativeToken, sdk.NewIntWithDecimal(1, 20))})
	cuKeeper.SetCU(ctx, fromCU)
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin := cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold := cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	fromCU = cuKeeper.GetCU(ctx, keygenFromAddr)
	// check fromcu.coinshold
	openFee := input.Tk.GetTokenInfo(ctx, ethSymbol).OpenFee
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check order & flows
	orderGot := input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
	if tcs[caseNo].Ok {
		checkKeygenOrder(t, input, orderGot, msg)
		checkKeygenFlow(t, input, res, msg)
	}

	caseNo = 3 // should ok, to cu not exist
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	// check to cu should exist
	assert.NotNil(t, cuKeeper.GetCU(ctx, toNoExist))
	openFee = input.Tk.GetTokenInfo(ctx, ethSymbol).OpenFee
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check order & flows
	if tcs[caseNo].Ok {
		orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
		checkKeygenOrder(t, input, orderGot, msg)
		checkKeygenFlow(t, input, res, msg)
	}

	caseNo = 4 // should ok, to = from cu
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	// check fromcu.coinshold
	openFee = input.Tk.GetTokenInfo(ctx, ethSymbol).OpenFee
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check order & flows
	if tcs[caseNo].Ok {
		orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
		checkKeygenOrder(t, input, orderGot, msg)
		checkKeygenFlow(t, input, res, msg)
	}

	caseNo = 5 // create address for OpCU ,should fail,from must be validator
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	openFee = sdk.ZeroInt()
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check order & flows
	if tcs[caseNo].Ok {
		orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
		checkKeygenOrder(t, input, orderGot, msg)
		checkKeygenFlow(t, input, res, msg)
	}

	caseNo = 6 // create address for OpCU,should ok, from must be validator
	validator = cuKeeper.GetCU(ctx, validatorAddr1)
	validator.AddCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewIntWithDecimal(1, 22))))
	cuKeeper.SetCU(ctx, validator)
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	// check fromcu.coinshold
	openFee = input.Tk.GetTokenInfo(ctx, ethSymbol).SysOpenFee
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check order & flows
	if tcs[caseNo].Ok {
		orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
		checkKeygenOrder(t, input, orderGot, msg)
		checkKeygenFlow(t, input, res, msg)
	}

	caseNo = 7 // should ok, pubkey exist
	input.ChainNode.On("ConvertAddressFromSerializedPubKey", "eth", pubkey.Bytes()).Return(pubkey.Address().String(), nil)
	tocu := cuKeeper.GetCU(ctx, tcs[caseNo].To)
	tocu.SetAssetPubkey(pubkey.Bytes(), 1)
	cuKeeper.SetCU(ctx, tocu)
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	// to has processing order
	assert.Equal(t, false, res.IsOK(), "case No:", caseNo)
	openFee = sdk.ZeroInt()
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	ol := input.Ok.GetProcessOrderListByType(input.Ctx, sdk.OrderTypeKeyGen)
	for _, orderid := range ol {
		input.Ok.RemoveProcessOrder(input.Ctx, sdk.OrderTypeKeyGen, orderid)
	}

	caseNo = 7 // should ok, pubkey exist
	input.ChainNode.On("ConvertAddressFromSerializedPubKey", "eth", pubkey.Bytes()).Return(pubkey.Address().String(), nil)
	tocu = cuKeeper.GetCU(ctx, tcs[caseNo].To)
	tocu.SetAssetPubkey(pubkey.Bytes(), 1)
	cuKeeper.SetCU(ctx, tocu)
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	openFee = sdk.ZeroInt()
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check the asset address created by chainode
	assert.EqualValues(t, pubkey.Address().String(), cuKeeper.GetCU(ctx, tcs[caseNo].To).GetAssetAddress(tcs[caseNo].Symbol.String(), 1))
	receip, err := input.Rk.GetReceiptFromResult(&res)
	assert.NotNil(t, err)
	// no flows
	assert.Nil(t, receip)
	orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
	// no order
	assert.Nil(t, orderGot)

	caseNo = 8 // should ok, keygen for subtoken ,the main net token  address exist
	input.ChainNode.On("SupportChain", usdtToken).Return(true)
	// eth address is from case7
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, tcs[caseNo].To)
	tocu.SetAssetPubkey(pubkey.Bytes(), 1)
	ethAddr := "ethaddr"
	tocu.SetAssetAddress("eth", ethAddr, 1)
	cuKeeper.SetCU(ctx, tocu)
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	openFee = sdk.ZeroInt()
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken).Sub(openFee)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	// check the asset address created by chainode
	assert.EqualValues(t, ethAddr, cuKeeper.GetCU(ctx, tcs[caseNo].To).GetAssetAddress(tcs[caseNo].Symbol.String(), 1))
	receip, err = input.Rk.GetReceiptFromResult(&res)
	assert.NotNil(t, err)
	// no flows
	assert.Nil(t, receip)
	orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
	// no order
	assert.Nil(t, orderGot)

	caseNo = 9 // should ok, use prekeygen order
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, tcs[caseNo].To)
	cuKeeper.SetCU(ctx, toCU)
	preKeyGenOrderID := uuid.NewV4().String()
	preKeyGenOrder := &sdk.OrderKeyGen{
		OrderBase: sdk.OrderBase{
			CUAddress: keygenFromAddr,
			ID:        preKeyGenOrderID,
			OrderType: sdk.OrderTypeKeyGen,
			Status:    sdk.OrderStatusSignFinish,
		},
		Pubkey: pubkey.Bytes(),
	}
	input.Ok.SetOrder(ctx, preKeyGenOrder)
	keygenkeeper.AddWaitAssignKeyGenOrderID(ctx, preKeyGenOrderID)
	input.ChainNode.On("ConvertAddressFromSerializedPubKey", "eth", preKeyGenOrder.Pubkey).Return(pubkey.Address().String(), nil)
	msg = types.NewMsgKeyGen(tcs[caseNo].OrderID, tcs[caseNo].Symbol, tcs[caseNo].From, tcs[caseNo].To)
	originCoin = cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken)
	originHold = cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenForTest(ctx, keygenkeeper, msg)
	assert.Equal(t, tcs[caseNo].Ok, res.IsOK(), "case No:", caseNo)
	openFee = input.Tk.GetTokenInfo(ctx, ethSymbol).OpenFee
	// 只扣费，不冻结
	assert.True(t, originHold.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoinsHold().AmountOf(sdk.NativeToken)))
	assert.True(t, originCoin.Equal(cuKeeper.GetCU(ctx, msg.From).GetCoins().AmountOf(sdk.NativeToken).Add(openFee)))
	assert.EqualValues(t, pubkey.Address().String(), cuKeeper.GetCU(ctx, tcs[caseNo].To).GetAssetAddress(tcs[caseNo].Symbol.String(), 1))
	orderGot = input.Ok.GetOrder(ctx, tcs[caseNo].OrderID)
	assert.Nil(t, orderGot)
	order := input.Ok.GetOrder(ctx, preKeyGenOrderID)
	assert.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
}

func checkKeygenOrder(t *testing.T, input testInput, orderGot sdk.Order, msg types.MsgKeyGen) {
	keyGenOrder, ok := orderGot.(*sdk.OrderKeyGen)
	assert.True(t, ok)
	assert.NotNil(t, keyGenOrder)
	assert.Equal(t, sdk.OrderStatusBegin, orderGot.GetOrderStatus())
	assert.Equal(t, sdk.OrderTypeKeyGen, orderGot.GetOrderType())
	assert.Equal(t, msg.Symbol.String(), orderGot.GetSymbol())
	assert.EqualValues(t, msg.From, orderGot.GetCUAddress())
	assert.EqualValues(t, msg.OrderID, orderGot.GetID())
	assert.EqualValues(t, len(input.Sk.GetAllValidators(input.Ctx)), len(keyGenOrder.KeyNodes))
	if input.Ck.GetCU(input.Ctx, msg.To).GetCUType() == sdk.CUTypeOp {
		assert.EqualValues(t, input.Tk.GetSysOpenFee(input.Ctx, msg.Symbol), keyGenOrder.OpenFee.AmountOf(sdk.NativeToken))
	} else {
		assert.EqualValues(t, input.Tk.GetOpenFee(input.Ctx, msg.Symbol), keyGenOrder.OpenFee.AmountOf(sdk.NativeToken))
	}
	assert.EqualValues(t, sdk.Majority23(len(input.Sk.GetAllValidators(input.Ctx))), keyGenOrder.SignThreshold)

}

func checkKeygenFlow(t *testing.T, input testInput, res sdk.Result, msg types.MsgKeyGen) {
	receip, err := input.Rk.GetReceiptFromResult(&res)
	assert.Nil(t, err)
	assert.NotNil(t, receip)
	assert.Equal(t, sdk.CategoryTypeKeyGen, receip.Category)
	assert.True(t, len(receip.Flows) > 2)
	orderFlow, ok := receip.Flows[0].(sdk.OrderFlow)
	assert.True(t, ok)
	assert.Equal(t, msg.To, orderFlow.CUAddress)
	assert.Equal(t, msg.Symbol, orderFlow.Symbol)
	assert.Equal(t, sdk.OrderTypeKeyGen, orderFlow.OrderType)
	assert.Equal(t, msg.OrderID, orderFlow.OrderID)
	assert.Equal(t, sdk.OrderStatusBegin, orderFlow.OrderStatus)

	keygenFlow, ok := receip.Flows[1].(sdk.KeyGenFlow)
	assert.True(t, ok)
	assert.Equal(t, msg.From, keygenFlow.From)
	assert.Equal(t, msg.OrderID, keygenFlow.OrderID)
	assert.Equal(t, msg.To, keygenFlow.To)
	assert.Equal(t, msg.Symbol, keygenFlow.Symbol)

	balanceFlow, ok := receip.Flows[2].(sdk.BalanceFlow)
	assert.True(t, ok)
	assert.Equal(t, msg.From, balanceFlow.CUAddress)
	assert.Equal(t, sdk.NativeToken, balanceFlow.Symbol.String())

}

func checkKeygenEvent(t *testing.T, input testInput, res sdk.Result, msg types.MsgKeyGen) {
	event := res.Events[0]
	assert.Equal(t, types.EventTypeKeyGen, event.Type)
	assert.Greater(t, len(event.Attributes), 3)
}

func TestHandleMsgKeyGenFinish(t *testing.T) {
	input := SetupTestInput()
	ctx := input.Ctx
	cuKeeper := input.Ck.(custodianunit.CUKeeper)
	keygenkeeper := input.Kk
	keygenFinishFromAddr := sdk.NewCUAddress()
	pubkey := ed25519.GenPrivKey().PubKey()
	to := sdk.CUAddress(pubkey.Address())

	input.ChainNode.On("SupportChain", "NOSupport").Return(false)

	// case1 token not support
	msg := types.NewMsgKeyGenWaitSign(keygenFinishFromAddr, uuid.NewV4().String(), pubkey.Bytes(), nil, nil, 1)
	res := keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())
	// case2 token not pupport by chainnode
	input.ChainNode.On("SupportChain", ethToken).Return(false).Once()
	msg = types.NewMsgKeyGenWaitSign(keygenFinishFromAddr, uuid.NewV4().String(), pubkey.Bytes(), nil, nil, 1)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())
	input.ChainNode.On("SupportChain", ethToken).Return(true)

	// case3 order not exist
	ethSymbol := sdk.Symbol(ethToken)
	msg = types.NewMsgKeyGenWaitSign(keygenFinishFromAddr, uuid.NewV4().String(), pubkey.Bytes(), nil, nil, 1)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// setup msg & keygenorder & validators
	validator1 := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, validatorAddr1)
	validator2 := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, validatorAddr2)
	cuKeeper.SetCU(ctx, validator1)
	cuKeeper.SetCU(ctx, validator2)
	msg = types.NewMsgKeyGenWaitSign(keygenFinishFromAddr, uuid.NewV4().String(), pubkey.Bytes(),
		[]sdk.CUAddress{validatorAddr1, validatorAddr2}, nil, 1)
	keygenorder := newTestKeyGenOrder(input, msg, to)
	input.Ok.SetOrder(input.Ctx, keygenorder)
	keygenFromCU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, keygenFromAddr)
	keygenFromCU.AddCoinsHold(sdk.NewCoins(sdk.Coin{sdk.NativeToken, input.Tk.GetTokenInfo(ctx, ethSymbol).OpenFee}))
	cuKeeper.SetCU(ctx, keygenFromCU)

	// case7 toCU is opcu,symbol != msg.symbol
	toCU := cuKeeper.NewOpCUWithAddress(ctx, "btc", to)
	cuKeeper.SetCU(ctx, toCU)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case8 toCU is opcu, keygenFinishFromAddr is not validator
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewOpCUWithAddress(ctx, "eth", to)
	cuKeeper.SetCU(ctx, toCU)
	fromCU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, keygenFinishFromAddr)
	cuKeeper.SetCU(ctx, fromCU)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case9 toCU is opcu, msg.keygenFinishFromAddr is validator,keygenorder.keygenFinishFromAddr is not validator
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewOpCUWithAddress(ctx, "eth", to)
	cuKeeper.SetCU(ctx, toCU)
	msg.From = validatorAddr1
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())
	msg.From = keygenFinishFromAddr

	// case10 toCU is opcu, msg.keygenFinishFromAddr is not validator,keygenorder.keygenFinishFromAddr is validator
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewOpCUWithAddress(ctx, "eth", to)
	cuKeeper.SetCU(ctx, toCU)
	keygenorder.CUAddress = validatorAddr1
	input.Ok.SetOrder(input.Ctx, keygenorder)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case11 toCU have asset address, keygenFinishFromAddr is validator
	cuKeeper.RemoveCU(ctx, toCU)
	msg.From = validatorAddr1
	keygenorder.CUAddress = validatorAddr1
	input.Ok.SetOrder(input.Ctx, keygenorder)
	toCU = cuKeeper.NewOpCUWithAddress(ctx, "eth", to)
	toCU.SetAssetAddress(ethToken, "0xc6452b4a3", 1)
	cuKeeper.SetCU(ctx, toCU)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())
	keygenorder.CUAddress = keygenFromAddr
	input.Ok.SetOrder(input.Ctx, keygenorder)

	// case12 chainnode ConvertAddress != msg.address
	msg.PubKey = pubkey.Bytes()
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewOpCUWithAddress(ctx, "eth", to)
	cuKeeper.SetCU(ctx, toCU)
	input.ChainNode.On("ConvertAddressFromSerializedPubKey", "eth", msg.PubKey).Return(pubkey.Address().String(), nil)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case13 msg no signature . verify sign failed
	fromCU.AddCoinsHold(sdk.NewCoins(sdk.Coin{sdk.NativeToken, input.Tk.GetTokenInfo(ctx, ethSymbol).OpenFee}))
	cuKeeper.SetCU(ctx, fromCU)
	msg.PubKey = pubkey.Bytes()
	msg.From = keygenFinishFromAddr
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, to)
	cuKeeper.SetCU(ctx, toCU)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case14 msg signature not enough. verify sign failed
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, to)
	cuKeeper.SetCU(ctx, toCU)
	sigs := []cutypes.StdSignature{}
	signmsg := types.NewMsgKeyGenWaitSign(msg.From, msg.OrderID, msg.PubKey, msg.KeyNodes, sigs, 1)
	sig2, _ := validatorPriv2.Sign(signmsg.GetSignBytes())
	msg.KeySigs = []cutypes.StdSignature{cutypes.StdSignature{Signature: sig2, PubKey: validatorPriv2.PubKey()}}
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case15 msg signature mismatch. verify sign failed
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, to)
	cuKeeper.SetCU(ctx, toCU)
	sigs = []cutypes.StdSignature{}
	signmsg = types.NewMsgKeyGenWaitSign(msg.From, msg.OrderID, msg.PubKey, msg.KeyNodes, sigs, 1)
	sig1, err := validatorPriv1.Sign(signmsg.GetSignBytes())
	assert.Nil(t, err)
	// sig2 is mismatch
	sig2, _ = validatorPriv2.Sign(msg.GetSignBytes())
	msg.KeySigs = []cutypes.StdSignature{cutypes.StdSignature{Signature: sig1, PubKey: validatorPriv1.PubKey()},
		cutypes.StdSignature{Signature: sig2, PubKey: validatorPriv2.PubKey()}}
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.False(t, res.IsOK())

	// case 16 ok ,to.cutype == user cu
	cuKeeper.RemoveCU(ctx, toCU)
	toCU = cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, to)
	cuKeeper.SetCU(ctx, toCU)
	sigs = []cutypes.StdSignature{}
	signmsg = types.NewMsgKeyGenWaitSign(msg.From, msg.OrderID, msg.PubKey, msg.KeyNodes, sigs, 1)
	sig1, err = validatorPriv1.Sign(signmsg.GetSignBytes())
	assert.Nil(t, err)
	sig2, _ = validatorPriv2.Sign(signmsg.GetSignBytes())
	msg.KeySigs = []cutypes.StdSignature{cutypes.StdSignature{Signature: sig1, PubKey: validatorPriv1.PubKey()},
		cutypes.StdSignature{Signature: sig2, PubKey: validatorPriv2.PubKey()}}
	originHold := cuKeeper.GetCU(ctx, keygenFromAddr).GetCoinsHold().AmountOf(sdk.NativeToken)
	res = keygen.HandleMsgKeyGenWaitSignForTest(ctx, keygenkeeper, msg)
	assert.True(t, res.IsOK())
	// check order closed
	orderGot := input.Ok.GetOrder(ctx, msg.OrderID)
	assert.Equal(t, sdk.OrderStatusWaitSign, orderGot.GetOrderStatus())
	// check keygenFinishFromAddr.coinshold (set in case 13)
	cu := cuKeeper.GetCU(ctx, keygenFromAddr)

	assert.True(t, originHold.Equal(cu.GetCoinsHold().AmountOf(sdk.NativeToken)))
}

func newTestKeyGenOrder(input testInput, msg types.MsgKeyGenWaitSign, to sdk.CUAddress) *sdk.OrderKeyGen {
	ordBase := sdk.OrderBase{
		CUAddress: keygenFromAddr,
		ID:        msg.OrderID,
		OrderType: sdk.OrderTypeKeyGen,
		Status:    sdk.OrderStatusBegin,
		Symbol:    ethToken,
	}

	tokenInfo := input.Tk.GetTokenInfo(input.Ctx, sdk.Symbol(ethToken))
	orderKeyGen := sdk.OrderKeyGen{
		OrderBase:        ordBase,
		KeyNodes:         msg.KeyNodes,
		SignThreshold:    3,
		To:               to,
		MultiSignAddress: "",
		OpenFee:          sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, tokenInfo.OpenFee)),
	}

	order := input.Ok.NewOrder(input.Ctx, &orderKeyGen)
	return (order).(*sdk.OrderKeyGen)

}

func SetupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	cdc.RegisterInterface((*exported.CustodianUnit)(nil), nil)
	cdc.RegisterConcrete(&custodianunit.BaseCU{}, "hbtcchain/cu/basecu", nil)
	receipt.RegisterCodec(cdc)
	order.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	cuKey := sdk.NewKVStoreKey("cuKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")
	tokenKey := sdk.NewKVStoreKey(token.ModuleName)
	keygenKey := sdk.NewKVStoreKey(keygen.ModuleName)
	orderkey := sdk.NewKVStoreKey("order")
	stakingKey := sdk.NewKVStoreKey(staking.StoreKey)
	stakingKeyT := sdk.NewTransientStoreKey(staking.TStoreKey)
	Supplykey := sdk.NewKVStoreKey(supply.StoreKey)
	transferKey := sdk.NewKVStoreKey(transfer.StoreKey)
	distrKey := sdk.NewKVStoreKey(distribution.StoreKey)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(cuKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(tokenKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keygenKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(orderkey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(Supplykey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKeyT, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(distrKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	rk := receipt.NewKeeper(cdc)
	tk := token.NewKeeper(tokenKey, cdc, subspace.NewSubspace(cdc, keyParams, tkeyParams, token.DefaultParamspace))
	ck := custodianunit.NewCUKeeper(cdc, cuKey, &tk, ps, cutypes.ProtoBaseCU)
	ok := order.NewKeeper(cdc, orderkey, subspace.NewSubspace(cdc, keyParams, tkeyParams, order.DefaultParamspace))
	chainnode := new(chainnode.MockChainnode)
	transferK := transfer.NewBaseKeeper(cdc, transferKey, ck, &tk, &ok, rk, nil, chainnode, pk.Subspace(transfer.DefaultParamspace), transfer.DefaultCodespace, nil)

	maccPerms := map[string][]string{
		custodianunit.FeeCollectorName: nil,
		distribution.ModuleName:        nil,
		stakingtypes.NotBondedPoolName: []string{supply.Burner, supply.Staking},
		stakingtypes.BondedPoolName:    []string{supply.Burner, supply.Staking},
		mint.ModuleName:                []string{supply.Minter},
		types.ModuleName:               {supply.Minter, supply.Burner},
	}
	supplyKeeper := supply.NewKeeper(cdc, Supplykey, ck, transferK, maccPerms)

	initTokens := sdk.TokensFromConsensusPower(100000)
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(int64(2))))

	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	feeCollectorAcc := supply.NewEmptyModuleAccount(custodianunit.FeeCollectorName)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)
	distrAcc := supply.NewEmptyModuleAccount(distribution.ModuleName)
	supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	supplyKeeper.SetModuleAccount(ctx, distrAcc)
	supplyKeeper.SetModuleAccount(ctx, bondPool)
	supplyKeeper.SetModuleAccount(ctx, notBondedPool)

	stakingSubspace := subspace.NewSubspace(cdc, keyParams, tkeyParams, staking.DefaultParamspace)
	stakingK := staking.NewKeeper(cdc, stakingKey, stakingKeyT, supplyKeeper, stakingSubspace, staking.DefaultCodespace)

	distrKeeper := distribution.NewKeeper(cdc, distrKey, pk.Subspace(distribution.DefaultParamspace), stakingK, supplyKeeper, distribution.DefaultCodespace,
		custodianunit.FeeCollectorName, nil)

	kk := keygen.NewKeeper(keygenKey, cdc, &tk, ck, &ok, rk, stakingK, distrKeeper, chainnode)

	ck.SetParams(ctx, cutypes.DefaultParams())
	//init token info
	for _, tokenInfo := range token.TestTokenData {
		token := token.NewTokenInfo(tokenInfo.Symbol, tokenInfo.Chain, tokenInfo.Issuer, tokenInfo.TokenType,
			tokenInfo.IsSendEnabled, tokenInfo.IsDepositEnabled, tokenInfo.IsWithdrawalEnabled, tokenInfo.Decimals,
			tokenInfo.TotalSupply, tokenInfo.CollectThreshold, tokenInfo.DepositThreshold, tokenInfo.OpenFee,
			tokenInfo.SysOpenFee, tokenInfo.WithdrawalFeeRate, tokenInfo.SysTransferNum, tokenInfo.OpCUSysTransferNum,
			tokenInfo.GasLimit, tokenInfo.GasPrice, tokenInfo.MaxOpCUNumber, tokenInfo.Confirmations, tokenInfo.IsNonceBased) //WithdrawalAddress and depositAddress will be added later.
		tk.SetTokenInfo(ctx, token)
	}

	//init staking info，set validators
	val1 := stakingtypes.NewValidator(sdk.ValAddress(validatorAddr1), ed25519.GenPrivKey().PubKey(), stakingtypes.Description{}, true)
	val2 := stakingtypes.NewValidator(sdk.ValAddress(validatorAddr2), ed25519.GenPrivKey().PubKey(), stakingtypes.Description{}, true)
	stakingK.SetValidator(ctx, val1)
	stakingK.SetValidator(ctx, val2)
	vals := []sdk.CUAddress{}
	vals = append(vals, validatorAddr1)
	vals = append(vals, validatorAddr2)
	stakingK.StartNewEpoch(ctx, vals)
	ctx = ctx.WithBlockHeight(1)

	feePool := distribution.InitialFeePool()
	feePool.CommunityPool = sdk.DecCoins{}
	distrKeeper.SetFeePool(ctx, feePool)

	return testInput{Cdc: cdc, Ctx: ctx, Ck: ck, Kk: kk, Tk: tk, Ok: ok, Rk: *rk, Sk: stakingK, Dk: distrKeeper, ChainNode: chainnode}
}

type testInput struct {
	Cdc       *codec.Codec
	Ctx       sdk.Context
	Kk        keygen.Keeper
	Tk        token.Keeper
	Ok        order.Keeper
	Rk        receipt.Keeper
	Ck        custodianunit.CUKeeperI
	Sk        staking.Keeper
	Dk        distribution.Keeper
	ChainNode *chainnode.MockChainnode
}
