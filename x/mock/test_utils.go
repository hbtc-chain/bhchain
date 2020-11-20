package mock

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/hbtc-chain/bhchain/codec"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/hbtc-chain/bhchain/baseapp"
	sdk "github.com/hbtc-chain/bhchain/types"
)

// BigInterval is a representation of the interval [lo, hi), where
// lo and hi are both of type sdk.Int
type BigInterval struct {
	lo sdk.Int
	hi sdk.Int
}

// RandFromBigInterval chooses an interval uniformly from the provided list of
// BigIntervals, and then chooses an element from an interval uniformly at random.
func RandFromBigInterval(r *rand.Rand, intervals []BigInterval) sdk.Int {
	if len(intervals) == 0 {
		return sdk.ZeroInt()
	}

	interval := intervals[r.Intn(len(intervals))]

	lo := interval.lo
	hi := interval.hi

	diff := hi.Sub(lo)
	result := sdk.NewIntFromBigInt(new(big.Int).Rand(r, diff.BigInt()))
	result = result.Add(lo)

	return result
}

// CheckBalance checks the balance of an CU.
func CheckBalance(t *testing.T, app *App, addr sdk.CUAddress, exp sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	require.Equal(t, exp, app.TransferKeeper.GetAllBalance(ctxCheck, addr))
}

// CheckGenTx checks a generated signed transaction. The result of the check is
// compared against the parameter 'expPass'. A test assertion is made using the
// parameter 'expPass' against the result. A corresponding result is returned.
func CheckGenTx(
	t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg,
	seq []uint64, expPass bool, priv ...crypto.PrivKey,
) sdk.Result {
	tx := GenTx(msgs, seq, priv...)
	res := app.Check(tx)

	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}

	return res
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(t *testing.T, cdc *codec.Codec, app *baseapp.BaseApp, header abci.Header, msgs []sdk.Msg, seq []uint64, expSimPass, expPass bool, priv ...crypto.PrivKey) (sdk.Result, abci.ResponseBeginBlock, abci.ResponseEndBlock) {

	tx := GenTx(msgs, seq, priv...)

	txBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)
	require.Nil(t, err)

	// Must simulate now as CheckTx doesn't run Msgs anymore
	res := app.Simulate(txBytes, tx)

	if expSimPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}

	// Simulate a sending a transaction and committing a block
	responseBeginBlock := app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Deliver(tx)

	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}

	responseEndBlock := app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return res, responseBeginBlock, responseEndBlock
}

func AdvanceBlock(app *baseapp.BaseApp, lastHeight int64, n int) ([]abci.ResponseBeginBlock, []abci.ResponseEndBlock) {
	responseBeginBlocks := make([]abci.ResponseBeginBlock, n)
	responseEndBlocks := make([]abci.ResponseEndBlock, n)
	for i := 0; i < n; i++ {
		responseBeginBlocks[i] = app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: lastHeight + int64(i) + 1}})
		responseEndBlocks[i] = app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
		fmt.Printf("current block height %d\n", app.LastBlockHeight())
	}
	return responseBeginBlocks, responseEndBlocks
}
