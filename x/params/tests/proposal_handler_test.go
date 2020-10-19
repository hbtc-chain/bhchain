package tests

import (
	"testing"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	distr "github.com/hbtc-chain/bhchain/x/distribution"
	"github.com/hbtc-chain/bhchain/x/gov"
	"github.com/hbtc-chain/bhchain/x/hrc20"
	"github.com/hbtc-chain/bhchain/x/mint"
	"github.com/hbtc-chain/bhchain/x/order"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/slashing"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/hbtc-chain/bhchain/x/transfer"
	"github.com/hbtc-chain/bhchain/x/transfer/keeper"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testEnv struct {
	cdc           *codec.Codec
	ctx           sdk.Context
	k             keeper.BaseKeeper
	ck            custodianunit.CUKeeper
	tk            token.Keeper
	ok            order.Keeper
	rk            receipt.Keeper
	stakingkeeper staking.Keeper
	supplyKeeper  supply.Keeper
	pk            params.Keeper
	mk            mint.Keeper
	hk            hrc20.Keeper
}

func setupTestEnv(t *testing.T) testEnv {
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(custodianunit.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyToken := sdk.NewKVStoreKey(token.StoreKey)
	keyReceipt := sdk.NewKVStoreKey(receipt.StoreKey)
	keyOrder := sdk.NewKVStoreKey(order.StoreKey)
	keyGov := sdk.NewKVStoreKey(gov.StoreKey)
	keySlash := sdk.NewKVStoreKey(slashing.StoreKey)
	keyTransfer := sdk.NewKVStoreKey(transfer.StoreKey)
	keyDistr := sdk.NewKVStoreKey(distr.StoreKey)
	keyMint := sdk.NewKVStoreKey(mint.StoreKey)
	keyHrc20 := sdk.NewKVStoreKey(hrc20.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyToken, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyReceipt, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyOrder, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySlash, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyTransfer, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMint, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyHrc20, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	//register cdc
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	custodianunit.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	params.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	token.RegisterCodec(cdc)
	receipt.RegisterCodec(cdc)
	order.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	slashing.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	transfer.RegisterCodec(cdc)
	mint.RegisterCodec(cdc)
	hrc20.RegisterCodec(cdc)

	feeCollectorAcc := supply.NewEmptyModuleAccount(custodianunit.FeeCollectorName)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[sdk.CUAddress([]byte("moduleAcc")).String()] = true
	blacklistedAddrs[feeCollectorAcc.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	tk := token.NewKeeper(keyToken, cdc, pk.Subspace(token.DefaultParamspace))
	rk := receipt.NewKeeper(cdc)
	ok := order.NewKeeper(cdc, keyOrder, pk.Subspace(order.DefaultParamspace))
	ck := custodianunit.NewCUKeeper(
		cdc, keyAcc, &tk, pk.Subspace(custodianunit.DefaultParamspace), custodianunit.ProtoBaseCU,
	)
	ck.SetParams(ctx, custodianunit.DefaultParams())

	bankKeeper := keeper.NewBaseKeeper(cdc, keyTransfer, ck, &tk, &ok, rk, nil, nil, pk.Subspace(transfer.DefaultParamspace), transfer.DefaultCodespace, blacklistedAddrs)
	bankKeeper.SetSendEnabled(ctx, true)

	maccPerms := map[string][]string{
		custodianunit.FeeCollectorName: nil,
		staking.NotBondedPoolName:      []string{supply.Burner, supply.Staking},
		staking.BondedPoolName:         []string{supply.Burner, supply.Staking},
		mint.ModuleName:                []string{supply.Minter},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, ck, bankKeeper, maccPerms)
	supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	supplyKeeper.SetModuleAccount(ctx, bondPool)
	supplyKeeper.SetModuleAccount(ctx, notBondedPool)

	stakingKeeper := staking.NewKeeper(cdc, keyStaking, tkeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	params := staking.DefaultParams()
	stakingKeeper.SetParams(ctx, params)
	mk := mint.NewKeeper(cdc, keyMint, pk.Subspace(mint.DefaultParamspace), stakingKeeper, supplyKeeper, custodianunit.FeeCollectorName)
	bankKeeper.SetStakingKeeper(ctx, stakingKeeper)
	tk.SetStakingKeeper(stakingKeeper)

	hk := hrc20.NewKeeper(cdc, keyHrc20, pk.Subspace(hrc20.DefaultParamspace), &tk, ck, nil, supplyKeeper, rk)

	return testEnv{cdc: cdc, ctx: ctx, k: *bankKeeper, ck: ck, tk: tk, ok: ok, rk: *rk, stakingkeeper: stakingKeeper, supplyKeeper: *supplyKeeper, pk: pk, mk: mk, hk: hk}
}

func testProposal(changes ...params.ParamChange) params.ParameterChangeProposal {
	return params.NewParameterChangeProposal(
		"Test",
		"description",
		changes,
	)
}

func TestModifyMintParamByProposal(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	mk := input.mk
	pk := input.pk

	mk.SetParams(ctx, mint.DefaultParams())
	p := mk.GetParams(ctx)
	require.Equal(t, mint.DefaultParams(), p)
	require.Equal(t, sdk.NativeToken, p.MintDenom)
	tp := testProposal(params.NewParamChange(mint.DefaultParamspace, string(mint.KeyMintDenom), `"btc"`))

	hdlr := params.NewParamChangeProposalHandler(input.pk)
	res := hdlr(input.ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)

	var param string
	ss, ok := pk.GetSubspace(mint.DefaultParamspace)
	require.True(t, ok)

	ss.Get(ctx, mint.KeyMintDenom, &param)
	require.Equal(t, param, "btc")
	require.Equal(t, param, mk.GetParams(ctx).MintDenom)
}

func TestModifyStakingParamByProposal(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	stakingKeeper := input.stakingkeeper
	pk := input.pk

	stakingKeeper.SetParams(ctx, staking.DefaultParams())
	p := stakingKeeper.GetParams(ctx)
	require.Equal(t, staking.DefaultParams(), p)
	require.Equal(t, staking.DefaultMaxValidators, p.MaxValidators)
	tp := testProposal(params.NewParamChange(staking.DefaultParamspace, string(staking.KeyMaxValidators), "10"))

	hdlr := params.NewParamChangeProposalHandler(input.pk)
	res := hdlr(input.ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)

	var param uint16
	ss, ok := pk.GetSubspace(staking.DefaultParamspace)
	require.True(t, ok)

	ss.Get(ctx, staking.KeyMaxValidators, &param)
	require.Equal(t, uint16(10), param)
	require.Equal(t, param, stakingKeeper.GetParams(ctx).MaxValidators)
}

func TestModifyTokenParamByProposal(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	tk := input.tk
	pk := input.pk

	tk.SetParams(ctx, token.DefaultParams())
	p := tk.GetParams(ctx)
	require.Equal(t, token.DefaultParams(), p)
	require.Equal(t, token.DefaultTokenCacheSize, p.TokenCacheSize)
	tp := testProposal(params.NewParamChange(token.DefaultParamspace, string(token.KeyTokenCacheSize), `"10"`))

	hdlr := params.NewParamChangeProposalHandler(pk)
	res := hdlr(input.ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)

	var param uint64
	ss, ok := pk.GetSubspace(token.DefaultParamspace)
	require.True(t, ok)

	ss.Get(ctx, token.KeyTokenCacheSize, &param)
	require.Equal(t, uint64(10), param)
	require.Equal(t, param, tk.GetParams(ctx).TokenCacheSize)
}

func TestModifyHr20ParamByProposal(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	pk := input.pk
	hk := input.hk

	hk.SetParams(ctx, hrc20.DefaultParams())
	p := hk.GetParams(ctx)
	require.Equal(t, hrc20.DefaultParams(), p)
	require.Equal(t, hrc20.DefaultIssueTokenFee, p.IssueTokenFee)
	tp := testProposal(params.NewParamChange(hrc20.DefaultParamspace, string(hrc20.KeyIssueTokenFee), `"21000000000000000"`))

	hdlr := params.NewParamChangeProposalHandler(pk)
	res := hdlr(input.ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)

	var param sdk.Int
	ss, ok := pk.GetSubspace(hrc20.DefaultParamspace)
	require.True(t, ok)
	ss.Get(ctx, hrc20.KeyIssueTokenFee, &param)
	require.Equal(t, sdk.NewInt(21000000000000000), param)
	require.Equal(t, param, hk.GetParams(ctx).IssueTokenFee)

}
