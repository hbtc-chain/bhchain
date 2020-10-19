package mapping

import (
	"github.com/hbtc-chain/bhchain/x/mapping/types"
	"reflect"
	"testing"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/params/subspace"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testEnv struct {
	mk       Keeper
	ck       custodianunit.CUKeeperI
	tk       *token.Keeper
	rk       *receipt.Keeper
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc      *codec.Codec // The wire codec for binary encoding/decoding
	ctx      sdk.Context
}

var testTokenInfo = []sdk.TokenInfo{
	{
		Symbol:              "btc",
		Chain:               "btc",
		Issuer:              "",
		TokenType:           sdk.UtxoBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		TotalSupply:         sdk.NewInt(2100),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},
	{
		Symbol:              "eth",
		Chain:               "eth",
		Issuer:              "",
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            18,
		TotalSupply:         sdk.NewInt(10000),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},

	//ERC20
	{
		Symbol:              "tbtc",
		Chain:               "eth",
		Issuer:              "0x123456",
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		TotalSupply:         sdk.NewInt(2100),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},
	{
		Symbol:              "tbtc2",
		Chain:               "eth",
		Issuer:              "0x12345678",
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		TotalSupply:         sdk.NewInt(1100), // Does not match total supply for BTC
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},
	{
		Symbol:              "tbtc3",
		Chain:               "eth",
		Issuer:              "0x1234567890",
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            18, // Does not match decimals for BTC
		TotalSupply:         sdk.NewInt(2100),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},
	{ // Identical to btc
		Symbol:              "tbtc4",
		Chain:               "eth",
		Issuer:              "0x123456789012",
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		TotalSupply:         sdk.NewInt(2100),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},
	{ // Identical to btc
		Symbol:              "tbtc5",
		Chain:               "eth",
		Issuer:              "0x12345678901234",
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		TotalSupply:         sdk.NewInt(2100),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1200),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	},
}

func setupUnitTestEnv() testEnv {
	db := dbm.NewMemDB()

	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	custodianunit.RegisterCodec(cdc)
	token.RegisterCodec(cdc)
	receipt.RegisterCodec(cdc)
	RegisterCodec(cdc)

	mappingKey := sdk.NewKVStoreKey(StoreKey)
	tokenKey := sdk.NewKVStoreKey(token.StoreKey)
	cuKey := sdk.NewKVStoreKey(custodianunit.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(mappingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tokenKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(cuKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, custodianunit.DefaultParamspace)
	rk := receipt.NewKeeper(cdc)
	tk := token.NewKeeper(tokenKey, cdc, subspace.NewSubspace(cdc, keyParams, tkeyParams, token.DefaultParamspace))
	ck := custodianunit.NewCUKeeper(cdc, cuKey, &tk, ps, custodianunit.ProtoBaseCU)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, sdk.CodespaceType("mapping"))
	mk := NewKeeper(mappingKey, cdc, &tk, ck, rk, pk.Subspace(types.DefaultParamspace))

	ck.SetParams(ctx, custodianunit.DefaultParams())

	for _, tokenInfo := range testTokenInfo {
		tk.SetTokenInfo(ctx, &tokenInfo)
	}

	return testEnv{
		mk:       mk,
		ck:       ck,
		tk:       &tk,
		rk:       rk,
		storeKey: tokenKey,
		cdc:      cdc,
		ctx:      ctx,
	}
}

func TestMappingInfo(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.mk

	assert.False(t, keeper.HasTargetSymbol(ctx, sdk.Symbol("btc")))

	mi := &MappingInfo{
		IssueSymbol:  sdk.Symbol("tbtc"),
		TargetSymbol: sdk.Symbol("btc"),
		TotalSupply:  sdk.NewInt(2100),
		IssuePool:    sdk.NewInt(2000),
		Enabled:      true,
	}
	keeper.SetMappingInfo(ctx, mi)

	assert.True(t, reflect.DeepEqual(mi, keeper.GetMappingInfo(ctx, mi.IssueSymbol)))
	assert.True(t, reflect.DeepEqual([]sdk.Symbol{sdk.Symbol("tbtc")}, keeper.GetIssueSymbols(ctx)))
	assert.True(t, keeper.HasTargetSymbol(ctx, sdk.Symbol("btc")))
}
