package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/evidence"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func (keeper BaseKeeper) verifyAccountBasedSignedTx(fromAddr, chain, symbol string, rawData, signedTx []byte) (sdk.Result, string) {
	txHash := ""
	verified, err := keeper.cn.VerifyAccountSignedTransaction(chain, symbol, fromAddr, signedTx)
	if err != nil || !verified {
		return sdk.ErrInvalidTx(fmt.Sprintf("VerifyAccountSignedTransaction fail:%v, err:%v", hex.EncodeToString(signedTx), err)).Result(), txHash
	}

	tx, err := keeper.cn.QueryAccountTransactionFromSignedData(chain, symbol, signedTx)
	if err != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("QueryAccountTransactionFromSignedData Error:%v", signedTx)).Result(), txHash
	}

	if tx.From != fromAddr {
		return sdk.ErrInvalidTx(fmt.Sprintf("from an unexpected address:%v, expected address:%v", tx.From, fromAddr)).Result(), txHash
	}

	rawTx, _, err := keeper.cn.QueryAccountTransactionFromData(chain, symbol, rawData)
	if err != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("QueryAccountTransactionFromData Error:%v", err)).Result(), txHash
	}

	if tx.To != rawTx.To {
		return sdk.ErrInvalidTx(fmt.Sprintf("to an unexpected address:%v, expected address:%v", tx.To, rawTx.To)).Result(), txHash
	}

	if !tx.Amount.Equal(rawTx.Amount) {
		return sdk.ErrInvalidTx(fmt.Sprintf("amount mismatch,expected:%v, actual:%v", rawTx.Amount, tx.Amount)).Result(), txHash
	}

	if !tx.GasPrice.Equal(rawTx.GasPrice) {
		return sdk.ErrInvalidTx(fmt.Sprintf("gasPrice mismatch, expected:%v, actual:%v", rawTx.GasPrice, tx.GasPrice)).Result(), txHash
	}

	if !tx.GasLimit.Equal(rawTx.GasLimit) {
		return sdk.ErrInvalidTx(fmt.Sprintf("gasLimit mismatch, expected:%v, actual:%v", rawTx.GasLimit, tx.GasLimit)).Result(), txHash
	}

	if tx.ContractAddress != rawTx.ContractAddress {
		return sdk.ErrInvalidTx(fmt.Sprintf("contract address mismatch, expected:%v, actual:%v", rawTx.ContractAddress, tx.ContractAddress)).Result(), txHash
	}
	txHash = tx.Hash

	return sdk.Result{}, txHash
}

func (keeper BaseKeeper) verifyUtxoBasedSignedTx(ctx sdk.Context, vins []*sdk.UtxoIn, opCUAddr sdk.CUAddress, chain, symbol string, rawData, signedTx []byte) (sdk.Result, string) {
	var err error
	if len(vins) == 0 {
		vins, err = keeper.cn.QueryUtxoInsFromData(chain, symbol, rawData)
		if err != nil {
			return sdk.ErrInvalidTx(err.Error()).Result(), ""
		}
		for _, vin := range vins {
			item := keeper.ck.GetDeposit(ctx, symbol, opCUAddr, vin.Hash, vin.Index)

			if item == sdk.DepositNil {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v does not exist", vin.Hash, vin.Index)).Result(), ""
			}

			if item.GetStatus() != sdk.DepositItemStatusInProcess {
				return sdk.ErrInvalidTx(fmt.Sprintf("vin %v %v status is %v, not in_process", vin.Hash, vin.Index, item.GetStatus())).Result(), ""
			}

			vin.Address = item.ExtAddress
			vin.Amount = item.Amount
		}
	}
	fromAddrs := make([]string, len(vins))
	for i, vin := range vins {
		fromAddrs[i] = vin.Address
	}
	verified, err := keeper.cn.VerifyUtxoSignedTransaction(chain, symbol, fromAddrs, signedTx, vins)
	if err != nil || !verified {
		return sdk.ErrInvalidTx(fmt.Sprintf("Fail to verify signed transaction:%v, err:%v", signedTx, err)).Result(), ""
	}

	rawTx, _, err := keeper.cn.QueryUtxoTransactionFromData(chain, symbol, rawData, vins)
	if err != nil {
		return sdk.ErrInvalidTx(err.Error()).Result(), ""
	}

	tx, err := keeper.cn.QueryUtxoTransactionFromSignedData(chain, symbol, signedTx, vins)
	if err != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("Fail to get transaction from signed transaction:%v", signedTx)).Result(), ""
	}

	if len(tx.Vins) != len(rawTx.Vins) {
		return sdk.ErrInvalidTx(fmt.Sprintf("vin length mismatch, expected:%v,  actual:%v", len(rawTx.Vins), len(tx.Vins))).Result(), ""
	}
	for i, in := range tx.Vins {
		if !in.Equal(*rawTx.Vins[i]) {
			return sdk.ErrInvalidTx(fmt.Sprintf("vin mismatch, expected:%v,  actual:%v", rawTx.Vins[i].String(), in.String())).Result(), ""
		}
	}

	if len(tx.Vouts) != len(rawTx.Vouts) {
		return sdk.ErrInvalidTx(fmt.Sprintf("vout length mismatch, expected:%v, actual:%v", len(rawTx.Vouts), len(tx.Vouts))).Result(), ""
	}
	for i, out := range tx.Vouts {
		if !out.Equal(*rawTx.Vouts[i]) {
			return sdk.ErrInvalidTx(fmt.Sprintf("vout mismatch, expected:%v, actual:%v", rawTx.Vouts[i].String(), out.String())).Result(), ""
		}
	}

	if !rawTx.CostFee.Equal(tx.CostFee) {
		return sdk.ErrInvalidTx(fmt.Sprintf("costFee mismatch, expected:%v, actual:%v", rawTx.CostFee, tx.CostFee)).Result(), ""
	}

	return sdk.Result{}, tx.Hash
}

