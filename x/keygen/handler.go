package keygen

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/hbtc-chain/bhchain/types"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset/exported"
	"github.com/hbtc-chain/bhchain/x/keygen/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgKeyGen:
			return handleMsgKeyGen(ctx, keeper, msg)
		case MsgPreKeyGen:
			return handleMsgPreKeyGen(ctx, keeper, msg)
		case MsgKeyGenWaitSign:
			return handleMsgKeyGenWaitSign(ctx, keeper, msg)
		case MsgKeyGenFinish:
			return handleMsgKeyGenFinish(ctx, keeper, msg)
		case MsgOpcuMigrationKeyGen:
			return handleMsgOpcuMigrationKeyGen(ctx, keeper, msg)
		case MsgNewOpCU:
			return handleMsgNewOpCU(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized token Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgKeyGen(ctx sdk.Context, keeper Keeper, msg MsgKeyGen) sdk.Result {
	ctx.Logger().Info("handleMsgKeyGen", "msg", msg)
	//user's fromAddr/toAddr cu account allow toAddr be nil,but opcu account must not be nil.
	symbol, fromAddr, toAddr := msg.Symbol, msg.From, msg.To
	ti := keeper.tk.GetIBCToken(ctx, symbol)
	if result := checkSymbol(symbol, ti, keeper); !result.IsOK() {
		return result
	}
	var toCUAst exported.CUIBCAsset
	fromCUAst := keeper.ik.GetCUIBCAsset(ctx, fromAddr)
	if fromAddr.Equals(toAddr) {
		toCUAst = fromCUAst
	} else {
		toCUAst = keeper.ik.GetCUIBCAsset(ctx, toAddr)
	}

	if toCUAst == nil {
		toCUAst = keeper.ik.NewCUIBCAssetWithAddress(ctx, sdk.CUTypeUser, toAddr)
	}

	toCU := keeper.ck.GetCU(ctx, toAddr)
	if toCU == nil {
		toCU = keeper.ck.GetOrNewCU(ctx, sdk.CUTypeUser, toAddr)
		keeper.ck.SetCU(ctx, toCU)
	}

	pubkeyEpochIndex := toCUAst.GetAssetPubkeyEpoch()
	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	feeCoin := getFeeCoin(toCUAst.GetCUType(), ti)
	if pubkeyEpochIndex > 0 {
		feeCoin = sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt())
	}
	have := keeper.trk.GetBalance(ctx, fromAddr, feeCoin.Denom)
	if have.LT(feeCoin.Amount) {
		return sdk.ErrInsufficientFee(fmt.Sprintf("From CU %s does not have enough fee. has:%v, need:%v", fromAddr.String(), have, feeCoin.Amount)).Result()
	}

	if result := checkOrderID(ctx, msg.OrderID, keeper); !result.IsOK() {
		return result
	}
	if toCUAst.GetAssetAddress(symbol.String(), curEpoch.Index) != "" {
		return sdk.ErrInvalidAddr(fmt.Sprintf("CU %s already has %v address", toAddr.String(), symbol)).Result()
	}

	if toCUAst.GetCUType() == sdk.CUTypeOp {
		if toCU.GetSymbol() != symbol.String() {
			return sdk.ErrInvalidSymbol(fmt.Sprintf("symbol:%v & Op CU.symbol:%v not equal", symbol, toCU.GetSymbol())).Result()
		}
		if !isValidator(curEpoch.KeyNodeSet, fromAddr) {
			return sdk.ErrInvalidAddr(fmt.Sprintf("from CU %s is not a validator", fromAddr.String())).Result()
		}
		if pubkeyEpochIndex > 0 {
			return sdk.ErrInvalidTx("OPCU has already keygen").Result()
		}
	}
	processOrderList := keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeKeyGen)
	if exist, orderID := checkCuKeyGenOrder(ctx, keeper, processOrderList, toAddr); exist {
		return sdk.ErrInvalidTx(fmt.Sprintf("ToCU %s has unfinished keygen order %s", toAddr.String(), orderID)).Result()
	}

	if pubkeyEpochIndex == curEpoch.Index {
		//6、如果subtoken，且tocu已有链上地址，直接copy。执行结束。
		chainAddr := toCUAst.GetAssetAddress(ti.Chain.String(), curEpoch.Index)
		if ti.Symbol != ti.Chain && chainAddr != "" {
			_ = toCUAst.SetAssetAddress(symbol.String(), chainAddr, curEpoch.Index)
			keeper.ik.SetCUIBCAsset(ctx, toCUAst)
			return sdk.Result{}
		}

		//7、toCU.AssetPubkey已经存在，address不存在，用AssetPubkey向chainnode要address.
		if pk := toCUAst.GetAssetPubkey(curEpoch.Index); pk != nil {
			mulAddress, err := keeper.cn.ConvertAddress(ti.Chain.String(), pk)
			if err != nil {
				return sdk.ErrInternal(fmt.Sprintf("chainnode err:%v", err)).Result()
			}
			if mulAddress != "" {
				err = toCUAst.SetAssetAddress(symbol.String(), mulAddress, curEpoch.Index)
				if err != nil {
					return sdk.ErrInternal(fmt.Sprintf("Set address error: %v", err)).Result()
				}
				if symbol != ti.Chain {
					if err := toCUAst.SetAssetAddress(ti.Chain.String(), mulAddress, curEpoch.Index); err != nil {
						return sdk.ErrInternal(fmt.Sprintf("Set chain asset address error: %v", err)).Result()
					}
				}
				keeper.ik.SetCUIBCAsset(ctx, toCUAst)

				// index extAddress to cuAddress
				keeper.ck.SetExtAddressWithCU(ctx, ti.Chain.String(), mulAddress, toCUAst.GetAddress())
				return sdk.Result{}
			}
		}
	}

	// 先从预生成公钥中绑定
	waitAssignKeyGens := keeper.GetWaitAssignKeyGenOrderIDs(ctx)
	for _, orderID := range waitAssignKeyGens {
		order := keeper.ok.GetOrder(ctx, orderID)
		if order.GetOrderType() != sdk.OrderTypeKeyGen || order.GetOrderStatus() != sdk.OrderStatusSignFinish {
			ctx.Logger().Error("Unexpected order", "type", order.GetOrderType(), "status", order.GetOrderStatus())
			keeper.DelWaitAssignKeyGenOrderID(ctx, orderID)
			continue
		}
		keygenOrder := order.(*sdk.OrderKeyGen)
		// 向 chainnode 获取 Address
		addr, err := keeper.cn.ConvertAddress(ti.Chain.String(), keygenOrder.Pubkey)
		if err != nil {
			ctx.Logger().Error("Convert address error", "chain", ti.Chain.String(), "pubkey", keygenOrder.Pubkey, "err", err)
			keeper.DelWaitAssignKeyGenOrderID(ctx, orderID)
			continue
		}
		// multisignaddress 写入to CU，PubKey 写入cu.AssetPubkey
		if result := setAddressAndPubkeyToCU(ctx, toCUAst, keeper, keygenOrder.Pubkey, addr, symbol.String(), ti.Chain.String(), curEpoch.Index); !result.IsOK() {
			return result
		}
		// 更新 order 状态
		keygenOrder.CUAddress = fromAddr
		keygenOrder.Symbol = symbol.String()
		keygenOrder.To = toAddr
		keygenOrder.OpenFee = feeCoin
		keygenOrder.Status = sdk.OrderStatusFinish
		keygenOrder.MultiSignAddress = addr
		keeper.ok.SetOrder(ctx, keygenOrder)
		keeper.DelWaitAssignKeyGenOrderID(ctx, orderID)

		// 清算 openfee, 生成 orderFlow, keyGenFinishFlow
		flows := make([]sdk.Flow, 0, 3+len(keygenOrder.KeyNodes))
		orderflow := keeper.rk.NewOrderFlow(symbol, fromAddr, orderID, sdk.OrderTypeKeyGen, sdk.OrderStatusFinish)
		keyGenFinishFlow := sdk.KeyGenFinishFlow{OrderID: orderID, ToAddr: toAddr.String(), IsPreKeyGen: true}
		flows = append(flows, orderflow, keyGenFinishFlow)

		if feeCoin.Amount.IsPositive() {
			_, flow, err := keeper.trk.SubCoin(ctx, fromAddr, feeCoin)
			if err != nil {
				return err.Result()
			}
			flows = append(flows, flow)
			keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(sdk.NewCoins(feeCoin)))
		}

		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
		result := sdk.Result{}
		keeper.rk.SaveReceiptToResult(receipt, &result)
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeKeyGenFinish,
				sdk.NewAttribute(types.AttributeKeyFrom, fromAddr.String()),
				sdk.NewAttribute(types.AttributeKeySender, msg.From.String()),
				sdk.NewAttribute(types.AttributeKeyTo, toCUAst.GetAddress().String()),
				sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol.String()),
				sdk.NewAttribute(types.AttributeKeyOrderID, orderID),
			),
		})

		result.Events = append(result.Events, ctx.EventManager().Events()...)
		return result

	}

	// 没有预生成的公钥则直接 keygen
	//hold openfee
	var transferFlows []sdk.Flow
	if pubkeyEpochIndex == 0 {
		var err sdk.Error
		transferFlows, err = keeper.trk.LockCoin(ctx, fromAddr, feeCoin)
		if err != nil {
			return err.Result()
		}
	}
	keeper.ik.SetCUIBCAsset(ctx, toCUAst)

	//8、生成KeyGenOrder， ID为msg.orderID
	excludedKeyNode := keeper.getExcludedKeyNode(ctx, curEpoch.KeyNodeSet)
	keynodes := make([]sdk.CUAddress, 0, len(curEpoch.KeyNodeSet))
	for _, val := range curEpoch.KeyNodeSet {
		if !val.Equals(excludedKeyNode) {
			keynodes = append(keynodes, val)
		}
	}
	if len(keynodes) == 0 {
		return sdk.ErrInsufficientValidatorNumForKeyGen("empty keynode list").Result()
	}
	order := keeper.ok.NewOrderKeyGen(ctx, fromAddr, msg.OrderID, symbol.String(), keynodes, uint64(sdk.Majority23(len(curEpoch.KeyNodeSet))), toAddr, feeCoin)
	keeper.ok.SetOrder(ctx, order)

	//9. generate orderflow, keygenflow, balanceFlows
	flows := make([]sdk.Flow, 0, 3)
	orderFlow := keeper.rk.NewOrderFlow(symbol, toAddr, msg.OrderID, sdk.OrderTypeKeyGen, sdk.OrderStatusBegin)
	keyGenFlow := sdk.KeyGenFlow{OrderID: msg.OrderID, Symbol: symbol, From: fromAddr, To: toAddr, IsPreKeyGen: false, ExcludedKeyNode: excludedKeyNode}
	flows = append(flows, orderFlow, keyGenFlow)
	flows = append(flows, transferFlows...)

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
	result := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &result)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeKeyGen,
			sdk.NewAttribute(types.AttributeKeyFrom, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyTo, msg.To.String()),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgKeyGenWaitSign(ctx sdk.Context, keeper Keeper, msg MsgKeyGenWaitSign) sdk.Result {
	ctx.Logger().Info("handleMsgKeyGenWaitSign", "msg", msg)
	//1. get order and apply base check
	order := keeper.ok.GetOrder(ctx, msg.OrderID)
	if order == nil {
		return sdk.ErrNotFoundOrder(fmt.Sprintf("The order of %s does not exist", msg.OrderID)).Result()
	}
	fromCUAddr, symbol, orderType, orderStatus := order.GetCUAddress(), order.GetSymbol(), order.GetOrderType(), order.GetOrderStatus()
	if orderType != sdk.OrderTypeKeyGen || orderStatus != sdk.OrderStatusBegin {
		return sdk.ErrInvalidTx(fmt.Sprintf("Order type: %v is not 'OrderTypeKeyGen', or its status:%v  is not 'begin'", orderType, orderStatus)).Result()
	}

	// 2.轮次必须一致
	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	if curEpoch.Index != msg.Epoch {
		return sdk.ErrInvalidTx(fmt.Sprintf("Keygen epoch %d is not equal to current epoch %d", msg.Epoch, curEpoch.Index)).Result()
	}

	//3、如果to CU是opCU，cu.symbol与msg.symbol必等,from CU必须是validator
	keyGenOrder, ok := order.(*sdk.OrderKeyGen)
	if !ok {
		return sdk.ErrInvalidTx("not a OrderKeyGen").Result()
	}

	toCUAst := keeper.ik.GetCUIBCAsset(ctx, keyGenOrder.To)
	if toCUAst != nil {
		if toCUAst.GetCUType() == sdk.CUTypeOp {
			toCU := keeper.ck.GetCU(ctx, keyGenOrder.To)
			if toCU.GetSymbol() != symbol {
				return sdk.ErrInvalidSymbol(fmt.Sprintf("%s is not Op CU %s's symbol:%v", symbol, toCU.GetAddress().String(), toCU.GetSymbol())).Result()
			}
			// order.from & msg.from may not equal
			if !isValidator(curEpoch.KeyNodeSet, fromCUAddr) || !isValidator(curEpoch.KeyNodeSet, msg.From) {
				return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%s or msg.From:%s is not a keynode", fromCUAddr.String(), msg.From.String())).Result()
			}
		}

		//4、检查tocu.symbol对应的address是否已经存在，
		if toCUAst.GetAssetAddress(symbol, msg.Epoch) != "" {
			return sdk.ErrInvalidTx("cu asset address already exist").Result()
		}
	} else {
		// 3.检查 from 必须为 validator
		if !isValidator(curEpoch.KeyNodeSet, msg.From) {
			return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%s is not a keynode", msg.From.String())).Result()
		}
	}

	// 6、验签
	signMsg := types.NewMsgKeyGenWaitSign(msg.From, msg.OrderID, msg.PubKey, msg.KeyNodes, []cutypes.StdSignature{}, msg.Epoch)
	if result := checkKeyNodesSigns(signMsg.GetSignBytes(), msg.KeySigs, keyGenOrder.KeyNodes); !result.IsOK() {
		return result
	}

	//7、修改KeyGenOrder等待验证签名，记录MultiSignAddress，生成KeyGenFinishFlow.
	keyGenOrder.Epoch = msg.Epoch
	keyGenOrder.SetOrderStatus(sdk.OrderStatusWaitSign)
	keyGenOrder.Pubkey = msg.PubKey
	keeper.ok.SetOrder(ctx, order)

	flows := make([]sdk.Flow, 0, 2)
	orderflow := keeper.rk.NewOrderFlow(sdk.Symbol(symbol), fromCUAddr, msg.OrderID, sdk.OrderTypeKeyGen, sdk.OrderStatusWaitSign)
	keyGenWaitSignFlow := sdk.KeyGenWaitSignFlow{OrderID: msg.OrderID, PubKey: msg.PubKey}
	flows = append(flows, orderflow, keyGenWaitSignFlow)

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
	result := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeKeyGenWaitSign,
			sdk.NewAttribute(types.AttributeKeyFrom, fromCUAddr.String()),
			sdk.NewAttribute(types.AttributeKeySender, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgKeyGenFinish(ctx sdk.Context, keeper Keeper, msg MsgKeyGenFinish) sdk.Result {
	ctx.Logger().Info("handleMsgKeyGenFinish", "msg", msg)
	//1. get order and apply base check
	order := keeper.ok.GetOrder(ctx, msg.OrderID)
	if order == nil {
		return sdk.ErrNotFoundOrder(fmt.Sprintf("The order of %s does not exist", msg.OrderID)).Result()
	}

	fromCUAddr, symbol, orderType, orderStatus := order.GetCUAddress(), order.GetSymbol(), order.GetOrderType(), order.GetOrderStatus()
	if orderType != sdk.OrderTypeKeyGen || orderStatus != sdk.OrderStatusWaitSign {
		return sdk.ErrInvalidTx(fmt.Sprintf("Order type: %v is not 'OrderTypeKeyGen', or its status:%v  is not 'begin'", orderType, orderStatus)).Result()
	}

	keyGenOrder, ok := order.(*sdk.OrderKeyGen)
	if !ok {
		return sdk.ErrInvalidTx("not a OrderKeyGen").Result()
	}

	// 2.轮次必须一致
	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	if curEpoch.Index != keyGenOrder.Epoch {
		return sdk.ErrInvalidTx(fmt.Sprintf("Keygen epoch %d is not equal to current epoch %d", keyGenOrder.Epoch, curEpoch.Index)).Result()
	}

	signHash := sdk.BytesToHash(keyGenOrder.Pubkey).Bytes()
	_, err := crypto.SigToPub(signHash, msg.Signature)
	if err != nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("keygen order get pubkey error")).Result()
	}

	// 6.生成 orderFlow, keyGenSignFinishFlow
	flows := make([]sdk.Flow, 0, 2)

	toCUAst := keeper.ik.GetCUIBCAsset(ctx, keyGenOrder.To)
	if toCUAst != nil {
		ti := keeper.tk.GetIBCToken(ctx, sdk.Symbol(symbol))
		addr, err := keeper.cn.ConvertAddress(ti.Chain.String(), keyGenOrder.Pubkey)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Convert address error, chain:%v, pubKey:%v, err:%v", ti.Chain.String(), keyGenOrder.Pubkey, err)).Result()

		}
		if result := setAddressAndPubkeyToCU(ctx, toCUAst, keeper, keyGenOrder.Pubkey, addr, symbol, ti.Chain.String(), keyGenOrder.Epoch); !result.IsOK() {
			return result
		}

		if toCUAst.GetCUType() == sdk.CUTypeOp && toCUAst.GetMigrationStatus() == sdk.MigrationKeyGenBegin {
			toCUAst.SetMigrationStatus(sdk.MigrationKeyGenFinish)
			keeper.ik.SetCUIBCAsset(ctx, toCUAst)
		}

		keyGenOrder.SetOrderStatus(sdk.OrderStatusFinish)
		orderFlow := keeper.rk.NewOrderFlow("", msg.Validator, msg.OrderID, sdk.OrderTypeKeyGen, sdk.OrderStatusFinish)
		keyGenSignFinishFlow := sdk.KeyGenFinishFlow{OrderID: msg.OrderID, IsPreKeyGen: false, ToAddr: keyGenOrder.To.String()}
		flows = append(flows, orderFlow, keyGenSignFinishFlow)

		//sub openfee
		openFee := keyGenOrder.OpenFee
		hasFee := openFee.IsPositive()
		if hasFee {
			_, balanceFlow, err := keeper.trk.SubCoinHold(ctx, fromCUAddr, openFee)
			if err != nil {
				return err.Result()
			}
			keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(sdk.NewCoins(openFee)))
			flows = append(flows, balanceFlow)
		}
	} else {
		keyGenOrder.SetOrderStatus(sdk.OrderStatusSignFinish)
		keeper.ok.RemoveProcessOrder(ctx, sdk.OrderTypeKeyGen, msg.OrderID)
		keeper.AddWaitAssignKeyGenOrderID(ctx, msg.OrderID)
		orderFlow := keeper.rk.NewOrderFlow("", msg.Validator, msg.OrderID, sdk.OrderTypeKeyGen, sdk.OrderStatusSignFinish)
		keyGenSignFinishFlow := sdk.KeyGenFinishFlow{OrderID: msg.OrderID, IsPreKeyGen: true, ToAddr: keyGenOrder.To.String()}
		flows = append(flows, orderFlow, keyGenSignFinishFlow)
	}

	keeper.ok.SetOrder(ctx, order)

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
	result := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeKeyGenFinish,
			sdk.NewAttribute(types.AttributeKeySender, msg.Validator.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, msg.OrderID),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgPreKeyGen(ctx sdk.Context, keeper Keeper, msg MsgPreKeyGen) sdk.Result {
	ctx.Logger().Info("handleMsgPreKeyGen", "msg", msg)
	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	if !curEpoch.MigrationFinished {
		return sdk.ErrInvalidTx("Cannot pre keygen in migration period").Result()
	}

	// 1.检查交易为验证人发出
	if !isValidator(curEpoch.KeyNodeSet, msg.From) {
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU %s is not a keynode", msg.From.String())).Result()
	}
	// 2.检查当前没有正在处理的 order
	processOrders := keeper.ok.GetProcessOrderList(ctx)
	if len(processOrders) > 0 {
		return sdk.ErrSystemBusy(fmt.Sprintf("Already has %d orders being processed", len(processOrders))).Result()
	}
	// 3.检查一次创建的订单数不能超过阈值
	if len(msg.OrderIDs) > MaxPreKeyGenOrders {
		return sdk.ErrPreKeyGenTooMany(fmt.Sprintf("Cannot create %d prekeygen orders in a time", MaxPreKeyGenOrders)).Result()
	}
	// 4.检查缓存的 keygen order 是否超出阈值
	waitAssignKeyGens := keeper.GetWaitAssignKeyGenOrderIDs(ctx)
	if len(waitAssignKeyGens)+len(msg.OrderIDs) > MaxWaitAssignKeyOrders {
		return sdk.ErrWaitAssignTooMany(fmt.Sprintf("Already has %d pre keygen orders, plus extra %d will exceed %d",
			len(waitAssignKeyGens), len(msg.OrderIDs), MaxWaitAssignKeyOrders)).Result()
	}
	// 5.orderID的检查，orderID是否合法uuid，全局查重
	for _, orderID := range msg.OrderIDs {
		if result := checkOrderID(ctx, orderID, keeper); !result.IsOK() {
			return result
		}
	}

	// 6.生成KeyGenOrder, orderflow, keygenflow
	excludedKeyNode := keeper.getExcludedKeyNode(ctx, curEpoch.KeyNodeSet)
	keynodes := make([]sdk.CUAddress, 0, len(curEpoch.KeyNodeSet))
	for _, val := range curEpoch.KeyNodeSet {
		if !val.Equals(excludedKeyNode) {
			keynodes = append(keynodes, val)
		}
	}
	if len(keynodes) == 0 {
		return sdk.ErrInsufficientValidatorNumForKeyGen("empty keynode list").Result()
	}
	flows := make([]sdk.Flow, 0, 2*len(msg.OrderIDs))
	for _, orderID := range msg.OrderIDs {
		zeroFee := sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt())
		order := keeper.ok.NewOrderKeyGen(ctx, msg.From, orderID, "", keynodes, uint64(sdk.Majority23(len(curEpoch.KeyNodeSet))), nil, zeroFee)
		keeper.ok.SetOrder(ctx, order)
		orderFlow := keeper.rk.NewOrderFlow("", nil, orderID, sdk.OrderTypeKeyGen, sdk.OrderStatusBegin)
		keyGenFlow := sdk.KeyGenFlow{OrderID: orderID, Symbol: "", From: msg.From, To: nil, IsPreKeyGen: true, ExcludedKeyNode: excludedKeyNode}
		flows = append(flows, orderFlow, keyGenFlow)
	}

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
	result := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &result)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePreKeyGen,
			sdk.NewAttribute(types.AttributeKeyFrom, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgOpcuMigrationKeyGen(ctx sdk.Context, keeper Keeper, msg MsgOpcuMigrationKeyGen) sdk.Result {
	ctx.Logger().Info("handleMsgOpcuMigrationKeyGen", "msg", msg)

	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	if curEpoch.MigrationFinished {
		return sdk.ErrInvalidTx("Migration has finished").Result()
	}

	if !isValidator(curEpoch.KeyNodeSet, msg.From) {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from CU %s is not a keynode", msg.From.String())).Result()
	}

	for _, orderID := range msg.OrderIDs {
		if result := checkOrderID(ctx, orderID, keeper); !result.IsOK() {
			return result
		}
	}

	excludedKeyNode := keeper.getExcludedKeyNode(ctx, curEpoch.KeyNodeSet)
	keynodes := make([]sdk.CUAddress, 0, len(curEpoch.KeyNodeSet))
	for _, val := range curEpoch.KeyNodeSet {
		if !val.Equals(excludedKeyNode) {
			keynodes = append(keynodes, val)
		}
	}
	if len(keynodes) == 0 {
		return sdk.ErrInsufficientValidatorNumForKeyGen("empty keynode list").Result()
	}

	threshold := uint64(sdk.Majority23(len(curEpoch.KeyNodeSet)))
	zeroFee := sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt())

	opcus := keeper.ck.GetOpCUs(ctx, "")
	if len(opcus) != len(msg.OrderIDs) {
		return sdk.ErrInvalidTx(fmt.Sprintf("Order id number %d is not equal to opcu number %d", len(msg.OrderIDs), len(opcus))).Result()
	}
	processOrderList := keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeKeyGen)
	var flows []sdk.Flow
	for i, opcu := range opcus {
		opcuAst := keeper.ik.GetCUIBCAsset(ctx, opcu.GetAddress())
		if opcuAst.GetMigrationStatus() != sdk.MigrationBegin {
			return sdk.ErrInvalidTx("Migration status is not Begin").Result()
		}
		opcuAst.SetMigrationStatus(sdk.MigrationKeyGenBegin)
		keeper.ik.SetCUIBCAsset(ctx, opcuAst)
		if exist, _ := checkCuKeyGenOrder(ctx, keeper, processOrderList, opcu.GetAddress()); exist {
			continue
		}
		order := keeper.ok.NewOrderKeyGen(ctx, msg.From, msg.OrderIDs[i], opcu.GetSymbol(), keynodes, threshold, opcu.GetAddress(), zeroFee)
		keeper.ok.SetOrder(ctx, order)
		orderFlow := keeper.rk.NewOrderFlow(sdk.Symbol(opcu.GetSymbol()), opcu.GetAddress(), msg.OrderIDs[i], sdk.OrderTypeKeyGen, sdk.OrderStatusBegin)
		keyGenFlow := sdk.KeyGenFlow{OrderID: msg.OrderIDs[i], Symbol: sdk.Symbol(opcu.GetSymbol()), From: msg.From, To: opcu.GetAddress(), IsPreKeyGen: false, ExcludedKeyNode: excludedKeyNode}
		flows = append(flows, orderFlow, keyGenFlow)
	}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
	result := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOpcuMigrationKeyGen,
			sdk.NewAttribute(types.AttributeKeyFrom, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyOrderIDs, strings.Join(msg.OrderIDs, ",")),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func checkSymbol(symbol sdk.Symbol, ti *sdk.IBCToken, keeper Keeper) sdk.Result {
	if ti == nil {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("not support token %v", symbol)).Result()
	}

	if ti.Chain.String() == sdk.NativeToken {
		return sdk.ErrInvalidTx("native token no need to key gen").Result()
	}

	return sdk.Result{Code: sdk.CodeOK}
}

