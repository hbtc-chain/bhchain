package keeper

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/orderbook"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	tokenKeeper   types.TokenKeeper
	tk            types.TransferKeeper
	rk            types.ReceiptKeeper
	sk            types.SupplyKeeper
	marketManager *orderbook.Manager
	paramstore    params.Subspace
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, tokenKeeper types.TokenKeeper,
	rk types.ReceiptKeeper, sk types.SupplyKeeper, tk types.TransferKeeper, paramstore params.Subspace) Keeper {
	k := Keeper{
		storeKey:    key,
		cdc:         cdc,
		tokenKeeper: tokenKeeper,
		rk:          rk,
		sk:          sk,
		tk:          tk,
		paramstore:  paramstore.WithKeyTable(ParamKeyTable()),
	}
	k.marketManager = orderbook.NewManager(k)
	return k
}

func (k Keeper) AddLiquidity(ctx sdk.Context, from sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol,
	maxTokenAAmount, maxTokenBAmount sdk.Int) sdk.Result {

	pair := k.GetTradingPair(ctx, dexID, tokenA, tokenB)
	if pair == nil && dexID != 0 {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s-%s trading pair does not exist in dex %d",
			tokenA, tokenB, dexID)).Result()
	}
	if pair != nil && pair.IsPublic && pair.DexID != 0 {
		pair = k.GetTradingPair(ctx, 0, tokenA, tokenB)
	}
	var needTokenA, needTokenB, liquidity sdk.Int
	if pair == nil || pair.TotalLiquidity.IsZero() {
		product := big.NewInt(0).Mul(maxTokenAAmount.BigInt(), maxTokenBAmount.BigInt())
		sqr := sdk.NewIntFromBigInt(big.NewInt(0).Sqrt(product))
		minimumLiquidity := k.MinimumLiquidity(ctx)
		liquidity = sqr.Sub(minimumLiquidity) // lock MinimumLiquidity permanently
		if !liquidity.IsPositive() {
			return sdk.ErrInvalidAmount("insufficient liquidity").Result()
		}
		if pair == nil {
			pair = types.NewDefaultTradingPair(tokenA, tokenB, minimumLiquidity)
		} else {
			pair.TotalLiquidity = minimumLiquidity
		}
		needTokenA, needTokenB = maxTokenAAmount, maxTokenBAmount
	} else {
		needTokenB = mulAndDiv(maxTokenAAmount, pair.TokenBAmount, pair.TokenAAmount)
		if needTokenB.GT(maxTokenBAmount) {
			needTokenA = mulAndDiv(maxTokenBAmount, pair.TokenAAmount, pair.TokenBAmount)
			if needTokenA.GT(maxTokenAAmount) {
				return sdk.ErrInvalidAmount("need amount exceeds expectation").Result()
			}
			needTokenB = maxTokenBAmount
		} else {
			needTokenA = maxTokenAAmount
		}
		liquidity = sdk.MinInt(mulAndDiv(needTokenA, pair.TotalLiquidity, pair.TokenAAmount), mulAndDiv(needTokenB, pair.TotalLiquidity, pair.TokenBAmount))
	}

	if !needTokenA.IsPositive() || !needTokenB.IsPositive() {
		return sdk.ErrInvalidTx(fmt.Sprintf("amount is too small, %s amount: %v, %s amount: %v",
			tokenA, tokenB, needTokenA.String(), needTokenB.String())).Result()
	}
	if !liquidity.IsPositive() {
		return sdk.ErrInvalidTx(fmt.Sprintf("liquidity %s is too small",
			liquidity.String())).Result()
	}

	need := sdk.NewCoins(sdk.NewCoin(tokenA.String(), needTokenA), sdk.NewCoin(tokenB.String(), needTokenB))
	_, flows, err := k.tk.SubCoins(ctx, from, need)
	if err != nil {
		return err.Result()
	}

	pair.TokenAAmount = pair.TokenAAmount.Add(needTokenA)
	pair.TokenBAmount = pair.TokenBAmount.Add(needTokenB)
	pair.TotalLiquidity = pair.TotalLiquidity.Add(liquidity)
	k.SaveTradingPair(ctx, pair)

	addrLiquidity := k.GetLiquidity(ctx, from, pair.DexID, pair.TokenA, pair.TokenB)
	addrLiquidity = addrLiquidity.Add(liquidity)
	k.saveLiquidity(ctx, from, pair.DexID, tokenA, tokenB, addrLiquidity)

	k.onUpdateLiquidity(ctx, from, pair.DexID, tokenA, tokenB, liquidity)

	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)
	eventLiquidity := types.NewEventLiquidity(from, pair.DexID, tokenA, tokenB, pair.TokenAAmount, pair.TokenBAmount, needTokenA, needTokenB)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddLiquidity,
			sdk.NewAttribute(types.AttributeKeyLiquidity, eventLiquidity.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) RemoveLiquidity(ctx sdk.Context, from sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol, liquidity sdk.Int) sdk.Result {
	pair := k.GetTradingPair(ctx, dexID, tokenA, tokenB)
	if pair == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s-%s trading pair does not exist in dex %d",
			tokenA.String(), tokenB.String(), dexID)).Result()
	}
	if pair.IsPublic && pair.DexID != 0 {
		pair = k.GetTradingPair(ctx, 0, tokenA, tokenB)
	}
	addrLiquidity := k.GetLiquidity(ctx, from, pair.DexID, pair.TokenA, pair.TokenB)
	if addrLiquidity.LT(liquidity) {
		return sdk.ErrInsufficientFunds(fmt.Sprintf("insufficient liquidity, has %s, need %s", addrLiquidity.String(), liquidity.String())).Result()
	}
	if liquidity.GT(pair.TotalLiquidity) {
		return sdk.ErrInternal("remove amount is larger than total liquidity").Result()
	}
	returnTokenA := mulAndDiv(liquidity, pair.TokenAAmount, pair.TotalLiquidity)
	returnTokenB := mulAndDiv(liquidity, pair.TokenBAmount, pair.TotalLiquidity)
	if !returnTokenA.IsPositive() || !returnTokenB.IsPositive() {
		return sdk.ErrInvalidTx(fmt.Sprintf("amount is too small, %s amount: %v, %s amount: %v",
			tokenA, tokenB, returnTokenA.String(), returnTokenB.String())).Result()
	}

	returnCoins := sdk.NewCoins(sdk.NewCoin(tokenA.String(), returnTokenA), sdk.NewCoin(tokenB.String(), returnTokenB))
	_, flows, err := k.tk.AddCoins(ctx, from, returnCoins)
	if err != nil {
		return err.Result()
	}

	pair.TokenAAmount = pair.TokenAAmount.Sub(returnTokenA)
	pair.TokenBAmount = pair.TokenBAmount.Sub(returnTokenB)
	pair.TotalLiquidity = pair.TotalLiquidity.Sub(liquidity)
	k.SaveTradingPair(ctx, pair)

	addrLiquidity = addrLiquidity.Sub(liquidity)
	k.saveLiquidity(ctx, from, pair.DexID, tokenA, tokenB, addrLiquidity)

	k.onUpdateLiquidity(ctx, from, pair.DexID, tokenA, tokenB, liquidity.Neg())

	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)
	eventLiquidity := types.NewEventLiquidity(from, pair.DexID, tokenA, tokenB, pair.TokenAAmount, pair.TokenBAmount, returnTokenA, returnTokenB)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveLiquidity,
			sdk.NewAttribute(types.AttributeKeyLiquidity, eventLiquidity.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) SwapExactIn(ctx sdk.Context, dexID uint32, from, referer, receiver sdk.CUAddress, amountIn, minAmountOut sdk.Int,
	path []sdk.Symbol) sdk.Result {

	amountOut, err := k.getAmountOut(ctx, dexID, amountIn, path)
	if err != nil {
		return sdk.ErrInvalidTx(err.Error()).Result()
	}
	if amountOut.LT(minAmountOut) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("insufficient amount out, min: %s, got: %s", minAmountOut.String(), amountOut.String())).Result()
	}

	return k.directSwap(ctx, dexID, from, referer, receiver, amountIn, path)
}

