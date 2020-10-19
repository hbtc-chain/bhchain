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
	cuKeeper      types.CUKeeper
	rk            types.ReceiptKeeper
	sk            types.SupplyKeeper
	marketManager *orderbook.Manager
	paramstore    params.Subspace
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, tokenKeeper types.TokenKeeper, cuKeeper types.CUKeeper,
	rk types.ReceiptKeeper, sk types.SupplyKeeper, tk types.TransferKeeper, paramstore params.Subspace) Keeper {
	k := Keeper{
		storeKey:    key,
		cdc:         cdc,
		tokenKeeper: tokenKeeper,
		cuKeeper:    cuKeeper,
		rk:          rk,
		sk:          sk,
		tk:          tk,
		paramstore:  paramstore.WithKeyTable(ParamKeyTable()),
	}
	k.marketManager = orderbook.NewManager(k)
	return k
}

func (k Keeper) CheckSymbol(ctx sdk.Context, symbol sdk.Symbol) sdk.Result {
	tokenInfo := k.tokenKeeper.GetTokenInfo(ctx, symbol)
	if tokenInfo == nil {
		return sdk.ErrUnSupportToken(fmt.Sprintf("token %s does not exist", symbol.String())).Result()
	}
	if !tokenInfo.IsSendEnabled {
		return sdk.ErrUnSupportToken(fmt.Sprintf("token %s is not enable to send", symbol)).Result()
	}
	return sdk.Result{}
}

