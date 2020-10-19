package transfer_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/mock"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/transfer"
	"github.com/hbtc-chain/bhchain/x/transfer/keeper"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

var moduleAccAddr, _ = sdk.CUAddressFromBase58("HBCW1KVVAy28Fb1HLz38jAJ7pg7Ga8N91MQm")

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp, err := getBenchmarkMockApp()
	supply.RegisterCodec(mapp.Cdc)
	receipt.RegisterCodec(mapp.Cdc)
	require.NoError(t, err)
	return mapp
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, keeper *keeper.BaseKeeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		bankGenesis := transfer.DefaultGenesisState()
		transfer.InitGenesis(ctx, keeper, bankGenesis)

		return abci.ResponseInitChain{}
	}
}

// getBenchmarkMockApp initializes a mock application for this module, for purposes of benchmarking
// Any long term API support commitments do not apply to this function.
func getBenchmarkMockApp() (*mock.App, error) {
	mapp := mock.NewApp()
	types.RegisterCodec(mapp.Cdc)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[moduleAccAddr.String()] = true
	keyTransfer := sdk.NewKVStoreKey(transfer.StoreKey)

	rk := mapp.ReceiptKeeper
	bankKeeper := keeper.NewBaseKeeper(mapp.Cdc, keyTransfer,
		mapp.CUKeeper, nil, nil, rk, nil, nil,
		mapp.ParamsKeeper.Subspace(types.DefaultParamspace),
		types.DefaultCodespace,
		blacklistedAddrs,
	)
	mapp.Router().AddRoute(types.RouterKey, transfer.NewHandler(bankKeeper))
	mapp.SetInitChainer(getInitChainer(mapp, bankKeeper))

	err := mapp.CompleteSetup()
	return mapp, err
}

func BenchmarkOneBankSendTxPerBlock(b *testing.B) {
	benchmarkApp, _ := getBenchmarkMockApp()

	// Add an CU at genesis
	acc := &custodianunit.BaseCU{
		Address: addr1,
		// Some value conceivably higher than the benchmarks would ever go
		Coins: sdk.Coins{sdk.NewInt64Coin("foocoin", 100000000000)},
	}
	accs := []custodianunit.CU{acc}

	// Construct genesis state
	mock.SetGenesis(benchmarkApp, accs)
	// Precompute all txs
	txs := mock.GenSequenceOfTxs([]sdk.Msg{sendMsg1}, []uint64{uint64(0)}, b.N, priv1)
	b.ResetTimer()
	// Run this with a profiler, so its easy to distinguish what time comes from
	// Committing, and what time comes from Check/Deliver Tx.
	for i := 0; i < b.N; i++ {
		benchmarkApp.BeginBlock(abci.RequestBeginBlock{})
		x := benchmarkApp.Check(txs[i])
		if !x.IsOK() {
			panic("something is broken in checking transaction")
		}
		benchmarkApp.Deliver(txs[i])
		benchmarkApp.EndBlock(abci.RequestEndBlock{})
		benchmarkApp.Commit()
	}
}

func BenchmarkOneBankMultiSendTxPerBlock(b *testing.B) {
	benchmarkApp, _ := getBenchmarkMockApp()

	// Add an CU at genesis
	acc := &custodianunit.BaseCU{
		Address: addr1,
		// Some value conceivably higher than the benchmarks would ever go
		Coins: sdk.Coins{sdk.NewInt64Coin("foocoin", 100000000000)},
	}
	accs := []custodianunit.CU{acc}

	// Construct genesis state
	mock.SetGenesis(benchmarkApp, accs)
	// Precompute all txs
	txs := mock.GenSequenceOfTxs([]sdk.Msg{multiSendMsg1}, []uint64{uint64(0)}, b.N, priv1)
	b.ResetTimer()
	// Run this with a profiler, so its easy to distinguish what time comes from
	// Committing, and what time comes from Check/Deliver Tx.
	for i := 0; i < b.N; i++ {
		benchmarkApp.BeginBlock(abci.RequestBeginBlock{})
		x := benchmarkApp.Check(txs[i])
		if !x.IsOK() {
			panic("something is broken in checking transaction")
		}
		benchmarkApp.Deliver(txs[i])
		benchmarkApp.EndBlock(abci.RequestEndBlock{})
		benchmarkApp.Commit()
	}
}