func checkOrderID(ctx sdk.Context, orderID string, keeper Keeper) sdk.Result {
	if keeper.ok.IsExist(ctx, orderID) {
		return sdk.ErrInvalidTx(fmt.Sprintf("OrderID: %v is duplicate", orderID)).Result()
	}
	return sdk.Result{Code: sdk.CodeOK}
}

func isValidator(vals []sdk.CUAddress, cuAddress sdk.CUAddress) bool {
	for _, v := range vals {
		if cuAddress.Equals(v) {
			return true
		}
	}
	return false
}

func checkKeyNodesSigns(msg []byte, sigs []cutypes.StdSignature, keyNodes []sdk.CUAddress) sdk.Result {
	nodes := make(map[string]bool, len(keyNodes))
	signedNodes := make(map[string]bool, len(keyNodes))
	for _, node := range keyNodes {
		nodes[node.String()] = true
	}
	for _, sig := range sigs {
		pubKey := sig.PubKey
		node := sdk.CUAddressFromPubKey(pubKey)
		if !pubKey.VerifyBytes(msg, sig.Signature) {
			return sdk.ErrInternal(fmt.Sprintf("Verify signature fail. pubKey: %v, sig:%v", pubKey, sig.Signature)).Result()
		}
		if nodes[node.String()] {
			signedNodes[node.String()] = true
		}
	}
	if len(signedNodes) != len(nodes) {
		return sdk.ErrInternal(fmt.Sprintf("not enough valid sign, key nodes:%v, signed nodes:%v", nodes, signedNodes)).Result()
	}
	return sdk.Result{Code: sdk.CodeOK}
}

