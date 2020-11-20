package receipt

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
)

// ReceiptKeeperI defines a module interface for receipt, i.e. all kind of flows.
// The receipt is part of Result which is stored in the block results.
type ReceiptKeeperI interface {
	// NewReceipt creates a new receipt with a list of flows
	NewReceipt(category CategoryType, flows []Flow) *sdk.Receipt

	// NewOrderFlow creates a new order flow
	NewOrderFlow(symbol sdk.Symbol, cuAddress CUAddress, orderID string, orderType sdk.OrderType,
		orderStatus sdk.OrderStatus) OrderFlow
	// NewBalanceFlow creates a new balance flow for an asset
	NewBalanceFlow(cuAddress CUAddress, symbol Symbol, orderID string, previousBalance,
		balanceChange, previousBalanceOnHold, balanceOnHoldChange Int) BalanceFlow

	NewDepositFlow(CuAddress, multisignedadress, symbol, txhash, orderID, memo string,
		index uint64, amount Int, depositType sdk.DepositType, epoch uint64) DepositFlow
	// SaveReceiptToResult saves the receipt into a result.
	SaveReceiptToResult(receipt *sdk.Receipt, result *Result) *Result

	GetReceiptFromResult(result *Result) (*sdk.Receipt, error)
}

var _ ReceiptKeeperI = (*Keeper)(nil)

type Keeper struct {
	cdc *codec.Codec
}

func NewKeeper(cdc *codec.Codec) *Keeper {
	return &Keeper{
		cdc: cdc,
	}
}

func (r *Keeper) NewReceipt(category CategoryType, flows []Flow) *sdk.Receipt {
	return &sdk.Receipt{
		Category: category,
		Flows:    flows,
	}
}

func (r *Keeper) NewOrderFlow(symbol sdk.Symbol, cuAddress sdk.CUAddress, orderID string, orderType sdk.OrderType,
	orderStatus sdk.OrderStatus) OrderFlow {
	return OrderFlow{
		Symbol:      symbol,
		CUAddress:   cuAddress,
		OrderID:     orderID,
		OrderType:   orderType,
		OrderStatus: orderStatus,
	}
}

func (r *Keeper) NewBalanceFlow(cuAddress sdk.CUAddress, symbol Symbol, orderID string, previousBalance,
	balanceChange, previousBalanceOnHold, balanceOnHoldChange Int) BalanceFlow {
	return BalanceFlow{
		CUAddress:             cuAddress,
		Symbol:                symbol,
		PreviousBalance:       previousBalance,
		BalanceChange:         balanceChange,
		PreviousBalanceOnHold: previousBalanceOnHold,
		BalanceOnHoldChange:   balanceOnHoldChange,
	}
}

func (r *Keeper) NewDepositFlow(CuAddress, multisignedadress, symbol, txhash, orderID, memo string,
	index uint64, amount Int, depositType sdk.DepositType, epoch uint64) DepositFlow {
	return DepositFlow{
		CuAddress:         CuAddress,
		Multisignedadress: multisignedadress,
		Symbol:            symbol,
		Index:             index,
		Txhash:            txhash,
		Amount:            amount,
		OrderID:           orderID,
		DepositType:       depositType,
		Memo:              memo,
		Epoch:             epoch,
	}
}

func (r *Keeper) NewDepositConfirmedFlow(validOrderIds, invalidOrderIds []string) sdk.DepositConfirmedFlow {
	return sdk.DepositConfirmedFlow{
		ValidOrderIDs:   validOrderIds,
		InValidOrderIDs: invalidOrderIds,
	}
}

func (r *Keeper) NewOrderRetryFlow(orderIDs []string, excludedKeyNode sdk.CUAddress) sdk.OrderRetryFlow {
	return sdk.OrderRetryFlow{
		OrderIDs:        orderIDs,
		ExcludedKeyNode: excludedKeyNode,
	}
}

func (r *Keeper) SaveReceiptToResult(receipt *sdk.Receipt, result *Result) *Result {
	if !result.IsOK() {
		// Contract: do not save receipt in failed result.
		return result
	}
	result.Data = r.cdc.MustMarshalBinaryLengthPrefixed(*receipt)
	return result
}

func (r *Keeper) GetReceiptFromResult(result *Result) (*sdk.Receipt, error) {
	var rc sdk.Receipt

	if err := ModuleCdc.UnmarshalBinaryLengthPrefixed(result.Data, &rc); err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *Keeper) NewCollectWaitSignFlow(orderIDs []string, rawData []byte) sdk.CollectWaitSignFlow {
	return sdk.CollectWaitSignFlow{
		OrderIDs: orderIDs,
		RawData:  rawData,
	}
}

