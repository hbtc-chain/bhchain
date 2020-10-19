package mapping

import (
	"bytes"
	"fmt"

	"github.com/hbtc-chain/bhchain/x/mapping/types"
	"github.com/hbtc-chain/bhchain/x/params"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/mapping/internal"
)

var (
	mappingStoreKeyPrefix        = []byte{0x01}
	targetSymbolsStoreKeyPrefix  = []byte{0x02}
	freeSwapInfoStoreKeyPrefix   = []byte{0x03}
	directSwapInfoStoreKeyPrefix = []byte{0x04}
	swapPoolKey                  = []byte("swap_pool_key")
)

// Keeper for mapping module
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramstore params.Subspace

	tk internal.TokenKeeper
	ck internal.CUKeeper
	rk internal.ReceiptKeeper
}

func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, tk internal.TokenKeeper, ck internal.CUKeeper,
	rk internal.ReceiptKeeper, paramstore params.Subspace) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		rk:         rk,
		ck:         ck,
		tk:         tk,
		paramstore: paramstore.WithKeyTable(ParamKeyTable()),
	}
}

func mappingStoreKey(symbol sdk.Symbol) []byte {
	return append(mappingStoreKeyPrefix, []byte(symbol.String())...)
}

func targetSymbolStoreKey(targetSymbol sdk.Symbol) []byte {
	return append(targetSymbolsStoreKeyPrefix, []byte(targetSymbol.String())...)
}

func freeSwapOrderStoreKey(orderID string) []byte {
	return append(freeSwapInfoStoreKeyPrefix, []byte(orderID)...)
}

func directSwapOrderStoreKey(orderID string) []byte {
	return append(directSwapInfoStoreKeyPrefix, []byte(orderID)...)
}

//Set the entire MappingInfo
func (k Keeper) SetMappingInfo(ctx sdk.Context, mappingInfo *MappingInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(mappingStoreKey(mappingInfo.IssueSymbol), k.cdc.MustMarshalBinaryBare(mappingInfo))
	targetSymbolKey := targetSymbolStoreKey(mappingInfo.TargetSymbol)
	if !store.Has(targetSymbolStoreKey(mappingInfo.TargetSymbol)) {
		store.Set(targetSymbolKey, []byte{})
	}
}

func (k Keeper) GetMappingInfo(ctx sdk.Context, issueSymbol sdk.Symbol) *MappingInfo {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(mappingStoreKey(issueSymbol)) {
		return nil
	}

	bz := store.Get(mappingStoreKey(issueSymbol))
	var mappingInfo MappingInfo
	k.cdc.MustUnmarshalBinaryBare(bz, &mappingInfo)
	return &mappingInfo
}

func (k Keeper) HasTargetSymbol(ctx sdk.Context, targetSymbol sdk.Symbol) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(targetSymbolStoreKey(targetSymbol))
}

func (k Keeper) GetIssueSymbolIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, mappingStoreKeyPrefix)
}

func (k Keeper) GetIssueSymbols(ctx sdk.Context) []sdk.Symbol {
	var symbols []sdk.Symbol
	iter := k.GetIssueSymbolIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		symbols = append(symbols, sdk.Symbol(string(bytes.TrimPrefix(iter.Key(), mappingStoreKeyPrefix))))
	}
	return symbols
}

func (k Keeper) GetFreeSwapOrderIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, freeSwapInfoStoreKeyPrefix)
}

func (k Keeper) GetDirectSwapOrderIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, directSwapInfoStoreKeyPrefix)
}

func (k Keeper) IsSwapOrderExist(ctx sdk.Context, OrderID string, swapType int) bool {
	store := ctx.KVStore(k.storeKey)
	if swapType == types.SwapTypeFree {
		return store.Has(freeSwapOrderStoreKey(OrderID))
	} else {
		return store.Has(directSwapOrderStoreKey(OrderID))
	}
}