func (k Keeper) AddLiquidity(ctx sdk.Context, from sdk.CUAddress, tokenA, tokenB sdk.Symbol, minTokenAAmount, minTokenBAmount sdk.Int) sdk.Result {
	if tokenA > tokenB {
		tokenA, tokenB = tokenB, tokenA
		minTokenAAmount, minTokenBAmount = minTokenBAmount, minTokenAAmount
	}
	pair := k.GetTradingPair(ctx, tokenA, tokenB)
	var needTokenA, needTokenB, liquidity sdk.Int
	if pair == nil {
		product := big.NewInt(0).Mul(minTokenAAmount.BigInt(), minTokenBAmount.BigInt())
		sqr := sdk.NewIntFromBigInt(big.NewInt(0).Sqrt(product))
		minimumLiquidity := k.MinimumLiquidity(ctx)
		liquidity = sqr.Sub(minimumLiquidity) // lock MinimumLiquidity permanently
		if !liquidity.IsPositive() {
			return sdk.ErrInvalidAmount("insufficient liquidity").Result()
		}
		pair = types.NewTradingPair(tokenA, tokenB, minimumLiquidity)
		needTokenA, needTokenB = minTokenAAmount, minTokenBAmount
	} else {
		needTokenB = mulAndDiv(minTokenAAmount, pair.TokenBAmount, pair.TokenAAmount)
		if needTokenB.LT(minTokenBAmount) {
			needTokenA = mulAndDiv(minTokenBAmount, pair.TokenAAmount, pair.TokenBAmount)
			if needTokenA.LT(minTokenAAmount) {
				return sdk.ErrInvalidAmount("insufficient min token amount").Result()
			}
			needTokenB = minTokenBAmount
		} else {
			needTokenA = minTokenAAmount
		}
		liquidity = sdk.MinInt(mulAndDiv(needTokenA, pair.TotalLiquidity, pair.TokenAAmount), mulAndDiv(needTokenB, pair.TotalLiquidity, pair.TokenBAmount))
	}
	cu := k.cuKeeper.GetCU(ctx, from)
	have := cu.GetCoins()
	need := sdk.NewCoins(sdk.NewCoin(tokenA.String(), needTokenA), sdk.NewCoin(tokenB.String(), needTokenB))
	if have.AmountOf(tokenA.String()).LT(needTokenA) || have.AmountOf(tokenB.String()).LT(needTokenB) {
		return sdk.ErrInsufficientFunds(fmt.Sprintf("insufficient funds, need:%v, have:%v", need, have)).Result()
	}
	cu.SubCoins(need)
	k.cuKeeper.SetCU(ctx, cu)

	pair.TokenAAmount = pair.TokenAAmount.Add(needTokenA)
	pair.TokenBAmount = pair.TokenBAmount.Add(needTokenB)
	pair.TotalLiquidity = pair.TotalLiquidity.Add(liquidity)
	k.saveTradingPair(ctx, pair)

	addrLiquidity := k.GetLiquidity(ctx, tokenA, tokenB, from)
	addrLiquidity = addrLiquidity.Add(liquidity)
	k.saveLiquidity(ctx, tokenA, tokenB, from, addrLiquidity)

	k.onUpdateLiquidity(ctx, tokenA, tokenB, from, liquidity)

	flows := make([]sdk.Flow, 0, 2)
	for _, flow := range cu.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)
	eventLiquidity := types.NewEventLiquidity(from, tokenA, tokenB, pair.TokenAAmount, pair.TokenBAmount, needTokenA, needTokenB)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddLiquidity,
			sdk.NewAttribute(types.AttributeKeyLiquidity, eventLiquidity.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) RemoveLiquidity(ctx sdk.Context, from sdk.CUAddress, tokenA, tokenB sdk.Symbol, liquidity sdk.Int) sdk.Result {
	if tokenA > tokenB {
		tokenA, tokenB = tokenB, tokenA
	}
	pair := k.GetTradingPair(ctx, tokenA, tokenB)
	if pair == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("no trading pair of %s-%s", tokenA.String(), tokenB.String())).Result()
	}
	addrLiquidity := k.GetLiquidity(ctx, tokenA, tokenB, from)
	if addrLiquidity.LT(liquidity) {
		return sdk.ErrInsufficientFunds(fmt.Sprintf("insufficient liquidity, has %s, need %s", addrLiquidity.String(), liquidity.String())).Result()
	}
	if liquidity.GT(pair.TotalLiquidity) {
		return sdk.ErrInternal("remove amount is larger than total liquidity").Result()
	}
	returnTokenA := mulAndDiv(liquidity, pair.TokenAAmount, pair.TotalLiquidity)
	returnTokenB := mulAndDiv(liquidity, pair.TokenBAmount, pair.TotalLiquidity)
	cu := k.cuKeeper.GetCU(ctx, from)
	returnCoins := sdk.NewCoins(sdk.NewCoin(tokenA.String(), returnTokenA), sdk.NewCoin(tokenB.String(), returnTokenB))
	cu.AddCoins(returnCoins)
	k.cuKeeper.SetCU(ctx, cu)

	pair.TokenAAmount = pair.TokenAAmount.Sub(returnTokenA)
	pair.TokenBAmount = pair.TokenBAmount.Sub(returnTokenB)
	pair.TotalLiquidity = pair.TotalLiquidity.Sub(liquidity)
	k.saveTradingPair(ctx, pair)

	addrLiquidity = addrLiquidity.Sub(liquidity)
	k.saveLiquidity(ctx, tokenA, tokenB, from, addrLiquidity)

	k.onUpdateLiquidity(ctx, tokenA, tokenB, from, liquidity.Neg())

	flows := make([]sdk.Flow, 0, 2)
	for _, flow := range cu.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)
	eventLiquidity := types.NewEventLiquidity(from, tokenA, tokenB, pair.TokenAAmount, pair.TokenBAmount, returnTokenA, returnTokenB)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveLiquidity,
			sdk.NewAttribute(types.AttributeKeyLiquidity, eventLiquidity.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) SwapExactIn(ctx sdk.Context, from, referer, receiver sdk.CUAddress, amountIn, minAmountOut sdk.Int,
	path []sdk.Symbol) sdk.Result {

	fromCU := k.cuKeeper.GetCU(ctx, from)
	have := fromCU.GetCoins()
	if have.AmountOf(path[0].String()).LT(amountIn) {
		return sdk.ErrInsufficientFunds(fmt.Sprintf("token %s is insufficient", path[0].String())).Result()
	}

	amountOut, err := k.getAmountOut(ctx, amountIn, path)
	if err != nil {
		return sdk.ErrInvalidTx(err.Error()).Result()
	}
	if amountOut.LT(minAmountOut) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("insufficient amount out, min: %s, got: %s", minAmountOut.String(), amountOut.String())).Result()
	}

	return k.DoSwap(ctx, from, referer, receiver, amountIn, path)
}

func (k Keeper) SwapExactOut(ctx sdk.Context, from, referer, receiver sdk.CUAddress, amountOut, maxAmountIn sdk.Int,
	path []sdk.Symbol) sdk.Result {

	fromCU := k.cuKeeper.GetCU(ctx, from)
	have := fromCU.GetCoins()
	if have.AmountOf(path[0].String()).LT(maxAmountIn) {
		return sdk.ErrInsufficientFunds(fmt.Sprintf("token %s is insufficient", path[0].String())).Result()
	}

	amountIn, err := k.getAmountIn(ctx, amountOut, path)
	if err != nil {
		return sdk.ErrInvalidTx(err.Error()).Result()
	}
	if amountIn.GT(maxAmountIn) {
		return sdk.ErrInvalidAmount(fmt.Sprintf("excessive amount in, max: %s, got: %s", maxAmountIn.String(), amountIn.String())).Result()
	}

	return k.DoSwap(ctx, from, referer, receiver, amountIn, path)
}

