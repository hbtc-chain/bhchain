package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func (k Keeper) canRepurchase(ctx sdk.Context, tokenIn sdk.Symbol, amount sdk.Int) bool {
	repurchaseToken := sdk.Symbol(k.RepurchaseToken(ctx))
	if repurchaseToken == "" {
		return false
	}
	if tokenIn == repurchaseToken {
		return true
	}
	pair := k.GetTradingPair(ctx, 0, tokenIn, repurchaseToken)
	return pair != nil
}

func (k Keeper) getFeeRates(ctx sdk.Context, pair *types.TradingPair) *types.FeeRate {
	if pair.DexID == 0 {
		return types.NewFeeRate(k.LpRewardRate(ctx), k.RepurchaseRate(ctx), k.RefererTransactionBonusRate(ctx))
	}
	if pair.IsPublic {
		return types.NewFeeRate(k.LpRewardRate(ctx), k.RepurchaseRate(ctx), pair.RefererRewardRate)
	}
	return types.NewFeeRate(pair.LPRewardRate, k.RepurchaseRate(ctx), pair.RefererRewardRate)
}

func (k Keeper) getRealInCoeff(ctx sdk.Context, pair *types.TradingPair) sdk.Dec {
	return sdk.OneDec().Sub(k.getFeeRates(ctx, pair).TotalFeeRate())
}

func (k Keeper) getReserves(ctx sdk.Context, pair *types.TradingPair, tokenA, tokenB sdk.Symbol) (sdk.Int, sdk.Int) {
	if pair.IsPublic && pair.DexID != 0 {
		pair = k.GetTradingPair(ctx, 0, tokenA, tokenB)
	}
	if pair == nil {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}
	if tokenA == pair.TokenA {
		return pair.TokenAAmount, pair.TokenBAmount
	}
	return pair.TokenBAmount, pair.TokenAAmount
}

func (k Keeper) getAmountOut(ctx sdk.Context, dexID uint32, amountIn sdk.Int, path []sdk.Symbol) (sdk.Int, error) {
	amountOut := amountIn
	for i := 0; i < len(path)-1; i++ {
		pair := k.GetTradingPair(ctx, dexID, path[i], path[i+1])
		if pair == nil {
			return sdk.ZeroInt(), fmt.Errorf("%s-%s trading pair does not exist in dex %d",
				path[i], path[i+1], dexID)
		}
		reserveIn, reserveOut := k.getReserves(ctx, pair, path[i], path[i+1])
		if !reserveIn.IsPositive() || !reserveOut.IsPositive() {
			return sdk.ZeroInt(), fmt.Errorf("%s-%s trading pair does not have enough liquidity", path[i], path[i+1])
		}

		coeff := k.getRealInCoeff(ctx, pair)
		amountOut = amountOut.ToDec().Mul(coeff).TruncateInt()
		amountOut = mulAndDiv(amountOut, reserveOut, reserveIn.Add(amountOut))
	}
	return amountOut, nil
}

func (k Keeper) getAmountIn(ctx sdk.Context, dexID uint32, amountOut sdk.Int, path []sdk.Symbol) (sdk.Int, error) {
	amountIn := amountOut
	for i := len(path) - 1; i > 0; i-- {
		pair := k.GetTradingPair(ctx, dexID, path[i-1], path[i])
		if pair == nil {
			return sdk.ZeroInt(), fmt.Errorf("%s-%s trading pair does not exist in dex %d",
				path[i-1], path[i], dexID)
		}
		reserveIn, reserveOut := k.getReserves(ctx, pair, path[i-1], path[i])
		if !reserveIn.IsPositive() || !reserveOut.IsPositive() {
			return sdk.ZeroInt(), fmt.Errorf("%s-%s trading pair does not have enough liquidity", path[i-1], path[i])
		}
		if reserveOut.LTE(amountIn) {
			return sdk.ZeroInt(), fmt.Errorf("insufficient reserve out, have %v, need %v", reserveOut.String(), amountIn.String())
		}

		coeff := k.getRealInCoeff(ctx, pair)
		amountIn = mulAndDiv(amountIn, reserveIn, reserveOut.Sub(amountIn))
		amountIn = amountIn.ToDec().Quo(coeff).TruncateInt().AddRaw(1)
	}
	return amountIn, nil
}

func (k Keeper) getDec(ctx sdk.Context, key []byte) (ret sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.ZeroDec()
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &ret)
	return
}

func (k Keeper) CheckSymbol(ctx sdk.Context, symbol sdk.Symbol) (sdk.Token, sdk.Result) {
	tokenInfo := k.tokenKeeper.GetToken(ctx, symbol)
	if tokenInfo == nil {
		return nil, sdk.ErrUnSupportToken(fmt.Sprintf("token %s does not exist", symbol.String())).Result()
	}
	if !tokenInfo.IsSendEnabled() {
		return tokenInfo, sdk.ErrUnSupportToken(fmt.Sprintf("token %s is not enable to send", symbol)).Result()
	}
	return tokenInfo, sdk.Result{}
}

func (k Keeper) SortTokens(ctx sdk.Context, symbolA, symbolB sdk.Symbol) (sdk.Symbol, sdk.Symbol, sdk.Result) {
	tokenA, result := k.CheckSymbol(ctx, symbolA)
	if !result.IsOK() {
		return symbolA, symbolB, result
	}

	tokenB, result := k.CheckSymbol(ctx, symbolB)
	if !result.IsOK() {
		return symbolA, symbolB, result
	}

	// sort by symbol
	if tokenA.GetWeight() == tokenB.GetWeight() {
		if symbolA > symbolB {
			return symbolB, symbolA, sdk.Result{}
		}
		return symbolA, symbolB, sdk.Result{}
	}

	// sort by weight
	if tokenA.GetWeight() < tokenB.GetWeight() {
		return symbolB, symbolA, sdk.Result{}
	}
	return symbolA, symbolB, sdk.Result{}
}

func (k Keeper) setDec(ctx sdk.Context, key []byte, d sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(d)
	store.Set(key, bz)
}

func (k Keeper) getFlowFromResult(result *sdk.Result) []sdk.Flow {
	receipt, _ := k.rk.GetReceiptFromResult(result)
	return receipt.Flows
}

func (k Keeper) calLimitSwapAmount(ctx sdk.Context, order *types.Order, pair *types.TradingPair) (sdk.Int, bool) {
	if !pair.TokenAAmount.IsPositive() || !pair.TokenBAmount.IsPositive() {
		return sdk.ZeroInt(), false
	}

	curPrice := pair.Price()
	amount := sdk.ZeroInt()
	var priceSuitable bool
	if order.Side == types.OrderSideBuy && order.Price.GT(curPrice) {
		amount = pair.TokenAAmount.ToDec().Mul(order.Price).TruncateInt().Sub(pair.TokenBAmount)
		priceSuitable = true
	} else if order.Side == types.OrderSideSell && order.Price.LT(curPrice) {
		amount = pair.TokenBAmount.ToDec().Quo(order.Price).TruncateInt().Sub(pair.TokenAAmount)
		priceSuitable = true
	}
	if amount.IsPositive() {
		return amount, priceSuitable
	}
	return sdk.ZeroInt(), priceSuitable
}

func mulAndDiv(amountA, amountB, amountC sdk.Int) sdk.Int {
	product := big.NewInt(0).Mul(amountA.BigInt(), amountB.BigInt())
	return sdk.NewIntFromBigInt(big.NewInt(0).Quo(product, amountC.BigInt()))
}
