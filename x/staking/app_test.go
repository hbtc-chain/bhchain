package staking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/mock"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/staking/types"
	"github.com/hbtc-chain/bhchain/x/supply"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

// getMockApp returns an initialized mock application for this module.
func getMockApp(t *testing.T) (*mock.App, Keeper) {
	mApp := mock.NewApp()

	RegisterCodec(mApp.Cdc)
	supply.RegisterCodec(mApp.Cdc)
	receipt.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyTransfer := sdk.NewKVStoreKey(transfer.StoreKey)

	feeCollector := supply.NewEmptyModuleAccount(custodianunit.FeeCollectorName)
	notBondedPool := supply.NewEmptyModuleAccount(types.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(types.BondedPoolName, supply.Burner, supply.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollector.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true

	rk := mApp.ReceiptKeeper
	bankKeeper := transfer.NewBaseKeeper(mApp.Cdc, keyTransfer, mApp.CUKeeper, nil, nil, rk, nil, nil, mApp.ParamsKeeper.Subspace(transfer.DefaultParamspace), transfer.DefaultCodespace, blacklistedAddrs)
	maccPerms := map[string][]string{
		custodianunit.FeeCollectorName: nil,
		types.NotBondedPoolName:        {supply.Burner, supply.Staking},
		types.BondedPoolName:           {supply.Burner, supply.Staking},
	}
	supplyKeeper := supply.NewKeeper(mApp.Cdc, keySupply, mApp.CUKeeper, bankKeeper, maccPerms)
	keeper := NewKeeper(mApp.Cdc, keyStaking, tkeyStaking, supplyKeeper, mApp.ParamsKeeper.Subspace(DefaultParamspace), DefaultCodespace)

	mApp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mApp.SetEndBlocker(getEndBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper, mApp.CUKeeper, supplyKeeper,
		[]supplyexported.ModuleAccountI{feeCollector, notBondedPool, bondPool}))

	require.NoError(t, mApp.CompleteSetup(keyStaking, tkeyStaking, keySupply))

	return mApp, keeper
}

// getEndBlocker returns a staking endblocker.
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := EndBlocker(ctx, keeper)

		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
			Events:           ctx.EventManager().ABCIEvents(),
		}
	}
}

// getInitChainer initializes the chainer of the mock app and sets the genesis
// state. It returns an empty ResponseInitChain.
func getInitChainer(mapp *mock.App, keeper Keeper, cuKeeper types.CUKeeper, supplyKeeper types.SupplyKeeper,
	blacklistedAddrs []supplyexported.ModuleAccountI) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		// set module accounts
		for _, macc := range blacklistedAddrs {
			supplyKeeper.SetModuleAccount(ctx, macc)
		}

		stakingGenesis := DefaultGenesisState()
		stakingGenesis.Params.ElectionPeriod = 5

		validators := InitGenesis(ctx, keeper, cuKeeper, supplyKeeper, stakingGenesis)
		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

//__________________________________________________________________________________________

func checkValidator(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.ValAddress, expFound bool) Validator {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	validator, found := keeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkActiveValidator(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.ValAddress, isBonded bool) Validator {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	validator, found := keeper.GetValidator(ctxCheck, addr)

	require.True(t, found)
	require.Equal(t, isBonded, validator.IsBonded())
	return validator
}

func checkDelegation(
	t *testing.T, mapp *mock.App, keeper Keeper, delegatorAddr sdk.CUAddress,
	validatorAddr sdk.ValAddress, expFound bool, expShares sdk.Dec,
) {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	delegation, found := keeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		require.True(t, found)
		require.True(sdk.DecEq(t, expShares, delegation.Shares))

		return
	}

	require.False(t, found)
}

