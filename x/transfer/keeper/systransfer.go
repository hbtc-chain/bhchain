package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func (keeper BaseKeeper) SysTransfer(ctx sdk.Context, fromCUAddr, toCUAddr sdk.CUAddress, toAddr, orderID, symbol string) sdk.Result {
	if keeper.ok.IsExist(ctx, orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("order %v already exists", orderID)).Result()
	}

	fromCUAst := keeper.ik.GetCUIBCAsset(ctx, fromCUAddr)
	if fromCUAst == nil {
		return sdk.ErrInvalidAccount(fromCUAddr.String()).Result()
	}
	if fromCUAst.GetCUType() != sdk.CUTypeOp {
		return sdk.ErrInvalidTx(fmt.Sprintf("systransfer from a non OP CU :%v", fromCUAddr.String())).Result()
	}
	if fromCUAst.GetMigrationStatus() != sdk.MigrationFinish {
		return sdk.ErrInvalidTx(fmt.Sprintf("systransfer from a migrating OP CU :%v", fromCUAddr.String())).Result()
	}

	tokenInfo := keeper.tk.GetIBCToken(ctx, sdk.Symbol(symbol))
	if tokenInfo == nil {
		return sdk.ErrUnSupportToken(symbol).Result()
	}
	chain := tokenInfo.Chain.String()
	if symbol == chain {
		return sdk.ErrInvalidTx(fmt.Sprintf("Not support systansfer chain's mainnet token")).Result()
	}

	toCUAst := keeper.ik.GetCUIBCAsset(ctx, toCUAddr)
	if toCUAst == nil {
		return sdk.ErrInvalidAccount(toCUAddr.String()).Result()
	}
	valid, canonicalToAddr := keeper.cn.ValidAddress(chain, symbol, toAddr)
	if !valid {
		return sdk.ErrInvalidAddr(fmt.Sprintf("%v is not a valid address", toAddr)).Result()
	}
	toCUAsset := toCUAst.GetAssetByAddr(symbol, canonicalToAddr)
	if toCUAsset == sdk.NilAsset {
		return sdk.ErrInvalidTx(fmt.Sprintf("%v does not belong to cu %v", toAddr, toCUAst.GetAddress().String())).Result()
	}
	if toCUAst.GetCUType() == sdk.CUTypeOp && toCUAst.GetMigrationStatus() == sdk.MigrationFinish {
		if toCUAsset.Epoch != keeper.sk.GetCurrentEpoch(ctx).Index {
			return sdk.ErrInvalidTx("Cannot sys transfer to last epoch addr").Result()
		}
	}

	if keeper.hasProcessingSysTransfer(ctx, toCUAddr, chain, canonicalToAddr) {
		return sdk.ErrInvalidTx(fmt.Sprintf("To OPCU %v has processing sys transfer of %s", toCUAddr.String(), chain)).Result()
	}

	//symbol check
	if !tokenInfo.WithdrawalEnabled || !tokenInfo.SendEnabled || !keeper.IsSendEnabled(ctx) {
		return sdk.ErrTransactionIsNotEnabled(fmt.Sprintf("%v's systransfer is not enabled temporary", symbol)).Result()
	}

	//chain check
	curEpoch := keeper.sk.GetCurrentEpoch(ctx)
	fromAddr := fromCUAst.GetAssetAddress(chain, curEpoch.Index)
	sendable := fromCUAst.IsEnabledSendTx(chain, fromAddr)
	if !sendable {
		return sdk.ErrInternal(fmt.Sprintf("%v %v is not sendable", fromCUAddr, chain)).Result()
	}

	chainTokenInfo := keeper.tk.GetIBCToken(ctx, sdk.Symbol(chain))
	if chainTokenInfo == nil {
		return sdk.ErrUnSupportToken(symbol).Result()
	}
	if !chainTokenInfo.WithdrawalEnabled || !chainTokenInfo.SendEnabled {
		return sdk.ErrTransactionIsNotEnabled(fmt.Sprintf("%v's systransfer is not enabled temporary", chain)).Result()
	}

	chainPrice := chainTokenInfo.GasPrice
	gasFee := chainPrice.Mul(tokenInfo.GasLimit)
	if !keeper.checkNeedSysTransfer(ctx, chain, canonicalToAddr, gasFee, toCUAst.GetCUType(), toCUAddr) {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s does not need systransfer", toCUAst.GetAddress())).Result()
	}

	var balanceFlows []sdk.Flow
	if toCUAst.GetCUType() == sdk.CUTypeUser {
		dlt := keeper.ik.GetDepositList(ctx, symbol, toCUAddr)
		waitCollectNum := sdk.ZeroInt()
		waitCollectItems := []sdk.DepositItem{}
		for _, item := range dlt {
			if item.Status == sdk.DepositItemStatusUnCollected {
				waitCollectNum = waitCollectNum.Add(item.Amount)
				waitCollectItems = append(waitCollectItems, item)
			}
		}

		if waitCollectNum.GT(sdk.ZeroInt()) {
			_, flow, err := keeper.AddCoin(ctx, toCUAddr, sdk.NewCoin(symbol, waitCollectNum))
			if err != nil {
				return err.Result()
			}
			balanceFlows = append(balanceFlows, flow)
			_, flow, err = keeper.SubCoin(ctx, toCUAddr, tokenInfo.CollectFee())
			if err != nil {
				return err.Result()
			}
			balanceFlows = append(balanceFlows, flow)

			waitCollectOrderIDs := keeper.getWaitCollectOrderIDs(ctx, toCUAddr.String(), symbol)
			if len(waitCollectItems) <= 0 {
				return sdk.ErrInternal(fmt.Sprintf("systransfer cu %v %v not have waitcollect order", toCUAddr.String(), symbol)).Result()
			}

			order := keeper.ok.GetOrder(ctx, waitCollectOrderIDs[0])
			collectOrder := order.(*sdk.OrderCollect)
			collectOrder.CostFee = tokenInfo.CollectFee().Amount
			keeper.ok.SetOrder(ctx, order)

			for _, item := range waitCollectItems {
				_ = keeper.ik.SetDepositStatus(ctx, symbol, toCUAddr, item.Hash, item.Index, sdk.DepositItemStatusWaitCollect)
			}
		}
	}

	var amount sdk.Int
	if toCUAst.GetCUType() == sdk.CUTypeOp {
		if toCUAst.GetMigrationStatus() == sdk.MigrationFinish {
			amount = tokenInfo.OpCUSysTransferAmount()
		} else {
			amount = tokenInfo.GasPrice.Mul(tokenInfo.GasLimit)
		}
	} else {
		amount = tokenInfo.SysTransferAmount()
	}

	//move (chain, amount) to assethold
	need := sdk.NewCoins(sdk.NewCoin(chain, amount))
	have := fromCUAst.GetAssetCoins()

	if have.AmountOf(chain).LT(amount) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("actual have %v, need %v", have, need)).Result()
	}

	sysTransferOrder := keeper.ok.NewOrderSysTransfer(ctx, fromCUAddr, orderID, symbol, amount, sdk.ZeroInt(), toCUAddr.String(), canonicalToAddr, fromCUAddr.String(), fromAddr)
	if sysTransferOrder == nil {
		return sdk.ErrInvalidOrder(fmt.Sprintf("Fail to create order:%v", orderID)).Result()
	}
	keeper.ok.SetOrder(ctx, sysTransferOrder)

	if tokenInfo.IsNonceBased {
		fromCUAst.SetEnableSendTx(false, chain, fromAddr)
	}
	fromCUAst.SubAssetCoins(need)
	fromCUAst.AddAssetCoinsHold(need)
	keeper.ik.SetCUIBCAsset(ctx, fromCUAst)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), sysTransferOrder.GetCUAddress(), sysTransferOrder.GetID(), sdk.OrderTypeSysTransfer, sdk.OrderStatusBegin))
	flows = append(flows, keeper.rk.NewSysTransferFlow(orderID, fromCUAddr.String(), toCUAddr.String(), fromAddr, canonicalToAddr, symbol, amount))
	flows = append(flows, balanceFlows...)
	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeSysTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSysTransfer,
			sdk.NewAttribute(types.AttributeKeySender, fromAddr),
			sdk.NewAttribute(types.AttributeKeyRecipient, toAddr),
			sdk.NewAttribute(types.AttributeKeySymbol, symbol),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	})

	return result
}

