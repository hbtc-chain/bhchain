package tests

import (
	"strings"
	"testing"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestDespoistEthToUserCUSuccessWithStandardAddress(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}
	symbol := token.EthToken
	chain := symbol

	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	require.Equal(t, true, tokenInfo.IsDepositEnabled)
	require.Equal(t, true, tokenInfo.IsSendEnabled)
	require.Equal(t, true, tokenInfo.IsWithdrawalEnabled)
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	hash := "0x84ed75bfad4b6d1c405a123990a1750974aa1f053394d442dfbc76090eeed44a"
	toAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	memo := "nice to see u"
	amt := sdk.TokensFromConsensusPower(1)
	orderID := uuid.NewV1().String()

	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	//mockCN.On("ValidAddress", chain, symbol, fromAddr).Return(true, true, fromAddr)
	toCU := ck.GetCU(ctx, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = toCU.AddAsset(symbol, toAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

	result := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
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
	require.Equal(t, uint64(0), df.Index)
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
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

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
	require.Equal(t, uint64(0), dls1[0].Index)

	require.NotPanics(t, func() {
		depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash, 0)
		require.Equal(t, dls1[0], depositItem)

	})

	dls2 := ck.GetDepositListByHash(ctx, symbol, toCUAddr, hash)
	require.Equal(t, dls1, dls2)

	//check userCU's coin
	toCU = ck.GetCU(ctx, toCUAddr)
	//require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
}

func TestDespoistEthToUserCUSuccessWithLowerToAddress(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}
	symbol := token.EthToken
	chain := symbol

	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	require.Equal(t, true, tokenInfo.IsDepositEnabled)
	require.Equal(t, true, tokenInfo.IsSendEnabled)
	require.Equal(t, true, tokenInfo.IsWithdrawalEnabled)
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	hash := "0x84ed75bfad4b6d1c405a123990a1750974aa1f053394d442dfbc76090eeed44a"
	toAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	memo := "nice to see u"
	amt := sdk.TokensFromConsensusPower(1)
	orderID := uuid.NewV1().String()

	mockCN.On("ValidAddress", chain, symbol, strings.ToLower(toAddr)).Return(true, toAddr) //to address is lower
	//mockCN.On("ValidAddress", chain, symbol, fromAddr).Return(true, true, fromAddr)
	toCU := ck.GetCU(ctx, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = toCU.AddAsset(symbol, toAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

	result := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), strings.ToLower(toAddr), hash, 0, amt, orderID, memo)
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
	require.Equal(t, uint64(0), df.Index)
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
	//	require.Equal(t, amt, bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	receipt, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls1))
	require.Equal(t, hash, dls1[0].Hash)
	require.Equal(t, amt, dls1[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls1[0].Status)
	require.Equal(t, memo, dls1[0].Memo)
	require.Equal(t, uint64(0), dls1[0].Index)

	require.NotPanics(t, func() {
		depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash, 0)
		require.Equal(t, dls1[0], depositItem)
	})

	dls2 := ck.GetDepositListByHash(ctx, symbol, toCUAddr, hash)
	require.Equal(t, dls1, dls2)

	//check userCU's coin
	toCU = ck.GetCU(ctx, toCUAddr)
	//	require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
}

