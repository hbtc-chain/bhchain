package mint

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
)

func TestBegingBlocker(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)
	keeper := input.keeper
	supplyKeeper := input.supplyKeeper

	//height = 1
	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	account := supplyKeeper.GetModuleAccount(ctx, custodianunit.FeeCollectorName)

	require.EqualValues(t, DefaultParams().MintPerBlock, input.trk.GetAllBalance(ctx, account.GetAddress()).AmountOf(sdk.DefaultBondDenom))

	//height = 2
	header = abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	ctx = ctx.WithBlockHeader(header)
	BeginBlocker(ctx, keeper)
	require.EqualValues(t, DefaultParams().MintPerBlock.Mul(sdk.NewInt(2)), input.trk.GetAllBalance(ctx, account.GetAddress()).AmountOf(sdk.DefaultBondDenom))

}
