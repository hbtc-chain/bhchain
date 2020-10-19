package keeper

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func (keeper BaseKeeper) Withdrawal(ctx sdk.Context, fromCUAddr sdk.CUAddress, toAddr, orderID, symbol string, amt, gasFee sdk.Int) sdk.Result {
	if sdk.IsIllegalOrderID(orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("invalid OrderID:%v", orderID)).Result()
	}

	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount(fromCUAddr.String()).Result()
	}
	if fromCU.GetCUType() != sdk.CUTypeUser {
		return sdk.ErrInvalidTx(fmt.Sprintf("withdrawal from a non user CU :%v", fromCUAddr)).Result()
	}

	if !sdk.Symbol(symbol).IsValidTokenName() {
		return sdk.ErrInvalidSymbol(symbol).Result()
	}

	if !keeper.tk.IsTokenSupported(ctx, sdk.Symbol(symbol)) {
		return sdk.ErrUnSupportToken(symbol).Result()
	}

	tokenInfo := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	chain := tokenInfo.Chain.String()

	valid, canonicalToAddr := keeper.cn.ValidAddress(chain, symbol, toAddr)
	if !valid {
		return sdk.ErrInvalidAddr(fmt.Sprintf("%v is not a valid address", toAddr)).Result()
	}

	toCUAddr, _ := keeper.ck.GetCUFromExtAddress(ctx, chain, canonicalToAddr)
	if toCUAddr != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("withdrawal to a chain CU :%v not support, use send cmd directly instead", toCUAddr)).Result()
	}

	if !tokenInfo.IsWithdrawalEnabled || !tokenInfo.IsSendEnabled || !keeper.GetSendEnabled(ctx) {
		return sdk.ErrTransactionIsNotEnabled(fmt.Sprintf("%v's withdraw is not enabled temporary", symbol)).Result()
	}

	if keeper.ok.IsExist(ctx, orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("order %v already exists", orderID)).Result()
	}

	if gasFee.LT(tokenInfo.WithDrawalFee().Amount) {
		return sdk.ErrInsufficientFee(fmt.Sprintf("need:%v, actual have:%v", tokenInfo.WithDrawalFee(), gasFee)).Result()
	}

	if !amt.IsPositive() {
		return sdk.ErrInvalidAmount(fmt.Sprintf("amt:%v", amt)).Result()
	}

	feeCoins := sdk.NewCoins(sdk.NewCoin(chain, gasFee))
	coins := sdk.NewCoins(sdk.NewCoin(symbol, amt))
	need := coins.Add(feeCoins)
	have := fromCU.GetCoins()

	if chain == symbol {
		if have.AmountOf(chain).LT(need.AmountOf(chain)) {
			return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, actual have:%v", need, have)).Result()
		}
	} else {
		if have.AmountOf(chain).LT(gasFee) || have.AmountOf(symbol).LT(amt) {
			return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, actual have:%v", need, have)).Result()
		}
	}

	withdrawalOrder := keeper.ok.NewOrderWithdrawal(ctx, fromCUAddr, orderID, symbol, amt, gasFee, sdk.ZeroInt(), canonicalToAddr, "", "")
	if withdrawalOrder == nil {
		return sdk.ErrInvalidOrder(fmt.Sprintf("Fail to create order:%v", orderID)).Result()
	}
	if tokenInfo.TokenType == sdk.UtxoBased {
		withdrawalOrder.WithdrawStatus = sdk.WithdrawStatusValid
	} else {
		withdrawalOrder.WithdrawStatus = sdk.WithdrawStatusUnconfirmed
	}
	keeper.ok.SetOrder(ctx, withdrawalOrder)

	//onhold needed coins
	fromCU.ResetBalanceFlows()
	fromCU.SubCoins(need)
	fromCU.AddCoinsHold(need)
	keeper.ck.SetCU(ctx, fromCU)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), withdrawalOrder.GetCUAddress(), withdrawalOrder.GetID(), sdk.OrderTypeWithdrawal, sdk.OrderStatusBegin))
	flows = append(flows, keeper.rk.NewWithdrawalFlow(orderID, fromCUAddr.String(), canonicalToAddr, symbol, amt, gasFee, withdrawalOrder.WithdrawStatus))

	for _, balanceFlow := range fromCU.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}

	fromCU.ResetBalanceFlows()
	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeWithdrawal, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

