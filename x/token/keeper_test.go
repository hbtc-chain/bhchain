package token

import (
	"testing"

	"github.com/hbtc-chain/bhchain/store"
	"github.com/hbtc-chain/bhchain/x/evidence"
	"github.com/hbtc-chain/bhchain/x/params"
	stakingtypes "github.com/hbtc-chain/bhchain/x/staking/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/hbtc-chain/bhchain/x/token/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

func setupUnitTestEnv() testEnv {
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

	tk := NewKeeper(tokenKey, cdc, pk.Subspace(DefaultParamspace))

	stakingKeeper := newMockStakingKeeper(validators)
	tk.SetStakingKeeper(stakingKeeper)
	evidenceKeeper := evidence.NewKeeper(cdc, keyEvid, pk.Subspace(evidence.DefaultParamspace), stakingKeeper)
	evidence.InitGenesis(ctx, evidenceKeeper, evidence.DefaultGenesisState())
	tk.SetEvidenceKeeper(evidenceKeeper)

	InitGenesis(ctx, tk, DefaultGenesisState())
	//for s := range TestTokenData {
	//	symbol := s
	//	tk.SetTokenInfo(ctx, newTokenInfo(symbol))
	//}

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
		keeper.SetTokenInfo(ctx, newTokenInfo(symbol))
	}

	for symbol, testToken := range TestTokenData {
		assert.Equal(t, testToken.Symbol, keeper.GetSymbol(ctx, symbol))
		assert.Equal(t, testToken.Issuer, keeper.GetIssuer(ctx, symbol))
		assert.Equal(t, testToken.Chain, keeper.GetChain(ctx, symbol))
		assert.Equal(t, testToken.TokenType, keeper.GetTokenType(ctx, symbol))
		assert.Equal(t, testToken.IsSendEnabled, keeper.IsSendEnabled(ctx, symbol))
		assert.Equal(t, testToken.IsDepositEnabled, keeper.IsDepositEnabled(ctx, symbol))
		assert.Equal(t, testToken.IsWithdrawalEnabled, keeper.IsWithdrawalEnabled(ctx, symbol))
		assert.Equal(t, testToken.Decimals, keeper.GetDecimals(ctx, symbol))
		assert.Equal(t, testToken.TotalSupply, keeper.GetTotalSupply(ctx, symbol))
		assert.Equal(t, testToken.CollectThreshold, keeper.GetCollectThreshold(ctx, symbol))
		assert.Equal(t, testToken.OpenFee, keeper.GetOpenFee(ctx, symbol))
		assert.Equal(t, testToken.SysOpenFee, keeper.GetSysOpenFee(ctx, symbol))
		assert.Equal(t, testToken.WithdrawalFeeRate, keeper.GetWithdrawalFeeRate(ctx, symbol))
		assert.Equal(t, testToken.DepositThreshold, keeper.GetDepositThreshold(ctx, symbol))
		assert.Equal(t, testToken.MaxOpCUNumber, keeper.GetMaxOpCUNumber(ctx, symbol))
		assert.Equal(t, testToken.GasLimit, keeper.GetGasLimit(ctx, symbol))
		assert.Equal(t, testToken.OpCUSysTransferAmount(), keeper.GetOpCUSystransferAmount(ctx, symbol))
		assert.Equal(t, testToken.SysTransferAmount(), keeper.GetSystransferAmount(ctx, symbol))
	}

	symbols := keeper.GetSymbols(ctx)
	assert.Contains(t, symbols, BtcToken)
	assert.Contains(t, symbols, EthToken)
	assert.Contains(t, symbols, UsdtToken)
}

