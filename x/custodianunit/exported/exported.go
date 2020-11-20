package exported

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// CustodianUnit is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of CustodianUnit numbers for replay protection for previously pruned CUs,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements CustodianUnit.
type CustodianUnit interface {
	GetAddress() sdk.CUAddress
	SetAddress(sdk.CUAddress) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetSequence() uint64
	SetSequence(uint64) error

	String() string

	GetCUType() sdk.CUType

	SetCUType(sdk.CUType) error
	// for Operation CU
	SetSymbol(symbol string) error
	GetSymbol() string

	Validate() error
}

// VestingCU defines an CustodianUnit type that vests coins via a vesting schedule.
type VestingCU interface {
	CustodianUnit

	// Delegation and undelegation accounting that returns the resulting base
	// coins amount.
	TrackDelegation(blockTime time.Time, amount sdk.Coins)
	TrackUndelegation(amount sdk.Coins)

	GetVestedCoins(blockTime time.Time) sdk.Coins
	GetVestingCoins(blockTime time.Time) sdk.Coins

	GetStartTime() int64
	GetEndTime() int64

	GetOriginalVesting() sdk.Coins
	GetDelegatedFree() sdk.Coins
	GetDelegatedVesting() sdk.Coins
}
