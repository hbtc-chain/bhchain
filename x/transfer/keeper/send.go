package keeper

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) (sdk.Result, sdk.Error)
	SendCoin(ctx sdk.Context, fromAddr sdk.CUAddress, toAddr sdk.CUAddress, amt sdk.Coin) (sdk.Result, []sdk.Flow, sdk.Error)
	SendCoins(ctx sdk.Context, fromAddr sdk.CUAddress, toAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, []sdk.Flow, sdk.Error)

	SubCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	SubCoinsHold(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	AddCoinsHold(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	LockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)
	UnlockCoin(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coin) ([]sdk.Flow, sdk.Error)

	IsSendEnabled(ctx sdk.Context) bool
	SetSendEnabled(ctx sdk.Context, enabled bool)

	GetBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string) sdk.Int
	GetAllBalance(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins

	GetHoldBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string) sdk.Int
	GetAllHoldBalance(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins

	Codespace() sdk.CodespaceType

	BlacklistedAddr(addr sdk.CUAddress) bool
}

func (keeper BaseKeeper) BlacklistedAddr(addr sdk.CUAddress) bool {
	return keeper.blacklistedAddrs[addr.String()]
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseKeeper) InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) (sdk.Result, sdk.Error) {
	// Safety check ensuring that when sending coins the keeper must maintain the
	// Check supply invariant and validity of Coins.
	result := sdk.Result{}
	if err := types.ValidateInputsOutputs(inputs, outputs); err != nil {
		return err.Result(), err
	}
	var flows []sdk.Flow
	for _, in := range inputs {
		_, inFlows, err := keeper.SubCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err.Result(), err
		}

		flows = append(flows, inFlows...)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMultiTransfer,
				sdk.NewAttribute(types.AttributeKeySender, in.Address.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, in.Coins.String()),
			),
		)
	}

	for _, out := range outputs {
		_, outFlows, err := keeper.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err.Result(), err
		}
		flows = append(flows, outFlows...)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMultiTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, out.Address.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, out.Coins.String()),
			),
		)
	}

	if len(flows) > 0 {
		receipt := keeper.rk.NewReceipt(sdk.CategoryTypeMultiTransfer, flows)
		keeper.rk.SaveReceiptToResult(receipt, &result)
	}

	return result, nil
}

func (keeper BaseKeeper) SendCoin(ctx sdk.Context, from, to sdk.CUAddress, coin sdk.Coin) (sdk.Result, []sdk.Flow, sdk.Error) {
	_, subFlow, err := keeper.SubCoin(ctx, from, coin)
	if err != nil {
		return err.Result(), nil, err
	}
	_, addFlow, err := keeper.AddCoin(ctx, to, coin)
	if err != nil {
		return err.Result(), nil, err
	}

	flows := []sdk.Flow{subFlow, addFlow}
	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result, flows, nil
}

func (keeper BaseKeeper) SendCoins(ctx sdk.Context, from, to sdk.CUAddress, coins sdk.Coins) (sdk.Result, []sdk.Flow, sdk.Error) {
	var flows []sdk.Flow
	for _, coin := range coins {
		_, fls, err := keeper.SendCoin(ctx, from, to, coin)
		if err != nil {
			return err.Result(), nil, err
		}
		flows = append(flows, fls...)
	}

	result := sdk.Result{}
	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeTransfer, flows)
	keeper.rk.SaveReceiptToResult(receipt, &result)

	return result, flows, nil
}

func (keeper BaseKeeper) GetBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.BalanceKey(addr, symbol))
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}
	var balance sdk.Int
	keeper.cdc.MustUnmarshalBinaryBare(bz, &balance)
	return balance
}

func (keeper BaseKeeper) GetHoldBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.HoldBalanceKey(addr, symbol))
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}
	var balance sdk.Int
	keeper.cdc.MustUnmarshalBinaryBare(bz, &balance)
	return balance
}

