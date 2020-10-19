package gov

import (
	"strings"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/stretchr/testify/require"
)

func TestInvalidMsg(t *testing.T) {
	k := Keeper{}
	h := NewHandler(k)

	res := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized gov message type"))
}

func TestSubmitProposalFailInsufficintInitDeposit(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(input.keeper)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	minDeposit := input.keeper.GetDepositParams(ctx).MinInitDeposit.AmountOf(sdk.DefaultBondDenom).Int64()
	for i := int64(0); i < minDeposit; i++ {
		newProposalMsg := NewMsgSubmitProposal(
			ContentFromProposalType("test", "test", ProposalTypeText),
			sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, i)},
			input.addrs[0], 0,
		)
		res := govHandler(ctx, newProposalMsg)
		require.False(t, res.IsOK())
	}

	newProposalMsg := NewMsgSubmitProposal(
		ContentFromProposalType("test", "test", ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, minDeposit)},
		input.addrs[0], 0,
	)
	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

}