// WithdrawalConfirm confirm account type withdrawal order
func (keeper BaseKeeper) WithdrawalConfirm(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, valid bool) sdk.Result {
	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("withdrawal from not a validator :%v", fromCUAddr)).Result()
	}

	order := keeper.ok.GetOrder(ctx, orderID)
	if order == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("orderid:%v does not exist", orderID)).Result()
	}
	withdrawOrder, ok := order.(*sdk.OrderWithdrawal)
	if !ok {
		return sdk.ErrInvalidTx(fmt.Sprintf("Not withdraw order")).Result()
	}

	if withdrawOrder.GetOrderStatus() != sdk.OrderStatusBegin {
		return sdk.ErrInvalidTx(fmt.Sprintf("invalid order status")).Result()
	}
	tokenInfo := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(withdrawOrder.Symbol))
	if tokenInfo == nil {
		return sdk.ErrInternal(fmt.Sprintf("token %s not exists", withdrawOrder.Symbol)).Result()
	}
	if tokenInfo.TokenType == sdk.UtxoBased {
		return sdk.ErrInvalidTx("unexpected token type").Result()
	}

	result := sdk.Result{}
	confirmedFirstTime, _, _ := keeper.evidenceKeeper.Vote(ctx, orderID, fromCUAddr, types.NewTxVote(0, valid), uint64(ctx.BlockHeight()))
	if confirmedFirstTime {
		var flows []sdk.Flow
		var balanceFlows []sdk.Flow

		if valid {
			withdrawOrder.WithdrawStatus = sdk.WithdrawStatusValid
		} else {
			withdrawOrder.WithdrawStatus = sdk.WithdrawStatusInvalid
			withdrawOrder.SetOrderStatus(sdk.OrderStatusCancel)

			// unfreeze user fund
			withdrawFromCU := keeper.ck.GetCU(ctx, withdrawOrder.CUAddress)
			withdrawFromCU.ResetBalanceFlows()

			feeCoins := sdk.NewCoins(sdk.NewCoin(tokenInfo.Chain.String(), withdrawOrder.GasFee))
			withdrawCoins := sdk.NewCoins(sdk.NewCoin(withdrawOrder.Symbol, withdrawOrder.Amount))
			totalCoins := withdrawCoins.Add(feeCoins)

			withdrawFromCU.SubCoinsHold(totalCoins)
			withdrawFromCU.AddCoins(totalCoins)

			keeper.ck.SetCU(ctx, withdrawFromCU)
			for _, balanceFlow := range withdrawFromCU.GetBalanceFlows() {
				balanceFlows = append(balanceFlows, balanceFlow)
			}
			withdrawFromCU.ResetBalanceFlows()
		}
		keeper.ok.SetOrder(ctx, withdrawOrder)

		flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(withdrawOrder.Symbol), withdrawOrder.GetCUAddress(), withdrawOrder.GetID(), sdk.OrderTypeWithdrawal, sdk.OrderStatusBegin))
		flows = append(flows, keeper.rk.NewWithdrawalConfirmFlow(orderID, withdrawOrder.WithdrawStatus))
		flows = append(flows, balanceFlows...)
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeWithdrawal, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}
	return result
}

