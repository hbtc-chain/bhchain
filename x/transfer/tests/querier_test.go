package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbtc-chain/bhchain/x/transfer/keeper"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	authtypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func TestBalances(t *testing.T) {
	input := setupTestInput(t)
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/bank/%s", keeper.QueryBalance),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(&input.k)

	res, err := querier(input.ctx, []string{"balances"}, req)
	require.NotNil(t, err)
	require.Nil(t, res)

	_, _, addr := authtypes.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryBalanceParams(addr))
	res, err = querier(input.ctx, []string{"balances"}, req)
	require.Nil(t, err) // the CU does not exist, no error returned anyway
	require.NotNil(t, res)

	var coins sdk.Coins
	require.NoError(t, input.cdc.UnmarshalJSON(res, &coins))
	require.True(t, coins.IsZero())

	acc := input.ck.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
	acc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("foo", 10)))
	input.ck.SetCU(input.ctx, acc)
	res, err = querier(input.ctx, []string{"balances"}, req)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NoError(t, input.cdc.UnmarshalJSON(res, &coins))
	require.True(t, coins.AmountOf("foo").Equal(sdk.NewInt(10)))
}

func TestQuerierRouteNotFound(t *testing.T) {
	input := setupTestInput(t)
	req := abci.RequestQuery{
		Path: "custom/bank/notfound",
		Data: []byte{},
	}

	querier := keeper.NewQuerier(&input.k)
	_, err := querier(input.ctx, []string{"notfound"}, req)
	require.Error(t, err)
}
