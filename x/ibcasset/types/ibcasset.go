package types

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/hbtc-chain/bhchain/base58"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset/exported"
)

type CUIBCAsset struct {
	Address          sdk.CUAddress       `json:"address" yaml:"address"`
	Type             sdk.CUType          `json:"cu_type" yaml:"type"`
	Assets           []sdk.Asset         `json:"assets" yaml:"assets"`
	AssetCoins       sdk.Coins           `json:"asset_coins" yaml:"asset_coins"`
	AssetCoinsHold   sdk.Coins           `json:"asset_coins_hold" yaml:"asset_coins_hold"`
	AssetPubkey      []byte              `json:"asset_pubkey" yaml:"asset_pubkey"`
	AssetPubkeyEpoch uint64              `json:"asset_pubkey_epoch" yaml:"asset_pubkey_epoch"`
	GasUsed          sdk.Coins           `json:"gas_used" yaml:"gas_used"`
	GasReceived      sdk.Coins           `json:"gas_received" yaml:"gas_received"`
	MigrationStatus  sdk.MigrationStatus `json:"migration_status" yaml:"migration_status"`
}

func ProtoBaseCUIBCAsset() exported.CUIBCAsset {
	return &CUIBCAsset{}
}

func (cuAst *CUIBCAsset) GetAddress() sdk.CUAddress {
	return cuAst.Address
}

func (cuAst *CUIBCAsset) SetAddress(cuaddress sdk.CUAddress) error {
	if len(cuAst.Address) != 0 {
		return errors.New("cannot override custodian unit address")
	}

	cuAst.Address = cuaddress
	return nil
}

func (cuAst *CUIBCAsset) GetCUType() sdk.CUType {
	return cuAst.Type
}

// SetCUType set the custodian unit type
// if the type of custodian unit already defined return error
func (cuAst *CUIBCAsset) SetCUType(cuType sdk.CUType) error {
	if cuAst.Type != 0 {
		return errors.New("cannot override custodian unit type")
	}
	cuAst.Type = cuType
	return nil
}

func (cuAst *CUIBCAsset) IsEnabledSendTx(chain string, addr string) bool {
	for _, ast := range cuAst.Assets {
		if ast.Denom == chain && ast.Address == addr {
			return ast.EnableSendTx
		}
	}
	return false
}

func (cuAst *CUIBCAsset) SetEnableSendTx(enabled bool, chain string, addr string) {
	for i, ast := range cuAst.Assets {
		if ast.Denom == chain && ast.Address == addr {
			cuAst.Assets[i].EnableSendTx = enabled
			break
		}
	}
}

func (cuAst *CUIBCAsset) GetAssets() []sdk.Asset {
	return cuAst.Assets
}

func (cuAst *CUIBCAsset) GetAsset(denom string, epoch uint64) sdk.Asset {
	for _, asset := range cuAst.Assets {
		if asset.Denom == denom && asset.Epoch == epoch {
			return asset
		}
	}
	return sdk.NilAsset
}

func (cuAst *CUIBCAsset) SetMigrationStatus(status sdk.MigrationStatus) {
	cuAst.MigrationStatus = status
}

func (cuAst *CUIBCAsset) GetMigrationStatus() sdk.MigrationStatus {
	return cuAst.MigrationStatus
}

func (cuAst *CUIBCAsset) GetAssetByAddr(denom string, addr string) sdk.Asset {
	for _, asset := range cuAst.Assets {
		if asset.Denom == denom && asset.Address == addr {
			return asset
		}
	}
	return sdk.NilAsset
}

func (cuAst *CUIBCAsset) GetAssetAddress(denom string, epoch uint64) string {
	as := cuAst.GetAsset(denom, epoch)
	if as == sdk.NilAsset {
		return ""
	}
	return as.Address
}

func (cuAst *CUIBCAsset) GetAssetPubkey(epoch uint64) []byte {
	if cuAst.AssetPubkeyEpoch != epoch {
		return nil
	}
	return cuAst.AssetPubkey
}

func (cuAst *CUIBCAsset) GetAssetPubkeyEpoch() uint64 {
	return cuAst.AssetPubkeyEpoch
}

func (cuAst *CUIBCAsset) SetAssetPubkey(pubkey []byte, epoch uint64) error {
	cuAst.AssetPubkey = pubkey
	cuAst.AssetPubkeyEpoch = epoch
	return nil
}

func (cuAst *CUIBCAsset) AddAsset(denom, address string, epoch uint64) error {
	for i := 0; i < len(cuAst.Assets); i++ {
		asset := cuAst.Assets[i]
		if asset.Denom == denom && asset.Epoch == epoch {
			return errors.New("asset already exist")
		}

		if asset.Denom == denom && asset.Epoch == 0 {
			cuAst.Assets[i].Address = address
			cuAst.Assets[i].Epoch = epoch
		}
	}

	//delete old epoch's assert info
	for i := 0; i < len(cuAst.Assets); i++ {
		asset := cuAst.Assets[i]
		if epoch >= 3 && epoch-asset.Epoch >= 2 {
			if !asset.GasRemained.IsZero() {
				cuAst.AddGasUsed(sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.GasRemained)))
			}
			cuAst.Assets = append(cuAst.Assets[:i], cuAst.Assets[i+1:]...)
			i--
		}
	}

	cuAst.Assets = append(cuAst.Assets, sdk.NewAsset(denom, address, epoch, true))

	return nil
}