func (k Keeper) CreateFreeSwapOrder(ctx sdk.Context, Owner sdk.CUAddress, swapInfo FreeSwapInfo, orderID string) sdk.Result {
	fromCU := k.ck.GetCU(ctx, Owner)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(swapPoolKey)
	var swapPool SwapPool
	if bz != nil {
		err := k.cdc.UnmarshalBinaryBare(bz, &swapPool)
		if err != nil {
			return sdk.ErrInvalidTx("UnmarshalBinaryBare swap pool err").Result()
		}
	} else {
		swapPool.SwapCoins = sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt()))
	}

	needCoin := sdk.NewCoins(sdk.NewCoin(swapInfo.SrcSymbol.String(), swapInfo.TotalAmount))
	if fromCU.GetCoins().AmountOf(swapInfo.SrcSymbol.String()).LT(swapInfo.TotalAmount) {
		return sdk.ErrInvalidTx(fmt.Sprintf("swap token not enough, need:%v, have:%v", needCoin, fromCU.GetCoins())).Result()
	}

	if swapInfo.MaxSwapAmount.Equal(sdk.ZeroInt()) {
		swapInfo.MaxSwapAmount = swapInfo.TotalAmount
	}

	fromCU.SubCoins(needCoin)
	swapPool.SwapCoins = swapPool.SwapCoins.Add(needCoin)
	freeSwapOrder := FreeSwapOrder{
		OrderId:      orderID,
		Owner:        Owner,
		SwapInfo:     swapInfo,
		RemainAmount: swapInfo.TotalAmount,
	}

	store.Set(freeSwapOrderStoreKey(orderID), k.cdc.MustMarshalBinaryBare(freeSwapOrder))
	store.Set(swapPoolKey, k.cdc.MustMarshalBinaryBare(swapPool))

	k.ck.SetCU(ctx, fromCU)

	var flows []sdk.Flow
	for _, balanceFlow := range fromCU.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}

	fromCU.ResetBalanceFlows()

	receipt := k.rk.NewReceipt(sdk.CategoryTypeQuickSwap, flows)
	res := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &res)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateFreeSwap,
			sdk.NewAttribute(types.AttributeKeyFrom, Owner.String()),
			sdk.NewAttribute(types.AttributeKeyIssueToken, swapInfo.SrcSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyTargetToken, swapInfo.TargetSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, orderID),
			sdk.NewAttribute(types.AttributeKeyAmount, swapInfo.TotalAmount.String()),
		),
	)

	res.Events = append(res.Events, ctx.EventManager().Events()...)

	return res
}

func (k Keeper) CreateDirectSwapOrder(ctx sdk.Context, Owner sdk.CUAddress, swapInfo DirectSwapInfo, orderID string) sdk.Result {
	fromCU := k.ck.GetCU(ctx, Owner)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(swapPoolKey)
	var swapPool SwapPool
	if bz != nil {
		err := k.cdc.UnmarshalBinaryBare(bz, &swapPool)
		if err != nil {
			return sdk.ErrInvalidTx("UnmarshalBinaryBare swap pool err").Result()
		}
	} else {
		swapPool.SwapCoins = sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.ZeroInt()))
	}

	needCoin := sdk.NewCoins(sdk.NewCoin(swapInfo.SrcSymbol.String(), swapInfo.Amount))
	if fromCU.GetCoins().AmountOf(swapInfo.SrcSymbol.String()).LT(swapInfo.Amount) {
		return sdk.ErrInvalidTx(fmt.Sprintf("swap token not enough, need:%v, have:%v", needCoin, fromCU.GetCoins())).Result()
	}

	fromCU.SubCoins(needCoin)
	swapPool.SwapCoins = swapPool.SwapCoins.Add(needCoin)
	directSwapOrder := DirectSwapOrder{
		OrderId:  orderID,
		Owner:    Owner,
		SwapInfo: swapInfo,
	}

	store.Set(directSwapOrderStoreKey(orderID), k.cdc.MustMarshalBinaryBare(directSwapOrder))
	store.Set(swapPoolKey, k.cdc.MustMarshalBinaryBare(swapPool))

	k.ck.SetCU(ctx, fromCU)

	var flows []sdk.Flow
	for _, balanceFlow := range fromCU.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}

	fromCU.ResetBalanceFlows()
	receipt := k.rk.NewReceipt(sdk.CategoryTypeQuickSwap, flows)
	res := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &res)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateDirectSwap,
			sdk.NewAttribute(types.AttributeKeyFrom, Owner.String()),
			sdk.NewAttribute(types.AttributeKeyIssueToken, swapInfo.SrcSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyTargetToken, swapInfo.TargetSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, orderID),
			sdk.NewAttribute(types.AttributeKeyAmount, swapInfo.Amount.String()),
		),
	)

	res.Events = append(res.Events, ctx.EventManager().Events()...)

	return res
}