func (keeper BaseKeeper) isMigrationFinishedBySymbol(ctx sdk.Context, symbol string, curEpoch sdk.Epoch) bool {
	if curEpoch.MigrationFinished {
		return true
	}

	opCUs := keeper.ck.GetOpCUs(ctx, symbol)
	for _, opCU := range opCUs {
		if opCU.GetMigrationStatus() != sdk.MigrationFinish {
			return false
		}
	}

	return true
}

func (keeper BaseKeeper) checkOpcusMigrationStatus(ctx sdk.Context, curEpoch sdk.Epoch) {
	symbols := keeper.tk.GetSymbols(ctx)
	for _, symbol := range symbols {
		ti := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
		if ti != nil && ti.Chain != sdk.NativeToken {
			if !keeper.isMigrationFinishedBySymbol(ctx, symbol, curEpoch) {
				return
			}
		}
	}

	keeper.sk.SetMigrationFinished(ctx)
}

func (keeper BaseKeeper) hasProcessingSysTransfer(ctx sdk.Context, opcu sdk.CUAddress, chain, toAddr string) bool {
	for _, orderID := range keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeSysTransfer) {
		order := keeper.ok.GetOrder(ctx, orderID)
		if order == nil {
			continue
		}
		sysTransferOrder := order.(*sdk.OrderSysTransfer)
		if sysTransferOrder.ToCU == opcu.String() && sysTransferOrder.ToAddress == toAddr {
			tokenInfo := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(order.GetSymbol()))
			if tokenInfo.Chain.String() == chain {
				return true
			}
		}
	}
	return false

}

func (keeper BaseKeeper) getWaitCollectOrderIDs(ctx sdk.Context, cuAddr string, symbol string) []string {
	waitCollectOrderIDs := []string{}
	for _, orderID := range keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeCollect) {
		order := keeper.ok.GetOrder(ctx, orderID)
		if order == nil {
			continue
		}
		collectOrder := order.(*sdk.OrderCollect)
		if collectOrder.CollectFromCU.String() == cuAddr && collectOrder.Symbol == symbol && collectOrder.Status == sdk.OrderStatusBegin {
			waitCollectOrderIDs = append(waitCollectOrderIDs, collectOrder.ID)
		}
	}

	return waitCollectOrderIDs
}

func (keeper BaseKeeper) checkNeedSysTransfer(ctx sdk.Context, chain, toAddr string, gasFee sdk.Int, cu exported.CustodianUnit) bool {
	//if cu.GetCUType() == sdk.CUTypeOp {
	//	ownedCoins := cu.GetAssetCoins()
	//	feeCoins := sdk.NewCoins(sdk.NewCoin(chain, gasFee.Mul(sdk.NewInt(types.MaxSystransferNum))))
	//	if ownedCoins.IsAllGTE(feeCoins) {
	//		return false
	//	}
	//} else {
	//	gasRemained := cu.GetGasRemained(chain, toAddr)
	//	if gasRemained.GTE(gasFee) {
	//		return false
	//	}
	//}

	for _, orderID := range keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeCollect, sdk.OrderTypeWithdrawal, sdk.OrderTypeOpcuAssetTransfer) {
		order := keeper.ok.GetOrder(ctx, orderID)
		if order == nil || order.GetSymbol() == chain || order.GetOrderStatus() != sdk.OrderStatusBegin {
			continue
		}
		switch orderDetail := order.(type) {
		case *sdk.OrderCollect:
			if orderDetail.CollectFromAddress == toAddr {
				return true
			}
		case *sdk.OrderWithdrawal:
			if cu.GetCUType() == sdk.CUTypeOp {
				return true
			}
		case *sdk.OrderOpcuAssetTransfer:
			if orderDetail.GetCUAddress().Equals(cu.GetAddress()) {
				return true
			}
		}
	}
	return false
}