func (k Keeper) SwapExactOut(ctx sdk.Context, dexID uint32, from, referer, receiver sdk.CUAddress, amountOut, maxAmountIn sdk.Int,
	path []sdk.Symbol) sdk.Result {

	amountIn, err := k.getAmountIn(ctx, dexID, amountOut, path)
	if err != nil {
		return sdk.ErrInvalidTx(err.Error()).Result()
	}
	if amountIn.GT(maxAmountIn) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("excessive amount in, max: %s, got: %s", maxAmountIn.String(), amountIn.String())).Result()
	}

	return k.directSwap(ctx, dexID, from, referer, receiver, amountIn, path)
}

func (k Keeper) directSwap(ctx sdk.Context, dexID uint32, from, referer, receiver sdk.CUAddress, amountIn sdk.Int, path []sdk.Symbol) sdk.Result {

	flows := make([]sdk.Flow, 0, 4)
	need := sdk.NewCoin(path[0].String(), amountIn)
	_, flow, err := k.tk.SubCoin(ctx, from, need)
	if err != nil {
		return err.Result()
	}
	flows = append(flows, flow)

	bonusCoins := sdk.NewCoins()
	repurchaseFunds := sdk.NewCoins()
	var swapEvents types.EventSwaps
	for i := 0; i < len(path)-1; i++ {
		pair := k.GetTradingPair(ctx, dexID, path[i], path[i+1])
		feeRate := k.getFeeRates(ctx, pair)
		if pair.IsPublic && pair.DexID != 0 {
			pair = k.GetTradingPair(ctx, 0, path[i], path[i+1])
		}
		amountOut, bonus, repurchaseFund, updatedPair := k.swap(ctx, feeRate, pair, path[i], amountIn, false)

		if bonus.IsPositive() {
			bonusCoins = bonusCoins.Add(sdk.NewCoins(sdk.NewCoin(path[i].String(), bonus)))
		}
		if repurchaseFund.IsPositive() {
			repurchaseFunds = repurchaseFunds.Add(sdk.NewCoins(sdk.NewCoin(path[i].String(), repurchaseFund)))
		}

		swapEvents = append(swapEvents, types.NewEventSwap(from, "", dexID, pair.TokenA, pair.TokenB, path[i],
			updatedPair.TokenAAmount, updatedPair.TokenBAmount, amountIn, amountOut))

		amountIn = amountOut
	}

	if amountIn.IsPositive() {
		_, flow, err = k.tk.AddCoin(ctx, receiver, sdk.NewCoin(path[len(path)-1].String(), amountIn))
		if err != nil {
			return err.Result()
		}
		flows = append(flows, flow)
	}

	if bonusCoins.IsValid() {
		_, refererFlows, err := k.tk.AddCoins(ctx, referer, bonusCoins)
		if err != nil {
			return err.Result()
		}
		flows = append(flows, refererFlows...)
	}

	k.addRepurchaseFunds(ctx, repurchaseFunds)

	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSwap,
			sdk.NewAttribute(types.AttributeKeySwapResult, swapEvents.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) LimitSwap(ctx sdk.Context, dexID uint32, orderID string, from, referer, receiver sdk.CUAddress, amountIn sdk.Int,
	price sdk.Dec, baseSymbol, quoteSymbol sdk.Symbol, side int, expiredAt int64) sdk.Result {

	pair := k.GetTradingPair(ctx, dexID, baseSymbol, quoteSymbol)
	if pair == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s-%s trading pair does not exist in dex %d", baseSymbol, quoteSymbol, dexID)).Result()
	}

	feeRate := k.getFeeRates(ctx, pair)
	realInCoeff := sdk.OneDec().Sub(feeRate.TotalFeeRate())
	realAmount := amountIn.ToDec().Mul(realInCoeff).TruncateDec()
	var amountOut sdk.Int
	if side == types.OrderSideBuy {
		amountOut = realAmount.Quo(price).TruncateInt()
	} else {
		amountOut = realAmount.Mul(price).TruncateInt()
	}
	if !amountOut.IsPositive() {
		return sdk.ErrInvalidTx("limit order amount is too small").Result()
	}

	if pair.IsPublic && dexID != 0 {
		pair = k.GetTradingPair(ctx, 0, baseSymbol, quoteSymbol)
	}
	if pair == nil || !pair.TokenAAmount.IsPositive() || !pair.TokenBAmount.IsPositive() {
		return sdk.ErrInvalidTx(fmt.Sprintf("%s-%s trading pair does not have enough liquidity", baseSymbol, quoteSymbol)).Result()
	}

	symbol := baseSymbol
	if side == types.OrderSideBuy {
		symbol = quoteSymbol
	}
	flows, err := k.tk.LockCoin(ctx, from, sdk.NewCoin(symbol.String(), amountIn))
	if err != nil {
		return err.Result()
	}

	order := &types.Order{
		DexID:       pair.DexID,
		OrderID:     orderID,
		CreatedTime: ctx.BlockTime().Unix(),
		ExpiredTime: expiredAt,
		From:        from,
		Referer:     referer,
		Receiver:    receiver,
		Price:       price,
		Side:        byte(side),
		BaseSymbol:  baseSymbol,
		QuoteSymbol: quoteSymbol,
		AmountIn:    amountIn,
		LockedFund:  amountIn,
		FeeRate:     feeRate,
	}
	k.saveOrder(ctx, order)

	balanceFlows, event, _, _ := k.limitSwap(ctx, order, pair)
	for _, flow := range balanceFlows {
		flows = append(flows, flow)
	}
	if order.Status != types.OrderStatusFilled {
		k.addUnfinishedOrder(ctx, order)
		ctx.GasMeter().ConsumeGas(k.LimitSwapMatchingGas(ctx).Uint64(), "limit order matching gas fee")
	}

	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)
	if event != nil {
		events := types.EventSwaps{event}
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeSwap,
				sdk.NewAttribute(types.AttributeKeySwapResult, events.String()),
			),
		})
		result.Events = append(result.Events, ctx.EventManager().Events()...)
	}

	return result
}