func (keeper BaseKeeper) WithdrawalWaitSign(ctx sdk.Context, opCUAddr sdk.CUAddress, orderIDs, signHashes []string, rawData []byte) sdk.Result {
	tokenInfo, withdrawalOrders, err := keeper.checkWithdrawalOrders(ctx, orderIDs, sdk.OrderStatusBegin)
	if err != nil {
		return err.Result()
	}

	symbol := keeper.ok.GetOrder(ctx, orderIDs[0]).GetSymbol()
	chain := tokenInfo.Chain.String()

	curEpoch := keeper.sk.GetCurrentEpoch(ctx)

	opCU := keeper.ck.GetCU(ctx, opCUAddr)
	if opCU == nil {
		return sdk.ErrInvalidAccount(fmt.Sprintf("CU %v does not exist", opCUAddr)).Result()
	}
	if opCU.GetMigrationStatus() != sdk.MigrationFinish {
		return sdk.ErrInvalidTx(fmt.Sprintf("OPCU %v is in migration", opCU)).Result()
	}
	fromAddr := opCU.GetAssetAddress(symbol, curEpoch.Index)
	err = keeper.checkWithdrawalOpCU(opCU, chain, symbol, true, fromAddr)
	if err != nil {
		return err.Result()
	}

	//Retrieve gas Price
	gasPrice := tokenInfo.GasPrice
	if chain != symbol {
		ti := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(chain))
		if ti == nil {
			return sdk.ErrInvalidSymbol(fmt.Sprintf("%s does not exist", chain)).Result()
		}
		gasPrice = ti.GasPrice
	}

	priceUpLimit := sdk.NewDecFromInt(gasPrice).Mul(PriceUpLimitRatio)
	priceLowLimit := sdk.NewDecFromInt(gasPrice).Mul(PriceLowLimitRatio)
	costFee := sdk.ZeroInt()
	utxoInNum := 0
	var coins sdk.Coins
	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		if len(orderIDs) > sdk.MaxVoutNum {
			return sdk.ErrInvalidTx(fmt.Sprintf("contains too many vouts %v", len(orderIDs))).Result()
		}

		//formulate the vins
		vins, err := keeper.cn.QueryUtxoInsFromData(chain, symbol, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		for _, vin := range vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCUAddr, vin.Hash, vin.Index)
			if item == sdk.DepositNil {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v does not exist", vin.Hash, vin.Index)).Result()
			}

			if item.GetStatus() != sdk.DepositItemStatusConfirmed {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v status is %v", vin.Hash, vin.Index, item.GetStatus())).Result()
			}

			vin.Address = item.ExtAddress
			vin.Amount = item.Amount

			utxoInNum++
		}

		tx, hashes, err := keeper.cn.QueryUtxoTransactionFromData(chain, symbol, rawData, vins)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		amt, sdkErr := keeper.checkDecodedUtxoTransaction(ctx, symbol, opCUAddr, withdrawalOrders, tx, fromAddr)
		if sdkErr != nil {
			return sdkErr.Result()
		}

		//Estimate SignedTx Size and calculate price
		size := sdk.EstimateSignedUtxoTxSize(len(tx.Vins), len(tx.Vouts)).ToDec()
		price := sdk.NewDecFromInt(tx.CostFee).MulInt64(sdk.KiloBytes).Quo(size)

		if price.GT(priceUpLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too high, actual:%v, uplimit:%v", price, priceUpLimit)).Result()
		}
		if price.LT(priceLowLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too low, actual:%v, lowlimit:%v", price, priceLowLimit)).Result()
		}

		// check hashes
		if len(hashes) != len(signHashes) {
			return sdk.ErrInvalidTx(fmt.Sprintf("signhashes's number mismatch, expected:%v, have:%v", len(hashes), len(signHashes))).Result()
		}

		for i := 0; i < len(hashes); i++ {
			if string(hashes[i]) != signHashes[i] {
				return sdk.ErrInvalidTx(fmt.Sprintf("mismatch hashes, expected:%v, have:%v", hashes[i], signHashes[i])).Result()
			}
		}
		costFee = tx.CostFee
		coins = sdk.NewCoins(sdk.NewCoin(symbol, amt)) //in BTC, locked the sum(Vin)

		for _, vin := range vins {
			_ = keeper.ck.SetDepositStatus(ctx, symbol, opCUAddr, vin.Hash, vin.Index, sdk.DepositItemStatusInProcess)
		}

	case sdk.AccountBased:
		if len(signHashes) != sdk.LimitAccountBasedOrderNum || len(orderIDs) != sdk.LimitAccountBasedOrderNum {
			return sdk.ErrInvalidTx(fmt.Sprintf("AccountBased token supports only one withdrawal at one time, ordernum:%v, signhashnum:%v", len(orderIDs), len(signHashes))).Result()
		}

		tx, hash, err := keeper.cn.QueryAccountTransactionFromData(chain, symbol, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		withdrawalOrder := withdrawalOrders[0]
		if withdrawalOrder.OpCUaddress != "" && withdrawalOrder.OpCUaddress != opCUAddr.String() {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected withdrawal reset from address:%v, expected:%v", opCUAddr.String(), withdrawalOrder.OpCUaddress)).Result()
		}

		if tx.To != withdrawalOrder.WithdrawToAddress {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected withdrawal to address:%v, expected:%v", tx.To, withdrawalOrder.WithdrawToAddress)).Result()
		}

		if !tx.Amount.Equal(withdrawalOrder.Amount) {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected withdrawal Amount:%v, expected:%v", tx.Amount, withdrawalOrder.Amount)).Result()
		}

		validContractAddr := ""
		if tokenInfo.Issuer != "" {
			_, validContractAddr = keeper.cn.ValidAddress(tokenInfo.Chain.String(), tokenInfo.Symbol.String(), tokenInfo.Issuer)
		}
		if tx.ContractAddress != validContractAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected withdrawal contract address:%v, expected:%v", tx.ContractAddress, tokenInfo.Issuer)).Result()
		}

		if string(hash) != signHashes[0] {
			return sdk.ErrInvalidTx(fmt.Sprintf("hash mismatch, expected:%v, have:%v", hash, signHashes[0])).Result()
		}

		if !tokenInfo.GasLimit.Equal(tx.GasLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas limit mismatch, expected:%v, have:%v", tokenInfo.GasLimit, tx.GasLimit)).Result()
		}

		if sdk.NewDecFromInt(tx.GasPrice).GT(priceUpLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too high, actual:%v, uplimit:%v", tx.GasPrice, priceUpLimit)).Result()
		}

		if sdk.NewDecFromInt(tx.GasPrice).LT(priceLowLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too low, actual:%v, lowlimit:%v", tx.GasPrice, priceLowLimit)).Result()
		}

		costFee = tx.GasPrice.Mul(tx.GasLimit)
		if costFee.GT(withdrawalOrder.GasFee) {
			return sdk.ErrGasOverflow(fmt.Sprintf("actual gas:%v > gas uplimit:%v", costFee, withdrawalOrder.GasFee)).Result()
		}

		nonce := opCU.GetNonce(tokenInfo.Chain.String(), fromAddr)
		if nonce != tx.Nonce {
			return sdk.ErrInvalidTx(fmt.Sprintf("tx nonce not equal, opcu :%v, rawdata:%v", nonce, tx.Nonce)).Result()
		}

		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, costFee))
		coins = sdk.NewCoins(sdk.NewCoin(symbol, tx.Amount))
		coins = coins.Add(feeCoins)

		//No need to check gr-gu for opCU, check AssetCoins directly
		have := opCU.GetAssetCoins()
		if chain == symbol {
			if have.AmountOf(chain).LT(coins.AmountOf(chain)) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, actual have:%v", coins, have)).Result()
			}
		} else {
			if have.AmountOf(chain).LT(costFee) || have.AmountOf(symbol).LT(tx.Amount) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, actual have:%v", coins, have)).Result()
			}
		}

		opCU.SetEnableSendTx(false, chain, fromAddr)

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update opCU's assetCoins and assetCoinsHold, Lock the opCU
	opCU.SubAssetCoins(coins)
	opCU.AddAssetCoinsHold(coins)
	keeper.ck.SetCU(ctx, opCU)

	//update order's status
	for _, orderID := range orderIDs {
		withdrawalOrder := keeper.ok.GetOrder(ctx, orderID).(*sdk.OrderWithdrawal)
		withdrawalOrder.OpCUaddress = opCUAddr.String()
		withdrawalOrder.FromAddress = fromAddr
		withdrawalOrder.CostFee = costFee
		withdrawalOrder.UtxoInNum = utxoInNum
		withdrawalOrder.OrderBase.Height = uint64(ctx.BlockHeight())
		withdrawalOrder.Status = sdk.OrderStatusWaitSign
		withdrawalOrder.RawData = make([]byte, len(rawData))
		copy(withdrawalOrder.RawData, rawData)
		keeper.ok.SetOrder(ctx, withdrawalOrder)
	}

	var flows []sdk.Flow
	withdrawalOrder := withdrawalOrders[0]
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), withdrawalOrder.GetCUAddress(), withdrawalOrder.GetID(), sdk.OrderTypeWithdrawal, sdk.OrderStatusWaitSign))
	flows = append(flows, keeper.rk.NewWithdrawalWaitSignFlow(orderIDs, opCUAddr.String(), fromAddr, rawData))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeWithdrawal, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

