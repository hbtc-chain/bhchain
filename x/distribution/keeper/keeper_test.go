package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestSetWithdrawAddr(t *testing.T) {
	ctx, _, _, keeper, _, _ := CreateTestInputDefault(t, false, 1000)

	keeper.SetWithdrawAddrEnabled(ctx, false)

	err := keeper.SetWithdrawAddr(ctx, delAddr1, delAddr2)
	require.NotNil(t, err)

	keeper.SetWithdrawAddrEnabled(ctx, true)

	err = keeper.SetWithdrawAddr(ctx, delAddr1, delAddr2)
	require.Nil(t, err)

	keeper.blacklistedAddrs[distrAcc.GetAddress().String()] = true
	require.Error(t, keeper.SetWithdrawAddr(ctx, delAddr1, distrAcc.GetAddress()))
}

func TestWithdrawValidatorCommission(t *testing.T) {
	ctx, _, tk, keeper, sk, _ := CreateTestInputDefault(t, false, 1000)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec(sk.BondDenom(ctx), sdk.NewDec(3).Quo(sdk.NewDec(2))),
	}

	// set module CU coins
	distrAcc := keeper.GetDistributionAccount(ctx)
	tk.AddCoins(ctx, distrAcc.GetAddress(), sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(2)),
		sdk.NewCoin(sk.BondDenom(ctx), sdk.NewInt(2)),
	))

	keeper.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// check initial balance
	balance := tk.GetAllBalance(ctx, sdk.CUAddress(valOpAddr3))
	expTokens := sdk.TokensFromConsensusPower(1000)
	expCoins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), expTokens))
	require.Equal(t, expCoins, balance)

	// set outstanding rewards
	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr3, valCommission)

	// set commission
	keeper.SetValidatorAccumulatedCommission(ctx, valOpAddr3, valCommission)

	// withdraw commission
	keeper.WithdrawValidatorCommission(ctx, valOpAddr3)

	// check balance increase
	balance = tk.GetAllBalance(ctx, sdk.CUAddress(valOpAddr3))
	require.Equal(t, sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(1)),
		sdk.NewCoin(sk.BondDenom(ctx), expTokens.AddRaw(1)),
	), balance)

	// check remainder
	remainder := keeper.GetValidatorAccumulatedCommission(ctx, valOpAddr3)
	require.True(t, remainder.IsEqual(sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(1).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec(sk.BondDenom(ctx), sdk.NewDec(1).Quo(sdk.NewDec(2))),
	}))

	require.True(t, true)
}

func TestGetTotalRewards(t *testing.T) {
	ctx, _, _, keeper, sk, _ := CreateTestInputDefault(t, false, 1000)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec(sk.BondDenom(ctx), sdk.NewDec(3).Quo(sdk.NewDec(2))),
	}

	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr1, valCommission)
	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr2, valCommission)

	expectedRewards := valCommission.MulDec(sdk.NewDec(2))
	totalRewards := keeper.GetTotalRewards(ctx)

	require.True(t, expectedRewards.IsEqual(totalRewards))
}
