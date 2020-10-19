package exported

import (
	sdk "github.com/hbtc-chain/bhchain/types"

	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

// ModuleAccountI defines an CustodianUnit interface for modules that hold tokens in an escrow
type ModuleAccountI interface {
	exported.CustodianUnit

	GetName() string
	GetPermissions() []string
	HasPermission(string) bool
}

// SupplyI defines an inflationary supply interface for modules that handle
// token supply.
type SupplyI interface {
	GetTotal() sdk.Coins
	GetBurned() sdk.Coins
	SetTotal(total sdk.Coins) SupplyI

	Inflate(amount sdk.Coins) SupplyI
	Deflate(amount sdk.Coins) SupplyI

	String() string
	ValidateBasic() error
}