func TestDespoistEthToUserCUSuccessWithUpperToAddress(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}
	symbol := token.EthToken
	chain := symbol
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	require.Equal(t, true, tokenInfo.IsDepositEnabled)
	require.Equal(t, true, tokenInfo.IsSendEnabled)
	require.Equal(t, true, tokenInfo.IsWithdrawalEnabled)
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	hash := "0x84ed75bfad4b6d1c405a123990a1750974aa1f053394d442dfbc76090eeed44a"
	toAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	memo := "nice to see u"
	amt := sdk.TokensFromConsensusPower(1)
	orderID := uuid.NewV1().String()

	mockCN.On("ValidAddress", chain, symbol, strings.ToUpper(toAddr)).Return(true, toAddr) //to address is lower
	//mockCN.On("ValidAddress", chain, symbol, fromAddr).Return(true, true, fromAddr)
	toCU := ck.GetCU(ctx, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = toCU.AddAsset(symbol, toAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

	result := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), strings.ToUpper(toAddr), hash, 0, amt, orderID, memo)
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
	require.Equal(t, uint64(0), df.Index)
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
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	receipt, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls1))
	require.Equal(t, hash, dls1[0].Hash)
	require.Equal(t, amt, dls1[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls1[0].Status)
	require.Equal(t, memo, dls1[0].Memo)
	require.Equal(t, uint64(0), dls1[0].Index)

	require.NotPanics(t, func() {
		depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash, 0)
		require.Equal(t, dls1[0], depositItem)
	})

	dls2 := ck.GetDepositListByHash(ctx, symbol, toCUAddr, hash)
	require.Equal(t, dls1, dls2)

	//check userCU's coin
	toCU = ck.GetCU(ctx, toCUAddr)
	//	require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
}

func TestDespoistEthToOPCUSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}
	symbol := token.EthToken
	chain := symbol
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))
	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	require.Equal(t, true, tokenInfo.IsDepositEnabled)
	require.Equal(t, true, tokenInfo.IsSendEnabled)
	require.Equal(t, true, tokenInfo.IsWithdrawalEnabled)
	tk.SetTokenInfo(ctx, tokenInfo)

	//
	toCUAddr, err := sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	require.Nil(t, err)

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)

	hash := "0x84ed75bfad4b6d1c405a123990a1750974aa1f053394d442dfbc76090eeed44a"
	toAddr := "0xc96d141c9110a8e61ed62caad8a7c858db15b82c"

	memo := "nice to see u"
	amt := sdk.TokensFromConsensusPower(1)
	orderID := uuid.NewV1().String()
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	//mockCN.On("ValidAddress", chain, symbol, fromAddr).Return(true, true, fromAddr)
	toCU := ck.NewCUWithAddress(ctx, sdk.CUTypeOp, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = toCU.SetAssetAddress(symbol, toAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

	result := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
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
	require.Equal(t, uint64(0), df.Index)
	require.Equal(t, sdk.DepositTypeOPCU, df.DepositType)

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
	require.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	//check receipt after
	receipt, err = rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid = receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of.OrderStatus)

	cf, valid := receipt.Flows[1].(sdk.DepositConfirmedFlow)
	require.True(t, valid)
	require.Equal(t, 1, len(cf.ValidOrderIDs))
	require.Equal(t, 0, len(cf.InValidOrderIDs))
	require.Equal(t, orderID, cf.ValidOrderIDs[0])

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	receipt, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	//check userCU's coin
	toCU = ck.GetCU(ctx, toCUAddr)
	//	require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))

}

func TestDespoistBtcToUserCUSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	mockCN = chainnode.MockChainnode{}
	symbol := token.BtcToken
	chain := symbol
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	require.Equal(t, true, tokenInfo.IsDepositEnabled)
	require.Equal(t, true, tokenInfo.IsSendEnabled)
	require.Equal(t, true, tokenInfo.IsWithdrawalEnabled)
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	hash := "9ae3c919d84f4b72802de6f4f4aa0d88abcc9fd57315ddf27b8e25f032e4a180"
	index := uint64(1)
	toAddr := "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"
	memo := ""
	amt := sdk.NewInt(85475551)
	orderID := uuid.NewV1().String()
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)

	toCU := ck.GetCU(ctx, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = toCU.AddAsset(symbol, toAddr, 1)
	require.Nil(t, err)
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
	require.Equal(t, uint64(1), df.Index)
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
	//	require.Equal(t, amt, bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	receipt, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls1))
	require.Equal(t, hash, dls1[0].Hash)
	require.Equal(t, amt, dls1[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls1[0].Status)
	require.Equal(t, memo, dls1[0].Memo)
	require.Equal(t, uint64(1), dls1[0].Index)

	require.NotPanics(t, func() {
		depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash, index)
		require.Equal(t, dls1[0], depositItem)
	})

	dls2 := ck.GetDepositListByHash(ctx, symbol, toCUAddr, hash)
	require.Equal(t, dls1, dls2)

	//check userCU's coin
	toCU = ck.GetCU(ctx, toCUAddr)
	//	require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
}