func (keeper BaseKeeper) WithdrawalSignFinish(ctx sdk.Context, orderIDs []string, signedTx []byte, tansHash string) sdk.Result {
	tokenInfo, withdrawalOrders, err := keeper.checkWithdrawalOrders(ctx, orderIDs, sdk.OrderStatusWaitSign)
	if err != nil {
		return err.Result()
	}

	order := withdrawalOrders[0]
	symbol := tokenInfo.Symbol.String()
	chain := tokenInfo.Chain.String()

	opCUAddr, _ := sdk.CUAddressFromBase58(withdrawalOrders[0].OpCUaddress)
	sendable := tokenInfo.TokenType == sdk.UtxoBased

	opCU := keeper.ck.GetCU(ctx, opCUAddr)
	if opCU == nil {
		return sdk.ErrInvalidAccount(fmt.Sprintf("CU %v does not exist", opCUAddr)).Result()
	}

	err = keeper.checkWithdrawalOpCU(opCU, chain, symbol, sendable, order.FromAddress)
	if err != nil {
		return err.Result()
	}
	var txHash string

	switch tokenInfo.TokenType {
	case sdk.UtxoBased:

		result, hash := keeper.verifyUtxoBasedSignedTx(ctx, nil, opCUAddr, chain, symbol, withdrawalOrders[0].RawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}

		txHash = hash

	case sdk.AccountBased:
		result, hash := keeper.verifyAccountBasedSignedTx(order.FromAddress, chain, symbol, withdrawalOrders[0].RawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}

		txHash = hash

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update order's status
	for _, orderID := range orderIDs {
		withdrawalOrder := keeper.ok.GetOrder(ctx, orderID).(*sdk.OrderWithdrawal)
		withdrawalOrder.Status = sdk.OrderStatusSignFinish
		withdrawalOrder.SignedTx = make([]byte, len(signedTx))
		withdrawalOrder.Txhash = txHash
		copy(withdrawalOrder.SignedTx, signedTx)
		keeper.ok.SetOrder(ctx, withdrawalOrder)
	}

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), order.GetID(), sdk.OrderTypeWithdrawal, sdk.OrderStatusSignFinish))
	flows = append(flows, keeper.rk.NewWithdrawalSignFinishFlow(orderIDs, signedTx, tansHash))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeWithdrawal, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) WithdrawalFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, costFee sdk.Int, valid bool) sdk.Result {
	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("withdrawal from not a validator :%v", fromCUAddr)).Result()
	}

	ord := keeper.ok.GetOrder(ctx, orderIDs[0])
	if ord == nil {
		err := sdk.ErrNotFoundOrder(fmt.Sprintf("orderid:%v does not exist", orderIDs[0]))
		return err.Result()
	}

	withdrawalOrder, valid := ord.(*sdk.OrderWithdrawal)
	if !valid {
		err := sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not withdrawal order", orderIDs[0]))
		return err.Result()
	}

	confirmedFirstTime, _, _ := keeper.evidenceKeeper.Vote(ctx, withdrawalOrder.Txhash, fromCUAddr, types.NewTxVote(costFee.Int64(), true), uint64(ctx.BlockHeight()))
	result := sdk.Result{}
	if !confirmedFirstTime {
		return result
	}

	tokenInfo, withdrawalOrders, err := keeper.checkWithdrawalOrders(ctx, orderIDs, sdk.OrderStatusSignFinish)
	if err != nil {
		return err.Result()
	}

	order := withdrawalOrders[0]
	symbol := tokenInfo.Symbol.String()
	chain := tokenInfo.Chain.String()
	opCUAddr, _ := sdk.CUAddressFromBase58(order.OpCUaddress)

	opCU := keeper.ck.GetCU(ctx, opCUAddr)
	if opCU == nil {
		return sdk.ErrInvalidAccount(fmt.Sprintf("CU %v does not exist", opCUAddr)).Result()
	}

	sendable := tokenInfo.TokenType == sdk.UtxoBased
	err = keeper.checkWithdrawalOpCU(opCU, chain, symbol, sendable, order.FromAddress)
	if err != nil {
		return err.Result()
	}

	hash := order.Txhash

	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		vins, err := keeper.cn.QueryUtxoInsFromData(chain, symbol, order.RawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		for _, vin := range vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCUAddr, vin.Hash, vin.Index)
			if item == sdk.DepositNil {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v does not exist", vin.Hash, vin.Index)).Result()
			}

			if item.GetStatus() != sdk.DepositItemStatusInProcess {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v status is %v", vin.Hash, vin.Index, item.GetStatus())).Result()
			}

			vin.Address = item.ExtAddress
			vin.Amount = item.Amount
		}

		tx, err := keeper.cn.QueryUtxoTransactionFromSignedData(chain, symbol, order.SignedTx, vins)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction from signed transaction:%v", order.SignedTx)).Result()
		}

		//check the change and update deposit item's status
		backAmount := sdk.ZeroInt()
		for i, vout := range tx.Vouts {
			if order.FromAddress == vout.Address {
				vin := sdk.NewUtxoIn(tx.Hash, uint64(i), vout.Amount, vout.Address)
				backAmount = backAmount.Add(vout.Amount)
				//formulate the changeback deposit item
				depositItem, err := sdk.NewDepositItem(vin.Hash, vin.Index, vin.Amount, vout.Address, "", sdk.DepositItemStatusConfirmed)
				if err != nil {
					return sdk.ErrInvalidOrder(fmt.Sprintf("fail to create deposit item, %v %v %v", vin.Hash, vin.Index, vin.Amount)).Result()
				}
				_ = keeper.ck.SaveDeposit(ctx, symbol, opCU.GetAddress(), depositItem)
			}
		}

		//delete used Vins from opCU
		inAmt := sdk.ZeroInt()
		for _, vin := range tx.Vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCU.GetAddress(), vin.Hash, vin.Index)
			if item == sdk.DepositNil {
				return sdk.ErrInvalidOrder(fmt.Sprintf("deposit item%v %v does not exist", vin.Hash, vin.Index)).Result()
			}
			keeper.ck.DelDeposit(ctx, symbol, opCUAddr, vin.Hash, vin.Index)
			inAmt = inAmt.Add(vin.Amount)
		}

		opCU.SubAssetCoinsHold(sdk.NewCoins(sdk.NewCoin(symbol, inAmt)))
		opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(symbol, backAmount)))

		//calculate Op's gr/gu
		grAmt := sdk.ZeroInt()
		for _, wOrder := range withdrawalOrders {
			grAmt = grAmt.Add(wOrder.GasFee) //accumulated each order's gasFee
		}
		opCU.AddGasReceived(sdk.NewCoins(sdk.NewCoin(chain, grAmt)))
		opCU.AddGasUsed(sdk.NewCoins(sdk.NewCoin(chain, costFee)))

	case sdk.AccountBased:
		tx, err := keeper.cn.QueryAccountTransactionFromSignedData(chain, symbol, order.SignedTx)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction from signed transaction:%v", order.SignedTx)).Result()
		}
		if tx.Hash != hash {
			return sdk.ErrInvalidTx(fmt.Sprintf("hash mismatch, expected: %v, actual:%v", hash, tx.Hash)).Result()
		}

		//update opcu's assetcoinshold, and refund unused gas fee if necessary
		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, order.CostFee))
		coins := sdk.NewCoins(sdk.NewCoin(symbol, tx.Amount))
		coins = coins.Add(feeCoins)
		opCU.SubAssetCoinsHold(coins)
		if !order.CostFee.Equal(costFee) {
			opCU.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, order.CostFee.Sub(costFee))))
		}

		//update order.CostFee, gasused, add user gived GasFee into gasReceived
		order.CostFee = costFee
		keeper.ok.SetOrder(ctx, order)

		opCU.AddGasReceived(sdk.NewCoins(sdk.NewCoin(chain, order.GasFee)))
		opCU.AddGasUsed(sdk.NewCoins(sdk.NewCoin(chain, costFee)))
		if tokenInfo.IsNonceBased {
			//don't update local nonce
			nonce := opCU.GetNonce(chain, order.FromAddress) + 1
			opCU.SetNonce(chain, nonce, order.FromAddress)
		}
		opCU.SetEnableSendTx(true, chain, order.FromAddress)

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update order's status and costFee
	var balanceFlows []sdk.Flow
	for _, orderID := range orderIDs {
		withdrawalOrder := keeper.ok.GetOrder(ctx, orderID).(*sdk.OrderWithdrawal)
		fromCUAddr := withdrawalOrder.GetCUAddress()
		fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
		fromCU.ResetBalanceFlows()
		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, withdrawalOrder.GasFee))
		withdrawCoins := sdk.NewCoins(sdk.NewCoin(symbol, withdrawalOrder.Amount))
		totalCoins := withdrawCoins.Add(feeCoins)

		fromCU.SubCoinsHold(totalCoins)

		//back gasfee if it is the mainnet asset
		if withdrawalOrder.GasFee.GT(costFee) && withdrawalOrder.Symbol == chain {
			fromCU.AddCoins(sdk.NewCoins(sdk.NewCoin(chain, withdrawalOrder.GasFee.Sub(costFee))))
			withdrawalOrder.GasFee = costFee
		}

		if valid {
			withdrawalOrder.Status = sdk.OrderStatusFinish
		} else {
			withdrawalOrder.Status = sdk.OrderStatusFailed
			fromCU.AddCoins(withdrawCoins) // tx failed, add back user's coin
		}
		keeper.ok.SetOrder(ctx, withdrawalOrder)

		keeper.ck.SetCU(ctx, fromCU)
		for _, balanceFlow := range fromCU.GetBalanceFlows() {
			balanceFlows = append(balanceFlows, balanceFlow)
		}
		fromCU.ResetBalanceFlows()
	}

	keeper.ck.SetCU(ctx, opCU)

	var flows []sdk.Flow
	withdrawalOrder = withdrawalOrders[0]
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), opCUAddr, withdrawalOrder.GetID(), sdk.OrderTypeWithdrawal, sdk.OrderStatusFinish))
	flows = append(flows, keeper.rk.NewWithdrawalFinishFlow(orderIDs, costFee, valid))
	flows = append(flows, balanceFlows...)

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeWithdrawal, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) checkWithdrawalOrders(ctx sdk.Context, orderIDs []string, orderStatus sdk.OrderStatus) (tokenInfo *sdk.TokenInfo, withdrawalOrders []*sdk.OrderWithdrawal, err sdk.Error) {
	order := keeper.ok.GetOrder(ctx, orderIDs[0])
	if order == nil {
		err = sdk.ErrNotFoundOrder(fmt.Sprintf("orderid:%v does not exist", orderIDs[0]))
		return
	}
	firstStatus := order.GetOrderStatus()

	withdrawalOrder, valid := order.(*sdk.OrderWithdrawal)
	if !valid {
		err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not withdrawal order", orderIDs[0]))
		return
	}

	symbol := order.GetSymbol()
	if !sdk.Symbol(symbol).IsValidTokenName() {
		err = sdk.ErrInvalidSymbol(symbol)
		return
	}

	if !keeper.tk.IsTokenSupported(ctx, sdk.Symbol(symbol)) {
		err = sdk.ErrUnSupportToken(symbol)
		return
	}

	tokenInfo = keeper.tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	chain := tokenInfo.Chain.String()
	if !tokenInfo.IsWithdrawalEnabled || !tokenInfo.IsSendEnabled || !keeper.GetSendEnabled(ctx) {
		err = sdk.ErrTransactionIsNotEnabled(fmt.Sprintf("%v's withdrawal is not enabled temporary", symbol))
		return
	}

	hash := withdrawalOrder.Txhash
	rawData := withdrawalOrder.RawData
	signedTx := withdrawalOrder.SignedTx

	orderIDsMap := map[string]struct{}{}
	for _, orderID := range orderIDs {
		_, exist := orderIDsMap[orderID]
		if !exist {
			orderIDsMap[orderID] = struct{}{}
		} else {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("duplicated orderIDs:%v", strings.Join(orderIDs, ",")))
			return
		}

		order := keeper.ok.GetOrder(ctx, orderID)
		if order == nil {
			err = sdk.ErrNotFoundOrder(orderID)
			return
		}

		withdrawalOrder, valid := order.(*sdk.OrderWithdrawal)
		if !valid {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not a withdrawal order", order))
			return
		}

		if !orderStatus.Match(withdrawalOrder.Status) {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v status does not match expctedStatus:%v", order, orderStatus))
			return
		}
		if firstStatus != withdrawalOrder.Status {
			err = sdk.ErrInvalidTx(fmt.Sprintf("orderid:%v's status is %v, not as expected %v", orderID, order.GetOrderStatus(), firstStatus))
			return
		}
		if withdrawalOrder.WithdrawStatus != sdk.WithdrawStatusValid {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %s is not confirmed yet or is invalid", orderID))
			return
		}

		if orderStatus == sdk.OrderStatusCancel {
			err = sdk.ErrInvalidAddr(fmt.Sprintf("order %v is canceled", order))
		}

		if orderStatus == sdk.OrderStatusWaitSign || orderStatus == sdk.OrderStatusSignFinish {
			if len(withdrawalOrder.RawData) == 0 {
				err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v RawData is empty", order))
				return
			}
		}

		if orderStatus == sdk.OrderStatusSignFinish {
			if len(withdrawalOrder.SignedTx) == 0 || withdrawalOrder.Txhash == "" {
				err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v SignTx or ext tx hash is empty", order))
				return
			}
		}

		if withdrawalOrder.Symbol != symbol {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("symbol mismatch, need:%v, actual:%v", symbol, withdrawalOrder.Symbol))
			return
		}

		if withdrawalOrder.Txhash != hash {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("hash mismatch, need:%v, actual:%v", hash, withdrawalOrder.Txhash))
			return
		}

		if bytes.Compare(withdrawalOrder.RawData, rawData) != 0 {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("rawData mismatch, need:%v, actual:%v", rawData, withdrawalOrder.RawData))
			return
		}

		if bytes.Compare(withdrawalOrder.SignedTx, signedTx) != 0 {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("signTx mismatch, need:%v, actual:%v", signedTx, withdrawalOrder.SignedTx))
			return
		}

		userCU := keeper.ck.GetCU(ctx, withdrawalOrder.CUAddress)
		if userCU == nil {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v's CU does not exist", orderID))
			return
		}

		if userCU.GetCUType() != sdk.CUTypeUser {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v CU type is not user type", orderID))
			return
		}

		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, withdrawalOrder.GasFee))
		amtCoins := sdk.NewCoins(sdk.NewCoin(symbol, withdrawalOrder.Amount))
		need := amtCoins.Add(feeCoins)
		have := userCU.GetCoinsHold()
		if chain == symbol {
			if have.AmountOf(chain).LT(need.AmountOf(chain)) {
				err = sdk.ErrInsufficientCoins(fmt.Sprintf("need %v, %v have %v", need, userCU.GetAddress(), have))
				return
			}
		} else {
			if have.AmountOf(chain).LT(withdrawalOrder.GasFee) || have.AmountOf(symbol).LT(withdrawalOrder.Amount) {
				err = sdk.ErrInsufficientCoins(fmt.Sprintf("need %v, %v have %v", need, userCU.GetAddress(), have))
				return
			}
		}

		withdrawalOrders = append(withdrawalOrders, withdrawalOrder)
	}

	return
}

