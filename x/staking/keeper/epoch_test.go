package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestGet(t *testing.T) {
	ctx, _, keeper, _ := CreateTestInput(t, false, 0)

	height := int64(254)
	ctx = ctx.WithBlockHeader(abci.Header{Height: height})
	vals := addrVals[2 : 2+4]
	epoch1 := keeper.GetCurrentEpoch(ctx)
	require.EqualValues(t, sdk.NewEpoch(1, uint64(0), 0, nil, true), epoch1)

	ctx = ctx.WithBlockHeader(abci.Header{Height: height + 1})
	epoch2 := keeper.StartNewEpoch(ctx, vals)
	require.EqualValues(t, sdk.NewEpoch(2, uint64(height+2), 0, vals, false), epoch2)

	epoch1, _ = keeper.GetEpoch(ctx, epoch1.Index)
	require.EqualValues(t, height+1, epoch1.EndBlockNum)

	ctx = ctx.WithBlockHeader(abci.Header{Height: height + 5})
	epoch3 := keeper.StartNewEpoch(ctx, vals)
	require.EqualValues(t, sdk.NewEpoch(3, uint64(height+6), 0, vals, false), epoch3)

	epoch := keeper.GetEpochByHeight(ctx, uint64(height+8))
	require.EqualValues(t, epoch3, epoch)

	epoch = keeper.GetEpochByHeight(ctx, uint64(height+2))
	res, _ := keeper.GetEpoch(ctx, epoch2.Index)
	require.EqualValues(t, res, epoch)
}
