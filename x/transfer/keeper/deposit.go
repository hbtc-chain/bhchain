package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

//Deposit
func (keeper BaseKeeper) Deposit(ctx sdk.Context, fromCUAddr, toCUAddr sdk.CUAddress, symbol sdk.Symbol, toAddr, hash string, index uint64, amt sdk.Int, orderID, memo string) sdk.Result {
	if sdk.IsIllegalOrderID(orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("deposit invalid OrderID %v", orderID)).Result()
	}

	fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
	if fromCU == nil {
		return sdk.ErrInvalidAccount(fromCUAddr.String()).Result()
	}
	if fromCU.GetCUType() == sdk.CUTypeOp {
		return sdk.ErrInvalidTx(fmt.Sprintf("fromCU %v is opcu", fromCUAddr)).Result()
	}

	tokenInfo := keeper.tk.GetIBCToken(ctx, symbol)
	if tokenInfo == nil {
		return sdk.ErrUnSupportToken(symbol.String()).Result()
	}
	chain := tokenInfo.Chain.String()
	if !tokenInfo.DepositEnabled || !tokenInfo.SendEnabled || !keeper.IsSendEnabled(ctx) {
		return sdk.ErrTransactionIsNotEnabled(fmt.Sprintf("%v's deposit is not enabled temporary", symbol)).Result()
	}

	if tokenInfo.TokenType == sdk.AccountBased && index != 0 {
		return sdk.ErrInvalidTx(fmt.Sprintf("deposit invalid index:%v", index)).Result()
	}

	if amt.LT(tokenInfo.DepositThreshold) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("desposit %v LT deposit threshold %v", amt, tokenInfo.DepositThreshold)).Result()
	}

	valid, canonicalToAddr := keeper.cn.ValidAddress(chain, symbol.String(), toAddr)
	if !valid {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is an invalid address", toAddr)).Result()
	}

	toCUAst := keeper.ik.GetCUIBCAsset(ctx, toCUAddr)
	if toCUAst == nil {
		toCUAst = keeper.ik.NewCUIBCAssetWithAddress(ctx, sdk.CUTypeUser, toCUAddr)
		keeper.ik.SetCUIBCAsset(ctx, toCUAst)
	}

	asset := toCUAst.GetAssetByAddr(symbol.String(), canonicalToAddr)
	if asset == sdk.NilAsset {
		asset = toCUAst.GetAssetByAddr(chain, canonicalToAddr)
		if symbol.String() != chain && asset != sdk.NilAsset {
			_ = toCUAst.SetAssetAddress(symbol.String(), canonicalToAddr, asset.Epoch)
			keeper.ik.SetCUIBCAsset(ctx, toCUAst)
		} else {
			return sdk.ErrInvalidTx(fmt.Sprintf("Deposit addr %s does not belong to CU %s", canonicalToAddr, toCUAst.GetAddress().String())).Result()
		}
	}

	if keeper.ok.IsExist(ctx, orderID) {
		return sdk.ErrInvalidOrder(fmt.Sprintf("order %v already exists", orderID)).Result()
	}

	if keeper.ik.IsDepositExist(ctx, symbol.String(), toCUAddr, hash, index) {
		return sdk.ErrInvalidTx(fmt.Sprintf("deposit %v %v %v %v item already exist", symbol, toCUAddr, hash, index)).Result()
	}

	//ProcessOrder should be optimized.
	processOrderList := keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeCollect)
	for _, id := range processOrderList {
		order := keeper.ok.GetOrder(ctx, id)
		if order != nil {
			collectOrder := order.(*sdk.OrderCollect)
			if collectOrder.Txhash == hash && collectOrder.Index == index {
				return sdk.ErrInvalidTx(fmt.Sprintf("Tx: %v is already exist and not finish", hash)).Result()
			}
		}
	}

	collectOrder := keeper.ok.NewOrderCollect(ctx, toCUAddr, orderID, symbol.String(),
		toCUAddr, canonicalToAddr, amt, sdk.ZeroInt(), sdk.ZeroInt(), hash, index, memo)
	keeper.ok.SetOrder(ctx, collectOrder)

	var flows []sdk.Flow
	flows = append(flows, keeper.rk.NewOrderFlow(symbol, toCUAddr, orderID, sdk.OrderTypeDeposit, sdk.OrderStatusBegin))
	var depositType = sdk.DepositTypeCU
	if toCUAst.GetCUType() == sdk.CUTypeOp {
		depositType = sdk.DepositTypeOPCU
	}
	flows = append(flows, keeper.rk.NewDepositFlow(toCUAddr.String(), canonicalToAddr, symbol.String(), hash, orderID, memo, index, amt, depositType, asset.Epoch))

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeDeposit, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result
}

