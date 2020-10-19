package tests

import (
	"errors"
	"testing"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestWithdrawalBtcSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	ctx = ctx.WithBlockHeight(10)
	validators := input.validators
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.BtcToken
	chain := token.BtcToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tokenInfo.GasPrice = sdk.NewInt(10000000 / 380)
	tk.SetTokenInfo(ctx, tokenInfo)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	d0Amt := sdk.NewInt(10000000)
	hash0 := "opcu_utxo_deposit_0"
	d0, err := sdk.NewDepositItem(hash0, 0, d0Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	d1Amt := sdk.NewInt(80000000)
	hash1 := "opcu_utxo_deposit_1"
	d1, err := sdk.NewDepositItem(hash1, 0, d1Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, btcOPCUAddr)
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	opCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)
	deposiList := sdk.DepositList{d0, d1}.Sort()

	ck.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), deposiList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, d0Amt.Add(d1Amt))))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, btcOPCUAddr)
	require.Equal(t, d0Amt.Add(d1.Amount), opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, deposiList, ck.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)

	amt := sdk.NewInt(80000000)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	err = user1CU.AddAsset(symbol, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", 1)
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, user1CU)

	toAddr := "mnRw8TRyxUVEv1CnfzpahuRr5BeWYsCGES"

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddress).Return(true, opCUBtcAddress)

	//Step1, Withdrawal
	ctx = ctx.WithBlockHeight(11)
	orderID := uuid.NewV1().String()
	withdrawalAmt := sdk.NewInt(40000000)
	gasFee := sdk.NewInt(10000)
	result := keeper.Withdrawal(ctx, user1CUAddr, toAddr, orderID, symbol, withdrawalAmt, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	o := ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusBegin, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt.Category)
	require.Equal(t, 3, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, user1CUAddr, of.CUAddress)

	wf, v := receipt.Flows[1].(sdk.WithdrawalFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr.String(), wf.FromCu)
	require.Equal(t, toAddr, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, withdrawalAmt, wf.Amount)
	require.Equal(t, gasFee, wf.GasFee)
	require.Equal(t, orderID, wf.OrderID)

	bf, v := receipt.Flows[2].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf.Symbol)
	require.Equal(t, amt, bf.PreviousBalance)
	require.Equal(t, withdrawalAmt.Add(gasFee).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, withdrawalAmt.Add(gasFee), bf.BalanceOnHoldChange)

	//Check user1 coins and coinsHold
	user1CU1 := ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, withdrawalAmt.Add(gasFee), user1CU1.GetCoinsHold().AmountOf(symbol))

	//Step2, WithdrawalWaitSign
	vin1 := sdk.NewUtxoIn(hash0, 0, d0Amt, opCUBtcAddress)
	vin2 := sdk.NewUtxoIn(hash1, 0, d1Amt, opCUBtcAddress)
	signHash1 := []byte("signHash1")
	signHash2 := []byte("signHash2")
	signHashes := [][]byte{signHash1, signHash2}
	withdrawalTxHash := "withdrawalTxHash"
	chainnodewithdrawalTx := &chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin1, &vin2},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: withdrawalAmt},
			{Address: opCUBtcAddress, Amount: d0Amt.Add(d1Amt).Sub(withdrawalAmt).Sub(gasFee)},
		},
		CostFee: gasFee,
	}

	ins := chainnodewithdrawalTx.Vins

	rawData := []byte("rawData")
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(chainnodewithdrawalTx, signHashes, nil)

	ctx = ctx.WithBlockHeight(1)
	result1 := keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{orderID}, []string{string(signHash1), string(signHash2)}, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, btcOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, v := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, user1CUAddr, of1.CUAddress)

	wwf, v := receipt1.Flows[1].(sdk.WithdrawalWaitSignFlow)
	require.True(t, v)
	require.Equal(t, []string{orderID}, wwf.OrderIDs)
	require.Equal(t, rawData, wwf.RawData)
	require.Equal(t, btcOPCUAddr.String(), wwf.OpCU)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, withdrawalAmt.Add(gasFee), user1CU1.GetCoinsHold().AmountOf(symbol))

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, btcOPCUAddr)
	require.Equal(t, d0Amt.Add(d1Amt), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoins().AmountOf(symbol))
	sendable := opCU.IsEnabledSendTx(chain, opCUBtcAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasReceived().AmountOf(symbol))

	//Step3, WithdrawalSignFinish
	signedData := []byte("signedData")
	chainnodewithdrawalTx.BlockHeight = 10000 //btc height = 10000
	mockCN.On("QueryUtxoInsFromData", chain, symbol, signedData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodewithdrawalTx, nil).Once()
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{opCUBtcAddress, opCUBtcAddress}, signedData, ins).Return(true, nil).Once()

	result2 := keeper.WithdrawalSignFinish(ctx, []string{orderID}, signedData, "")
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, btcOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, rawData, o.(*sdk.OrderWithdrawal).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderWithdrawal).SignedTx)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt2.Category)
	require.Equal(t, 2, len(receipt2.Flows))

	of2, v := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, user1CUAddr, of2.CUAddress)

	wsf, v := receipt2.Flows[1].(sdk.WithdrawalSignFinishFlow)
	require.True(t, v)
	require.Equal(t, []string{orderID}, wsf.OrderIDs)
	require.Equal(t, signedData, wsf.SignedTx)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, withdrawalAmt.Add(gasFee), user1CU1.GetCoinsHold().AmountOf(symbol))

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, btcOPCUAddr)
	require.Equal(t, d0Amt.Add(d1Amt), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUBtcAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasReceived().AmountOf(symbol))

	//Step4, WithdrawalFinish
	chainnodewithdrawalTx.Status = chainnode.StatusSuccess //btc height = 10000
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodewithdrawalTx, nil)

	//1st confirm
	result3 := keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{orderID}, gasFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//2nd confirm
	result3 = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{orderID}, gasFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//3rd confirm
	result3 = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{orderID}, gasFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, btcOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, rawData, o.(*sdk.OrderWithdrawal).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderWithdrawal).SignedTx)
	require.Equal(t, gasFee, o.(*sdk.OrderWithdrawal).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt3.Category)
	require.Equal(t, 3, len(receipt3.Flows))

	of3, v := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, btcOPCUAddr, of3.CUAddress)

	wff, v := receipt3.Flows[1].(sdk.WithdrawalFinishFlow)
	require.True(t, v)
	require.Equal(t, []string{orderID}, wff.OrderIDs)
	require.Equal(t, gasFee, wff.CostFee)

	bf3, v := receipt3.Flows[2].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, withdrawalAmt.Add(gasFee), bf3.PreviousBalanceOnHold)
	require.Equal(t, withdrawalAmt.Add(gasFee).Neg(), bf3.BalanceOnHoldChange)

	item := ck.GetDeposit(ctx, symbol, opCU.GetAddress(), withdrawalTxHash, 1)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, d0Amt.Add(d1Amt).Sub(withdrawalAmt).Sub(gasFee), item.Amount)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU1.GetCoinsHold().AmountOf(symbol))

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, btcOPCUAddr)
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, d0Amt.Add(d1Amt).Sub(withdrawalAmt).Sub(gasFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUBtcAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, gasFee, opCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, gasFee, opCU.GetGasReceived().AmountOf(symbol))

}

