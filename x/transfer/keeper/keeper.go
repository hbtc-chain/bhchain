package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hbtc-chain/bhchain/codec"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
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
	CollectSignFinish(ctx sdk.Context, orderIDs []string, signedTx []byte, txHash string) sdk.Result
	CollectFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, costFee sdk.Int) sdk.Result

	Withdrawal(ctx sdk.Context, fromCU sdk.CUAddress, toAddr, orderID, symbol string, amt, gasFee sdk.Int) sdk.Result
	WithdrawalConfirm(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, valid bool) sdk.Result

	WithdrawalWaitSign(ctx sdk.Context, opCUAddr sdk.CUAddress, orderIDs, signHashes []string, rawData []byte) sdk.Result
	WithdrawalSignFinish(ctx sdk.Context, orderIDs []string, signedTx []byte, txHash string) sdk.Result
	WithdrawalFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, costFee sdk.Int, valid bool) sdk.Result
	CancelWithdrawal(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string) sdk.Result

	SysTransfer(ctx sdk.Context, fromCUAddr, toCUAddr sdk.CUAddress, toAddr, orderID, symbol string) sdk.Result
	SysTransferWaitSign(ctx sdk.Context, orderID, signHash string, rawData []byte) sdk.Result
	SysTransferSignFinish(ctx sdk.Context, orderID string, signedTx []byte) sdk.Result
	SysTransferFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, costFee sdk.Int) sdk.Result

	OpcuAssetTransfer(ctx sdk.Context, opCUAddr sdk.CUAddress, toAddr, orderID, symbol string, items []sdk.TransferItem) sdk.Result
	OpcuAssetTransferWaitSign(ctx sdk.Context, orderID string, signHashes []string, rawData []byte) sdk.Result
	OpcuAssetTransferSignFinish(ctx sdk.Context, orderID string, signedTx []byte, txHash string) sdk.Result
	OpcuAssetTransferFinish(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderID string, costFee sdk.Int) sdk.Result

	OrderRetry(ctx sdk.Context, fromCUAddr sdk.CUAddress, orderIDs []string, retryTimes uint32, evidences []types.EvidenceValidator) sdk.Result
}

// BaseKeeper manages transfers between accounts. It implements the ReceiptKeeper interface.
type BaseKeeper struct {
	BaseSendKeeper
	cdc            *codec.Codec
	storeKey       sdk.StoreKey
	ok             types.OrderKeeper
	sk             types.StakingKeeper
	evidenceKeeper types.EvidenceKeeper
	cn             types.Chainnode
	paramSpace     params.Subspace
}

func (keeper *BaseKeeper) SetEvidenceKeeper(evidenceKeeper types.EvidenceKeeper) {
	keeper.evidenceKeeper = evidenceKeeper
}

// NewBaseKeeper returns a new BaseKeeper
func NewBaseKeeper(cdc *codec.Codec, key sdk.StoreKey, ck types.CUKeeper, tk types.TokenKeeper, ok types.OrderKeeper, rk types.ReceiptKeeper, sk types.StakingKeeper, cn types.Chainnode,
	paramSpace params.Subspace,
	codespace sdk.CodespaceType, blacklistedAddrs map[string]bool) *BaseKeeper {

	ps := paramSpace.WithKeyTable(types.ParamKeyTable())
	return &BaseKeeper{
		BaseSendKeeper: NewBaseSendKeeper(ck, rk, tk, ps, codespace, blacklistedAddrs),
		cdc:            cdc,
		storeKey:       key,
		ok:             ok,
		sk:             sk,
		cn:             cn,
		paramSpace:     ps,
	}
}

