package tests

import (
	"errors"
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/hbtc-chain/bhchain/x/transfer/types"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token"
)

func TestSysTransferEthToUserHaveEthSuccess(t *testing.T) {
	t.Skip("currently systransfer does not support chain's main token")
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.EthToken
	chain := token.EthToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.SysTransferNum = sdk.NewInt(100)
	tokenInfo.GasLimit = sdk.NewInt(100)
	tokenInfo.GasPrice = sdk.NewInt(100)
	tk.SetTokenInfo(ctx, tokenInfo)

	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	opCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	opCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = opCU.SetAssetAddress(symbol, opCUEthAddress, 1)
	require.Nil(t, err)

	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetCoinsHold().AmountOf(symbol))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	user1CUEthAddress := "0x81b7E08F65Bdf5648606c89998A9CC8164397647"

	amt := sdk.NewInt(100)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = user1CU.AddAsset(symbol, user1CUEthAddress, 1)
	ck.SetCU(ctx, user1CU)

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, user1CUEthAddress, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	mockCN.On("ValidAddress", chain, symbol, user1CUEthAddress).Return(true, user1CUEthAddress)
	mockCN.On("ValidAddress", chain, symbol, opCUEthAddress).Return(true, opCUEthAddress)

	//Step1, SysTransfer
	ethSysTranasferOrderID := uuid.NewV1().String()
	ctx = ctx.WithBlockHeight(11)
	orderID := ethSysTranasferOrderID
	sysTransferAmt := tokenInfo.SysTransferAmount()
	//	gasFee := sdk.NewInt(1300000)
	result := keeper.SysTransfer(ctx, opCU.GetAddress(), user1CUAddr, user1CUEthAddress, orderID, symbol)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	order := ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, sdk.ZeroInt(), order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, user1CUEthAddress, order.(*sdk.OrderSysTransfer).ToAddress)
	require.Equal(t, opCU.GetAddress().String(), order.(*sdk.OrderSysTransfer).OpCUaddress)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of.CUAddress)

	wf, valid := receipt.Flows[1].(sdk.SysTransferFlow)
	require.True(t, valid)
	require.Equal(t, opCU.GetAddress().String(), wf.FromCU)
	require.Equal(t, user1CU.GetAddress().String(), wf.ToCU)
	require.Equal(t, user1CUEthAddress, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, sysTransferAmt, wf.Amount)
	require.Equal(t, orderID, wf.OrderID)

	//Check opCU coins and coinsHold
	opCU2 := ck.GetCU(ctx, opCU.GetAddress())
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt), opCU2.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sysTransferAmt, opCU2.GetAssetCoinsHold().AmountOf(symbol))
	sendable := opCU2.IsEnabledSendTx(chain, opCU2.GetAssetAddress(chain, 1))
	require.Equal(t, false, sendable)

	//Step2, SysTransferWaitSign
	sysTransferTxHash := "sysTransferTxHash"
	chainnodeSysTransferlTx := chainnode.ExtAccountTransaction{
		Hash:     sysTransferTxHash,
		From:     opCUEthAddress,
		To:       user1CUEthAddress,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(100),
		GasPrice: sdk.NewInt(100), // 10Gwei
	}
	suggestGasFee := sdk.NewInt(100).MulRaw(100)

	rawData := []byte("rawData")
	signHash := "signHash"
	mockCN.On("ValidAddress", chain, symbol, user1CUEthAddress).Return(true, user1CUEthAddress)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeSysTransferlTx, []byte(signHash), nil)

	ctx = ctx.WithBlockHeight(20)
	result1 := keeper.SysTransferWaitSign(ctx, orderID, signHash, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, suggestGasFee, order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, user1CUEthAddress, order.(*sdk.OrderSysTransfer).ToAddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, valid := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of1.CUAddress)

	wwf, valid := receipt1.Flows[1].(sdk.SysTransferWaitSignFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wwf.OrderID)
	require.Equal(t, rawData, wwf.RawData)

	//check OPCU's coins, coinsHold and status
	opCU3 := ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), opCU3.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), opCU3.GetAssetCoins().AmountOf(symbol))

	//Step3, SysTransferSignFinish
	signedData := []byte("signedData")
	chainnodeSysTransferlTx.BlockHeight = 10000 //eth height = 10000

	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCUEthAddress, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result2 := keeper.SysTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt2.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of2, valid := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of2.CUAddress)

	wsf, valid := receipt2.Flows[1].(sdk.SysTransferSignFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wsf.OrderID)
	require.Equal(t, signedData, wsf.SignedTx)

	//Check user1 coins and coinsHold
	user1CU1 := ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasReceived())

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCU.GetAssetAddress(chain, 1))
	require.Equal(t, false, sendable)
	require.Equal(t, sdk.Coins(nil), opCU.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), opCU.GetGasReceived())

	//Step4, SysTransferFinish
	costFee := sdk.NewInt(8000)
	chainnodeSysTransferlTx.Status = chainnode.StatusSuccess
	chainnodeSysTransferlTx.CostFee = costFee
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result3 := keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)
	require.Equal(t, costFee, order.(*sdk.OrderSysTransfer).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt3.Category)
	require.Equal(t, 2, len(receipt3.Flows))

	of3, valid := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of3.CUAddress)

	sff, valid := receipt3.Flows[1].(sdk.SysTransferFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, sff.OrderID)
	require.Equal(t, costFee, sff.CostFee)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, sdk.NewInt(100), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU1.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sysTransferAmt, user1CU1.GetGasReceived().AmountOf(symbol))

	//check systransfer deposit item
	item := ck.GetDeposit(ctx, symbol, user1CUAddr, sysTransferTxHash, 0)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, sysTransferAmt, item.Amount)

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(costFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCU.GetAssetAddress(chain, 1))
	require.Equal(t, true, sendable)
	require.Equal(t, sysTransferAmt.Add(costFee), opCU.GetGasUsed().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasReceived().AmountOf(chain))
}

func TestSysTransferEthToUserHaveUsdtSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	ethSymbol := token.EthToken
	usdtSymbol := token.UsdtToken
	chain := token.EthToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup eth token info
	ethTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	ethTokenInfo.SysTransferNum = sdk.NewInt(100)
	ethTokenInfo.GasLimit = sdk.NewInt(100)
	ethTokenInfo.GasPrice = sdk.NewInt(100)
	tk.SetTokenInfo(ctx, ethTokenInfo)

	//setup usdt token info
	usdtTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.UsdtToken))
	usdtTokenInfo.SysTransferNum = sdk.NewInt(3)
	usdtTokenInfo.GasLimit = sdk.NewInt(100)
	usdtTokenInfo.GasPrice = sdk.NewInt(800)
	tk.SetTokenInfo(ctx, usdtTokenInfo)

	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	opCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	opCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = opCU.SetAssetAddress(ethSymbol, opCUEthAddress, 1)
	require.Nil(t, err)

	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(ethSymbol, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU.GetAssetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoinsHold().AmountOf(ethSymbol))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	user1CUEthAddress := "0x81b7E08F65Bdf5648606c89998A9CC8164397647"

	user1EthAmt := sdk.NewInt(1)
	user1UsdtAmt := sdk.NewInt(100)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(usdtSymbol, user1UsdtAmt)))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(ethSymbol, user1EthAmt)))
	err = user1CU.AddAsset(ethSymbol, user1CUEthAddress, 1)
	err = user1CU.AddAsset(usdtSymbol, user1CUEthAddress, 1)
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, user1CU)

	require.Equal(t, user1UsdtAmt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(usdtSymbol))
	require.Equal(t, user1EthAmt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(ethSymbol))
	require.Equal(t, user1CUEthAddress, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(ethSymbol, 1))
	require.Equal(t, user1CUEthAddress, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(usdtSymbol, 1))

	// set collect order
	collectOrder := ok.NewOrderCollect(ctx, user1CUAddr, uuid.NewV1().String(), "usdt",
		user1CUAddr, user1CUEthAddress, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), "", 0, "")
	ok.SetOrder(ctx, collectOrder)

	//Step1, SysTransfer
	sysTranasferOrderID := uuid.NewV1().String()
	ctx = ctx.WithBlockHeight(11)
	orderID := sysTranasferOrderID
	sysTransferAmt := usdtTokenInfo.SysTransferAmount()
	mockCN.On("ValidAddress", chain, usdtSymbol, user1CUEthAddress).Return(true, user1CUEthAddress).Once()
	// gasFee := sdk.NewInt(1300000)
	result := keeper.SysTransfer(ctx, opCU.GetAddress(), user1CUAddr, user1CUEthAddress, orderID, usdtSymbol)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	order := ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, sdk.ZeroInt(), order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, user1CUEthAddress, order.(*sdk.OrderSysTransfer).ToAddress)
	require.Equal(t, opCU.GetAddress().String(), order.(*sdk.OrderSysTransfer).OpCUaddress)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of.CUAddress)

	wf, valid := receipt.Flows[1].(sdk.SysTransferFlow)
	require.True(t, valid)
	require.Equal(t, opCU.GetAddress().String(), wf.FromCU)
	require.Equal(t, user1CU.GetAddress().String(), wf.ToCU)
	require.Equal(t, user1CUEthAddress, wf.ToAddr)
	require.Equal(t, usdtSymbol, wf.Symbol)
	require.Equal(t, sysTransferAmt, wf.Amount)
	require.Equal(t, orderID, wf.OrderID)

	//Check opCU coins and coinsHold
	opCU2 := ck.GetCU(ctx, opCU.GetAddress())
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt), opCU2.GetAssetCoins().AmountOf(ethSymbol))
	require.Equal(t, sysTransferAmt, opCU2.GetAssetCoinsHold().AmountOf(ethSymbol))
	sendable := opCU2.IsEnabledSendTx(chain, opCU2.GetAssetAddress(chain, 1))
	require.Equal(t, false, sendable)

	//Step2, SysTransferWaitSign
	sysTransferTxHash := "sysTransferTxHash"
	chainnodeSysTransferlTx := chainnode.ExtAccountTransaction{
		Hash:     sysTransferTxHash,
		From:     opCUEthAddress,
		To:       user1CUEthAddress,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(100),
		GasPrice: sdk.NewInt(100), // 10Gwei
	}
	suggestGasFee := sdk.NewInt(100).MulRaw(100)

	rawData := []byte("rawData")
	signHash := "signHash"

	//mockCN.On("ValidAddress", chain, usdtSymbol, user1CUEthAddress).Return(true, true, user1CUEthAddress)
	mockCN.On("ValidAddress", chain, ethSymbol, opCUEthAddress).Return(true, opCUEthAddress)
	mockCN.On("ValidAddress", chain, ethSymbol, user1CUEthAddress).Return(true, user1CUEthAddress)
	mockCN.On("QueryAccountTransactionFromData", chain, ethSymbol, rawData).Return(&chainnodeSysTransferlTx, []byte(signHash), nil)

	ctx = ctx.WithBlockHeight(20)
	result1 := keeper.SysTransferWaitSign(ctx, orderID, signHash, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, suggestGasFee, order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, user1CUEthAddress, order.(*sdk.OrderSysTransfer).ToAddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, valid := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of1.CUAddress)

	wwf, valid := receipt1.Flows[1].(sdk.SysTransferWaitSignFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wwf.OrderID)
	require.Equal(t, rawData, wwf.RawData)

	//check OPCU's coins, coinsHold and status
	opCU3 := ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), opCU3.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), opCU3.GetAssetCoins().AmountOf(ethSymbol))

	//Step3, SysTransferSignFinish
	signedData := []byte("signedData")
	chainnodeSysTransferlTx.BlockHeight = 10000 //eth height = 10000

	mockCN.On("VerifyAccountSignedTransaction", chain, ethSymbol, opCUEthAddress, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, ethSymbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result2 := keeper.SysTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt2.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of2, valid := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of2.CUAddress)

	wsf, valid := receipt2.Flows[1].(sdk.SysTransferSignFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wsf.OrderID)
	require.Equal(t, signedData, wsf.SignedTx)

	//Check user1 coins and coinsHold
	user1CU1 := ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasReceived())

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), opCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), opCU.GetAssetCoins().AmountOf(ethSymbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUEthAddress)
	require.Equal(t, false, sendable)
	require.Equal(t, sdk.Coins(nil), opCU.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), opCU.GetGasReceived())

	//Step4, SysTransferFinish
	costFee := sdk.NewInt(8000)
	chainnodeSysTransferlTx.Status = chainnode.StatusSuccess
	chainnodeSysTransferlTx.CostFee = costFee
	mockCN.On("QueryAccountTransactionFromSignedData", chain, ethSymbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result3 := keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, opCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)
	require.Equal(t, costFee, order.(*sdk.OrderSysTransfer).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt3.Category)
	require.Equal(t, 2, len(receipt3.Flows))

	of3, valid := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, opCU.GetAddress(), of3.CUAddress)

	sff, valid := receipt3.Flows[1].(sdk.SysTransferFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, sff.OrderID)
	require.Equal(t, costFee, sff.CostFee)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, user1EthAmt, user1CU1.GetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), user1CU1.GetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, sysTransferAmt, user1CU1.GetGasReceived().AmountOf(ethSymbol))

	//check systransfer deposit item
	item := ck.GetDeposit(ctx, ethSymbol, user1CUAddr, sysTransferTxHash, 0)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, sysTransferAmt, item.Amount)

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(costFee), opCU.GetAssetCoins().AmountOf(ethSymbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUEthAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, sysTransferAmt.Add(costFee), opCU.GetGasUsed().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasReceived().AmountOf(chain))
}