func TestWithdrawalBtcSuccessWithMultiOrders(t *testing.T) {
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
	symbol := token.BtcToken
	chain := token.BtcToken
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tokenInfo.GasPrice = sdk.NewInt(10000000 / 580)
	tk.SetTokenInfo(ctx, tokenInfo)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	d0Amt := sdk.NewInt(10000000)
	hash0 := "opcu_utxo_deposit_0"
	d0, err := sdk.NewDepositItem(hash0, 0, d0Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	d1Amt := sdk.NewInt(80000000)
	hash1 := "opcu_utxo_deposit_1"
	d1, err := sdk.NewDepositItem(hash1, 0, d1Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	opCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)
	deposiList := sdk.DepositList{d0, d1}.Sort()

	ck.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), deposiList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, d0Amt.Add(d1Amt))))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	require.Equal(t, d0Amt.Add(d1.Amount), opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, deposiList, ck.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	amt := sdk.NewInt(80000000)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	err = user1CU.AddAsset(symbol, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", 1)
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, user1CU)
	toAddr := "mnRw8TRyxUVEv1CnfzpahuRr5BeWYsCGES"

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	/*Withdrawal*/
	user1WithdrawalOrderID1 := uuid.NewV1().String()
	user1WithdrawalOrderID2 := uuid.NewV1().String()
	user1WithdrawalOrderID3 := uuid.NewV1().String()
	user1WithdrawalOrderID4 := uuid.NewV1().String()
	user1WithdrawalOrderID5 := uuid.NewV1().String()

	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddress).Return(true, opCUBtcAddress)

	result := keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID1, symbol, sdk.NewInt(1000001), sdk.NewInt(100000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID2, symbol, sdk.NewInt(1000002), sdk.NewInt(100000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID3, symbol, sdk.NewInt(1000003), sdk.NewInt(100000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID4, symbol, sdk.NewInt(1000004), sdk.NewInt(100000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID5, symbol, sdk.NewInt(1000005), sdk.NewInt(100000))
	require.Equal(t, sdk.CodeOK, result.Code)

	o := ok.GetOrder(ctx, user1WithdrawalOrderID5)
	require.NotNil(t, o)
	require.Equal(t, user1WithdrawalOrderID5, o.GetID())
	require.Equal(t, sdk.OrderStatusBegin, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt.Category)
	require.Equal(t, 3, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, user1WithdrawalOrderID5, of.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, user1CUAddr, of.CUAddress)

	wf, v := receipt.Flows[1].(sdk.WithdrawalFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr.String(), wf.FromCu)
	require.Equal(t, toAddr, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, sdk.NewInt(1000005), wf.Amount)
	require.Equal(t, sdk.NewInt(100000), wf.GasFee)
	require.Equal(t, user1WithdrawalOrderID5, wf.OrderID)

	bf, v := receipt.Flows[2].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf.Symbol)
	require.Equal(t, amt.SubRaw(4400010), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(1100005).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.NewInt(4400010), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1100005), bf.BalanceOnHoldChange)

	//Check user1 coins and coinsHold
	user1CU1 := ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.SubRaw(5500015), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(5500015), user1CU1.GetCoinsHold().AmountOf(symbol))

	//Step2, WithdrawalWaitSign
	vin1 := sdk.NewUtxoIn(hash0, 0, d0Amt, opCUBtcAddress)
	vin2 := sdk.NewUtxoIn(hash1, 0, d1Amt, opCUBtcAddress)
	costFee := sdk.NewInt(10000)
	signHash1 := []byte("signHash1")
	signHash2 := []byte("signHash2")
	signHashes := [][]byte{signHash1, signHash2}
	withdrawalTxHash := "withdrawalTxHash"

	chainnodewithdrawalTx := &chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin1, &vin2},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: sdk.NewInt(1000001)},
			{Address: toAddr, Amount: sdk.NewInt(1000002)},
			{Address: toAddr, Amount: sdk.NewInt(1000003)},
			{Address: toAddr, Amount: sdk.NewInt(1000004)},
			{Address: toAddr, Amount: sdk.NewInt(1000005)},
			{Address: opCUBtcAddress, Amount: d0Amt.Add(d1Amt).SubRaw(6000015).Sub(costFee)},
			{Address: opCUBtcAddress, Amount: sdk.NewInt(1000000)},
		},
		CostFee: costFee,
	}
	ins := chainnodewithdrawalTx.Vins

	rawData := []byte("rawData")
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(chainnodewithdrawalTx, signHashes, nil)

	ctx = ctx.WithBlockHeight(20)
	result1 := keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, []string{string(signHash1), string(signHash2)}, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	o = ok.GetOrder(ctx, user1WithdrawalOrderID1)
	require.NotNil(t, o)
	require.Equal(t, user1WithdrawalOrderID1, o.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, btcOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, v := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, user1WithdrawalOrderID1, of1.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, user1CUAddr, of1.CUAddress)

	wwf, v := receipt1.Flows[1].(sdk.WithdrawalWaitSignFlow)
	require.True(t, v)
	require.Equal(t, []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, wwf.OrderIDs)
	require.Equal(t, rawData, wwf.RawData)
	require.Equal(t, btcOPCUAddr.String(), wwf.OpCU)

	//Check user1 coins and coinsHold
	user1CU = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.SubRaw(5500015), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(5500015), user1CU.GetCoinsHold().AmountOf(symbol))

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, btcOPCUAddr)
	require.Equal(t, d0Amt.Add(d1Amt), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoins().AmountOf(symbol))
	sendable := opCU.IsEnabledSendTx(chain, opCUBtcAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetGasReceived().AmountOf(symbol))

	//Step3, WithdrawalSignFinish
	signedData := []byte("signedData")
	chainnodewithdrawalTx.BlockHeight = 10000 //btc height = 10000
	mockCN.On("QueryUtxoInsFromData", chain, symbol, signedData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodewithdrawalTx, nil).Once()
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{opCUBtcAddress, opCUBtcAddress}, signedData, ins).Return(true, nil).Once()

	result2 := keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, signedData, "")
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	o = ok.GetOrder(ctx, user1WithdrawalOrderID1)
	require.NotNil(t, o)
	require.Equal(t, user1WithdrawalOrderID1, o.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, btcOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, rawData, o.(*sdk.OrderWithdrawal).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderWithdrawal).SignedTx)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt2.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of2, v := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, user1WithdrawalOrderID1, of2.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, user1CUAddr, of2.CUAddress)

	wsf, v := receipt2.Flows[1].(sdk.WithdrawalSignFinishFlow)
	require.True(t, v)
	require.Equal(t, []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, wsf.OrderIDs)
	require.Equal(t, signedData, wsf.SignedTx)

	//Step4, WithdrawalFinish
	chainnodewithdrawalTx.Status = chainnode.StatusSuccess //btc height = 10000
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodewithdrawalTx, nil).Once()

	//1st confirm
	result3 := keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, costFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//2nd confirm
	result3 = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, costFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//3rd confirm
	result3 = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, costFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	o = ok.GetOrder(ctx, user1WithdrawalOrderID1)
	require.NotNil(t, o)
	require.Equal(t, user1WithdrawalOrderID1, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, btcOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, rawData, o.(*sdk.OrderWithdrawal).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderWithdrawal).SignedTx)
	require.Equal(t, costFee, o.(*sdk.OrderWithdrawal).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt3.Category)
	require.Equal(t, 7, len(receipt3.Flows))

	of3, v := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, user1WithdrawalOrderID1, of3.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, btcOPCUAddr, of3.CUAddress)

	wff, v := receipt3.Flows[1].(sdk.WithdrawalFinishFlow)
	require.True(t, v)
	require.Equal(t, []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5}, wff.OrderIDs)
	require.Equal(t, costFee, wff.CostFee)

	bf3, v := receipt3.Flows[2].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, sdk.NewInt(74499985), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, sdk.NewInt(5500015), bf3.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1100001).Neg(), bf3.BalanceOnHoldChange)

	bf3, v = receipt3.Flows[3].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, sdk.NewInt(74499985), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, sdk.NewInt(4400014), bf3.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1100002).Neg(), bf3.BalanceOnHoldChange)

	bf3, v = receipt3.Flows[4].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, sdk.NewInt(74499985), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, sdk.NewInt(3300012), bf3.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1100003).Neg(), bf3.BalanceOnHoldChange)

	bf3, v = receipt3.Flows[5].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, sdk.NewInt(74499985), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, sdk.NewInt(2200009), bf3.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1100004).Neg(), bf3.BalanceOnHoldChange)

	bf3, v = receipt3.Flows[6].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, sdk.NewInt(74499985), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, sdk.NewInt(1100005), bf3.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1100005).Neg(), bf3.BalanceOnHoldChange)

	//check changeback deposit items
	item := ck.GetDeposit(ctx, symbol, opCU.GetAddress(), withdrawalTxHash, 5)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, d0Amt.Add(d1Amt).SubRaw(6000015).Sub(costFee), item.Amount)

	item = ck.GetDeposit(ctx, symbol, opCU.GetAddress(), withdrawalTxHash, 6)
	require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
	require.Equal(t, sdk.NewInt(1000000), item.Amount)

	//Check user1 coins and coinsHold
	user1CU = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.SubRaw(5500015), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetCoinsHold().AmountOf(symbol))

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, btcOPCUAddr)
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, d0Amt.Add(d1Amt).SubRaw(5000015).Sub(costFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUBtcAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, costFee, opCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(500000), opCU.GetGasReceived().AmountOf(symbol))

}

