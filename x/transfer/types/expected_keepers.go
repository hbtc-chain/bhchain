package types

import (
	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	evidencetypes "github.com/hbtc-chain/bhchain/x/evidence/exported"
	ibcexported "github.com/hbtc-chain/bhchain/x/ibcasset/exported"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

// CUKeeper defines the CustodianUnit contract that must be fulfilled when
// creating a x/bank keeper.
type CUKeeper interface {
	NewCUWithAddress(ctx sdk.Context, cuType sdk.CUType, addr sdk.CUAddress) exported.CustodianUnit

	GetCU(ctx sdk.Context, addr sdk.CUAddress) exported.CustodianUnit
	GetOrNewCU(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CustodianUnit
	GetAllCUs(ctx sdk.Context) []exported.CustodianUnit
	SetCU(ctx sdk.Context, acc exported.CustodianUnit)
	IterateCUs(ctx sdk.Context, process func(exported.CustodianUnit) bool)

	GetOpCUs(ctx sdk.Context, symbol string) []exported.CustodianUnit
	GetCUFromExtAddress(ctx sdk.Context, symbol, extAddress string) (sdk.CUAddress, error)
}

type IBCAssetKeeper interface {
	GetCUIBCAsset(context sdk.Context, addresses sdk.CUAddress) ibcexported.CUIBCAsset
	NewCUIBCAssetWithAddress(ctx sdk.Context, cuType sdk.CUType, cuaddr sdk.CUAddress) ibcexported.CUIBCAsset
	SetCUIBCAsset(ctx sdk.Context, cuAst ibcexported.CUIBCAsset)

	//Deposit operation
	GetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress) sdk.DepositList
	GetDepositListByHash(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string) sdk.DepositList
	SetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress, list sdk.DepositList)
	SaveDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, deposit sdk.DepositItem) error
	DelDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64)
	SetDepositStatus(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64, status sdk.DepositItemStatus) error
	GetDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) sdk.DepositItem
	IsDepositExist(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) bool
}

type TokenKeeper interface {
	GetIBCToken(ctx sdk.Context, symbol sdk.Symbol) *sdk.IBCToken
}

type ReceiptKeeper interface {
	// NewReceipt creates a new receipt with a list of flows
	NewReceipt(category sdk.CategoryType, flows []sdk.Flow) *sdk.Receipt

	NewOrderFlow(symbol sdk.Symbol, cuAddress sdk.CUAddress, orderID string, orderType sdk.OrderType,
		orderStatus sdk.OrderStatus) sdk.OrderFlow
	NewBalanceFlow(cuAddress sdk.CUAddress, symbol sdk.Symbol, orderID string, previousBalance,
		balanceChange, previousBalanceOnHold, balanceOnHoldChange sdk.Int) sdk.BalanceFlow
	NewDepositFlow(CuAddress, multisignedadress, symbol, txhash, orderID, memo string,
		index uint64, amount sdk.Int, depositType sdk.DepositType, epoch uint64) sdk.DepositFlow
	NewDepositConfirmedFlow(validOrderIds, invalidOrderIds []string) sdk.DepositConfirmedFlow
	NewOrderRetryFlow(orderIDs []string, excludedKeyNode sdk.CUAddress) sdk.OrderRetryFlow

	NewCollectWaitSignFlow(orderIDs []string, rawData []byte) sdk.CollectWaitSignFlow
	NewCollectSignFinishFlow(orderIDs []string, signedTx []byte) sdk.CollectSignFinishFlow
	NewCollectFinishFlow(orderIDs []string, costFee sdk.Int) sdk.CollectFinishFlow
	NewWithdrawalFlow(orderID, fromcu, toaddr, symbol string, amount, gasFee sdk.Int, status sdk.WithdrawStatus) sdk.WithdrawalFlow
	NewWithdrawalConfirmFlow(orderID string, status sdk.WithdrawStatus) sdk.WithdrawalConfirmFlow
	NewWithdrawalWaitSignFlow(orderIDs []string, opcu, fromAddr string, rawData []byte) sdk.WithdrawalWaitSignFlow
	NewWithdrawalSignFinishFlow(orderIDs []string, signedTx []byte) sdk.WithdrawalSignFinishFlow
	NewWithdrawalFinishFlow(orderIDs []string, costFee sdk.Int, valid bool) sdk.WithdrawalFinishFlow
	NewSysTransferFlow(orderID, fromcu, tocu, fromAddr, toaddr, symbol string, amount sdk.Int) sdk.SysTransferFlow
	NewSysTransferWaitSignFlow(orderID string, rawData []byte) sdk.SysTransferWaitSignFlow
	NewSysTransferSignFinishFlow(orderID string, signedTx []byte) sdk.SysTransferSignFinishFlow
	NewSysTransferFinishFlow(orderID string, costFee sdk.Int) sdk.SysTransferFinishFlow
	NewOpcuAssetTransferFlow(orderID, fromcu, fromAddr, toaddr, symbol string, items []sdk.TransferItem) sdk.OpcuAssetTransferFlow
	NewOpcuAssetTransferWaitSignFlow(orderID string, rawData []byte) sdk.OpcuAssetTransferWaitSignFlow
	NewOpcuAssetTransferSignFinishFlow(orderID string, signedTx []byte) sdk.OpcuAssetTransferSignFinishFlow
	NewOpcuAssetTransferFinishFlow(orderID string, costFee sdk.Int) sdk.OpcuAssetTransferFinishFlow

	SaveReceiptToResult(receipt *sdk.Receipt, result *sdk.Result) *sdk.Result
	GetReceiptFromResult(result *sdk.Result) (*sdk.Receipt, error)
}

