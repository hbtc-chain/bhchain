package mint

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesisStateEqual(t *testing.T) {
	gs1 := DefaultGenesisState()
	gs2 := DefaultGenesisState()

	require.True(t, gs1.Equal(gs2))
	gs2.Params.Inflation = sdk.NewDecWithPrec(1, 1)
	require.False(t, gs1.Equal(gs2))

	gs3 := NewGenesisState(gs1.Minter, gs1.Params)
	require.True(t, gs1.Equal(gs3))

	gs4 := NewGenesisState(gs2.Minter, gs2.Params)
	require.True(t, gs2.Equal(gs4))
}

func TestGenesisStateIsEmpty(t *testing.T) {
	gs1 := GenesisState{}
	require.True(t, gs1.IsEmpty())

	gs1.Params.MintDenom = sdk.DefaultBondDenom
	require.False(t, gs1.IsEmpty())

	gs1 = DefaultGenesisState()
	require.False(t, gs1.IsEmpty())
}