func (keeper BaseKeeper) checkDecodedUtxoTransaction(ctx sdk.Context, symbol string, opCUAddr sdk.CUAddress, withdrawalOrders []*sdk.OrderWithdrawal, tx *chainnode.ExtUtxoTransaction, fromAddr string) (sdk.Int, sdk.Error) {
	inAmt := sdk.ZeroInt()
	outAmt := sdk.ZeroInt()

	opCU := keeper.ck.GetCU(ctx, opCUAddr)

	for _, vin := range tx.Vins {
		if fromAddr != vin.Address {
			return sdk.ZeroInt(), sdk.ErrInvalidTx(fmt.Sprintf("Unexpected Vin address:%v, expected:%v", vin.Address, fromAddr))
		}

		item := keeper.ck.GetDeposit(ctx, symbol, opCUAddr, vin.Hash, vin.Index)
		if item == sdk.DepositNil {
			return sdk.ZeroInt(), sdk.ErrUnknownUtxo(vin.String())
		}

		inAmt = inAmt.Add(vin.Amount)
	}

	need := sdk.NewCoins(sdk.NewCoin(symbol, inAmt))
	owned := opCU.GetAssetCoins()
	if owned.AmountOf(symbol).LT(inAmt) {
		return sdk.ZeroInt(), sdk.ErrInsufficientCoins(fmt.Sprintf("opCU has insufficient coins, expected: %v, actual have:%v", need, owned))
	}

	for i := range withdrawalOrders {
		withdrawalOrder := withdrawalOrders[i]
		vout := tx.Vouts[i]

		if vout.Address != withdrawalOrder.WithdrawToAddress {
			return sdk.ZeroInt(), sdk.ErrInvalidTx(fmt.Sprintf("Unexpected Vout address:%v, expected:%v", vout.Address, withdrawalOrder.WithdrawToAddress))
		}

		if !vout.Amount.Equal(withdrawalOrder.Amount) {
			return sdk.ZeroInt(), sdk.ErrInvalidTx(fmt.Sprintf("Unexpected Vout Amount:%v, expected:%v", vout.Amount, withdrawalOrder.Amount))

		}
	}
	//support serveral changeback
	if len(tx.Vouts) > len(withdrawalOrders) {
		for i := len(withdrawalOrders); i < len(tx.Vouts); i++ {
			if tx.Vouts[i].Address != fromAddr {
				return sdk.ZeroInt(), sdk.ErrInvalidTx(fmt.Sprintf("Unexpected Changeback address:%v, expected:%v", tx.Vouts[i].Address, fromAddr))
			}
		}
	}

	for _, vout := range tx.Vouts {
		outAmt = outAmt.Add(vout.Amount)
	}

	calculatedFee := inAmt.Sub(outAmt)
	if !tx.CostFee.Equal(calculatedFee) {
		return sdk.ZeroInt(), sdk.ErrInvalidTx(fmt.Sprintf("Unexpected Gas:%v, expected:%v", calculatedFee, tx.CostFee))
	}

	return inAmt, nil
}

