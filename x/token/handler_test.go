package token

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

var ethSymbol = sdk.Symbol("eth")

var symbols = []sdk.Symbol{"eth", "btc", "usdt"}

type Args struct {
	symbol sdk.Symbol
	height uint64
}

func Test_handleMsgSynGasPriceNormal(t *testing.T) {
	env := setupUnitTestEnv()
	gasPrice := sdk.NewInt(100)
	for i := 0; i < 4; i++ {
		got := handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, i, 95, gasPrice, ethSymbol))
		require.Equal(t, sdk.CodeOK, got.Code, got)
	}

	tokenInfo := env.tk.GetTokenInfo(env.ctx, ethSymbol)
	require.Equal(t, gasPrice.String(), tokenInfo.GasPrice.String())
}

func Test_handleMsgSynGasPriceMulti(t *testing.T) {
	env := setupUnitTestEnv()
	gasPrice := sdk.NewInt(100)

	args := make([]Args, 0)
	for _, symbol := range symbols {
		for height := 1; height < 10; height++ {
			args = append(args, Args{symbol: symbol, height: uint64(100 - height)})
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(args), func(i, j int) { args[i], args[j] = args[j], args[i] })
	for _, arg := range args {
		for i := 0; i < 4; i++ {
			got := handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, i, arg.height, gasPrice, arg.symbol))
			require.Equal(t, sdk.CodeOK, got.Code)
		}
	}

	for _, symbol := range symbols {
		tokenInfo := env.tk.GetTokenInfo(env.ctx, symbol)
		require.Equal(t, gasPrice, tokenInfo.GasPrice)
	}
}

func Test_handleMsgSynGasPriceRandomPrice(t *testing.T) {
	env := setupUnitTestEnv()

	args := make([]Args, 0)
	for _, symbol := range symbols {
		for height := 1; height < 10; height++ {
			args = append(args, Args{symbol: symbol, height: uint64(100 - height)})
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(args), func(i, j int) { args[i], args[j] = args[j], args[i] })
	for _, arg := range args {
		for i := 0; i < 4; i++ {
			got := handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, i, arg.height, sdk.NewInt(rand.Int63()), arg.symbol))
			require.Equal(t, sdk.CodeOK, got.Code)
		}
	}
}

func Test_handleMsgSynGasPriceMissing(t *testing.T) {
	env := setupUnitTestEnv()

	gasPrice := sdk.NewInt(100)
	for i := 0; i < 4; i++ {
		got := handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, i, 95, gasPrice, ethSymbol))
		require.Equal(t, sdk.CodeOK, got.Code)
	}

	tokenInfo := env.tk.GetTokenInfo(env.ctx, ethSymbol)
	require.Equal(t, gasPrice.String(), tokenInfo.GasPrice.String())
}

func Test_handleMsgSynGasPriceAllDifferent(t *testing.T) {
	env := setupUnitTestEnv()

	gasPrice := sdk.NewInt(100)
	got := handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, 0, 95, gasPrice, ethSymbol))
	require.Equal(t, sdk.CodeOK, got.Code)

	got = handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, 1, 95, gasPrice.MulRaw(2), ethSymbol))
	require.Equal(t, sdk.CodeOK, got.Code)

	got = handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, 2, 95, gasPrice.MulRaw(3), ethSymbol))
	require.Equal(t, sdk.CodeOK, got.Code)

	got = handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, 3, 95, gasPrice.MulRaw(4), ethSymbol))
	require.Equal(t, sdk.CodeOK, got.Code)

	tokenInfo := env.tk.GetTokenInfo(env.ctx, ethSymbol)
	// default
	require.Equal(t, sdk.NewInt(1000), tokenInfo.GasPrice)
}

func Test_handleMsgSynGasPriceInvalidHeight(t *testing.T) {
	env := setupUnitTestEnv()

	gasPrice := sdk.NewInt(100)
	got := handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, 0, 100, gasPrice, ethSymbol))
	require.Equal(t, sdk.CodeInvalidTx, got.Code)

	got = handleMsgSynGasPrice(env.ctx, env.tk, newSyncGasPriceMsg(env, 0, 10, gasPrice, ethSymbol))
	require.Equal(t, sdk.CodeInvalidTx, got.Code)
}

func Test_handleMsgSynGasPricePunish(t *testing.T) {
	env := setupUnitTestEnv()
	gasPrice := sdk.ZeroInt()
	ctx := env.ctx
	for i := 1; i <= 11; i++ {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 10)
		for i := 0; i < 4; i++ {
			if i == 0 {
				gasPrice = sdk.NewInt(10000)
			} else {
				gasPrice = sdk.NewInt(100)
			}
			got := handleMsgSynGasPrice(ctx, env.tk, newSyncGasPriceMsg(env, i, uint64(ctx.BlockHeight()-1), gasPrice, ethSymbol))
			require.Equal(t, sdk.CodeOK, got.Code, sdk.CodeToDefaultMsg(got.Code))
		}
		tokenInfo := env.tk.GetTokenInfo(env.ctx, ethSymbol)
		require.Equal(t, "100", tokenInfo.GasPrice.String())
	}
	ctx = ctx.WithBlockHeight(1000)
	env.mockStakingKeeper.On("JailByOperator", mock.Anything, env.validators[0].OperatorAddress)
	env.mockStakingKeeper.On("SlashByOperator", mock.Anything, env.validators[0].OperatorAddress, mock.Anything, mock.Anything)
	env.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	env.mockStakingKeeper.AssertNumberOfCalls(t, "JailByOperator", 1)
	env.mockStakingKeeper.AssertNumberOfCalls(t, "SlashByOperator", 1)
}

func newSyncGasPriceMsg(env testEnv, validatorIndex int, height uint64, gasPrice sdk.Int, symbol sdk.Symbol) types.MsgSynGasPrice {
	return types.MsgSynGasPrice{
		From:   sdk.CUAddress(env.validators[validatorIndex].OperatorAddress).String(),
		Height: height,
		GasPrice: []sdk.TokensGasPrice{
			{
				Chain:    symbol.String(),
				GasPrice: gasPrice,
			},
		},
	}
}
