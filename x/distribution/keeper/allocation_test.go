package keeper

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking"
)

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())

	require.True(t, sh(ctx, msg).IsOK())
	val := sk.Validator(ctx, valOpAddr1)

	// allocate tokens
	tokens := sdk.DecCoins{
		{sdk.DefaultBondDenom, sdk.NewDec(10)},
	}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// check commission
	expected := sdk.DecCoins{
		{sdk.DefaultBondDenom, sdk.NewDec(5)},
	}
	require.Equal(t, expected, k.GetValidatorAccumulatedCommission(ctx, val.GetOperator()))

	// check current rewards
	require.Equal(t, expected, k.GetValidatorCurrentRewards(ctx, val.GetOperator()).Rewards)
}

func TestAllocateTokensToManyValidators(t *testing.T) {
	ctx, ak, tk, k, sk, supplyKeeper := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create second validator with 0% commission
	commission = staking.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)
	tk.AddCoins(ctx, feeCollector.GetAddress(), fees)

	ak.SetCU(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}

	mockEpoch := sdk.Epoch{}
	mockEpoch.KeyNodeSet = append(mockEpoch.KeyNodeSet, sdk.CUAddress(valOpAddr1), sdk.CUAddress(valOpAddr2))
	sk.On("GetCurrentEpoch", mock.Anything).Return(mockEpoch)
	k.AllocateTokens(ctx, 200, 200, valConsAddr2, votes)

	// 98 outstanding rewards (100 less 2 to community pool)
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(495, 1)}}, k.GetValidatorOutstandingRewards(ctx, valOpAddr1))
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(505, 1)}}, k.GetValidatorOutstandingRewards(ctx, valOpAddr2))
	// 2 community pool coins
	require.Equal(t, sdk.DecCoins(nil), k.GetFeePool(ctx).CommunityPool)
	// 50% commission for first proposer, 49.5 * 50% = 24.75
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(2475, 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
	// zero commission for second proposer
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	// just staking.proportional for first proposer less commission = 49.5 * 50% = 24.75
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(2475, 2)}}, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards)
	// proposer reward + staking.proportional for second proposer = 50.5 * 100%
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(505, 1)}}, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards)
}

func TestAllocateTokensTruncation(t *testing.T) {
	communityTax := sdk.NewDec(0)
	ctx, ak, tk, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000000, communityTax)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 10% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(110)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create second validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create third validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr3, valConsPk3,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   11,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}
	abciValС := abci.Validator{
		Address: valConsPk3.Address(),
		Power:   10,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsZero())
	require.True(t, k.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(634195840)))

	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)

	tk.AddCoins(ctx, feeCollector.GetAddress(), fees)

	ak.SetCU(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValС,
			SignedLastBlock: true,
		},
	}

	sk.On("GetCurrentEpoch", mock.Anything).Return(
		sdk.Epoch{
			KeyNodeSet: []sdk.CUAddress{sdk.CUAddress(valOpAddr1), sdk.CUAddress(valOpAddr2)},
		})
	k.AllocateTokens(ctx, 31, 31, valConsAddr2, votes)

	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsValid())
}