func checkCuKeyGenOrder(ctx sdk.Context, keeper Keeper, processOrderList []string, cu sdk.CUAddress) (bool, string) {
	for _, id := range processOrderList {
		order := keeper.ok.GetOrder(ctx, id)
		if order != nil {
			keyGenOrder, ok := order.(*sdk.OrderKeyGen)
			if ok && keyGenOrder.To.Equals(cu) {
				return true, id
			}
		}
	}
	return false, ""
}

func getFeeCoin(cutype sdk.CUType, ti *sdk.IBCToken) sdk.Coin {
	openFee := sdk.ZeroInt()
	if cutype == sdk.CUTypeOp {
		openFee = ti.SysOpenFee
	} else {
		openFee = ti.OpenFee
	}

	return sdk.NewCoin(sdk.NativeToken, openFee)
}

func setAddressAndPubkeyToCU(ctx sdk.Context, cuAst exported.CUIBCAsset, keeper Keeper,
	pubkey []byte, address string, symbol, chain string, epoch uint64) sdk.Result {
	if err := cuAst.SetAssetAddress(symbol, address, epoch); err != nil {
		return sdk.ErrInternal(fmt.Sprintf("Set asset address error: %v", err)).Result()
	}
	if symbol != chain {
		if err := cuAst.SetAssetAddress(chain, address, epoch); err != nil {
			return sdk.ErrInternal(fmt.Sprintf("Set chain asset address error: %v", err)).Result()
		}
	}

	if err := cuAst.SetAssetPubkey(pubkey, epoch); err != nil {
		return sdk.ErrInternal(fmt.Sprintf("Set asset public key error: %v", err)).Result()
	}
	keeper.ck.SetExtAddressWithCU(ctx, chain, address, cuAst.GetAddress())
	keeper.ik.SetCUIBCAsset(ctx, cuAst)

	return sdk.Result{Code: sdk.CodeOK}
}

