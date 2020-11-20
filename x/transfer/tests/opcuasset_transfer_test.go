package tests

import (
	"errors"
	"strconv"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
)

func TestOpcuAstTransferBtcSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	ik := input.ik
	newTestCU := func(cu custodianunit.CU) *testCU {
		return newTestCU(ctx, input.trk, input.ik, cu)
	}

	ctx = ctx.WithBlockHeight(10)
	validators := input.validators
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := "btc"
	chain := "btc"
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "btc")))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "eth")))

	// setup epoch
	var valAddr []sdk.CUAddress
	for _, val := range validators {
		valAddr = append(valAddr, sdk.CUAddress(val.OperatorAddress))
	}
	input.stakingkeeper.StartNewEpoch(ctx, valAddr)
	ctx = ctx.WithBlockHeight(11)

	//setup token
	tokenInfo := tk.GetIBCToken(ctx, sdk.Symbol("btc"))
	tokenInfo.GasPrice = sdk.NewInt(10000000 / 380)
	tk.SetToken(ctx, tokenInfo)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	newEpochAddress := "miaTKppfCyfW1FdRYfmUuDwbjKMNcRtSMb"
	depositList := sdk.DepositList{}
	totalAmount := sdk.ZeroInt()
	for i := 1; i <= 7; i++ {
		amount := sdk.NewInt(int64(10000000 * i))
		hash := "opcu_utxo_deposit_" + strconv.Itoa(i)
		d, err := sdk.NewDepositItem(hash, 0, amount, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
		require.Nil(t, err)
		depositList = append(depositList, d)
		totalAmount = totalAmount.Add(amount)
	}

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	err = opCU.SetAssetAddress(symbol, newEpochAddress, 2)
	opCU.SetAssetPubkey(pubkey.Bytes(), 2)
	require.Nil(t, err)

	ik.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), depositList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, totalAmount)))
	ck.SetCU(ctx, opCU)

	opCU1 := newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	require.Equal(t, totalAmount, opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, depositList, ik.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))

	mockCN.On("ValidAddress", chain, symbol, newEpochAddress).Return(true, newEpochAddress)

	//Step1, OpcuAssetTransfer
	orderID := uuid.NewV1().String()
	var transferItems []sdk.TransferItem
	transferAmount := sdk.ZeroInt()
	for i := 0; i < 6; i++ {
		transferItems = append(transferItems, sdk.TransferItem{
			Hash:   depositList[i].Hash,
			Amount: depositList[i].Amount,
			Index:  depositList[i].Index,
		})
		transferAmount = transferAmount.Add(depositList[i].Amount)
	}
	result := keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	o := ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusBegin, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, btcOPCUAddr, o.GetCUAddress())

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, btcOPCUAddr, of.CUAddress)

	wf, v := receipt.Flows[1].(sdk.OpcuAssetTransferFlow)
	require.True(t, v)
	require.Equal(t, btcOPCUAddr.String(), wf.Opcu)
	require.Equal(t, opCUBtcAddress, wf.FromAddr)
	require.Equal(t, newEpochAddress, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, orderID, wf.OrderID)

	opcu := newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	require.Equal(t, sdk.MigrationAssetBegin, opcu.GetMigrationStatus())
	depositList = ik.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr))
	for i, item := range depositList {
		if i < 6 {
			require.Equal(t, sdk.DepositItemStatusInProcess, item.Status)
		} else {
			require.Equal(t, sdk.DepositItemStatusConfirmed, item.Status)
		}
	}

	// Step2, OpcuAssetTransferWaitSign
	var vins []*sdk.UtxoIn
	var signHashes [][]byte
	var hashes [][]byte
	for i := 0; i < 6; i++ {
		vin := sdk.NewUtxoIn(depositList[i].Hash, depositList[i].Index, depositList[i].Amount, opCUBtcAddress)
		vins = append(vins, &vin)
		signHashes = append(signHashes, []byte(strconv.Itoa(i)))
		hashes = append(hashes, []byte(strconv.Itoa(i)))
	}
	opcuAstTransferTxHash := "opcuAstTransferTxHash"
	gasFee := sdk.NewInt(25000)
	chainnodeTx := &chainnode.ExtUtxoTransaction{
		Hash: opcuAstTransferTxHash,
		Vins: vins,
		Vouts: []*sdk.UtxoOut{
			{Address: newEpochAddress, Amount: transferAmount.Sub(gasFee)},
		},
		CostFee: gasFee,
	}

	ins := chainnodeTx.Vins

	rawData := []byte("rawData")
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(chainnodeTx, signHashes, nil)

	ctx = ctx.WithBlockHeight(12)
	result1 := keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, btcOPCUAddr, o.GetCUAddress())

	// check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, v := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, btcOPCUAddr, of1.CUAddress)

	wwf, v := receipt1.Flows[1].(sdk.OpcuAssetTransferWaitSignFlow)
	require.True(t, v)
	require.Equal(t, orderID, wwf.OrderID)
	require.Equal(t, rawData, wwf.RawData)

	// Step3, OpcuAssetTransferSignFinish
	signedData := []byte("signedData")
	chainnodeTx.BlockHeight = 10000 //btc height = 10000
	mockCN.On("QueryUtxoInsFromData", chain, symbol, signedData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx, nil).Once()
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, mock.Anything, signedData, ins).Return(true, nil).Once()

	result2 := keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	// check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, btcOPCUAddr, o.GetCUAddress())
	require.Equal(t, rawData, o.(*sdk.OrderOpcuAssetTransfer).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderOpcuAssetTransfer).SignedTx)

	// check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt2.Category)
	require.Equal(t, 2, len(receipt2.Flows))

	of2, v := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, btcOPCUAddr, of2.CUAddress)

	wsf, v := receipt2.Flows[1].(sdk.OpcuAssetTransferSignFinishFlow)
	require.True(t, v)
	require.Equal(t, orderID, wsf.OrderID)
	require.Equal(t, signedData, wsf.SignedTx)

	//Step4, OpcuAssetTransferFinish
	chainnodeTx.Status = chainnode.StatusSuccess //btc height = 10000
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx, nil)

	//1st confirm
	result3 := keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//2nd confirm
	result3 = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//3rd confirm
	result3 = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	// check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, btcOPCUAddr, o.GetCUAddress())
	require.Equal(t, rawData, o.(*sdk.OrderOpcuAssetTransfer).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderOpcuAssetTransfer).SignedTx)
	require.Equal(t, gasFee, o.(*sdk.OrderOpcuAssetTransfer).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt3.Category)
	require.Equal(t, 2, len(receipt3.Flows))

	of3, v := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, btcOPCUAddr, of3.CUAddress)

	wff, v := receipt3.Flows[1].(sdk.OpcuAssetTransferFinishFlow)
	require.True(t, v)
	require.Equal(t, orderID, wff.OrderID)
	require.Equal(t, gasFee, wff.CostFee)

	opcu = newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	require.Equal(t, sdk.MigrationAssetBegin, opcu.GetMigrationStatus())
	depositList = ik.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr))
	require.Equal(t, 2, len(depositList))
	require.True(t, newEpochAddress == depositList[0].ExtAddress || newEpochAddress == depositList[1].ExtAddress)
	require.Equal(t, sdk.DepositItemStatusConfirmed, depositList[0].Status)
	require.Equal(t, sdk.DepositItemStatusConfirmed, depositList[1].Status)
}