// DelegateCoins performs delegation by deducting amt coins from an CustodianUnit with
// address addr. For vesting accounts, delegations amounts are tracked for both
// vesting and vested coins.
// The coins are then transferred from the delegator address to a ModuleAccount address.
// If any of the delegation amounts are negative, an error is returned.
func (keeper *BaseKeeper) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {

	var sdkErr sdk.Error
	delegatorAcc := keeper.ck.GetCU(ctx, delegatorAddr)
	if delegatorAcc == nil || delegatorAcc.GetCUType() == sdk.CUTypeOp {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("CustodianUnit %s does not exist", delegatorAddr))
		return sdkErr.Result(), sdkErr
	}

	moduleAcc := keeper.ck.GetCU(ctx, moduleAccAddr)
	if moduleAcc == nil {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("module CustodianUnit %s does not exist", moduleAccAddr))
		return sdkErr.Result(), sdkErr
	}

	if !amt.IsValid() {
		sdkErr = sdk.ErrInvalidCoins(amt.String())
		return sdkErr.Result(), sdkErr
	}

	oldCoins := delegatorAcc.GetCoins()

	_, hasNeg := oldCoins.SafeSub(amt)
	if hasNeg {
		sdkErr = sdk.ErrInsufficientCoins(
			fmt.Sprintf("insufficient CustodianUnit funds; %s < %s", oldCoins, amt))
		return sdkErr.Result(), sdkErr
	}

	if err := trackDelegation(delegatorAcc, ctx.BlockHeader().Time, amt); err != nil {
		sdkErr = sdk.ErrInternal(fmt.Sprintf("failed to track delegation: %v", err))
		return sdkErr.Result(), sdkErr

	}

	keeper.ck.SetCU(ctx, delegatorAcc)

	_, flows, err := keeper.AddCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		sdkErr = err
		return sdkErr.Result(), sdkErr
	}
	result := sdk.Result{}
	if len(flows) > 0 {
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}

	return result, nil
}

// UndelegateCoins performs undelegation by crediting amt coins to an CustodianUnit with
// address addr. For vesting accounts, undelegation amounts are tracked for both
// vesting and vested coins.
// The coins are then transferred from a ModuleAccount address to the delegator address.
// If any of the undelegation amounts are negative, an error is returned.
func (keeper *BaseKeeper) UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {
	var sdkErr sdk.Error
	delegatorAcc := keeper.ck.GetCU(ctx, delegatorAddr)
	if delegatorAcc == nil || delegatorAcc.GetCUType() == sdk.CUTypeOp {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("CustodianUnit %s does not exist", delegatorAddr))
		return sdkErr.Result(), sdkErr
	}

	moduleAcc := keeper.ck.GetCU(ctx, moduleAccAddr)
	if moduleAcc == nil {
		sdkErr = sdk.ErrUnknownAddress(fmt.Sprintf("module CustodianUnit %s does not exist", moduleAccAddr))
		return sdkErr.Result(), sdkErr
	}

	if !amt.IsValid() {
		sdkErr = sdk.ErrInvalidCoins(amt.String())
		return sdkErr.Result(), sdkErr
	}

	oldCoins := moduleAcc.GetCoins()

	newCoins, hasNeg := oldCoins.SafeSub(amt)
	if hasNeg {
		sdkErr = sdk.ErrInsufficientCoins(
			fmt.Sprintf("insufficient CustodianUnit funds; %s < %s", oldCoins, amt))
		return sdkErr.Result(), sdkErr
	}

	flows, err := keeper.SetCoins(ctx, moduleAccAddr, newCoins)
	if err != nil {
		sdkErr = err
		return sdkErr.Result(), sdkErr
	}

	if err := trackUndelegation(delegatorAcc, amt); err != nil {
		sdkErr = sdk.ErrInternal(fmt.Sprintf("failed to track undelegation: %v", err))
		return sdkErr.Result(), sdkErr
	}

	keeper.ck.SetCU(ctx, delegatorAcc)
	var result sdk.Result
	if len(flows) > 0 {
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}

	return result, nil
}