func (keeper BaseKeeper) SysTransferWaitSign(ctx sdk.Context, orderID string, signHash []byte, rawData []byte) sdk.Result {
	order := keeper.ok.GetOrder(ctx, orderID)
	if order == nil {
		return sdk.ErrNotFoundOrder(orderID).Result()
	}
	sysTransferOrder, valid := order.(*sdk.OrderSysTransfer)
	if !valid {
		return sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not systransfer order", orderID)).Result()
	}

	fromCUAddr := order.GetCUAddress()
	fromCUAst := keeper.ik.GetCUIBCAsset(ctx, fromCUAddr)

	tokenInfo, err := keeper.checkSysTransferOrder(ctx, sysTransferOrder, sdk.OrderStatusBegin)
	if err != nil {
		return err.Result()
	}

	symbol := order.GetSymbol()
	chain := tokenInfo.Chain.String()

	chainTokenInfo := keeper.tk.GetIBCToken(ctx, tokenInfo.Chain)
	priceUpLimit := sdk.NewDecFromInt(chainTokenInfo.GasPrice).Mul(PriceUpLimitRatio)
	priceLowLimit := sdk.NewDecFromInt(chainTokenInfo.GasPrice).Mul(PriceLowLimitRatio)

	gasFee := sdk.ZeroInt()
	var coins sdk.Coins
	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		return sdk.ErrInvalidTx("Not support UtxoBased systransfer temporary").Result()

	case sdk.AccountBased:
		tx, hash, err := keeper.cn.QueryAccountTransactionFromData(chain, chain, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		if tx.To != sysTransferOrder.ToAddress {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected systransfer to address:%v, expected:%v", tx.To, sysTransferOrder.ToAddress)).Result()
		}

		if !tx.Amount.Equal(sysTransferOrder.Amount) {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected systransfer Amount:%v, expected:%v", tx.Amount, sysTransferOrder.Amount)).Result()
		}

		validContractAddr := ""
		if chainTokenInfo.Issuer != "" {
			_, validContractAddr = keeper.cn.ValidAddress(chainTokenInfo.Chain.String(), chainTokenInfo.Symbol.String(), chainTokenInfo.Issuer)
		}
		if tx.ContractAddress != validContractAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected withdrawal contract address:%v, expected:%v", tx.ContractAddress, chainTokenInfo.Issuer)).Result()
		}

		if !bytes.Equal(hash, signHash) {
			return sdk.ErrInvalidTx(fmt.Sprintf("hash mismatch, expected:%v, have:%v", string(hash), signHash)).Result()
		}

		if !chainTokenInfo.GasLimit.Equal(tx.GasLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas limit mismatch, expected:%v, have:%v", tokenInfo.GasLimit, tx.GasLimit)).Result()
		}

		if sdk.NewDecFromInt(tx.GasPrice).GT(priceUpLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too high, actual:%v, uplimit:%v", tx.GasPrice, priceUpLimit)).Result()
		}

		if sdk.NewDecFromInt(tx.GasPrice).LT(priceLowLimit) {
			return sdk.ErrInvalidTx(fmt.Sprintf("gas price is too low, actual:%v, lowlimit:%v", tx.GasPrice, priceLowLimit)).Result()
		}

		nonce := fromCUAst.GetNonce(chain, sysTransferOrder.FromAddress)
		if nonce != tx.Nonce {
			return sdk.ErrInvalidTx(fmt.Sprintf("tx nonce not equal, cu :%v, rawdata:%v", nonce, tx.Nonce)).Result()
		}

		gasFee = tx.GasPrice.Mul(tx.GasLimit)
		coins = sdk.NewCoins(sdk.NewCoin(chain, gasFee))

		have := fromCUAst.GetAssetCoins()
		if have.AmountOf(chain).LT(gasFee) {
			return sdk.ErrInsufficientCoins(fmt.Sprintf("actual have %v, need %v", have, coins)).Result()
		}

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//move feecoins to onhold
	fromCUAst.SubAssetCoins(coins)
	fromCUAst.AddAssetCoinsHold(coins)
	keeper.ik.SetCUIBCAsset(ctx, fromCUAst)

	sysTransferOrder.CostFee = gasFee //record the gasfee for future use, OrderSysTranfer has no GasPrice and GasLimit, use CostFee record
	sysTransferOrder.Status = sdk.OrderStatusWaitSign
	sysTransferOrder.RawData = rawData
	keeper.ok.SetOrder(ctx, sysTransferOrder)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), orderID, sdk.OrderTypeSysTransfer, sdk.OrderStatusWaitSign))
	flows = append(flows, keeper.rk.NewSysTransferWaitSignFlow(orderID, rawData))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeSysTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

