// nolint:deadcode unused
package gov

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	distrkeeper "github.com/hbtc-chain/bhchain/x/distribution/keeper"
	distrtype "github.com/hbtc-chain/bhchain/x/distribution/types"
	"github.com/hbtc-chain/bhchain/x/genaccounts"
	"github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/mock"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

var (
	valTokens  = sdk.TokensFromConsensusPower(4200000)
	initTokens = sdk.TokensFromConsensusPower(1000000000)
	valCoins   = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

type testInput struct {
	mApp     *mock.App
	keeper   Keeper
	router   Router
	sk       staking.Keeper
	dk       DistributionKeeper
	tk       transfer.Keeper
	addrs    []sdk.CUAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []custodianunit.CU) testInput {
	mApp := mock.NewApp()

	staking.RegisterCodec(mApp.Cdc)
	types.RegisterCodec(mApp.Cdc)
	supply.RegisterCodec(mApp.Cdc)
	distrtype.RegisterCodec(mApp.Cdc)
	receipt.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tKeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyDistr := sdk.NewKVStoreKey(distrtype.StoreKey)
	keyTransfer := mApp.KeyTransfer

	govAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Burner)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[govAcc.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true

	pk := mApp.ParamsKeeper

	rtr := NewRouter().
		AddRoute(RouterKey, ProposalHandler)

	bk := mApp.TransferKeeper

	maccPerms := map[string][]string{
		types.ModuleName:          []string{supply.Burner},
		staking.NotBondedPoolName: []string{supply.Burner, supply.Staking},
		staking.BondedPoolName:    []string{supply.Burner, supply.Staking},
		distrtype.ModuleName:      []string{supply.Burner, supply.Staking},
	}
	supplyKeeper := supply.NewKeeper(mApp.Cdc, keySupply, mApp.CUKeeper, bk, maccPerms)
	sk := staking.NewKeeper(mApp.Cdc, keyStaking, tKeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)

	sk.SetTransferKeeper(mApp.TransferKeeper)

	dk := distrkeeper.NewKeeper(mApp.Cdc, keyDistr, pk.Subspace(distrkeeper.DefaultParamspace), sk, supplyKeeper, mApp.TransferKeeper, distrtype.DefaultCodespace, custodianunit.FeeCollectorName, blacklistedAddrs)

	keeper := NewKeeper(mApp.Cdc, keyGov, pk, pk.Subspace(DefaultParamspace), supplyKeeper, sk, dk, mApp.TransferKeeper, DefaultCodespace, rtr)

	mApp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mApp.SetEndBlocker(getEndBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper, sk, *supplyKeeper, dk, genAccs, genState,
		[]supplyexported.ModuleAccountI{govAcc, notBondedPool, bondPool}))

	require.NoError(t, mApp.CompleteSetup(keyStaking, tKeyStaking, keyGov, keySupply, keyDistr, keyTransfer))

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

	return testInput{mApp, keeper, rtr, sk, dk, bk, addrs, pubKeys, privKeys}
}

// gov and staking endblocker
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{}
	}
}

// gov and staking initchainer
func getInitChainer(mapp *mock.App, keeper Keeper, stakingKeeper staking.Keeper, supplyKeeper supply.Keeper, distrKeeper distrkeeper.Keeper, accs []custodianunit.CU, genState GenesisState,
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

		//inital distribution without call distribution.InitGenesis to avoid import cycle
		distrGenesis := distrtype.DefaultGenesisState()
		distrKeeper.SetFeePool(ctx, distrGenesis.FeePool)

		if genState.IsEmpty() {
			//set gov MinInitDeposit to a small value to pass test case
			govGenesis := DefaultGenesisState()
			govGenesis.DepositParams.MinInitDeposit = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)}
			InitGenesis(ctx, keeper, supplyKeeper, govGenesis)
		} else {
			InitGenesis(ctx, keeper, supplyKeeper, genState)
		}
		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

// Sorts Addresses
func SortAddresses(addrs []sdk.CUAddress) {
	var byteAddrs [][]byte
	for _, addr := range addrs {
		byteAddrs = append(byteAddrs, addr.Bytes())
	}
	SortByteArrays(byteAddrs)
	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// implement `Interface` in sort package.
type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// Public
func SortByteArrays(src [][]byte) [][]byte {
	sorted := sortByteArrays(src)
	sort.Sort(sorted)
	return sorted
}

func testProposal() Content {
	return NewTextProposal("Test", "description")
}

const contextKeyBadProposal = "contextKeyBadProposal"

// badProposalHandler implements a governance proposal handler that is identical
// to the actual handler except this fails if the context doesn't contain a value
// for the key contextKeyBadProposal or if the value is false.
func badProposalHandler(ctx sdk.Context, c Content) sdk.Result {
	switch c.ProposalType() {
	case ProposalTypeText:
		v := ctx.Value(contextKeyBadProposal)

		if v == nil || !v.(bool) {
			return sdk.ErrInternal("proposal failed").Result()
		}

		return sdk.Result{}

	default:
		errMsg := fmt.Sprintf("unrecognized gov proposal type: %s", c.ProposalType())
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

// checks if two proposals are equal (note: slow, for tests only)
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	return bytes.Equal(types.ModuleCdc.MustMarshalBinaryBare(proposalA),
		types.ModuleCdc.MustMarshalBinaryBare(proposalB))
}

var (
	pubkeys = []crypto.PubKey{
		ed25519.GenPrivKey().PubKey(),
		ed25519.GenPrivKey().PubKey(),
		ed25519.GenPrivKey().PubKey(),
		ed25519.GenPrivKey().PubKey(),
	}

	testDescription     = staking.NewDescription("T", "E", "S", "T")
	testCommissionRates = staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func createValidators(t *testing.T, stakingHandler sdk.Handler, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {

		valTokens := sdk.TokensFromConsensusPower(powerAmt[i])
		valCreateMsg := staking.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			testDescription, testCommissionRates, sdk.OneInt(),
		)

		res := stakingHandler(ctx, valCreateMsg)
		require.True(t, res.IsOK())
	}
}

func readProposalID(t *testing.T, res sdk.Result) uint64 {
	require.True(t, res.IsOK())
	for _, event := range res.Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == "proposal_id" {
					proposalID, _ := strconv.ParseInt(string(attribute.Value), 10, 64)
					return uint64(proposalID)
				}
			}
		}
	}
	return 0
}