func TestOpcuAstTransferBtcError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	ik := input.ik
	newTestCU := func(cu custodianunit.CU) *testCU {
		return newTestCU(ctx, input.trk, input.ik, cu)
	}

	ctx = ctx.WithBlockHeight(10)
	validators := input.validators
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := "btc"
	chain := "btc"
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "btc")))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "eth")))

	//setup token
	tokenInfo := tk.GetIBCToken(ctx, sdk.Symbol("btc"))
	tokenInfo.GasPrice = sdk.NewInt(10000000 / 380)
	tk.SetToken(ctx, tokenInfo)

	//setup OpCU
	opCUBtcAddress := "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg"
	newEpochAddress := "miaTKppfCyfW1FdRYfmUuDwbjKMNcRtSMb"
	depositList := sdk.DepositList{}
	totalAmount := sdk.ZeroInt()
	for i := 1; i <= 7; i++ {
		amount := sdk.NewInt(int64(10000000 * i))
		hash := "opcu_utxo_deposit_" + strconv.Itoa(i)
		d, err := sdk.NewDepositItem(hash, 0, amount, opCUBtcAddress, "", sdk.DepositItemStatusConfirmed)
		require.Nil(t, err)
		depositList = append(depositList, d)
		totalAmount = totalAmount.Add(amount)
	}

	btcOPCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)
	opCU := newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUBtcAddress, 1)
	err = opCU.SetAssetAddress(symbol, newEpochAddress, 2)
	opCU.SetAssetPubkey(pubkey.Bytes(), 2)
	require.Nil(t, err)

	ik.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), depositList)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, totalAmount)))
	ck.SetCU(ctx, opCU)

	opCU1 := newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	require.Equal(t, totalAmount, opCU1.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, depositList, ik.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr)))

	mockCN.On("ValidAddress", chain, symbol, newEpochAddress).Return(true, newEpochAddress)
	mockCN.On("ValidAddress", chain, symbol, opCUBtcAddress).Return(true, opCUBtcAddress)

	orderID := uuid.NewV1().String()
	var transferItems []sdk.TransferItem
	transferAmount := sdk.ZeroInt()
	for i := 0; i < len(depositList); i++ {
		transferItems = append(transferItems, sdk.TransferItem{
			Hash:   depositList[i].Hash,
			Amount: depositList[i].Amount,
			Index:  depositList[i].Index,
		})
		transferAmount = transferAmount.Add(depositList[i].Amount)
	}
	// 不在轮换期
	result := keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// setup epoch
	var valAddr []sdk.CUAddress
	for _, val := range validators {
		valAddr = append(valAddr, sdk.CUAddress(val.OperatorAddress))
	}
	input.stakingkeeper.StartNewEpoch(ctx, valAddr)
	ctx = ctx.WithBlockHeight(11)

	// to addr 周期不对
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, opCUBtcAddress, orderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeInvalidAddress, result.Code)
	// transfer items 数量不足
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems[:4])
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// transfer items 数量过多
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// transfer items 重复
	originItem := transferItems[0]
	transferItems[0] = transferItems[1]
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	transferItems[0] = originItem

	// symbol 不存在
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, "invalidcoin", transferItems[:6])
	require.Equal(t, sdk.CodeUnsupportToken, result.Code)

	// order id 存在
	duplicatedOrderID := uuid.NewV1().String()
	duplicatdcOrder := &sdk.OrderOpcuAssetTransfer{
		OrderBase: sdk.OrderBase{
			CUAddress: btcOPCUAddr,
			ID:        duplicatedOrderID,
			OrderType: sdk.OrderTypeOpcuAssetTransfer,
			Symbol:    "btc",
		},
	}
	ok.SetOrder(ctx, duplicatdcOrder)
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, duplicatedOrderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "already exists")

	// cu 不存在
	cuAddr, _ := sdk.CUAddressFromBase58("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe")
	result = keeper.OpcuAssetTransfer(ctx, cuAddr, newEpochAddress, orderID, symbol, transferItems[:4])
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	// transfer item 不存在
	correctHash := transferItems[0].Hash
	transferItems[0].Hash = "wrong_hash"
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	transferItems[0].Hash = correctHash

	// utxo 状态不正确
	depositList[0].Status = sdk.DepositItemStatusInProcess
	ik.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), depositList)
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// utxo 地址周期不对
	depositList[0].Status = sdk.DepositItemStatusConfirmed
	depositList[0].ExtAddress = newEpochAddress
	ik.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), depositList)
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// 所有都正确，通过
	depositList[0].ExtAddress = opCUBtcAddress
	ik.SetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr), depositList)
	result = keeper.OpcuAssetTransfer(ctx, btcOPCUAddr, newEpochAddress, orderID, symbol, transferItems[:6])
	require.Equal(t, sdk.CodeOK, result.Code)

	// OpcuAssetTransferWaitSign
	var vins []*sdk.UtxoIn
	var signHashes [][]byte
	var hashes [][]byte
	for i := 0; i < 6; i++ {
		vin := sdk.NewUtxoIn(depositList[i].Hash, depositList[i].Index, depositList[i].Amount, opCUBtcAddress)
		vins = append(vins, &vin)
		signHashes = append(signHashes, []byte(strconv.Itoa(i)))
		hashes = append(hashes, []byte(strconv.Itoa(i)))
	}
	opcuAstTransferTxHash := "opcuAstTransferTxHash"
	gasFee := sdk.NewInt(25000)
	chainnodeTx := &chainnode.ExtUtxoTransaction{
		Hash: opcuAstTransferTxHash,
		Vins: vins,
		Vouts: []*sdk.UtxoOut{
			{Address: newEpochAddress, Amount: transferAmount.Sub(gasFee)},
		},
		CostFee: gasFee,
	}

	ctx = ctx.WithBlockHeight(12)
	ins := chainnodeTx.Vins
	rawData := []byte("rawData")

	// rawData 不正确
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, errors.New("fail to QueryUtxoInsFromData")).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// vins 数量不对
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins[:5], nil).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// vins 不存在
	correctHash = ins[0].Hash
	ins[0].Hash = "wrong_hash"
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	ins[0].Hash = correctHash

	// QueryUtxoTransactionFromData 失败
	mockCN.On("QueryUtxoInsFromData", chain, symbol, rawData).Return(ins, nil)
	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(chainnodeTx, signHashes, errors.New("QueryUtxoTransactionFromDataError")).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	mockCN.On("QueryUtxoTransactionFromData", chain, symbol, rawData, ins).Return(chainnodeTx, signHashes, nil)
	// len(hashes) != len(signHashes)
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes[:5], rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "signhashes's number mismatch")

	//hash mismatch
	tmpHash := hashes[0]
	hashes[0] = []byte("wrong_hash")
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "mismatch hashes")
	hashes[0] = tmpHash

	//price is too hight
	size := sdk.EstimateSignedUtxoTxSize(6, 1).ToDec()
	actualPrice := sdk.NewDecFromInt(gasFee).Quo(size).Mul(sdk.NewDec(sdk.KiloBytes))
	tokenPrice := actualPrice.Quo(sdk.NewDecWithPrec(12, 1))
	tokenInfo.GasPrice = tokenPrice.TruncateInt()
	tk.SetToken(ctx, tokenInfo)
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too high")

	//price is too low
	tokenPrice = actualPrice.Quo(sdk.NewDecWithPrec(8, 1))
	tokenInfo.GasPrice = tokenPrice.TruncateInt().AddRaw(1)
	tk.SetToken(ctx, tokenInfo)
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too low")

	//len(hashes) != len(signHashes)
	tokenInfo.GasPrice = actualPrice.TruncateInt()
	tk.SetToken(ctx, tokenInfo)

	//everything is ok
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	// OpcuAssetTransferSignFinish
	signedData := []byte("signedData")
	chainnodeTx.BlockHeight = 10000 //btc height = 10000

	// VerifyUtxoSignedTransaction error
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, mock.Anything, signedData, ins).Return(true, errors.New("VerifyUtxoSignedTransactionError")).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "VerifyUtxoSignedTransactionError")

	// VerifyUtxoSignedTransaction,  verified = false
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, mock.Anything, signedData, ins).Return(false, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// QueryUtxoTransactionFromSignedData error
	mockCN.On("VerifyUtxoSignedTransaction", chain, symbol, mock.Anything, signedData, ins).Return(true, nil)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx, errors.New("QueryUtxoTransactionFromSignedDataError")).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	chainnodeTx2 := &chainnode.ExtUtxoTransaction{
		Hash: opcuAstTransferTxHash,
		Vins: vins,
		Vouts: []*sdk.UtxoOut{
			{Address: newEpochAddress, Amount: transferAmount.Sub(gasFee)},
		},
		CostFee: gasFee,
	}
	// vins mismatch
	chainnodeTx2.Vins = append(vins, vins[0])
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	chainnodeTx2.Vins = vins

	// vout mismatch
	chainnodeTx2.Vouts = append(chainnodeTx2.Vouts, chainnodeTx2.Vouts[0])
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	chainnodeTx2.Vouts = chainnodeTx2.Vouts[:1]

	// cost fee mismatch
	chainnodeTx2.CostFee = gasFee.AddRaw(1)
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	chainnodeTx2.CostFee = gasFee

	// everything is ok
	mockCN.On("QueryUtxoTransactionFromSignedData", chain, symbol, signedData, ins).Return(chainnodeTx2, nil)
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result.Code)

	// OpcuAssetTransferFinish
	// empty order
	opcuAstTransferOrder := ok.GetOrder(ctx, orderID)
	ok.DeleteOrder(ctx, opcuAstTransferOrder)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeNotFoundOrder, result.Code)
	require.Contains(t, result.Log, "does not exist")

	orderStatus := opcuAstTransferOrder.GetOrderStatus()
	opcuAstTransferOrder.SetOrderStatus(sdk.OrderStatusFinish)
	ok.SetOrder(ctx, opcuAstTransferOrder)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	opcuAstTransferOrder.SetOrderStatus(orderStatus)
	ok.SetOrder(ctx, opcuAstTransferOrder)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	// check order
	o := ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, btcOPCUAddr, o.GetCUAddress())
	require.Equal(t, rawData, o.(*sdk.OrderOpcuAssetTransfer).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderOpcuAssetTransfer).SignedTx)
	require.Equal(t, gasFee, o.(*sdk.OrderOpcuAssetTransfer).CostFee)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of.OrderStatus)
	require.Equal(t, btcOPCUAddr, of.CUAddress)

	wff, v := receipt.Flows[1].(sdk.OpcuAssetTransferFinishFlow)
	require.True(t, v)
	require.Equal(t, orderID, wff.OrderID)
	require.Equal(t, gasFee, wff.CostFee)

	opcu := newTestCU(ck.GetCU(ctx, btcOPCUAddr))
	require.Equal(t, sdk.MigrationAssetBegin, opcu.GetMigrationStatus())
	depositList = ik.GetDepositList(ctx, symbol, sdk.CUAddress(btcOPCUAddr))
	require.Equal(t, 2, len(depositList))
	require.True(t, newEpochAddress == depositList[0].ExtAddress || newEpochAddress == depositList[1].ExtAddress)
	require.Equal(t, sdk.DepositItemStatusConfirmed, depositList[0].Status)
	require.Equal(t, sdk.DepositItemStatusConfirmed, depositList[1].Status)
}

func TestOpcuAstTransferEthSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	newTestCU := func(cu custodianunit.CU) *testCU {
		return newTestCU(ctx, input.trk, input.ik, cu)
	}

	ctx = ctx.WithBlockHeight(10)
	validators := input.validators
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := "eth"
	chain := "eth"
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "btc")))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "eth")))

	// setup epoch
	var valAddr []sdk.CUAddress
	for _, val := range validators {
		valAddr = append(valAddr, sdk.CUAddress(val.OperatorAddress))
	}
	input.stakingkeeper.StartNewEpoch(ctx, valAddr)
	ctx = ctx.WithBlockHeight(11)

	//setup token
	tokenInfo := tk.GetIBCToken(ctx, sdk.Symbol("eth"))
	tokenInfo.GasLimit = sdk.NewInt(10000)
	tokenInfo.GasPrice = sdk.NewInt(100)
	tokenInfo.SysTransferNum = sdk.NewInt(3)
	tk.SetToken(ctx, tokenInfo)

	//setup OpCU
	opCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	newEpochAddress := "0x4543429c2110d850BA382C815fD2FeC9E821ee96"

	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUEthAddress, 1)
	err = opCU.SetAssetAddress(symbol, newEpochAddress, 2)
	opCU.SetAssetPubkey(pubkey.Bytes(), 2)
	require.Nil(t, err)

	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU1 := newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU1.GetAssetCoins().AmountOf(symbol))

	mockCN.On("ValidAddress", chain, symbol, newEpochAddress).Return(true, newEpochAddress)

	//Step1, OpcuAssetTransfer
	orderID := uuid.NewV1().String()
	transferItems := []sdk.TransferItem{
		{Amount: opCUEthAmt},
	}
	result := keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	o := ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusBegin, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, ethOPCUAddr, o.GetCUAddress())

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusBegin, of.OrderStatus)
	require.Equal(t, ethOPCUAddr, of.CUAddress)

	wf, v := receipt.Flows[1].(sdk.OpcuAssetTransferFlow)
	require.True(t, v)
	require.Equal(t, ethOPCUAddr.String(), wf.Opcu)
	require.Equal(t, opCUEthAddress, wf.FromAddr)
	require.Equal(t, newEpochAddress, wf.ToAddr)
	require.Equal(t, symbol, wf.Symbol)
	require.Equal(t, orderID, wf.OrderID)

	opcu := newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	require.Equal(t, sdk.MigrationAssetBegin, opcu.GetMigrationStatus())

	// Step2, OpcuAssetTransferWaitSign
	opcuAstTransferTxHash := "opcuAstTransferTxHash"
	gasFee := sdk.NewInt(25000)
	chainnodeTx := &chainnode.ExtAccountTransaction{
		Hash:     opcuAstTransferTxHash,
		From:     opCUEthAddress,
		To:       newEpochAddress,
		Amount:   opCUEthAmt.SubRaw(1000000),
		Nonce:    0,
		GasLimit: sdk.NewInt(10000),
		GasPrice: sdk.NewInt(100),
	}

	rawData := []byte("rawData")
	signHash := []byte("signHash")
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(chainnodeTx, []byte(signHash), nil)

	ctx = ctx.WithBlockHeight(12)
	result1 := keeper.OpcuAssetTransferWaitSign(ctx, orderID, [][]byte{signHash}, rawData)
	require.Equal(t, sdk.CodeOK, result1.Code)

	//check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusWaitSign, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, ethOPCUAddr, o.GetCUAddress())

	// check receipt
	receipt1, err := rk.GetReceiptFromResult(&result1)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt1.Category)
	require.Equal(t, 2, len(receipt1.Flows))

	of1, v := receipt1.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of1.Symbol.String())
	require.Equal(t, orderID, of1.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of1.OrderType)
	require.Equal(t, sdk.OrderStatusWaitSign, of1.OrderStatus)
	require.Equal(t, ethOPCUAddr, of1.CUAddress)

	wwf, v := receipt1.Flows[1].(sdk.OpcuAssetTransferWaitSignFlow)
	require.True(t, v)
	require.Equal(t, orderID, wwf.OrderID)
	require.Equal(t, rawData, wwf.RawData)

	// Step3, OpcuAssetTransferSignFinish
	signedData := []byte("signedData")
	chainnodeTx.BlockHeight = 10000 //btc height = 10000
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, opCUEthAddress, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx, nil).Once()

	result2 := keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result2.Code)

	// check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusSignFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, ethOPCUAddr, o.GetCUAddress())
	require.Equal(t, rawData, o.(*sdk.OrderOpcuAssetTransfer).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderOpcuAssetTransfer).SignedTx)

	// check receipt
	receipt2, err := rk.GetReceiptFromResult(&result2)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt2.Category)
	require.Equal(t, 2, len(receipt2.Flows))

	of2, v := receipt2.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of2.Symbol.String())
	require.Equal(t, orderID, of2.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of2.OrderType)
	require.Equal(t, sdk.OrderStatusSignFinish, of2.OrderStatus)
	require.Equal(t, ethOPCUAddr, of2.CUAddress)

	wsf, v := receipt2.Flows[1].(sdk.OpcuAssetTransferSignFinishFlow)
	require.True(t, v)
	require.Equal(t, orderID, wsf.OrderID)
	require.Equal(t, signedData, wsf.SignedTx)

	//Step4, OpcuAssetTransferFinish
	costFee := sdk.NewInt(800000)
	chainnodeTx.CostFee = costFee
	chainnodeTx.Status = chainnode.StatusSuccess
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx, nil).Once()

	//1st confirm
	result3 := keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//2nd confirm
	result3 = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	//3rd confirm
	result3 = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result3.Code)

	// check order
	o = ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, ethOPCUAddr, o.GetCUAddress())
	require.Equal(t, rawData, o.(*sdk.OrderOpcuAssetTransfer).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderOpcuAssetTransfer).SignedTx)
	require.Equal(t, gasFee, o.(*sdk.OrderOpcuAssetTransfer).CostFee)

	//check receipt
	receipt3, err := rk.GetReceiptFromResult(&result3)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt3.Category)
	require.Equal(t, 2, len(receipt3.Flows))

	of3, v := receipt3.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of3.Symbol.String())
	require.Equal(t, orderID, of3.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of3.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of3.OrderStatus)
	require.Equal(t, ethOPCUAddr, of3.CUAddress)

	wff, v := receipt3.Flows[1].(sdk.OpcuAssetTransferFinishFlow)
	require.True(t, v)
	require.Equal(t, orderID, wff.OrderID)
	require.Equal(t, gasFee, wff.CostFee)

	opcu = newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	require.Equal(t, sdk.MigrationFinish, opcu.GetMigrationStatus())
}