func (k Keeper) SwapSymbol(ctx sdk.Context, fromCUAddr sdk.CUAddress, swapType int, orderID string, swapAmount sdk.Int) sdk.Result {
	fromCU := k.ck.GetCU(ctx, fromCUAddr)
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(swapPoolKey)
	var swapPool SwapPool
	if bz != nil {
		err := k.cdc.UnmarshalBinaryBare(bz, &swapPool)
		if err != nil {
			return sdk.ErrInvalidTx("UnmarshalBinaryBare swap pool err").Result()
		}
	} else {
		return sdk.ErrInvalidTx("Get Swap Pool Err").Result()
	}

	var flows []sdk.Flow
	if swapType == types.SwapTypeFree {
		var order FreeSwapOrder
		bz = store.Get(freeSwapOrderStoreKey(orderID))
		if bz == nil {
			return sdk.ErrInvalidOrder(fmt.Sprintf("swap order not exitst:%v", orderID)).Result()
		}

		k.cdc.UnmarshalBinaryBare(bz, &order)
		if order.SwapInfo.ExpiredTime != 0 && ctx.BlockTime().Unix() > order.SwapInfo.ExpiredTime {
			return sdk.ErrInvalidTx("swap time is expired").Result()
		}

		if swapAmount.LT(order.SwapInfo.MinSwapAmount) || swapAmount.GT(order.SwapInfo.MaxSwapAmount) {
			return sdk.ErrInvalidTx("swap amount not fit").Result()
		}

		if order.RemainAmount.LT(swapAmount) {
			return sdk.ErrInvalidAmount(fmt.Sprintf("reamain coin amount not enough:%v, %v", order.RemainAmount, swapAmount)).Result()
		}

		tokenInfo := k.tk.GetTokenInfo(ctx, order.SwapInfo.SrcSymbol)
		needCoinAmt := sdk.NewDecFromInt(swapAmount.Mul(order.SwapInfo.SwapPrice).Quo(sdk.NewIntWithDecimal(1, int(tokenInfo.Decimals)))).TruncateInt()
		needCoin := sdk.NewCoins(sdk.NewCoin(order.SwapInfo.TargetSymbol.String(), needCoinAmt))
		if fromCU.GetCoins().AmountOf(order.SwapInfo.TargetSymbol.String()).LT(needCoinAmt) {
			return sdk.ErrInvalidAmount(fmt.Sprintf("swap coin not enough:%v", swapAmount)).Result()
		}

		swapCoin := sdk.NewCoins(sdk.NewCoin(order.SwapInfo.SrcSymbol.String(), swapAmount))
		toCU := k.ck.GetCU(ctx, order.Owner)
		fromCU.SubCoins(needCoin)
		fromCU.AddCoins(swapCoin)
		swapPool.SwapCoins = swapPool.SwapCoins.Sub(swapCoin)
		toCU.AddCoins(needCoin)
		order.RemainAmount = order.RemainAmount.Sub(swapAmount)
		if order.RemainAmount.Equal(sdk.ZeroInt()) {
			store.Delete(freeSwapOrderStoreKey(orderID))
		} else {
			store.Set(freeSwapOrderStoreKey(orderID), k.cdc.MustMarshalBinaryBare(order))
		}

		k.ck.SetCU(ctx, fromCU)
		k.ck.SetCU(ctx, toCU)
		if swapPool.SwapCoins.Empty() {
			store.Delete(swapPoolKey)
		} else {
			store.Set(swapPoolKey, k.cdc.MustMarshalBinaryBare(swapPool))
		}

		for _, balanceFlow := range toCU.GetBalanceFlows() {
			flows = append(flows, balanceFlow)
		}

		toCU.ResetBalanceFlows()

	} else {
		var order DirectSwapOrder
		bz = store.Get(directSwapOrderStoreKey(orderID))
		if bz == nil {
			return sdk.ErrInvalidOrder(fmt.Sprintf("swap order not exitst:%v", orderID)).Result()
		}

		k.cdc.UnmarshalBinaryBare(bz, &order)
		if order.SwapInfo.ExpiredTime != 0 && ctx.BlockTime().Unix() > order.SwapInfo.ExpiredTime {
			return sdk.ErrInvalidTx("swap time is expired").Result()
		}

		if order.SwapInfo.ReceiveAddr != fromCUAddr.String() {
			return sdk.ErrInvalidAddr(fmt.Sprintf("swap addr is not expected:%v, %v", order.SwapInfo.ReceiveAddr, fromCUAddr.String())).Result()
		}

		needCoin := sdk.NewCoins(sdk.NewCoin(order.SwapInfo.TargetSymbol.String(), order.SwapInfo.SwapAmount))
		if fromCU.GetCoins().AmountOf(order.SwapInfo.TargetSymbol.String()).LT(order.SwapInfo.SwapAmount) {
			return sdk.ErrInvalidAmount(fmt.Sprintf("swap coin not enough:%v", swapAmount)).Result()
		}

		swapCoin := sdk.NewCoins(sdk.NewCoin(order.SwapInfo.SrcSymbol.String(), order.SwapInfo.Amount))
		toCU := k.ck.GetCU(ctx, order.Owner)
		fromCU.SubCoins(needCoin)
		fromCU.AddCoins(swapCoin)
		swapPool.SwapCoins = swapPool.SwapCoins.Sub(swapCoin)
		toCU.AddCoins(needCoin)

		k.ck.SetCU(ctx, fromCU)
		k.ck.SetCU(ctx, toCU)
		store.Delete(directSwapOrderStoreKey(orderID))

		if swapPool.SwapCoins.Empty() {
			store.Delete(swapPoolKey)
		} else {
			store.Set(swapPoolKey, k.cdc.MustMarshalBinaryBare(swapPool))
		}

		for _, balanceFlow := range toCU.GetBalanceFlows() {
			flows = append(flows, balanceFlow)
		}

		toCU.ResetBalanceFlows()
	}

	for _, balanceFlow := range fromCU.GetBalanceFlows() {
		flows = append(flows, balanceFlow)
	}

	fromCU.ResetBalanceFlows()

	receipt := k.rk.NewReceipt(sdk.CategoryTypeQuickSwap, flows)
	res := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &res)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapSymbol,
			sdk.NewAttribute(types.AttributeKeyFrom, fromCUAddr.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, orderID),
			sdk.NewAttribute(types.AttributeKeySwapType, string(swapType)),
			sdk.NewAttribute(types.AttributeKeyAmount, swapAmount.String()),
		),
	)

	res.Events = append(res.Events, ctx.EventManager().Events()...)

	return res
}