//Handle
func handleMsgNewOpCU(ctx sdk.Context, keeper Keeper, msg MsgNewOpCU) sdk.Result {
	ctx.Logger().Info("handleMsgNewOpCU", "msg", msg)

	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	if !curEpoch.MigrationFinished {
		return sdk.ErrInvalidTx("Cannot new opcu in migration period").Result()
	}

	ti := keeper.tk.GetIBCToken(ctx, sdk.Symbol(msg.Symbol))
	if ti == nil {
		return sdk.ErrUnSupportToken(fmt.Sprintf("token %s not support", msg.Symbol)).Result()
	}
	if !msg.OpCUAddress.IsValidAddr() {
		return sdk.ErrInvalidAddr("invalid CU address").Result()
	}
	opcuAddress := msg.OpCUAddress

	// op cu address already used
	if c := keeper.ck.GetCU(ctx, opcuAddress); c != nil {
		return sdk.ErrInvalidAddr("the CU address already exist").Result()
	}

	if !isValidator(curEpoch.KeyNodeSet, msg.From) {
		return sdk.ErrInvalidAddr(fmt.Sprintf("from CU %s is not a keynode", msg.From)).Result()
	}

	cuLimit := ti.MaxOpCUNumber
	opcuCount := len(keeper.ck.GetOpCUs(ctx, msg.Symbol))
	if uint64(opcuCount) >= cuLimit {
		return sdk.ErrInternal(fmt.Sprintf("too many operation CU of %s", msg.Symbol)).Result()
	}

	opcu := keeper.ck.NewOpCUWithAddress(ctx, msg.Symbol, opcuAddress)
	if opcu == nil {
		return sdk.ErrInternal("create operation CU failed").Result()
	}

	opcuAst := keeper.ik.NewCUIBCAssetWithAddress(ctx, sdk.CUTypeOp, opcuAddress)
	if opcuAst == nil {
		return sdk.ErrInternal("create operation CU failed").Result()
	}

	keeper.ck.SetCU(ctx, opcu)
	keeper.ik.SetCUIBCAsset(ctx, opcuAst)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeKeyNewOPCU,
			sdk.NewAttribute(types.AttributeKeySender, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyTo, msg.OpCUAddress.String()),
		),
	})

	return sdk.Result{}
}