func TestStakingMsgs(t *testing.T) {
	mApp, keeper := getMockApp(t)

	genTokens := sdk.TokensFromConsensusPower(420000)
	bondTokens := sdk.TokensFromConsensusPower(100000)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc1 := &custodianunit.BaseCU{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	acc2 := &custodianunit.BaseCU{
		Address: addr2,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []custodianunit.CU{acc1, acc2}

	mock.SetGenesis(mApp, accs)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})

	// create validator
	description := NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := NewMsgCreateValidator(
		sdk.ValAddress(addr1), priv1.PubKey(), bondCoin, description, commissionRates, sdk.OneInt(), false,
	)

	mock.AdvanceBlock(mApp.BaseApp, mApp.LastBlockHeight(), 3)
	header := abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{createValidatorMsg}, []uint64{0}, true, true, priv1)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, mApp, keeper, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1), validator.OperatorAddress)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// edit the validator
	description = NewDescription("bar_moniker", "", "", "")
	editValidatorMsg := NewMsgEditValidator(sdk.ValAddress(addr1), description, nil, nil, nil)

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{editValidatorMsg}, []uint64{1}, true, true, priv1)

	validator = checkValidator(t, mApp, keeper, sdk.ValAddress(addr1), true)
	require.Equal(t, description, validator.Description)

	// delegate
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})
	delegateMsg := NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{delegateMsg}, []uint64{0}, true, true, priv2)
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
	checkDelegation(t, mApp, keeper, addr2, sdk.ValAddress(addr1), true, bondTokens.ToDec())

	// begin unbonding
	beginUnbondingMsg := NewMsgUndelegate(addr2, sdk.ValAddress(addr1), bondCoin)
	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, []uint64{1}, true, true, priv2)

	// delegation should exist anymore
	checkDelegation(t, mApp, keeper, addr2, sdk.ValAddress(addr1), false, sdk.Dec{})

	// balance should be the same because bonding not yet complete
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
}

type env struct {
	mApp      *mock.App
	basePower int64
	keeper    Keeper
	n         int
	privs     []crypto.PrivKey
	accs      []custodianunit.CU
	bondCoin  sdk.Coin
	genCoin   sdk.Coin
}

func TestStakingUpdate(t *testing.T) {
	env := setupValidators(t)
	// unbond
	ubdIndex := 9
	beginUnbondingMsg := NewMsgUndelegate(env.accs[ubdIndex].GetAddress(), sdk.ValAddress(env.accs[ubdIndex].GetAddress()), env.bondCoin)
	header := abci.Header{Height: env.mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, env.mApp.Cdc, env.mApp.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, []uint64{1}, true, true, env.privs[ubdIndex])

	mock.AdvanceBlock(env.mApp.BaseApp, env.mApp.LastBlockHeight(), 2)
	// still the same
	for i := 0; i < env.n; i++ {
		checkActiveValidator(t, env.mApp, env.keeper, sdk.ValAddress(env.accs[i].GetAddress()), true)
	}

	_, responseEndBlocks := mock.AdvanceBlock(env.mApp.BaseApp, env.mApp.LastBlockHeight(), 1)
	require.Len(t, responseEndBlocks[0].ValidatorUpdates, 1)
	require.EqualValues(t, tmtypes.TM2PB.PubKey(env.privs[ubdIndex].PubKey()), responseEndBlocks[0].ValidatorUpdates[0].PubKey)
	require.EqualValues(t, env.basePower*(int64(ubdIndex)), responseEndBlocks[0].ValidatorUpdates[0].Power)
	require.Empty(t, responseEndBlocks[0].Events)
	// still the same
	for i := 0; i < env.n; i++ {
		checkActiveValidator(t, env.mApp, env.keeper, sdk.ValAddress(env.accs[i].GetAddress()), true)
	}

	// unbond all
	beginUnbondingMsg = NewMsgUndelegate(env.accs[ubdIndex].GetAddress(), sdk.ValAddress(env.accs[ubdIndex].GetAddress()), env.bondCoin.MulRaw(int64(ubdIndex)))
	header = abci.Header{Height: env.mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, env.mApp.Cdc, env.mApp.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, []uint64{2}, true, true, env.privs[ubdIndex])
	fmt.Printf("current block height %d\n", env.mApp.LastBlockHeight())

	mock.AdvanceBlock(env.mApp.BaseApp, env.mApp.LastBlockHeight(), 3)
	// still the same
	for i := 0; i < env.n; i++ {
		validator := checkActiveValidator(t, env.mApp, env.keeper, sdk.ValAddress(env.accs[i].GetAddress()), true)
		require.Equal(t, i == ubdIndex, validator.Jailed)
	}

	_, responseEndBlocks = mock.AdvanceBlock(env.mApp.BaseApp, env.mApp.LastBlockHeight(), 1)
	require.Len(t, responseEndBlocks[0].ValidatorUpdates, 1)

	require.EqualValues(t, tmtypes.TM2PB.PubKey(env.privs[ubdIndex].PubKey()), responseEndBlocks[0].ValidatorUpdates[0].PubKey)
	require.EqualValues(t, 0, responseEndBlocks[0].ValidatorUpdates[0].Power)
	require.Len(t, responseEndBlocks[0].Events, 1)
	require.EqualValues(t, types.EventTypeMigrationBegin, responseEndBlocks[0].Events[0].Type)
	// n - 11 - 1 become bonded and ubdIndex become unbonding
	for i := 0; i < env.n; i++ {
		checkActiveValidator(t, env.mApp, env.keeper, sdk.ValAddress(env.accs[i].GetAddress()), i != ubdIndex)
	}

}