/*
user1 has 80000000btc, withdrawal 60000000btc with 10000000btc as gas
opCU  has 80000000btc, withdrawal 60000000btc used 2000000btc as gas
opCU has 2 utxo deposit items,
 {opcu_utxo_deposit_0, 0, 10000000}
 {opcu_utxo_deposit_1, 0, 80000000}
fininaly has 2 utxo deposit items
 {opcu_utxo_deposit_0, 0, 10000000}
 {withdrawalTxHash, 1, 18000000}
*/

func TestWithdrawalBtcError(t *testing.T) {
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
	symbol := token.BtcToken
	chain := token.BtcToken
	validators := input.validators
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tokenInfo.GasPrice = sdk.NewInt(10000000 / 380)
	tk.SetTokenInfo(ctx, tokenInfo)

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	require.Nil(t, err)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	d0Amt := sdk.NewInt(10000000)
	hash0 := "opcu_utxo_deposit_0"
	d0, err := sdk.NewDepositItem(hash0, 0, d0Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	d1Amt := sdk.NewInt(80000000)
	hash1 := "opcu_utxo_deposit_1"
	d1, err := sdk.NewDepositItem(hash1, 0, d1Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	opCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)
	deposiList := sdk.DepositList{d0, d1}.Sort()

	ck.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), deposiList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, d0Amt.Add(d1Amt))))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	require.Equal(t, d0Amt.Add(d1.Amount), opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, deposiList, ck.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))
	require.Equal(t, d0, ck.GetDeposit(ctx, symbol, btcOPCUAddr, d0.Hash, d0.Index))
	require.Equal(t, d1, ck.GetDeposit(ctx, symbol, btcOPCUAddr, d1.Hash, d1.Index))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	amt := sdk.NewInt(80000000)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	err = user1CU.AddAsset(symbol, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", 1)
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, user1CU)
	toAddr := "mnRw8TRyxUVEv1CnfzpahuRr5BeWYsCGES"

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	/*Withdrawal*/

	//illegal uuid
	mockCN.On("SupportChain", chain).Return(true)
	result := keeper.Withdrawal(ctx, user1CUAddr, toAddr, "illegaleOrderID", symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//from CU not exist
	user1WithdrawalOrderID := uuid.NewV1().String()
	cuAddr, _ := sdk.CUAddressFromBase58("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe")
	result = keeper.Withdrawal(ctx, cuAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	//from CU is not a user CU
	result = keeper.Withdrawal(ctx, btcOPCUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "withdrawal from a non user CU")

	//invalid token name
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, "Fcoin", sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidSymbol, result.Code)

	//upsupport token
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, "fcoin", sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeUnsupportToken, result.Code)

	//token's sendenable is false
	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//token's withdrawenable is false
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//tranfer's sendenable is false
	tokenInfo.IsWithdrawalEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)
	keeper.SetSendEnabled(ctx, false)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//orderid already exist
	keeper.SetSendEnabled(ctx, true)
	duplicatedOrderID := uuid.NewV1().String()
	duplicatdcOrder := &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        duplicatedOrderID,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    token.BtcToken,
		},
	}
	ok.SetOrder(ctx, duplicatdcOrder)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, duplicatedOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "already exists")

	//toaddress is not valid
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(false, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidAddress, result.Code)
	require.Contains(t, result.Log, "is not a valid address")

	//gasFee LT WithdrawalFeeRate
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(69999000), sdk.NewInt(100))
	require.Equal(t, sdk.CodeInsufficientFee, result.Code)

	//amt is negative
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000).Neg(), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeAmountError, result.Code)

	//amt is zero
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.ZeroInt(), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeAmountError, result.Code)

	//withdrawal more than coins
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(70000001), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInsufficientCoins, result.Code)

	//every things is ok
	withdrawalAmt := sdk.NewInt(60000000)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, withdrawalAmt, sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeOK, result.Code)

	/*WithdrawalWaitSign*/
	costFee := sdk.NewInt(2000000)
	rawData := []byte("rawData")
	vin := sdk.NewUtxoIn(d1.Hash, d1.Index, d1.Amount, opCUBtcAddress)
	signHash1 := []byte("signHash1")
	signHashes := [][]byte{signHash1}
	withdrawalTxHash := "withdrawalTxHash"
	chainnodeWithdrawalTx := chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: withdrawalAmt},
			{Address: opCUBtcAddress, Amount: d1Amt.Sub(withdrawalAmt).Sub(costFee)},
		},
		CostFee: costFee,
	}

	//QueryUtxoInsFromData error
	ins := chainnodeWithdrawalTx.Vins
	require.Equal(t, d1.Hash, ins[0].Hash)
	require.Equal(t, d1.Amount, ins[0].Amount)
	require.Equal(t, d1.Index, uint64(ins[0].Index))
	require.Equal(t, opCUBtcAddress, ins[0].Address)
	require.Equal(t, d1, ck.GetDeposit(ctx, symbol, btcOPCUAddr, ins[0].Hash, uint64(ins[0].Index)))

	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddress).Return(true, opCUBtcAddress)
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, errors.New("Fail to QueryUtxoInsFromData")).Once()
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//Empty vins
	nonExistVins := []*sdk.UtxoIn{
		&sdk.UtxoIn{
			Hash:    "nonexist",
			Index:   1,
			Amount:  sdk.NewInt(100),
			Address: opCUBtcAddress,
		},
	}
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(nonExistVins, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//Vins status is not DepositItemStatusConfirmed
	nonCollecectdVins := []*sdk.UtxoIn{
		&sdk.UtxoIn{
			Hash:    "noncollected",
			Index:   1,
			Amount:  sdk.NewInt(100),
			Address: opCUBtcAddress,
		},
	}
	nonCollectDepositItem, _ := sdk.NewDepositItem(nonCollecectdVins[0].Hash, nonCollecectdVins[0].Index, nonCollecectdVins[0].Amount, opCUBtcAddress, " ", sdk.DepositItemStatusConfirmed)
	ck.SetDepositList(ctx, symbol, btcOPCUAddr, sdk.DepositList{d0, d1, nonCollectDepositItem})

	//QuerUtxoTransactionFromData fail
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeWithdrawalTx, signHashes, errors.New("QueryUtxoTransactionFromDataError")).Once()
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//price is too high
	actualPrice := sdk.NewDecFromInt(costFee).Quo(sdk.NewDec(230)).Mul(sdk.NewDec(sdk.KiloBytes))
	tokenPrice := actualPrice.Quo(sdk.NewDecWithPrec(12, 1))
	tokenInfo.GasPrice = tokenPrice.TruncateInt()
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeWithdrawalTx, signHashes, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too high")

	//price is too low
	tokenPrice = actualPrice.Quo(sdk.NewDecWithPrec(8, 1))
	tokenInfo.GasPrice = tokenPrice.TruncateInt().AddRaw(1)
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeWithdrawalTx, signHashes, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too low")

	//len(hashes) != len(signHashes)
	tokenInfo.GasPrice = actualPrice.TruncateInt()
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeWithdrawalTx, signHashes, nil)
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1), "noexisthash"}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "signhashes's number mismatch")

	//hash mismatch
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{"mismatchedhash"}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "mismatch hashes")

	//everything is ok
	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	/*WithdrawalSignFinish*/
	signedTx := []byte("signedTx")
	chainnodeWithdrawalTx1 := chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: withdrawalAmt},
			{Address: opCUBtcAddress, Amount: d1Amt.Sub(withdrawalAmt).Sub(costFee)},
		},
		CostFee: costFee,
	}

	//VerifyUtxoSignedTransaction err
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{opCUBtcAddress}, signedTx, ins).Return(true, errors.New("VerifyUtxoSignedTransactionError")).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "VerifyUtxoSignedTransactionError")

	//VerifyUtxoSignedTransaction,  verified = false
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{opCUBtcAddress}, signedTx, ins).Return(false, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//QueryUtxoTransactionFromSignedData, err
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{opCUBtcAddress}, signedTx, ins).Return(true, nil).Once()
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeWithdrawalTx1, errors.New("QueryUtxoTransactionFromSignedDataError")).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//Vins mimatch
	vin1 := sdk.NewUtxoIn(hash1, 1, d1Amt, opCUBtcAddress)
	chainnodeWithdrawalTx1 = chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin1},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: withdrawalAmt},
			{Address: opCUBtcAddress, Amount: d1Amt.Sub(withdrawalAmt).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{opCUBtcAddress}, signedTx, ins).Return(true, nil)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//Vout mismatch
	chainnodeWithdrawalTx1 = chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: withdrawalAmt.SubRaw(1)},
			{Address: opCUBtcAddress, Amount: d1Amt.Sub(withdrawalAmt).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//tx.CostFee != order.CostFee
	chainnodeWithdrawalTx1 = chainnode.ExtUtxoTransaction{
		Hash: withdrawalTxHash,
		Vins: []*sdk.UtxoIn{&vin},
		Vouts: []*sdk.UtxoOut{
			{Address: toAddr, Amount: withdrawalAmt},
			{Address: opCUBtcAddress, Amount: d1Amt.Sub(withdrawalAmt).Sub(costFee)},
		},
		CostFee: costFee.SubRaw(1),
	}
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//everything is ok
	chainnodeWithdrawalTx1.CostFee = costFee
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeWithdrawalTx1, nil)
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeOK, result.Code)

	//QueryUtxoTransaction err
	result = keeper.WithdrawalFinish(ctx, fromCUAddr, []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//empty order
	withdrawalOrder := ok.GetOrder(ctx, user1WithdrawalOrderID)
	ok.DeleteOrder(ctx, withdrawalOrder)
	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeNotFoundOrder, result.Code)
	require.Contains(t, result.Log, "does not exist")

	//order status is sdk.OrderStatusFinish
	orderStatus := withdrawalOrder.GetOrderStatus()
	withdrawalOrder.SetOrderStatus(sdk.OrderStatusFinish)
	ok.SetOrder(ctx, withdrawalOrder)
	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	withdrawalOrder.SetOrderStatus(orderStatus)
	ok.SetOrder(ctx, withdrawalOrder)
	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	//everything is ok
	//Check user1CU coins
	user1CU = ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	require.Equal(t, sdk.NewInt(10000000), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoinsHold().AmountOf(symbol))

	//check opCU's coins, coinsonhold, deposit items
	opCU = ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(28000000), opCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(symbol))
	sendable := opCU.IsEnabledSendTx(chain, opCUBtcAddress)
	require.True(t, sendable)
	require.Equal(t, sdk.NewInt(10000000), opCU.GetGasReceived().AmountOf(symbol))
	require.Equal(t, costFee, opCU.GetGasUsed().AmountOf(symbol))

	ck.DelDeposit(ctx, symbol, btcOPCUAddr, nonCollectDepositItem.Hash, nonCollectDepositItem.Index)
	depositList := ck.GetDepositList(ctx, symbol, btcOPCUAddr)
	require.Equal(t, d0, depositList[0])
	require.Equal(t, withdrawalTxHash, depositList[1].Hash)
	require.Equal(t, sdk.NewInt(18000000), depositList[1].Amount)
	require.Equal(t, uint64(1), depositList[1].Index)
	require.Equal(t, sdk.DepositItemStatusConfirmed, depositList[1].Status)

}