func (keeper BaseKeeper) GetAllBalance(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins {
	var ret sdk.Coins
	store := ctx.KVStore(keeper.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.BalanceKeyPrefix(addr))
	for ; iter.Valid(); iter.Next() {
		var balance sdk.Int
		keeper.cdc.MustUnmarshalBinaryBare(iter.Value(), &balance)
		ret = ret.Add(sdk.NewCoins(sdk.NewCoin(types.GetSymbolFromBalanceKey(iter.Key()), balance)))
	}
	return ret
}

func (keeper BaseKeeper) GetAllHoldBalance(ctx sdk.Context, addr sdk.CUAddress) sdk.Coins {
	var ret sdk.Coins
	store := ctx.KVStore(keeper.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.HoldBalanceKeyPrefix(addr))
	for ; iter.Valid(); iter.Next() {
		var balance sdk.Int
		keeper.cdc.MustUnmarshalBinaryBare(iter.Value(), &balance)
		ret = ret.Add(sdk.NewCoins(sdk.NewCoin(types.GetSymbolFromHoldBalanceKey(iter.Key()), balance)))
	}
	return ret
}

func (keeper BaseKeeper) AddCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error) {
	if !coin.IsValid() {
		return coin, sdk.BalanceFlow{}, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coin %s", coin.String()))
	}

	before := keeper.GetBalance(ctx, addr, coin.Denom)
	after := before.Add(coin.Amount)
	keeper.setBalance(ctx, addr, coin.Denom, after)
	return sdk.NewCoin(coin.Denom, after), sdk.BalanceFlow{
		CUAddress:             addr,
		Symbol:                sdk.Symbol(coin.Denom),
		PreviousBalance:       before,
		BalanceChange:         coin.Amount,
		PreviousBalanceOnHold: sdk.ZeroInt(),
		BalanceOnHoldChange:   sdk.ZeroInt(),
	}, nil
}

func (keeper BaseKeeper) AddCoins(ctx sdk.Context, addr sdk.CUAddress, coins sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error) {
	if !coins.IsValid() {
		return coins, nil, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins %s", coins.String()))
	}
	var (
		newCoins sdk.Coins
		flows    []sdk.Flow
	)

	for _, coin := range coins {
		newCoin, flow, err := keeper.AddCoin(ctx, addr, coin)
		if err != nil {
			return coins, nil, err
		}
		newCoins = newCoins.Add(sdk.NewCoins(newCoin))
		flows = append(flows, flow)
	}
	return newCoins, flows, nil
}

func (keeper BaseKeeper) SubCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error) {
	if !coin.IsValid() {
		return coin, sdk.BalanceFlow{}, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coin %s", coin.String()))
	}

	before := keeper.GetBalance(ctx, addr, coin.Denom)
	if before.LT(coin.Amount) {
		return coin, sdk.BalanceFlow{}, sdk.ErrInsufficientCoins(fmt.Sprintf("token %s balance not enough, has %s, need %s", coin.Denom, before.String(), coin.Amount.String()))
	}
	after := before.Sub(coin.Amount)
	keeper.setBalance(ctx, addr, coin.Denom, after)
	return sdk.NewCoin(coin.Denom, after), sdk.BalanceFlow{
		CUAddress:             addr,
		Symbol:                sdk.Symbol(coin.Denom),
		PreviousBalance:       before,
		BalanceChange:         coin.Amount.Neg(),
		PreviousBalanceOnHold: sdk.ZeroInt(),
		BalanceOnHoldChange:   sdk.ZeroInt(),
	}, nil
}

func (keeper BaseKeeper) SubCoins(ctx sdk.Context, addr sdk.CUAddress, coins sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error) {
	if !coins.IsValid() {
		return coins, nil, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins %s", coins.String()))
	}
	var (
		newCoins sdk.Coins
		flows    []sdk.Flow
	)

	for _, coin := range coins {
		newCoin, flow, err := keeper.SubCoin(ctx, addr, coin)
		if err != nil {
			return coins, nil, err
		}
		newCoins = newCoins.Add(sdk.NewCoins(newCoin))
		flows = append(flows, flow)
	}
	return newCoins, flows, nil
}

func (keeper BaseKeeper) AddCoinHold(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error) {
	if !coin.IsValid() {
		return coin, sdk.BalanceFlow{}, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coin %s", coin.String()))
	}

	before := keeper.GetHoldBalance(ctx, addr, coin.Denom)
	after := before.Add(coin.Amount)
	keeper.setHoldBalance(ctx, addr, coin.Denom, after)
	return sdk.NewCoin(coin.Denom, after), sdk.BalanceFlow{
		CUAddress:             addr,
		Symbol:                sdk.Symbol(coin.Denom),
		PreviousBalanceOnHold: before,
		BalanceOnHoldChange:   coin.Amount,
		PreviousBalance:       sdk.ZeroInt(),
		BalanceChange:         sdk.ZeroInt(),
	}, nil
}

func (keeper BaseKeeper) AddCoinsHold(ctx sdk.Context, addr sdk.CUAddress, coins sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error) {
	if !coins.IsValid() {
		return coins, nil, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins %s", coins.String()))
	}
	var (
		newCoins sdk.Coins
		flows    []sdk.Flow
	)

	for _, coin := range coins {
		newCoin, flow, err := keeper.AddCoinHold(ctx, addr, coin)
		if err != nil {
			return coins, nil, err
		}
		newCoins = newCoins.Add(sdk.NewCoins(newCoin))
		flows = append(flows, flow)
	}
	return newCoins, flows, nil
}

