package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func (keeper BaseKeeper) OpcuAssetTransfer(ctx sdk.Context, opCUAddr sdk.CUAddress, toAddr, orderID, symbol string, items []sdk.TransferItem) sdk.Result {
	curEpoch := keeper.sk.GetCurrentEpoch(ctx)
	if curEpoch.MigrationFinished {
		return sdk.ErrInvalidTx("not the right val epoch").Result()
	}

	if sdk.IsIllegalOrderID(orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("invalid OrderID:%v", orderID)).Result()
	}
	if keeper.ok.IsExist(ctx, orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("order %v already exists", orderID)).Result()
	}
	hasItem := make(map[string]bool)
	for _, item := range items {
		key := fmt.Sprintf("%s-%d", item.Hash, item.Index)
		if hasItem[key] {
			return sdk.ErrInvalidTx("duplicated transfer items").Result()
		}
		hasItem[key] = true
	}

	if keeper.hasUnfinishedOrder(ctx, opCUAddr) {
		return sdk.ErrInvalidTx("Opcu has unfinished order").Result()
	}

	opCU := keeper.ck.GetCU(ctx, opCUAddr)
	if opCU == nil {
		return sdk.ErrInvalidAccount(opCUAddr.String()).Result()
	}

	if opCU.GetCUType() != sdk.CUTypeOp {
		return sdk.ErrInvalidTx(fmt.Sprintf("opcutransfer from a non op CU :%v", opCUAddr)).Result()
	}

	if !sdk.Symbol(symbol).IsValidTokenName() {
		return sdk.ErrInvalidSymbol(symbol).Result()
	}
	if !keeper.tk.IsTokenSupported(ctx, sdk.Symbol(symbol)) {
		return sdk.ErrUnSupportToken(symbol).Result()
	}

	tokenInfo := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
	chain := tokenInfo.Chain.String()
	fromAddr := opCU.GetAssetAddress(chain, curEpoch.Index-1)
	if fromAddr == "" {
		opCU.SetMigrationStatus(sdk.MigrationFinish)
		keeper.ck.SetCU(ctx, opCU)
		keeper.checkOpcusMigrationStatus(ctx, curEpoch)
		return sdk.Result{}
	}
	if !opCU.IsEnabledSendTx(chain, fromAddr) {
		return sdk.ErrInvalidTx(fmt.Sprintf("opcutransfer not sendable tx now :%v", opCUAddr)).Result()
	}

	valid, canonicalToAddr := keeper.cn.ValidAddress(chain, symbol, toAddr)
	if !valid {
		return sdk.ErrInvalidAddr(fmt.Sprintf("%v is not a valid address", toAddr)).Result()
	}

	toAsset := opCU.GetAssetByAddr(symbol, canonicalToAddr)
	if toAsset == sdk.NilAsset {
		return sdk.ErrInvalidAddr(fmt.Sprintf("%v does not belong to cu %v", canonicalToAddr, opCU.GetAddress().String())).Result()
	}
	if toAsset.Epoch != curEpoch.Index {
		return sdk.ErrInvalidAddr("to addr not belong to currenct epoch").Result()
	}

	if tokenInfo.TokenType == sdk.UtxoBased {
		if items[0].Amount.IsZero() {
			if keeper.checkUtxoOpcuAstTransferFinish(ctx, fromAddr, symbol, opCU) {
				opCU.SetMigrationStatus(sdk.MigrationFinish)
				keeper.ck.SetCU(ctx, opCU)
				keeper.checkOpcusMigrationStatus(ctx, curEpoch)
				return sdk.Result{}
			}
			return sdk.ErrInvalidTx("Opcu transfer items are empty").Result()
		}

		depositList := keeper.ck.GetDepositList(ctx, symbol, opCU.GetAddress())
		depositList = depositList.Filter(func(d sdk.DepositItem) bool {
			return d.ExtAddress == fromAddr && d.Status == sdk.DepositItemStatusConfirmed
		})

		if len(items) > sdk.MaxVinNum {
			return sdk.ErrInvalidTx(fmt.Sprintf("Opcu transfer too many utxoins(%x) one time", len(items))).Result()
		}
		if len(items) != sdk.MaxVinNum && len(depositList) > len(items) {
			return sdk.ErrInvalidTx(fmt.Sprintf("Opcu transfer utxo number %d is not enouch", len(items))).Result()
		}

		sum := sdk.ZeroInt()
		for _, item := range items {
			depositItem := keeper.ck.GetDeposit(ctx, symbol, opCUAddr, item.Hash, item.Index)
			if depositItem == sdk.DepositNil || !depositItem.Amount.Equal(item.Amount) ||
				depositItem.Status == sdk.DepositItemStatusInProcess || depositItem.ExtAddress != fromAddr {
				return sdk.ErrInvalidTx(fmt.Sprintf("Invalid DepositItem(%v)", item.Hash)).Result()
			}
			sum = sum.Add(item.Amount)
		}

		if sum.LTE(keeper.utxoOpcuAstTransferThreshold(len(items), tokenInfo)) {
			for _, item := range items {
				keeper.ck.DelDeposit(ctx, symbol, opCUAddr, item.Hash, item.Index)
			}
			burnedCoins := sdk.NewCoins(sdk.NewCoin(symbol, sum))
			opCU.SubAssetCoins(burnedCoins)
			opCU.AddGasUsed(burnedCoins)
			keeper.ck.SetCU(ctx, opCU)
			if keeper.checkUtxoOpcuAstTransferFinish(ctx, fromAddr, symbol, opCU) {
				opCU.SetMigrationStatus(sdk.MigrationFinish)
				keeper.ck.SetCU(ctx, opCU)
				keeper.checkOpcusMigrationStatus(ctx, curEpoch)
			}
			return sdk.Result{}
		}

		for _, item := range items {
			_ = keeper.ck.SetDepositStatus(ctx, symbol, opCUAddr, item.Hash, item.Index, sdk.DepositItemStatusInProcess)
		}

	} else if tokenInfo.TokenType == sdk.AccountBased {
		if len(items) > sdk.LimitAccountBasedOrderNum {
			return sdk.ErrInvalidTx("opcu transfer is locked, not suitable").Result()
		}

		opcuSymbol := opCU.GetSymbol()
		status := opCU.GetMigrationStatus()
		if opcuSymbol != symbol && symbol == chain && status == sdk.MigrationKeyGenFinish {
			return sdk.ErrInvalidTx(fmt.Sprintf("opcu transfer should transfer symbol(%v) first", opcuSymbol)).Result()
		}

		have := opCU.GetAssetCoins().AmountOf(symbol)
		if !items[0].Amount.Equal(have) {
			return sdk.ErrInvalidTx(fmt.Sprintf("opcu transfer amount not equal,need:%v, actual have:%v", items[0].Amount, have)).Result()
		}

		if symbol == chain {
			if items[0].Amount.LT(tokenInfo.SysTransferAmount()) {
				opCU.SubAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, items[0].Amount)))
				opCU.AddGasUsed(sdk.NewCoins(sdk.NewCoin(chain, items[0].Amount)))
				opCU.SetMigrationStatus(sdk.MigrationFinish)
				keeper.ck.SetCU(ctx, opCU)
				keeper.checkOpcusMigrationStatus(ctx, curEpoch)
				return sdk.Result{}
			}
		} else {
			if items[0].Amount.IsZero() {
				opCU.SetMigrationStatus(sdk.MigrationMainTokenFinish)
				keeper.ck.SetCU(ctx, opCU)
				return sdk.Result{}
			}
		}

		opCU.SetEnableSendTx(false, chain, fromAddr)
	} else {
		return sdk.ErrInvalidTx(fmt.Sprintf("UnSupported tokenType:%v", tokenInfo.TokenType)).Result()
	}

	opCUAstTransferOrder := keeper.ok.NewOrderOpcuAssetTransfer(ctx, opCUAddr, orderID, symbol, items, canonicalToAddr)
	if opCUAstTransferOrder == nil {
		return sdk.ErrInvalidOrder(fmt.Sprintf("Fail to create order:%v", orderID)).Result()
	}
	keeper.ok.SetOrder(ctx, opCUAstTransferOrder)

	//onhold needed coins
	opCU.SetMigrationStatus(sdk.MigrationAssetBegin)
	keeper.ck.SetCU(ctx, opCU)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), opCUAddr, orderID, sdk.OrderTypeOpcuAssetTransfer, sdk.OrderStatusBegin))
	flows = append(flows, keeper.rk.NewOpcuAssetTransferFlow(orderID, opCUAddr.String(), fromAddr, canonicalToAddr, symbol, items))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeOpcuAssetTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