func TestWithdrawalBtcError1(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	//ok := input.ok
	//rk := input.rk
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.BtcToken
	chain := token.BtcToken

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tk.SetTokenInfo(ctx, tokenInfo)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	d0Amt := sdk.NewInt(10000000)
	hash0 := "opcu_utxo_deposit_0"
	d0, err := sdk.NewDepositItem(hash0, 0, d0Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	d1Amt := sdk.NewInt(80000000)
	hash1 := "opcu_utxo_deposit_1"
	d1, err := sdk.NewDepositItem(hash1, 0, d1Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	require.Nil(t, err)
	deposiList := sdk.DepositList{d0, d1}.Sort()

	ck.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), deposiList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, d0Amt.Add(d1Amt))))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	require.Equal(t, d0Amt.Add(d1.Amount), opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, deposiList, ck.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	amt := sdk.NewInt(80000000)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	err = user1CU.AddAsset(symbol, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", 1)
	ck.SetCU(ctx, user1CU)
	toAddr := "mnRw8TRyxUVEv1CnfzpahuRr5BeWYsCGES"

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	/*Withdrawal*/
	user1WithdrawalOrderID1 := uuid.NewV1().String()
	user1WithdrawalOrderID2 := uuid.NewV1().String()
	user1WithdrawalOrderID3 := uuid.NewV1().String()
	user1WithdrawalOrderID4 := uuid.NewV1().String()
	user1WithdrawalOrderID5 := uuid.NewV1().String()
	user1WithdrawalOrderID6 := uuid.NewV1().String()
	user1WithdrawalOrderID7 := uuid.NewV1().String()

	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddress).Return(true, opCUBtcAddress)

	result := keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID1, symbol, sdk.NewInt(1000001), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID2, symbol, sdk.NewInt(1000002), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID3, symbol, sdk.NewInt(1000003), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID4, symbol, sdk.NewInt(1000004), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID5, symbol, sdk.NewInt(1000005), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID6, symbol, sdk.NewInt(1000006), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID7, symbol, sdk.NewInt(1000007), sdk.NewInt(10000))
	require.Equal(t, sdk.CodeOK, result.Code)

	/*WithdrawalWaitSign*/
	rawData := []byte("rawData")
	signHash1 := []byte("signHash1")

	result = keeper.WithdrawalWaitSign(ctx, btcOPCUAddr, []string{user1WithdrawalOrderID1, user1WithdrawalOrderID2, user1WithdrawalOrderID3, user1WithdrawalOrderID4, user1WithdrawalOrderID5, user1WithdrawalOrderID6, user1WithdrawalOrderID7}, []string{string(signHash1)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "contains too many vouts")
}