type OrderKeeper interface {
	NewOrder(ctx sdk.Context, order sdk.Order) sdk.Order
	GetOrder(ctx sdk.Context, orderID string) sdk.Order
	SetOrder(ctx sdk.Context, order sdk.Order)
	DeleteOrder(ctx sdk.Context, order sdk.Order)
	IsExist(ctx sdk.Context, orderID string) bool
	NewOrderCollect(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string, collectFromCU sdk.CUAddress,
		collectFromAddress string, amount, gasPrice, gasLimit sdk.Int, txHash string, index uint64, memo string) *sdk.OrderCollect
	NewOrderWithdrawal(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
		amount, gasFee, costFee sdk.Int, withdrawToAddr, opCUAddr, txHash string) *sdk.OrderWithdrawal
	NewOrderSysTransfer(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
		amount, costFee sdk.Int, toCU, toAddr, opCUAddr, fromAddr string) *sdk.OrderSysTransfer
	NewOrderOpcuAssetTransfer(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
		items []sdk.TransferItem, toAddr string) *sdk.OrderOpcuAssetTransfer
	RemoveProcessOrder(ctx sdk.Context, orderType sdk.OrderType, orderID string)
	GetProcessOrderListByType(ctx sdk.Context, orderTypes ...sdk.OrderType) []string
}

type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool)
	GetEpochByHeight(ctx sdk.Context, height uint64) sdk.Epoch
	GetCurrentEpoch(ctx sdk.Context) sdk.Epoch
	SetMigrationFinished(ctx sdk.Context)
	IsActiveKeyNode(ctx sdk.Context, addr sdk.CUAddress) (bool, int)
}

type EvidenceKeeper interface {
	HandleBehaviour(ctx sdk.Context, behaviourKey string, validator sdk.ValAddress, height uint64, normal bool)
	Vote(sdk.Context, string, sdk.CUAddress, evidencetypes.Vote, uint64) (bool, bool, []*evidencetypes.VoteItem)
	VoteWithCustomBox(ctx sdk.Context, voteID string, voter sdk.CUAddress, vote evidencetypes.Vote, height uint64, newVoteBox evidencetypes.NewVoteBox) (bool, bool, []*evidencetypes.VoteItem)
}

type Chainnode interface {
	SupportChain(chain string) bool
	ValidAddress(chain, symbol, address string) (bool, string)
	QueryBalance(chain, symbol, address, contractAddress string, blockHeight uint64) (sdk.Int, error)
	QueryUtxo(chain, symbol string, vin *sdk.UtxoIn) (bool, error)
	QueryGasPrice(chain string) (sdk.Int, error)
	QueryUtxoTransaction(chain, symbol, hash string, asynMode bool) (*chainnode.ExtUtxoTransaction, error)
	QueryAccountTransaction(chain, symbol, hash string, asynMode bool) (*chainnode.ExtAccountTransaction, error)
	VerifyUtxoSignedTransaction(chain, symbol string, address []string, signedTxData []byte, vins []*sdk.UtxoIn) (bool, error)
	VerifyAccountSignedTransaction(chain, symbol string, address string, signedTxData []byte) (bool, error)
	QueryAccountTransactionFromSignedData(chain, symbol string, signedTxData []byte) (*chainnode.ExtAccountTransaction, error)
	QueryUtxoTransactionFromSignedData(chain, symbol string, signedTxData []byte, vins []*sdk.UtxoIn) (*chainnode.ExtUtxoTransaction, error)
	QueryAccountTransactionFromData(chain, symbol string, rawData []byte) (*chainnode.ExtAccountTransaction, []byte, error)
	QueryUtxoTransactionFromData(chain, symbol string, rawData []byte, vins []*sdk.UtxoIn) (*chainnode.ExtUtxoTransaction, [][]byte, error)
	QueryUtxoInsFromData(chain, symbol string, data []byte) ([]*sdk.UtxoIn, error)
}
