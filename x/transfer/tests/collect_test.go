package tests

import (
	"errors"
	"testing"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestColletEthSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.GasLimit = sdk.NewInt(21000)
	tokenInfo.GasPrice = sdk.NewInt(10000000000)
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	//step1, deposit 2 ETH items
	chain := token.EthToken
	symbol := chain
	//deposit item1
	hash1 := "0x84ed75bfad4b6d1c405a123990a1750974aa1f053394d442dfbc76090eeed44a"
	fromAddr1 := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	toAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"

	memo1 := "nice to see u"
	amt1 := sdk.TokensFromConsensusPower(1)
	orderID1 := uuid.NewV1().String()

	//deposit item2
	hash2 := "0x2ac020bb869c2f3fc404b799ec38338b25c3bd1a2438d541f008101e8bb40dc0"

	memo2 := "see u again"
	amt2 := sdk.TokensFromConsensusPower(2)
	orderID2 := uuid.NewV1().String()

	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	mockCN.On("ValidAddress", chain, symbol, fromAddr1).Return(true, fromAddr1)

	toCU := ck.GetCU(ctx, toCUAddr)
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	toCU.SetAssetPubkey(pub.Bytes(), 1)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	err = toCU.AddAsset(symbol, toAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

	result1 := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(token.EthToken), toAddr, hash1, 0, amt1, orderID1, memo1)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	order := ok.GetOrder(ctx, orderID1)
	require.NotNil(t, order)
	require.Equal(t, orderID1, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, toCUAddr, order.GetCUAddress())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//no deposit item now
	dls1 := ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 0, len(dls1))

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID1, of.OrderID)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, toCUAddr, of.CUAddress)

	df, valid := receipt.Flows[1].(sdk.DepositFlow)
	require.True(t, valid)
	require.Equal(t, toCUAddr.String(), df.CuAddress)
	require.Equal(t, memo1, df.Memo)
	require.Equal(t, toAddr, df.Multisignedadress)
	require.Equal(t, symbol, df.Symbol)
	require.Equal(t, hash1, df.Txhash)
	require.Equal(t, orderID1, df.OrderID)
	require.Equal(t, amt1, df.Amount)
	require.Equal(t, uint64(0), df.Index)
	require.Equal(t, sdk.DepositTypeCU, df.DepositType)

	//check CU's coins and assetcoins
	toCU = ck.GetCU(ctx, toCUAddr)
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))

	result2 := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(token.EthToken), toAddr, hash2, 0, amt2, orderID2, memo2)
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check order
	order = ok.GetOrder(ctx, orderID2)
	require.NotNil(t, order)
	require.Equal(t, orderID2, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, toCUAddr, order.GetCUAddress())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//no deposit item now
	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 0, len(dls1))

	//check receipt
	receipt, err = rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid = receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID2, of.OrderID)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, toCUAddr, of.CUAddress)

	df, valid = receipt.Flows[1].(sdk.DepositFlow)
	require.True(t, valid)
	require.Equal(t, toCUAddr.String(), df.CuAddress)
	require.Equal(t, memo2, df.Memo)
	require.Equal(t, toAddr, df.Multisignedadress)
	require.Equal(t, symbol, df.Symbol)
	require.Equal(t, hash2, df.Txhash)
	require.Equal(t, orderID2, df.OrderID)
	require.Equal(t, amt2, df.Amount)
	require.Equal(t, uint64(0), df.Index)
	require.Equal(t, sdk.DepositTypeCU, df.DepositType)

	//check CU's coins and assetcoins
	toCU = ck.GetCU(ctx, toCUAddr)
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))

	//1st confirm
	result := keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{orderID1, orderID2}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	order = ok.GetOrder(ctx, orderID1)
	require.NotNil(t, order)
	require.Equal(t, orderID1, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//2nd confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[1].OperatorAddress), []string{orderID1, orderID2}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	order = ok.GetOrder(ctx, orderID2)
	require.NotNil(t, order)
	require.Equal(t, orderID2, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//3rd confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[2].OperatorAddress), []string{orderID1, orderID2}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	order = ok.GetOrder(ctx, orderID1)
	require.NotNil(t, order)
	require.Equal(t, orderID1, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	order = ok.GetOrder(ctx, orderID2)
	require.NotNil(t, order)
	require.Equal(t, orderID2, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	//check receipt after deposit confirmed
	receipt, err = rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 4, len(receipt.Flows))

	of, valid = receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	//require.Equal(t, orderID2, of.OrderID)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of.OrderStatus)
	//require.Equal(t, toCUAddr, of.CUAddress)

	cf, valid := receipt.Flows[1].(sdk.DepositConfirmedFlow)
	require.True(t, valid)
	require.Equal(t, 2, len(cf.ValidOrderIDs))
	require.Equal(t, 0, len(cf.InValidOrderIDs))
	require.Equal(t, orderID1, cf.ValidOrderIDs[0])
	require.Equal(t, orderID2, cf.ValidOrderIDs[1])

	bf, valid := receipt.Flows[2].(sdk.BalanceFlow)
	require.True(t, valid)
	require.Equal(t, toCUAddr, bf.CUAddress)
	require.Equal(t, symbol, bf.Symbol.String())
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf, valid = receipt.Flows[3].(sdk.BalanceFlow)
	require.True(t, valid)
	require.Equal(t, toCUAddr, bf.CUAddress)
	require.Equal(t, symbol, bf.Symbol.String())
	require.Equal(t, amt2, bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID1, orderID2}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)

	_, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 2, len(dls1))
	require.Equal(t, hash1, dls1[1].Hash)
	require.Equal(t, amt1, dls1[1].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls1[1].Status)
	require.Equal(t, memo1, dls1[1].Memo)
	require.Equal(t, uint64(0), dls1[1].Index)

	depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash1, 0)
	require.Equal(t, dls1[1], depositItem)

	depositItem = ck.GetDeposit(ctx, symbol, toCUAddr, hash2, 0)
	require.Equal(t, dls1[0], depositItem)

	//check deposit item
	dls := ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 2, len(dls))
	require.Equal(t, hash1, dls[1].Hash)
	require.Equal(t, amt1, dls[1].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls[1].Status)
	require.Equal(t, memo1, dls[1].Memo)
	require.Equal(t, uint64(0), dls[1].Index)
	require.Equal(t, hash2, dls[0].Hash)
	require.Equal(t, amt2, dls[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls[0].Status)
	require.Equal(t, memo2, dls[0].Memo)
	require.Equal(t, uint64(0), dls[0].Index)

	//check CU's coins and assetcoins
	toCU = ck.GetCU(ctx, toCUAddr)
	require.Equal(t, amt1.Add(amt2), toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))

	//step2, CollectWaitSign
	gas := sdk.NewInt(21000).Mul(sdk.NewInt(10000000000))
	gasOriginal := gas
	collectFromCUAddr := toCUAddr
	collectFromCU := toCU
	collectFromAddr := toAddr
	collectFromCU.AddGasReceived(sdk.NewCoins(sdk.NewCoin(chain, gas)))
	ck.SetCU(ctx, collectFromCU)

	collectToCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	collectToCU := ck.GetCU(ctx, collectToCUAddr)
	require.Nil(t, err)
	collectToAddr := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	err = collectToCU.SetAssetAddress(symbol, collectToAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, collectToCU)
	rawData := []byte("rawdata")
	collectTxHash := "collectTxHash"

	collectAmt := amt1.Add(amt2).Sub(gas)

	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   collectAmt,
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10000000000), // 10Gwei
	}

	mockCN.On("ValidAddress", chain, symbol, collectToAddr).Return(true, collectToAddr)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil)
	result3 := keeper.CollectWaitSign(ctx, collectToCUAddr, []string{orderID1, orderID2}, rawData)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//check receipts
	receipt, err = rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeCollect, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))
	orderFlow, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderStatusWaitSign, orderFlow.OrderStatus)
	require.Equal(t, sdk.OrderTypeCollect, orderFlow.OrderType)

	cwf, valid := receipt.Flows[1].(sdk.CollectWaitSignFlow)
	require.True(t, valid)
	require.Equal(t, []string{orderID1, orderID2}, cwf.OrderIDs)
	require.Equal(t, rawData, cwf.RawData)

	//check collect orders
	collectOrder1 := ok.GetOrder(ctx, orderID1).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusWaitSign, collectOrder1.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, collectOrder1.GetOrderType())
	require.Equal(t, symbol, collectOrder1.GetSymbol())
	require.Equal(t, fromCUAddr, collectOrder1.GetCUAddress())
	require.Equal(t, sdk.NewInt(10000000000), collectOrder1.GasPrice)
	require.Equal(t, sdk.NewInt(21000), collectOrder1.GasLimit)
	require.Equal(t, collectToCUAddr, collectOrder1.CollectToCU)
	require.Equal(t, rawData, collectOrder1.RawData)

	collectOrder2 := ok.GetOrder(ctx, orderID2).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusWaitSign, collectOrder2.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, collectOrder2.GetOrderType())
	require.Equal(t, symbol, collectOrder2.GetSymbol())
	require.Equal(t, fromCUAddr, collectOrder2.GetCUAddress())
	require.Equal(t, sdk.NewInt(10000000000), collectOrder2.GasPrice)
	require.Equal(t, sdk.NewInt(21000), collectOrder2.GasLimit)
	require.Equal(t, collectToCUAddr, collectOrder2.CollectToCU)
	require.Equal(t, rawData, collectOrder2.RawData)

	//check deposit item
	dls = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, sdk.DepositItemStatusInProcess, dls[0].Status)
	require.Equal(t, sdk.DepositItemStatusInProcess, dls[1].Status)

	//check collectFromCU's coins
	collectFromCU = ck.GetCU(ctx, collectFromCUAddr)
	require.Equal(t, amt1.Add(amt2), collectFromCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetCoinsHold().AmountOf(symbol))

	//check collectToCU's coins
	collectToCU = ck.GetCU(ctx, collectToCUAddr)
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoinsHold().AmountOf(symbol))

	//step3, CollectSignFinish
	chainnodeCollectTx.From = collectFromAddr
	signedData := []byte("singedTx")
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, collectFromAddr, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()

	result4 := keeper.CollectSignFinish(ctx, []string{orderID1, orderID2}, signedData, "")
	require.Equal(t, sdk.CodeOK, result4.Code)

	//check receipts
	receipt4, err := rk.GetReceiptFromResult(&result4)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeCollect, receipt4.Category)
	require.Equal(t, 2, len(receipt4.Flows))
	orderFlow4, valid := receipt4.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderStatusSignFinish, orderFlow4.OrderStatus)
	require.Equal(t, sdk.OrderTypeCollect, orderFlow4.OrderType)

	sff, valid := receipt4.Flows[1].(sdk.CollectSignFinishFlow)
	require.True(t, valid)
	require.Equal(t, []string{orderID1, orderID2}, sff.OrderIDs)
	require.Equal(t, signedData, sff.SignedTx)

	//check collect orders
	collectOrder1 = ok.GetOrder(ctx, orderID1).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusSignFinish, collectOrder1.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, collectOrder1.GetOrderType())
	require.Equal(t, symbol, collectOrder1.GetSymbol())
	require.Equal(t, fromCUAddr, collectOrder1.GetCUAddress())
	require.Equal(t, sdk.NewInt(10000000000), collectOrder1.GasPrice)
	require.Equal(t, sdk.NewInt(21000), collectOrder1.GasLimit)
	require.Equal(t, collectToCUAddr, collectOrder1.CollectToCU)
	require.Equal(t, rawData, collectOrder1.RawData)
	require.Equal(t, signedData, collectOrder1.SignedTx)

	collectOrder2 = ok.GetOrder(ctx, orderID2).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusSignFinish, collectOrder2.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, collectOrder2.GetOrderType())
	require.Equal(t, symbol, collectOrder2.GetSymbol())
	require.Equal(t, fromCUAddr, collectOrder2.GetCUAddress())
	require.Equal(t, sdk.NewInt(10000000000), collectOrder2.GasPrice)
	require.Equal(t, sdk.NewInt(21000), collectOrder2.GasLimit)
	require.Equal(t, collectToCUAddr, collectOrder2.CollectToCU)
	require.Equal(t, rawData, collectOrder2.RawData)
	require.Equal(t, signedData, collectOrder2.SignedTx)

	//check deposit item
	dls = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, sdk.DepositItemStatusInProcess, dls[0].Status)
	require.Equal(t, sdk.DepositItemStatusInProcess, dls[1].Status)

	//check collectFromCU's coins
	collectFromCU = ck.GetCU(ctx, collectFromCUAddr)
	require.Equal(t, amt1.Add(amt2), collectFromCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetCoinsHold().AmountOf(symbol))

	//check collectToCU's coins
	collectToCU = ck.GetCU(ctx, collectToCUAddr)
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoinsHold().AmountOf(symbol))

	//step4, CollectFinish
	costFee := sdk.NewInt(10000000000).MulRaw(10000)
	chainnodeCollectTx.CostFee = costFee
	chainnodeCollectTx.Status = chainnode.StatusSuccess
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()
	mockCN.On("QueryAccountTransaction", chain, symbol, collectTxHash, mock.Anything).Return(&chainnodeCollectTx, nil)

	//1st confirm
	result5 := keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{orderID1, orderID2}, costFee)
	require.Equal(t, sdk.CodeOK, result5.Code)

	//2nd confirm
	result5 = keeper.CollectFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{orderID1, orderID2}, costFee)
	require.Equal(t, sdk.CodeOK, result5.Code)

	//3rd confirm
	result5 = keeper.CollectFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{orderID1, orderID2}, costFee)
	require.Equal(t, sdk.CodeOK, result5.Code)

	//check receipts
	receipt5, err := rk.GetReceiptFromResult(&result5)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeCollect, receipt5.Category)
	require.Equal(t, 2, len(receipt5.Flows))

	orderFlow5, valid := receipt5.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderStatusFinish, orderFlow5.OrderStatus)
	require.Equal(t, sdk.OrderTypeCollect, orderFlow5.OrderType)

	ff, valid := receipt5.Flows[1].(sdk.CollectFinishFlow)
	require.True(t, valid)
	require.Equal(t, []string{orderID1, orderID2}, ff.OrderIDs)
	require.Equal(t, costFee, ff.CostFee)

	//check collect orders
	collectOrder1 = ok.GetOrder(ctx, orderID1).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusFinish, collectOrder1.GetOrderStatus())
	collectOrder2 = ok.GetOrder(ctx, orderID2).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusFinish, collectOrder2.GetOrderStatus())

	//check deposit item
	dls = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, sdk.DepositItemStatusConfirmed, dls[0].Status)
	require.Equal(t, sdk.DepositItemStatusConfirmed, dls[1].Status)

	//check new generate deposit item
	collectDepositItem := ck.GetDeposit(ctx, symbol, collectToCUAddr, collectTxHash, 0)
	require.Equal(t, sdk.DepositItemStatusConfirmed, collectDepositItem.Status)
	require.Equal(t, collectAmt, collectDepositItem.Amount)

	//check collectFromCU's coins
	collectFromCU = ck.GetCU(ctx, collectFromCUAddr)
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, gas.Add(gasOriginal), collectFromCU.GetGasReceived().AmountOf(chain))
	require.Equal(t, costFee, collectFromCU.GetGasUsed().AmountOf(chain))

	//check collectToCU's coins
	collectToCU = ck.GetCU(ctx, collectToCUAddr)
	require.Equal(t, collectAmt, collectToCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoinsHold().AmountOf(symbol))

	require.Equal(t, collectAmt.Add(gas), amt1.Add(amt2))

	//4th confirm
	result5 = keeper.CollectFinish(ctx, sdk.CUAddress(validators[3].GetOperator()), []string{orderID1, orderID2}, costFee)
	require.Equal(t, sdk.CodeOK, result5.Code)
}