func TestModifySend(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	keeper.SetTokenInfo(ctx, newTokenInfo(sdk.Symbol(BtcToken)))
	assert.True(t, keeper.IsSendEnabled(ctx, sdk.Symbol(BtcToken)))

	keeper.EnableSend(ctx, sdk.Symbol(BtcToken))
	assert.True(t, keeper.IsSendEnabled(ctx, sdk.Symbol(BtcToken)))

	keeper.DisableSend(ctx, sdk.Symbol(BtcToken))
	assert.False(t, keeper.IsSendEnabled(ctx, sdk.Symbol(BtcToken)))
}

func newTokenInfo(symbol sdk.Symbol) *sdk.TokenInfo {
	t, ok := TestTokenData[symbol]
	if !ok {
		return nil
	}
	return NewTokenInfo(t.Symbol, t.Chain, t.Issuer, t.TokenType,
		t.IsSendEnabled, t.IsDepositEnabled, t.IsWithdrawalEnabled, t.Decimals,
		t.TotalSupply, t.CollectThreshold, t.DepositThreshold, t.OpenFee,
		t.SysOpenFee, t.WithdrawalFeeRate, t.SysTransferNum, t.OpCUSysTransferNum,
		t.GasLimit, t.GasPrice, t.MaxOpCUNumber, t.Confirmations, t.IsNonceBased)

}

func TestTokenInfoEncoding(t *testing.T) {
	input := setupUnitTestEnv()
	keeper := input.tk

	for _, info := range TestTokenData {
		bz := keeper.cdc.MustMarshalBinaryBare(info)
		var tokenInfo sdk.TokenInfo
		keeper.cdc.MustUnmarshalBinaryBare(bz, &tokenInfo)
		assert.Equal(t, info, tokenInfo)
	}
}

func TestKeeper_IsErc20Utxo(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	for symbol := range TestTokenData {
		keeper.SetTokenInfo(ctx, newTokenInfo(symbol))
	}

	assert.True(t, keeper.IsSubToken(ctx, sdk.Symbol(UsdtToken)))
	assert.False(t, keeper.IsSubToken(ctx, sdk.Symbol(BtcToken)))
	assert.False(t, keeper.IsSubToken(ctx, sdk.Symbol(EthToken)))

	assert.True(t, keeper.IsUtxoBased(ctx, sdk.Symbol(BtcToken)))
	assert.False(t, keeper.IsUtxoBased(ctx, sdk.Symbol(UsdtToken)))
	assert.False(t, keeper.IsUtxoBased(ctx, sdk.Symbol(EthToken)))
}

func TestKeeper_EnableSendDepositWithdrawal(t *testing.T) {
	input := setupUnitTestEnv()
	keeper := input.tk
	ctx := input.ctx
	for symbol := range TestTokenData {
		keeper.SetTokenInfo(ctx, newTokenInfo(symbol))
	}

	keeper.DisableSend(ctx, sdk.Symbol(EthToken))
	keeper.DisableWithdrawal(ctx, sdk.Symbol(EthToken))
	keeper.DisableDeposit(ctx, sdk.Symbol(EthToken))
	onSend := keeper.IsSendEnabled(ctx, sdk.Symbol(EthToken))
	onWithdrawal := keeper.IsWithdrawalEnabled(ctx, sdk.Symbol(EthToken))
	onDeposit := keeper.IsDepositEnabled(ctx, sdk.Symbol(EthToken))
	assert.EqualValues(t, false, onSend, onWithdrawal, onDeposit)

	keeper.EnableSend(ctx, sdk.Symbol(EthToken))
	keeper.EnableWithdrawal(ctx, sdk.Symbol(EthToken))
	keeper.EnableDeposit(ctx, sdk.Symbol(EthToken))
	onSend = keeper.IsSendEnabled(ctx, sdk.Symbol(EthToken))
	onWithdrawal = keeper.IsWithdrawalEnabled(ctx, sdk.Symbol(EthToken))
	onDeposit = keeper.IsDepositEnabled(ctx, sdk.Symbol(EthToken))
	assert.EqualValues(t, true, onSend, onWithdrawal, onDeposit)
}
