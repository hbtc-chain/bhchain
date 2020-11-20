package keeper

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func (k Keeper) CalculateEarning(ctx sdk.Context, addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol) sdk.Int {
	tokenA, tokenB = k.SortToken(tokenA, tokenB)
	liquidity := k.GetLiquidity(ctx, addr, dexID, tokenA, tokenB)
	globalMask := k.getDec(ctx, types.GlobalMaskKey(dexID, tokenA, tokenB))
	addrMask := k.getDec(ctx, types.AddrMaskKey(addr, dexID, tokenA, tokenB))
	return globalMask.Mul(liquidity.ToDec()).Sub(addrMask).TruncateInt()
}

func (k Keeper) ClaimEarning(ctx sdk.Context, addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol) sdk.Result {
	tokenA, tokenB = k.SortToken(tokenA, tokenB)
	earning := k.CalculateEarning(ctx, addr, dexID, tokenA, tokenB)
	if !earning.IsPositive() {
		return sdk.Result{}
	}

	var flows []sdk.Flow
	referer := k.GetReferer(ctx, addr)
	if referer == nil {
		result, _ := k.sk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.NativeDefiToken, earning)))
		flows = k.getFlowFromResult(&result)
	} else {
		refererShare := earning.ToDec().Mul(k.RefererMiningBonusRate(ctx)).TruncateInt()
		selfShare := earning.Sub(refererShare)
		result, _ := k.sk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, referer, sdk.NewCoins(sdk.NewCoin(sdk.NativeDefiToken, refererShare)))
		flows = k.getFlowFromResult(&result)
		result, _ = k.sk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.NativeDefiToken, selfShare)))
		flows = append(flows, k.getFlowFromResult(&result)...)
	}

	addrMaskKey := types.AddrMaskKey(addr, dexID, tokenA, tokenB)
	addrMask := k.getDec(ctx, addrMaskKey)
	addrMask = addrMask.Add(earning.ToDec())
	k.setDec(ctx, addrMaskKey, addrMask)

	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawEarning,
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, earning.String()),
		),
	})
	result.Events = append(result.Events, ctx.EventManager().Events()...)

	return result
}

func (k Keeper) Mining(ctx sdk.Context) {
	amount := k.getMiningAmount(ctx)
	if !amount.IsPositive() {
		return
	}

	defiToken := k.tokenKeeper.GetToken(ctx, sdk.NativeDefiToken)
	if defiToken == nil {
		return
	}
	circulation := k.sk.GetSupply(ctx).GetTotal().AmountOf(sdk.NativeDefiToken)
	burned := k.sk.GetSupply(ctx).GetBurned().AmountOf(sdk.NativeDefiToken)
	maxMining := defiToken.GetTotalSupply().Sub(circulation).Sub(burned)
	amount = sdk.MinInt(amount, maxMining)
	if !amount.IsPositive() {
		return
	}

	k.sk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.NativeDefiToken, amount)))
	k.distribute(ctx, amount)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeMining,
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	})
}

func (k Keeper) distribute(ctx sdk.Context, amount sdk.Int) {
	totalWeight := sdk.ZeroInt()
	miningWeights := k.MiningWeights(ctx)
	for _, w := range miningWeights {
		totalWeight = totalWeight.Add(w.Weight)
	}

	remaining := amount
	for i, w := range miningWeights {
		distribution := remaining
		if i < len(miningWeights)-1 {
			distribution = amount.Mul(w.Weight).Quo(totalWeight)
			remaining = remaining.Sub(distribution)
		}
		tokenA, tokenB := k.SortToken(w.TokenA, w.TokenB)
		k.onMining(ctx, w.DexID, tokenA, tokenB, distribution)
	}
}

func (k Keeper) getMiningAmount(ctx sdk.Context) sdk.Int {
	height := uint64(ctx.BlockHeight())
	amount := sdk.ZeroInt()
	for _, plan := range k.MiningPlans(ctx) {
		if plan.StartHeight > height {
			break
		}
		amount = plan.MiningPerBlock
	}
	return amount
}

func (k Keeper) onUpdateLiquidity(ctx sdk.Context, addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol, liquidity sdk.Int) {
	// update total share
	totalShareKey := types.TotalShareKey(dexID, tokenA, tokenB)
	totalShare := k.getDec(ctx, totalShareKey)
	totalShare = totalShare.Add(liquidity.ToDec())
	k.setDec(ctx, totalShareKey, totalShare)

	// update addr mast
	globalMask := k.getDec(ctx, types.GlobalMaskKey(dexID, tokenA, tokenB))
	if globalMask.IsPositive() {
		addrMaskKey := types.AddrMaskKey(addr, dexID, tokenA, tokenB)
		addrMask := k.getDec(ctx, addrMaskKey)
		addrMask = addrMask.Add(globalMask.Mul(liquidity.ToDec()))
		k.setDec(ctx, addrMaskKey, addrMask)
	}
}

func (k Keeper) onMining(ctx sdk.Context, dexID uint32, tokenA, tokenB sdk.Symbol, amount sdk.Int) {
	totalShare := k.getDec(ctx, types.TotalShareKey(dexID, tokenA, tokenB))
	if totalShare.IsPositive() {
		globalMaskKey := types.GlobalMaskKey(dexID, tokenA, tokenB)
		globalMask := k.getDec(ctx, globalMaskKey)
		globalMask = globalMask.Add(amount.ToDec().Quo(totalShare))
		k.setDec(ctx, globalMaskKey, globalMask)
	}
}