func (k Keeper) swap(ctx sdk.Context, feeRate *types.FeeRate, pair *types.TradingPair, tokenIn sdk.Symbol, amountIn sdk.Int,
	isRepurchasing bool) (sdk.Int, sdk.Int, sdk.Int, *types.TradingPair) {

	amountInDec := sdk.NewDecFromInt(amountIn)
	lpReward, refererBonus, repurchaseFund := sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	if !isRepurchasing {
		lpReward = amountInDec.Mul(feeRate.LPRewardRate).TruncateInt()
		repurchaseFund = amountInDec.Mul(feeRate.RepurchaseRate).TruncateInt()
		refererBonus = amountInDec.Mul(feeRate.RefererRewardRate).TruncateInt()
		if !k.canRepurchase(ctx, tokenIn, repurchaseFund) {
			lpReward = lpReward.Add(repurchaseFund)
			repurchaseFund = sdk.ZeroInt()
		}
	}
	realAmountIn := amountIn.Sub(lpReward).Sub(refererBonus).Sub(repurchaseFund)

	var amountOut sdk.Int
	if pair.TokenA == tokenIn {
		amountOut = mulAndDiv(realAmountIn, pair.TokenBAmount, realAmountIn.Add(pair.TokenAAmount))
		pair.TokenAAmount = pair.TokenAAmount.Add(realAmountIn).Add(lpReward)
		pair.TokenBAmount = pair.TokenBAmount.Sub(amountOut)
	} else {
		amountOut = mulAndDiv(realAmountIn, pair.TokenAAmount, realAmountIn.Add(pair.TokenBAmount))
		pair.TokenBAmount = pair.TokenBAmount.Add(realAmountIn).Add(lpReward)
		pair.TokenAAmount = pair.TokenAAmount.Sub(amountOut)
	}
	k.SaveTradingPair(ctx, pair)
	return amountOut, refererBonus, repurchaseFund, pair
}