func TestOpcuAstTransferEthError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	newTestCU := func(cu custodianunit.CU) *testCU {
		return newTestCU(ctx, input.trk, input.ik, cu)
	}

	ctx = ctx.WithBlockHeight(10)
	validators := input.validators
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	symbol := "eth"
	chain := "eth"
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "btc")))
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, "eth")))

	//setup token
	tokenInfo := tk.GetIBCToken(ctx, sdk.Symbol("eth"))
	tokenInfo.GasLimit = sdk.NewInt(10000)
	tokenInfo.GasPrice = sdk.NewInt(100)
	tokenInfo.SysTransferNum = sdk.NewInt(3)
	tk.SetToken(ctx, tokenInfo)

	//setup OpCU
	opCUEthAddress := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	newEpochAddress := "0x4543429c2110d850BA382C815fD2FeC9E821ee96"

	ethOPCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)
	opCU := newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	err = opCU.SetAssetAddress(symbol, opCUEthAddress, 1)
	err = opCU.SetAssetAddress(symbol, newEpochAddress, 2)
	opCU.SetAssetPubkey(pubkey.Bytes(), 2)
	require.Nil(t, err)

	opCUEthAmt := sdk.NewInt(90000000)
	opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, opCUEthAmt)))
	ck.SetCU(ctx, opCU)

	opCU1 := newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	require.Equal(t, opCUEthAmt, opCU1.GetAssetCoins().AmountOf(symbol))

	mockCN.On("ValidAddress", chain, symbol, newEpochAddress).Return(true, newEpochAddress)
	mockCN.On("ValidAddress", chain, symbol, opCUEthAddress).Return(true, opCUEthAddress)

	// OpcuAssetTransfer
	orderID := uuid.NewV1().String()
	transferItems := []sdk.TransferItem{
		{Amount: opCUEthAmt},
	}
	// 不在轮换期
	result := keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// setup epoch
	var valAddr []sdk.CUAddress
	for _, val := range validators {
		valAddr = append(valAddr, sdk.CUAddress(val.OperatorAddress))
	}
	input.stakingkeeper.StartNewEpoch(ctx, valAddr)
	ctx = ctx.WithBlockHeight(11)

	// to addr 周期不对
	result = keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, opCUEthAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidAddress, result.Code)
	// transfer items 数量过多
	result = keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, orderID, symbol, append(transferItems, transferItems[0]))
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// symbol 不存在
	result = keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, orderID, "invalidcoin", transferItems)
	require.Equal(t, sdk.CodeUnsupportToken, result.Code)

	// order id 存在
	duplicatedOrderID := uuid.NewV1().String()
	duplicatdcOrder := &sdk.OrderOpcuAssetTransfer{
		OrderBase: sdk.OrderBase{
			CUAddress: ethOPCUAddr,
			ID:        duplicatedOrderID,
			OrderType: sdk.OrderTypeOpcuAssetTransfer,
			Symbol:    "eth",
		},
	}
	ok.SetOrder(ctx, duplicatdcOrder)
	result = keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, duplicatedOrderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "already exists")

	// cu 不存在
	cuAddr, _ := sdk.CUAddressFromBase58("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe")
	result = keeper.OpcuAssetTransfer(ctx, cuAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)

	// transfer item 金额不对
	correctAmount := transferItems[0].Amount
	transferItems[0].Amount = transferItems[0].Amount.SubRaw(1)
	result = keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	transferItems[0].Amount = correctAmount

	// 所有都正确，通过
	result = keeper.OpcuAssetTransfer(ctx, ethOPCUAddr, newEpochAddress, orderID, symbol, transferItems)
	require.Equal(t, sdk.CodeOK, result.Code)

	// OpcuAssetTransferWaitSign
	opcuAstTransferTxHash := "opcuAstTransferTxHash"
	gasFee := sdk.NewInt(25000)
	chainnodeTx := &chainnode.ExtAccountTransaction{
		Hash:     opcuAstTransferTxHash,
		From:     opCUEthAddress,
		To:       newEpochAddress,
		Amount:   opCUEthAmt.SubRaw(1000000),
		Nonce:    0,
		GasLimit: sdk.NewInt(10000),
		GasPrice: sdk.NewInt(100),
	}

	rawData := []byte("rawData")
	signHash := []byte("signHash")
	hashes := [][]byte{signHash}

	ctx = ctx.WithBlockHeight(12)

	// rawData 不正确
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(chainnodeTx, signHash, errors.New("Fail to QueryAccountTransactionFromData")).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// amount 不正确
	correctAmount = chainnodeTx.Amount
	chainnodeTx.Amount = chainnodeTx.Amount.AddRaw(1)
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(chainnodeTx, signHash, nil).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Unexpected opcu asset transfer Amount")
	chainnodeTx.Amount = correctAmount

	// contract 不正确
	correctContract := chainnodeTx.ContractAddress
	chainnodeTx.ContractAddress = "wrong_contract"
	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(chainnodeTx, signHash, nil).Once()
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "Unexpected opcu asset transfer contract address")
	chainnodeTx.ContractAddress = correctContract

	mockCN.On("QueryAccountTransactionFromData", chain, symbol, rawData).Return(chainnodeTx, signHash, nil)
	// len(hashes) != len(signHashes)
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, append(hashes, hashes[0]), rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "AccountBased token supports only one opcutastransfer at one time")

	//hash mismatch
	correctHash := hashes[0]
	hashes[0] = []byte("wrong_hash")
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "hash mismatch")
	hashes[0] = correctHash

	//price is too hight
	correctPrice := tokenInfo.GasPrice
	tokenInfo.GasPrice = sdk.NewDecFromInt(correctPrice).Quo(sdk.NewDecWithPrec(12, 1)).TruncateInt()
	tk.SetToken(ctx, tokenInfo)
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too high")

	//price is too low
	tokenInfo.GasPrice = sdk.NewDecFromInt(correctPrice).Quo(sdk.NewDecWithPrec(8, 1)).TruncateInt().AddRaw(1)
	tk.SetToken(ctx, tokenInfo)
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "gas price is too low")

	tokenInfo.GasPrice = correctPrice
	tk.SetToken(ctx, tokenInfo)

	//everything is ok
	result = keeper.OpcuAssetTransferWaitSign(ctx, orderID, hashes, rawData)
	require.Equal(t, sdk.CodeOK, result.Code)

	// OpcuAssetTransferSignFinish
	signedData := []byte("signedData")
	chainnodeTx.BlockHeight = 10000 //btc height = 10000

	// VerifyAccountSignedTransaction error
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, mock.Anything, signedData).Return(true, errors.New("VerifyAccountSignedTransactionError")).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "VerifyAccountSignedTransaction")

	// VerifyAccountSignedTransaction,  verified = false
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, mock.Anything, signedData).Return(false, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	// QueryAccountTransactionFromSignedData error
	mockCN.On("VerifyAccountSignedTransaction", chain, symbol, mock.Anything, signedData).Return(true, nil)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx, errors.New("QueryAccountTransactionFromSignedData")).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)

	chainnodeTx2 := &chainnode.ExtAccountTransaction{
		Hash:     opcuAstTransferTxHash,
		From:     opCUEthAddress,
		To:       newEpochAddress,
		Amount:   opCUEthAmt.SubRaw(1000000),
		Nonce:    0,
		GasLimit: sdk.NewInt(10000),
		GasPrice: sdk.NewInt(100),
	}

	// toAddr mismatch
	chainnodeTx2.To = "wrong_to"
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	chainnodeTx2.To = newEpochAddress

	// amount mismatch
	correctAmount = chainnodeTx2.Amount
	chainnodeTx2.Amount = chainnodeTx2.Amount.AddRaw(1)
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	chainnodeTx2.Amount = correctAmount

	// contract mismatch
	correctContract = chainnodeTx2.ContractAddress
	chainnodeTx2.ContractAddress = "wrong_contract"
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	chainnodeTx2.ContractAddress = correctContract

	// everything is ok
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(chainnodeTx2, nil).Once()
	result = keeper.OpcuAssetTransferSignFinish(ctx, orderID, signedData)
	require.Equal(t, sdk.CodeOK, result.Code)

	// OpcuAssetTransferFinish
	// empty order
	opcuAstTransferOrder := ok.GetOrder(ctx, orderID)
	ok.DeleteOrder(ctx, opcuAstTransferOrder)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeNotFoundOrder, result.Code)
	require.Contains(t, result.Log, "does not exist")

	orderStatus := opcuAstTransferOrder.GetOrderStatus()
	opcuAstTransferOrder.SetOrderStatus(sdk.OrderStatusFinish)
	ok.SetOrder(ctx, opcuAstTransferOrder)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	opcuAstTransferOrder.SetOrderStatus(orderStatus)
	ok.SetOrder(ctx, opcuAstTransferOrder)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[0].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[1].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)
	result = keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validators[2].GetOperator()), orderID, gasFee)
	require.Equal(t, sdk.CodeOK, result.Code)

	// check order
	o := ok.GetOrder(ctx, orderID)
	require.NotNil(t, o)
	require.Equal(t, orderID, o.GetID())
	require.Equal(t, sdk.OrderStatusFinish, o.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, o.GetOrderType())
	require.Equal(t, symbol, o.GetSymbol())
	require.Equal(t, ethOPCUAddr, o.GetCUAddress())
	require.Equal(t, rawData, o.(*sdk.OrderOpcuAssetTransfer).RawData)
	require.Equal(t, signedData, o.(*sdk.OrderOpcuAssetTransfer).SignedTx)
	require.Equal(t, gasFee, o.(*sdk.OrderOpcuAssetTransfer).CostFee)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeOpcuAssetTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, v := receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, v)
	require.Equal(t, symbol, of.Symbol.String())
	require.Equal(t, orderID, of.OrderID)
	require.Equal(t, sdk.OrderTypeOpcuAssetTransfer, of.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of.OrderStatus)
	require.Equal(t, ethOPCUAddr, of.CUAddress)

	wff, v := receipt.Flows[1].(sdk.OpcuAssetTransferFinishFlow)
	require.True(t, v)
	require.Equal(t, orderID, wff.OrderID)
	require.Equal(t, gasFee, wff.CostFee)

	opcu := newTestCU(ck.GetCU(ctx, ethOPCUAddr))
	require.Equal(t, sdk.MigrationFinish, opcu.GetMigrationStatus())
}