func (keeper BaseKeeper) SubCoinHold(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) (sdk.Coin, sdk.Flow, sdk.Error) {
	if !coin.IsValid() {
		return coin, sdk.BalanceFlow{}, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coin %s", coin.String()))
	}

	before := keeper.GetHoldBalance(ctx, addr, coin.Denom)
	if before.LT(coin.Amount) {
		return coin, sdk.BalanceFlow{}, sdk.ErrInsufficientCoins(fmt.Sprintf("token %s balance not enough, has %s, need %s", coin.Denom, before.String(), coin.Amount.String()))
	}
	after := before.Sub(coin.Amount)
	keeper.setHoldBalance(ctx, addr, coin.Denom, after)
	return sdk.NewCoin(coin.Denom, after), sdk.BalanceFlow{
		CUAddress:             addr,
		Symbol:                sdk.Symbol(coin.Denom),
		PreviousBalanceOnHold: before,
		BalanceOnHoldChange:   coin.Amount.Neg(),
		PreviousBalance:       sdk.ZeroInt(),
		BalanceChange:         sdk.ZeroInt(),
	}, nil
}

func (keeper BaseKeeper) SubCoinsHold(ctx sdk.Context, addr sdk.CUAddress, coins sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error) {
	if !coins.IsValid() {
		return coins, nil, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins %s", coins.String()))
	}
	var (
		newCoins sdk.Coins
		flows    []sdk.Flow
	)

	for _, coin := range coins {
		newCoin, flow, err := keeper.SubCoinHold(ctx, addr, coin)
		if err != nil {
			return coins, nil, err
		}
		newCoins = newCoins.Add(sdk.NewCoins(newCoin))
		flows = append(flows, flow)
	}
	return newCoins, flows, nil
}

func (keeper BaseKeeper) LockCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) ([]sdk.Flow, sdk.Error) {
	var flows []sdk.Flow

	_, flow, err := keeper.SubCoin(ctx, addr, coin)
	if err != nil {
		return nil, err
	}
	flows = append(flows, flow)

	_, flow, err = keeper.AddCoinHold(ctx, addr, coin)
	if err != nil {
		return nil, err
	}
	flows = append(flows, flow)
	return flows, nil
}

func (keeper BaseKeeper) LockCoins(ctx sdk.Context, addr sdk.CUAddress, coins sdk.Coins) ([]sdk.Flow, sdk.Error) {
	if !coins.IsValid() {
		return nil, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins %s", coins.String()))
	}
	var flows []sdk.Flow
	for _, coin := range coins {
		fls, err := keeper.LockCoin(ctx, addr, coin)
		if err != nil {
			return nil, err
		}
		flows = append(flows, fls...)
	}
	return flows, nil
}

func (keeper BaseKeeper) UnlockCoin(ctx sdk.Context, addr sdk.CUAddress, coin sdk.Coin) ([]sdk.Flow, sdk.Error) {
	var flows []sdk.Flow
	_, flow, err := keeper.SubCoinHold(ctx, addr, coin)
	if err != nil {
		return nil, err
	}
	flows = append(flows, flow)

	_, flow, err = keeper.AddCoin(ctx, addr, coin)
	if err != nil {
		return nil, err
	}
	flows = append(flows, flow)
	return flows, nil
}

func (keeper BaseKeeper) UnlockCoins(ctx sdk.Context, addr sdk.CUAddress, coins sdk.Coins) ([]sdk.Flow, sdk.Error) {
	if !coins.IsValid() {
		return nil, sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins %s", coins.String()))
	}
	var flows []sdk.Flow
	for _, coin := range coins {
		fls, err := keeper.UnlockCoin(ctx, addr, coin)
		if err != nil {
			return nil, err
		}
		flows = append(flows, fls...)
	}
	return flows, nil
}

func (keeper BaseKeeper) setBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string, balance sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.BalanceKey(addr, symbol)
	if balance.IsZero() {
		store.Delete(key)
	} else {
		bz := keeper.cdc.MustMarshalBinaryBare(balance)
		store.Set(key, bz)
	}
}

func (keeper BaseKeeper) setHoldBalance(ctx sdk.Context, addr sdk.CUAddress, symbol string, balance sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.HoldBalanceKey(addr, symbol)
	if balance.IsZero() {
		store.Delete(key)
	} else {
		bz := keeper.cdc.MustMarshalBinaryBare(balance)
		store.Set(key, bz)
	}
}