func (r *Keeper) NewCollectSignFinishFlow(orderIDs []string, signedTx []byte) sdk.CollectSignFinishFlow {
	return sdk.CollectSignFinishFlow{
		OrderIDs: orderIDs,
		SignedTx: signedTx,
	}
}

func (r *Keeper) NewCollectFinishFlow(orderIDs []string, costFee Int) sdk.CollectFinishFlow {
	return sdk.CollectFinishFlow{
		OrderIDs: orderIDs,
		CostFee:  costFee,
	}
}

func (r *Keeper) NewWithdrawalFlow(orderID, fromcu, toaddr, symbol string, amount, gasFee sdk.Int, status sdk.WithdrawStatus) sdk.WithdrawalFlow {
	return sdk.WithdrawalFlow{
		OrderID:        orderID,
		FromCu:         fromcu,
		ToAddr:         toaddr,
		Symbol:         symbol,
		Amount:         amount,
		GasFee:         gasFee,
		WithdrawStatus: status,
	}
}

func (r *Keeper) NewWithdrawalConfirmFlow(orderID string, status sdk.WithdrawStatus) sdk.WithdrawalConfirmFlow {
	return sdk.WithdrawalConfirmFlow{
		OrderID:        orderID,
		WithdrawStatus: status,
	}
}

func (r *Keeper) NewWithdrawalWaitSignFlow(orderIDs []string, opcu, fromAddr string, rawData []byte) sdk.WithdrawalWaitSignFlow {
	return sdk.WithdrawalWaitSignFlow{
		OrderIDs: orderIDs,
		OpCU:     opcu,
		FromAddr: fromAddr,
		RawData:  rawData,
	}
}

func (r *Keeper) NewWithdrawalSignFinishFlow(orderIDs []string, signedTx []byte) sdk.WithdrawalSignFinishFlow {
	return sdk.WithdrawalSignFinishFlow{
		OrderIDs: orderIDs,
		SignedTx: signedTx,
	}
}

func (r *Keeper) NewWithdrawalFinishFlow(orderIDs []string, costFee sdk.Int, valid bool) sdk.WithdrawalFinishFlow {
	return sdk.WithdrawalFinishFlow{
		OrderIDs: orderIDs,
		CostFee:  costFee,
		Valid:    valid,
	}
}

func (r *Keeper) NewSysTransferFlow(orderID, fromcu, tocu, fromAddr, toaddr, symbol string, amount Int) sdk.SysTransferFlow {
	return sdk.SysTransferFlow{
		OrderID:  orderID,
		FromCU:   fromcu,
		ToCU:     tocu,
		FromAddr: fromAddr,
		ToAddr:   toaddr,
		Symbol:   symbol,
		Amount:   amount,
	}
}

func (r *Keeper) NewSysTransferWaitSignFlow(orderID string, rawData []byte) sdk.SysTransferWaitSignFlow {
	return sdk.SysTransferWaitSignFlow{
		OrderID: orderID,
		RawData: rawData,
	}
}

func (r *Keeper) NewSysTransferSignFinishFlow(orderID string, signedTx []byte) sdk.SysTransferSignFinishFlow {
	return sdk.SysTransferSignFinishFlow{
		OrderID:  orderID,
		SignedTx: signedTx,
	}
}

func (r *Keeper) NewSysTransferFinishFlow(orderID string, costFee Int) sdk.SysTransferFinishFlow {
	return sdk.SysTransferFinishFlow{
		OrderID: orderID,
		CostFee: costFee,
	}
}

func (r *Keeper) NewOpcuAssetTransferFlow(orderID, opcu, fromAddr, toaddr, symbol string, items []sdk.TransferItem) sdk.OpcuAssetTransferFlow {
	flow := sdk.OpcuAssetTransferFlow{
		OrderID:       orderID,
		Opcu:          opcu,
		FromAddr:      fromAddr,
		ToAddr:        toaddr,
		Symbol:        symbol,
		TransferItems: make([]sdk.TransferItem, len(items)),
	}

	copy(flow.TransferItems, items)

	return flow
}

func (r *Keeper) NewOpcuAssetTransferWaitSignFlow(orderID string, rawData []byte) sdk.OpcuAssetTransferWaitSignFlow {
	return sdk.OpcuAssetTransferWaitSignFlow{
		OrderID: orderID,
		RawData: rawData,
	}
}

func (r *Keeper) NewOpcuAssetTransferSignFinishFlow(orderID string, signedTx []byte) sdk.OpcuAssetTransferSignFinishFlow {
	return sdk.OpcuAssetTransferSignFinishFlow{
		OrderID:  orderID,
		SignedTx: signedTx,
	}
}

func (r *Keeper) NewOpcuAssetTransferFinishFlow(orderID string, costFee Int) sdk.OpcuAssetTransferFinishFlow {
	return sdk.OpcuAssetTransferFinishFlow{
		OrderID: orderID,
		CostFee: costFee,
	}
}
