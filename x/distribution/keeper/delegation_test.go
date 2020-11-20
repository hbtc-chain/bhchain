package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking"
)

func TestCalculateRewardsBasic(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// historical count should be 2 (once for validator init, once for delegation init)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// historical count should be 2 still
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// allocate some rewards
	initial := int64(100000)
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 2)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsAfterSlash(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valPower := int64(1000000)
	valTokens := sdk.TokensFromConsensusPower(valPower)
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, sdk.OneInt())
	got := sh(ctx, msg)
	require.True(t, got.IsOK(), "%v", got)

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// retrieve validator
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, initial.ToDec()}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial.QuoRaw(2).ToDec()}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial.QuoRaw(2).ToDec()}},
		k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	power := int64(1000000)
	valTokens := sdk.TokensFromConsensusPower(power)
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, initial.ToDec()}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50% again
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial.ToDec()}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial.ToDec()}},
		k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegator(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := int64(200000)
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.CUAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	require.True(t, sh(ctx, msg2).IsOK())
	del2 := sk.Delegation(ctx, sdk.CUAddress(valOpAddr2), valOpAddr1)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	balancePower := int64(10000000)
	balanceTokens := sdk.TokensFromConsensusPower(balancePower)
	ctx, _, tk, k, sk, _ := CreateTestInputDefault(t, false, balancePower)
	sh := staking.NewHandler(sk.Keeper)

	// set module CU coins
	distrAcc := k.GetDistributionAccount(ctx)
	tk.AddCoins(ctx, distrAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens)))

	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator with 50% commission
	power := int64(1000000)
	valTokens := sdk.TokensFromConsensusPower(power)
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(
		valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
		staking.Description{}, commission, sdk.OneInt(),
	)
	require.True(t, sh(ctx, msg).IsOK())

	// assert correct initial balance
	expTokens := balanceTokens.Sub(valTokens)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, expTokens)},
		tk.GetAllBalance(ctx, sdk.CUAddress(valOpAddr1)),
	)

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(100000)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	k.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (initial + latest for delegation)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// withdraw rewards
	_, _, err := k.WithdrawDelegationRewards(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)
	require.Nil(t, err)

	// historical count should still be 2 (added one record, cleared one)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// assert correct balance
	exp := balanceTokens.Sub(valTokens).Add(initial.QuoRaw(2))
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		tk.GetAllBalance(ctx, sdk.CUAddress(valOpAddr1)),
	)

	// withdraw commission
	_, _, err = k.WithdrawValidatorCommission(ctx, valOpAddr1)
	require.Nil(t, err)

	// assert correct balance
	exp = balanceTokens.Sub(valTokens).Add(initial)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		tk.GetAllBalance(ctx, sdk.CUAddress(valOpAddr1)),
	)
}

func TestCalculateRewardsAfterManySlashesInSameBlock(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	power := int64(1000000)
	valTokens := sdk.TokensFromConsensusPower(power)
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(100000).ToDec()
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, initial}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))

	// slash the validator by 50% again
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegatorMultiSlash(t *testing.T) {
	ctx, _, _, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	power := int64(1000000)
	valTokens := sdk.TokensFromConsensusPower(power)
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(300000).ToDec()
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, initial}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// second delegation
	delTokens := sdk.TokensFromConsensusPower(1000000)
	msg2 := staking.NewMsgDelegate(sdk.CUAddress(valOpAddr2), valOpAddr1,
		sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	require.True(t, sh(ctx, msg2).IsOK())
	del2 := sk.Delegation(ctx, sdk.CUAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator again
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 2/3 initial (half initial first period, 1/6 initial second period)
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial.QuoInt64(2).Add(initial.QuoInt64(6))}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be initial / 3
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial.QuoInt64(3)}}, rewards)

	// commission should be equal to initial (twice 50% commission, unaffected by slashing)
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, initial}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegatorMultWithdraw(t *testing.T) {
	ctx, _, tk, k, sk, _ := CreateTestInputDefault(t, false, 10000000)
	sh := staking.NewHandler(sk.Keeper)
	initial := int64(20)

	// set module CU coins
	distrAcc := k.GetDistributionAccount(ctx)
	tk.AddCoins(ctx, distrAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))

	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	totalRewards := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(initial*2))}
	tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(initial))}

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (validator init, delegation init)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.CUAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	require.True(t, sh(ctx, msg2).IsOK())

	// historical count should be 3 (second delegation init)
	require.Equal(t, uint64(3), k.GetValidatorHistoricalReferenceCount(ctx))

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)
	del2 := sk.Delegation(ctx, sdk.CUAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk.Keeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws
	_, _, err := k.WithdrawDelegationRewards(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)
	require.Nil(t, err)

	// second delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.CUAddress(valOpAddr2), valOpAddr1)
	require.Nil(t, err)

	// historical count should be 3 (validator init + two delegations)
	require.Equal(t, uint64(3), k.GetValidatorHistoricalReferenceCount(ctx))

	// validator withdraws commission

	_, _, err = k.WithdrawValidatorCommission(ctx, valOpAddr1)
	require.Nil(t, err)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())

	totalRewards = totalRewards.Add(tokens)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws again
	k.WithdrawDelegationRewards(ctx, sdk.CUAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 4)}}, rewards)

	// commission should be half initial
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))

	totalRewards = k.GetValidatorOutstandingRewards(ctx, valOpAddr1).Add(tokens)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 2)}}, rewards)

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
}