func TestStakingJail(t *testing.T) {
	env := setupValidators(t)

	// jail 1/6
	jailIndex := []int{7, 10}
	jailedCons := make([]sdk.ConsAddress, len(jailIndex))
	for i, index := range jailIndex {
		jailedCons[i] = sdk.ConsAddress(env.privs[index].PubKey().Address())
	}
	responseEndBlock := JailValidator(env.mApp, env.keeper, env.mApp.LastBlockHeight()+1, jailedCons)
	for _, index := range jailIndex {
		validator := checkActiveValidator(t, env.mApp, env.keeper, sdk.ValAddress(env.accs[index].GetAddress()), false)
		require.True(t, validator.Jailed)
	}

	require.Len(t, responseEndBlock.ValidatorUpdates, 2)

	require.Len(t, responseEndBlock.Events, 2)
	require.EqualValues(t, types.EventTypeMigrationBegin, responseEndBlock.Events[0].Type)

	// jail one more
	jailIndex = []int{8}
	jailedCons = make([]sdk.ConsAddress, len(jailIndex))
	for i, index := range jailIndex {
		jailedCons[i] = sdk.ConsAddress(env.privs[index].PubKey().Address())
	}
	responseEndBlock = JailValidator(env.mApp, env.keeper, env.mApp.LastBlockHeight()+1, jailedCons)
	// should not update because last migration is not finished
	for _, index := range jailIndex {
		validator := checkActiveValidator(t, env.mApp, env.keeper, sdk.ValAddress(env.accs[index].GetAddress()), true)
		require.True(t, validator.Jailed)
	}

	require.Len(t, responseEndBlock.ValidatorUpdates, 0)

	require.Len(t, responseEndBlock.Events, 0)
}