func TestColletEthError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	//rk := input.rk
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	validators := input.validators

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = true
	tokenInfo.IsDepositEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)
	chain := token.EthToken
	symbol := token.EthToken

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	require.Nil(t, err)

	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	user1CU := ck.GetCU(ctx, user1CUAddr)
	require.NotNil(t, user1CU)
	user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	user1BtcAddr := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	user1CU.SetAssetAddress(token.EthToken, user1EthAddr, 1)
	user1CU.SetAssetAddress(token.BtcToken, user1BtcAddr, 1)

	user1EthDeposit1, err := sdk.NewDepositItem("user1EthDeposit1", 0, sdk.NewInt(1000000000), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	user1EthDeposit2, err := sdk.NewDepositItem("user1EthDeposit2", 0, sdk.NewInt(2000000000), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(token.EthToken, sdk.NewInt(3000000000))))
	user1CU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(token.EthToken, sdk.NewInt(3000000000))))
	ck.SetCU(ctx, user1CU)

	ck.SetDepositList(ctx, token.EthToken, user1CUAddr, sdk.DepositList{user1EthDeposit1, user1EthDeposit2})
	user1EthDepositList := ck.GetDepositList(ctx, token.EthToken, user1CUAddr)
	require.Equal(t, 2, len(user1EthDepositList))

	user1EthOrderID1 := uuid.NewV1().String()
	order := &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1EthOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1EthAddr,
		Txhash:             user1EthDeposit1.Hash,
		Index:              user1EthDeposit1.Index,
		Amount:             user1EthDeposit1.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user1EthOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1EthOrder1)

	user1EthOrderID2 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1EthOrderID2,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1EthAddr,
		Txhash:             user1EthDeposit2.Hash,
		Index:              user1EthDeposit2.Index,
		Amount:             user1EthDeposit2.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user1EthOrder2 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1EthOrder2)

	user1BtcDeposit1, err := sdk.NewDepositItem("user1BtcDeposit1", 0, sdk.NewInt(3000000000), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	user1BtcDeposit2, err := sdk.NewDepositItem("user1BtcDeposit2", 0, sdk.NewInt(4000000000), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	ck.SetDepositList(ctx, token.BtcToken, user1CUAddr, sdk.DepositList{user1BtcDeposit1, user1BtcDeposit2})
	user1CU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(token.BtcToken, sdk.NewInt(7000000000))))
	ck.SetCU(ctx, user1CU)
	user1BtcDepositList := ck.GetDepositList(ctx, token.BtcToken, user1CUAddr)

	user1BtcOrderID1 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.BtcToken,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1BtcAddr,
		Txhash:             user1BtcDeposit1.Hash,
		Index:              user1BtcDeposit1.Index,
		Amount:             user1BtcDeposit1.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user1BtcOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder1)

	user1BtcOrderID2 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID2,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.BtcToken,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1BtcAddr,
		Txhash:             user1BtcDeposit2.Hash,
		Index:              user1BtcDeposit2.Index,
		Amount:             user1BtcDeposit2.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user1BtcOrder2 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder2)
	require.Equal(t, 2, len(user1BtcDepositList))
	require.Equal(t, user1BtcOrder1.GetID(), ok.GetOrder(ctx, user1BtcOrderID1).GetID())

	mockCN.On("SupportChain", symbol).Return(true)
	rawData := []byte("rawData")

	/*---CollectWaitSign----*/
	ctx = ctx.WithBlockHeight(10)
	//noexist CU
	noExistCUAddr, _ := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	result := keeper.CollectWaitSign(ctx, noExistCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	//CU is not opCU
	user2CUAddr, err := sdk.CUAddressFromBase58("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q")
	result = keeper.CollectWaitSign(ctx, user2CUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// opCU only support btc
	btcOpCUAddr, _ := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, " does not support symbol")

	//opCU doesnot have eth address
	ethOpCUAddr, _ := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, " does not have eth's address")

	//opCU's address is not valid
	collectToAddr := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	mockCN.On("ValidAddress", chain, symbol, collectToAddr).Return(false, "").Once()
	ethOpCU := ck.GetCU(ctx, ethOpCUAddr)
	err = ethOpCU.SetAssetAddress(symbol, collectToAddr, 1)
	ck.SetCU(ctx, ethOpCU)
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidAddress, result.Code, result)
	require.Contains(t, result.Log, "is not a valid address")

	////QueryAccountTransaction err
	collectTxHash := "collectTxHash"
	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       "notCollectToAddr",
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10),
	}

	mockCN.On("ValidAddress", chain, symbol, collectToAddr).Return(true, collectToAddr)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, errors.New("err")).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Fail to get transaction")

	//tx.To != collectToAddr
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "collet to an unexpected address")

	tokenInfo.GasLimit = sdk.NewInt(21000)
	tk.SetTokenInfo(ctx, tokenInfo)

	// gasLimit != tokenInfo.GasLimit
	chainnodeCollectTx = chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21001),
		GasPrice: sdk.NewInt(10), //
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas limit")

	//totalAmt < tx.Amout
	chainnodeCollectTx = chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).AddRaw(1),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), //
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInsufficientCoins, result.Code)

	// tx.Amount + gasFee > assetCoins
	user1CU.SubAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, sdk.NewInt(1))))
	ck.SetCU(ctx, user1CU)
	chainnodeCollectTx = chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), //
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// tx.Amount less than collect threshold
	user1CU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, sdk.NewInt(1))))
	ck.SetCU(ctx, user1CU)
	chainnodeCollectTx = chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), //
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Contains(t, result.Log, "less than threshold")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	tokenInfo.CollectThreshold = sdk.NewInt(3000000000)
	tk.SetTokenInfo(ctx, tokenInfo)

	//gas price too high
	tokenInfo.GasPrice = sdk.NewInt(6)
	tk.SetTokenInfo(ctx, tokenInfo)
	chainnodeCollectTx = chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), //
	}
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too high")

	//gas price is too low
	tokenInfo.GasPrice = sdk.NewInt(22)
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil).Once()
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too low")

	//everything is ok for waitsign
	chainnodeCollectTx = chainnode.ExtAccountTransaction{
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), //
	}
	tokenInfo.GasPrice = sdk.NewInt(10)
	tk.SetTokenInfo(ctx, tokenInfo)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(&chainnodeCollectTx, []byte{0x0}, nil)
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	user1CU = ck.GetCU(ctx, user1CUAddr)
	sendable := user1CU.IsEnabledSendTx(chain, user1CU.GetAssetAddress(chain, 1))
	require.Equal(t, false, sendable)

	/*---CollectSignFinish----*/
	ctx = ctx.WithBlockHeight(20)
	//VerifyAccountSignedTransaction, err
	chainnodeCollectTx.From = user1EthAddr
	signedTx := []byte("singedTx")
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, user1EthAddr, signedTx).Return(true, errors.New("err")).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Fail to verify signed transaction")

	//VerifyAccountSignedTransaction verified =false
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, user1EthAddr, signedTx).Return(false, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Fail to verify signed transaction")

	//QueryAccountTransactionFromSignedData, err
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, user1EthAddr, signedTx).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx, errors.New("err")).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "Fail to get transaction")

	//tx.From != fromAddr
	chainnodeCollectTx1 := chainnode.ExtAccountTransaction{
		From:     "not from Addrss",
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), // 10Gwei
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "collet from an unexpected address")

	//tx.To != toAddr
	chainnodeCollectTx1 = chainnode.ExtAccountTransaction{
		From:     user1EthAddr,
		Hash:     collectTxHash,
		To:       "not to collectAddress",
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), // 10Gwei
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//require.Contains(t, result.Log, "collet to an unexpected address")

	//amount mismatch
	chainnodeCollectTx1 = chainnode.ExtAccountTransaction{
		From:     user1EthAddr,
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(1), // 10Gwei
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "amount mismatch")

	//gasPrice != tx.GasPrice
	chainnodeCollectTx1 = chainnode.ExtAccountTransaction{
		From:     user1EthAddr,
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(11), // 10Gwei
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gasPrice mismatch")

	//gasLimit != tx.GasLimit
	chainnodeCollectTx1 = chainnode.ExtAccountTransaction{
		From:     user1EthAddr,
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(20000),
		GasPrice: sdk.NewInt(10), // 10Gwei
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gasLimit mismatch")

	//everything is ok for collectsignFinish
	chainnodeCollectTx1 = chainnode.ExtAccountTransaction{
		From:     user1EthAddr,
		Hash:     collectTxHash,
		To:       collectToAddr,
		Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
		Nonce:    0,
		GasLimit: sdk.NewInt(21000),
		GasPrice: sdk.NewInt(10), // 10Gwei
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1EthOrderID1, user1EthOrderID2}, signedTx, "")
	require.Equal(t, sdk.CodeOK, result.Code)

	/*---CollectFinish----*/
	ctx = ctx.WithBlockHeight(30)

	//chainnodeCollectTx2 := chainnode.ExtAccountTransaction{
	//	Status:   chainnode.StatusSuccess,
	//	From:     user1EthAddr,
	//	Hash:     collectTxHash,
	//	To:       collectToAddr,
	//	Amount:   user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount).SubRaw(210000),
	//	Nonce:    1,
	//	GasLimit: sdk.NewInt(21000),
	//	GasPrice: sdk.NewInt(10),
	//	CostFee:  sdk.NewInt(160000),
	//}

	//not a validator
	result = keeper.CollectFinish(ctx, fromCUAddr, []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "collect from not a validator")

	//empty order
	collectOrder1 := ok.GetOrder(ctx, user1EthOrderID1)
	ok.DeleteOrder(ctx, collectOrder1)
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeNotFoundOrder, result.Code)
	require.Contains(t, result.Log, "does not exist")

	//order status is mismatch
	orderStatus := collectOrder1.GetOrderStatus()
	collectOrder1.SetOrderStatus(sdk.OrderStatusFinish)
	ok.SetOrder(ctx, collectOrder1)
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeOK, result.Code)
	//	require.Contains(t, result.Log, "not as expected")

	collectOrder1.SetOrderStatus(orderStatus)
	ok.SetOrder(ctx, collectOrder1)

	//1st confirm
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeOK, result.Code)

	//2nd confirm
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[1].OperatorAddress), []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeOK, result.Code)

	//3rd confirm
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedTx).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[2].OperatorAddress), []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeOK, result.Code)

	//4th confirm
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{user1EthOrderID1, user1EthOrderID2}, sdk.NewInt(160000))
	require.Equal(t, sdk.CodeOK, result.Code)

	//check user1's coins, coinsOnhold,AssetCoins, gr, gu
	user1CU = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, sdk.NewInt(3000000000), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt().String(), user1CU.GetAssetCoins().AmountOf(symbol).String())
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(160000), user1CU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(210000), user1CU.GetGasReceived().AmountOf(symbol))

	//check btcOpCU's coins
	ethOpCU = ck.GetCU(ctx, ethOpCUAddr)
	require.Equal(t, sdk.ZeroInt(), ethOpCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), ethOpCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(3000000000).SubRaw(210000), ethOpCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), ethOpCU.GetAssetCoinsHold().AmountOf(symbol))

	//check orders stautus
	require.Equal(t, sdk.OrderStatusFinish, ok.GetOrder(ctx, user1EthOrderID1).(*sdk.OrderCollect).Status)
	require.Equal(t, sdk.OrderStatusFinish, ok.GetOrder(ctx, user1EthOrderID2).(*sdk.OrderCollect).Status)
	require.Equal(t, sdk.OrderStatusBegin, ok.GetOrder(ctx, user1BtcOrderID1).(*sdk.OrderCollect).Status)
	require.Equal(t, sdk.OrderStatusBegin, ok.GetOrder(ctx, user1BtcOrderID2).(*sdk.OrderCollect).Status)

	//check deposit status

	require.Equal(t, sdk.DepositItemStatusConfirmed, ck.GetDeposit(ctx, symbol, user1CUAddr, "user1EthDeposit1", 0).Status)
	require.Equal(t, sdk.DepositItemStatusConfirmed, ck.GetDeposit(ctx, symbol, user1CUAddr, "user1EthDeposit2", 0).Status)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, ck.GetDeposit(ctx, token.BtcToken, user1CUAddr, "user1BtcDeposit1", 0).Status)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, ck.GetDeposit(ctx, token.BtcToken, user1CUAddr, "user1BtcDeposit2", 0).Status)

}

func TestCollectBtcSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}
	chain := token.BtcToken
	symbol := chain

	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tokenInfo.GasPrice = sdk.NewInt(551000 / 190) //Adjust gasPrice according txsize calculation
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	hash := "9ae3c919d84f4b72802de6f4f4aa0d88abcc9fd57315ddf27b8e25f032e4a180"
	index := uint64(1)
	fromAddr := "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"
	toAddr := "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"
	memo := ""
	amt := sdk.NewInt(85475551)
	orderID := uuid.NewV1().String()

	// success
	vin := &sdk.UtxoIn{Hash: hash, Index: index, Amount: amt, Address: toAddr}
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	mockCN.On("ValidAddress", chain, symbol, fromAddr).Return(true, fromAddr)

	//step1, Deposit 1 BTC item
	toCU := ck.GetCU(ctx, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	err = toCU.AddAsset(symbol, toAddr, 1)
	require.Nil(t, err)
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

	result := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, index, amt, orderID, memo)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	order := ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, symbol, order.GetSymbol())
	require.Equal(t, toCUAddr, order.GetCUAddress())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//no deposit item now
	dls1 := ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 0, len(dls1))

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, toCUAddr, of.CUAddress)

	df, valid := receipt.Flows[1].(sdk.DepositFlow)
	require.True(t, valid)
	require.Equal(t, toCUAddr.String(), df.CuAddress)
	require.Equal(t, memo, df.Memo)
	require.Equal(t, toAddr, df.Multisignedadress)
	require.Equal(t, symbol, df.Symbol)
	require.Equal(t, hash, df.Txhash)
	require.Equal(t, orderID, df.OrderID)
	require.Equal(t, amt, df.Amount)
	require.Equal(t, index, df.Index)
	require.Equal(t, sdk.DepositTypeCU, df.DepositType)

	//check CU's coins and assetcoins
	toCU = ck.GetCU(ctx, toCUAddr)
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))

	//1st confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//2nd confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[1].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))

	//3rd confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[2].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	//check receipt after deposit confirmed
	receipt, err = rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 3, len(receipt.Flows))

	of, valid = receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of.OrderStatus)

	cf, valid := receipt.Flows[1].(sdk.DepositConfirmedFlow)
	require.True(t, valid)
	require.Equal(t, 1, len(cf.ValidOrderIDs))
	require.Equal(t, 0, len(cf.InValidOrderIDs))
	require.Equal(t, orderID, cf.ValidOrderIDs[0])

	bf, valid := receipt.Flows[2].(sdk.BalanceFlow)
	require.True(t, valid)
	require.Equal(t, toCUAddr, bf.CUAddress)
	require.Equal(t, symbol, bf.Symbol.String())
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	//require.Equal(t, amt, bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	_, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls1))
	require.Equal(t, hash, dls1[0].Hash)
	require.Equal(t, amt, dls1[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls1[0].Status)
	require.Equal(t, memo, dls1[0].Memo)
	require.Equal(t, uint64(1), dls1[0].Index)

	depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash, index)
	require.Equal(t, dls1[0], depositItem)

	//check deposit item
	dls := ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls))
	require.Equal(t, hash, dls[0].Hash)
	require.Equal(t, amt, dls[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls[0].Status)
	require.Equal(t, memo, dls[0].Memo)
	require.Equal(t, index, dls[0].Index)

	//check CU's coins and assetcoins
	toCU = ck.GetCU(ctx, toCUAddr)
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))

	//step2, CollectWaitSign
	collectFromCUAddr := toCUAddr
	collectFromCU := toCU
	collectFromAddr := toAddr
	collectToCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	collectToCU := ck.GetCU(ctx, collectToCUAddr)
	require.Nil(t, err)
	require.Nil(t, err)
	collectToAddr := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	collectToCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = collectToCU.SetAssetAddress(symbol, collectToAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, collectToCU)
	rawData := []byte("rawdata")
	collectTxHash := "collectTxHash"
	signBytes := []byte("signbytes")

	collectFee := sdk.NewInt(551)
	collectAmt := amt.Sub(collectFee)

	chainnodeCollectTx := &chainnode.ExtUtxoTransaction{
		Hash:    collectTxHash,
		Vins:    []*sdk.UtxoIn{vin},
		Vouts:   []*sdk.UtxoOut{{Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg", Amount: collectAmt}},
		CostFee: collectFee,
	}

	ins := chainnodeCollectTx.Vins

	mockCN.On("ValidAddress", chain, symbol, collectToAddr).Return(true, collectToAddr)
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, chainnodeCollectTx.Vins).Return(chainnodeCollectTx, [][]byte{signBytes}, nil).Once()

	result2 := keeper.CollectWaitSign(ctx, collectToCUAddr, []string{orderID}, rawData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	//check receipts
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeCollect, receipt2.Category)
	require.Equal(t, 2, len(receipt2.Flows))
	orderFlow2, valid := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderTypeCollect, orderFlow2.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, orderFlow2.OrderStatus)

	cwf, valid := receipt2.Flows[1].(sdk.CollectWaitSignFlow)
	require.True(t, valid)
	require.Equal(t, []string{orderID}, cwf.OrderIDs)
	require.Equal(t, rawData, cwf.RawData)

	//check collect orders
	collectOrder1 := ok.GetOrder(ctx, orderID).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusWaitSign, collectOrder1.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, collectOrder1.GetOrderType())
	require.Equal(t, symbol, collectOrder1.GetSymbol())
	require.Equal(t, fromCUAddr, collectOrder1.GetCUAddress())
	require.Equal(t, collectToCUAddr, collectOrder1.CollectToCU)
	require.Equal(t, rawData, collectOrder1.RawData)

	//check deposit item
	dls = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, sdk.DepositItemStatusInProcess, dls[0].Status)

	//check collectFromCU's coins
	collectFromCU = ck.GetCU(ctx, collectFromCUAddr)
	require.Equal(t, amt, collectFromCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetGasReceived().AmountOf(symbol))

	//check collectToCU's coins
	collectToCU = ck.GetCU(ctx, collectToCUAddr)
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetGasReceived().AmountOf(symbol))

	//step3, CollectSignFinish
	signedData := []byte("signedTx")
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{collectFromAddr}, signedData, ins).Return(true, nil).Once()
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeCollectTx, nil).Once()
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(chainnodeCollectTx, [][]byte{signBytes}, nil).Once()

	result4 := keeper.CollectSignFinish(ctx, []string{orderID}, signedData, "")
	require.Equal(t, sdk.CodeOK, result4.Code)

	//check receipts
	receipt4, err := rk.GetReceiptFromResult(&result4)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeCollect, receipt4.Category)
	require.Equal(t, 2, len(receipt4.Flows))

	orderFlow4, valid := receipt4.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderTypeCollect, orderFlow4.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, orderFlow4.OrderStatus)

	sff, valid := receipt4.Flows[1].(sdk.CollectSignFinishFlow)
	require.True(t, valid)
	require.Equal(t, []string{orderID}, sff.OrderIDs)
	require.Equal(t, signedData, sff.SignedTx)

	//check collect orders
	collectOrder1 = ok.GetOrder(ctx, orderID).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusSignFinish, collectOrder1.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, collectOrder1.GetOrderType())
	require.Equal(t, symbol, collectOrder1.GetSymbol())
	require.Equal(t, fromCUAddr, collectOrder1.GetCUAddress())
	require.Equal(t, collectToCUAddr, collectOrder1.CollectToCU)
	require.Equal(t, rawData, collectOrder1.RawData)
	require.Equal(t, signedData, collectOrder1.SignedTx)

	//check deposit item
	dls = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, sdk.DepositItemStatusInProcess, dls[0].Status)

	//check collectFromCU's coins
	collectFromCU = ck.GetCU(ctx, collectFromCUAddr)
	require.Equal(t, amt, collectFromCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetGasReceived().AmountOf(symbol))

	//check collectToCU's coins
	collectToCU = ck.GetCU(ctx, collectToCUAddr)
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetGasReceived().AmountOf(symbol))

	//step4, CollectFinish
	chainnodeCollectTx.Status = chainnode.StatusSuccess
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{collectFromAddr}, signedData, ins).Return(true, nil).Once()
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeCollectTx, nil).Once()

	//1st confirm
	result5 := keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), []string{orderID}, collectFee)
	require.Equal(t, sdk.CodeOK, result5.Code)

	//2nd confirm
	result5 = keeper.CollectFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), []string{orderID}, collectFee)
	require.Equal(t, sdk.CodeOK, result5.Code)

	//3rd confirm
	result5 = keeper.CollectFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), []string{orderID}, collectFee)
	require.Equal(t, sdk.CodeOK, result5.Code)

	//check receipts
	receipt5, err := rk.GetReceiptFromResult(&result5)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeCollect, receipt5.Category)
	require.Equal(t, 2, len(receipt5.Flows))

	orderFlow5, valid := receipt5.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderStatusFinish, orderFlow5.OrderStatus)
	require.Equal(t, sdk.OrderTypeCollect, orderFlow5.OrderType)

	ff, valid := receipt5.Flows[1].(sdk.CollectFinishFlow)
	require.True(t, valid)
	require.Equal(t, []string{orderID}, ff.OrderIDs)
	require.Equal(t, collectFee, ff.CostFee)

	//check collect orders
	collectOrder1 = ok.GetOrder(ctx, orderID).(*sdk.OrderCollect)
	require.Equal(t, sdk.OrderStatusFinish, collectOrder1.GetOrderStatus())

	//check deposit item
	dls = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, sdk.DepositItemStatusConfirmed, dls[0].Status)

	//check new generate deposit item
	collectDepositItem := ck.GetDeposit(ctx, symbol, collectToCUAddr, collectTxHash, 0)
	require.Equal(t, sdk.DepositItemStatusConfirmed, collectDepositItem.Status)
	require.Equal(t, collectAmt, collectDepositItem.Amount)

	//check collectFromCU's coins
	collectFromCU = ck.GetCU(ctx, collectFromCUAddr)
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectFromCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, collectFee, collectFromCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, collectFee, collectFromCU.GetGasReceived().AmountOf(symbol))

	//check collectToCU's coins
	collectToCU = ck.GetCU(ctx, collectToCUAddr)
	require.Equal(t, collectAmt, collectToCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), collectToCU.GetGasReceived().AmountOf(symbol))

	//4th confirm
	result5 = keeper.CollectFinish(ctx, sdk.CUAddress(validators[3].GetOperator()), []string{orderID}, collectFee)
	require.Equal(t, sdk.CodeOK, result5.Code)
}