func (k Keeper) DoSwap(ctx sdk.Context, from, referer, receiver sdk.CUAddress, amountIn sdk.Int, path []sdk.Symbol) sdk.Result {
	fromCU := k.cuKeeper.GetCU(ctx, from)
	need := sdk.NewCoins(sdk.NewCoin(path[0].String(), amountIn))
	fromCU.SubCoins(need)
	k.cuKeeper.SetCU(ctx, fromCU)

	bonusCoins := sdk.NewCoins()
	repurchaseFunds := sdk.NewCoins()
	var swapEvents types.EventSwaps
	for i := 0; i < len(path)-1; i++ {
		pair := k.GetTradingPair(ctx, path[i], path[i+1])
		amountOut, bonus, repurchaseFund, updatedPair := k.swap(ctx, pair, path[i], amountIn, false)

		bonusCoins = bonusCoins.Add(sdk.NewCoins(sdk.NewCoin(path[i].String(), bonus)))
		if repurchaseFund.IsPositive() {
			repurchaseFunds = repurchaseFunds.Add(sdk.NewCoins(sdk.NewCoin(path[i].String(), repurchaseFund)))
		}

		swapEvents = append(swapEvents, types.NewEventSwap(from, "", pair.TokenA, pair.TokenB, path[i],
			updatedPair.TokenAAmount, updatedPair.TokenBAmount, amountIn, amountOut))

		amountIn = amountOut
	}

	receiverCU := k.cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, receiver)
	receiverCU.AddCoins(sdk.NewCoins(sdk.NewCoin(path[len(path)-1].String(), amountIn)))
	k.cuKeeper.SetCU(ctx, receiverCU)

	refererCU := k.cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, referer)
	refererCU.AddCoins(bonusCoins)
	k.cuKeeper.SetCU(ctx, refererCU)

	burnedAmount := k.repurchaseAndBurn(ctx, repurchaseFunds)

	flows := make([]sdk.Flow, 0, 4)
	for _, flow := range fromCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	for _, flow := range receiverCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	for _, flow := range refererCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	receipt := k.rk.NewReceipt(sdk.CategoryTypeOpenswap, flows)
	result := sdk.Result{}
	k.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSwap,
			sdk.NewAttribute(types.AttributeKeySwapResult, swapEvents.String()),
			sdk.NewAttribute(types.AttributeKeyBurned, burnedAmount.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func (k Keeper) LimitSwap(ctx sdk.Context, orderID string, from, referer, receiver sdk.CUAddress, amountIn sdk.Int, price sdk.Dec,
	baseSymbol, quoteSymbol sdk.Symbol, side int, expiredAt int64) sdk.Result {

	pair := k.GetTradingPair(ctx, baseSymbol, quoteSymbol)
	if pair == nil {
		return sdk.ErrInvalidTx(fmt.Sprintf("no trading pair of %s-%s", baseSymbol.String(), quoteSymbol.String())).Result()
	}

	fromCU := k.cuKeeper.GetCU(ctx, from)
	have := fromCU.GetCoins()
	symbol := baseSymbol
	if side == types.OrderSideBuy {
		symbol = quoteSymbol
	}
	need := sdk.NewCoins(sdk.NewCoin(symbol.String(), amountIn))
	if have.AmountOf(symbol.String()).LT(amountIn) {
		return sdk.ErrInsufficientFunds("insufficient funds").Result()
	}
	fromCU.SubCoins(need)
	fromCU.AddCoinsHold(need)
	k.cuKeeper.SetCU(ctx, fromCU)
	flows := make([]sdk.Flow, 0, 2)
	for _, flow := range fromCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}

	order := types.NewOrder(orderID, ctx.BlockTime().Unix(), expiredAt, from, referer, receiver,
		baseSymbol, quoteSymbol, price, amountIn, byte(side))
	k.CreateOrder(ctx, order)

	_, burned, balanceFlows, event, _ := k.limitSwap(ctx, order, pair)
	for _, flow := range balanceFlows {
		flows = append(flows, flow)
	}
	if order.LockedFund.IsZero() {
		k.finishOrderWithStatus(ctx, order, types.OrderStatusFilled)
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
				sdk.NewAttribute(types.AttributeKeyBurned, burned.String()),
			),
		})
		result.Events = append(result.Events, ctx.EventManager().Events()...)
	}

	return result

}