func (k Keeper) limitSwap(ctx sdk.Context, order *types.Order, pair *types.TradingPair) ([]sdk.Flow, *types.EventSwap, *types.TradingPair, bool) {
	maxAmountIn, priceSuitable := k.calLimitSwapAmount(ctx, order, pair)
	if maxAmountIn.IsZero() {
		return nil, nil, pair, priceSuitable
	}

	realAmountInCoeff := sdk.OneDec().Sub(order.FeeRate.TotalFeeRate())
	realMaxAmountIn := maxAmountIn.ToDec().Quo(realAmountInCoeff).TruncateInt()
	if realMaxAmountIn.GT(order.LockedFund) {
		realMaxAmountIn = order.LockedFund
	}

	tokenIn, tokenOut := pair.TokenA, pair.TokenB
	if order.Side == types.OrderSideBuy {
		tokenIn, tokenOut = tokenOut, tokenIn
	}
	amountOut, bonus, repurchaseFund, pair := k.swap(ctx, order.FeeRate, pair, tokenIn, realMaxAmountIn, false)

	swapEvent := types.NewEventSwap(order.From, order.OrderID, pair.DexID, pair.TokenA, pair.TokenB, tokenIn,
		pair.TokenAAmount, pair.TokenBAmount, realMaxAmountIn, amountOut)

	flows := make([]sdk.Flow, 0, 3)
	if realMaxAmountIn.IsPositive() {
		_, flow, _ := k.tk.SubCoinHold(ctx, order.From, sdk.NewCoin(tokenIn.String(), realMaxAmountIn))
		flows = append(flows, flow)
	}

	if amountOut.IsPositive() {
		_, flow, _ := k.tk.AddCoin(ctx, order.Receiver, sdk.NewCoin(tokenOut.String(), amountOut))
		flows = append(flows, flow)
	}

	if bonus.IsPositive() {
		_, flow, _ := k.tk.AddCoin(ctx, order.Referer, sdk.NewCoin(tokenIn.String(), bonus))
		flows = append(flows, flow)
	}

	if repurchaseFund.IsPositive() {
		k.addRepurchaseFunds(ctx, sdk.NewCoins(sdk.NewCoin(tokenIn.String(), repurchaseFund)))
	}

	order.LockedFund = order.LockedFund.Sub(realMaxAmountIn)
	if order.LockedFund.IsZero() {
		order.Status = types.OrderStatusFilled
		order.FinishedTime = ctx.BlockTime().Unix()
	} else {
		order.Status = types.OrderStatusPartiallyFilled
	}
	k.saveOrder(ctx, order)

	return flows, swapEvent, pair, priceSuitable
}