func TestDespoistBtcToOPCUSuccess(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	validators := input.validators
	symbol := token.BtcToken
	chain := symbol
	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))
	pubkey := ed25519.GenPrivKey().PubKey()

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	require.Equal(t, true, tokenInfo.IsDepositEnabled)
	require.Equal(t, true, tokenInfo.IsSendEnabled)
	require.Equal(t, true, tokenInfo.IsWithdrawalEnabled)
	tk.SetTokenInfo(ctx, tokenInfo)

	toCUAddr, err := sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	require.Nil(t, err)

	fromCUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)

	hash := "9ae3c919d84f4b72802de6f4f4aa0d88abcc9fd57315ddf27b8e25f032e4a180"
	index := uint64(1)
	toAddr := "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9"
	//canonicalFromAddr := fromAddr
	canonicalToAddr := toAddr
	memo := ""
	amt := sdk.NewInt(85475551)
	orderID := uuid.NewV1().String()
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, canonicalToAddr)
	//mockCN.On("ValidAddress", chain, symbol, fromAddr).Return(true, true, canonicalFromAddr)

	toCU := ck.GetCU(ctx, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(symbol, 1))
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	err = toCU.SetAssetAddress(symbol, canonicalToAddr, 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, canonicalToAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(symbol, 1))

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
	require.Equal(t, uint64(1), df.Index)
	require.Equal(t, sdk.DepositTypeOPCU, df.DepositType)

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
	require.Equal(t, sdk.OrderStatusFinish, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	//check receipt after deposit confirmed
	receipt, err = rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	of, valid = receipt.Flows[0].(sdk.OrderFlow)
	require.True(t, valid)
	require.Equal(t, sdk.OrderTypeDeposit, of.OrderType)
	require.Equal(t, sdk.OrderStatusFinish, of.OrderStatus)

	cf, valid := receipt.Flows[1].(sdk.DepositConfirmedFlow)
	require.True(t, valid)
	require.Equal(t, 1, len(cf.ValidOrderIDs))
	require.Equal(t, 0, len(cf.InValidOrderIDs))
	require.Equal(t, orderID, cf.ValidOrderIDs[0])

	//4th confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[3].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)
	receipt, err = rk.GetReceiptFromResult(&result)
	require.NotNil(t, err)

	dls1 = ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls1))
	require.Equal(t, hash, dls1[0].Hash)
	require.Equal(t, amt, dls1[0].Amount)
	require.Equal(t, sdk.DepositItemStatusConfirmed, dls1[0].Status)
	require.Equal(t, memo, dls1[0].Memo)
	require.Equal(t, uint64(1), dls1[0].Index)

	require.NotPanics(t, func() {
		depositItem := ck.GetDeposit(ctx, symbol, toCUAddr, hash, index)
		require.Equal(t, dls1[0], depositItem)
	})

	dls2 := ck.GetDepositListByHash(ctx, symbol, toCUAddr, hash)
	require.Equal(t, dls1, dls2)

	//check userCU's coin
	toCU = ck.GetCU(ctx, toCUAddr)
	//	require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
	require.Equal(t, amt, toCU.GetAssetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetAssetCoinsHold().AmountOf(symbol))
}

//{"FEBF0CA4CB4897C9A27A54275E612FEF275752AE", "HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463"},
//{"3BDA7843C6CE02FB1B274DF18F58E04354750EB8", "HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463"},
//{"CCE353A7008DD9E838691E5921D935848A0410F8", "HBCesEjnbc7wu2m6dTL8ekd45VrVAQzqYD7J"},
//{"4DBC8579C8A7453E7547A496AB07FB48F435B1F0", "HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"},

