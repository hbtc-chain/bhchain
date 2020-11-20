package hrc10

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc10/types"
	"github.com/stretchr/testify/assert"
)

func TestQueryParams(t *testing.T) {
	expectedStr := "{\n  \"issue_token_fee\": \"1000000000000000000\"\n}"
	input := setupTestEnv(t)
	hk := input.hrc10k
	ctx := input.ctx
	cdc := input.cdc

	assert.Equal(t, types.DefaultParams(), hk.GetParams(ctx))

	bz, err := queryParams(ctx, hk)
	assert.Nil(t, err)
	assert.Equal(t, expectedStr, string(bz))

	var params types.Params
	err1 := cdc.UnmarshalJSON(bz, &params)
	assert.Nil(t, err1)
	assert.Equal(t, types.DefaultParams(), params)

	//modify the params
	params.IssueTokenFee = sdk.NewInt(200000000)
	hk.SetParams(ctx, params)
	assert.Equal(t, params, hk.GetParams(ctx))
}
