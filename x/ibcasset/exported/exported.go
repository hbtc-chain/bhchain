package exported

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

//
// Many complex conditions can be used in the concrete struct which implements cuibcasset.
type CUIBCAsset interface {
	GetAddress() sdk.CUAddress
	SetAddress(sdk.CUAddress) error // errors if already set.

	GetCUType() sdk.CUType
	SetCUType(sdk.CUType) error

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

	IsEnabledSendTx(chain string, addr string) bool
	SetEnableSendTx(enabled bool, chain string, addr string)

	GetNonce(chain string, addr string) uint64
	SetNonce(chain string, nonce uint64, addr string)

	SetMigrationStatus(status sdk.MigrationStatus)
	GetMigrationStatus() sdk.MigrationStatus
}
