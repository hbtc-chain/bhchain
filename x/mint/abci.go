package mint

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/mint/internal/types"
)

func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter & params
	params := k.GetParams(ctx)
	//minter := k.GetMinter(ctx)

	// mintedCoins
	mintedCoins := sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, params.MintPerBlock))

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector CU
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			//sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoins.AmountOf(k.GetParams(ctx).MintDenom).String()),
		),
	)
}