func (cuAst *CUIBCAsset) SetAssetAddress(denom, address string, epoch uint64) error {
	for i, asset := range cuAst.Assets {
		if asset.Denom == denom && asset.Epoch == 0 {
			cuAst.Assets[i].Address = address
			cuAst.Assets[i].Epoch = epoch
			return nil
		}
	}
	return cuAst.AddAsset(denom, address, epoch)
}

func (cuAst *CUIBCAsset) GetNonce(denom string, addr string) uint64 {
	for _, asset := range cuAst.Assets {
		if asset.Denom == denom && asset.Address == addr {
			return asset.Nonce
		}
	}

	return 0
}

func (cuAst *CUIBCAsset) SetNonce(denom string, nonce uint64, addr string) {
	for i, asset := range cuAst.Assets {
		if asset.Denom == denom && asset.Address == addr {
			cuAst.Assets[i].Nonce = nonce
			return
		}
	}
}

func (cuAst *CUIBCAsset) GetAssetCoinsHold() sdk.Coins {
	return cuAst.AssetCoinsHold
}
func (cuAst *CUIBCAsset) AddAssetCoinsHold(coins sdk.Coins) sdk.Coins {
	cuAst.AssetCoinsHold = cuAst.AssetCoinsHold.Add(coins)
	return cuAst.AssetCoinsHold
}
func (cuAst *CUIBCAsset) SubAssetCoinsHold(coins sdk.Coins) sdk.Coins {
	cuAst.AssetCoinsHold = cuAst.AssetCoinsHold.Sub(coins)
	return cuAst.AssetCoinsHold
}

func (cuAst *CUIBCAsset) GetAssetCoins() sdk.Coins {
	return cuAst.AssetCoins
}

func (cuAst *CUIBCAsset) AddAssetCoins(coins sdk.Coins) sdk.Coins {
	cuAst.AssetCoins = cuAst.AssetCoins.Add(coins)
	return cuAst.AssetCoins
}

func (cuAst *CUIBCAsset) SubAssetCoins(coins sdk.Coins) sdk.Coins {
	cuAst.AssetCoins = cuAst.AssetCoins.Sub(coins)
	return cuAst.AssetCoins
}

func (cuAst *CUIBCAsset) GetGasRemained(chain string, addr string) sdk.Int {
	asset := cuAst.GetAssetByAddr(chain, addr)
	if asset == sdk.NilAsset {
		return sdk.ZeroInt()
	}
	return asset.GasRemained
}

func (cuAst *CUIBCAsset) AddGasRemained(chain string, addr string, amt sdk.Int) {
	for i, asset := range cuAst.Assets {
		if asset.Denom == chain && asset.Address == addr {
			cuAst.Assets[i].GasRemained = asset.GasRemained.Add(amt)
			return
		}
	}
}

func (cuAst *CUIBCAsset) SubGasRemained(chain string, addr string, amt sdk.Int) {
	for i, asset := range cuAst.Assets {
		if asset.Denom == chain && asset.Address == addr {
			cuAst.Assets[i].GasRemained = asset.GasRemained.Sub(amt)
			return
		}
	}
}

func (cuAst *CUIBCAsset) GetGasUsed() sdk.Coins {
	return cuAst.GasUsed
}

func (cuAst *CUIBCAsset) AddGasUsed(coins sdk.Coins) sdk.Coins {
	cuAst.GasUsed = cuAst.GasUsed.Add(coins)
	return cuAst.GasUsed
}

func (cuAst *CUIBCAsset) SubGasUsed(coins sdk.Coins) sdk.Coins {
	cuAst.GasUsed = cuAst.GasUsed.Sub(coins)
	return cuAst.GasUsed
}

func (cuAst *CUIBCAsset) GetGasReceived() sdk.Coins {
	return cuAst.GasReceived
}

func (cuAst *CUIBCAsset) AddGasReceived(coins sdk.Coins) sdk.Coins {
	cuAst.GasReceived = cuAst.GasReceived.Add(coins)
	return cuAst.GasReceived
}

func (cuAst *CUIBCAsset) SubGasReceived(coins sdk.Coins) sdk.Coins {
	cuAst.GasReceived = cuAst.GasReceived.Sub(coins)
	return cuAst.GasReceived
}

// MarshalYAML returns the YAML representation of an custodianunit.
func (cuAst *CUIBCAsset) MarshalYAML() (interface{}, error) {
	var bs []byte
	var err error
	var assetPubkey string
	if cuAst.AssetPubkey != nil {
		assetPubkey = base58.Encode(cuAst.AssetPubkey)
	}
	bs, err = yaml.Marshal(struct {
		Address          sdk.CUAddress
		Type             sdk.CUType
		Assets           []sdk.Asset
		AssetCoins       sdk.Coins
		AssetCoinsHold   sdk.Coins
		AssetPubkey      string
		AssetPubkeyEpoch uint64
		GasUsed          sdk.Coins
		GasReceived      sdk.Coins
		MigrationStatu   sdk.MigrationStatus
	}{
		Address:          cuAst.Address,
		Type:             cuAst.Type,
		AssetCoins:       cuAst.AssetCoins,
		AssetCoinsHold:   cuAst.AssetCoinsHold,
		GasUsed:          cuAst.GasUsed,
		GasReceived:      cuAst.GasReceived,
		Assets:           cuAst.Assets,
		AssetPubkey:      assetPubkey,
		AssetPubkeyEpoch: cuAst.AssetPubkeyEpoch,
		MigrationStatu:   cuAst.MigrationStatus,
	})
	if err != nil {
		return nil, err
	}

	return string(bs), err
}