func (k Keeper) RepurchaseAndBurn(ctx sdk.Context) sdk.Int {
	if ctx.BlockHeight()%k.RepurchaseDuration(ctx) != 0 {
		return sdk.ZeroInt()
	}

	store := ctx.KVStore(k.storeKey)
	repurchaseFunds := k.getRepurchaseFunds(ctx)
	totalRepurchaseAmount := sdk.ZeroInt()
	repurchaseToken := k.RepurchaseToken(ctx)
	for _, coin := range repurchaseFunds {
		switch coin.Denom {
		case repurchaseToken:
			totalRepurchaseAmount = totalRepurchaseAmount.Add(coin.Amount)
		default:
			pair := k.GetTradingPair(ctx, 0, sdk.Symbol(coin.Denom), sdk.Symbol(repurchaseToken))
			if pair != nil {
				feeRate := k.getFeeRates(ctx, pair)
				amount, _, _, _ := k.swap(ctx, feeRate, pair, sdk.Symbol(coin.Denom), coin.Amount, true)
				totalRepurchaseAmount = totalRepurchaseAmount.Add(amount)
			}
		}

		store.Delete(types.RepurchaseFundKey(coin.Denom))
	}

	if totalRepurchaseAmount.IsPositive() {
		burnedCoins := sdk.NewCoins(sdk.NewCoin(repurchaseToken, totalRepurchaseAmount))
		k.tk.AddCoins(ctx, types.ModuleCUAddress, burnedCoins)
		k.sk.BurnCoins(ctx, types.ModuleName, burnedCoins)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRepurchase,
				sdk.NewAttribute(types.AttributeKeySymbol, repurchaseToken),
				sdk.NewAttribute(types.AttributeKeyAmount, totalRepurchaseAmount.String()),
			),
		})
	}
	return totalRepurchaseAmount
}

