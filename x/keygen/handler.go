package keygen

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
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
	ti := keeper.tk.GetTokenInfo(ctx, symbol)
	if result := checkSymbol(symbol, ti, keeper); !result.IsOK() {
		return result
	}
	var toCU exported.CustodianUnit
	fromCU := keeper.ck.GetCU(ctx, fromAddr)
	if fromAddr.Equals(toAddr) {
		toCU = fromCU
	} else {
		toCU = keeper.ck.GetOrNewCU(ctx, sdk.CUTypeUser, toAddr)
	}

	pubkeyEpochIndex := toCU.GetAssetPubkeyEpoch()
	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	feeCoin := getFeeCoin(toCU.GetCUType(), ti)
	feeCoins := sdk.NewCoins(feeCoin)
	if pubkeyEpochIndex > 0 {
		feeCoin = sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt())
	}
	fromCoins := fromCU.GetCoins()
	if fromCoins.AmountOf(feeCoin.Denom).LT(feeCoin.Amount) {
		return sdk.ErrInsufficientFee(fmt.Sprintf("From CU: %v no enough fee. nativetoken:%v,openfee:%v", fromAddr, fromCoins.AmountOf(sdk.NativeToken), feeCoin.Amount)).Result()
	}

	if result := checkOrderID(ctx, msg.OrderID, keeper); !result.IsOK() {
		return result
	}
	if toCU.GetAssetAddress(symbol.String(), curEpoch.Index) != "" {
		return sdk.ErrInvalidAddr(fmt.Sprintf("CU%v already have %v address", toAddr, symbol)).Result()
	}

	if toCU.GetCUType() == sdk.CUTypeOp {
		if toCU.GetSymbol() != symbol.String() {
			return sdk.ErrInvalidSymbol(fmt.Sprintf("symbol:%v & Op CU.symbol:%v not equal", symbol, toCU.GetSymbol())).Result()
		}
		if !isValidator(curEpoch.KeyNodeSet, fromAddr) {
			return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%v is not a validator", fromAddr)).Result()
		}
		if pubkeyEpochIndex > 0 {
			return sdk.ErrInvalidTx("OPCU has already keygen").Result()
		}
	}
	processOrderList := keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeKeyGen)
	if checkCuKeyGenOrder(ctx, keeper, processOrderList, toAddr) {
		return sdk.ErrInvalidTx(fmt.Sprintf("ToCU: %v is already exist and not finish", toAddr.String())).Result()
	}

	if pubkeyEpochIndex == curEpoch.Index {
		//6、如果subtoken，且tocu已有链上地址，直接copy。执行结束。
		chainAddr := toCU.GetAssetAddress(ti.Chain.String(), curEpoch.Index)
		if keeper.tk.IsSubToken(ctx, symbol) && chainAddr != "" {
			_ = toCU.SetAssetAddress(symbol.String(), chainAddr, curEpoch.Index)
			keeper.ck.SetCU(ctx, toCU)
			return sdk.Result{}
		}

		//7、toCU.AssetPubkey已经存在，address不存在，用AssetPubkey向chainnode要address.
		if pk := toCU.GetAssetPubkey(curEpoch.Index); pk != nil {
			mulAddress, err := keeper.cn.ConvertAddress(ti.Chain.String(), pk)
			if err != nil {
				return sdk.ErrInternal(fmt.Sprintf("chainnode err:%v", err)).Result()
			}
			if mulAddress != "" {
				err = toCU.SetAssetAddress(symbol.String(), mulAddress, curEpoch.Index)
				if err != nil {
					return sdk.ErrInternal(fmt.Sprintf("Set address error: %v", err)).Result()
				}
				if symbol != ti.Chain {
					if err := toCU.SetAssetAddress(ti.Chain.String(), mulAddress, curEpoch.Index); err != nil {
						return sdk.ErrInternal(fmt.Sprintf("Set chain asset address error: %v", err)).Result()
					}
				}
				keeper.ck.SetCU(ctx, toCU)

				// index extAddress to cuAddress
				keeper.ck.SetExtAddresseWithCU(ctx, ti.Chain.String(), mulAddress, toCU.GetAddress())
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
		if result := setAddressAndPubkeyToCU(ctx, toCU, keeper, keygenOrder.Pubkey, addr, symbol.String(), ti.Chain.String(), curEpoch.Index); !result.IsOK() {
			return result
		}
		// 更新 order 状态
		keygenOrder.CUAddress = fromAddr
		keygenOrder.Symbol = symbol.String()
		keygenOrder.To = toAddr
		keygenOrder.OpenFee = feeCoins
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
			fromCU.SubCoins(feeCoins)
			keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(feeCoins))
		}

		if len(fromCU.GetBalanceFlows()) > 0 {
			flows = append(flows, fromCU.GetBalanceFlows()[0])
			fromCU.ResetBalanceFlows()
		}

		keeper.ck.SetCU(ctx, fromCU)

		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeKeyGen, flows)
		result := sdk.Result{}
		keeper.rk.SaveReceiptToResult(receipt, &result)
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeKeyGenFinish,
				sdk.NewAttribute(types.AttributeKeyFrom, fromAddr.String()),
				sdk.NewAttribute(types.AttributeKeySender, msg.From.String()),
				sdk.NewAttribute(types.AttributeKeyTo, toCU.GetAddress().String()),
				sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol.String()),
				sdk.NewAttribute(types.AttributeKeyOrderID, orderID),
			),
		})

		result.Events = append(result.Events, ctx.EventManager().Events()...)
		return result

	}

	// 没有预生成的公钥则直接 keygen
	//hold openfee
	if pubkeyEpochIndex == 0 {
		fromCU.SubCoins(feeCoins)
		fromCU.AddCoinsHold(feeCoins)
		keeper.ck.SetCU(ctx, fromCU)
	}
	keeper.ck.SetCU(ctx, toCU)

	//8、生成KeyGenOrder， ID为msg.orderID
	vals := curEpoch.KeyNodeSet
	if len(vals) == 0 {
		return sdk.ErrInsufficientValidatorNumForKeyGen(fmt.Sprintf("validator's number:%v", len(vals))).Result()
	}
	keynodes := make([]sdk.CUAddress, len(vals))
	for i, val := range vals {
		keynodes[i] = val
	}
	order := keeper.ok.NewOrderKeyGen(ctx, fromAddr, msg.OrderID, symbol.String(), keynodes, uint64(sdk.Majority23(len(vals))), toAddr, feeCoins)
	keeper.ok.SetOrder(ctx, order)

	//9. generate orderflow, keygenflow, balanceFlows
	flows := make([]sdk.Flow, 0, 3)
	orderFlow := keeper.rk.NewOrderFlow(symbol, toAddr, msg.OrderID, sdk.OrderTypeKeyGen, sdk.OrderStatusBegin)
	keyGenFlow := sdk.KeyGenFlow{OrderID: msg.OrderID, Symbol: symbol, From: fromAddr, To: toAddr, IsPreKeyGen: false}
	flows = append(flows, orderFlow, keyGenFlow)
	bFlows := fromCU.GetBalanceFlows()
	fromCU.ResetBalanceFlows()
	for _, bFlow := range bFlows {
		flows = append(flows, bFlow)
	}

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

	toCU := keeper.ck.GetCU(ctx, keyGenOrder.To)
	if toCU != nil {
		if toCU.GetCUType() == sdk.CUTypeOp {
			if toCU.GetSymbol() != symbol {
				return sdk.ErrInvalidSymbol(fmt.Sprintf("%v is not Op CU.symbol:%v", symbol, toCU.GetSymbol())).Result()
			}
			// order.from & msg.from may not equal
			if !isValidator(curEpoch.KeyNodeSet, fromCUAddr) || !isValidator(curEpoch.KeyNodeSet, msg.From) {
				return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%v or msg.From:%v is not a validator", fromCUAddr, msg.From)).Result()
			}
		}

		//4、检查tocu.symbol对应的address是否已经存在，
		if toCU.GetAssetAddress(symbol, msg.Epoch) != "" {
			return sdk.ErrInvalidTx("cu asset address already exist").Result()
		}
	} else {
		// 3.检查 from 必须为 validator
		if !isValidator(curEpoch.KeyNodeSet, msg.From) {
			return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%v is not a validator", msg.From)).Result()
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

	toCU := keeper.ck.GetCU(ctx, keyGenOrder.To)
	if toCU != nil {
		ti := keeper.tk.GetTokenInfo(ctx, sdk.Symbol(symbol))
		addr, err := keeper.cn.ConvertAddress(ti.Chain.String(), keyGenOrder.Pubkey)
		if err != nil {
			return sdk.ErrInvalidTx(fmt.Sprintf("Convert address error, chain:%v, pubKey:%v, err:%v", ti.Chain.String(), keyGenOrder.Pubkey, err)).Result()

		}
		if result := setAddressAndPubkeyToCU(ctx, toCU, keeper, keyGenOrder.Pubkey, addr, symbol, ti.Chain.String(), keyGenOrder.Epoch); !result.IsOK() {
			return result
		}

		if toCU.GetCUType() == sdk.CUTypeOp && toCU.GetMigrationStatus() == sdk.MigrationKeyGenBegin {
			toCU.SetMigrationStatus(sdk.MigrationKeyGenFinish)
			keeper.ck.SetCU(ctx, toCU)
		}

		keyGenOrder.SetOrderStatus(sdk.OrderStatusFinish)

		//sub openfee
		fromCU := keeper.ck.GetCU(ctx, fromCUAddr)
		openFee := keyGenOrder.OpenFee
		hasFee := openFee.AmountOf(sdk.NativeToken).IsPositive()
		if hasFee {
			fromCU.SubCoinsHold(openFee)
			keeper.ck.SetCU(ctx, fromCU)
			keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(openFee))
		}

		orderFlow := keeper.rk.NewOrderFlow("", msg.Validator, msg.OrderID, sdk.OrderTypeKeyGen, sdk.OrderStatusFinish)
		keyGenSignFinishFlow := sdk.KeyGenFinishFlow{OrderID: msg.OrderID, IsPreKeyGen: false, ToAddr: keyGenOrder.To.String()}
		flows = append(flows, orderFlow, keyGenSignFinishFlow)
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
		return sdk.ErrInvalidAddr(fmt.Sprintf("From CU:%v is not a validator", msg.From)).Result()
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
	vals := curEpoch.KeyNodeSet
	if len(vals) == 0 {
		return sdk.ErrInsufficientValidatorNumForKeyGen(fmt.Sprintf("validator's number:%v", len(vals))).Result()
	}
	keynodes := make([]sdk.CUAddress, len(vals))
	for i, val := range vals {
		keynodes[i] = val
	}
	flows := make([]sdk.Flow, 0, 2*len(msg.OrderIDs))
	for _, orderID := range msg.OrderIDs {
		order := keeper.ok.NewOrderKeyGen(ctx, msg.From, orderID, "", keynodes, uint64(sdk.Majority23(len(vals))), nil, nil)
		keeper.ok.SetOrder(ctx, order)
		orderFlow := keeper.rk.NewOrderFlow("", nil, orderID, sdk.OrderTypeKeyGen, sdk.OrderStatusBegin)
		keyGenFlow := sdk.KeyGenFlow{OrderID: orderID, Symbol: "", From: msg.From, To: nil, IsPreKeyGen: true}
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
		return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%v is not a validator", msg.From)).Result()
	}

	for _, orderID := range msg.OrderIDs {
		if result := checkOrderID(ctx, orderID, keeper); !result.IsOK() {
			return result
		}
	}

	// 生成 keynodes
	vals := curEpoch.KeyNodeSet
	if len(vals) == 0 {
		return sdk.ErrInsufficientValidatorNumForKeyGen(fmt.Sprintf("validator's number:%v", len(vals))).Result()
	}
	keynodes := make([]sdk.CUAddress, len(vals))
	for i, val := range vals {
		keynodes[i] = val
	}
	threshold := uint64(sdk.Majority23(len(vals)))
	zeroFee := sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt()))

	opcus := keeper.ck.GetOpCUs(ctx, "")
	if len(opcus) != len(msg.OrderIDs) {
		return sdk.ErrInvalidTx(fmt.Sprintf("Order id number %d is not equal to opcu number %d", len(msg.OrderIDs), len(opcus))).Result()
	}
	processOrderList := keeper.ok.GetProcessOrderListByType(ctx, sdk.OrderTypeKeyGen)
	var flows []sdk.Flow
	for i, opcu := range opcus {
		if opcu.GetMigrationStatus() != sdk.MigrationBegin {
			return sdk.ErrInvalidTx("Migration status is not Begin").Result()
		}
		opcu.SetMigrationStatus(sdk.MigrationKeyGenBegin)
		keeper.ck.SetCU(ctx, opcu)
		if checkCuKeyGenOrder(ctx, keeper, processOrderList, opcu.GetAddress()) {
			continue
		}
		order := keeper.ok.NewOrderKeyGen(ctx, msg.From, msg.OrderIDs[i], opcu.GetSymbol(), keynodes, threshold, opcu.GetAddress(), zeroFee)
		keeper.ok.SetOrder(ctx, order)
		orderFlow := keeper.rk.NewOrderFlow(sdk.Symbol(opcu.GetSymbol()), opcu.GetAddress(), msg.OrderIDs[i], sdk.OrderTypeKeyGen, sdk.OrderStatusBegin)
		keyGenFlow := sdk.KeyGenFlow{OrderID: msg.OrderIDs[i], Symbol: sdk.Symbol(opcu.GetSymbol()), From: msg.From, To: opcu.GetAddress(), IsPreKeyGen: false}
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

func checkSymbol(symbol sdk.Symbol, ti *sdk.TokenInfo, keeper Keeper) sdk.Result {
	if ti == nil {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("not support token %v", symbol)).Result()
	}

	if ti.Chain.String() == sdk.NativeToken {
		return sdk.ErrInvalidTx("native token no need to key gen").Result()
	}

	if !keeper.cn.SupportChain(ti.Chain.String()) {
		return sdk.ErrInvalidTx("chain not supported").Result()
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

func checkCuKeyGenOrder(ctx sdk.Context, keeper Keeper, processOrderList []string, cu sdk.CUAddress) bool {
	for _, id := range processOrderList {
		order := keeper.ok.GetOrder(ctx, id)
		if order != nil {
			keyGenOrder, ok := order.(*sdk.OrderKeyGen)
			if ok && keyGenOrder.To.Equals(cu) {
				return true
			}
		}
	}
	return false
}

func getFeeCoin(cutype sdk.CUType, ti *sdk.TokenInfo) sdk.Coin {
	openFee := sdk.ZeroInt()
	if cutype == sdk.CUTypeOp {
		openFee = ti.SysOpenFee
	} else {
		openFee = ti.OpenFee
	}

	return sdk.NewCoin(sdk.NativeToken, openFee)
}

func setAddressAndPubkeyToCU(ctx sdk.Context, cu exported.CustodianUnit, keeper Keeper,
	pubkey []byte, address string, symbol, chain string, epoch uint64) sdk.Result {
	if err := cu.SetAssetAddress(symbol, address, epoch); err != nil {
		return sdk.ErrInternal(fmt.Sprintf("Set asset address error: %v", err)).Result()
	}
	if symbol != chain {
		if err := cu.SetAssetAddress(chain, address, epoch); err != nil {
			return sdk.ErrInternal(fmt.Sprintf("Set chain asset address error: %v", err)).Result()
		}
	}

	if err := cu.SetAssetPubkey(pubkey, epoch); err != nil {
		return sdk.ErrInternal(fmt.Sprintf("Set asset public key error: %v", err)).Result()
	}
	keeper.ck.SetExtAddresseWithCU(ctx, chain, address, cu.GetAddress())
	keeper.ck.SetCU(ctx, cu)

	return sdk.Result{Code: sdk.CodeOK}
}

//Handle
func handleMsgNewOpCU(ctx sdk.Context, keeper Keeper, msg MsgNewOpCU) sdk.Result {
	ctx.Logger().Info("handleMsgNewOpCU", "msg", msg)

	curEpoch := keeper.vk.GetCurrentEpoch(ctx)
	if !curEpoch.MigrationFinished {
		return sdk.ErrInvalidTx("Cannot new opcu in migration period").Result()
	}

	symbol := sdk.Symbol(msg.Symbol)
	if !keeper.tk.IsTokenSupported(ctx, symbol) {
		return sdk.ErrUnSupportToken(msg.Symbol).Result()
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
		return sdk.ErrInvalidAddr(fmt.Sprintf("from CU:%v is not a validator", msg.From)).Result()
	}
	// too many op cu of symbol
	var cuLimit uint64

	cuLimit = keeper.tk.GetMaxOpCUNumber(ctx, symbol)
	opcuCount := len(keeper.ck.GetOpCUs(ctx, msg.Symbol))
	if uint64(opcuCount) >= cuLimit {
		return sdk.ErrInternal(fmt.Sprintf("too many operation CU of %s", msg.Symbol)).Result()
	}

	// TODO add event ?
	opcu := keeper.ck.NewOpCUWithAddress(ctx, msg.Symbol, opcuAddress)
	if opcu == nil {
		return sdk.ErrInternal("create operation CU failed").Result()
	}
	keeper.ck.SetCU(ctx, opcu)

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