func (k Keeper) CancelSwap(ctx sdk.Context, fromCUAddr sdk.CUAddress, swapType int, orderID string) sdk.Result {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(swapPoolKey)
	var swapPool SwapPool
	if bz != nil {
		err := k.cdc.UnmarshalBinaryBare(bz, &swapPool)
		if err != nil {
			return sdk.ErrInvalidTx("UnmarshalBinaryBare swap pool err").Result()
		}
	} else {
		return sdk.ErrInvalidTx("Get Swap Pool Err").Result()
	}

	var flows []sdk.Flow
	if swapType == types.SwapTypeFree {
		var order FreeSwapOrder
		bz = store.Get(freeSwapOrderStoreKey(orderID))
		if bz == nil {
			return sdk.ErrInvalidOrder(fmt.Sprintf("swap order not exitst:%v", orderID)).Result()
		}

		k.cdc.UnmarshalBinaryBare(bz, &order)

		if order.Owner.String() != fromCUAddr.String() {
			return sdk.ErrInvalidAddr(fmt.Sprintf("swap order addr err:%v, %v", order.Owner, fromCUAddr)).Result()
		}

		if order.RemainAmount.GT(sdk.ZeroInt()) {
			toCU := k.ck.GetCU(ctx, order.Owner)
			cancelCoins := sdk.NewCoins(sdk.NewCoin(order.SwapInfo.SrcSymbol.String(), order.RemainAmount))
			toCU.AddCoins(cancelCoins)
			swapPool.SwapCoins = swapPool.SwapCoins.Sub(cancelCoins)
			k.ck.SetCU(ctx, toCU)
			if swapPool.SwapCoins.Empty() {
				store.Delete(swapPoolKey)
			} else {
				store.Set(swapPoolKey, k.cdc.MustMarshalBinaryBare(swapPool))
			}
			for _, balanceFlow := range toCU.GetBalanceFlows() {
				flows = append(flows, balanceFlow)
			}

			toCU.ResetBalanceFlows()
		}

		store.Delete(freeSwapOrderStoreKey(orderID))
	} else {
		var order DirectSwapOrder
		bz = store.Get(directSwapOrderStoreKey(orderID))
		if bz == nil {
			return sdk.ErrInvalidOrder(fmt.Sprintf("swap order not exitst:%v", orderID)).Result()
		}

		k.cdc.UnmarshalBinaryBare(bz, &order)
		if order.Owner.String() != fromCUAddr.String() {
			return sdk.ErrInvalidAddr(fmt.Sprintf("swap order addr err:%v, %v", order.Owner, fromCUAddr)).Result()
		}

		toCU := k.ck.GetCU(ctx, order.Owner)
		cancelCoins := sdk.NewCoins(sdk.NewCoin(order.SwapInfo.SrcSymbol.String(), order.SwapInfo.Amount))
		toCU.AddCoins(cancelCoins)
		swapPool.SwapCoins = swapPool.SwapCoins.Sub(cancelCoins)
		if swapPool.SwapCoins.Empty() {
			store.Delete(swapPoolKey)
		} else {
			store.Set(swapPoolKey, k.cdc.MustMarshalBinaryBare(swapPool))
		}
		k.ck.SetCU(ctx, toCU)

		store.Delete(directSwapOrderStoreKey(orderID))
		for _, balanceFlow := range toCU.GetBalanceFlows() {
			flows = append(flows, balanceFlow)
		}

		toCU.ResetBalanceFlows()
	}

	receipt := k.rk.NewReceipt(sdk.CategoryTypeQuickSwap, flows)
	res := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &res)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCancelSwap,
			sdk.NewAttribute(types.AttributeKeyFrom, fromCUAddr.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, orderID),
			sdk.NewAttribute(types.AttributeKeySwapType, string(swapType)),
		),
	)

	res.Events = append(res.Events, ctx.EventManager().Events()...)

	return res
}