func (keeper BaseKeeper) ConfirmedDeposit(ctx sdk.Context, fromCUAddr sdk.CUAddress, validOrderIDs, invalidOrderIDs []string) sdk.Result {
	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("depositconfirm from not a validator :%v", fromCUAddr)).Result()
	}

	validOrderFlows, confirmedValidOrderIDs := keeper.processDepositOrderIDs(ctx, fromCUAddr, validOrderIDs, true)

	invalidOrderFlows, confirmedInvalidOrderIDs := keeper.processDepositOrderIDs(ctx, fromCUAddr, invalidOrderIDs, false)

	result := sdk.Result{}
	if len(confirmedValidOrderIDs) > 0 || len(confirmedInvalidOrderIDs) > 0 {
		var flows []sdk.Flow
		flows = append(flows, keeper.rk.NewOrderFlow("", nil, "", sdk.OrderTypeDeposit, sdk.OrderStatusFinish))
		flows = append(flows, keeper.rk.NewDepositConfirmedFlow(confirmedValidOrderIDs, confirmedInvalidOrderIDs))
		flows = append(flows, validOrderFlows...)
		flows = append(flows, invalidOrderFlows...)
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeDeposit, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}
	return result
}

func (keeper BaseKeeper) processDepositOrderIDs(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, valid bool) ([]sdk.Flow, []string) {
	var confirmedOrderIDs []string
	var flows []sdk.Flow
	for _, id := range orderIDs {
		order := keeper.ok.GetOrder(ctx, id)
		if order == nil || order.GetOrderType() != sdk.OrderTypeCollect {
			continue
		}
		collectOrder, ok := order.(*sdk.OrderCollect)
		if !ok {
			continue
		}
		confirmedFirstTime, _, _ := keeper.evidenceKeeper.Vote(ctx, id, fromCUAddr, types.NewTxVote(0, valid), uint64(ctx.BlockHeight()))
		if confirmedFirstTime {
			balanceFlows, err := keeper.confirmDepositOrder(ctx, collectOrder, valid)
			if err != nil {
				continue
			}
			confirmedOrderIDs = append(confirmedOrderIDs, id)
			flows = append(flows, balanceFlows...)
		}

	}
	return flows, confirmedOrderIDs
}

func (keeper BaseKeeper) confirmDepositOrder(ctx sdk.Context, order *sdk.OrderCollect, valid bool) ([]sdk.Flow, error) {
	var flows []sdk.Flow

	order.DepositStatus = sdk.DepositConfirmed
	if !valid {
		order.SetOrderStatus(sdk.OrderStatusFinish)
		keeper.ok.SetOrder(ctx, order)
		return flows, nil
	}

	toCUAst := keeper.ik.GetCUIBCAsset(ctx, order.CollectFromCU)
	collectAmt := order.Amount

	depositItemStatus := sdk.DepositItemStatusUnCollected
	haveWaitCollectItem := false
	dlt := keeper.ik.GetDepositList(ctx, order.Symbol, order.CollectFromCU)
	for _, item := range dlt {
		if item.GetStatus() == sdk.DepositItemStatusWaitCollect {
			haveWaitCollectItem = true
			break
		}
	}

	if toCUAst.GetCUType() == sdk.CUTypeOp {
		//update order status
		order.SetOrderStatus(sdk.OrderStatusFinish)
		depositItemStatus = sdk.DepositItemStatusConfirmed
	} else {
		tokenInfo := keeper.tk.GetIBCToken(ctx, sdk.Symbol(order.Symbol))
		gasFee := tokenInfo.GasPrice.Mul(tokenInfo.GasLimit)

		if toCUAst.GetGasRemained(tokenInfo.Chain.String(), order.CollectFromAddress).GTE(gasFee) {
			//if enough gas for collect, set deposit status to wait collect
			depositItemStatus = sdk.DepositItemStatusWaitCollect
		} else {
			//if not enough gas for collect, for main token, 1x gasFee needed, for non-main token, 2x gasFee need
			if tokenInfo.Symbol.String() == tokenInfo.Chain.String() {
				if !haveWaitCollectItem {
					collectAmt = collectAmt.Sub(tokenInfo.CollectFee().Amount)
					order.CostFee = tokenInfo.CollectFee().Amount
				}
				depositItemStatus = sdk.DepositItemStatusWaitCollect

			} else {
				if haveWaitCollectItem {
					depositItemStatus = sdk.DepositItemStatusWaitCollect
				} else {
					collectFee := tokenInfo.CollectFee()
					if keeper.GetBalance(ctx, toCUAst.GetAddress(), tokenInfo.Chain.String()).GTE(collectFee.Amount) {
						depositItemStatus = sdk.DepositItemStatusWaitCollect
						_, flow, err := keeper.SubCoin(ctx, toCUAst.GetAddress(), collectFee)
						if err != nil {
							return nil, err
						}
						flows = append(flows, flow)
						order.CostFee = collectFee.Amount
					}
				}
			}
		}
	}

	keeper.ok.SetOrder(ctx, order)

	//Add to deposit item
	depositItem, _ := sdk.NewDepositItem(order.Txhash, order.Index, order.Amount, order.CollectFromAddress, order.Memo, depositItemStatus)
	err := keeper.ik.SaveDeposit(ctx, order.Symbol, order.CollectFromCU, depositItem)
	if err != nil {
		return nil, err
	}

	if depositItemStatus == sdk.DepositItemStatusWaitCollect {
		_, flow, err := keeper.AddCoin(ctx, toCUAst.GetAddress(), sdk.NewCoin(order.Symbol, collectAmt))
		if err != nil {
			return nil, err
		}
		flows = append(flows, flow)
	}

	toCUAst.AddAssetCoins(sdk.NewCoins(sdk.NewCoin(order.Symbol, order.Amount)))
	keeper.ik.SetCUIBCAsset(ctx, toCUAst)
	return flows, nil
}