func TestDespoistEthToUserCUError(t *testing.T) {
	input := setupTestInput(t)
	keeper := input.k
	ctx := input.ctx
	ck := input.ck
	tk := input.tk
	ok := input.ok
	rk := input.rk
	//cn := input.cn
	mockCN = chainnode.MockChainnode{}
	chain := token.EthToken
	symbol := chain
	validators := input.validators
	pubkey := ed25519.GenPrivKey().PubKey()

	require.Equal(t, 1, len(ck.GetOpCUs(ctx, symbol)))

	toCUAddr, err := sdk.CUAddressFromBase58("HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463")
	require.Nil(t, err)
	fromCUAddr := toCUAddr

	tokenInfo := tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tk.SetTokenInfo(ctx, tokenInfo)

	hash := "0x84ed75bfad4b6d1c405a123990a1750974aa1f053394d442dfbc76090eeed44a"
	index := uint64(0)
	//fromAddr := "0xd139E358aE9cB5424B2067da96F94cC938343446"
	toAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	memo := "nice to see u"
	amt := sdk.TokensFromConsensusPower(1)
	orderID := "b8d36bd2-6ab6-11ea-b7ed-f218982f5e9c"

	require.Equal(t, 0, len(ok.GetProcessOrderList(ctx)))

	//illegal id
	result := keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, "illegalid", memo)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.Contains(t, result.Log, "invalid OrderID")

	//fromCU does not exist
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)
	require.Contains(t, result.Log, "HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463")

	//toCU does not exist
	fromCUAddr, err = sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	require.Nil(t, err)
	toCUAddr, err = sdk.CUAddressFromBase58("HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463")
	require.Nil(t, err)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
	require.Equal(t, sdk.CodeInvalidAccount, result.Code)
	require.Contains(t, result.Log, "HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463")

	//token doesnot support
	bheos := "bheos"
	toCU := ck.NewCUWithAddress(ctx, sdk.CUTypeUser, toCUAddr)
	require.Equal(t, "", toCU.GetAssetAddress(bheos, 1))
	err = toCU.AddAsset(bheos, toAddr, 1)
	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	require.Nil(t, err)
	ck.SetCU(ctx, toCU)
	require.Equal(t, toAddr, ck.GetCU(ctx, toCUAddr).GetAssetAddress(bheos, 1))
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(bheos), toAddr, hash, 0, sdk.NewInt(100), orderID, memo)
	require.Equal(t, sdk.CodeUnsupportToken, result.Code)

	//token's IsSendEnabled = false
	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//token's IsDepositEnable = false
	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tokenInfo.IsSendEnabled = true
	tokenInfo.IsDepositEnabled = false
	tk.SetTokenInfo(ctx, tokenInfo)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)

	//transfer's SendEnable = false
	tokenInfo = tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	tokenInfo.IsDepositEnabled = true
	tk.SetTokenInfo(ctx, tokenInfo)
	keeper.SetSendEnabled(ctx, false)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID, memo)
	require.Equal(t, sdk.CodeTransactionIsNotEnabled, result.Code)
	require.Equal(t, 0, len(ok.GetProcessOrderList(ctx)))

	//amt < DepositThreshold
	keeper.SetSendEnabled(ctx, true)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, tokenInfo.DepositThreshold.Sub(sdk.NewInt(1)), orderID, memo)
	require.Equal(t, sdk.CodeInsufficientCoins, result.Code)

	//toAddr is different
	mockCN.On("ValidAddress", chain, symbol, "0xc96d141c9110a8e61ed62caad8a7c858db15b82d").Return(true, "0xC96d141C9110a8e61ED62CAaD8A7c858dB15B82D").Once()
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), "0xc96d141c9110a8e61ed62caad8a7c858db15b82d", hash, 0, amt, orderID, memo)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//	require.Contains(t, result.Log, "address is 0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c, not equal to")
	require.Equal(t, 0, len(ok.GetProcessOrderList(ctx)))

	//uuid already exist
	mockCN.On("ValidAddress", chain, symbol, toAddr).Return(true, toAddr)
	orderID1 := uuid.NewV1().String()
	order1 := &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: toCUAddr,
			ID:        orderID1,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
		},
	}
	//order1 := ok.NewOrderCollect(ctx, toCUAddr, orderID1, symbol, toCUAddr, toCU.GetAssetAddress(symbol, 1), amt, sdk.ZeroInt(), sdk.ZeroInt(), hash, height, index, memo)
	ok.SetOrder(ctx, order1)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, 0, amt, orderID1, memo)
	//	require.Equal(t, sdk.CodeInvalidOrder, result.Code)
	//	require.Contains(t, result.Log, "already exists")

	//Deposit already exist
	deposit := ck.GetDeposit(ctx, symbol, toCUAddr, hash, index)
	require.NoError(t, err)
	require.Equal(t, sdk.DepositNil, deposit)
	d := sdk.DepositItem{
		Hash:   hash,
		Index:  index,
		Amount: amt,
	}
	ck.SaveDeposit(ctx, symbol, toCUAddr, d)
	deposit = ck.GetDeposit(ctx, symbol, toCUAddr, d.Hash, d.Index)
	require.NoError(t, err)
	require.Equal(t, d, deposit)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, index, amt, orderID, memo)
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//	require.Contains(t, result.Log, "item already exist")
	ck.DelDeposit(ctx, symbol, toCUAddr, d.Hash, d.Index)

	toCU.SetAssetPubkey(pubkey.Bytes(), 1)
	toCU.AddAsset(symbol, toAddr, 1)
	ck.SetCU(ctx, toCU)
	result = keeper.Deposit(ctx, fromCUAddr, toCUAddr, sdk.Symbol(symbol), toAddr, hash, index, amt, orderID, memo)
	require.Equal(t, sdk.CodeOK, result.Code)

	//check order
	order := ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositUnconfirm, int(order.(*sdk.OrderCollect).DepositStatus))
	require.Equal(t, symbol, order.GetSymbol())

	/*Confirm deposit*/
	//Not from a validator
	result = keeper.ConfirmedDeposit(ctx, toCUAddr, []string{orderID}, []string{})
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	//	require.Contains(t, result.Log, "depositconfirm from not a validator")

	//for a non validator
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{uuid.NewV1().String()}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)

	//1st confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[0].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)

	//2nd confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[1].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)

	//3rd confirm
	result = keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validators[2].OperatorAddress), []string{orderID}, []string{})
	require.Equal(t, sdk.CodeOK, result.Code)

	//check deposit item
	dls1 := ck.GetDepositList(ctx, symbol, toCUAddr)
	require.Equal(t, 1, len(dls1))
	require.Equal(t, hash, dls1[0].Hash)
	require.Equal(t, amt, dls1[0].Amount)
	require.Equal(t, sdk.DepositItemStatusWaitCollect, dls1[0].Status)
	require.Equal(t, memo, dls1[0].Memo)
	require.Equal(t, uint64(0), dls1[0].Index)

	dls2 := ck.GetDepositListByHash(ctx, symbol, toCUAddr, hash)
	require.Equal(t, dls1, dls2)

	//check receipt
	receipt, err := rk.GetReceiptFromResult(&result)
	require.Nil(t, err)
	require.Equal(t, sdk.CategoryTypeDeposit, receipt.Category)
	require.Equal(t, 3, len(receipt.Flows))

	of, valid := receipt.Flows[0].(sdk.OrderFlow)
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
	//	require.Equal(t, amt, bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	//check order
	order = ok.GetOrder(ctx, orderID)
	require.NotNil(t, order)
	require.Equal(t, orderID, order.GetID())
	require.Equal(t, sdk.OrderStatusBegin, order.GetOrderStatus())
	require.Equal(t, sdk.OrderTypeCollect, order.GetOrderType())
	require.Equal(t, sdk.DepositConfirmed, int(order.(*sdk.OrderCollect).DepositStatus))

	//check CU coins
	toCU = ck.GetCU(ctx, toCUAddr)
	//	require.Equal(t, amt, toCU.GetCoins().AmountOf(symbol))
	require.Equal(t, sdk.ZeroInt(), toCU.GetCoinsHold().AmountOf(symbol))
}