func (k Keeper) GetLiquidity(ctx sdk.Context, addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LiquidityKey(addr, dexID, tokenA, tokenB))
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}
	var d sdk.Int
	k.cdc.MustUnmarshalBinaryBare(bz, &d)
	return d
}

func (k Keeper) GetAddrUnfinishedOrders(ctx sdk.Context, addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol) []*types.Order {
	tokenA, tokenB, _ = k.SortTokens(ctx, tokenA, tokenB)
	prefix := types.UnfinishedOrderKeyPrefixWithPair(addr, dexID, tokenA, tokenB)
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	var orders []*types.Order
	for ; iter.Valid(); iter.Next() {
		orderID := types.GetOrderIDFromUnfinishedOrderKey(iter.Key())
		order := k.GetOrder(ctx, orderID)
		orders = append(orders, order)
	}
	iter.Close()
	sort.Sort(sort.Reverse(types.OrderByCreatedTime(orders)))
	return orders
}

func (k Keeper) getAddrAllLiquidity(ctx sdk.Context, addr sdk.CUAddress, dexID *uint32) []*types.AddrLiquidity {
	var ret []*types.AddrLiquidity
	var prefix []byte
	if dexID == nil {
		prefix = types.AddrLiquidityKeyPrefix(addr)
	} else {
		prefix = types.AddrLiquidityKeyPrefixWithDexID(addr, *dexID)
	}
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	for ; iter.Valid(); iter.Next() {
		dexID, tokenA, tokenB := types.DecodeLiquidityKey(iter.Key())
		pair := k.GetTradingPair(ctx, dexID, tokenA, tokenB)
		var liquidity sdk.Int
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &liquidity)
		ret = append(ret, types.NewAddrLiquidity(pair, liquidity))
	}
	iter.Close()
	return ret
}

func (k Keeper) saveLiquidity(ctx sdk.Context, addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol, liquidity sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(liquidity)
	store.Set(types.LiquidityKey(addr, dexID, tokenA, tokenB), bz)
}

func (k Keeper) GetReferer(ctx sdk.Context, addr sdk.CUAddress) (ret sdk.CUAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RefererKey(addr))
	if len(bz) == 0 {
		return nil
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &ret)
	return
}

func (k Keeper) BindReferer(ctx sdk.Context, addr, referer sdk.CUAddress) {
	if referer.Equals(addr) {
		return
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(referer)
	store.Set(types.RefererKey(addr), bz)
}

func (k Keeper) getRepurchaseFund(ctx sdk.Context, symbol string) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RepurchaseFundKey(symbol))
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}
	var amount sdk.Int
	k.cdc.MustUnmarshalBinaryBare(bz, &amount)
	return amount
}

func (k Keeper) getRepurchaseFunds(ctx sdk.Context) sdk.Coins {
	funds := sdk.NewCoins()
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.RepurchaseFundKeyPrefix)
	for ; iter.Valid(); iter.Next() {
		var amount sdk.Int
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &amount)
		symbol := types.GetSymbolFromRepurchaseFundKey(iter.Key())
		funds = funds.Add(sdk.NewCoins(sdk.NewCoin(symbol, amount)))
	}
	iter.Close()
	return funds
}

func (k Keeper) addRepurchaseFunds(ctx sdk.Context, coins sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	for _, coin := range coins {
		amount := coin.Amount.Add(k.getRepurchaseFund(ctx, coin.Denom))
		bz := k.cdc.MustMarshalBinaryBare(amount)
		store.Set(types.RepurchaseFundKey(coin.Denom), bz)
	}
}
