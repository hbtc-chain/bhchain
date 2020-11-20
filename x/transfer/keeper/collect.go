package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

var (
	PriceUpLimitRatio  = sdk.NewDecWithPrec(12, 1) //gas price uplimit 1.2x
	PriceLowLimitRatio = sdk.NewDecWithPrec(8, 1)  //gas price lowlimit 0.8x
)

func (keeper BaseKeeper) CollectWaitSign(ctx sdk.Context, toCUAddr sdk.CUAddress, orderIDs []string, rawData []byte) sdk.Result {
	toCUAst := keeper.ik.GetCUIBCAsset(ctx, toCUAddr)
	if toCUAst == nil {
		return sdk.ErrInvalidAccount(toCUAddr.String()).Result()
	}
	if toCUAst.GetCUType() != sdk.CUTypeOp {
		return sdk.ErrInvalidTx(fmt.Sprintf("collect to a non OP Cu :%v", toCUAddr)).Result()
	}

	if toCUAst.GetMigrationStatus() != sdk.MigrationFinish {
		return sdk.ErrInvalidTx(fmt.Sprintf("To OPCU %v is in migration", toCUAddr)).Result()
	}

	//basic  check and retrieve information
	totalCoins, tokenInfo, vins, collectOrders, depositItems, err := keeper.checkCollectOrders(ctx, orderIDs, sdk.OrderStatusBegin, sdk.DepositItemStatusWaitCollect)
	if err != nil {
		return err.Result()
	}

	toCU := keeper.ck.GetCU(ctx, toCUAddr)
	symbol := collectOrders[0].Symbol
	chain := tokenInfo.Chain.String()
	if toCU.GetSymbol() != symbol {
		return sdk.ErrInvalidTx(fmt.Sprintf("OP CU %v does not support symbol:%v", toCUAddr, symbol)).Result()
	}
	curEpoch := keeper.sk.GetCurrentEpoch(ctx)

	//Retrieve gas Price
	gasPrice := tokenInfo.GasPrice
	if chain != symbol {
		ti := keeper.tk.GetIBCToken(ctx, sdk.Symbol(chain))
		if ti == nil {
			return sdk.ErrInvalidSymbol(fmt.Sprintf("%s does not exist", chain)).Result()
		}
		gasPrice = ti.GasPrice
	}

	toAddr := toCUAst.GetAssetAddress(symbol, curEpoch.Index)
	if toAddr == "" {
		return sdk.ErrInvalidTx(fmt.Sprintf("OP CU %v does not have %v's address", toCUAddr, symbol)).Result()
	}

	valid, _ := keeper.cn.ValidAddress(chain, symbol, toAddr)
	if !valid {
		return sdk.ErrInvalidAddr(fmt.Sprintf("%v is not a valid address for %v", toAddr, symbol)).Result()
	}

	priceUpLimit := sdk.NewDecFromInt(gasPrice).Mul(PriceUpLimitRatio)
	priceLowLimit := sdk.NewDecFromInt(gasPrice).Mul(PriceLowLimitRatio)

	var inAmt, outAmt, gasLimit sdk.Int
	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		tx, _, err := keeper.cn.QueryUtxoTransactionFromData(chain, symbol, rawData, vins)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction:%v,err:%v", rawData, err)).Result()
		}

		if tx.CostFee.IsZero() {
			return sdk.ErrInsufficientFee(fmt.Sprintf("QueryUtxoTransactionFromData, costFee =0, tx:%v, rawData:%v", tx, hex.EncodeToString(rawData))).Result()
		}

		inAmt = sdk.ZeroInt()
		for i, in := range tx.Vins {
			if !in.Equal(*vins[i]) {
				return sdk.ErrInvalidTx(fmt.Sprintf("expected Vin %v, actual %v", vins[i].String(), in.String())).Result()
			}
			inAmt = inAmt.Add(in.Amount)
		}

		if len(tx.Vins) > sdk.MaxVinNum {
			return sdk.ErrInvalidTx(fmt.Sprintf("contains too many vin %v", len(tx.Vins))).Result()
		}
		if inAmt.LT(tokenInfo.CollectThreshold) {
			return sdk.ErrInvalidTx(fmt.Sprintf("collect amount %s less than threshold", inAmt.String())).Result()
		}

		outAmt = sdk.ZeroInt()
		for _, out := range tx.Vouts {
			if out.Address != toAddr {
				return sdk.ErrInvalidTx(fmt.Sprintf("collect to an unexpect address %v, expected address:%v ", out.Address, toAddr)).Result()
			}
			outAmt = outAmt.Add(out.Amount)
		}
		//tx.CostFee must be not Nil in utxo
		if !inAmt.Equal(outAmt.Add(tx.CostFee)) {
			return sdk.ErrInvalidTx(fmt.Sprintf("inAmt:%v != outAmt:%v + fee:%v", inAmt, outAmt, tx.CostFee)).Result()
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

		for i := range collectOrders {
			userCUAst := keeper.ik.GetCUIBCAsset(ctx, collectOrders[i].CollectFromCU)
			coins := sdk.NewCoins(sdk.NewCoin(collectOrders[i].Symbol, collectOrders[i].Amount))
			userCUAst.SubAssetCoins(coins)
			userCUAst.AddAssetCoinsHold(coins)
			keeper.ik.SetCUIBCAsset(ctx, userCUAst)
		}

	case sdk.AccountBased:
		fromCUAst := keeper.ik.GetCUIBCAsset(ctx, collectOrders[0].CollectFromCU)
		sendable := fromCUAst.IsEnabledSendTx(chain, collectOrders[0].CollectFromAddress)
		//support only one collect at one time
		if !sendable {
			return sdk.ErrInternal(fmt.Sprintf("%v %v sendable is false", fromCUAst.GetAddress(), chain)).Result()
		}

		tx, _, err := keeper.cn.QueryAccountTransactionFromData(chain, symbol, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction:%v,err:%v", rawData, err)).Result()
		}

		if tx.To != toAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("collet to an unexpected address:%v, expected address:%v", tx.To, toAddr)).Result()
		}

		validContractAddr := ""
		if tokenInfo.Issuer != "" {
			_, validContractAddr = keeper.cn.ValidAddress(tokenInfo.Chain.String(), tokenInfo.Symbol.String(), tokenInfo.Issuer)
		}
		if tx.ContractAddress != validContractAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected withdrawal contract address:%v, expected:%v", tx.ContractAddress, tokenInfo.Issuer)).Result()
		}

		if !tokenInfo.GasLimit.Equal(tx.GasLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas limit %v != expected %v", tx.GasLimit, tokenInfo.GasLimit)).Result()
		}

		nonce := fromCUAst.GetNonce(tokenInfo.Chain.String(), collectOrders[0].CollectFromAddress)
		if nonce != tx.Nonce {
			return sdk.ErrInvalidTx(fmt.Sprintf("cu(%v) tx nonce not equal, cunoce :%v, rawdata:%v", fromCUAst.GetAddress(), nonce, tx.Nonce)).Result()
		}

		fee := tx.GasPrice.Mul(tx.GasLimit)
		coins := sdk.NewCoins(sdk.NewCoin(symbol, tx.Amount))
		gasRemained := fromCUAst.GetGasRemained(chain, collectOrders[0].CollectFromAddress)
		//for erc20, check gr-gu, if gr-gu < estimated gas, error
		if chain != symbol {
			if !totalCoins.IsEqual(coins) {
				return sdk.ErrInvalidTx(fmt.Sprintf("total amount:%v != outAmt:%v ", totalCoins, coins)).Result()
			}
			if gasRemained.LT(fee) {
				return sdk.ErrInsufficientFee(fmt.Sprintf("need:%v, actual:%v", fee, gasRemained)).Result()
			}
			if tx.Amount.LT(tokenInfo.CollectThreshold) {
				return sdk.ErrInvalidTx(fmt.Sprintf("collect amount %s less than threshold", tx.Amount.String())).Result()
			}
		} else {
			needAmount := fee.Add(tx.Amount)
			have := gasRemained.Add(totalCoins.AmountOf(chain))
			//consider remained gas
			if have.LT(needAmount) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("token %s is insufficient, need:%v, actual:%v", chain, needAmount, have)).Result()
			}
			if needAmount.LT(tokenInfo.CollectThreshold) {
				return sdk.ErrInvalidTx(fmt.Sprintf("collect amount %s less than threshold", needAmount.String())).Result()
			}
		}

		if sdk.NewDecFromInt(tx.GasPrice).GT(priceUpLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too high, actual:%v, uplimit:%v", tx.GasPrice, priceUpLimit)).Result()
		}

		if sdk.NewDecFromInt(tx.GasPrice).LT(priceLowLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too low, actual:%v, lowlimit:%v", tx.GasPrice, priceLowLimit)).Result()
		}

		gasPrice = tx.GasPrice
		gasLimit = tx.GasLimit

		fromCUAst.SubAssetCoins(totalCoins)
		fromCUAst.AddAssetCoinsHold(totalCoins)
		if tokenInfo.IsNonceBased {
			//Lock user CU
			fromCUAst.SetEnableSendTx(false, chain, collectOrders[0].CollectFromAddress)
		}
		keeper.ik.SetCUIBCAsset(ctx, fromCUAst)

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update collectOrder status and deposit item status
	for i := 0; i < len(orderIDs); i++ {
		collectOrders[i].SetOrderStatus(sdk.OrderStatusWaitSign)
		collectOrders[i].CollectToCU = toCUAddr
		collectOrders[i].GasLimit = gasLimit
		collectOrders[i].GasPrice = gasPrice
		collectOrders[i].RawData = make([]byte, len(rawData))
		copy(collectOrders[i].RawData, rawData)
		keeper.ok.SetOrder(ctx, collectOrders[i])
		_ = keeper.ik.SetDepositStatus(ctx, symbol, collectOrders[i].CUAddress, depositItems[i].Hash, depositItems[i].Index, sdk.DepositItemStatusInProcess)
	}

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), collectOrders[0].CollectFromCU, collectOrders[0].GetID(), sdk.OrderTypeCollect, sdk.OrderStatusWaitSign))
	flows = append(flows, keeper.rk.NewCollectWaitSignFlow(orderIDs, rawData))
	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeCollect, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) CollectSignFinish(ctx sdk.Context, orderIDs []string, signedTx []byte) sdk.Result {
	//basic  check
	_, tokenInfo, vins, collectOrders, _, err := keeper.checkCollectOrders(ctx, orderIDs, sdk.OrderStatusWaitSign, sdk.DepositItemStatusInProcess)
	if err != nil {
		return err.Result()
	}
	symbol := collectOrders[0].Symbol
	chain := tokenInfo.Chain.String()
	rawData := collectOrders[0].RawData

	var txHash string
	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		result, hash := keeper.verifyUtxoBasedSignedTx(ctx, vins, nil, chain, symbol, rawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}
		txHash = hash
	case sdk.AccountBased:
		result, hash := keeper.verifyAccountBasedSignedTx(collectOrders[0].CollectFromAddress, chain, symbol, collectOrders[0].RawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}
		txHash = hash
	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update collectOrder status and coins
	for i := 0; i < len(orderIDs); i++ {
		collectOrders[i].SetOrderStatus(sdk.OrderStatusSignFinish)
		collectOrders[i].ExtTxHash = txHash
		collectOrders[i].SignedTx = make([]byte, len(signedTx))
		copy(collectOrders[i].SignedTx, signedTx)
		keeper.ok.SetOrder(ctx, collectOrders[i])
	}

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), collectOrders[0].CollectFromCU, collectOrders[0].GetID(), sdk.OrderTypeCollect, sdk.OrderStatusSignFinish))
	flows = append(flows, keeper.rk.NewCollectSignFinishFlow(orderIDs, signedTx))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeCollect, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) CollectFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, costFee sdk.Int) sdk.Result {
	//basic  check
	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("collect from not a validator :%v", fromCUAddr)).Result()
	}

	totalCoins, tokenInfo, vins, collectOrders, _, err := keeper.checkCollectOrders(ctx, orderIDs, sdk.OrderStatusSignFinish, sdk.DepositItemStatusInProcess, sdk.DepositItemStatusConfirmed)
	if err != nil {
		return err.Result()
	}

	confirmedFirstTime, _, _ := keeper.evidenceKeeper.Vote(ctx, collectOrders[0].ExtTxHash, fromCUAddr, types.NewTxVote(costFee.Int64(), true), uint64(ctx.BlockHeight()))
	result := sdk.Result{}
	if !confirmedFirstTime {
		return result
	}

	symbol := collectOrders[0].Symbol
	chain := tokenInfo.Chain.String()
	toCUAddr := collectOrders[0].CollectToCU
	signedTx := collectOrders[0].SignedTx
	outAmt := sdk.ZeroInt()

	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		localTx, err := keeper.cn.QueryUtxoTransactionFromSignedData(chain, symbol, signedTx, vins)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction from signed transaction:%v", signedTx)).Result()
		}

		if !localTx.CostFee.Equal(costFee) {
			return sdk.ErrInvalidTx(fmt.Sprintf("unmatch costFee, expected:%v, actua:%v", localTx.CostFee, costFee)).Result()
		}

		//update fromCU's asset, collectOrders' status and deposit status
		for i := range collectOrders {
			userCUAst := keeper.ik.GetCUIBCAsset(ctx, collectOrders[i].CollectFromCU)
			coins := sdk.NewCoins(sdk.NewCoin(collectOrders[i].Symbol, collectOrders[i].Amount))
			userCUAst.SubAssetCoinsHold(coins)

			//record the gr/gu in userCU[0]
			if i == 0 {
				gr := sdk.NewCoins(sdk.NewCoin(symbol, costFee))
				userCUAst.AddGasReceived(gr)
				userCUAst.AddGasUsed(gr)
			}

			keeper.ik.SetCUIBCAsset(ctx, userCUAst)
			_ = keeper.ik.SetDepositStatus(ctx, collectOrders[i].Symbol, collectOrders[i].CollectFromCU, collectOrders[i].Txhash, collectOrders[i].Index, sdk.DepositItemStatusConfirmed)
			collectOrders[i].Status = sdk.OrderStatusFinish
			keeper.ok.SetOrder(ctx, collectOrders[i])
		}

		for i, vout := range localTx.Vouts {
			if !vout.Equal(*localTx.Vouts[i]) {
				return sdk.ErrInvalidTx(fmt.Sprintf("vout mismatch, expected:%v, actual:%v", localTx.Vins[i].String(), vout.String())).Result()
			}
			outAmt = outAmt.Add(vout.Amount)

			item, err := sdk.NewDepositItem(localTx.Hash, uint64(i), vout.Amount, vout.Address, "", sdk.DepositItemStatusConfirmed)
			if err != nil {
				return sdk.ErrInvalidTx(fmt.Sprintf("fail to create deposit item, %v %v %v", localTx.Hash, i, vout.Amount)).Result()
			}
			_ = keeper.ik.SaveDeposit(ctx, symbol, toCUAddr, item)
		}

		//update collectToCU's asset
		opCUAst := keeper.ik.GetCUIBCAsset(ctx, toCUAddr)
		coins := sdk.NewCoins(sdk.NewCoin(symbol, outAmt))
		opCUAst.AddAssetCoins(coins)
		keeper.ik.SetCUIBCAsset(ctx, opCUAst)

	case sdk.AccountBased:
		localTx, err := keeper.cn.QueryAccountTransactionFromSignedData(chain, symbol, signedTx)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction:%v, err:%v", signedTx, err)).Result()
		}

		userCUAst := keeper.ik.GetCUIBCAsset(ctx, collectOrders[0].CollectFromCU)
		//coins := sdk.NewCoins(sdk.NewCoin(symbol, totalCoins))
		userCUAst.SubAssetCoinsHold(totalCoins)
		//update gr if necessary
		if chain == symbol {
			gr := totalCoins.Sub(sdk.NewCoins(sdk.NewCoin(symbol, localTx.Amount)))
			userCUAst.AddGasReceived(gr)
			userCUAst.AddGasRemained(symbol, collectOrders[0].CollectFromAddress, gr.AmountOf(symbol))
		}
		gu := sdk.NewCoins(sdk.NewCoin(chain, costFee)) //update gasused
		userCUAst.AddGasUsed(gu)
		userCUAst.SubGasRemained(chain, collectOrders[0].CollectFromAddress, costFee)

		if tokenInfo.IsNonceBased {
			//don't update local nonce for trx
			nonce := userCUAst.GetNonce(tokenInfo.Chain.String(), collectOrders[0].CollectFromAddress) + 1
			userCUAst.SetNonce(tokenInfo.Chain.String(), nonce, collectOrders[0].CollectFromAddress)
			userCUAst.SetEnableSendTx(true, chain, collectOrders[0].CollectFromAddress)
		}
		keeper.ik.SetCUIBCAsset(ctx, userCUAst)

		for i := range collectOrders {
			_ = keeper.ik.SetDepositStatus(ctx, collectOrders[i].Symbol, collectOrders[i].CollectFromCU, collectOrders[i].Txhash, collectOrders[i].Index, sdk.DepositItemStatusConfirmed)
			collectOrders[i].Status = sdk.OrderStatusFinish
			keeper.ok.SetOrder(ctx, collectOrders[i])
		}

		//update collectToCU's asset
		opCUAst := keeper.ik.GetCUIBCAsset(ctx, toCUAddr)
		coins := sdk.NewCoins(sdk.NewCoin(symbol, localTx.Amount))
		opCUAst.AddAssetCoins(coins)
		keeper.ik.SetCUIBCAsset(ctx, opCUAst)

		//add deposititem into opcu, status is collected
		item, err := sdk.NewDepositItem(localTx.Hash, uint64(0), localTx.Amount, localTx.To, "", sdk.DepositItemStatusConfirmed)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("fail to create deposit item, %v %v %v", localTx.Hash, 0, localTx.Amount)).Result()
		}
		_ = keeper.ik.SaveDeposit(ctx, symbol, toCUAddr, item)

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidOrder("Not support AccountSharedBased temporary").Result()
	}

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), collectOrders[0].CollectFromCU, collectOrders[0].GetID(), sdk.OrderTypeCollect, sdk.OrderStatusFinish))
	flows = append(flows, keeper.rk.NewCollectFinishFlow(orderIDs, costFee))
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeCollect, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) checkCollectOrders(ctx sdk.Context, orderIDs []string, orderStatus sdk.OrderStatus, depositStatuses ...sdk.DepositItemStatus) (coins sdk.Coins, tokenInfo *sdk.IBCToken,
	vins []*sdk.UtxoIn, collectOrders []*sdk.OrderCollect, depositItems []*sdk.DepositItem, err sdk.Error) {
	order := keeper.ok.GetOrder(ctx, orderIDs[0])
	if order == nil {
		err = sdk.ErrNotFoundOrder(fmt.Sprintf("orderid:%v does not exist", orderIDs[0]))
		return
	}

	collectOrder, valid := order.(*sdk.OrderCollect)
	if !valid {
		err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not collect order", orderIDs[0]))
		return
	}

	symbol := order.GetSymbol()

	tokenInfo = keeper.tk.GetIBCToken(ctx, sdk.Symbol(symbol))
	if tokenInfo == nil {
		err = sdk.ErrUnSupportToken(symbol)
		return
	}
	if !tokenInfo.WithdrawalEnabled || !tokenInfo.SendEnabled || !keeper.IsSendEnabled(ctx) {
		err = sdk.ErrTransactionIsNotEnabled(fmt.Sprintf("%v's collect is not enabled temporary", symbol))
		return
	}

	var checkCUAddr sdk.CUAddress
	var fromExtAddr string
	rawData := collectOrder.RawData
	signedTx := collectOrder.SignedTx
	extTxHash := collectOrder.ExtTxHash

	for i, orderID := range orderIDs {
		order := keeper.ok.GetOrder(ctx, orderID)
		if order == nil {
			err = sdk.ErrNotFoundOrder(orderID)
			return
		}
		collectOrder, valid := order.(*sdk.OrderCollect)
		if !valid {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not a Collect Order", order))
			return
		}

		if collectOrder.DepositStatus != sdk.DepositConfirmed {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order deposit status %v != expctedStatus", collectOrder.DepositStatus))
			return
		}

		if orderStatus != collectOrder.Status && orderStatus != sdk.OrderStatusSignFinish {
			err = sdk.ErrInvalidTx(fmt.Sprintf("orderid:%v's status is %v, not as expected %v", orderID, order.GetOrderStatus(), orderStatus))
			return
		}

		if orderStatus == sdk.OrderStatusWaitSign || orderStatus == sdk.OrderStatusSignFinish {
			if len(collectOrder.RawData) == 0 {
				err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v RawData is empty", order))
				return
			}
		}

		if orderStatus == sdk.OrderStatusSignFinish {
			if len(collectOrder.SignedTx) == 0 || collectOrder.ExtTxHash == "" {
				err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v SignTx or ext tx hash is empty", order))
				return
			}
		}

		if collectOrder.Symbol != symbol {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("symbol mismatch, need:%v, actual:%v", symbol, collectOrder.Symbol))
			return
		}

		if bytes.Compare(collectOrder.RawData, rawData) != 0 {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("rawData mismatch, need:%v, actual:%v", rawData, collectOrder.RawData))
			return
		}

		if bytes.Compare(collectOrder.SignedTx, signedTx) != 0 {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("signedTx mismatch, need:%v, actual:%v", signedTx, collectOrder.SignedTx))
			return
		}

		if collectOrder.ExtTxHash != extTxHash {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("extTxHash mismatch, need:%v, actual:%v", extTxHash, collectOrder.ExtTxHash))
			return
		}

		userCU := keeper.ck.GetCU(ctx, collectOrder.CUAddress)
		if userCU == nil {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v's CU does not exist", order))
			return
		}

		if userCU.GetCUType() != sdk.CUTypeUser {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v CU type is not user type", order))
			return
		}
		depositItem := keeper.ik.GetDeposit(ctx, symbol, collectOrder.CUAddress, collectOrder.Txhash, collectOrder.Index)
		if depositItem == sdk.DepositNil {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v's deposit item %v %v does not exist", order, collectOrder.Txhash, collectOrder.Index))
			return
		}

		if !depositItem.GetStatus().In(depositStatuses) {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %s's deposit item %v %v status %v is not in expected statuses:%v ", order.GetID(), collectOrder.Txhash, collectOrder.Index, depositItem.GetStatus(), depositStatuses))
			return
		}

		switch tokenInfo.TokenType {
		case sdk.AccountBased:
			if i == 0 {
				checkCUAddr = collectOrder.CUAddress //record checkCUAddr for consistence check
				fromExtAddr = collectOrder.CollectFromAddress
			} else {
				if !checkCUAddr.Equals(collectOrder.CUAddress) {
					err = sdk.ErrInvalidOrder(fmt.Sprintf("Different CU in one collect order for AccountBased token: %v, %v", checkCUAddr, collectOrder.CUAddress))
					return
				}
				if fromExtAddr != collectOrder.CollectFromAddress {
					err = sdk.ErrInvalidOrder(fmt.Sprintf("Different FromAddr in one collect order for AccountBased token: %v, %v", fromExtAddr, collectOrder.CollectFromAddress))
					return
				}
			}

		case sdk.UtxoBased:
			utxo := sdk.NewUtxoIn(depositItem.Hash, depositItem.Index, depositItem.Amount, depositItem.ExtAddress)
			vins = append(vins, &utxo)
		}

		coins = coins.Add(sdk.NewCoins(sdk.NewCoin(symbol, depositItem.Amount)))
		collectOrders = append(collectOrders, collectOrder)
		depositItems = append(depositItems, &depositItem)
	}
	return
}