func (keeper BaseKeeper) OpcuAssetTransferWaitSign(ctx sdk.Context, orderID string, signHashes []string, rawData []byte) sdk.Result {
	tokenInfo, order, err := keeper.checkOpcuTransferOrder(ctx, orderID, sdk.OrderStatusBegin)
	if err != nil {
		return err.Result()
	}

	curEpoch := keeper.sk.GetCurrentEpoch(ctx)

	symbol := order.GetSymbol()
	chain := tokenInfo.Chain.String()

	opCU := keeper.ck.GetCU(ctx, order.GetCUAddress())
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

	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		//formulate the vins
		vins, err := keeper.cn.QueryUtxoInsFromData(chain, symbol, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}
		for _, vin := range vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCU.GetAddress(), vin.Hash, vin.Index)
			if item == sdk.DepositNil {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v does not exist", vin.Hash, vin.Index)).Result()
			}

			if item.GetStatus() != sdk.DepositItemStatusInProcess {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v status is %v", vin.Hash, vin.Index, item.GetStatus())).Result()
			}

			vin.Address = item.ExtAddress
			vin.Amount = item.Amount
		}

		if len(vins) != len(order.TransfertItems) {
			return sdk.ErrInvalidTx(fmt.Sprintf("opcu transfer vins(%d) not match", len(vins))).Result()
		}

		for _, vin := range vins {
			found := false
			for _, item := range order.TransfertItems {
				if vin.Hash == item.Hash && vin.Index == item.Index && vin.Amount.Equal(item.Amount) {
					found = true
					break
				}
			}

			if !found {
				return sdk.ErrInvalidTx(fmt.Sprintf("opcu transfer vin(%v) not found in order", vin.Hash)).Result()
			}
		}

		tx, hashes, err := keeper.cn.QueryUtxoTransactionFromData(chain, symbol, rawData, vins)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
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

		if len(tx.Vouts) != 1 {
			return sdk.ErrInvalidTx("vout number should be 1").Result()
		}

		if tx.Vouts[0].Address != order.ToAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("mismatch vout Addr, expected:%v, got:%v", order.ToAddr, tx.Vouts[0].Address)).Result()
		}

	case sdk.AccountBased:
		if len(signHashes) != sdk.LimitAccountBasedOrderNum {
			return sdk.ErrInvalidTx(fmt.Sprintf("AccountBased token supports only one opcutastransfer at one time, signhashnum:%v", len(signHashes))).Result()
		}

		tx, hash, err := keeper.cn.QueryAccountTransactionFromData(chain, symbol, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		if tx.To != order.ToAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected opcuTransfer to address:%v, expected:%v", tx.To, order.ToAddr)).Result()
		}

		expectAmount := tx.Amount
		if order.Symbol == chain {
			expectAmount = tx.Amount.Add(tx.GasPrice.Mul(tx.GasLimit))
		}

		if !expectAmount.Equal(order.TransfertItems[0].Amount) {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected opcu asset transfer Amount:%v, expected:%v", expectAmount, order.TransfertItems[0].Amount)).Result()
		}

		validContractAddr := ""
		if tokenInfo.Issuer != "" {
			_, validContractAddr = keeper.cn.ValidAddress(tokenInfo.Chain.String(), tokenInfo.Symbol.String(), tokenInfo.Issuer)
		}
		if tx.ContractAddress != validContractAddr {
			return sdk.ErrInvalidTx(fmt.Sprintf("Unexpected opcu asset transfer contract address:%v, expected:%v", tx.ContractAddress, tokenInfo.Issuer)).Result()
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

		lastAsset := opCU.GetAsset(tokenInfo.Chain.String(), curEpoch.Index-1)
		if lastAsset == sdk.NilAsset {
			return sdk.ErrInvalidTx("asset not found").Result()
		}
		if lastAsset.Nonce != tx.Nonce {
			return sdk.ErrInvalidTx(fmt.Sprintf("tx nonce not equal, opcu :%v, rawdata:%v", lastAsset.Nonce, tx.Nonce)).Result()
		}

		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, tx.GasPrice.Mul(tx.GasLimit)))
		coins := sdk.NewCoins(sdk.NewCoin(symbol, tx.Amount))
		coins = coins.Add(feeCoins)

		have := opCU.GetAssetCoins()
		if chain == symbol {
			if have.AmountOf(chain).LT(coins.AmountOf(chain)) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("opCU has insufficient coins, need:%v, actual have:%v", coins, have)).Result()
			}
		} else {
			if have.AmountOf(chain).LT(feeCoins.AmountOf(chain)) || have.AmountOf(symbol).LT(tx.Amount) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("opCU has insufficient coins, need:%v, actual have:%v", coins, have)).Result()
			}
		}

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	order.Status = sdk.OrderStatusWaitSign
	order.RawData = make([]byte, len(rawData))
	copy(order.RawData, rawData)
	keeper.ok.SetOrder(ctx, order)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), orderID, sdk.OrderTypeOpcuAssetTransfer, sdk.OrderStatusWaitSign))
	flows = append(flows, keeper.rk.NewOpcuAssetTransferWaitSignFlow(orderID, rawData))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeOpcuAssetTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)
	return result
}

