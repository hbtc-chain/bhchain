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

	GetCoins() sdk.Coins
	SetCoins(coins sdk.Coins) error
	AddCoins(coins sdk.Coins) sdk.Coins
	SubCoins(coins sdk.Coins) sdk.Coins

	GetCoinsHold() sdk.Coins
	SetCoinsHold(coins sdk.Coins) error
	AddCoinsHold(coins sdk.Coins) sdk.Coins
	SubCoinsHold(coins sdk.Coins) sdk.Coins

	String() string

	GetCUType() sdk.CUType

	SetCUType(sdk.CUType) error
	// for Operation CU
	GetSymbol() string

	GetAssets() []sdk.Asset

	GetAsset(denom string, epoch uint64) sdk.Asset
	GetAssetByAddr(denom string, addr string) sdk.Asset

	AddAsset(denom, address string, epoch uint64) error

	GetAssetAddress(denom string, epoch uint64) string

	SetAssetAddress(denom, address string, epoch uint64) error

	GetAssetCoinsHold() sdk.Coins
	AddAssetCoinsHold(coins sdk.Coins) sdk.Coins
	SubAssetCoinsHold(coins sdk.Coins) sdk.Coins

	GetAssetCoins() sdk.Coins
	AddAssetCoins(coins sdk.Coins) sdk.Coins
	SubAssetCoins(coins sdk.Coins) sdk.Coins

	GetAssetPubkey(epoch uint64) []byte
	SetAssetPubkey(pubkey []byte, epoch uint64) error
	GetAssetPubkeyEpoch() uint64

	GetGasUsed() sdk.Coins
	AddGasUsed(coins sdk.Coins) sdk.Coins
	SubGasUsed(coins sdk.Coins) sdk.Coins

	GetGasReceived() sdk.Coins
	AddGasReceived(coins sdk.Coins) sdk.Coins
	SubGasReceived(coins sdk.Coins) sdk.Coins

	GetGasRemained(chain string, addr string) sdk.Int
	AddGasRemained(chain string, addr string, amt sdk.Int)
	SubGasRemained(chain string, addr string, amt sdk.Int)

	GetBalanceFlows() []sdk.BalanceFlow

	ResetBalanceFlows()

	IsEnabledSendTx(chain string, addr string) bool
	SetEnableSendTx(enabled bool, chain string, addr string)

	GetNonce(chain string, addr string) uint64
	SetNonce(chain string, nonce uint64, addr string)

	SetMigrationStatus(status sdk.MigrationStatus)
	GetMigrationStatus() sdk.MigrationStatus

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