func setupValidators(t *testing.T) *env {
	n := 13

	// init n validators
	mApp, keeper := getMockApp(t)

	genTokens := sdk.TokensFromConsensusPower(21000000)
	basePower := int64(500000)
	bondTokens := sdk.TokensFromConsensusPower(basePower)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	msgs := make([]sdk.Msg, 0, 3*n)
	seqs := make([]uint64, n)
	privs := make([]crypto.PrivKey, n)
	accs := make([]custodianunit.CU, n)
	for i := 0; i < n; i++ {
		priv := secp256k1.GenPrivKey()
		acc := &custodianunit.BaseCU{
			Address: sdk.CUAddress(priv.PubKey().Address()),
			Coins:   sdk.Coins{genCoin.MulRaw(int64(i + 1))},
		}
		// create validator
		description := NewDescription(fmt.Sprintf("foo_moniker-%d", i), "", "", "")
		createValidatorMsg := NewMsgCreateValidator(
			sdk.ValAddress(acc.Address), priv.PubKey(), bondCoin.MulRaw(int64(i+1)), description, commissionRates, sdk.OneInt(), false,
		)
		valAddr := sdk.ValAddress(acc.Address)
		heartbeatMsg := NewMsgKeyNodeHeartbeat(1, valAddr)
		isKeyNode := true
		editValidatorMsg := NewMsgEditValidator(valAddr, description, nil, nil, &isKeyNode)
		privs[i] = priv
		accs[i] = acc
		msgs = append(msgs, createValidatorMsg, heartbeatMsg, editValidatorMsg)
	}
	mock.SetGenesis(mApp, accs)
	for i := 0; i < n; i++ {
		mock.CheckBalance(t, mApp, accs[i].GetAddress(), sdk.Coins{genCoin.MulRaw(int64(i + 1))})
	}

	// check current validators, should be zero
	for i := 0; i < n; i++ {
		checkValidator(t, mApp, keeper, sdk.ValAddress(accs[i].GetAddress()), false)
	}

	// create validators for all
	mock.AdvanceBlock(mApp.BaseApp, mApp.LastBlockHeight(), 1)
	header := abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, msgs, seqs, true, true, privs...)
	for i := 0; i < n; i++ {
		mock.CheckBalance(t, mApp, accs[i].GetAddress(), sdk.Coins{genCoin.Sub(bondCoin).MulRaw(int64(i + 1))})
	}

	// check current validators, should be found but not bonded
	for i := 0; i < n; i++ {
		checkActiveValidator(t, mApp, keeper, sdk.ValAddress(accs[i].GetAddress()), false)
	}

	mock.AdvanceBlock(mApp.BaseApp, mApp.LastBlockHeight(), 1)
	// still not bonded
	for i := 0; i < n; i++ {
		checkActiveValidator(t, mApp, keeper, sdk.ValAddress(accs[i].GetAddress()), false)
	}
	mock.AdvanceBlock(mApp.BaseApp, mApp.LastBlockHeight(), 1)
	// check current validators, last 11 become bonded
	for i := 0; i < n; i++ {
		fmt.Printf("check validator %d\n", i)
		checkActiveValidator(t, mApp, keeper, sdk.ValAddress(accs[i].GetAddress()), true)
	}

	SetMigrationFinished(mApp, keeper, mApp.LastBlockHeight()+1)

	return &env{
		mApp:      mApp,
		keeper:    keeper,
		n:         n,
		privs:     privs,
		accs:      accs,
		basePower: basePower,
		bondCoin:  bondCoin,
		genCoin:   genCoin,
	}
}

func SetMigrationFinished(mApp *mock.App, k Keeper, height int64) abci.ResponseEndBlock {
	header := abci.Header{Height: height}
	mApp.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mApp.NewContext(false, header)
	k.SetMigrationFinished(ctx)
	responseEndBlock := mApp.BaseApp.EndBlock(abci.RequestEndBlock{})
	mApp.BaseApp.Commit()
	return responseEndBlock
}

func JailValidator(mApp *mock.App, k Keeper, height int64, consAddrs []sdk.ConsAddress) abci.ResponseEndBlock {
	header := abci.Header{Height: height}
	mApp.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mApp.NewContext(false, header)
	for _, consAddr := range consAddrs {
		k.Jail(ctx, consAddr)
	}
	responseEndBlock := mApp.BaseApp.EndBlock(abci.RequestEndBlock{})
	mApp.BaseApp.Commit()
	return responseEndBlock
}