func (keeper BaseKeeper) SysTransferSignFinish(ctx sdk.Context, orderID string, signedTx []byte) sdk.Result {
	order := keeper.ok.GetOrder(ctx, orderID)
	if order == nil {
		return sdk.ErrNotFoundOrder(orderID).Result()
	}

	sysTransferOrder, valid := order.(*sdk.OrderSysTransfer)
	if !valid {
		return sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not systransfer order", orderID)).Result()
	}

	tokenInfo, err := keeper.checkSysTransferOrder(ctx, sysTransferOrder, sdk.OrderStatusWaitSign)
	if err != nil {
		return err.Result()
	}

	symbol := order.GetSymbol()
	chain := tokenInfo.Chain.String()

	var txHash string

	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		return sdk.ErrInvalidTx("Not support UtxoBased systransfer temporary").Result()

	case sdk.AccountBased:
		result, hash := keeper.verifyAccountBasedSignedTx(sysTransferOrder.FromAddress, chain, chain, sysTransferOrder.RawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}

		txHash = hash
	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	sysTransferOrder.TxHash = txHash
	sysTransferOrder.Status = sdk.OrderStatusSignFinish
	sysTransferOrder.SignedTx = signedTx
	keeper.ok.SetOrder(ctx, sysTransferOrder)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), orderID, sdk.OrderTypeSysTransfer, sdk.OrderStatusSignFinish))
	flows = append(flows, keeper.rk.NewSysTransferSignFinishFlow(orderID, signedTx))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeSysTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