func (keeper *BaseKeeper) SetStakingKeeper(ctx sdk.Context, stakingKeeper types.StakingKeeper) {
	keeper.sk = stakingKeeper
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
		if orderType == sdk.OrderTypeKeyGen {
			order.SetOrderStatus(sdk.OrderStatusBegin)
			keeper.ok.SetOrder(ctx, order)
		}
		keeper.setOrderRetryTimes(ctx, txID, retryTimes)
		//add flow
		var flows []sdk.Flow
		flows = append(flows, keeper.rk.NewOrderFlow(sdk.Symbol(order.GetSymbol()), fromCUAddr, orderIDs[0], order.GetOrderType(), sdk.OrderStatusWaitSign))
		flows = append(flows, keeper.rk.NewOrderRetryFlow(orderIDs))
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeOrderRetry, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}
	if confirmed {
		keeper.handleOrderRetryEvidences(ctx, txID, retryTimes, validVotes)
	}

	return result
}

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) (sdk.Result, sdk.Error)
	SendCoins(ctx sdk.Context, fromAddr sdk.CUAddress, toAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)

	SubtractCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	SetCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) ([]sdk.Flow, sdk.Error)

	GetSendEnabled(ctx sdk.Context) bool
	SetSendEnabled(ctx sdk.Context, enabled bool)

	BlacklistedAddr(addr sdk.CUAddress) bool
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper
	rk         types.ReceiptKeeper
	tk         types.TokenKeeper
	paramSpace params.Subspace

	// list of addresses that are restricted from receiving transactions
	blacklistedAddrs map[string]bool
}

// NewBaseSendKeeper returns a new BaseSendKeeper.
func NewBaseSendKeeper(ck types.CUKeeper, rk types.ReceiptKeeper, tk types.TokenKeeper, paramSpace params.Subspace, codespace sdk.CodespaceType,
	blacklistedAddrs map[string]bool) BaseSendKeeper {

	return BaseSendKeeper{
		BaseViewKeeper:   NewBaseViewKeeper(ck, codespace),
		rk:               rk,
		tk:               tk,
		paramSpace:       paramSpace,
		blacklistedAddrs: blacklistedAddrs,
	}
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseSendKeeper) InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) (sdk.Result, sdk.Error) {
	// Safety check ensuring that when sending coins the keeper must maintain the
	// Check supply invariant and validity of Coins.
	result := sdk.Result{}
	if err := types.ValidateInputsOutputs(inputs, outputs); err != nil {
		return err.Result(), err
	}
	var flows []sdk.Flow
	for _, in := range inputs {
		_, inFlows, err := keeper.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err.Result(), err
		}

		for _, flow := range inFlows {
			flows = append(flows, flow)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMultiTransfer,
				sdk.NewAttribute(types.AttributeKeySender, in.Address.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, in.Coins.String()),
			),
		)
	}

	for _, out := range outputs {
		_, outFlows, err := keeper.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err.Result(), err
		}
		for _, flow := range outFlows {
			flows = append(flows, flow)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMultiTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, out.Address.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, out.Coins.String()),
			),
		)
	}

	if len(flows) > 0 {
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeMultiTransfer, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}

	return result, nil
}

// SendCoins moves coins from one CustodianUnit to another
func (keeper BaseSendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.CUAddress, toAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {
	result := sdk.Result{}
	_, fromFlows, err := keeper.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err.Result(), err
	}

	_, toFlows, err := keeper.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err.Result(), err
	}

	var flows []sdk.Flow
	for _, flow := range fromFlows {
		flows = append(flows, flow)
	}

	for _, flow := range toFlows {
		flows = append(flows, flow)
	}

	if len(flows) > 0 {
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, toAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		),
	},
	)

	return result, nil
}

// SubtractCoins subtracts amt from the coins at the addr.
//
// CONTRACT: If the CustodianUnit is a vesting CustodianUnit, the amount has to be spendable.
func (keeper BaseSendKeeper) SubtractCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error) {
	if !amt.IsValid() {
		return nil, nil, sdk.ErrInvalidCoins(amt.String())
	}

	cu := keeper.ck.GetCU(ctx, addr)
	if cu != nil && cu.GetCUType() == sdk.CUTypeOp {
		return nil, nil, sdk.ErrInvalidCUType(addr.String())
	}

	oldCoins, spendableCoins := sdk.NewCoins(), sdk.NewCoins()

	if cu != nil {
		oldCoins = cu.GetCoins()
		spendableCoins = oldCoins
	}

	// For non-vesting accounts, spendable coins will simply be the original coins.
	// So the check here is sufficient instead of subtracting from oldCoins.
	_, hasNeg := spendableCoins.SafeSub(amt)
	if hasNeg {
		return amt, nil, sdk.ErrInsufficientCoins(
			fmt.Sprintf("insufficient CustodianUnit funds; %s < %s", spendableCoins, amt),
		)
	}

	newCoins := oldCoins.Sub(amt) // should not panic as spendable coins was already checked
	flows, err := keeper.SetCoins(ctx, addr, newCoins)

	return newCoins, flows, err
}