/*
opcu has 90000000 eth,
user1 has 80000000 eth
user1 withdrawal 60000000, eth with gas = 1300000
settle suggest 1000000 eth as gas in withdrawalwaitsign
actual used 800000 eth as gas in withdrawalfinish

*/

func TestWithdrawalEthSuccess(t *testing.T) {
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

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tokenInfo.GasLimit = sdk.NewInt(10000)
	tokenInfo.GasPrice = sdk.NewInt(100)
	tk.SetTokenInfo(ctx, tokenInfo)

	//fromCUAddr, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	//require.Nil(t, err)

	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	opCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = opCU.SetAssetAddress(symbol, opCUEthAddress, 1)
	require.Nil(t, err)
	//deposiList := sdk.DepositList{d0, d1}.Sort()

	//ck.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), deposiList)
	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU1.GetAssetCoins().AmountOf(symbol))
	//require.Equal(t, deposiList, ck.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)

	amt := sdk.NewInt(80000000)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	err = user1CU.AddAsset(symbol, "0x81b7e08f65bdf5648606c89998a9cc8164397647", 1)

	ck.SetCU(ctx, user1CU)

	withdrawalToAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, "0x81b7e08f65bdf5648606c89998a9cc8164397647", ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	mockCN.On("ValidAddress", chain, symbol, withdrawalToAddr).Return(true, withdrawalToAddr)
	mockCN.On("ValidAddress", chain, symbol, opCUEthAddress).Return(true, opCUEthAddress)

	//Step1, Withdrawal
	ctx = ctx.WithBlockHeight(11)
	orderID := uuid.NewV1().String()
	withdrawalAmt := sdk.NewInt(60000000)
	gasFee := sdk.NewInt(1300000)
	result := keeper.Withdrawal(ctx, user1CUAddr, withdrawalToAddr, orderID, symbol, withdrawalAmt, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	o := ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusBegin, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, withdrawalAmt, o.(*sdk.OrderWithdrawal).Amount)
	require.Equal(t, gasFee, o.(*sdk.OrderWithdrawal).GasFee)
	require.Equal(t, sdk.ZeroInt(), o.(*sdk.OrderWithdrawal).CostFee)
	require.Equal(t, withdrawalToAddr, o.(*sdk.OrderWithdrawal).WithdrawToAddress)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt.Category)
	require.Equal(t, 3, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, user1CUAddr, of.CUAddress)

	wf, v := receipt.Flows[1].(sdk.WithdrawalFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr.String(), wf.FromCu)
	require.Equal(t, withdrawalToAddr, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, withdrawalAmt, wf.Amount)
	require.Equal(t, gasFee, wf.GasFee)
	require.Equal(t, orderID, wf.OrderID)

	bf, v := receipt.Flows[2].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf.Symbol)
	require.Equal(t, amt, bf.PreviousBalance)
	require.Equal(t, withdrawalAmt.Add(gasFee).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, withdrawalAmt.Add(gasFee), bf.BalanceOnHoldChange)

	//Check user1 coins and coinsHold
	user1CU1 := ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, withdrawalAmt.Add(gasFee), user1CU1.GetCoinsHold().AmountOf(symbol))

	// confirm withdraw
	for i := 0; i < 3; i++ {
		result := keeper.WithdrawalConfirm(ctx, sdk.CUAddress(validators[i].GetOperator()), orderID, true)
		require.Equalf(t, sdk.CodeOK, result.Code, "i:%d, log:%s", i, result.Log)
	}

	//Step2, WithdrawalWaitSign
	withdrawalTxHash := "withdrawalTxHash"
	chainnodewithdrawalTx := chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddress,
		To:       withdrawalToAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(10000),
		GasPrice: sdk.NewInt(100),
	}
	suggestGasFee := sdk.NewInt(10000).MulRaw(100)

	rawData := []byte("rawData")
	signHash := "signHash"
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodewithdrawalTx, []byte(signHash), nil)

	ctx = ctx.WithBlockHeight(20)
	result1 := keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{orderID}, []string{signHash}, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, suggestGasFee, o.(*sdk.OrderWithdrawal).CostFee)
	require.Equal(t, withdrawalAmt, o.(*sdk.OrderWithdrawal).Amount)
	require.Equal(t, gasFee, o.(*sdk.OrderWithdrawal).GasFee)
	require.Equal(t, withdrawalToAddr, o.(*sdk.OrderWithdrawal).WithdrawToAddress)

	//check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, v := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, user1CUAddr, of1.CUAddress)

	wwf, v := receipt1.Flows[1].(sdk.WithdrawalWaitSignFlow)
	require.True(t, v)
	require.Equal(t, []string{orderID}, wwf.OrderIDs)
	require.Equal(t, rawData, wwf.RawData)
	require.Equal(t, ethOPCUAddr.String(), wwf.OpCU)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, withdrawalAmt.Add(gasFee), user1CU1.GetCoinsHold().AmountOf(symbol))

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, withdrawalAmt.Add(suggestGasFee), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCUEthAmt.Sub(withdrawalAmt).Sub(suggestGasFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable := opCU.IsEnabledSendTx(chain, opCUEthAddress)
	require.Equal(t, false, sendable)

	//Step3, WithdrawalSignFinish
	signedData := []byte("signedData")
	chainnodewithdrawalTx.BlockHeight = 10000 //eth height = 10000

	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCUEthAddress, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodewithdrawalTx, nil).Once()

	result2 := keeper.WithdrawalSignFinish(ctx, []string{orderID}, signedData, "")
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, rawData, o.(*sdk.OrderWithdrawal).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderWithdrawal).SignedTx)
	require.Equal(t, withdrawalTxHash, o.(*sdk.OrderWithdrawal).Txhash)

	//check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt2.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of2, v := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, user1CUAddr, of2.CUAddress)

	wsf, v := receipt2.Flows[1].(sdk.WithdrawalSignFinishFlow)
	require.True(t, v)
	require.Equal(t, []string{orderID}, wsf.OrderIDs)
	require.Equal(t, signedData, wsf.SignedTx)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, withdrawalAmt.Add(gasFee), user1CU1.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasReceived())

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, withdrawalAmt.Add(suggestGasFee), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCUEthAmt.Sub(withdrawalAmt).Sub(suggestGasFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUEthAddress)
	require.Equal(t, false, sendable)
	require.Equal(t, sdk.Coins(nil), opCU.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), opCU.GetGasReceived())

	//Step4, WithdrawalFinish
	costFee := sdk.NewInt(800000)
	chainnodewithdrawalTx.CostFee = costFee
	chainnodewithdrawalTx.Status = chainnode.StatusSuccess
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodewithdrawalTx, nil).Once()

	result3 := keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{orderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{orderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	result3 = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{orderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeWithdrawal, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, user1CUAddr, o.GetCUAddress())
	require.Equal(t, ethOPCUAddr.String(), o.(*sdk.OrderWithdrawal).OpCUaddress)
	require.Equal(t, rawData, o.(*sdk.OrderWithdrawal).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderWithdrawal).SignedTx)
	require.Equal(t, gasFee, o.(*sdk.OrderWithdrawal).GasFee)
	require.Equal(t, costFee, o.(*sdk.OrderWithdrawal).CostFee)
	require.Equal(t, withdrawalTxHash, o.(*sdk.OrderWithdrawal).Txhash)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeWithdrawal, receipt3.Category)
	require.Equal(t, 3, len(receipt3.Flows))

	of3, v := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeWithdrawal, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, ethOPCUAddr, of3.CUAddress)

	wff, v := receipt3.Flows[1].(sdk.WithdrawalFinishFlow)
	require.True(t, v)
	require.Equal(t, []string{orderID}, wff.OrderIDs)
	require.Equal(t, costFee, wff.CostFee)

	bf3, v := receipt3.Flows[2].(sdk.BalanceFlow)
	require.True(t, v)
	require.Equal(t, user1CUAddr, bf3.CUAddress)
	require.Equal(t, sdk.Symbol(symbol), bf3.Symbol)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), bf3.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf3.BalanceChange)
	require.Equal(t, withdrawalAmt.Add(gasFee), bf3.PreviousBalanceOnHold)
	require.Equal(t, withdrawalAmt.Add(gasFee).Neg(), bf3.BalanceOnHoldChange)

	//Check user1 coins and coinsHold
	user1CU1 = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, amt.Sub(withdrawalAmt).Sub(gasFee), user1CU1.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU1.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasUsed())
	require.Equal(t, sdk.Coins(nil), user1CU1.GetGasReceived())

	//check OPCU's coins, coinsHold and status
	opCU = ck.GetCU(ctx, ethOPCUAddr)
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, opCUEthAmt.Sub(withdrawalAmt).Sub(costFee), opCU.GetAssetCoins().AmountOf(symbol))
	sendable = opCU.IsEnabledSendTx(chain, opCUEthAddress)
	require.Equal(t, true, sendable)
	require.Equal(t, costFee, opCU.GetGasUsed().AmountOf(chain))
	require.Equal(t, gasFee, opCU.GetGasReceived().AmountOf(chain))

}

