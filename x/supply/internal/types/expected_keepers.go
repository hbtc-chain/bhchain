package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

// CUKeeper defines the expected CustodianUnit keeper (noalias)
type CUKeeper interface {
	IterateCUs(ctx sdk.Context, process func(exported.CustodianUnit) (stop bool))
	GetCU(sdk.Context, sdk.CUAddress) exported.CustodianUnit
	SetCU(sdk.Context, exported.CustodianUnit)
	NewCU(sdk.Context, exported.CustodianUnit) exported.CustodianUnit
}

// BankKeeper defines the expected bank keeper (noalias)
type TransferKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.CUAddress, toAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, []sdk.Flow, sdk.Error)
	DelegateCoins(ctx sdk.Context, fromAdd, toAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)
	UndelegateCoins(ctx sdk.Context, fromAddr, toAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error)

	SubCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.CUAddress, amt sdk.Coins) (sdk.Coins, []sdk.Flow, sdk.Error)
}
