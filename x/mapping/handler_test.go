package mapping

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

func TestMappingSwap(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.mk
	ck := input.ck
	rk := input.rk
	tk := input.tk
	symbol := sdk.Symbol("tbtc")
	denom := symbol.ToDenomName()
	targetSymbol := sdk.Symbol("btc")
	targetDenom := targetSymbol.ToDenomName()
	from, _ := sdk.CUAddressFromBase58("HBCZSkjCGQggAT28GcQednHbpJyfxHhmeTCH")

	// Prepare mapping
	mappingInfo := &MappingInfo{
		IssueSymbol:  symbol,
		TargetSymbol: targetSymbol,
		TotalSupply:  sdk.NewInt(2100),
		IssuePool:    sdk.NewInt(2100),
		Enabled:      true,
	}
	cu := ck.NewCUWithAddress(ctx, sdk.CUTypeUser, from)
	_ = cu.SetCoins(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2100))))
	ck.SetCU(ctx, cu)
	keeper.SetMappingInfo(ctx, mappingInfo)

	// Swap from target
	msgIssueNotFound := NewMsgMappingSwap(
		from,
		sdk.Symbol("notfound"),
		sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(10))))
	res := handleMsgMappingSwap(ctx, keeper, msgIssueNotFound)
	assert.False(t, res.IsOK())

	msgInvalidSwapAmountZero := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.ZeroInt())))
	res = handleMsgMappingSwap(ctx, keeper, msgInvalidSwapAmountZero)
	assert.False(t, res.IsOK())

	msgInvalidSwapAmountDenom := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin("notexist", sdk.NewInt(10))))
	res = handleMsgMappingSwap(ctx, keeper, msgInvalidSwapAmountDenom)
	assert.False(t, res.IsOK())

	msgCoinsFromNil := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(10))))
	res = handleMsgMappingSwap(ctx, keeper, msgCoinsFromNil)
	assert.False(t, res.IsOK())

	_ = cu.SetCoins(sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(500))))
	ck.SetCU(ctx, cu)

	msgCoinsFromNotEnough := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(510)))) // > 500
	res = handleMsgMappingSwap(ctx, keeper, msgCoinsFromNotEnough)
	assert.False(t, res.IsOK())

	_ = cu.SetCoins(sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(2200))))
	ck.SetCU(ctx, cu)

	msgPoolFromNotEnough := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(2101))))
	res = handleMsgMappingSwap(ctx, keeper, msgPoolFromNotEnough)
	assert.False(t, res.IsOK())

	_ = cu.SetCoins(sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(500))))
	ck.SetCU(ctx, cu)

	msgTarget := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(targetDenom, sdk.NewInt(100))))

	// Disable mapping
	mi := keeper.GetMappingInfo(ctx, symbol)
	mi.Enabled = false
	keeper.SetMappingInfo(ctx, mi)
	res = handleMsgMappingSwap(ctx, keeper, msgTarget)
	assert.False(t, res.IsOK())
	// Enable back
	mi = keeper.GetMappingInfo(ctx, symbol)
	mi.Enabled = true
	keeper.SetMappingInfo(ctx, mi)

	// Disable issue token send
	issueTokenInfo := tk.GetTokenInfo(ctx, symbol)
	issueTokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, issueTokenInfo)
	res = handleMsgMappingSwap(ctx, keeper, msgTarget)
	assert.False(t, res.IsOK())
	// Enable back
	issueTokenInfo.IsSendEnabled = true
	tk.SetTokenInfo(ctx, issueTokenInfo)

	// Disabled target token send
	targetTokenInfo := tk.GetTokenInfo(ctx, targetSymbol)
	targetTokenInfo.IsSendEnabled = false
	tk.SetTokenInfo(ctx, targetTokenInfo)
	res = handleMsgMappingSwap(ctx, keeper, msgTarget)
	assert.False(t, res.IsOK())
	// Enable back
	targetTokenInfo.IsSendEnabled = true
	tk.SetTokenInfo(ctx, targetTokenInfo)

	res = handleMsgMappingSwap(ctx, keeper, msgTarget)
	assert.True(t, res.IsOK())
	cu = ck.GetCU(ctx, from)
	assert.True(t, cu.GetCoins().IsEqual(sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
		sdk.NewCoin(targetDenom, sdk.NewInt(400)))))
	mi = keeper.GetMappingInfo(ctx, symbol)
	assert.True(t, mi.IssuePool.Equal(sdk.NewInt(2100-100)))
	receipt, err := rk.GetReceiptFromResult(&res)
	assert.NoError(t, err)
	flowId := 0
	assert.Equal(t, targetSymbol, receipt.Flows[flowId].(sdk.BalanceFlow).Symbol)
	assert.Equal(t, from, receipt.Flows[flowId].(sdk.BalanceFlow).CUAddress)
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalance.Equal(
		sdk.NewInt(500)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceChange.Equal(
		sdk.NewInt(-100)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalanceOnHold.IsZero())
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceOnHoldChange.IsZero())
	flowId = 1
	assert.Equal(t, symbol, receipt.Flows[flowId].(sdk.BalanceFlow).Symbol)
	assert.Equal(t, from, receipt.Flows[flowId].(sdk.BalanceFlow).CUAddress)
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalance.Equal(
		sdk.NewInt(0)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceChange.Equal(
		sdk.NewInt(100)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalanceOnHold.IsZero())
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceOnHoldChange.IsZero())
	flowId = 2
	assert.Equal(t, symbol, receipt.Flows[flowId].(MappingBalanceFlow).IssueSymbol)
	assert.True(t, receipt.Flows[flowId].(MappingBalanceFlow).PreviousIssuePool.Equal(
		sdk.NewInt(2100)))
	assert.True(t, receipt.Flows[flowId].(MappingBalanceFlow).IssuePoolChange.Equal(
		sdk.NewInt(-100)))

	// Swap from issue
	msgInvalidSwapAmountZero = NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(denom, sdk.ZeroInt())))
	res = handleMsgMappingSwap(ctx, keeper, msgInvalidSwapAmountZero)
	assert.False(t, res.IsOK())

	msgInvalidSwapAmountDenom = NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin("notexist", sdk.NewInt(10))))
	res = handleMsgMappingSwap(ctx, keeper, msgInvalidSwapAmountDenom)
	assert.False(t, res.IsOK())

	msgCoinsFromNotEnough = NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(110)))) // > 100
	res = handleMsgMappingSwap(ctx, keeper, msgCoinsFromNotEnough)
	assert.False(t, res.IsOK())

	_ = cu.SetCoins(sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(200)),
		sdk.NewCoin(targetDenom, sdk.NewInt(400))))
	ck.SetCU(ctx, cu)
	msgPoolFromNotEnough = NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(101))))
	res = handleMsgMappingSwap(ctx, keeper, msgPoolFromNotEnough)
	assert.False(t, res.IsOK())

	_ = cu.SetCoins(sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
		sdk.NewCoin(targetDenom, sdk.NewInt(400))))
	ck.SetCU(ctx, cu)

	msgIssue := NewMsgMappingSwap(
		from,
		symbol,
		sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(30))))
	res = handleMsgMappingSwap(ctx, keeper, msgIssue)
	assert.True(t, res.IsOK())
	cu = ck.GetCU(ctx, from)
	assert.True(t, cu.GetCoins().IsEqual(sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100-30)),
		sdk.NewCoin(targetDenom, sdk.NewInt(400+30)))))
	mi = keeper.GetMappingInfo(ctx, symbol)
	assert.True(t, mi.IssuePool.Equal(sdk.NewInt(2100-100+30)))
	receipt, err = rk.GetReceiptFromResult(&res)
	assert.NoError(t, err)
	flowId = 0
	assert.Equal(t, symbol, receipt.Flows[flowId].(sdk.BalanceFlow).Symbol)
	assert.Equal(t, from, receipt.Flows[flowId].(sdk.BalanceFlow).CUAddress)
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalance.Equal(
		sdk.NewInt(100)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceChange.Equal(
		sdk.NewInt(-30)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalanceOnHold.IsZero())
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceOnHoldChange.IsZero())
	flowId = 1
	assert.Equal(t, targetSymbol, receipt.Flows[flowId].(sdk.BalanceFlow).Symbol)
	assert.Equal(t, from, receipt.Flows[flowId].(sdk.BalanceFlow).CUAddress)
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalance.Equal(
		sdk.NewInt(400)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceChange.Equal(
		sdk.NewInt(30)))
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).PreviousBalanceOnHold.IsZero())
	assert.True(t, receipt.Flows[flowId].(sdk.BalanceFlow).BalanceOnHoldChange.IsZero())
	flowId = 2
	assert.Equal(t, symbol, receipt.Flows[flowId].(MappingBalanceFlow).IssueSymbol)
	assert.True(t, receipt.Flows[flowId].(MappingBalanceFlow).PreviousIssuePool.Equal(
		sdk.NewInt(2000)))
	assert.True(t, receipt.Flows[flowId].(MappingBalanceFlow).IssuePoolChange.Equal(
		sdk.NewInt(30)))
}