func (keeper BaseKeeper) hasUnfinishedOrder(ctx sdk.Context, opcu sdk.CUAddress) bool {
	for _, orderID := range keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeCollect, sdk.OrderTypeWithdrawal, sdk.OrderTypeSysTransfer) {
		order := keeper.ok.GetOrder(ctx, orderID)
		if order == nil {
			continue
		}
		switch orderDetail := order.(type) {
		case *sdk.OrderCollect:
			if orderDetail.CollectToCU.Equals(opcu) {
				return true
			}
		case *sdk.OrderWithdrawal:
			if orderDetail.OpCUaddress == opcu.String() {
				return true
			}
		case *sdk.OrderSysTransfer:
			if orderDetail.OpCUaddress == opcu.String() || orderDetail.ToCU == opcu.String() {
				return true
			}
		}
	}
	return false
}

func (keeper BaseKeeper) handleOrderRetryEvidences(ctx sdk.Context, txID string, retryTimes uint32, votes []*evidence.VoteItem) {
	if keeper.hasEvidenceHandled(ctx, txID, retryTimes) {
		return
	}
	evidenceCounter := make(map[types.EvidenceValidator]int)
	for _, vote := range votes {
		if evids, ok := vote.Vote.([]types.EvidenceValidator); ok {
			for _, e := range evids {
				evidenceCounter[e]++
			}
		}
	}
	curEpoch := keeper.sk.GetCurrentEpoch(ctx)
	valsNum, threshold := len(curEpoch.KeyNodeSet), sdk.Majority23(len(curEpoch.KeyNodeSet))
	var maliciousValidators []sdk.CUAddress
	for validator, count := range evidenceCounter {
		if count >= threshold {
			validatorAddr, _ := sdk.CUAddressFromBase58(validator.Validator)
			maliciousValidators = append(maliciousValidators, validatorAddr)
		}
	}
	// 若存在恶意验证人，或者收集完全部的投票，则进行行为标记
	if len(maliciousValidators) > 0 || len(votes) == valsNum {
		for _, val := range curEpoch.KeyNodeSet {
			var found bool
			for _, maliciousValidator := range maliciousValidators {
				if maliciousValidator.Equals(val) {
					found = true
				}
			}
			keeper.evidenceKeeper.HandleBehaviour(ctx, evidence.DsignBehaviourKey, sdk.ValAddress(val), uint64(ctx.BlockHeight()), !found)
		}
		keeper.markEvidenceHandled(ctx, txID, retryTimes)
	}
}

func (keeper BaseKeeper) utxoOpcuAstTransferThreshold(vinNum int, tokenInfo *sdk.TokenInfo) sdk.Int {
	txSize := sdk.EstimateSignedUtxoTxSize(vinNum, 1)
	return txSize.Mul(tokenInfo.GasPrice).Quo(sdk.NewInt(sdk.KiloBytes))
}

func (keeper BaseKeeper) getOrderRetryTimes(ctx sdk.Context, txID string) uint32 {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(append(types.OrderRetryTimesPrefix, []byte(txID)...))
	if len(bz) == 0 {
		return 0
	}
	return binary.BigEndian.Uint32(bz)
}

func (keeper BaseKeeper) setOrderRetryTimes(ctx sdk.Context, txID string, retryTimes uint32) {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, retryTimes)
	store := ctx.KVStore(keeper.storeKey)
	store.Set(append(types.OrderRetryTimesPrefix, []byte(txID)...), buf)
}

func (keeper BaseKeeper) markEvidenceHandled(ctx sdk.Context, txID string, retryTimes uint32) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.GetOrderRetryEvidenceHandledKey(txID, retryTimes), []byte{})
}

func (keeper BaseKeeper) hasEvidenceHandled(ctx sdk.Context, txID string, retryTimes uint32) bool {
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(types.GetOrderRetryEvidenceHandledKey(txID, retryTimes))
}

func getOrderRawData(order sdk.Order) []byte {
	switch underlyingOrder := order.(type) {
	case *sdk.OrderCollect:
		return underlyingOrder.RawData
	case *sdk.OrderWithdrawal:
		return underlyingOrder.RawData
	case *sdk.OrderSysTransfer:
		return underlyingOrder.RawData
	case *sdk.OrderOpcuAssetTransfer:
		return underlyingOrder.RawData
	}
	return nil
}
