package order

import (
	"encoding/binary"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/order/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

type Keeper struct {
	key sdk.StoreKey

	// The codec codec for binary encoding/decoding of CustodianUnits.
	cdc *codec.Codec

	paramSubspace params.Subspace
}

func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace,
) Keeper {
	return Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
	}
}

func (k *Keeper) NewOrder(ctx sdk.Context, order sdk.Order) sdk.Order {
	if order == nil || order.GetCUAddress() == nil {
		return nil
	}
	if k.IsExist(ctx, order.GetID()) {
		return nil
	}
	order.SetOrderStatus(sdk.OrderStatusBegin)
	k.AddProcessOrder(ctx, order)
	return order
}

func (k *Keeper) AddProcessOrder(ctx sdk.Context, order sdk.Order) {
	store := ctx.KVStore(k.key)
	store.Set(processOrderKey(order.GetOrderType(), order.GetID()), []byte{0})
}

func (k *Keeper) RemoveProcessOrder(ctx sdk.Context, orderType sdk.OrderType, orderID string) {
	store := ctx.KVStore(k.key)
	store.Delete(processOrderKey(orderType, orderID))
}

// GetProcessOrderList get processing order of all type.
// Consider use GetProcessOrderListByType to narrow down the returned order to save gas
func (k *Keeper) GetProcessOrderList(ctx sdk.Context) []string {
	processOrderList := []string{}
	store := ctx.KVStore(k.key)
	iterator := sdk.KVStorePrefixIterator(store, types.ProcessOrderStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Key()
		_, orderID := decodeProcessOrderKey(bz)
		processOrderList = append(processOrderList, orderID)
	}

	return processOrderList
}

func (k *Keeper) GetProcessOrderListByType(ctx sdk.Context, orderTypes ...sdk.OrderType) []string {
	processOrderList := []string{}
	for _, orderType := range orderTypes {
		processOrderList = append(processOrderList, k.getProcessOrderListBySingleType(ctx, orderType)...)
	}

	return processOrderList
}

func (k *Keeper) getProcessOrderListBySingleType(ctx sdk.Context, orderType sdk.OrderType) []string {
	processOrderList := []string{}
	store := ctx.KVStore(k.key)
	iterator := sdk.KVStorePrefixIterator(store, processOrderPrefixKey(orderType))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Key()
		_, orderID := decodeProcessOrderKey(bz)
		processOrderList = append(processOrderList, orderID)
	}
	return processOrderList
}

func (k *Keeper) NewOrderKeyGen(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
	keyNodes []sdk.CUAddress, signThreshold uint64, to sdk.CUAddress, openFee sdk.Coins) *sdk.OrderKeyGen {
	ordBase := sdk.OrderBase{
		CUAddress: from,
		ID:        orderID,
		OrderType: sdk.OrderTypeKeyGen,
		Symbol:    symbol,
		Height:    uint64(ctx.BlockHeight()),
	}
	orderKeyGen := sdk.OrderKeyGen{
		OrderBase:        ordBase,
		KeyNodes:         keyNodes,
		SignThreshold:    signThreshold,
		To:               to,
		MultiSignAddress: "",
		OpenFee:          openFee,
	}
	order := k.NewOrder(ctx, &orderKeyGen)
	if order == nil {
		return &sdk.OrderKeyGen{}
	}
	return (order).(*sdk.OrderKeyGen)
}

func (k *Keeper) NewOrderCollect(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
	collectFromCU sdk.CUAddress, collectFromAddress string, amount, gasPrice, gasLimit sdk.Int, txHash string, index uint64, memo string) *sdk.OrderCollect {
	ordBase := sdk.OrderBase{
		CUAddress: from,
		ID:        orderID,
		OrderType: sdk.OrderTypeCollect,
		Symbol:    symbol,
		Height:    uint64(ctx.BlockHeight()),
	}

	neworder := sdk.OrderCollect{
		OrderBase:          ordBase,
		CollectFromCU:      collectFromCU,
		CollectFromAddress: collectFromAddress,
		Amount:             amount,
		GasPrice:           gasPrice,
		GasLimit:           gasLimit,
		CostFee:            sdk.ZeroInt(),
		Txhash:             txHash,
		Index:              index,
		Memo:               memo,
		DepositStatus:      sdk.DepositUnconfirm,
	}
	order := k.NewOrder(ctx, &neworder)
	if order == nil {
		return &sdk.OrderCollect{}
	}
	return (order).(*sdk.OrderCollect)
}

func (k *Keeper) NewOrderWithdrawal(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
	amount, gasFee, costFee sdk.Int, withdrawToAddr, opCUAddr, txHash string) *sdk.OrderWithdrawal {
	ordBase := sdk.OrderBase{
		CUAddress: from,
		ID:        orderID,
		OrderType: sdk.OrderTypeWithdrawal,
		Symbol:    symbol,
		Height:    uint64(ctx.BlockHeight()),
	}
	neworder := sdk.OrderWithdrawal{
		OrderBase:         ordBase,
		Amount:            amount,
		GasFee:            gasFee,
		CostFee:           costFee,
		WithdrawToAddress: withdrawToAddr,
		OpCUaddress:       opCUAddr,
		Txhash:            txHash,
		UtxoInNum:         0,
	}
	order := k.NewOrder(ctx, &neworder)
	if order == nil {
		return &sdk.OrderWithdrawal{}
	}
	return (order).(*sdk.OrderWithdrawal)
}

