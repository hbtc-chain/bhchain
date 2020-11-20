package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderKeygen(t *testing.T) {
	cuAddr, err := CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	assert.Nil(t, err)
	oneOrderBase := OrderBase{
		ID:        "UUID9",
		OrderType: OrderTypeKeyGen,
		Status:    OrderStatusBegin,
		Symbol:    "btc",
		CUAddress: cuAddr,
	}

	keyNodes := []CUAddress{NewCUAddress(), NewCUAddress()}
	to := NewCUAddress()
	firstOrder := OrderKeyGen{
		OrderBase:        oneOrderBase,
		KeyNodes:         keyNodes,
		SignThreshold:    uint64(3),
		To:               to,
		OpenFee:          NewCoin(NativeToken, NewInt(10)),
		MultiSignAddress: "0x12321cae2b",
	}

	orderBase := OrderBase{}
	secondOrder := OrderKeyGen{}
	secondOrder.OrderBase = orderBase
	secondOrder.SetOrderType(OrderTypeKeyGen)
	secondOrder.SetOrderStatus(OrderStatusBegin)
	secondOrder.SetID("UUID9")
	secondOrder.Symbol = "btc"
	secondOrder.CUAddress = cuAddr
	secondOrder.KeyNodes = keyNodes
	secondOrder.SignThreshold = uint64(3)
	secondOrder.To = to
	secondOrder.OpenFee = NewCoin(NativeToken, NewInt(10))
	secondOrder.MultiSignAddress = "0x12321cae2b"

	b := reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	copyOrder := firstOrder.DeepCopy()
	thirdOrder := (copyOrder).(*OrderKeyGen)
	b = reflect.DeepEqual(firstOrder, *thirdOrder)
	assert.True(t, b)

	// mod thirdOrder the firstOrder not changed
	thirdOrder.ID = "UUID10"
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
	thirdOrder.ID = "UUID9"

	thirdOrder.KeyNodes = []CUAddress{NewCUAddress(), NewCUAddress()}
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
	thirdOrder.KeyNodes = keyNodes

	thirdOrder.SignThreshold = uint64(7)
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
	thirdOrder.SignThreshold = uint64(3)

	thirdOrder.Symbol = "changed symbol"
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
	thirdOrder.Symbol = "btc"

	temp := thirdOrder.To
	thirdOrder.To = NewCUAddress()
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
	thirdOrder.To = temp
}

func TestOrderSysTransfer(t *testing.T) {

}

func TestOrderWithdrawal(t *testing.T) {
	cuAddr, err := CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	assert.Nil(t, err)
	oneOrderBase := OrderBase{
		ID:        "UUID9",
		OrderType: OrderTypeWithdrawal,
		Status:    OrderStatusBegin,
		Symbol:    "BHBTC",
		CUAddress: cuAddr,
	}

	firstOrder := OrderWithdrawal{
		OrderBase:         oneOrderBase,
		Amount:            NewInt(20),
		GasFee:            NewInt(3),
		WithdrawToAddress: "17tyo8fzvrkbSn4YoBn9TPeqqanMzfKbhy",
		OpCUaddress:       "opcuaddress1",
		Txhash:            "txhash1",
		RawData:           []byte("raw data"),
		SignedTx:          []byte("signed tx"),
	}

	orderBase := OrderBase{}
	secondOrder := OrderWithdrawal{}
	secondOrder.OrderBase = orderBase
	secondOrder.SetOrderType(OrderTypeWithdrawal)
	secondOrder.SetOrderStatus(OrderStatusBegin)
	secondOrder.SetID("UUID9")
	secondOrder.Symbol = "BHBTC"
	secondOrder.CUAddress = cuAddr
	secondOrder.Amount = NewInt(20)
	secondOrder.GasFee = NewInt(3)
	secondOrder.WithdrawToAddress = "17tyo8fzvrkbSn4YoBn9TPeqqanMzfKbhy"
	secondOrder.OpCUaddress = "opcuaddress1"
	secondOrder.Txhash = "txhash1"
	secondOrder.RawData = []byte("raw data")
	secondOrder.SignedTx = []byte("signed tx")

	b := reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	copyOrder := firstOrder.DeepCopy()
	thirdOrder := (copyOrder).(*OrderWithdrawal)
	b = reflect.DeepEqual(firstOrder, *thirdOrder)
	assert.True(t, b)

	// mod thirdOrder the firstOrder not changed
	thirdOrder.ID = "UUID10"
	thirdOrder.SignedTx = []byte("changed SignedTx")
	thirdOrder.Amount = NewInt(1)
	thirdOrder.Symbol = "changed symbol"
	thirdOrder.WithdrawToAddress = "changed WithdrawToAddress"
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
}

func TestOrderCollect(t *testing.T) {
	cuAddr, err := CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	assert.Nil(t, err)
	cuAddr1, err := CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	assert.Nil(t, err)
	oneOrderBase := OrderBase{
		ID:        "UUID9",
		OrderType: OrderTypeCollect,
		Status:    OrderStatusBegin,
		Symbol:    "BHBTC",
		CUAddress: cuAddr,
	}

	firstOrder := OrderCollect{
		OrderBase:          oneOrderBase,
		CollectFromCU:      cuAddr,
		CollectFromAddress: "38apd1KQmQcf3o2rkQD21WApSJAtMpdP9H",
		CollectToCU:        cuAddr1,
		Amount:             NewInt(20),
		GasPrice:           NewInt(2),
		GasLimit:           NewInt(1),
		Txhash:             "txhash1",
		RawData:            []byte("Raw data"),
		SignedTx:           []byte("Signed Tx"),
	}

	orderBase := OrderBase{}
	secondOrder := OrderCollect{}
	secondOrder.OrderBase = orderBase
	secondOrder.SetOrderType(OrderTypeCollect)
	secondOrder.SetOrderStatus(OrderStatusBegin)
	secondOrder.CUAddress = cuAddr
	secondOrder.SetID("UUID9")
	secondOrder.Symbol = "BHBTC"
	secondOrder.CollectFromCU = cuAddr
	secondOrder.CollectFromAddress = "38apd1KQmQcf3o2rkQD21WApSJAtMpdP9H"
	secondOrder.CollectToCU = cuAddr1
	secondOrder.Amount = NewInt(20)
	secondOrder.GasPrice = NewInt(2)
	secondOrder.GasLimit = NewInt(1)
	secondOrder.Txhash = "txhash1"
	secondOrder.RawData = []byte("Raw data")
	secondOrder.SignedTx = []byte("Signed Tx")

	b := reflect.DeepEqual(firstOrder, secondOrder)
	//	assert.True(t, b)
	copyOrder := firstOrder.DeepCopy()
	thirdOrder := (copyOrder).(*OrderCollect)
	b = reflect.DeepEqual(firstOrder, *thirdOrder)
	assert.True(t, b)

	// mod thirdOrder the firstOrder not changed
	thirdOrder.ID = "UUID10"
	thirdOrder.SignedTx = []byte("changed SignedTx")

	thirdOrder.Amount = NewInt(1)
	thirdOrder.Symbol = "changed symbol"
	thirdOrder.CollectFromAddress = "changed CollectFromAddress"
	b = reflect.DeepEqual(firstOrder, secondOrder)
	assert.True(t, b)
	b = reflect.DeepEqual(firstOrder, thirdOrder)
	assert.False(t, b)
}
