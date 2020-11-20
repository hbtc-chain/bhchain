package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hbtc-chain/bhchain/x/transfer/keeper"

	sdk "github.com/hbtc-chain/bhchain/types"
	authtypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func TestBalances(t *testing.T) {
	input := setupTestInput(t)
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/bank/%s", types.QueryBalance),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(input.k)

	res, err := querier(input.ctx, []string{"balances"}, req)
	require.NotNil(t, err)
	require.Nil(t, res)

	_, _, addr := authtypes.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryAllBalanceParams(addr))
	res, err = querier(input.ctx, []string{"balances"}, req)
	require.Nil(t, err) // the CU does not exist, no error returned anyway
	require.NotNil(t, res)

	var coins types.ResAllBalance
	require.NoError(t, input.cdc.UnmarshalJSON(res, &coins))
	require.True(t, coins.Available.IsZero())

	acc := input.ck.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
	testSetCUCoins(input.ctx, input.trk, acc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("foo", 10)))
	//acc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("foo", 10)))
	input.ck.SetCU(input.ctx, acc)
	res, err = querier(input.ctx, []string{"balances"}, req)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NoError(t, input.cdc.UnmarshalJSON(res, &coins))
	require.True(t, coins.Available.AmountOf("foo").Equal(sdk.NewInt(10)))
}

func TestQuerierRouteNotFound(t *testing.T) {
	input := setupTestInput(t)
	req := abci.RequestQuery{
		Path: "custom/bank/notfound",
		Data: []byte{},
	}

	querier := keeper.NewQuerier(input.k)
	_, err := querier(input.ctx, []string{"notfound"}, req)
	require.Error(t, err)
}
