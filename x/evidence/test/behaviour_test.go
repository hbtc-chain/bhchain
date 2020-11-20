package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence"
	"github.com/hbtc-chain/bhchain/x/evidence/keeper"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
	"github.com/hbtc-chain/bhchain/x/evidence/types/mocks"
	"github.com/hbtc-chain/bhchain/x/params"
	stakingtypes "github.com/hbtc-chain/bhchain/x/staking/types"
)

type testEnv struct {
	ctx           sdk.Context
	keeper        keeper.Keeper
	validators    []stakingtypes.Validator
	stakingKeeper *mocks.StakingKeeper
}

func TestKeeperHandleBehaviourAllMis(t *testing.T) {
	startBlock := 5
	env := setupUnitTestEnv()
	max := int(env.keeper.MaxMisbehaviourCount(env.ctx, types.VoteBehaviourKey))

	env.stakingKeeper.On("JailByOperator", mock.Anything, env.validators[0].OperatorAddress)
	env.stakingKeeper.On("SlashByOperator", mock.Anything, env.validators[0].OperatorAddress, int64(startBlock+max), mock.Anything)
	env.stakingKeeper.On("SlashByOperator", mock.Anything, env.validators[0].OperatorAddress, int64(startBlock+(max+1)*2-1), mock.Anything)

	for i := 0; i < (max+1)*2; i++ {
		env.keeper.HandleBehaviour(env.ctx, types.VoteBehaviourKey, env.validators[0].OperatorAddress, uint64(startBlock+i), false)
	}
	env.stakingKeeper.AssertExpectations(t)
	env.stakingKeeper.AssertNumberOfCalls(t, "JailByOperator", 2)
}

func TestKeeperHandleBehaviourRevert(t *testing.T) {
	startBlock := 5

	env := setupUnitTestEnv()

	window := int(env.keeper.BehaviourWindow(env.ctx, types.VoteBehaviourKey))

	for i := 0; i < window*2; i++ {
		env.keeper.HandleBehaviour(env.ctx, types.VoteBehaviourKey, env.validators[0].OperatorAddress, uint64(startBlock+i), i >= window/10)
		validatorBehavior := env.keeper.GetValidatorBehaviour(env.ctx, types.VoteBehaviourKey, env.validators[0].OperatorAddress)
		if i < window/10 {
			require.EqualValues(t, i+1, validatorBehavior.MisbehaviourCounter)
		} else if i < window {
			require.EqualValues(t, window/10, validatorBehavior.MisbehaviourCounter)
		} else if i < window+window/10 {
			require.EqualValues(t, window/10-(i-window+1), validatorBehavior.MisbehaviourCounter)
		} else {
			require.EqualValues(t, 0, validatorBehavior.MisbehaviourCounter)
		}
	}

	env.stakingKeeper.AssertNumberOfCalls(t, "JailByOperator", 0)
}

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.CUAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.CUAddress(priv2.PubKey().Address())
	priv3 = secp256k1.GenPrivKey()
	addr3 = sdk.CUAddress(priv3.PubKey().Address())
	priv4 = secp256k1.GenPrivKey()
	addr4 = sdk.CUAddress(priv4.PubKey().Address())
)

func setupUnitTestEnv() *testEnv {
	validators := []stakingtypes.Validator{
		{
			OperatorAddress: sdk.ValAddress(addr1),
			ConsPubKey:      priv1.PubKey(),
		},
		{
			OperatorAddress: sdk.ValAddress(addr2),
			ConsPubKey:      priv2.PubKey(),
		},
		{
			OperatorAddress: sdk.ValAddress(addr3),
			ConsPubKey:      priv3.PubKey(),
		},
		{
			OperatorAddress: sdk.ValAddress(addr4),
			ConsPubKey:      priv4.PubKey(),
		},
	}

	cdc := codec.New()
	types.RegisterCodec(cdc)
	// declare keys
	keeperKey := sdk.NewKVStoreKey(types.ModuleName)
	paramsKey := sdk.NewKVStoreKey("params")
	paramsTransientKey := sdk.NewTransientStoreKey("transient_params")

	// load store
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(paramsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keeperKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(paramsTransientKey, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// ctx
	ctx := sdk.NewContext(ms, abci.Header{
		Height: 100,
	}, false, log.NewNopLogger())

	// instantiate keeper
	pk := params.NewKeeper(cdc, paramsKey, paramsTransientKey, "bhchain")
	paramSpace := pk.Subspace(types.DefaultParamspace)

	stakingKeeper := &mocks.StakingKeeper{}

	keeper := keeper.NewKeeper(cdc, keeperKey, paramSpace, stakingKeeper)

	// init genesis state
	evidence.InitGenesis(ctx, keeper, types.DefaultGenesisState())

	return &testEnv{
		ctx:           ctx,
		keeper:        keeper,
		validators:    validators,
		stakingKeeper: stakingKeeper,
	}
}
