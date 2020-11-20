package token

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/hbtc-chain/bhchain/store"
	"github.com/hbtc-chain/bhchain/x/evidence"
	"github.com/hbtc-chain/bhchain/x/params"
	stakingtypes "github.com/hbtc-chain/bhchain/x/staking/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/hbtc-chain/bhchain/x/token/types"

	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/tendermint/tm-db"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
)

type testEnv struct {
	tk                Keeper
	storeKey          sdk.StoreKey
	cdc               *codec.Codec
	ctx               sdk.Context
	validators        []stakingtypes.Validator
	evidenceKeeper    evidence.Keeper
	mockStakingKeeper *mockStakingKeeper
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

type mockStakingKeeper struct {
	mock.Mock
	validators []stakingtypes.Validator
}

func newMockStakingKeeper(validators []stakingtypes.Validator) *mockStakingKeeper {
	return &mockStakingKeeper{validators: validators}
}

func (m *mockStakingKeeper) IsActiveKeyNode(_ sdk.Context, addr sdk.CUAddress) (bool, int) {
	for _, validator := range m.validators {
		if sdk.CUAddress(validator.OperatorAddress).Equals(addr) {
			return true, len(m.validators)
		}
	}
	return false, len(m.validators)
}

func (m *mockStakingKeeper) GetEpochByHeight(ctx sdk.Context, height uint64) sdk.Epoch {
	valSet := make([]sdk.CUAddress, 0)
	for _, validator := range m.validators {
		valSet = append(valSet, sdk.CUAddress(validator.OperatorAddress))
	}

	return sdk.Epoch{
		Index:             1,
		StartBlockNum:     1,
		EndBlockNum:       0,
		KeyNodeSet:        valSet,
		MigrationFinished: true,
	}
}

func (m *mockStakingKeeper) SlashByOperator(_a0 sdk.Context, _a1 sdk.ValAddress, _a2 int64, _a3 sdk.Dec) {
	m.Called(_a0, _a1, _a2, _a3)
}

func (m *mockStakingKeeper) JailByOperator(ctx sdk.Context, operator sdk.ValAddress) {
	m.Called(ctx, operator)
}

func (m *mockStakingKeeper) GetCurrentEpoch(ctx sdk.Context) sdk.Epoch {
	valSet := make([]sdk.CUAddress, 0)
	for _, validator := range m.validators {
		valSet = append(valSet, sdk.CUAddress(validator.OperatorAddress))
	}
	return sdk.Epoch{
		Index:             1,
		StartBlockNum:     1,
		EndBlockNum:       0,
		KeyNodeSet:        valSet,
		MigrationFinished: true,
	}
}

func setupUnitTestEnv(initDefaultTokens ...bool) testEnv {
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

	db := dbm.NewMemDB()
	tokenKey := sdk.NewKVStoreKey(ModuleName)
	addrKey := sdk.NewKVStoreKey("address") //address.ModuleName, use literal name to avoid import cycle
	cuKey := sdk.NewKVStoreKey("cu")        //cu.ModuleName, use literal name to avoid import cycle
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")
	keyEvid := sdk.NewKVStoreKey(evidence.StoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tokenKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(addrKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(cuKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyEvid, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	cdc := codec.New()
	types.RegisterCodec(cdc)
	evidence.RegisterCodec(cdc)
	ctx := sdk.NewContext(ms, abci.Header{
		Height: 100,
	}, false, log.NewNopLogger())

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, "bhchain")

	tk := NewKeeper(tokenKey, cdc)

	stakingKeeper := newMockStakingKeeper(validators)
	tk.SetStakingKeeper(stakingKeeper)
	evidenceKeeper := evidence.NewKeeper(cdc, keyEvid, pk.Subspace(evidence.DefaultParamspace), stakingKeeper)
	evidence.InitGenesis(ctx, evidenceKeeper, evidence.DefaultGenesisState())
	tk.SetEvidenceKeeper(evidenceKeeper)
	if len(initDefaultTokens) > 0 && initDefaultTokens[0] {
		InitGenesis(ctx, tk, DefaultGenesisState())
	} else {
		for _, t := range TestTokenData {
			tk.SetToken(ctx, t)
		}
	}

	return testEnv{
		tk:                tk,
		storeKey:          tokenKey,
		cdc:               cdc,
		ctx:               ctx,
		validators:        validators,
		evidenceKeeper:    evidenceKeeper,
		mockStakingKeeper: stakingKeeper,
	}
}

func TestSetTokenInfo(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	for symbol := range TestTokenData {
		keeper.SetToken(ctx, newTokenInfo(symbol))
	}

	for symbol, testToken := range TestIBCTokens {
		tk := keeper.GetIBCToken(ctx, symbol)
		assert.Equal(t, testToken.Symbol, tk.Symbol)
		assert.Equal(t, testToken.Issuer, tk.Issuer)
		assert.Equal(t, testToken.Chain, tk.Chain)
		assert.Equal(t, testToken.TokenType, tk.TokenType)
		assert.Equal(t, testToken.SendEnabled, tk.SendEnabled)
		assert.Equal(t, testToken.DepositEnabled, tk.DepositEnabled)
		assert.Equal(t, testToken.WithdrawalEnabled, tk.WithdrawalEnabled)
		assert.Equal(t, testToken.Decimals, tk.Decimals)
		assert.Equal(t, testToken.TotalSupply, tk.TotalSupply)
		assert.Equal(t, testToken.CollectThreshold, tk.CollectThreshold)
		assert.Equal(t, testToken.OpenFee, tk.OpenFee)
		assert.Equal(t, testToken.SysOpenFee, tk.SysOpenFee)
		assert.Equal(t, testToken.WithdrawalFeeRate, tk.WithdrawalFeeRate)
		assert.Equal(t, testToken.DepositThreshold, tk.DepositThreshold)
		assert.Equal(t, testToken.MaxOpCUNumber, tk.MaxOpCUNumber)
		assert.Equal(t, testToken.GasLimit, tk.GasLimit)
		assert.Equal(t, testToken.OpCUSysTransferAmount(), tk.OpCUSysTransferAmount())
		assert.Equal(t, testToken.SysTransferAmount(), tk.SysTransferAmount())
	}

	//set token info does not update ibc symbol list
	//symbols := keeper.GetIBCTokenSymbols(ctx)
	//
	//assert.Contains(t, symbols, sdk.Symbol("btc"))
	//assert.Contains(t, symbols, sdk.Symbol("eth"))
	//assert.Contains(t, symbols, sdk.Symbol("usdt"))
}

func newTokenInfo(symbol sdk.Symbol) sdk.Token {
	t, ok := TestTokenData[symbol]
	if !ok {
		return nil
	}
	return t
	// nt := *t
	// return &nt
}

func TestTokenInfoEncoding(t *testing.T) {
	input := setupUnitTestEnv()
	keeper := input.tk

	for _, info := range TestTokenData {
		bz := keeper.cdc.MustMarshalBinaryBare(info)
		var tokenInfo sdk.Token
		keeper.cdc.MustUnmarshalBinaryBare(bz, &tokenInfo)
		assert.Equal(t, info, tokenInfo)
	}
}

func TestKeeper_IsErc20Utxo(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	for symbol := range TestTokenData {
		keeper.SetToken(ctx, newTokenInfo(symbol))
	}

	assert.True(t, keeper.IsSubToken(ctx, testUsdtSymbol))
	assert.False(t, keeper.IsSubToken(ctx, "aa"))
	assert.False(t, keeper.IsSubToken(ctx, "bb"))

	assert.True(t, keeper.GetIBCToken(ctx, testBtcSymbol).TokenType == sdk.UtxoBased)
	assert.False(t, keeper.GetIBCToken(ctx, testUsdtSymbol).TokenType == sdk.UtxoBased)
	assert.False(t, keeper.GetIBCToken(ctx, testEthSymbol).TokenType == sdk.UtxoBased)
}