func TestColletBtcError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	//rk := input.rk
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := token.BtcToken
	chain := token.BtcToken
	validators := input.validators

	pubkey := ed25519.GenPrivKey().PubKey()

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tk.SetTokenInfo(ctx, tokenInfo)

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	require.Nil(t, err)

	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	user1CU := ck.GetCU(ctx, user1CUAddr)
	require.NotNil(t, user1CU)
	user1EthAddr := "0xc96d141c9110a8e61ed62caad8a7c858db15b82c"
	user1BtcAddr := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	user1CU.SetAssetPubkey(pubkey.Bytes(), 1)
	user1CU.SetAssetAddress(token.EthToken, user1EthAddr, 1)
	user1CU.SetAssetAddress(token.BtcToken, user1BtcAddr, 1)

	user1BtcDeposit1, err := sdk.NewDepositItem("user1BtcDeposit1", 0, sdk.NewInt(3000000000), user1BtcAddr, "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	user1BtcDeposit2, err := sdk.NewDepositItem("user1BtcDeposit2", 1, sdk.NewInt(4000000000), user1BtcAddr, "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	ck.SetDepositList(ctx, token.BtcToken, user1CUAddr, sdk.DepositList{user1BtcDeposit1, user1BtcDeposit2})
	user1CU.AddCoins(sdk.NewCoins(sdk.NewCoin(token.BtcToken, sdk.NewInt(7000000000))))
	user1CU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(token.BtcToken, sdk.NewInt(7000000000))))
	ck.SetCU(ctx, user1CU)
	user1BtcDepositList := ck.GetDepositList(ctx, token.BtcToken, user1CUAddr)

	user1BtcOrderID1 := uuid.NewV1().String()
	order := &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1BtcAddr,
		Txhash:             user1BtcDeposit1.Hash,
		Index:              user1BtcDeposit1.Index,
		Amount:             user1BtcDeposit1.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user1BtcOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder1)

	user1BtcOrderID2 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID2,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1BtcAddr,
		Txhash:             user1BtcDeposit2.Hash,
		Index:              user1BtcDeposit2.Index,
		Amount:             user1BtcDeposit2.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user1BtcOrder2 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder2)
	require.Equal(t, 2, len(user1BtcDepositList))
	require.Equal(t, user1BtcOrder1.GetID(), ok.GetOrder(ctx, user1BtcOrderID1).GetID())

	//setup user2
	user2CUAddr, err := sdk.CUAddressFromBase58("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q")
	require.Nil(t, err)
	user2CU := ck.GetCU(ctx, user2CUAddr)
	require.NotNil(t, user2CU)
	user2EthAddr := "0xd139e358ae9cb5424b2067da96f94cc938343446"
	user2BtcAddr := "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"
	user2CU.SetAssetPubkey(pubkey.Bytes(), 1)
	user2CU.SetAssetAddress(token.BtcToken, user2BtcAddr, 1)
	user2CU.SetAssetAddress(token.EthToken, user2EthAddr, 1)

	user2BtcDeposit1, err := sdk.NewDepositItem("user2BtcDeposit1", 0, sdk.NewInt(5000000000), user2BtcAddr, "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	user2BtcDeposit2, err := sdk.NewDepositItem("user2BtcDeposit2", 1, sdk.NewInt(6000000000), user2BtcAddr, "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	ck.SetDepositList(ctx, token.BtcToken, user2CUAddr, sdk.DepositList{user2BtcDeposit1, user2BtcDeposit2})
	user2CU.AddCoins(sdk.NewCoins(sdk.NewCoin(token.BtcToken, sdk.NewInt(11000000000))))
	user2CU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(token.BtcToken, sdk.NewInt(11000000000))))
	ck.SetCU(ctx, user2CU)
	user2BtcDepositList := ck.GetDepositList(ctx, token.BtcToken, user2CUAddr)

	user2BtcOrderID1 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user2CUAddr,
			ID:        user2BtcOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.BtcToken,
		},
		CollectFromCU:      user2CUAddr,
		CollectFromAddress: user2BtcAddr,
		Txhash:             user2BtcDeposit1.Hash,
		Index:              user2BtcDeposit1.Index,
		Amount:             user2BtcDeposit1.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user2BtcOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user2BtcOrder1)

	user2BtcOrderID2 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user2CUAddr,
			ID:        user2BtcOrderID2,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.BtcToken,
		},
		CollectFromCU:      user2CUAddr,
		CollectFromAddress: user2BtcAddr,
		Txhash:             user2BtcDeposit2.Hash,
		Index:              user2BtcDeposit2.Index,
		Amount:             user2BtcDeposit2.Amount,
		DepositStatus:      sdk.DepositConfirmed,
	}
	user2BtcOrder2 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user2BtcOrder2)
	require.Equal(t, 2, len(user2BtcDepositList))
	require.Equal(t, user2BtcOrder1.GetID(), ok.GetOrder(ctx, user2BtcOrderID1).GetID())

	mockCN.On("SupportChain", chain).Return(true)

	rawData := []byte("rawData")

	/*---CollectWaitSign----*/
	//noexist CU
	noExistCUAddr, _ := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	result := keeper.CollectWaitSign(ctx, noExistCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	//CU is not opCU
	user3CUAddr, _ := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	cu := ck.NewCUWithAddress(ctx, sdk.CUTypeUser, user3CUAddr)
	cu.SetAssetPubkey(pubkey.Bytes(), 1)
	ck.SetCU(ctx, cu)
	result = keeper.CollectWaitSign(ctx, user3CUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	//opCU only support eth
	ethOpCUAddr, _ := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	result = keeper.CollectWaitSign(ctx, ethOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "does not support symbol")

	//opCU does not have btc address
	btcOpCUAddr, _ := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, " does not have btc's address")

	//opCU's address is not valid
	collectToAddr := "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"
	mockCN.On("ValidAddress", chain, symbol, collectToAddr).Return(false, collectToAddr).Once()
	btcOpCU := ck.GetCU(ctx, btcOpCUAddr)
	err = btcOpCU.SetAssetAddress(symbol, collectToAddr, 1)
	ck.SetCU(ctx, btcOpCU)
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidAddress, result.Code)
	require.Contains(t, result.Log, "is not a valid address")

	//QueryUtxoTransactionFromData err
	collectTxHash := "collectTxHash"
	costFee := sdk.NewInt(5800)
	chainnodeCollectTx := chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	signHash1 := []byte("signHash1")
	signHash2 := []byte("signHash2")
	signHash3 := []byte("signHash3")

	ins := chainnodeCollectTx.Vins
	mockCN.On("ValidAddress", chain, symbol, collectToAddr).Return(true, collectToAddr)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, errors.New("QueryUtxoTransactionFromDataError")).Once()
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "QueryUtxoTransactionFromDataError")

	//Vin's order mismatch
	chainnodeCollectTx = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, nil).Once()
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "expected Vin")

	//Vout address mismatch
	chainnodeCollectTx = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao8", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, nil).Once()
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "collect to an unexpect address")

	//amount mismatch
	chainnodeCollectTx = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee).SubRaw(1)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, nil).Once()
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "!= outAmt")

	//gas price too high
	actualPrice := sdk.NewDecFromInt(costFee).Quo(sdk.NewDec(490)).Mul(sdk.NewDec(sdk.KiloBytes))
	tokenPrice := actualPrice.Quo(sdk.NewDecWithPrec(12, 1))
	tokenInfo.GasPrice = tokenPrice.TruncateInt()
	tk.SetTokenInfo(ctx, tokenInfo)

	chainnodeCollectTx = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, nil).Once()
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too high")

	//gas price roo low
	tokenPrice = actualPrice.Quo(sdk.NewDecWithPrec(8, 1))
	tokenInfo.GasPrice = tokenPrice.TruncateInt().AddRaw(1)
	tk.SetTokenInfo(ctx, tokenInfo)
	chainnodeCollectTx = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, nil).Once()
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too low")

	//everything is ok
	tokenPrice = actualPrice
	tokenInfo.GasPrice = tokenPrice.TruncateInt()
	tk.SetTokenInfo(ctx, tokenInfo)

	//tokenInfo.GasLimit = sdk.NewInt(10000)
	//tk.SetTokenInfo(ctx, tokenInfo)
	chainnodeCollectTx = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(&chainnodeCollectTx, [][]byte{signHash1, signHash2, signHash3}, nil)
	result = keeper.CollectWaitSign(ctx, btcOpCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	/*---CollectSignFinish----*/
	//VerifyUtxoSignedTransaction, err
	signedTx := []byte("singedTx")
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{user1BtcAddr, user1BtcAddr, user2BtcAddr}, signedTx, ins).Return(true, errors.New("err")).Once()
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Fail to verify signed transaction")

	//VerifyUtxoSignedTransaction verified =false
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{user1BtcAddr, user1BtcAddr, user2BtcAddr}, signedTx, ins).Return(false, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Fail to verify signed transaction")

	//QueryUtxoTransactionFromSignedData, err
	chainnodeCollectTx1 := chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, []string{user1BtcAddr, user1BtcAddr, user2BtcAddr}, signedTx, ins).Return(true, nil)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeCollectTx1, errors.New("err")).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Fail to get transaction")

	//Vin's mismatch
	chainnodeCollectTx1 = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "vin mismatch, expected")

	//Vout mismtach
	chainnodeCollectTx1 = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao8", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "vout mismatch, expected")

	//costFee mismatch
	chainnodeCollectTx1 = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee.SubRaw(1),
	}
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "costFee mismatch, expected")

	//everything is ok for collectsignFinish
	chainnodeCollectTx1 = chainnode.ExtUtxoTransaction{
		Hash:   collectTxHash,
		Status: chainnode.StatusSuccess,
		Vins: []*sdk.UtxoIn{
			{Hash: "user1BtcDeposit1", Index: 0, Amount: sdk.NewInt(3000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user1BtcDeposit2", Index: 1, Amount: sdk.NewInt(4000000000), Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"},
			{Hash: "user2BtcDeposit1", Index: 0, Amount: sdk.NewInt(5000000000), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"},
		},
		Vouts: []*sdk.UtxoOut{
			{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(12000000000).Sub(costFee)},
		},
		CostFee: costFee,
	}
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectSignFinish(ctx, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, signedTx, "")
	require.Equal(t, sdk.CodeOK, result.Code)

	/*---CollectFinish----*/
	//not a validator
	result = keeper.CollectFinish(ctx, fromCUAddr, []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "collect from not a validator")

	//empty order
	collectOrder1 := ok.GetOrder(ctx, user1BtcOrderID1)
	ok.DeleteOrder(ctx, collectOrder1)
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeNotFoundOrder, result.Code)
	require.Contains(t, result.Log, "does not exist")

	//order status is OrderStatusFinish
	orderStatus := collectOrder1.GetOrderStatus()
	collectOrder1.SetOrderStatus(sdk.OrderStatusFinish)
	ok.SetOrder(ctx, collectOrder1)
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)
	//require.Contains(t, result.Log, "not as expected")

	collectOrder1.SetOrderStatus(orderStatus)
	ok.SetOrder(ctx, collectOrder1)

	//1st confirm
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//2nd confirm
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[1].OperatorAddress), []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//3rd confirm
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedTx, ins).Return(&chainnodeCollectTx1, nil).Once()
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[2].OperatorAddress), []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//4nd confirm
	result = keeper.CollectFinish(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{user1BtcOrderID1, user1BtcOrderID2, user2BtcOrderID1}, costFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check user1's coins, coinsOnhold， AssetCoins，
	user1CU = ck.GetCU(ctx, user1CUAddr)
	require.Equal(t, sdk.NewInt(7000000000), user1CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user1CU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, costFee, user1CU.GetGasUsed().AmountOf(symbol))
	require.Equal(t, costFee, user1CU.GetGasReceived().AmountOf(symbol))

	//check user2's coins, coinsOnhold
	user2CU = ck.GetCU(ctx, user2CUAddr)
	require.Equal(t, sdk.NewInt(11000000000), user2CU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user2CU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(6000000000), user2CU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user2CU.GetAssetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user2CU.GetGasReceived().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), user2CU.GetGasUsed().AmountOf(symbol))

	//check btcOpCU's coins
	btcOpCU = ck.GetCU(ctx, btcOpCUAddr)
	require.Equal(t, sdk.ZeroInt(), btcOpCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), btcOpCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, sdk.NewInt(12000000000).Sub(costFee), btcOpCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), btcOpCU.GetAssetCoinsHold().AmountOf(symbol))

	//check orders stautus
	require.Equal(t, sdk.OrderStatusFinish, ok.GetOrder(ctx, user1BtcOrderID1).(*sdk.OrderCollect).Status)
	require.Equal(t, sdk.OrderStatusFinish, ok.GetOrder(ctx, user1BtcOrderID2).(*sdk.OrderCollect).Status)
	require.Equal(t, sdk.OrderStatusFinish, ok.GetOrder(ctx, user2BtcOrderID1).(*sdk.OrderCollect).Status)
	require.Equal(t, sdk.OrderStatusBegin, ok.GetOrder(ctx, user2BtcOrderID2).(*sdk.OrderCollect).Status)

	//check deposit status

	require.Equal(t, sdk.DepositItemStatusConfirmed, ck.GetDeposit(ctx, symbol, user1CUAddr, "user1BtcDeposit1", 0).Status)
	require.Equal(t, sdk.DepositItemStatusConfirmed, ck.GetDeposit(ctx, symbol, user1CUAddr, "user1BtcDeposit2", 1).Status)
	require.Equal(t, sdk.DepositItemStatusConfirmed, ck.GetDeposit(ctx, symbol, user2CUAddr, "user2BtcDeposit1", 0).Status)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, ck.GetDeposit(ctx, symbol, user2CUAddr, "user2BtcDeposit2", 1).Status)

}