func (keeper BaseKeeper) checkWithdrawalOpCU(opCU exported.CustodianUnit, chain, symbol string, sendable bool, fromAddr string) (err sdk.Error) {
	if opCU == nil {
		err = sdk.ErrInvalidAccount("CU is nil")
		return
	}

	opCUAddr := opCU.GetAddress()
	if opCU.GetCUType() != sdk.CUTypeOp {
		err = sdk.ErrInvalidAccount(fmt.Sprintf("CU %v is not a OPCU", opCUAddr))
		return
	}
	if opCU.GetSymbol() != symbol {
		err = sdk.ErrInvalidAccount(fmt.Sprintf("symbol mismatch, expected:%v, got:%v", symbol, opCU.GetSymbol()))
		return
	}

	valid, canonicalFromAddr := keeper.cn.ValidAddress(chain, symbol, fromAddr)
	if !valid {
		err = sdk.ErrInvalidAddr(fmt.Sprintf("%v's address %v is not a valid address", opCUAddr, fromAddr))
		return
	}

	if canonicalFromAddr != fromAddr {
		err = sdk.ErrInvalidAddr(fmt.Sprintf("%v's address %v is not a canonical address", opCUAddr, fromAddr))
		return
	}

	sendStatus := opCU.IsEnabledSendTx(chain, fromAddr)
	if sendStatus != sendable {
		err = sdk.ErrInternal(fmt.Sprintf("lockStatus mismatch, expected %v, actual %v", sendable, sendStatus))
		return
	}

	return
}