func (k Keeper) GetFreeSwapOrder(ctx sdk.Context, orderID string) *FreeSwapOrder {
	var order FreeSwapOrder
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(freeSwapOrderStoreKey(orderID))
	if bz == nil {
		return nil
	}
	k.cdc.UnmarshalBinaryBare(bz, &order)
	return &order
}

func (k Keeper) GetFreeSwapOrders(ctx sdk.Context) []FreeSwapOrder {
	var orders []FreeSwapOrder
	iter := k.GetFreeSwapOrderIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var order FreeSwapOrder
		k.cdc.UnmarshalBinaryBare(iter.Value(), &order)
		orders = append(orders, order)
	}
	return orders
}

func (k Keeper) GetDirectSwapOrder(ctx sdk.Context, orderID string) *DirectSwapOrder {
	var order DirectSwapOrder
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(directSwapOrderStoreKey(orderID))
	if bz == nil {
		return nil
	}
	k.cdc.UnmarshalBinaryBare(bz, &order)
	return &order
}

func (k Keeper) GetDirectSwapOrders(ctx sdk.Context) []DirectSwapOrder {
	var orders []DirectSwapOrder
	iter := k.GetDirectSwapOrderIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var order DirectSwapOrder
		k.cdc.UnmarshalBinaryBare(iter.Value(), &order)
		orders = append(orders, order)
	}
	return orders
}
