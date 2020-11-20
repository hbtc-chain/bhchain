package mint

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/genaccounts"
	"github.com/hbtc-chain/bhchain/x/mint/internal/types"
	"github.com/hbtc-chain/bhchain/x/mock"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

var (
	valTokens  = sdk.TokensFromConsensusPower(500000)
	initTokens = sdk.TokensFromConsensusPower(1000000000)
	valCoins   = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	//initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

type testInput struct {
	mApp         *mock.App
	keeper       Keeper
	sk           staking.Keeper
	supplyKeeper supply.Keeper
	trk          transfer.Keeper
	addrs        []sdk.CUAddress
	pubKeys      []crypto.PubKey
	privKeys     []crypto.PrivKey
}

type testInputForProposal struct {
	ctx    sdk.Context
	cdc    *codec.Codec
	keeper Keeper
}

func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []custodianunit.CU) testInput {
	mApp := mock.NewApp()

	staking.RegisterCodec(mApp.Cdc)
	types.RegisterCodec(mApp.Cdc)
	supply.RegisterCodec(mApp.Cdc)
	receipt.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tKeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyMint := sdk.NewKVStoreKey(StoreKey)
	keyTransfer := mApp.KeyTransfer

	feeCollectorAcc := supply.NewEmptyModuleAccount(custodianunit.FeeCollectorName)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)
	minterAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Minter)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true
	blacklistedAddrs[minterAcc.String()] = true

	pk := mApp.ParamsKeeper
	bk := mApp.TransferKeeper
	maccPerms := map[string][]string{
		custodianunit.FeeCollectorName: nil,
		types.ModuleName:               []string{supply.Minter},
		staking.NotBondedPoolName:      []string{supply.Burner, supply.Staking},
		staking.BondedPoolName:         []string{supply.Burner, supply.Staking},
	}
	supplyKeeper := supply.NewKeeper(mApp.Cdc, keySupply, mApp.CUKeeper, bk, maccPerms)
	sk := staking.NewKeeper(mApp.Cdc, keyStaking, tKeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	sk.SetTransferKeeper(bk)

	keeper := NewKeeper(mApp.Cdc, keyMint, pk.Subspace(DefaultParamspace), sk, supplyKeeper, custodianunit.FeeCollectorName)

	mApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mApp.SetBeginBlocker(getBeginBlocker(keeper))

	mApp.SetInitChainer(getInitChainer(mApp, keeper, sk, *supplyKeeper, genAccs, genState,
		[]supplyexported.ModuleAccountI{feeCollectorAcc, minterAcc, notBondedPool, bondPool}))

	require.NoError(t, mApp.CompleteSetup(keyStaking, tKeyStaking, keySupply, keyMint, keyTransfer))

	var (
		addrs    []sdk.CUAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	genAccounts := []genaccounts.GenesisCU{}
	if genAccs == nil || len(genAccs) == 0 {
		genAccounts, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs, valCoins)
	} else {
		for _, acc := range genAccs {
			genAccounts = append(genAccounts, genaccounts.GenesisCU{
				Type:     acc.GetCUType(),
				PubKey:   acc.GetPubKey(),
				Sequence: acc.GetSequence(),
			})
		}
	}

	mock.SetGenesis(mApp, genAccounts)

	return testInput{mApp, keeper, sk, *supplyKeeper, bk, addrs, pubKeys, privKeys}
}

//mint endblocker
func getBeginBlocker(keeper Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		BeginBlocker(ctx, keeper)
		return abci.ResponseBeginBlock{}
	}
}

// gov and staking initchainer
func getInitChainer(mapp *mock.App, keeper Keeper, stakingKeeper staking.Keeper, supplyKeeper supply.Keeper, accs []custodianunit.CU, genState GenesisState,
	blacklistedAddrs []supplyexported.ModuleAccountI) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakingGenesis := staking.DefaultGenesisState()

		totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(int64(len(mapp.GenesisAccounts)))))
		supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

		// set module accounts
		for _, macc := range blacklistedAddrs {
			supplyKeeper.SetModuleAccount(ctx, macc)
		}

		validators := staking.InitGenesis(ctx, stakingKeeper, mapp.CUKeeper, supplyKeeper, stakingGenesis)

		//inital minter
		var mintGenesis = DefaultGenesisState()
		if !genState.IsEmpty() {
			mintGenesis = genState
		}

		InitGenesis(ctx, keeper, mintGenesis)

		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}
