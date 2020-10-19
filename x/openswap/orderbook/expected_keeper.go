package orderbook

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

type OpenswapKeeper interface {
	IteratorAllUnfinishedOrder(ctx sdk.Context, f func(*types.Order))
}
