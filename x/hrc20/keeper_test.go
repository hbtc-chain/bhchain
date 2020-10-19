package hrc20

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSetParams(t *testing.T) {
	input := setupTestEnv(t)
	hk := input.hrc20k
	ctx := input.ctx

	param := hk.GetParams(ctx)

	assert.Equal(t, types.DefaultParams(), param)

	param.IssueTokenFee = sdk.NewInt(560000)
	hk.SetParams(ctx, param)

	param1 := hk.GetParams(ctx)
	assert.Equal(t, param, param1)
}
