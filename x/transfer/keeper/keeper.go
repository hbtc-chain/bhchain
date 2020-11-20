package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hbtc-chain/bhchain/codec"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

var _ Keeper = (*BaseKeeper)(nil)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper

	SetEvidenceKeeper(evidenceKeeper types.EvidenceKeeper)
	DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	Deposit(ctx sdk.Context, fromCU, toCUAddr sdk.CUAddress, symbol sdk.Symbol, toAddr, hash string, index uint64, amt sdk.Int, orderID, memo string) sdk.Result
	ConfirmedDeposit(ctx sdk.Context, fromCUAddr sdk.CUAddress, validOrderIDs, invalidOrderIds []string) sdk.Result
	CollectWaitSign(ctx sdk.Context, toCUAddr sdk.CUAddress, orderIDs []string, rawData []byte) sdk.Result
	CollectSignFinish(ctx sdk.Context, orderIDs []string, signedTx []byte) sdk.Result
	CollectFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, costFee sdk.Int) sdk.Result

	Withdrawal(ctx sdk.Context, fromCU sdk.CUAddress, toAddr, orderID, symbol string, amt, gasFee sdk.Int) sdk.Result
	WithdrawalConfirm(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, valid bool) sdk.Result

	WithdrawalWaitSign(ctx sdk.Context, opCUAddr sdk.CUAddress, orderIDs []string, signHashes [][]byte, rawData []byte) sdk.Result
	WithdrawalSignFinish(ctx sdk.Context, orderIDs []string, signedTx []byte) sdk.Result
	WithdrawalFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, costFee sdk.Int, valid bool) sdk.Result
	CancelWithdrawal(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string) sdk.Result

	SysTransfer(ctx sdk.Context, fromCUAddr, toCUAddr sdk.CUAddress, toAddr, orderID, symbol string) sdk.Result
	SysTransferWaitSign(ctx sdk.Context, orderID string, signHash []byte, rawData []byte) sdk.Result
	SysTransferSignFinish(ctx sdk.Context, orderID string, signedTx []byte) sdk.Result
	SysTransferFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, costFee sdk.Int) sdk.Result

	OpcuAssetTransfer(ctx sdk.Context, opCUAddr sdk.CUAddress, toAddr, orderID, symbol string, items []sdk.TransferItem) sdk.Result
	OpcuAssetTransferWaitSign(ctx sdk.Context, orderID string, signHashes [][]byte, rawData []byte) sdk.Result
	OpcuAssetTransferSignFinish(ctx sdk.Context, orderID string, signedTx []byte) sdk.Result
	OpcuAssetTransferFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, costFee sdk.Int) sdk.Result

	OrderRetry(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, retryTimes uint32, evidences []types.EvidenceValidator) sdk.Result
}

// BaseKeeper manages transfers between accounts. It implements the ReceiptKeeper interface.
type BaseKeeper struct {
	cdc              *codec.Codec
	storeKey         sdk.StoreKey
	ok               types.OrderKeeper
	sk               types.StakingKeeper
	rk               types.ReceiptKeeper
	tk               types.TokenKeeper
	ck               types.CUKeeper
	ik               types.IBCAssetKeeper
	evidenceKeeper   types.EvidenceKeeper
	cn               types.Chainnode
	paramSpace       params.Subspace
	codespace        sdk.CodespaceType
	blacklistedAddrs map[string]bool
}

func NewBaseKeeper(cdc *codec.Codec, key sdk.StoreKey, ck types.CUKeeper, ik types.IBCAssetKeeper, tk types.TokenKeeper, ok types.OrderKeeper,
	rk types.ReceiptKeeper, sk types.StakingKeeper, cn types.Chainnode,
	paramSpace params.Subspace, codespace sdk.CodespaceType, blacklistedAddrs map[string]bool) *BaseKeeper {

	ps := paramSpace.WithKeyTable(types.ParamKeyTable())
	return &BaseKeeper{
		cdc:              cdc,
		storeKey:         key,
		ok:               ok,
		sk:               sk,
		rk:               rk,
		tk:               tk,
		ck:               ck,
		ik:               ik,
		cn:               cn,
		paramSpace:       ps,
		codespace:        codespace,
		blacklistedAddrs: blacklistedAddrs,
	}
}

func (keeper *BaseKeeper) SetEvidenceKeeper(evidenceKeeper types.EvidenceKeeper) {
	keeper.evidenceKeeper = evidenceKeeper
}

func (keeper *BaseKeeper) SetStakingKeeper(sk types.StakingKeeper) {
	keeper.sk = sk
}

// DelegateCoins performs delegation by deducting amt coins from an CustodianUnit with
// address addr. For vesting accounts, delegations amounts are tracked for both
// vesting and vested coins.
// The coins are then transferred from the delegator address to a ModuleAccount address.
// If any of the delegation amounts are negative, an error is returned.
func (keeper *BaseKeeper) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {

	var sdkErr sdk.Error
	delegatorAcc := keeper.ck.GetOrNewCU(ctx, sdk.CUTypeUser, delegatorAddr)
	if delegatorAcc == nil || delegatorAcc.GetCUType() == sdk.CUTypeOp {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("CustodianUnit %s does not exist", delegatorAddr))
		return sdkErr.Result(), sdkErr
	}

	moduleAcc := keeper.ck.GetCU(ctx, moduleAccAddr)
	if moduleAcc == nil {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("module CustodianUnit %s does not exist", moduleAccAddr))
		return sdkErr.Result(), sdkErr
	}

	_, flows, err := keeper.SendCoins(ctx, delegatorAddr, moduleAccAddr, amt)
	if err != nil {
		return err.Result(), err
	}

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result, nil
}

