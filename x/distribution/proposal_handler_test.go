package distribution

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/distribution/keeper"
	"github.com/hbtc-chain/bhchain/x/distribution/types"
	"github.com/hbtc-chain/bhchain/x/params"
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
	ctx, cuKeeper, tk, keeper, _, supplyKeeper := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	// add coins to the module CU
	macc := keeper.GetDistributionAccount(ctx)
	tk.AddCoins(ctx, macc.GetAddress(), amount)

	supplyKeeper.SetModuleAccount(ctx, macc)

	CU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, recipient)

	require.True(t, tk.GetAllBalance(ctx, CU.GetAddress()).IsZero())
	cuKeeper.SetCU(ctx, CU)

	feePool := keeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoins(amount)
	keeper.SetFeePool(ctx, feePool)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.Equal(t, sdk.CodeOK, hdlr(ctx, tp).Code)
	require.Equal(t, tk.GetAllBalance(ctx, recipient), amount)
}

func TestProposalHandlerFailed(t *testing.T) {
	ctx, cuKeeper, tk, keeper, _, _ := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	CU := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, recipient)
	require.True(t, tk.GetAllBalance(ctx, CU.GetAddress()).IsZero())
	cuKeeper.SetCU(ctx, CU)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.NotEqual(t, sdk.CodeOK, hdlr(ctx, tp).Code)
	require.True(t, tk.GetAllBalance(ctx, recipient).IsZero())
}

func TestChangeParamsProposal(t *testing.T) {
	testCases := []struct {
		key    []byte
		number string
		except bool
	}{
		{keeper.ParamStoreKeyKeyNodeReward, `"0.05"`, true},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.15"`, true},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.1"`, true},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.02"`, false},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.20"`, false},
		{keeper.ParamStoreKeyKeyNodeReward, `"0"`, false},
		{keeper.ParamStoreKeyKeyNodeReward, `"1"`, false},

		{keeper.ParamStoreKeyKeyNodeReward, `"0.05"`, true},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.15"`, true},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.12"`, true},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.02"`, false},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.20"`, false},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.00001"`, false},
		{keeper.ParamStoreKeyKeyNodeReward, `"0.222222"`, false},
	}

	for _, c := range testCases {
		pc := params.NewParamChange(DefaultParamspace, string(c.key), c.number)
		pcp := params.NewParameterChangeProposal("change distr", "change distr", []params.ParamChange{pc})
		if c.except {
			require.Nil(t, pcp.ValidateBasic())
		} else {
			require.Error(t, pcp.ValidateBasic())
		}
	}
}