func (keeper BaseKeeper) SysTransferFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, costFee sdk.Int) sdk.Result {
	order := keeper.ok.GetOrder(ctx, orderID)
	if order == nil {
		return sdk.ErrNotFoundOrder(orderID).Result()
	}
	sysTransferOrder, valid := order.(*sdk.OrderSysTransfer)
	if !valid {
		return sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not systransfer order", orderID)).Result()
	}

	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("systransfer from not a validator :%v", fromCUAddr)).Result()
	}

	tokenInfo, err := keeper.checkSysTransferOrder(ctx, sysTransferOrder, sdk.OrderStatusSignFinish)
	if err != nil {
		return err.Result()
	}

	result := sdk.Result{}
	confirmedFirstTime, _, _ := keeper.evidenceKeeper.Vote(ctx, sysTransferOrder.TxHash, fromCUAddr, types.NewTxVote(costFee.Int64(), true), uint64(ctx.BlockHeight()))
	if !confirmedFirstTime {
		return result
	}

	symbol := order.GetSymbol()
	chain := tokenInfo.Chain.String()
	opCUAddr, _ := sdk.CUAddressFromBase58(sysTransferOrder.OpCUaddress)

	opCUAst := keeper.ik.GetCUIBCAsset(ctx, opCUAddr)

	toCUAddr, _ := sdk.CUAddressFromBase58(sysTransferOrder.ToCU)
	toCUAst := keeper.ik.GetCUIBCAsset(ctx, toCUAddr)

	var balanceFlows []sdk.Flow
	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		return sdk.ErrInvalidTx("Not support UtxoBased systransfer temporary").Result()
	case sdk.AccountBased:
		localTx, err := keeper.cn.QueryAccountTransactionFromSignedData(chain, chain, sysTransferOrder.SignedTx)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}
		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, sysTransferOrder.CostFee))
		coins := sdk.NewCoins(sdk.NewCoin(chain, localTx.Amount))
		opCUAst.SubAssetCoinsHold(coins.Add(feeCoins))
		opCUAst.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, sysTransferOrder.CostFee.Sub(costFee))))
		opCUAst.AddGasUsed(coins.Add(sdk.NewCoins(sdk.NewCoin(chain, costFee))))

		//update order.CostFee
		sysTransferOrder.CostFee = costFee
		if toCUAst.GetCUType() == sdk.CUTypeUser {
			toCUAst.AddGasReceived(sdk.NewCoins(sdk.NewCoin(chain, sysTransferOrder.Amount)))
			toCUAst.AddGasRemained(chain, sysTransferOrder.ToAddress, sysTransferOrder.Amount)
			usedFee := costFee.Add(sysTransferOrder.Amount)
			waitCollectOrderIDs := keeper.getWaitCollectOrderIDs(ctx, toCUAddr.String(), symbol)
			if len(waitCollectOrderIDs) > 0 {
				order := keeper.ok.GetOrder(ctx, waitCollectOrderIDs[0])
				collectOrder := order.(*sdk.OrderCollect)
				if collectOrder.CostFee.GT(usedFee) {
					_, flow, err := keeper.AddCoin(ctx, toCUAddr, sdk.NewCoin(chain, collectOrder.CostFee.Sub(usedFee)))
					if err != nil {
						return err.Result()
					}
					balanceFlows = append(balanceFlows, flow)
					collectOrder.CostFee = usedFee
					keeper.ok.SetOrder(ctx, order)
				}

			}
		} else {
			toCUAst.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, sysTransferOrder.Amount)))
		}
		keeper.ok.SetOrder(ctx, sysTransferOrder)

		//add deposit item into toCU, status is collected
		item, err := sdk.NewDepositItem(localTx.Hash, uint64(0), localTx.Amount, localTx.To, "", sdk.DepositItemStatusConfirmed)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("fail to create deposit item, %v %v %v", localTx.Hash, 0, localTx.Amount)).Result()
		}
		_ = keeper.ik.SaveDeposit(ctx, chain, toCUAddr, item)

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update order's status
	sysTransferOrder.Status = sdk.OrderStatusFinish
	keeper.ok.SetOrder(ctx, sysTransferOrder)

	if tokenInfo.IsNonceBased {
		//don't update local nonce
		nonce := opCUAst.GetNonce(chain, sysTransferOrder.FromAddress) + 1
		opCUAst.SetNonce(chain, nonce, sysTransferOrder.FromAddress)
		opCUAst.SetEnableSendTx(true, chain, sysTransferOrder.FromAddress)
	}
	keeper.ik.SetCUIBCAsset(ctx, opCUAst)
	keeper.ik.SetCUIBCAsset(ctx, toCUAst)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), orderID, sdk.OrderTypeSysTransfer, sdk.OrderStatusFinish))
	flows = append(flows, keeper.rk.NewSysTransferFinishFlow(orderID, costFee))
	flows = append(flows, balanceFlows...)

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeSysTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