/*
opcu has 90000000 eth,
user1 has 80000000 eth
user1 withdrawal 60000000 eth with gas = 10000000
settle suggest 2100000 eth as gas in withdrawalwaitsign
actual used 1400000 eth as gas in withdrawalfinish
finially, opCU gr = 10000000eth, gu:= 1400000

*/

func TestWithdrawalEthError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	//ok := input.ok
	//rk := input.rk
	ctx = ctx.WithBlockHeight(10)
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	validators := input.validators
	symbol := token.EthToken
	chain := token.EthToken

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.BtcToken)))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, token.EthToken)))

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tokenInfo.GasLimit = sdk.NewInt(21000)
	tk.SetTokenInfo(ctx, tokenInfo)

	//fromCUAddr, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	//require.Nil(t, err)
	//setup OpCU
	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	opCUEthAddr := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = opCU.SetAssetAddress(symbol, opCUEthAddr, 1)
	require.Nil(t, err)

	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU1 := ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU1.GetAssetCoins().AmountOf(symbol))

	//set UserCU
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	amt := sdk.NewInt(80000000)
	user1CU := ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(symbol, amt)))
	err = user1CU.AddAsset(symbol, "0x81b7e08f65bdf5648606c89998a9cc8164397647", 1)
	ck.SetCU(ctx, user1CU)

	toAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"

	require.Equal(t, amt, ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetCoinsHold().AmountOf(symbol))
	require.Equal(t, "0x81b7e08f65bdf5648606c89998a9cc8164397647", ck.GetCU(ctx, sdk.CUAddress(user1CUAddr)).GetAssetAddress(symbol, 1))

	/*Withdrawal*/
	user1WithdrawalOrderID := uuid.NewV1().String()
	//illegal uuid
	result := keeper.Withdrawal(ctx, user1CUAddr, toAddr, "illegaleOrderID", symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//from CU not exist
	cuAddr, _ := sdk.CUAddressFromBase58("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe")
	result = keeper.Withdrawal(ctx, cuAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	//from CU is not a user CU
	result = keeper.Withdrawal(ctx, ethOPCUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "withdrawal from a non user CU")

	//upsupport token
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, "fcoin", sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeUnsupportToken, result.Code)

	//token's sendenable is false
	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//token withdrawal enable is false
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//tranfer's sendenable is false
	tokenInfo.IsWithdrawalEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)
	keeper.SetSendEnabled(ctx, false)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//toaddress is not valid
	keeper.SetSendEnabled(ctx, true)
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(false, "").Once()
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInvalidAddress, result.Code)
	require.Contains(t, result.Log, "is not a valid address")

	//gasFee LT WithdrawalFeeRate
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(69999000), sdk.NewInt(1000))
	require.Equal(t, sdk.CodeInsufficientFee, result.Code)

	//amt is negative
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(60000000).Neg(), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeAmountError, result.Code)

	//amt is zero
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.ZeroInt(), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeAmountError, result.Code)

	//withdrawal more than owned coins
	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, sdk.NewInt(70000001), sdk.NewInt(10000000))
	require.Equal(t, sdk.CodeInsufficientCoins, result.Code)

	//every things is ok
	withdrawalAmt := sdk.NewInt(60000000)
	gasFee := sdk.NewInt(10000000)
	gasPrice := sdk.NewInt(100)
	costFee := sdk.NewInt(1400000)

	result = keeper.Withdrawal(ctx, user1CUAddr, toAddr, user1WithdrawalOrderID, symbol, withdrawalAmt, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	user1CU = ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	require.Equal(t, sdk.NewInt(10000000), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(70000000), user1CU.GetCoinsHold().AmountOf(symbol))

	// confirm withdraw
	for i := 0; i < 3; i++ {
		result := keeper.WithdrawalConfirm(ctx, sdk.CUAddress(validators[i].GetOperator()), user1WithdrawalOrderID, true)
		require.Equalf(t, sdk.CodeOK, result.Code, "i:%d, log:%s", i, result.Log)
	}

	/*WithdrawalWaitSign*/
	rawData := []byte("rawData")
	signHash := []byte("signHash")
	withdrawalTxHash := "withdrawalTxHash"
	chainnodeWithdrawalTx := chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}

	//len(signHash) != 1
	mockCN.On("ValidAddress", chain, symbol, opCUEthAddr).Return(true, opCUEthAddr)
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash), "noexisthhash"}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "AccountBased token supports only one withdrawal at one time")

	//QuerAccountTransactionFromData fail
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, errors.New("QueryAccountTransactionFromDataError")).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "QueryAccountTransactionFromDataError")

	//toAddr mismatch
	chainnodeWithdrawalTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       "toAddr mismatch",
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Unexpected withdrawal to address")

	//amount mismatch
	chainnodeWithdrawalTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   withdrawalAmt.SubRaw(1),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Unexpected withdrawal Amount")

	//signhash mismatch
	chainnodeWithdrawalTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    1,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, []byte("signhash mismatch"), nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "hash mismatch")

	//gaslimit mismatch
	chainnodeWithdrawalTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000).AddRaw(1),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas limit mismatch")

	//price is too high
	chainnodeWithdrawalTx = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}

	tokenInfo.GasPrice = gasPrice.MulRaw(10).QuoRaw(12)
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too high")

	//price is too low
	tokenInfo.GasPrice = gasPrice.MulRaw(10).QuoRaw(8).AddRaw(1)
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, nil).Once()
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too low")

	//everything is ok
	tokenInfo.GasPrice = gasPrice
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeWithdrawalTx, signHash, nil)
	result = keeper.WithdrawalWaitSign(ctx, ethOPCUAddr, []string{user1WithdrawalOrderID}, []string{string(signHash)}, rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	/*WithdrawalSignFinish*/
	signedTx := []byte("signedTx")
	chainnodeWithdrawalTx1 := chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}

	//VerifyAccountSignedTransaction err
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCUEthAddr, signedTx).Return(true, errors.New("VerifyAccountSignedTransactionError")).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "VerifyAccountSignedTransactionError")

	//VerifyAccountSignedTransaction,  verified = false
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCUEthAddr, signedTx).Return(false, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "VerifyAccountSignedTransaction fail")

	//QueryAccountTransactionFromSignedData, err
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCUEthAddr, signedTx).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, errors.New("QueryAccountTransactionFromSignedData")).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "QueryAccountTransactionFromSignedData Error")

	//toAddr mismatch
	chainnodeWithdrawalTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       "toAddr mismatch",
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "Unexpected withdrawal to address")

	//amount mismatch
	chainnodeWithdrawalTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   withdrawalAmt.AddRaw(1),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//	require.Contains(t, result.Log, "Unexpected withdrawal Amount")

	//gasPrice mimatch
	chainnodeWithdrawalTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice.AddRaw(1),
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "gas price mismatch")

	//gasLimit mismatch
	chainnodeWithdrawalTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     opCUEthAddr,
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000).SubRaw(1),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "gas limit mismatch")

	//from address mismatch
	chainnodeWithdrawalTx1 = chainnode.ExtAccountTransaction{
		Hash:     withdrawalTxHash,
		From:     "fromAddr mismatch",
		To:       toAddr,
		Amount:   withdrawalAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: gasPrice,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, nil).Once()
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "from addr mismatch")

	//everything is ok
	chainnodeWithdrawalTx1.From = opCUEthAddr
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeWithdrawalTx1, nil)
	result = keeper.WithdrawalSignFinish(ctx, []string{user1WithdrawalOrderID}, signedTx, "")
	require.Equal(t, sdk.CodeOK, result.Code)

	/*WithdrawalFinish*/

	//1st confirm
	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	//2nd confirm
	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	//3rd confirm
	result = keeper.WithdrawalFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{user1WithdrawalOrderID}, costFee, true)
	require.Equal(t, sdk.CodeOK, result.Code)

	//Check user1CU coins
	user1CU = ck.GetCU(ctx, sdk.CUAddress(user1CUAddr))
	require.Equal(t, sdk.NewInt(10000000), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoinsHold().AmountOf(symbol))

	//check opCU's coins, coinsonhold, deposit items
	opCU = ck.GetCU(ctx, sdk.CUAddress(ethOPCUAddr))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(28600000), opCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), opCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(10000000), opCU.GetGasReceived().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(1400000), opCU.GetGasUsed().AmountOf(symbol))
	sendable := opCU.IsEnabledSendTx(chain, opCUEthAddr)
	require.True(t, sendable)

}