func TestCheckCollectOrders(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	//rk := input.rk
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}

	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	user1CU := ck.GetCU(ctx, user1CUAddr)
	require.NotNil(t, user1CU)
	user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	user1BtcAddr := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	user1CU.SetAssetAddress(token.EthToken, user1EthAddr, 1)
	user1CU.SetAssetAddress(token.BtcToken, user1BtcAddr, 1)

	user1EthDeposit1, err := sdk.NewDepositItem("user1EthDeposit1", 0, sdk.NewInt(1000000000), "", "", sdk.DepositItemStatusUnCollected)
	require.Nil(t, err)
	user1EthDeposit2, err := sdk.NewDepositItem("user1EthDeposit2", 0, sdk.NewInt(2000000000), "", "", sdk.DepositItemStatusUnCollected)
	require.Nil(t, err)

	ck.SetDepositList(ctx, token.EthToken, user1CUAddr, sdk.DepositList{user1EthDeposit1, user1EthDeposit2})
	user1EthDepositList := ck.GetDepositList(ctx, token.EthToken, user1CUAddr)
	require.Equal(t, 2, len(user1EthDepositList))

	user1EthOrderID1 := uuid.NewV1().String()
	order := &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1EthOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.EthToken,
		},
		CollectFromCU: user1CUAddr,
		Txhash:        user1EthDeposit1.Hash,
		Index:         user1EthDeposit1.Index,
		Amount:        user1EthDeposit1.Amount,
		DepositStatus: sdk.DepositConfirmed,
	}

	user1EthOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1EthOrder1)

	user1EthOrderID2 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1EthOrderID2,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.EthToken,
		},
		CollectFromCU: user1CUAddr,
		Txhash:        user1EthDeposit2.Hash,
		Index:         user1EthDeposit2.Index,
		Amount:        user1EthDeposit2.Amount,
		DepositStatus: sdk.DepositConfirmed,
	}

	user1EthOrder2 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1EthOrder2)

	user1BtcDeposit1, err := sdk.NewDepositItem("user1BtcDeposit1", 0, sdk.NewInt(3000000000), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	user1BtcDeposit2, err := sdk.NewDepositItem("user1BtcDeposit2", 0, sdk.NewInt(4000000000), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	ck.SetDepositList(ctx, token.BtcToken, user1CUAddr, sdk.DepositList{user1BtcDeposit1, user1BtcDeposit2})
	user1BtcDepositList := ck.GetDepositList(ctx, token.BtcToken, user1CUAddr)

	user1BtcOrderID1 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.BtcToken,
		},
		CollectFromCU: user1CUAddr,
		Txhash:        user1BtcDeposit1.Hash,
		Index:         user1BtcDeposit1.Index,
		Amount:        user1BtcDeposit1.Amount,
		DepositStatus: sdk.DepositConfirmed,
	}

	user1BtcOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder1)

	user1BtcOrderID2 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1BtcOrderID2,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.BtcToken,
		},
		CollectFromCU: user1CUAddr,
		Txhash:        user1BtcDeposit2.Hash,
		Index:         user1BtcDeposit2.Index,
		Amount:        user1BtcDeposit2.Amount,
		DepositStatus: sdk.DepositConfirmed,
	}

	user1BtcOrder2 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1BtcOrder2)

	require.Equal(t, 2, len(user1BtcDepositList))
	require.Equal(t, user1BtcOrder1.GetID(), ok.GetOrder(ctx, user1BtcOrderID1).GetID())

	//setup user2
	user2CUAddr, err := sdk.CUAddressFromBase58("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q")
	require.Nil(t, err)
	user2CU := ck.GetCU(ctx, user2CUAddr)
	require.NotNil(t, user2CU)
	user2EthAddr := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	user2BtcAddr := "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"
	user2CU.SetAssetAddress(token.BtcToken, user2BtcAddr, 1)
	user2CU.SetAssetAddress(token.EthToken, user2EthAddr, 1)

	user2EthDeposit1, err := sdk.NewDepositItem("user2EthDeposit1", 0, sdk.NewInt(100000002), "", "", sdk.DepositItemStatusWaitCollect)
	require.Nil(t, err)
	ck.SetDepositList(ctx, token.EthToken, user2CUAddr, sdk.DepositList{user2EthDeposit1})

	user2EthOrderID1 := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user2CUAddr,
			ID:        user2EthOrderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    token.EthToken,
		},
		CollectFromCU: user2CUAddr,
		Txhash:        user2EthDeposit1.Hash,
		Index:         user2EthDeposit1.Index,
		Amount:        user2EthDeposit1.Amount,
		DepositStatus: sdk.DepositConfirmed,
	}
	user2EthOrder1 := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user2EthOrder1)

	//noexist order
	_, _, _, _, _, sdkErr := keeper.CheckCollectOrders(ctx, []string{uuid.NewV1().String()}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeNotFoundOrder, sdkErr.Code())

	//not a collect order
	notColletOrderID := uuid.NewV1().String()
	notCollectOrder := &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        notColletOrderID,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    "Eth",
		},
	}

	ok.SetOrder(ctx, ok.NewOrder(ctx, notCollectOrder))
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{notColletOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())

	//symbol is illegal
	user1ErrOrderID := uuid.NewV1().String()
	order = &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1ErrOrderID,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    "Eth",
		},
		CollectFromCU: user1CUAddr,
		Txhash:        "not exist deposit",
		Index:         user1EthDeposit1.Index,
		Amount:        user1EthDeposit1.Amount,
		DepositStatus: sdk.DepositConfirmed,
	}

	user1ErrOrder := ok.NewOrder(ctx, order)
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidSymbol, sdkErr.Code())

	//symbol is not support by token
	user1ErrOrder.(*sdk.OrderCollect).Symbol = "notsupport"
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeUnsupportToken, sdkErr.Code())

	//transfer sendenable is false
	keeper.SetSendEnabled(ctx, false)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, sdkErr.Code())

	//token's sendenable is false
	keeper.SetSendEnabled(ctx, true)
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, sdkErr.Code())

	//token's withdrawalenable is false
	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, sdkErr.Code())

	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(token.EthToken))
	tokenInfo.IsWithdrawalEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)

	//duplicated  order's
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1EthOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())

	//the second order does not exist
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, uuid.NewV1().String()}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	//require.Equal(t, sdk.CodeNotFoundOrder, sdkErr.Code())

	//the second order is not a collect order
	user1WithdrawalOrderID := uuid.NewV1().String()
	order1 := &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        user1WithdrawalOrderID,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    "Eth",
		},
		Txhash: user1EthDeposit1.Hash,
		Amount: user1EthDeposit1.Amount,
	}
	user1WithdrawalOrder := ok.NewOrder(ctx, order1)
	ok.SetOrder(ctx, user1WithdrawalOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1WithdrawalOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	//	require.Contains(t, sdkErr.Error(), "is not a Collect Order")

	//the second order status is not expected
	user1ErrOrder.(*sdk.OrderCollect).Symbol = token.EthToken
	user1ErrOrder.SetOrderStatus(sdk.OrderStatusWaitSign)
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	//	require.Contains(t, sdkErr.Error(), "doesn't match expctedStatus")

	//the second order symbol is not expected
	user1ErrOrder.(*sdk.OrderCollect).Symbol = token.BtcToken
	user1ErrOrder.SetOrderStatus(sdk.OrderStatusBegin)
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "symbol mismatch")

	//the second user CU does not exist
	user1ErrOrder.(*sdk.OrderCollect).Symbol = token.EthToken
	cuAddr, _ := sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
	user1ErrOrder.(*sdk.OrderCollect).CUAddress = cuAddr
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "CU does not exist")

	//the second user CU is not user type cu
	cuAddr, _ = sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	user1ErrOrder.(*sdk.OrderCollect).CUAddress = cuAddr
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "CU type is not user type")

	//the second user CU deposit item does not exist
	user1ErrOrder.(*sdk.OrderCollect).CUAddress = user2CUAddr
	ok.SetOrder(ctx, user1ErrOrder)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "does not exist")

	// the second user CU deposit staus is not expected
	user1ErrOrder.(*sdk.OrderCollect).CUAddress = user1CUAddr
	user1ErrOrder.(*sdk.OrderCollect).Txhash = user1EthDeposit2.Hash
	user1ErrOrder.(*sdk.OrderCollect).Amount = user1EthDeposit2.Amount
	user1ErrOrder.(*sdk.OrderCollect).Index = user1EthDeposit2.Index
	ok.SetOrder(ctx, user1ErrOrder)
	ck.SetDepositStatus(ctx, token.EthToken, user1CUAddr, user1EthDeposit2.Hash, user1EthDeposit2.Index, sdk.DepositItemStatusConfirmed)
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1ErrOrderID}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "status is not in expected statuses")
	ck.SetDepositStatus(ctx, token.EthToken, user1CUAddr, user1EthDeposit2.Hash, user1EthDeposit2.Index, sdk.DepositItemStatusWaitCollect)

	//the second cu's address is not consistent with the first one
	_, _, _, _, _, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user2EthOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Equal(t, sdk.CodeInvalidOrder, sdkErr.Code())
	require.Contains(t, sdkErr.Error(), "Different CU in one collect order for AccountBased token")

	amt, _, vins, collectOrders, deposits, sdkErr := keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusUnCollected)
	require.Nil(t, sdkErr)
	require.Equal(t, user1EthDeposit1.Amount, amt.AmountOf(token.EthToken))
	require.Nil(t, vins)
	require.Equal(t, user1EthOrder1.GetID(), collectOrders[0].GetID())
	require.Equal(t, &user1EthDeposit1, deposits[0])

	amt, _, vins, collectOrders, deposits, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1EthOrderID1, user1EthOrderID2}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Nil(t, sdkErr)
	require.Equal(t, user1EthDeposit1.Amount.Add(user1EthDeposit2.Amount), amt.AmountOf(token.EthToken))
	require.Nil(t, vins)
	require.Equal(t, user1EthOrder1.GetID(), collectOrders[0].GetID())
	require.Equal(t, user1EthOrder2.GetID(), collectOrders[1].GetID())
	require.Equal(t, &user1EthDeposit1, deposits[0])
	require.Equal(t, &user1EthDeposit2, deposits[1])

	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(token.BtcToken))
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsWithdrawalEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)

	amt, _, vins, collectOrders, deposits, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1BtcOrderID1}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Nil(t, sdkErr)
	require.Equal(t, user1BtcDeposit1.Amount, amt.AmountOf(token.BtcToken))
	require.Equal(t, 1, len(vins))
	require.Equal(t, "user1BtcDeposit1", vins[0].Hash)
	require.Equal(t, user1BtcOrder1.GetID(), collectOrders[0].GetID())
	require.Equal(t, &user1BtcDeposit1, deposits[0])

	amt, _, vins, collectOrders, deposits, sdkErr = keeper.CheckCollectOrders(ctx, []string{user1BtcOrderID1, user1BtcOrderID2}, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	require.Nil(t, sdkErr)
	require.Equal(t, user1BtcDeposit1.Amount.Add(user1BtcDeposit2.Amount), amt.AmountOf(token.BtcToken))
	require.Equal(t, 2, len(vins))
	require.Equal(t, "user1BtcDeposit1", vins[0].Hash)
	require.Equal(t, "user1BtcDeposit2", vins[1].Hash)
	require.Equal(t, user1BtcOrder1.GetID(), collectOrders[0].GetID())
	require.Equal(t, user1BtcOrder2.GetID(), collectOrders[1].GetID())
	require.Equal(t, &user1BtcDeposit1, deposits[0])
	require.Equal(t, &user1BtcDeposit2, deposits[1])
}