func (keeper BaseKeeper) checkSysTransferOrder(ctx sdk.Context, sysTransferOrder *sdk.OrderSysTransfer, orderStatus sdk.OrderStatus) (tokenInfo *sdk.IBCToken, err sdk.Error) {

	symbol := sysTransferOrder.GetSymbol()
	//symbol check
	tokenInfo = keeper.tk.GetIBCToken(ctx, sdk.Symbol(symbol))
	if tokenInfo == nil {
		err = sdk.ErrUnSupportToken(fmt.Sprintf("%s does not exist", symbol))
		return
	}

	//chain check
	chain := tokenInfo.Chain.String()
	chainTokenInfo := keeper.tk.GetIBCToken(ctx, sdk.Symbol(chain))
	if chainTokenInfo == nil {
		err = sdk.ErrUnSupportToken(fmt.Sprintf("%s does not exist", chain))
		return
	}

	if !orderStatus.Match(sysTransferOrder.Status) {
		err = sdk.ErrInvalidOrder(fmt.Sprintf("order status %d doesn't match expctedStatus:%d", sysTransferOrder.Status, orderStatus))
		return
	}

	toCUAddr, _ := sdk.CUAddressFromBase58(sysTransferOrder.ToCU)

	toCU := keeper.ck.GetCU(ctx, toCUAddr)
	if toCU == nil {
		err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v's to CU does not exist", sysTransferOrder))
		return
	}

	// check AssetCoinsHold only if order is not at terminated status
	if !sysTransferOrder.GetOrderStatus().Terminated() {
		need := sdk.NewCoins(sdk.NewCoin(tokenInfo.Chain.String(), sysTransferOrder.Amount))
		fromCUAst := keeper.ik.GetCUIBCAsset(ctx, sysTransferOrder.GetCUAddress())
		have := fromCUAst.GetAssetCoinsHold()

		if have.AmountOf(tokenInfo.Chain.String()).LT(sysTransferOrder.Amount) {
			err = sdk.ErrInsufficientCoins(fmt.Sprintf("need %v, have %v", need, have))
			return
		}
	}

	return
}