func TestCheckWithdrawalOrders(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	mockCN = chainnode.MockChainnode{}

	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	user1CU := ck.GetCU(ctx, user1CUAddr)
	require.NotNil(t, user1CU)
	user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	user1BtcAddr := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	user1CU.SetAssetAddress(token.EthToken, user1EthAddr, 1)
	user1CU.SetAssetAddress(token.BtcToken, user1BtcAddr, 1)
	user1CU.SetCoinsHold(sdk.NewCoins(sdk.NewCoin(token.BtcToken, sdk.NewInt(10000000))))
	ck.SetCU(ctx, user1CU)

	user1BtcOrderID1 := uuid.NewV1().String()
	order := &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID1,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    token.BtcToken,
		},
		Amount:            sdk.NewInt(10000000).SubRaw(100000),
		GasFee:            sdk.NewInt(100000),
		WithdrawToAddress: "0x81b7E08F65Bdf5648606c89998A9CC8164397647",
		Txhash:            "txHash",
		RawData:           []byte("rawData"),
		SignedTx:          []byte("signedTx"),
		WithdrawStatus:    sdk.WithdrawStatusValid,
	}

	user1EthOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1EthOrder1)

	//Non exist order
	_, _, sdkErr := keeper.CheckWithdrawalOrders(ctx, []string{uuid.NewV1().String()}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeNotFoundOrder, sdkErr.Code())

	//not a withdrawal order
	user1NonWithdrawalID := uuid.NewV1().String()
	collectOrder := &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1NonWithdrawalID,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    token.BtcToken,
		},
		Amount:   sdk.NewInt(10000000).SubRaw(100000),
		Txhash:   "txHash",
		RawData:  []byte("rawData"),
		SignedTx: []byte("signedTx"),
	}

	user1NonWithdrawalOrder1 := ok.NewOrder(ctx, collectOrder)
	ok.SetOrder(ctx, user1NonWithdrawalOrder1)

	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1NonWithdrawalID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())

	//symbol is illegal
	user1ErrOrderID := uuid.NewV1().String()
	order = &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1ErrOrderID,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    "Btc",
		},
		Amount:            sdk.NewInt(10000000).SubRaw(100000),
		GasFee:            sdk.NewInt(100000),
		WithdrawToAddress: "0x81b7E08F65Bdf5648606c89998A9CC8164397647",
		Txhash:            "txHash",
		RawData:           []byte("rawData"),
		SignedTx:          []byte("signedTx"),
		WithdrawStatus:    sdk.WithdrawStatusValid,
	}
	user1ErrOrder := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidSymbol, sdkErr.Code())

	//symbol is not support by token
	user1ErrOrder.(*sdk.OrderWithdrawal).Symbol = "notsupport"
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeUnsupportToken, sdkErr.Code())

	//transferkeeper's sendable is false
	keeper.SetSendEnabled(ctx, false)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "ithdrawal is not enabled temporary")

	//token's sendable is false
	keeper.SetSendEnabled(ctx, true)
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "withdrawal is not enabled temporary")

	//token's withdrawalenble is false
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "withdrawal is not enabled temporary")

	//duplicated order ID is
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = true
	tokenInfo.IsDepositEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1BtcOrderID1}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())

	//the second order does not exist
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, uuid.NewV1().String()}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeNotFoundOrder, sdkErr.Code(), sdkErr)

	//the second order is not a withdrawal order
	user1CollectOrderID := uuid.NewV1().String()
	order1 := &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1CollectOrderID,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    "btc",
		},
		Amount: sdk.NewInt(100),
	}
	user1CollectOrder := ok.NewOrder(ctx, order1)
	ok.SetOrder(ctx, user1CollectOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1CollectOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "is not a withdrawal order")

	//the second order status is not expected
	user1ErrOrder.(*sdk.OrderWithdrawal).Symbol = token.BtcToken
	user1ErrOrder.SetOrderStatus(sdk.OrderStatusWaitSign)
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "status does not match expctedStatus")

	//the second order symbol is not expected
	user1ErrOrder.(*sdk.OrderWithdrawal).Symbol = token.EthToken
	user1ErrOrder.SetOrderStatus(sdk.OrderStatusBegin)
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "symbol mismatch")

	//the second order hash is not expceted
	user1ErrOrder.(*sdk.OrderWithdrawal).Symbol = token.BtcToken
	user1ErrOrder.(*sdk.OrderWithdrawal).Txhash = "txHash1"
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "hash mismatch")

	//the second order rawData is not expceted
	user1ErrOrder.(*sdk.OrderWithdrawal).Txhash = "txHash"
	user1ErrOrder.(*sdk.OrderWithdrawal).RawData = []byte("rawData1")
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "rawData mismatch")

	//the second order signTx is not expceted
	user1ErrOrder.(*sdk.OrderWithdrawal).RawData = []byte("rawData")
	user1ErrOrder.(*sdk.OrderWithdrawal).SignedTx = []byte("signTx1")
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "signTx mismatch")

	//the second user CU does not exist
	user1ErrOrder.(*sdk.OrderWithdrawal).SignedTx = []byte("signedTx")
	cuAddr, _ := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	user1ErrOrder.(*sdk.OrderWithdrawal).CUAddress = cuAddr
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "CU does not exist")

	//the second user CU is not user type cu
	cuAddr, _ = sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	user1ErrOrder.(*sdk.OrderWithdrawal).CUAddress = cuAddr
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, sdkErr = keeper.CheckWithdrawalOrders(ctx, []string{user1BtcOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "CU type is not user type")

}