// AddCoins adds amt to the coins at the addr.
func (keeper BaseSendKeeper) AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error) {
	cu := keeper.ck.GetCU(ctx, addr)

	if !amt.IsValid() {
		return nil, nil, sdk.ErrInvalidCoins(amt.String())
	}

	oldCoins := sdk.NewCoins()
	if cu != nil {
		oldCoins = cu.GetCoins()
	}

	newCoins := oldCoins.Add(amt)
	if newCoins.IsAnyNegative() {
		return amt, nil, sdk.ErrInsufficientCoins(
			fmt.Sprintf("insufficient CustodianUnit funds; %s < %s", oldCoins, amt),
		)
	}

	flows, err := keeper.SetCoins(ctx, addr, newCoins)
	return newCoins, flows, err
}

// SetCoins sets the coins at the addr.
func (keeper BaseSendKeeper) SetCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) ([]sdk.Flow, sdk.Error) {
	if !amt.IsValid() {
		return nil, sdk.ErrInvalidCoins(amt.String())
	}

	cu := keeper.ck.GetCU(ctx, addr)
	if cu == nil {
		cu = keeper.ck.NewCUWithAddress(ctx, sdk.CUTypeUser, addr) //TODO(Keep), Should get Type from set coins, or??
	}

	err := cu.SetCoins(amt)
	if err != nil {
		panic(err)
	}
	keeper.ck.SetCU(ctx, cu)

	flows := []sdk.Flow{}
	for _, balanceFlow := range cu.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}
	cu.ResetBalanceFlows()
	return flows, nil
}

// BlacklistedAddr checks if a given address is blacklisted (i.e restricted from
// receiving funds)
func (keeper BaseSendKeeper) BlacklistedAddr(addr sdk.CUAddress) bool {
	return keeper.blacklistedAddrs[addr.String()]
}

var _ ViewKeeper = (*BaseViewKeeper)(nil)

// ViewKeeper defines a module interface that facilitates read only access to
// CustodianUnit balances.
type ViewKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins
	HasCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) bool

	Codespace() sdk.CodespaceType
}

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	ck        types.CUKeeper
	codespace sdk.CodespaceType
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(ck types.CUKeeper, codespace sdk.CodespaceType) BaseViewKeeper {
	return BaseViewKeeper{ck: ck, codespace: codespace}
}

// Logger returns a module-specific logger.
func (keeper BaseViewKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetCoins returns the coins at the addr.
func (keeper BaseViewKeeper) GetCoins(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins {
	acc := keeper.ck.GetCU(ctx, addr)
	if acc == nil {
		return sdk.NewCoins()
	}
	return acc.GetCoins()
}

// HasCoins returns whether or not an CustodianUnit has at least amt coins.
func (keeper BaseViewKeeper) HasCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) bool {
	return keeper.GetCoins(ctx, addr).IsAllGTE(amt)
}

// Codespace returns the keeper's codespace.
func (keeper BaseViewKeeper) Codespace() sdk.CodespaceType {
	return keeper.codespace
}

// CONTRACT: assumes that amt is valid.
func trackDelegation(acc exported.CustodianUnit, blockTime time.Time, amt sdk.Coins) error {
	vacc, ok := acc.(exported.VestingCU)
	if ok {
		// TODO: return error on CustodianUnit.TrackDelegation
		vacc.TrackDelegation(blockTime, amt)
		return nil
	}

	return acc.SetCoins(acc.GetCoins().Sub(amt))
}

// CONTRACT: assumes that amt is valid.
func trackUndelegation(acc exported.CustodianUnit, amt sdk.Coins) error {
	vacc, ok := acc.(exported.VestingCU)

	if ok {
		// TODO: return error on CustodianUnit.TrackUndelegation
		vacc.TrackUndelegation(amt)
		return nil
	}

	return acc.SetCoins(acc.GetCoins().Add(amt))
}
