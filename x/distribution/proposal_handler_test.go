package distribution

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/distribution/types"
	"github.com/stretchr/testify/require"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.CUAddress(delPk1.Address())

	amount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
)

func testProposal(recipient sdk.CUAddress, amount sdk.Coins) types.CommunityPoolSpendProposal {
	return types.NewCommunityPoolSpendProposal(
		"Test",
		"description",
		recipient,
		amount,
	)
}

func TestProposalHandlerPassed(t *testing.T) {
	ctx, cuKeeper, keeper, _, supplyKeeper := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	// add coins to the module CU
	macc := keeper.GetDistributionAccount(ctx)
	err := macc.SetCoins(macc.GetCoins().Add(amount))
	require.NoError(t, err)

	supplyKeeper.SetModuleAccount(ctx, macc)

	CU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, recipient)
	require.True(t, CU.GetCoins().IsZero())
	cuKeeper.SetCU(ctx, CU)

	feePool := keeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoins(amount)
	keeper.SetFeePool(ctx, feePool)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.Equal(t, sdk.CodeOK, hdlr(ctx, tp).Code)
	require.Equal(t, cuKeeper.GetCU(ctx, recipient).GetCoins(), amount)
}

func TestProposalHandlerFailed(t *testing.T) {
	ctx, cuKeeper, keeper, _, _ := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	CU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, recipient)
	require.True(t, CU.GetCoins().IsZero())
	cuKeeper.SetCU(ctx, CU)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.NotEqual(t, sdk.CodeOK, hdlr(ctx, tp).Code)
	require.True(t, cuKeeper.GetCU(ctx, recipient).GetCoins().IsZero())
}