func (k *Keeper) NewOrderSysTransfer(ctx sdk.Context, from sdk.CUAddress, orderID string, symbol string,
	amount, costFee sdk.Int, toCU, toAddr, opCUAddr, fromAddr string) *sdk.OrderSysTransfer {
	ordBase := sdk.OrderBase{
		CUAddress: from,
		ID:        orderID,
		OrderType: sdk.OrderTypeSysTransfer,
		Symbol:    symbol,
		Height:    uint64(ctx.BlockHeight()),
	}
	neworder := sdk.OrderSysTransfer{
		OrderBase:   ordBase,
		Amount:      amount,
		CostFee:     costFee,
		ToCU:        toCU,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		OpCUaddress: opCUAddr,
	}
	order := k.NewOrder(ctx, &neworder)
	if order == nil {
		return &sdk.OrderSysTransfer{}
	}
	return (order).(*sdk.OrderSysTransfer)
}

func (k *Keeper) NewOrderOpcuAssetTransfer(ctx sdk.Context, opcu sdk.CUAddress, orderID string, symbol string,
	items []sdk.TransferItem, toAddr string) *sdk.OrderOpcuAssetTransfer {
	ordBase := sdk.OrderBase{
		CUAddress: opcu,
		ID:        orderID,
		OrderType: sdk.OrderTypeOpcuAssetTransfer,
		Symbol:    symbol,
		Height:    uint64(ctx.BlockHeight()),
	}
	neworder := sdk.OrderOpcuAssetTransfer{
		OrderBase:      ordBase,
		TransfertItems: make([]sdk.TransferItem, len(items)),
		ToAddr:         toAddr,
	}
	copy(neworder.TransfertItems, items)
	order := k.NewOrder(ctx, &neworder)
	if order == nil {
		return &sdk.OrderOpcuAssetTransfer{}
	}
	return (order).(*sdk.OrderOpcuAssetTransfer)
}

func (k *Keeper) GetOrder(ctx sdk.Context, orderID string) sdk.Order {
	store := ctx.KVStore(k.key)
	bz := store.Get(orderKey(orderID))
	if bz == nil {
		return nil
	}
	var o sdk.Order
	k.cdc.MustUnmarshalBinaryBare(bz, &o)

	return o
}

func (k *Keeper) GetOrderByStatus(ctx sdk.Context, orderID string, status sdk.OrderStatus) []sdk.Order {
	orders := []sdk.Order{}
	f := func(o sdk.Order) (stop bool) {
		if o.GetOrderStatus() == status {
			orders = append(orders, o)
		}
		return false
	}
	k.iterateOrders(ctx, []byte(""), f)
	return orders
}

func (k *Keeper) SetOrder(ctx sdk.Context, order sdk.Order) {
	if order == nil || !order.GetCUAddress().IsValidAddr() {
		return
	}

	store := ctx.KVStore(k.key)
	bz := k.cdc.MustMarshalBinaryBare(order)

	store.Set(orderKey(order.GetID()), bz)

	if order.GetOrderStatus().Terminated() {
		k.RemoveProcessOrder(ctx, order.GetOrderType(), order.GetID())
	}
}

func (k *Keeper) DeleteOrder(ctx sdk.Context, order sdk.Order) {
	if order == nil || !order.GetCUAddress().IsValidAddr() {
		return
	}
	store := ctx.KVStore(k.key)
	store.Delete(orderKey(order.GetID()))
}

func (k *Keeper) iterateOrders(ctx sdk.Context, prefix []byte, process func(order sdk.Order) (stop bool)) {
	store := ctx.KVStore(k.key)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		as := k.decodeOrder(val)
		if process(as) {
			return
		}
		iter.Next()
	}
}

func (k *Keeper) decodeOrder(bz []byte) (order sdk.Order) {
	k.cdc.MustUnmarshalBinaryBare(bz, &order)
	return
}

// GetNextOrderNumber Returns and increments the global order number counter
func (k *Keeper) IsExist(ctx sdk.Context, uuid string) bool {
	store := ctx.KVStore(k.key)
	return store.Has(orderKey(uuid))
}

func orderKey(uuid string) []byte {
	// OrderNumber Key :  OrderStoreKeyPrefix + OrderIDKey
	k := append(types.OrderStoreKeyPrefix, []byte(uuid)...)
	return k
}

// prefix + orderType + orderID
func processOrderKey(orderType sdk.OrderType, orderID string) []byte {
	key := processOrderPrefixKey(orderType)
	key = append(key, []byte(orderID)...)
	return key
}

func processOrderPrefixKey(orderType sdk.OrderType) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(orderType))
	key := append(types.ProcessOrderStoreKeyPrefix, buf...)
	return key
}

func decodeProcessOrderKey(bz []byte) (orderType sdk.OrderType, orderID string) {
	prefixLen := len(types.ProcessOrderStoreKeyPrefix)
	orderType = sdk.OrderType(binary.BigEndian.Uint16(bz[prefixLen : prefixLen+2]))
	orderID = string(bz[prefixLen+2:])
	return
}
