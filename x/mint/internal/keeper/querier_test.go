package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbtc-chain/bhchain/x/mint/internal/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewQuerier(t *testing.T) {
	input := newTestInput(t)
	querier := NewQuerier(input.mintKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(input.ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)

	_, err = querier(input.ctx, []string{"foo"}, query)
	require.Error(t, err)
}

func TestQueryParams(t *testing.T) {
	input := newTestInput(t)

	var params types.Params

	res, sdkErr := queryParams(input.ctx, input.mintKeeper)

	require.NoError(t, sdkErr)

	err := input.cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, input.mintKeeper.GetParams(input.ctx), params)
}
