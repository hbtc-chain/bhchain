package receipt

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/tendermint/tendermint/crypto"
	"testing"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	//dbm "github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tm-db"
)

func TestSDKResult(t *testing.T) {
	_, _, cuAddr := testGenKey()

	input := setupTestInput()
	result := sdk.Result{}
	flows := []Flow{
		input.rek.NewBalanceFlow(cuAddr, sdk.NativeToken, "", sdk.NewInt(10), sdk.NewInt(10), sdk.ZeroInt(), sdk.ZeroInt()),
	}
	receipt := input.rek.NewReceipt(CategoryTypeTransfer, flows)
	result = *input.rek.SaveReceiptToResult(receipt, &result)

	receipt2, err := input.rek.GetReceiptFromResult(&result)
	assert.NoError(t, err)
	assert.Equal(t, receipt, receipt2)
}

func TestKeeperGetSet(t *testing.T) {
	input := setupTestInput()
	var data []byte

	// StatusFailed
	result0 := testMsgHandler(CategoryTypeTransfer)
	_, err := input.rek.GetReceiptFromResult(result0)
	assert.Nil(t, err)

	// success
	result2 := testMsgHandler(CategoryTypeKeyGen)
	receipt2, err := input.rek.GetReceiptFromResult(result2)
	assert.Nil(t, err)
	data = append(data, result2.Data...)
	var receiptTemp sdk.Receipt
	input.cdc.UnmarshalBinaryLengthPrefixed(data, &receiptTemp)
	assert.Equal(t, receiptTemp, *receipt2)

}

// test negative sdk.Int encode & decode by amino
func TestNegativeIntDecode(t *testing.T) {
	input := setupTestInput()
	result1 := testMsgHandler(CategoryTypeWithdrawal)
	receipt1, err := input.rek.GetReceiptFromResult(result1)
	assert.Nil(t, err)
	balFlow := (receipt1.Flows[0]).(BalanceFlow)
	bal, _ := sdk.NewIntFromString("21888242871839275222246405745257275088548364400416034343698204186575808495617")
	balWithdrawal, b := sdk.NewIntFromString("-21888242871839275222246405745257275088548364400416034343698204186575808495617")
	assert.True(t, b)
	assert.Equal(t, bal, balFlow.PreviousBalance)
	assert.True(t, balFlow.BalanceChange.IsNegative())
	assert.Equal(t, balWithdrawal, balFlow.BalanceChange)
}

func testMsgHandler(categoryType CategoryType) *Result {
	input := setupTestInput()

	switch categoryType {
	case CategoryTypeTransfer:
		ftl1 := []Flow{}
		ftl := input.rek.NewBalanceFlow(input.cuaddress, "btc", "", sdk.NewInt(100), sdk.NewInt(10),
			sdk.NewInt(0), sdk.NewInt(10))
		ftl1 = append(ftl1, ftl)
		receipt1 := input.rek.NewReceipt(CategoryTypeTransfer, ftl1)

		return input.rek.SaveReceiptToResult(receipt1, &Result{})
	case CategoryTypeKeyGen: //TODO(liyong): to add more order status after order module ready
		fol1 := []Flow{
			OrderFlow{
				CUAddress:   input.cuaddress,
				OrderID:     "uuidtest", //FIXME(liyong.zhang): call order uuid function
				OrderType:   sdk.OrderTypeKeyGen,
				OrderStatus: sdk.OrderStatusBegin,
			},
			OrderFlow{
				CUAddress:   input.cuaddress,
				OrderID:     "uuidtest1", //FIXME(liyong.zhang): call order uuid function
				OrderType:   sdk.OrderTypeKeyGen,
				OrderStatus: sdk.OrderStatusFinish,
			},
			BalanceFlow{
				CUAddress:             input.cuaddress,
				Symbol:                "btc",
				PreviousBalance:       sdk.NewInt(100),
				BalanceChange:         sdk.NewInt(10),
				PreviousBalanceOnHold: sdk.NewInt(0),
				BalanceOnHoldChange:   sdk.NewInt(10),
			},
		}
		receipt1 := input.rek.NewReceipt(CategoryTypeKeyGen, fol1)
		return input.rek.SaveReceiptToResult(receipt1, &sdk.Result{})
	case CategoryTypeDeposit:
		ftl := []Flow{}
		ftl1 := input.rek.NewBalanceFlow(input.cuaddress, "btc", "", sdk.NewInt(100), sdk.NewInt(10),
			sdk.NewInt(0), sdk.NewInt(10))
		ftl = append(ftl, ftl1)
		receipt1 := input.rek.NewReceipt(CategoryTypeDeposit, ftl)
		return input.rek.SaveReceiptToResult(receipt1, &sdk.Result{})
	case CategoryTypeWithdrawal:
		bal, _ := sdk.NewIntFromString("21888242871839275222246405745257275088548364400416034343698204186575808495617")
		balWithdrawal, _ := sdk.NewIntFromString("-21888242871839275222246405745257275088548364400416034343698204186575808495617")

		ftl := []Flow{}
		ftl1 := input.rek.NewBalanceFlow(input.cuaddress, "btc", "", bal, balWithdrawal,
			sdk.NewInt(0), sdk.NewInt(10))
		ftl2 := input.rek.NewOrderFlow("tempbtc", input.cuaddress, "uuidtemp", sdk.OrderTypeWithdrawal, sdk.OrderStatusBegin)
		ftl = append(ftl, ftl1)
		ftl = append(ftl, ftl2)
		receipt1 := input.rek.NewReceipt(CategoryTypeDeposit, ftl)
		return input.rek.SaveReceiptToResult(receipt1, &sdk.Result{})
	default:
		return &sdk.Result{}
	}

}

type testInput struct {
	cdc       *codec.Codec
	ctx       sdk.Context
	rek       *Keeper
	cuaddress sdk.CUAddress
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()
	cdc := codec.New()
	RegisterCodec(cdc)
	ms := store.NewCommitMultiStore(db)

	rek := NewKeeper(cdc)
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	_, _, cuAddr := testGenKey()

	return testInput{cdc: cdc, ctx: ctx, rek: rek, cuaddress: cuAddr}
}

func testGenKey() (crypto.PrivKey, crypto.PubKey, sdk.CUAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.CUAddress(pub.Address())
	return key, pub, addr
}
