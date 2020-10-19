package hrc20

import (
	"testing"

	"github.com/hbtc-chain/bhchain/x/hrc20/types"
	"github.com/stretchr/testify/assert"
)

func TestNewGenesisState(t *testing.T) {
	genState := NewGenesisState(types.DefaultParams())
	err := ValidateGenesis(genState)
	assert.Nil(t, err)

	assert.Equal(t, types.DefaultParams(), genState.Params)

	input := setupTestEnv(t)
	ctx := input.ctx
	keeper := input.hrc20k

	InitGenesis(ctx, keeper, genState)

	genState1 := ExportGenesis(ctx, keeper)
	assert.Equal(t, genState, genState1)
}

func TestDefaultGensisState(t *testing.T) {
	input := setupTestEnv(t)
	ctx := input.ctx
	keeper := input.hrc20k

	InitGenesis(ctx, keeper, DefaultGenesisState())
	genState1 := ExportGenesis(ctx, keeper)
	assert.Equal(t, DefaultGenesisState(), genState1)
}