func (keeper BaseKeeper) OpcuAssetTransferSignFinish(ctx sdk.Context, orderID string, signedTx []byte, txHash string) sdk.Result {
	tokenInfo, order, err := keeper.checkOpcuTransferOrder(ctx, orderID, sdk.OrderStatusWaitSign)
	if err != nil {
		return err.Result()
	}

	curEpoch := keeper.sk.GetCurrentEpoch(ctx)

	symbol := tokenInfo.Symbol.String()
	chain := tokenInfo.Chain.String()

	opCU := keeper.ck.GetCU(ctx, order.GetCUAddress())

	lastAsset := opCU.GetAsset(symbol, curEpoch.Index-1)
	if lastAsset == sdk.NilAsset {
		return sdk.ErrInvalidTx("asset not found").Result()
	}
	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		result, hash := keeper.verifyUtxoBasedSignedTx(ctx, nil, order.GetCUAddress(), chain, symbol, order.RawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}

		txHash = hash
	case sdk.AccountBased:
		result, hash := keeper.verifyAccountBasedSignedTx(lastAsset.Address, chain, symbol, order.RawData, signedTx)
		if result.Code != sdk.CodeOK {
			return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, result.Log)).Result()
		}
		txHash = hash

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	order.Status = sdk.OrderStatusSignFinish
	order.SignedTx = make([]byte, len(signedTx))
	copy(order.SignedTx, signedTx)
	order.Txhash = txHash
	keeper.ok.SetOrder(ctx, order)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), orderID, sdk.OrderTypeOpcuAssetTransfer, sdk.OrderStatusSignFinish))
	flows = append(flows, keeper.rk.NewOpcuAssetTransferSignFinishFlow(orderID, signedTx, txHash))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeOpcuAssetTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) OpcuAssetTransferFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, costFee sdk.Int) sdk.Result {
	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("opcu asset transfer from not a validator :%v", fromCUAddr)).Result()
	}

	tokenInfo, order, err := keeper.checkOpcuTransferOrder(ctx, orderID, sdk.OrderStatusSignFinish)
	if err != nil {
		return err.Result()
	}

	confirmedFirstTime, _, _ := keeper.evidenceKeeper.Vote(ctx, order.Txhash, fromCUAddr, types.NewTxVote(costFee.Int64(), true), uint64(ctx.BlockHeight()))

	result := sdk.Result{}
	if !confirmedFirstTime {
		return result
	}

	symbol := tokenInfo.Symbol.String()
	chain := tokenInfo.Chain.String()

	opCU := keeper.ck.GetCU(ctx, order.GetCUAddress())
	curEpoch := keeper.sk.GetCurrentEpoch(ctx)
	lastAsset := opCU.GetAsset(chain, curEpoch.Index-1)
	if lastAsset == sdk.NilAsset {
		return sdk.ErrInvalidTx("asset not found").Result()
	}

	switch tokenInfo.TokenType {
	case sdk.UtxoBased:
		vins, err := keeper.cn.QueryUtxoInsFromData(chain, symbol, order.RawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result()
		}

		for _, vin := range vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCU.GetAddress(), vin.Hash, vin.Index)
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
		for i, vout := range tx.Vouts {
			if order.ToAddr == vout.Address {
				vin := sdk.NewUtxoIn(tx.Hash, uint64(i), vout.Amount, vout.Address)
				//formulate the changeback deposit item
				depositItem, err := sdk.NewDepositItem(vin.Hash, vin.Index, vin.Amount, vout.Address, "", sdk.DepositItemStatusConfirmed)
				if err != nil {
					return sdk.ErrInvalidOrder(fmt.Sprintf("fail to create deposit item, %v %v %v", vin.Hash, vin.Index, vin.Amount)).Result()
				}
				_ = keeper.ck.SaveDeposit(ctx, symbol, opCU.GetAddress(), depositItem)
			}
		}

		//delete used Vins from opCU
		for _, vin := range tx.Vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCU.GetAddress(), vin.Hash, vin.Index)
			if item == sdk.DepositNil {
				return sdk.ErrInvalidOrder(fmt.Sprintf("deposit item%v %v does not exist", vin.Hash, vin.Index)).Result()
			}
			keeper.ck.DelDeposit(ctx, symbol, order.GetCUAddress(), vin.Hash, vin.Index)
		}

		opCU.SubAssetCoins(sdk.NewCoins(sdk.NewCoin(chain, costFee)))
		opCU.AddGasUsed(sdk.NewCoins(sdk.NewCoin(chain, costFee)))

		if keeper.checkUtxoOpcuAstTransferFinish(ctx, lastAsset.Address, symbol, opCU) {
			opCU.SetMigrationStatus(sdk.MigrationFinish)
		}

	case sdk.AccountBased:
		//update opcu's assetcoinshold, and refund unused gas fee if necessary
		feeCoins := sdk.NewCoins(sdk.NewCoin(chain, costFee))
		opCU.SubAssetCoins(feeCoins)
		opCU.AddGasUsed(feeCoins)

		if symbol != chain {
			opCU.SetMigrationStatus(sdk.MigrationMainTokenFinish)
		} else {
			opCU.SetMigrationStatus(sdk.MigrationFinish)
		}

		if tokenInfo.IsNonceBased {
			opCU.SetNonce(chain, lastAsset.Nonce+1, lastAsset.Address)
		}
		opCU.SetEnableSendTx(true, chain, lastAsset.Address)

	case sdk.AccountSharedBased:
		return sdk.ErrInvalidTx("Not support AccountSharedBased temporary").Result()
	}

	//update order's status and costFee
	order.Status = sdk.OrderStatusFinish
	order.CostFee = costFee
	keeper.ok.SetOrder(ctx, order)
	keeper.ck.SetCU(ctx, opCU)

	keeper.checkOpcusMigrationStatus(ctx, curEpoch)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(symbol), order.GetCUAddress(), orderID, sdk.OrderTypeOpcuAssetTransfer, sdk.OrderStatusFinish))
	flows = append(flows, keeper.rk.NewOpcuAssetTransferFinishFlow(orderID, costFee))

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeOpcuAssetTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) checkOpcuTransferOrder(ctx sdk.Context, orderID string, orderStatus sdk.OrderStatus) (tokenInfo *sdk.TokenInfo, opcuTransferOrder *sdk.OrderOpcuAssetTransfer, err sdk.Error) {
	order := keeper.ok.GetOrder(ctx, orderID)
	if order == nil {
		err = sdk.ErrNotFoundOrder(fmt.Sprintf("orderid:%v does not exist", orderID))
		return
	}

	opcuTransferOrder, valid := order.(*sdk.OrderOpcuAssetTransfer)
	if !valid {
		err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v is not opcutransfer order", orderID))
		return
	}

	symbol := order.GetSymbol()
	tokenInfo = keeper.tk.GetTokenInfo(ctx, sdk.Symbol(symbol))

	if !orderStatus.Match(opcuTransferOrder.Status) {
		err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v status doesn't match expctedStatus:%v", opcuTransferOrder, orderStatus))
		return
	}

	if orderStatus == sdk.OrderStatusWaitSign || orderStatus == sdk.OrderStatusSignFinish {
		if len(opcuTransferOrder.RawData) == 0 {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v RawData is empty", order))
			return
		}
	}

	if orderStatus == sdk.OrderStatusSignFinish {
		if len(opcuTransferOrder.SignedTx) == 0 || opcuTransferOrder.Txhash == "" {
			err = sdk.ErrInvalidOrder(fmt.Sprintf("order %v SignTx or Txhash is empty", order))
			return
		}
	}

	return
}

func (keeper BaseKeeper) checkUtxoOpcuAstTransferFinish(ctx sdk.Context, lastAddr, symbol string, opCU exported.CustodianUnit) bool {
	depositList := keeper.ck.GetDepositList(ctx, symbol, opCU.GetAddress())
	depositList = depositList.Filter(func(d sdk.DepositItem) bool {
		return d.ExtAddress == lastAddr
	})
	return len(depositList) == 0
}