func (k Keeper) swap(ctx sdk.Context, pair *types.TradingPair, tokenIn sdk.Symbol, amountIn sdk.Int, isRepurchasing bool) (sdk.Int, sdk.Int, sdk.Int, *types.TradingPair) {

	amountInDec := sdk.NewDecFromInt(amountIn)
	refererBonus, fee, repurchaseFund := sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	if !isRepurchasing {
		refererBonus = amountInDec.Mul(k.RefererTransactionBonusRate(ctx)).TruncateInt()
		fee = amountInDec.Mul(k.FeeRate(ctx)).TruncateInt()
		repurchaseFund = amountInDec.Mul(k.RepurchaseRate(ctx)).TruncateInt()
		if !k.canRepurchase(ctx, tokenIn) {
			fee = fee.Add(repurchaseFund)
			repurchaseFund = sdk.ZeroInt()
		}
	}
	realAmountIn := amountIn.Sub(refererBonus).Sub(fee).Sub(repurchaseFund)

	var amountOut sdk.Int
	if pair.TokenA == tokenIn {
		amountOut = mulAndDiv(realAmountIn, pair.TokenBAmount, realAmountIn.Add(pair.TokenAAmount))
		pair.TokenAAmount = pair.TokenAAmount.Add(realAmountIn).Add(fee)
		pair.TokenBAmount = pair.TokenBAmount.Sub(amountOut)
	} else {
		amountOut = mulAndDiv(realAmountIn, pair.TokenAAmount, realAmountIn.Add(pair.TokenBAmount))
		pair.TokenBAmount = pair.TokenBAmount.Add(realAmountIn).Add(fee)
		pair.TokenAAmount = pair.TokenAAmount.Sub(amountOut)
	}
	k.saveTradingPair(ctx, pair)
	return amountOut, refererBonus, repurchaseFund, pair
}

func (k Keeper) limitSwap(ctx sdk.Context, order *types.Order, pair *types.TradingPair) (sdk.Int, sdk.Int, []sdk.Flow, *types.EventSwap, *types.TradingPair) {
	maxAmountIn := calLimitSwapAmount(order, pair)
	if maxAmountIn.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt(), nil, nil, pair
	}

	realAmountInCoeff := k.getRealInCoeff(ctx)
	realMaxMountIn := maxAmountIn.ToDec().Quo(realAmountInCoeff).TruncateInt()
	if realMaxMountIn.GT(order.LockedFund) {
		realMaxMountIn = order.LockedFund
	}

	tokenIn, tokenOut := pair.TokenA, pair.TokenB
	if order.Side == types.OrderSideBuy {
		tokenIn, tokenOut = tokenOut, tokenIn
	}
	amountOut, bonus, repurchaseFund, pair := k.swap(ctx, pair, tokenIn, realMaxMountIn, false)

	swapEvent := types.NewEventSwap(order.From, order.OrderID, pair.TokenA, pair.TokenB, tokenIn,
		pair.TokenAAmount, pair.TokenBAmount, realMaxMountIn, amountOut)

	fromCU := k.cuKeeper.GetCU(ctx, order.From)
	fromCU.SubCoinsHold(sdk.NewCoins(sdk.NewCoin(tokenIn.String(), realMaxMountIn)))
	k.cuKeeper.SetCU(ctx, fromCU)

	receiverCU := k.cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, order.Receiver)
	receiverCU.AddCoins(sdk.NewCoins(sdk.NewCoin(tokenOut.String(), amountOut)))
	k.cuKeeper.SetCU(ctx, receiverCU)

	refererCU := k.cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, order.Referer)
	refererCU.AddCoins(sdk.NewCoins(sdk.NewCoin(tokenIn.String(), bonus)))
	k.cuKeeper.SetCU(ctx, refererCU)

	burned := sdk.ZeroInt()
	if repurchaseFund.IsPositive() {
		burned = k.repurchaseAndBurn(ctx, sdk.NewCoins(sdk.NewCoin(tokenIn.String(), repurchaseFund)))
	}

	flows := make([]sdk.Flow, 0, 3)
	for _, flow := range fromCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	for _, flow := range receiverCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}
	for _, flow := range refererCU.GetBalanceFlows() {
		flows = append(flows, flow)
	}

	order.LockedFund = order.LockedFund.Sub(realMaxMountIn)
	order.Status = types.OrderStatusPartiallyFilled
	k.saveOrder(ctx, order)

	return realMaxMountIn, burned, flows, swapEvent, pair
}