func (keeper BaseKeeper) CancelWithdrawal(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string) sdk.Result {
	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount(fromCUAddr.String()).Result()
	}

	order := keeper.ok.GetOrder(ctx, orderID)
	if order == nil {
		return sdk.ErrInvalidOrder(fmt.Sprintf("Get WithDrawalOrder(%v) Err", orderID)).Result()
	}

	withdrawalOrder, valid := order.(*sdk.OrderWithdrawal)
	if !valid {
		return sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not withdrawal order", orderID)).Result()
	}

	if withdrawalOrder.CUAddress.String() != fromCUAddr.String() {
		return sdk.ErrInvalidAddr(fmt.Sprintf("cancel withdrawal order invalid addr(%v:%v)", fromCUAddr, withdrawalOrder.CUAddress)).Result()
	}

	if withdrawalOrder.Status != sdk.OrderStatusBegin {
		return sdk.ErrInvalidTx(fmt.Sprintf("cancel withdrawal order status not ok:%v", withdrawalOrder.Status)).Result()
	}

	if withdrawalOrder.WithdrawStatus != sdk.WithdrawStatusValid {
		return sdk.ErrInvalidTx("cancel withdrawal not confirmed").Result()
	}

	tokenInfo := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(withdrawalOrder.Symbol))
	chain := tokenInfo.Chain.String()

	feeCoins := sdk.NewCoins(sdk.NewCoin(chain, withdrawalOrder.GasFee))
	coins := sdk.NewCoins(sdk.NewCoin(withdrawalOrder.Symbol, withdrawalOrder.Amount))
	need := coins.Add(feeCoins)

	have := fromCU.GetCoinsHold()
	if chain == withdrawalOrder.Symbol {
		if have.AmountOf(chain).LT(need.AmountOf(chain)) {
			return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, actual have:%v", need, have)).Result()
		}
	} else {
		if have.AmountOf(chain).LT(withdrawalOrder.GasFee) || have.AmountOf(withdrawalOrder.Symbol).LT(withdrawalOrder.Amount) {
			return sdk.ErrInsufficientCoins(fmt.Sprintf("need:%v, actual have:%v", need, have)).Result()
		}
	}

	fromCU.ResetBalanceFlows()
	fromCU.SubCoinsHold(need)
	fromCU.AddCoins(need)
	keeper.ck.SetCU(ctx, fromCU)

	withdrawalOrder.SetOrderStatus(sdk.OrderStatusCancel)
	keeper.ok.SetOrder(ctx, withdrawalOrder)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(withdrawalOrder.Symbol), withdrawalOrder.GetCUAddress(), withdrawalOrder.GetID(), sdk.OrderTypeWithdrawal, sdk.OrderStatusCancel))

	for _, balanceFlow := range fromCU.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeWithdrawal, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}
