package hrc10

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	distr "github.com/hbtc-chain/bhchain/x/distribution"
	"github.com/hbtc-chain/bhchain/x/gov"
	"github.com/hbtc-chain/bhchain/x/hrc10/types"
	"github.com/hbtc-chain/bhchain/x/mint"
	"github.com/hbtc-chain/bhchain/x/order"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/slashing"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

type testEnv struct {
	cdc            *codec.Codec
	ctx            sdk.Context
	transferKeeper transfer.Keeper
	ck             custodianunit.CUKeeper
	tk             token.Keeper
	ok             order.Keeper
	rk             receipt.Keeper
	stakingkeeper  staking.Keeper
	supplyKeeper   supply.Keeper
	pk             params.Keeper
	mk             mint.Keeper
	dk             distr.Keeper
	hrc10k         Keeper
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
	keyHrc10 := sdk.NewKVStoreKey(StoreKey)

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
	ms.MountStoreWithDB(keyHrc10, sdk.StoreTypeIAVL, db)
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
	RegisterCodec(cdc)

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
	tk := token.NewKeeper(keyToken, cdc)
	rk := receipt.NewKeeper(cdc)
	ok := order.NewKeeper(cdc, keyOrder, pk.Subspace(order.DefaultParamspace))
	ck := custodianunit.NewCUKeeper(
		cdc, keyAcc, pk.Subspace(custodianunit.DefaultParamspace), custodianunit.ProtoBaseCU,
	)
	ck.SetParams(ctx, custodianunit.DefaultParams())

	bankKeeper := transfer.NewBaseKeeper(cdc, keyTransfer, ck, nil, &tk, &ok, rk, nil, nil, pk.Subspace(transfer.DefaultParamspace), transfer.DefaultCodespace, blacklistedAddrs)
	bankKeeper.SetSendEnabled(ctx, true)

	maccPerms := map[string][]string{
		custodianunit.FeeCollectorName: nil,
		distr.ModuleName:               nil,
		staking.NotBondedPoolName:      []string{supply.Burner, supply.Staking},
		staking.BondedPoolName:         []string{supply.Burner, supply.Staking},
		mint.ModuleName:                []string{supply.Minter},
		types.ModuleName:               {supply.Minter, supply.Burner},
	}

	initPower := int64(10)
	numValidators := 4
	initTokens := sdk.TokensFromConsensusPower(initPower)
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(int64(numValidators))))
	bankKeeper.AddCoins(ctx, notBondedPool.GetAddress(), totalSupply)

	supplyKeeper := supply.NewKeeper(cdc, keySupply, ck, bankKeeper, maccPerms)
	supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	supplyKeeper.SetModuleAccount(ctx, bondPool)
	supplyKeeper.SetModuleAccount(ctx, notBondedPool)
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	stakingKeeper := staking.NewKeeper(cdc, keyStaking, tkeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	stakingParams := staking.DefaultParams()
	stakingKeeper.SetParams(ctx, stakingParams)

	distrKeeper := distr.NewKeeper(cdc, keyDistr, pk.Subspace(distr.DefaultParamspace), stakingKeeper, supplyKeeper, bankKeeper, distr.DefaultCodespace, custodianunit.FeeCollectorName, nil)

	mk := mint.NewKeeper(cdc, keyMint, pk.Subspace(mint.DefaultParamspace), stakingKeeper, supplyKeeper, custodianunit.FeeCollectorName)
	bankKeeper.SetStakingKeeper(stakingKeeper)
	tk.SetStakingKeeper(stakingKeeper)

	for _, ti := range token.TestTokenData {
		tk.CreateToken(ctx, ti)
	}

	hrc10k := NewKeeper(cdc, keyHrc10, pk.Subspace(DefaultParamspace), &tk, distrKeeper, supplyKeeper, rk, bankKeeper)
	hrc10k.SetParams(ctx, types.DefaultParams())

	// set the community pool to pay back the constant fee
	feePool := distr.InitialFeePool()
	feePool.CommunityPool = sdk.DecCoins{}
	distrKeeper.SetFeePool(ctx, feePool)

	return testEnv{cdc: cdc, ctx: ctx, transferKeeper: bankKeeper, ck: ck, tk: tk, ok: ok, rk: *rk, stakingkeeper: stakingKeeper, supplyKeeper: *supplyKeeper, pk: pk, mk: mk, dk: distrKeeper, hrc10k: hrc10k}
}

func getSymbolFromNewTokenRes(t *testing.T, res sdk.Result) string {
	require.True(t, res.IsOK())
	symbol := ""
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			if string(attribute.Key) == types.AttributeKeySymbol {
				symbol = string(attribute.Value)
			}
		}
	}
	require.NotEqual(t, symbol, "")
	return symbol
}