func TestCheckWithdrawalOpCU(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	chain := token.BtcToken
	symbol := token.BtcToken
	mockCN = chainnode.MockChainnode{}

	//setup btc opcu
	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	opCU := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	opCUBtcAddr := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	err = opCU.SetAssetAddress(symbol, opCUBtcAddr, 1)
	ck.SetCU(ctx, opCU)

	//cu does not exist
	cuAddr, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	cu := ck.GetCU(ctx, cuAddr)
	require.Nil(t, err)
	sdkErr := keeper.CheckWithdrawalOpCU(ctx, cu, chain, symbol, true, opCUBtcAddr)
	require.Equal(t, sdk.CodeInvalidAccount, sdkErr.Code())

	//cu is not a optype
	cuAddr, err = sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	cu = ck.GetCU(ctx, cuAddr)
	sdkErr = keeper.CheckWithdrawalOpCU(ctx, cu, chain, symbol, true, opCUBtcAddr)
	require.Equal(t, sdk.CodeInvalidAccount, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "is not a OPCU")

	//symbol mismatch
	cuAddr, err = sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	cu = ck.GetCU(ctx, cuAddr)
	sdkErr = keeper.CheckWithdrawalOpCU(ctx, cu, chain, symbol, true, opCUBtcAddr)
	require.Equal(t, sdk.CodeInvalidAccount, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "symbol mismatch, expected")

	//valid = false
	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddr).Return(false, "").Once()
	cuAddr, err = sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	cu = ck.GetCU(ctx, cuAddr)
	sdkErr = keeper.CheckWithdrawalOpCU(ctx, cu, chain, symbol, true, opCUBtcAddr)
	require.Equal(t, sdk.CodeInvalidAddress, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "is not a valid address")

	//lock status check
	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddr).Return(true, opCUBtcAddr)
	ctx = ctx.WithBlockHeight(10)
	opCU = ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	opCU.SetEnableSendTx(false, chain, opCUBtcAddr)
	ck.SetCU(ctx, opCU)
	sdkErr = keeper.CheckWithdrawalOpCU(ctx, opCU, chain, symbol, false, opCUBtcAddr)
	require.Nil(t, sdkErr)

	opCU.SetEnableSendTx(true, chain, opCUBtcAddr)
	ck.SetCU(ctx, opCU)
	sdkErr = keeper.CheckWithdrawalOpCU(ctx, opCU, chain, symbol, true, opCUBtcAddr)
	require.Nil(t, sdkErr)

}

func TestCheckDecodedUtxoTransaction(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	ctx = ctx.WithBlockHeight(10)
	mockCN = chainnode.MockChainnode{}
	symbol := token.BtcToken
	chain := token.BtcToken

	//setup token
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.WithdrawalFeeRate = sdk.NewDecWithPrec(1, 2)
	tk.SetTokenInfo(ctx, tokenInfo)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	d0Amt := sdk.NewInt(10000000)
	hash0 := "opcu_utxo_deposit_0"
	d0, err := sdk.NewDepositItem(hash0, 0, d0Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)
	d1Amt := sdk.NewInt(80000000)
	hash1 := "opcu_utxo_deposit_1"
	d1, err := sdk.NewDepositItem(hash1, 0, d1Amt, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
	require.Nil(t, err)

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := ck.GetCU(ctx, sdk.CUAddress(btcOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	require.Nil(t, err)
	deposiList := sdk.DepositList{d0, d1}.Sort()

	ck.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), deposiList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, d0Amt.Add(d1Amt))))
	ck.SetCU(ctx, opCU)

	//setup userCUs
	user1CUAddr, _ := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	user1CU := ck.GetCU(ctx, user1CUAddr)
	user1CU.SetAssetAddress(symbol, "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP", 1)

	user2CUAddr, _ := sdk.CUAddressFromBase58("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q")
	user2CU := ck.GetCU(ctx, user2CUAddr)
	user2CU.SetAssetAddress(symbol, "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", 1)

	//setup withdrawal orders
	user1BtcOrderID1 := uuid.NewV1().String()
	order := &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID1,
			OrderType: sdk.OrderTypeWithdrawal,
			Status:    sdk.OrderStatusBegin,
			Symbol:    token.BtcToken,
		},
		Amount:            sdk.NewInt(6000000),
		GasFee:            sdk.NewInt(10000),
		WithdrawToAddress: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP",
	}
	user1BtcOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder1)
	require.NotNil(t, ok.GetOrder(ctx, user1BtcOrderID1))
	user1BtcWithdrawalOrder1 := ok.GetOrder(ctx, user1BtcOrderID1).(*sdk.OrderWithdrawal)

	user2BtcOrderID1 := uuid.NewV1().String()
	order = &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user2CUAddr,
			ID:        user2BtcOrderID1,
			OrderType: sdk.OrderTypeWithdrawal,
			Status:    sdk.OrderStatusBegin,
			Symbol:    token.BtcToken,
		},
		Amount:            sdk.NewInt(3000000),
		GasFee:            sdk.NewInt(10000),
		WithdrawToAddress: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9",
	}

	user2BtcOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user2BtcOrder1)
	require.NotNil(t, ok.GetOrder(ctx, user2BtcOrderID1))
	user2BtcWithdrawalOrder1 := ok.GetOrder(ctx, user2BtcOrderID1).(*sdk.OrderWithdrawal)

	//vin address mimatch
	tx := &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: hash1, Index: 0, Amount: d1Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gk"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(1000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		CostFee: sdk.NewInt(100),
	}

	_, sdkErr := keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Equal(t, sdk.CodeInvalidTx, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "Unexpected Vin address")

	//non exist deposit
	tx = &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "noexist", Index: 0, Amount: d1Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(1000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		CostFee: sdk.NewInt(100),
	}

	_, sdkErr = keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Equal(t, sdk.CodeUnknownUtxo, sdkErr.Code())

	// utxo's total amount exceed assetcoins
	tx = &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(900000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		CostFee: sdk.NewInt(100000),
	}

	opCU.SubAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, sdk.NewInt(1).Add(d1Amt))))
	ck.SetCU(ctx, opCU)
	_, sdkErr = keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Equal(t, sdk.CodeInsufficientCoins, sdkErr.Code())

	//	has more than 1 change back, ok
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, sdk.NewInt(1).Add(d1Amt))))
	ck.SetCU(ctx, opCU)

	tx = &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(500000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Amount: sdk.NewInt(400000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		CostFee: sdk.NewInt(100000),
	}
	_, sdkErr = keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Nil(t, sdkErr)

	//vout amount mismatch
	tx = &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000001), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(900000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		CostFee: sdk.NewInt(100000),
	}
	_, sdkErr = keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Equal(t, sdk.CodeInvalidTx, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "Unexpected Vout Amount")

	//vout address mismatch
	tx = &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(900000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gk"},
		},
		CostFee: sdk.NewInt(100000),
	}
	_, sdkErr = keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Equal(t, sdk.CodeInvalidTx, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "Unexpected Changeback address")

	//calculate Fee mismatch
	tx = &chainnode.ExtUtxoTransaction{
		Vins: []*sdk.UtxoIn{
			{Hash: hash0, Index: 0, Amount: d0Amt, Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Amount: sdk.NewInt(6000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Amount: sdk.NewInt(3000000), Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"},
			{Amount: sdk.NewInt(900000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		CostFee: sdk.NewInt(100001),
	}
	_, sdkErr = keeper.CheckDecodedUtxoTransaction(ctx, chain, symbol, btcOPCUAddr, []*sdk.OrderWithdrawal{user1BtcWithdrawalOrder1, user2BtcWithdrawalOrder1}, tx, opCUBtcAddress)
	require.Equal(t, sdk.CodeInvalidTx, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "Unexpected Gas")
}