func TestSysTransferEthToEthOpCUSuccess(t *testing.T) {
	t.Skip("currently systransfer does not support chain's main token")
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.EthToken
	chain := token.EthToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.OpCUSysTransferNum = sdk.NewInt(200)
	tokenInfo.GasLimit = sdk.NewInt(100)
	tokenInfo.GasPrice = sdk.NewInt(100)
	tk.SetTokenInfo(ctx, tokenInfo)

	//set ethOPCU1
	ethOPCUAddr1, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU1 := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr1))
	opCU1EthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = opCU1.SetAssetAddress(symbol, opCU1EthAddress, 1)
	opCU1.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)

	opCU1EthAmt := sdk.NewInt(90000000)
	item, err := sdk.NewDepositItem("opcu1deposithash", 0, opCU1EthAmt, opCU1EthAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	ck.SaveDeposit(ctx, symbol, ethOPCUAddr1, item)

	opCU1.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCU1EthAmt)))
	ck.SetCU(ctx, opCU1)

	opCU1 = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr1))
	require.Equal(t, opCU1EthAmt, opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetCoinsHold().AmountOf(symbol))

	//setup ethopcu2
	ethOPCUAddr2, err := sdk.CUAddressFromBase58("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx")
	opCU2 := ck.NewOpCUWithAddress(ctx, "eth", ethOPCUAddr2)
	opCU2EthAddress := "0x81b7E08F65Bdf5648606c89998A9CC8164397647"
	err = opCU2.SetAssetAddress(symbol, opCU2EthAddress, 1)
	opCU2.SetAssetPubkey(pubkey.Bytes(), 1)
	opCU2EthAmt := sdk.NewInt(100)
	item, err = sdk.NewDepositItem("opcu2deposithash", 0, opCU2EthAmt, opCU2EthAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	ck.SaveDeposit(ctx, symbol, ethOPCUAddr2, item)
	opCU2.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCU2EthAmt)))
	ck.SetCU(ctx, opCU2)

	opCU2 = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr2))
	require.Equal(t, opCU2EthAmt, opCU2.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetCoinsHold().AmountOf(symbol))

	mockCN.On("ValidAddress", chain, symbol, opCU2EthAddress).Return(true, opCU2EthAddress)
	mockCN.On("ValidAddress", chain, symbol, opCU1EthAddress).Return(true, opCU1EthAddress)

	//Step1, SysTransfer
	ethSysTranasferOrderID := uuid.NewV1().String()
	ctx = ctx.WithBlockHeight(11)
	orderID := ethSysTranasferOrderID
	sysTransferAmt := tokenInfo.OpCUSysTransferAmount()
	//	gasFee := sdk.NewInt(1300000)
	result := keeper.SysTransfer(ctx, ethOPCUAddr1, ethOPCUAddr2, opCU2EthAddress, orderID, symbol)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	order := ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU1.GetAddress(), order.GetCUAddress())
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, sdk.ZeroInt(), order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, opCU2EthAddress, order.(*sdk.OrderSysTransfer).ToAddress)
	require.Equal(t, opCU1.GetAddress().String(), order.(*sdk.OrderSysTransfer).OpCUaddress)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, opCU1.GetAddress(), of.CUAddress)

	wf, valid := receipt.Flows[1].(sdk.SysTransferFlow)
	require.True(t, valid)
	require.Equal(t, opCU1.GetAddress().String(), wf.FromCU)
	require.Equal(t, opCU2.GetAddress().String(), wf.ToCU)
	require.Equal(t, opCU2EthAddress, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, sysTransferAmt, wf.Amount)
	require.Equal(t, orderID, wf.OrderID)

	//Check opCU coins and coinsHold
	opCU1 = ck.GetCU(ctx, opCU1.GetAddress())
	require.Equal(t, opCU1EthAmt.Sub(sysTransferAmt), opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sysTransferAmt, opCU1.GetAssetCoinsHold().AmountOf(symbol))
	sendable := opCU1.IsEnabledSendTx(chain, opCU1EthAddress)
	require.Equal(t, false, sendable)

	//Step2, SysTransferWaitSign
	sysTransferTxHash := "sysTransferTxHash"
	chainnodeSysTransferlTx := chainnode.ExtAccountTransaction{
		Hash:     sysTransferTxHash,
		From:     opCU1EthAddress,
		To:       opCU2EthAddress,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(100),
		GasPrice: sdk.NewInt(100), // 10Gwei
	}
	suggestGasFee := sdk.NewInt(100).MulRaw(100)

	rawData := []byte("rawData")
	signHash := "signHash"
	mockCN.On("ValidAddress", chain, symbol, opCU2EthAddress).Return(true, true)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeSysTransferlTx, []byte(signHash), nil)

	ctx = ctx.WithBlockHeight(20)
	result1 := keeper.SysTransferWaitSign(ctx, orderID, signHash, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU1.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr1.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, suggestGasFee, order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, opCU2EthAddress, order.(*sdk.OrderSysTransfer).ToAddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, valid := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, opCU1.GetAddress(), of1.CUAddress)

	wwf, valid := receipt1.Flows[1].(sdk.SysTransferWaitSignFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wwf.OrderID)
	require.Equal(t, rawData, wwf.RawData)

	//check OPCU's coins, coinsHold and status
	opCU1 = ck.GetCU(ctx, ethOPCUAddr1)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), opCU1.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCU1EthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), opCU1.GetAssetCoins().AmountOf(symbol))

	//Step3, SysTransferSignFinish
	signedData := []byte("signedData")
	chainnodeSysTransferlTx.BlockHeight = 10000 //eth height = 10000

	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCU1EthAddress, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result2 := keeper.SysTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU1.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr1.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt2.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of2, valid := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, opCU1.GetAddress(), of2.CUAddress)

	wsf, valid := receipt2.Flows[1].(sdk.SysTransferSignFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wsf.OrderID)
	require.Equal(t, signedData, wsf.SignedTx)

	//Check user1 coins and coinsHold
	opCU2 = ck.GetCU(ctx, opCU2.GetAddress())
	require.Equal(t, opCU2EthAmt, opCU2.GetAssetCoins().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetAssetCoinsHold().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetGasUsed().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetGasReceived().AmountOf(chain))

	//check OPCU's coins, coinsHold and status
	opCU1 = ck.GetCU(ctx, ethOPCUAddr1)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), opCU1.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCU1EthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), opCU1.GetAssetCoins().AmountOf(symbol))
	sendable = opCU1.IsEnabledSendTx(chain, opCU1EthAddress)
	require.Equal(t, false, sendable)
	require.Equal(t, sdk.Coins(nil), opCU1.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), opCU1.GetGasReceived())

	//Step4, SysTransferFinish
	costFee := sdk.NewInt(8000)
	chainnodeSysTransferlTx.Status = chainnode.StatusSuccess
	chainnodeSysTransferlTx.CostFee = costFee
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result3 := keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, opCU1.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr1.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)
	require.Equal(t, costFee, order.(*sdk.OrderSysTransfer).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt3.Category)
	require.Equal(t, 2, len(receipt3.Flows))

	of3, valid := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, opCU1.GetAddress(), of3.CUAddress)

	sff, valid := receipt3.Flows[1].(sdk.SysTransferFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, sff.OrderID)
	require.Equal(t, costFee, sff.CostFee)

	//Check user1 coins and coinsHold
	opCU2 = ck.GetCU(ctx, ethOPCUAddr2)
	require.Equal(t, opCU2EthAmt.Add(sysTransferAmt), opCU2.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetGasReceived().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU2.GetGasUsed().AmountOf(symbol))

	//check systransfer deposit item
	item = ck.GetDeposit(ctx, symbol, ethOPCUAddr2, sysTransferTxHash, 0)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, sysTransferAmt, item.Amount)

	//check OPCU's coins, coinsHold and status
	opCU1 = ck.GetCU(ctx, ethOPCUAddr1)
	require.Equal(t, sdk.ZeroInt(), opCU1.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCU1EthAmt.Sub(sysTransferAmt).Sub(costFee), opCU1.GetAssetCoins().AmountOf(symbol))
	sendable = opCU1.IsEnabledSendTx(chain, opCU1EthAddress)
	require.Equal(t, sysTransferAmt.Add(costFee), opCU1.GetGasUsed().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU1.GetGasReceived().AmountOf(chain))
}

func TestSysTransferEthToUsdtOpCUSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	ethSymbol := token.EthToken
	usdtSymbol := token.UsdtToken
	chain := token.EthToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup eth token info
	ethTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	ethTokenInfo.OpCUSysTransferNum = sdk.NewInt(20)
	ethTokenInfo.GasLimit = sdk.NewInt(100)
	ethTokenInfo.GasPrice = sdk.NewInt(100)
	tk.SetTokenInfo(ctx, ethTokenInfo)

	//setup usdt token info
	usdtTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.UsdtToken))
	usdtTokenInfo.OpCUSysTransferNum = sdk.NewInt(800)
	usdtTokenInfo.GasLimit = sdk.NewInt(100) //usdt gaslimt must = the gaslimit
	usdtTokenInfo.GasPrice = sdk.NewInt(800)
	tk.SetTokenInfo(ctx, usdtTokenInfo)

	//set ethOPCU
	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	ethOPCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	ethOPCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = ethOPCU.SetAssetAddress(ethSymbol, ethOPCUEthAddress, 1)
	ethOPCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)

	ethOPCUEthAmt := sdk.NewInt(90000000)
	item, err := sdk.NewDepositItem("opcu1deposithash", 0, ethOPCUEthAmt, ethOPCUEthAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	ck.SaveDeposit(ctx, ethSymbol, ethOPCUAddr, item)

	ethOPCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(ethSymbol, ethOPCUEthAmt)))
	ck.SetCU(ctx, ethOPCU)

	ethOPCU = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, ethOPCUEthAmt, ethOPCU.GetAssetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetCoinsHold().AmountOf(ethSymbol))

	//setup usdtOPCU
	usdtOPCUAddr, err := sdk.CUAddressFromBase58("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx")
	usdtOPCU := ck.NewOpCUWithAddress(ctx, "usdt", usdtOPCUAddr)
	usdtOpCUEthAddress := "0x81b7E08F65Bdf5648606c89998A9CC8164397647"
	err = usdtOPCU.SetAssetAddress(ethSymbol, usdtOpCUEthAddress, 1)
	err = usdtOPCU.SetAssetAddress(usdtSymbol, usdtOpCUEthAddress, 1)
	usdtOPCU.SetAssetPubkey(pubkey.Bytes(), 1)
	usdtOPCUEthAmt := sdk.NewInt(10)
	item, err = sdk.NewDepositItem("usdtopcuethdeposithash", 0, usdtOPCUEthAmt, usdtOpCUEthAddress, "", sdk.DepositItemStatusConfirmed)
	ck.SaveDeposit(ctx, ethSymbol, usdtOPCUAddr, item)
	usdtOPCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(ethSymbol, usdtOPCUEthAmt)))

	usdtOPCUUsdtAmt := sdk.NewInt(12345)
	item, err = sdk.NewDepositItem("usdtopcuusdtdeposithash", 0, usdtOPCUUsdtAmt, usdtOpCUEthAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	ck.SaveDeposit(ctx, usdtSymbol, usdtOPCUAddr, item)
	usdtOPCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(usdtSymbol, usdtOPCUUsdtAmt)))
	ck.SetCU(ctx, usdtOPCU)

	usdtOPCU = ck.GetCU(ctx, sdk.CUAddress(usdtOPCUAddr))
	require.Equal(t, usdtOPCUEthAmt, usdtOPCU.GetAssetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoinsHold().AmountOf(ethSymbol))

	require.Equal(t, usdtOPCUUsdtAmt, usdtOPCU.GetAssetCoins().AmountOf(usdtSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetAssetCoinsHold().AmountOf(usdtSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoins().AmountOf(usdtSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoinsHold().AmountOf(usdtSymbol))

	userCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	// set withdraw order
	withdrawOrder := ok.NewOrderWithdrawal(ctx, userCUAddr, uuid.NewV1().String(), "usdt",
		sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), "", "HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx", "")
	withdrawOrder.FromAddress = usdtOpCUEthAddress
	ok.SetOrder(ctx, withdrawOrder)

	//Step1, SysTransfer eth from eth oopcu to usdt opcu
	ethSysTranasferOrderID := uuid.NewV1().String()
	ctx = ctx.WithBlockHeight(11)
	orderID := ethSysTranasferOrderID
	sysTransferAmt := usdtTokenInfo.OpCUSysTransferAmount()
	mockCN.On("ValidAddress", chain, usdtSymbol, usdtOpCUEthAddress).Return(true, usdtOpCUEthAddress).Once()

	//	gasFee := sdk.NewInt(1300000)
	result := keeper.SysTransfer(ctx, ethOPCUAddr, usdtOPCUAddr, usdtOpCUEthAddress, orderID, usdtSymbol)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	order := ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, ethOPCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, sdk.ZeroInt(), order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, usdtOpCUEthAddress, order.(*sdk.OrderSysTransfer).ToAddress)
	require.Equal(t, ethOPCU.GetAddress().String(), order.(*sdk.OrderSysTransfer).OpCUaddress)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, ethOPCU.GetAddress(), of.CUAddress)

	wf, valid := receipt.Flows[1].(sdk.SysTransferFlow)
	require.True(t, valid)
	require.Equal(t, ethOPCU.GetAddress().String(), wf.FromCU)
	require.Equal(t, usdtOPCU.GetAddress().String(), wf.ToCU)
	require.Equal(t, usdtOpCUEthAddress, wf.ToAddr)
	require.Equal(t, usdtSymbol, wf.Symbol)
	require.Equal(t, sysTransferAmt, wf.Amount)
	require.Equal(t, orderID, wf.OrderID)

	//Check opCU coins and coinsHold
	ethOPCU = ck.GetCU(ctx, ethOPCU.GetAddress())
	require.Equal(t, ethOPCUEthAmt.Sub(sysTransferAmt), ethOPCU.GetAssetCoins().AmountOf(ethSymbol))
	require.Equal(t, sysTransferAmt, ethOPCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	sendable := ethOPCU.IsEnabledSendTx(chain, ethOPCUEthAddress)
	require.Equal(t, false, sendable)

	//Step2, SysTransferWaitSign
	sysTransferTxHash := "sysTransferTxHash"
	chainnodeSysTransferlTx := chainnode.ExtAccountTransaction{
		Hash:     sysTransferTxHash,
		From:     ethOPCUEthAddress,
		To:       usdtOpCUEthAddress,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(100),
		GasPrice: sdk.NewInt(100), // 10Gwei
	}

	suggestGasFee := sdk.NewInt(100).MulRaw(100)
	rawData := []byte("rawData")
	signHash := "signHash"

	//mockCN.On("ValidAddress", chain, usdtSymbol, usdtOpCUEthAddress).Return(true, true, usdtOpCUEthAddress)
	mockCN.On("ValidAddress", chain, ethSymbol, ethOPCUEthAddress).Return(true, ethOPCUEthAddress)
	mockCN.On("ValidAddress", chain, ethSymbol, usdtOpCUEthAddress).Return(true, true)
	mockCN.On("QueryAccountTransactionFromData", chain, ethSymbol, rawData).Return(&chainnodeSysTransferlTx, []byte(signHash), nil)

	ctx = ctx.WithBlockHeight(20)
	result1 := keeper.SysTransferWaitSign(ctx, orderID, signHash, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, ethOPCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, suggestGasFee, order.(*sdk.OrderSysTransfer).CostFee)
	require.Equal(t, sysTransferAmt, order.(*sdk.OrderSysTransfer).Amount)
	require.Equal(t, usdtOpCUEthAddress, order.(*sdk.OrderSysTransfer).ToAddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, valid := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, ethOPCU.GetAddress(), of1.CUAddress)

	wwf, valid := receipt1.Flows[1].(sdk.SysTransferWaitSignFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wwf.OrderID)
	require.Equal(t, rawData, wwf.RawData)

	//check OPCU's coins, coinsHold and status
	ethOPCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), ethOPCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, ethOPCUEthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), ethOPCU.GetAssetCoins().AmountOf(ethSymbol))

	//Step3, SysTransferSignFinish
	signedData := []byte("signedData")
	chainnodeSysTransferlTx.BlockHeight = 10000 //eth height = 10000

	mockCN.On("VerifyAccountSignedTransaction", chain, ethSymbol, ethOPCUEthAddress, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, ethSymbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result2 := keeper.SysTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, ethOPCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt2.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of2, valid := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, ethOPCU.GetAddress(), of2.CUAddress)

	wsf, valid := receipt2.Flows[1].(sdk.SysTransferSignFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, wsf.OrderID)
	require.Equal(t, signedData, wsf.SignedTx)

	//Check user1 coins and coinsHold
	usdtOPCU = ck.GetCU(ctx, usdtOPCU.GetAddress())
	require.Equal(t, usdtOPCUEthAmt, usdtOPCU.GetAssetCoins().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetAssetCoinsHold().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetGasUsed().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetGasReceived().AmountOf(chain))

	//check OPCU's coins, coinsHold and status
	ethOPCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sysTransferAmt.Add(suggestGasFee), ethOPCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, ethOPCUEthAmt.Sub(sysTransferAmt).Sub(suggestGasFee), ethOPCU.GetAssetCoins().AmountOf(ethSymbol))
	sendable = ethOPCU.IsEnabledSendTx(chain, ethOPCUEthAddress)
	require.Equal(t, false, sendable)
	require.Equal(t, sdk.Coins(nil), ethOPCU.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), ethOPCU.GetGasReceived())

	//Step4, SysTransferFinish
	costFee := sdk.NewInt(8000)
	chainnodeSysTransferlTx.Status = chainnode.StatusSuccess
	chainnodeSysTransferlTx.CostFee = costFee
	mockCN.On("QueryAccountTransactionFromSignedData", chain, ethSymbol, signedData).Return(&chainnodeSysTransferlTx, nil).Once()

	result3 := keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, costFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeSysTransfer, order.GetOrderType())
	require.Equal(t, usdtSymbol, order.GetSymbol())
	require.Equal(t, ethOPCU.GetAddress(), order.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), order.(*sdk.OrderSysTransfer).OpCUaddress)
	require.Equal(t, rawData, order.(*sdk.OrderSysTransfer).RawData)
	require.Equal(t, signedData, order.(*sdk.OrderSysTransfer).SignedTx)
	require.Equal(t, costFee, order.(*sdk.OrderSysTransfer).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeSysTransfer, receipt3.Category)
	require.Equal(t, 2, len(receipt3.Flows))

	of3, valid := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, usdtSymbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeSysTransfer, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, ethOPCU.GetAddress(), of3.CUAddress)

	sff, valid := receipt3.Flows[1].(sdk.SysTransferFinishFlow)
	require.True(t, valid)
	require.Equal(t, orderID, sff.OrderID)
	require.Equal(t, costFee, sff.CostFee)

	//Check usdt coins and coinsHold
	usdtOPCU = ck.GetCU(ctx, usdtOPCUAddr)
	require.Equal(t, usdtOPCUEthAmt.Add(sysTransferAmt), usdtOPCU.GetAssetCoins().AmountOf(ethSymbol))
	require.Equal(t, usdtOPCUUsdtAmt, usdtOPCU.GetAssetCoins().AmountOf(usdtSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoins().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetGasReceived().AmountOf(ethSymbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetGasUsed().AmountOf(ethSymbol))

	//check systransfer deposit item
	item = ck.GetDeposit(ctx, ethSymbol, usdtOPCUAddr, sysTransferTxHash, 0)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, sysTransferAmt, item.Amount)

	//check eth OPCU's coins, coinsHold and status
	ethOPCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetAssetCoinsHold().AmountOf(ethSymbol))
	require.Equal(t, ethOPCUEthAmt.Sub(sysTransferAmt).Sub(costFee), ethOPCU.GetAssetCoins().AmountOf(ethSymbol))
	sendable = ethOPCU.IsEnabledSendTx(chain, ethOPCUEthAddress)
	require.Equal(t, sysTransferAmt.Add(costFee), ethOPCU.GetGasUsed().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetGasReceived().AmountOf(chain))
}

