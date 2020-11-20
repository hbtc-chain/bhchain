package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func TestKeeper(t *testing.T) {
	input := setupTestInput(t)
	ctx := input.ctx
	rk := input.rk

	addr := sdk.NewCUAddress()
	addr2 := sdk.NewCUAddress()
	addr3 := sdk.NewCUAddress()
	cu := input.ck.NewCUWithAddress(ctx, sdk.CUTypeUser, addr)

	// Test GetCoins/SetCoins
	input.ck.SetCU(ctx, cu)
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins()))

	input.k.SubCoins(ctx, addr, input.k.GetAllBalance(ctx, addr))
	//testSetCUCoins(ctx, trk, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	_, flows, err := input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf := flows[0].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(10), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())

	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test AddCoins
	coins, flows, err := input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf = flows[0].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, sdk.NewInt(10), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(15), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 25))))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 25)), coins)

	coins, flows, err = input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf = flows[0].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(15), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15), sdk.NewInt64Coin("foocoin", 25))))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15)), coins)

	// Test SubtractCoins
	coins, flows, err = input.k.SubCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf = flows[0].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, sdk.NewInt(25), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(10).Neg(), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())

	coins, flows, err = input.k.SubCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf = flows[0].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, sdk.NewInt(15), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(5).Neg(), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15))))

	_, flows, err = input.k.SubCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 11)))
	require.NotNil(t, err)
	require.Equal(t, 0, len(flows))
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15))))

	coins, flows, err = input.k.SubCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf = flows[0].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, sdk.NewInt(10), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(10).Neg(), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())
	require.Empty(t, coins)
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	//require.False(t, input.k.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 1))))

	// Test SendCoins
	result, _, err := input.k.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5)))
	require.Nil(t, err)

	//check receipts
	receipt, err1 := rk.GetReceiptFromResult(&result)
	require.Nil(t, err1)

	require.Equal(t, sdk.CategoryTypeTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))
	bf = receipt.Flows[0].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr, bf.CUAddress)
	require.Equal(t, sdk.NewInt(15), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(5).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[1].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr2, bf.CUAddress)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(5), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, input.k.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	result, _, err2 := input.k.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	require.Implements(t, (*sdk.Error)(nil), err2)
	require.NotEqual(t, sdk.CodeOK, result.Code)
	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, input.k.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	_, flows, err = input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 30)))
	require.Nil(t, err)
	require.Equal(t, 1, len(flows))
	bf = flows[0].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalance)
	require.True(t, bf.PreviousBalanceOnHold.IsZero())
	require.Equal(t, sdk.NewInt(30), bf.BalanceChange)
	require.True(t, bf.BalanceOnHoldChange.IsZero())

	result, _, err = input.k.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)))
	require.Nil(t, err)

	//check receipts
	receipt, err1 = rk.GetReceiptFromResult(&result)
	require.Nil(t, err1)

	require.Equal(t, sdk.CategoryTypeTransfer, receipt.Category)
	require.Equal(t, 4, len(receipt.Flows))
	bf = receipt.Flows[0].(sdk.BalanceFlow)

	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, addr, bf.CUAddress)
	require.Equal(t, sdk.NewInt(30), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(10).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[1].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, addr2, bf.CUAddress)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(10), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[2].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr, bf.CUAddress)
	require.Equal(t, sdk.NewInt(10), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(5).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[3].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr2, bf.CUAddress)
	require.Equal(t, sdk.NewInt(5), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(5), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5))))
	require.True(t, input.k.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10))))

	// Test InputOutputCoins
	input1 := types.NewInput(addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 2)))
	output1 := types.NewOutput(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 2)))
	result, err = input.k.InputOutputCoins(ctx, []types.Input{input1}, []types.Output{output1})
	require.Nil(t, err)

	//check receipts
	receipt, err1 = rk.GetReceiptFromResult(&result)
	require.Nil(t, err1)

	require.Equal(t, sdk.CategoryTypeMultiTransfer, receipt.Category)
	require.Equal(t, 2, len(receipt.Flows))

	bf = receipt.Flows[0].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr2, bf.CUAddress)
	require.Equal(t, sdk.NewInt(10), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(2).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[1].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr, bf.CUAddress)
	require.Equal(t, sdk.NewInt(5), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(2), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 7))))
	require.True(t, input.k.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 8))))

	inputs := []types.Input{
		types.NewInput(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 3))),
		types.NewInput(addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 3), sdk.NewInt64Coin("foocoin", 2))),
	}

	outputs := []types.Output{
		types.NewOutput(addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 1))),
		types.NewOutput(addr3, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5))),
	}

	result, err = input.k.InputOutputCoins(ctx, inputs, outputs)
	require.Nil(t, err)

	//check receipts
	receipt, err1 = rk.GetReceiptFromResult(&result)
	require.Nil(t, err1)

	require.Equal(t, sdk.CategoryTypeMultiTransfer, receipt.Category)
	require.Equal(t, 6, len(receipt.Flows))

	bf = receipt.Flows[0].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr, bf.CUAddress)
	require.Equal(t, sdk.NewInt(7), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(3).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[1].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, addr2, bf.CUAddress)
	require.Equal(t, sdk.NewInt(10), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(3).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[2].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr2, bf.CUAddress)
	require.Equal(t, sdk.NewInt(8), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(2).Neg(), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[3].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, addr, bf.CUAddress)
	require.Equal(t, sdk.NewInt(20), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(1), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[4].(sdk.BalanceFlow)
	require.Equal(t, "barcoin", bf.Symbol.String())
	require.Equal(t, addr3, bf.CUAddress)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(2), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = receipt.Flows[5].(sdk.BalanceFlow)
	require.Equal(t, "foocoin", bf.Symbol.String())
	require.Equal(t, addr3, bf.CUAddress)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalance)
	require.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	require.Equal(t, sdk.NewInt(5), bf.BalanceChange)
	require.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	require.True(t, input.k.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 21), sdk.NewInt64Coin("foocoin", 4))))
	require.True(t, input.k.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 7), sdk.NewInt64Coin("foocoin", 6))))
	require.True(t, input.k.GetAllBalance(ctx, addr3).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5))))
}

func TestSendKeeper(t *testing.T) {
	input := setupTestInput(t)
	ctx := input.ctx

	sendKeeper := input.k
	input.k.SetSendEnabled(ctx, true)

	addr := sdk.NewCUAddress()
	addr2 := sdk.NewCUAddress()
	cu := input.ck.NewCUWithAddress(ctx, sdk.CUTypeUser, addr)

	// Test GetCoins/SetCoins
	input.ck.SetCU(ctx, cu)
	require.True(t, sendKeeper.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins()))

	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, sendKeeper.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	//require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	//require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	//require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	//require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))
	input.k.SubCoins(ctx, addr, sendKeeper.GetAllBalance(ctx, addr))
	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15)))

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, sendKeeper.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	_, _, err := sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	require.Implements(t, (*sdk.Error)(nil), err)
	require.True(t, sendKeeper.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 30)))
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, sendKeeper.GetAllBalance(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5))))
	require.True(t, sendKeeper.GetAllBalance(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10))))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	_, _, err = sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{"FOOCOIN", sdk.NewInt(-5)}})
	require.Error(t, err)
}