// UndelegateCoins performs undelegation by crediting amt coins to an CustodianUnit with
// address addr. For vesting accounts, undelegation amounts are tracked for both
// vesting and vested coins.
// The coins are then transferred from a ModuleAccount address to the delegator address.
// If any of the undelegation amounts are negative, an error is returned.
func (keeper *BaseKeeper) UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {
	var sdkErr sdk.Error
	delegatorAcc := keeper.ck.GetOrNewCU(ctx, sdk.CUTypeUser, delegatorAddr)
	if delegatorAcc == nil || delegatorAcc.GetCUType() == sdk.CUTypeOp {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("CustodianUnit %s does not exist", delegatorAddr))
		return sdkErr.Result(), sdkErr
	}

	moduleAcc := keeper.ck.GetCU(ctx, moduleAccAddr)
	if moduleAcc == nil {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("module CustodianUnit %s does not exist", moduleAccAddr))
		return sdkErr.Result(), sdkErr
	}

	_, flows, err := keeper.SendCoins(ctx, moduleAccAddr, delegatorAddr, amt)
	if err != nil {
		return err.Result(), err
	}

	var result sdk.Result
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result, nil
}

func (keeper *BaseKeeper) OrderRetry(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, retryTimes uint32, evidences []types.EvidenceValidator) sdk.Result {
	result := sdk.Result{}
	bValidator, _ := keeper.sk.IsActiveKeyNode(ctx, fromCUAddr)
	if !bValidator {
		return sdk.ErrInvalidTx(fmt.Sprintf("resetorder from not a validator :%v", fromCUAddr)).Result()
	}

	sort.Strings(orderIDs)
	order := keeper.ok.GetOrder(ctx, orderIDs[0])
	if order == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("invalid reset order data orderid:%v", orderIDs[0])).Result()
	}

	orderType := order.GetOrderType()
	if orderType == sdk.OrderTypeKeyGen {
		if order.GetOrderStatus() != sdk.OrderStatusBegin && order.GetOrderStatus() != sdk.OrderStatusWaitSign {
			return sdk.ErrInvalidTx(fmt.Sprintf("invalid reset order data orderid:%v", orderIDs[0])).Result()
		}
	} else {
		rawData := getOrderRawData(order)
		var expectOrderIDs []string
		for _, orderID := range keeper.ok.GetProcessOrderListByType(ctx, orderType) {
			processOrder := keeper.ok.GetOrder(ctx, orderID)
			if processOrder.GetOrderStatus() != sdk.OrderStatusWaitSign || bytes.Compare(getOrderRawData(processOrder), rawData) != 0 {
				continue
			}
			expectOrderIDs = append(expectOrderIDs, orderID)
		}
		sort.Strings(expectOrderIDs)
		if len(orderIDs) != len(expectOrderIDs) {
			return sdk.ErrInvalidTx("invalid order list").Result()
		}
		for i, orderID := range expectOrderIDs {
			if orderID != orderIDs[i] {
				return sdk.ErrInvalidTx("invalid order list").Result()
			}
		}
	}

	txID := strings.Join(orderIDs, "&")
	currentRetryTimes := keeper.getOrderRetryTimes(ctx, txID)
	// 只允许为当前轮和上一轮投票
	if retryTimes > currentRetryTimes+1 || retryTimes < currentRetryTimes {
		return sdk.ErrInvalidTx("invalid order retry times").Result()
	}
	voteID := fmt.Sprintf("%s-%d", txID, retryTimes)
	firstConfirmed, confirmed, validVotes := keeper.evidenceKeeper.VoteWithCustomBox(ctx, voteID, fromCUAddr, evidences, uint64(ctx.BlockHeight()), types.NewOrderRetryVoteBox)
	if firstConfirmed {
		var excludedKeyNode sdk.CUAddress
		if orderType == sdk.OrderTypeKeyGen {
			keyNodes := keeper.sk.GetCurrentEpoch(ctx).KeyNodeSet
			excludedKeyNode = keeper.getExcludedKeyNode(ctx, keyNodes)
			var i int
			for _, keyNode := range keyNodes {
				if !keyNode.Equals(excludedKeyNode) {
					keyNodes[i] = keyNode
					i++
				}
			}
			keygenOrder := order.(*sdk.OrderKeyGen)
			keygenOrder.KeyNodes = keyNodes[:i]
			order.SetOrderStatus(sdk.OrderStatusBegin)
			keeper.ok.SetOrder(ctx, order)
		}
		keeper.setOrderRetryTimes(ctx, txID, retryTimes)
		//add flow
		var flows []sdk.Flow
		flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(order.GetSymbol()), fromCUAddr, orderIDs[0], order.GetOrderType(), sdk.OrderStatusWaitSign))
		flows = append(flows, keeper.rk.NewOrderRetryFlow(orderIDs, excludedKeyNode))
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeOrderRetry, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}
	if confirmed {
		keeper.handleOrderRetryEvidences(ctx, txID, retryTimes, validVotes)
	}

	return result
}

// Codespace returns the keeper's codespace.
func (keeper BaseKeeper) Codespace() sdk.CodespaceType {
	return keeper.codespace
}
