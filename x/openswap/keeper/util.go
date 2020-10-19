package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

func (k Keeper) canRepurchase(ctx sdk.Context, symbol sdk.Symbol) bool {
	if symbol == sdk.NativeDefiToken {
		return true
	}
	pair := k.GetTradingPair(ctx, symbol, sdk.NativeDefiToken)
	if pair != nil {
		return true
	}
	pair = k.GetTradingPair(ctx, symbol, types.RepurchaseRoutingCoin)
	if pair == nil {
		return false
	}
	pair = k.GetTradingPair(ctx, types.RepurchaseRoutingCoin, sdk.NativeDefiToken)
	return pair != nil
}

func (k Keeper) getRealInCoeff(ctx sdk.Context) sdk.Dec {
	return sdk.OneDec().Sub(k.RefererTransactionBonusRate(ctx)).Sub(k.FeeRate(ctx)).Sub(k.RepurchaseRate(ctx))
}

func (k Keeper) getAmountOut(ctx sdk.Context, amountIn sdk.Int, path []sdk.Symbol) (sdk.Int, error) {
	coeff := k.getRealInCoeff(ctx)
	amountOut := amountIn
	for i := 0; i < len(path)-1; i++ {
		reserveIn, reserveOut, err := k.getReserve(ctx, path[i], path[i+1])
		if err != nil {
			return sdk.ZeroInt(), err
		}

		amountOut = amountOut.ToDec().Mul(coeff).TruncateInt()
		amountOut = mulAndDiv(amountOut, reserveOut, reserveIn.Add(amountOut))
	}
	return amountOut, nil
}

func (k Keeper) getAmountIn(ctx sdk.Context, amountOut sdk.Int, path []sdk.Symbol) (sdk.Int, error) {
	coeff := k.getRealInCoeff(ctx)
	amountIn := amountOut
	for i := len(path) - 1; i > 0; i-- {
		reserveIn, reserveOut, err := k.getReserve(ctx, path[i-1], path[i])
		if err != nil {
			return sdk.ZeroInt(), err
		}
		if reserveOut.LTE(amountOut) {
			return sdk.ZeroInt(), fmt.Errorf("insufficient reserve out, have %v, need %v", reserveOut.String(), amountOut.String())
		}

		amountIn = mulAndDiv(amountIn, reserveIn, reserveOut.Sub(amountIn))
		amountIn = amountIn.ToDec().Quo(coeff).TruncateInt()
	}
	return amountIn, nil
}

func (k Keeper) getReserve(ctx sdk.Context, tokenA, tokenB sdk.Symbol) (sdk.Int, sdk.Int, error) {
	pair := k.GetTradingPair(ctx, tokenA, tokenB)
	if pair == nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("%s-%s not found", tokenA.String(), tokenB.String())
	}
	if tokenA == pair.TokenA {
		return pair.TokenAAmount, pair.TokenBAmount, nil
	}
	return pair.TokenBAmount, pair.TokenAAmount, nil
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

func (k Keeper) setDec(ctx sdk.Context, key []byte, d sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(d)
	store.Set(key, bz)
}

func (k Keeper) getFlowFromResult(result *sdk.Result) []sdk.Flow {
	receipt, _ := k.rk.GetReceiptFromResult(result)
	return receipt.Flows
}

func mulAndDiv(amountA, amountB, amountC sdk.Int) sdk.Int {
	product := big.NewInt(0).Mul(amountA.BigInt(), amountB.BigInt())
	return sdk.NewIntFromBigInt(big.NewInt(0).Quo(product, amountC.BigInt()))
}

func calLimitSwapAmount(order *types.Order, pair *types.TradingPair) sdk.Int {
	amount := sdk.ZeroInt()
	curPrice := pair.Price()
	if order.Side == types.OrderSideBuy && order.Price.GT(curPrice) {
		amount = pair.TokenAAmount.ToDec().Mul(order.Price).TruncateInt().Sub(pair.TokenBAmount)
	} else if order.Side == types.OrderSideSell && order.Price.LT(curPrice) {
		amount = pair.TokenBAmount.ToDec().Quo(order.Price).TruncateInt().Sub(pair.TokenAAmount)
	}
	if amount.IsPositive() {
		return amount
	}
	return sdk.ZeroInt()
}
