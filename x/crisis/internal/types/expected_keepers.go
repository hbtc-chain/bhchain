package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

// SupplyKeeper defines the expected supply keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error)
}