func TestSysTransferEthToUserCUError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	//rk := input.rk
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.UsdtToken
	chain := token.EthToken
	validators := input.validators
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup eth token info
	ethTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(chain))
	ethTokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(2, 0)
	ethTokenInfo.OpCUSysTransferNum = sdk.NewInt(100)
	ethTokenInfo.GasLimit = sdk.NewInt(21000)
	ethTokenInfo.GasPrice = sdk.NewInt(1000)
	tk.SetTokenInfo(ctx, ethTokenInfo)

	//setup usdt token info
	usdtTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	usdtTokenInfo.OpCUSysTransferNum = sdk.NewInt(10)
	usdtTokenInfo.GasLimit = sdk.NewInt(21000) //usdt gaslimt must = the gaslimit
	usdtTokenInfo.GasPrice = sdk.NewInt(800)
	tk.SetTokenInfo(ctx, usdtTokenInfo)

	_, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	require.Nil(t, err)

	//setup eth OpCU
	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	opCUEthAddr := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = opCU.SetAssetAddress(chain, opCUEthAddr, 1)
	opCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)

	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU1.GetAssetCoins().AmountOf(chain))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	amt := sdk.NewInt(10)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	toAddr := "0x81b7e08f65bdf5648606c89998a9cc8164397647"
	err = user1CU.AddAsset(symbol, toAddr, 1)
	err = user1CU.AddAsset(chain, toAddr, 1)
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, user1CU)

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, toAddr, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	/*SysTransfer*/
	sysTransferAmt := usdtTokenInfo.SysTransferAmount()
	//illegal orderID
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr).Once()
	result := keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, "illegaleOrderID", symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//from CU not exist
	user1SysTransferOrderID := uuid.NewV1().String()
	cuAddr, _ := sdk.CUAddressFromBase58("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe")
	result = keeper.SysTransfer(ctx, cuAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	//from CU is not a op CU
	result = keeper.SysTransfer(ctx, user1CUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "systransfer from a non OP CU")

	//toCU doesnot exist
	result = keeper.SysTransfer(ctx, ethOPCUAddr, cuAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	//upsupport token
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, "fcoin")
	require.Equal(t, sdk.CodeUnsupportToken, result.Code)

	//symbol is mainnet token
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, chain)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Not support systansfer chain's mainnet token")

	//usdt's sendenable is false
	usdtTokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	usdtTokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, usdtTokenInfo)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//usdt's withdrawalenable is false
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	usdtTokenInfo.IsSendEnabled = true
	usdtTokenInfo.IsWithdrawalEnabled = false
	tk.SetTokenInfo(ctx, usdtTokenInfo)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//eth's  sendenable is false
	usdtTokenInfo.IsWithdrawalEnabled = true
	tk.SetTokenInfo(ctx, usdtTokenInfo)
	ethTokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(chain))
	ethTokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, ethTokenInfo)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//eth's withdrawalenable is false
	ethTokenInfo.IsSendEnabled = true
	ethTokenInfo.IsWithdrawalEnabled = false
	tk.SetTokenInfo(ctx, ethTokenInfo)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//tranfer's sendenable is false
	ethTokenInfo.IsWithdrawalEnabled = true
	tk.SetTokenInfo(ctx, ethTokenInfo)
	keeper.SetSendEnabled(ctx, false)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//toCU has enough gasFee, no need transfer
	keeper.SetSendEnabled(ctx, true)
	gasPrice := ethTokenInfo.GasPrice
	gasLimit := usdtTokenInfo.GasLimit
	user1CU.AddGasRemained(chain, toAddr, gasPrice.Mul(gasLimit))
	ck.SetCU(ctx, user1CU)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//toCU has no order to invoke systransfer
	user1CU.SubGasRemained(chain, toAddr, gasPrice.Mul(gasLimit))
	ck.SetCU(ctx, user1CU)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// eth collect order doesn't invoke systransfer
	collectOrder := ok.NewOrderCollect(ctx, user1CUAddr, uuid.NewV1().String(), "eth",
		user1CUAddr, toAddr, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), "", 0, "")
	ok.SetOrder(ctx, collectOrder)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	collectOrder.Symbol = "usdt"
	ok.SetOrder(ctx, collectOrder)

	//order already exist
	duplicatedOrderID := uuid.NewV1().String()
	duplicatdcOrder := &sdk.OrderSysTransfer{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        duplicatedOrderID,
			OrderType: sdk.OrderTypeSysTransfer,
			Symbol:    token.BtcToken,
		},
	}
	ok.SetOrder(ctx, duplicatdcOrder)

	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, duplicatedOrderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//systransfer more than asset coins
	usdtTokenInfo.OpCUSysTransferNum = sdk.NewInt(1)
	tk.SetTokenInfo(ctx, usdtTokenInfo)
	opCU.SubAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, opCUEthAmt.Sub(sysTransferAmt).AddRaw(1))))
	ck.SetCU(ctx, opCU)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeInsufficientCoins, result.Code)

	//every things is ok
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, opCUEthAmt.Sub(sysTransferAmt).AddRaw(1))))
	ck.SetCU(ctx, opCU)
	result = keeper.SysTransfer(ctx, ethOPCUAddr, user1CUAddr, toAddr, user1SysTransferOrderID, symbol)
	require.Equal(t, sdk.CodeOK, result.Code)

	gasPrice = ethTokenInfo.GasPrice
	costFee := sdk.NewInt(1400000)

	/*WithdrawalWaitSign*/
	rawData := []byte("rawData")
	signHash := []byte("signHash")
	withdrawalTxHash := "withdrawalTxHash"
	chainnodeSysTransferTx := chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}

	//order not found
	mockCN.On("ValidAddress", chain, chain, opCUEthAddr).Return(true, opCUEthAddr)
	result = keeper.SysTransferWaitSign(ctx, "user1SysTransferOrderID", string(signHash), rawData)
	require.Equal(t, sdk.CodeNotFoundOrder, result.Code)

	//toAddr mismatch
	chainnodeSysTransferTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       "toAddr mismatch",
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, signHash, nil).Once()
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "Unexpected systransfer to address")

	//amount mismatch
	chainnodeSysTransferTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   sysTransferAmt.SubRaw(1),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, signHash, nil).Once()
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "Unexpected systransfer Amount")

	//gaslimit mismatch
	chainnodeSysTransferTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000).AddRaw(1),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, signHash, nil).Once()
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//price is too high
	ethTokenInfo.GasPrice = gasPrice.MulRaw(10).QuoRaw(12)
	tk.SetTokenInfo(ctx, ethTokenInfo)
	chainnodeSysTransferTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, signHash, nil).Once()
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//price is too low
	ethTokenInfo.GasPrice = gasPrice.MulRaw(10).QuoRaw(8).AddRaw(1)
	tk.SetTokenInfo(ctx, ethTokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, signHash, nil).Once()
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// signHash not match
	ethTokenInfo.GasPrice = gasPrice
	tk.SetTokenInfo(ctx, ethTokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, []byte("wrong hash"), nil).Once()
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Contains(t, result.Log, "hash mismatch")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//everything is ok
	mockCN.On("QueryAccountTransactionFromData", chain, chain, rawData).Return(&chainnodeSysTransferTx, signHash, nil)
	result = keeper.SysTransferWaitSign(ctx, user1SysTransferOrderID, string(signHash), rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	/*SysTransferSignFinish*/
	signedTx := []byte("signedTx")
	chainnodeSysTransferTx1 := chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}

	//VerifyAccountSignedTransaction err
	mockCN.On("VerifyAccountSignedTransaction", chain, chain, opCUEthAddr, signedTx).Return(true, errors.New("VerifyAccountSignedTransactionError")).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//VerifyAccountSignedTransaction,  verified = false
	mockCN.On("VerifyAccountSignedTransaction", chain, chain, opCUEthAddr, signedTx).Return(false, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//QueryAccountTransactionFromSignedData, err
	mockCN.On("VerifyAccountSignedTransaction", chain, chain, opCUEthAddr, signedTx).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, errors.New("QueryAccountTransactionFromSignedDataError")).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//toAddr mismatch
	chainnodeSysTransferTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       "toAddr mismatch",
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//amount mismatch
	chainnodeSysTransferTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   sysTransferAmt.AddRaw(1),
		Nonce:    1,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//gasPrice mimatch
	chainnodeSysTransferTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice.AddRaw(1),
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//gasLimit mismatch
	chainnodeSysTransferTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000).SubRaw(1),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//from address mismatch
	chainnodeSysTransferTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     "fromAddr mismatch",
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//everything is ok
	chainnodeSysTransferTx1.From = opCUEthAddr
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx1, nil).Once()
	result = keeper.SysTransferSignFinish(ctx, user1SysTransferOrderID, signedTx)
	require.Equal(t, sdk.CodeOK, result.Code)

	/*SystranferFinish*/
	chainnodeSysTransferTx2 := chainnode.ExtAccountTransaction{
		Status:   chainnode.StatusSuccess,
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   sysTransferAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
		CostFee:  costFee,
	}

	//1st comfirm
	result = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), user1SysTransferOrderID, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//2nd comfirm
	result = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[1].OperatorAddress), user1SysTransferOrderID, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//3nd comfirm
	mockCN.On("QueryAccountTransactionFromSignedData", chain, chain, signedTx).Return(&chainnodeSysTransferTx2, nil).Once()
	result = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[2].OperatorAddress), user1SysTransferOrderID, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//4th confirm
	result = keeper.SysTransferFinish(ctx, sdk.CUAddress(validators[3].OperatorAddress), user1SysTransferOrderID, costFee)
	require.Equal(t, sdk.CodeOK, result.Code, result)

	//Check user1CU coins
	user1CU = ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	require.Equal(t, sdk.NewInt(10), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sysTransferAmt, user1CU.GetGasReceived().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetGasUsed().AmountOf(chain))

	//check opCU's coins, coinsonhold, deposit items
	opCU = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoins().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoinsHold().AmountOf(chain))
	require.Equal(t, opCUEthAmt.Sub(sysTransferAmt).Sub(costFee), opCU.GetAssetCoins().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(chain))
	require.Equal(t, sysTransferAmt.Add(costFee), opCU.GetGasUsed().AmountOf(chain))
	sendable := opCU.IsEnabledSendTx(chain, opCUEthAddr)
	require.True(t, sendable)
}

