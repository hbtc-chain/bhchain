package mint

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBegingBlocker(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)
	keeper := input.keeper
	supplyKeeper := input.supplyKeeper

	//height = 1
	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	require.Equal(t, sdk.ZeroInt(), supplyKeeper.GetModuleAccount(ctx, custodianunit.FeeCollectorName).GetCoins().AmountOf(sdk.DefaultBondDenom))

	//height = 2
	header = abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	ctx = ctx.WithBlockHeader(header)
	BeginBlocker(ctx, keeper)
	require.Equal(t, sdk.ZeroInt(), supplyKeeper.GetModuleAccount(ctx, custodianunit.FeeCollectorName).GetCoins().AmountOf(sdk.DefaultBondDenom))

}