func (k Keeper) repurchaseAndBurn(ctx sdk.Context, repurchaseFunds sdk.Coins) sdk.Int {
	totalRepurchaseAmount := sdk.ZeroInt()
	for _, coin := range repurchaseFunds {
		switch coin.Denom {
		case sdk.NativeDefiToken:
			totalRepurchaseAmount = totalRepurchaseAmount.Add(coin.Amount)
		case types.RepurchaseRoutingCoin:
			pair := k.GetTradingPair(ctx, types.RepurchaseRoutingCoin, sdk.NativeDefiToken)
			amount, _, _, _ := k.swap(ctx, pair, types.RepurchaseRoutingCoin, coin.Amount, true)
			totalRepurchaseAmount = totalRepurchaseAmount.Add(amount)
		default:
			pair := k.GetTradingPair(ctx, sdk.Symbol(coin.Denom), types.RepurchaseRoutingCoin)
			amount, _, _, _ := k.swap(ctx, pair, sdk.Symbol(coin.Denom), coin.Amount, true)
			pair = k.GetTradingPair(ctx, types.RepurchaseRoutingCoin, sdk.NativeDefiToken)
			amount, _, _, _ = k.swap(ctx, pair, types.RepurchaseRoutingCoin, amount, true)
			totalRepurchaseAmount = totalRepurchaseAmount.Add(amount)
		}
	}
	if totalRepurchaseAmount.IsPositive() {
		burnedCoins := sdk.NewCoins(sdk.NewCoin(sdk.NativeDefiToken, totalRepurchaseAmount))
		k.tk.AddCoins(ctx, types.ModuleCUAddress, burnedCoins)
		k.sk.BurnCoins(ctx, types.ModuleName, burnedCoins)
	}
	return totalRepurchaseAmount
}

func (k Keeper) GetTradingPair(ctx sdk.Context, tokenA, tokenB sdk.Symbol) *types.TradingPair {
	if tokenA > tokenB {
		tokenA, tokenB = tokenB, tokenA
	}
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TradingPairKey(tokenA, tokenB))
	if len(bz) == 0 {
		return nil
	}
	var pair types.TradingPair
	k.cdc.MustUnmarshalBinaryBare(bz, &pair)
	return &pair
}

func (k Keeper) getAllTradingPairs(ctx sdk.Context) []*types.TradingPair {
	var ret []*types.TradingPair
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.TradingPairKeyPrefix)
	for ; iter.Valid(); iter.Next() {
		var pair types.TradingPair
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &pair)
		ret = append(ret, &pair)
	}
	iter.Close()
	return ret
}

func (k Keeper) saveTradingPair(ctx sdk.Context, pair *types.TradingPair) {
	bz := k.cdc.MustMarshalBinaryBare(pair)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.TradingPairKey(pair.TokenA, pair.TokenB), bz)
}

// require tokenA < tokenB
func (k Keeper) GetLiquidity(ctx sdk.Context, tokenA, tokenB sdk.Symbol, addr sdk.CUAddress) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LiquidityKey(tokenA, tokenB, addr))
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}
	var d sdk.Int
	k.cdc.MustUnmarshalBinaryBare(bz, &d)
	return d
}

func (k Keeper) GetAddrUnfinishedOrders(ctx sdk.Context, tokenA, tokenB sdk.Symbol, addr sdk.CUAddress) []*types.Order {
	if tokenA > tokenB {
		tokenA, tokenB = tokenB, tokenA
	}
	prefix := types.GetUnfinishedOrderKeyPrefix(tokenA, tokenB, addr)
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

func (k Keeper) getAddrAllLiquidity(ctx sdk.Context, addr sdk.CUAddress) []*types.AddrLiquidity {
	var ret []*types.AddrLiquidity
	store := ctx.KVStore(k.storeKey)
	prefix := types.AddrLiquidityKeyPrefix(addr)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	for ; iter.Valid(); iter.Next() {
		tokenA, tokenB := types.DecodeTokensFromLiquidityKey(iter.Key(), prefix)
		pair := k.GetTradingPair(ctx, tokenA, tokenB)
		var liquidity sdk.Int
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &liquidity)
		ret = append(ret, types.NewAddrLiquidity(pair, liquidity))
	}
	iter.Close()
	return ret
}

func (k Keeper) saveLiquidity(ctx sdk.Context, tokenA, tokenB sdk.Symbol, addr sdk.CUAddress, d sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(d)
	store.Set(types.LiquidityKey(tokenA, tokenB, addr), bz)
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