func TestSysTransferEthToOPCUError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.UsdtToken
	chain := token.EthToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup eth token info
	ethTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	ethTokenInfo.OpCUSysTransferNum = sdk.NewInt(200)
	ethTokenInfo.GasLimit = sdk.NewInt(100)
	ethTokenInfo.GasPrice = sdk.NewInt(100)
	tk.SetTokenInfo(ctx, ethTokenInfo)

	//setup usdt token info
	usdtTokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	usdtTokenInfo.OpCUSysTransferNum = sdk.NewInt(5)
	usdtTokenInfo.GasLimit = sdk.NewInt(100) //usdt gaslimt must = the gaslimit
	usdtTokenInfo.GasPrice = sdk.NewInt(8000)
	tk.SetTokenInfo(ctx, usdtTokenInfo)

	//set ethOPCU
	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	ethOPCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	ethOPCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = ethOPCU.SetAssetAddress(chain, ethOPCUEthAddress, 1)
	ethOPCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)

	ethOPCUEthAmt := sdk.NewInt(90000000)
	item, err := sdk.NewDepositItem("opcu1deposithash", 0, ethOPCUEthAmt, ethOPCUEthAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	ck.SaveDeposit(ctx, chain, ethOPCUAddr, item)

	ethOPCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, ethOPCUEthAmt)))
	ck.SetCU(ctx, ethOPCU)

	ethOPCU = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, ethOPCUEthAmt, ethOPCU.GetAssetCoins().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetAssetCoinsHold().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetCoins().AmountOf(chain))
	require.Equal(t, sdk.ZeroInt(), ethOPCU.GetCoinsHold().AmountOf(chain))

	//setup usdtopcu
	usdtOPCUAddr, err := sdk.CUAddressFromBase58("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx")
	usdtOPCU := ck.NewOpCUWithAddress(ctx, symbol, usdtOPCUAddr)
	usdtOPCUEthAddress := "0x81b7E08F65Bdf5648606c89998A9CC8164397647"
	err = usdtOPCU.SetAssetAddress(symbol, usdtOPCUEthAddress, 1)
	err = usdtOPCU.SetAssetAddress(chain, usdtOPCUEthAddress, 1)
	usdtOPCU.SetAssetPubkey(pubkey.Bytes(), 1)
	usdtOPCUUsdtAmt := sdk.NewInt(100)
	item, err = sdk.NewDepositItem("usdtopcudeposithash", 0, usdtOPCUUsdtAmt, usdtOPCUEthAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	ck.SaveDeposit(ctx, symbol, usdtOPCUAddr, item)
	usdtOPCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, usdtOPCUUsdtAmt)))
	ck.SetCU(ctx, usdtOPCU)

	usdtOPCU = ck.GetCU(ctx, sdk.CUAddress(usdtOPCUAddr))
	require.Equal(t, usdtOPCUUsdtAmt, usdtOPCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), usdtOPCU.GetCoinsHold().AmountOf(symbol))

	mockCN.On("ValidAddress", chain, chain, usdtOPCUEthAddress).Return(true, usdtOPCUEthAddress)
	mockCN.On("ValidAddress", chain, chain, ethOPCUEthAddress).Return(true, ethOPCUEthAddress)

	ethSysTranasferOrderID := uuid.NewV1().String()
	ctx = ctx.WithBlockHeight(11)
	orderID := ethSysTranasferOrderID
	gasPrice := ethTokenInfo.GasPrice
	gasLimit := usdtTokenInfo.GasLimit

	//usdt opcu has  has enough
	usdtOPCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, gasPrice.Mul(gasLimit).MulRaw(types.MaxSystransferNum))))
	ck.SetCU(ctx, usdtOPCU)
	mockCN.On("ValidAddress", chain, symbol, usdtOPCUEthAddress).Return(true, usdtOPCUEthAddress).Once()
	result := keeper.SysTransfer(ctx, ethOPCUAddr, usdtOPCUAddr, usdtOPCUEthAddress, orderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "does not need systransfer")

	// usdt opcu has not order to invoke systransfer
	usdtOPCU.SubAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, gasPrice.Mul(gasLimit))))
	ck.SetCU(ctx, usdtOPCU)
	mockCN.On("ValidAddress", chain, symbol, usdtOPCUEthAddress).Return(true, usdtOPCUEthAddress).Once()
	result = keeper.SysTransfer(ctx, ethOPCUAddr, usdtOPCUAddr, usdtOPCUEthAddress, orderID, symbol)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "does not need systransfer")

	userCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)

	// eth withdraw order doesn't invoke systransfer
	withdrawOrder := ok.NewOrderWithdrawal(ctx, userCUAddr, uuid.NewV1().String(), "eth",
		sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), "", "HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx", "")
	withdrawOrder.FromAddress = usdtOPCUEthAddress
	ok.SetOrder(ctx, withdrawOrder)

	withdrawOrder.Symbol = "usdt"
	ok.SetOrder(ctx, withdrawOrder)
}
