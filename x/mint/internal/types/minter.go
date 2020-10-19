package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"math/big"
)

const (
	NativeTokenDecimals = 18 //TODO(Keep), should retrieve from token
)

var PrecisionMutipler = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(NativeTokenDecimals), nil))

// Minter represents the minting state.
type Minter struct{}

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter() Minter {
	return Minter{}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter() Minter {
	return NewMinter()
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// validate minter
func ValidateMinter(minter Minter) error {
	return nil
}

func (m Minter) String() string {
	return ""
}
