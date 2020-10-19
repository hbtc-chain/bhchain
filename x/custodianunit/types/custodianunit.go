package types

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"

	"github.com/hbtc-chain/bhchain/base58"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

var _ exported.CustodianUnit = (*BaseCU)(nil)

type BaseCU struct {
	Type             sdk.CUType          `json:"cu_type" yaml:"type"`
	Address          sdk.CUAddress       `json:"address" yaml:"address"`
	PubKey           crypto.PubKey       `json:"public_key" yaml:"public_key"`
	Sequence         uint64              `json:"sequence" yaml:"sequence"`
	Coins            sdk.Coins           `json:"coins" yaml:"coins"`
	CoinsHold        sdk.Coins           `json:"coins_hold" yaml:"coins_hold"`
	Assets           []sdk.Asset         `json:"assets" yaml:"assets"`
	AssetCoins       sdk.Coins           `json:"asset_coins" yaml:"asset_coins"`
	AssetCoinsHold   sdk.Coins           `json:"asset_coins_hold" yaml:"asset_coins_hold"`
	AssetPubkey      []byte              `json:"asset_pubkey" yaml:"asset_pubkey"`
	AssetPubkeyEpoch uint64              `json:"asset_pubkey_epoch" yaml:"asset_pubkey_epoch"`
	GasUsed          sdk.Coins           `json:"gas_used" yaml:"gas_used"`
	GasReceived      sdk.Coins           `json:"gas_received" yaml:"gas_received"`
	MigrationStatus  sdk.MigrationStatus `json:"migration_status" yaml:"migration_status"`

	balanceFlows []sdk.BalanceFlow
}

// NewBaseCU creates a new BaseCU object
func NewBaseCU(cutype sdk.CUType, address sdk.CUAddress, coins, assetCoins, assetCoinsHold, gasUsed, gasReceived, coinsHold sdk.Coins,
	pubKey crypto.PubKey, sequence uint64, assets []sdk.Asset, assetPubkey []byte) *BaseCU {
	return &BaseCU{
		Type:            cutype,
		Address:         address,
		Coins:           coins,
		CoinsHold:       coinsHold,
		PubKey:          pubKey,
		Sequence:        sequence,
		AssetCoins:      assetCoins,
		AssetCoinsHold:  assetCoinsHold,
		GasUsed:         gasUsed,
		GasReceived:     gasReceived,
		Assets:          assets,
		AssetPubkey:     assetPubkey,
		MigrationStatus: sdk.MigrationFinish,
	}
}

func ProtoBaseCU() exported.CustodianUnit {
	return &BaseCU{}
}

//Keeper independent constructor, only for genesis
func NewBaseCUWithAddress(cuaddr sdk.CUAddress, cuType sdk.CUType) BaseCU {
	return BaseCU{
		Type:            cuType,
		Address:         cuaddr,
		MigrationStatus: sdk.MigrationFinish,
	}
}

func NewBaseCUWithPubkey(pub crypto.PubKey, cuType sdk.CUType) BaseCU {
	cu := NewBaseCUWithAddress(sdk.CUAddress(pub.Address()), cuType)
	cu.SetPubKey(pub)
	return cu
}

func (bcu *BaseCU) GetAddress() sdk.CUAddress {
	return bcu.Address
}

func (bcu *BaseCU) SetAddress(cuaddress sdk.CUAddress) error {
	if len(bcu.Address) != 0 {
		return errors.New("cannot override custodian unit address")
	}

	bcu.Address = cuaddress
	return nil
}

func (bcu *BaseCU) GetCUType() sdk.CUType {
	return bcu.Type
}

// SetCUType set the custodian unit type
// if the type of custodian unit already defined return error
func (bcu *BaseCU) SetCUType(cuType sdk.CUType) error {
	if bcu.Type != 0 {
		return errors.New("cannot override custodian unit type")
	}
	bcu.Type = cuType
	return nil
}

func (bcu *BaseCU) GetPubKey() crypto.PubKey {
	return bcu.PubKey
}

func (bcu *BaseCU) SetPubKey(pub crypto.PubKey) error {
	if bcu.PubKey != nil {
		if !bcu.PubKey.Equals(pub) {
			return errors.New("Setting a different public key")
		}
	} else {
		if pub == nil {
			return errors.New("Setting a nil public key")
		}
		bcu.PubKey = pub
	}
	return nil
}

// GetSymbol for operation CU . return first asset.symbol
func (bcu *BaseCU) GetSymbol() string {
	if bcu.Type == sdk.CUTypeOp && len(bcu.Assets) > 0 {
		return bcu.Assets[0].Denom
	}
	return ""
}

// SetSymbol for operation CU . set first asset.symbol
func (bcu *BaseCU) SetSymbol(symbol string, epoch uint64) error {
	if bcu.Type != sdk.CUTypeOp {
		return errors.New("only opreation custodianunit can set symbol")
	}
	if len(bcu.Assets) == 0 {
		bcu.AddAsset(symbol, "", epoch)
		return nil
	}
	if bcu.Assets[0].Denom == "" {
		bcu.Assets[0].Denom = symbol
	}
	return errors.New("custodianunit asset not empty")
}

func (bcu *BaseCU) GetSequence() uint64 {
	return bcu.Sequence
}

func (bcu *BaseCU) SetSequence(seq uint64) error {
	bcu.Sequence = seq
	return nil
}

func (bcu *BaseCU) IsEnabledSendTx(chain string, addr string) bool {
	for _, ast := range bcu.Assets {
		if ast.Denom == chain && ast.Address == addr {
			return ast.EnableSendTx
		}
	}
	return false
}

func (bcu *BaseCU) SetEnableSendTx(enabled bool, chain string, addr string) {
	for i, ast := range bcu.Assets {
		if ast.Denom == chain && ast.Address == addr {
			bcu.Assets[i].EnableSendTx = enabled
			break
		}
	}
}

func (bcu *BaseCU) GetCoins() sdk.Coins {
	return bcu.Coins
}

func (bcu *BaseCU) SetCoins(coins sdk.Coins) error {
	changedCoins, _ := coins.SafeSub(bcu.Coins)

	for _, changedCoin := range changedCoins {
		denom := changedCoin.Denom
		changeAmt := changedCoin.Amount
		bcu.addBalanceFlow(
			denom,
			bcu.Coins.AmountOf(denom),
			changeAmt,
			bcu.CoinsHold.AmountOf(denom),
			sdk.ZeroInt())
	}
	bcu.Coins = coins
	return nil
}

func (bcu *BaseCU) GetCoinsHold() sdk.Coins {
	return bcu.CoinsHold
}

func (bcu *BaseCU) SetCoinsHold(coins sdk.Coins) error {
	changedCoinsHold, _ := coins.SafeSub(bcu.CoinsHold)

	for _, changedCoin := range changedCoinsHold {
		denom := changedCoin.Denom
		changeAmt := changedCoin.Amount
		bcu.addBalanceFlow(
			denom,
			bcu.Coins.AmountOf(denom),
			sdk.ZeroInt(),
			bcu.CoinsHold.AmountOf(denom),
			changeAmt)
	}
	bcu.CoinsHold = coins
	return nil
}

func (bcu *BaseCU) AddCoins(coins sdk.Coins) sdk.Coins {
	for _, coin := range coins {
		denom := coin.Denom
		bcu.addBalanceFlow(coin.Denom, bcu.Coins.AmountOf(denom), coin.Amount,
			bcu.CoinsHold.AmountOf(denom), sdk.ZeroInt())
	}
	bcu.Coins = bcu.Coins.Add(coins)
	return bcu.Coins
}

func (bcu *BaseCU) SubCoins(coins sdk.Coins) sdk.Coins {
	for _, coin := range coins {
		denom := coin.Denom
		bcu.addBalanceFlow(coin.Denom, bcu.Coins.AmountOf(denom), coins.AmountOf(denom).Neg(),
			bcu.CoinsHold.AmountOf(denom), sdk.ZeroInt())
	}
	bcu.Coins = bcu.Coins.Sub(coins)
	return bcu.Coins
}

func (bcu *BaseCU) AddCoinsHold(coins sdk.Coins) sdk.Coins {
	for _, coin := range coins {
		denom := coin.Denom
		bcu.addBalanceFlow(coin.Denom, bcu.Coins.AmountOf(denom), sdk.ZeroInt(),
			bcu.CoinsHold.AmountOf(denom), coin.Amount)
	}
	bcu.CoinsHold = bcu.CoinsHold.Add(coins)
	return bcu.CoinsHold
}

func (bcu *BaseCU) SubCoinsHold(coins sdk.Coins) sdk.Coins {
	for _, coin := range coins {
		denom := coin.Denom
		bcu.addBalanceFlow(coin.Denom, bcu.Coins.AmountOf(denom), sdk.ZeroInt(),
			bcu.CoinsHold.AmountOf(denom), coin.Amount.Neg())
	}
	bcu.CoinsHold = bcu.CoinsHold.Sub(coins)
	return bcu.CoinsHold
}

func (bcu *BaseCU) GetAssets() []sdk.Asset {
	return bcu.Assets
}

func (bcu *BaseCU) GetAsset(denom string, epoch uint64) sdk.Asset {
	for _, asset := range bcu.Assets {
		if asset.Denom == denom && asset.Epoch == epoch {
			return asset
		}
	}
	return sdk.NilAsset
}

func (bcu *BaseCU) GetAssetByAddr(denom string, addr string) sdk.Asset {
	for _, asset := range bcu.Assets {
		if asset.Denom == denom && asset.Address == addr {
			return asset
		}
	}
	return sdk.NilAsset
}

func (bcu *BaseCU) GetAssetAddress(denom string, epoch uint64) string {
	as := bcu.GetAsset(denom, epoch)
	if as == sdk.NilAsset {
		return ""
	}
	return as.Address
}

func (bcu *BaseCU) GetAssetPubkey(epoch uint64) []byte {
	if bcu.AssetPubkeyEpoch != epoch {
		return nil
	}
	return bcu.AssetPubkey
}

func (bcu *BaseCU) GetAssetPubkeyEpoch() uint64 {
	return bcu.AssetPubkeyEpoch
}

func (bcu *BaseCU) SetAssetPubkey(pubkey []byte, epoch uint64) error {
	bcu.AssetPubkey = pubkey
	bcu.AssetPubkeyEpoch = epoch
	return nil
}

func (bcu *BaseCU) AddAsset(denom, address string, epoch uint64) error {
	for i := 0; i < len(bcu.Assets); i++ {
		asset := bcu.Assets[i]
		if asset.Denom == denom && asset.Epoch == epoch {
			return errors.New("asset already exist")
		}

		if asset.Denom == denom && asset.Epoch == 0 {
			bcu.Assets[i].Address = address
			bcu.Assets[i].Epoch = epoch
		}
	}

	//delete old epoch's assert info
	for i := 0; i < len(bcu.Assets); i++ {
		asset := bcu.Assets[i]
		if epoch >= 3 && epoch-asset.Epoch >= 2 {
			if !asset.GasRemained.IsZero() {
				bcu.AddGasUsed(sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.GasRemained)))
			}
			bcu.Assets = append(bcu.Assets[:i], bcu.Assets[i+1:]...)
			i--
		}
	}

	bcu.Assets = append(bcu.Assets, sdk.NewAsset(denom, address, epoch, true))

	return nil
}

func (bcu *BaseCU) SetAssetAddress(denom, address string, epoch uint64) error {
	for i, asset := range bcu.Assets {
		if asset.Denom == denom && asset.Epoch == 0 {
			bcu.Assets[i].Address = address
			bcu.Assets[i].Epoch = epoch
			return nil
		}
	}
	return bcu.AddAsset(denom, address, epoch)
}

func (bcu *BaseCU) GetNonce(denom string, addr string) uint64 {
	for _, asset := range bcu.Assets {
		if asset.Denom == denom && asset.Address == addr {
			return asset.Nonce
		}
	}

	return 0
}

func (bcu *BaseCU) SetNonce(denom string, nonce uint64, addr string) {
	for i, asset := range bcu.Assets {
		if asset.Denom == denom && asset.Address == addr {
			bcu.Assets[i].Nonce = nonce
			return
		}
	}
}

func (bcu *BaseCU) GetAssetCoinsHold() sdk.Coins {
	return bcu.AssetCoinsHold
}
func (bcu *BaseCU) AddAssetCoinsHold(coins sdk.Coins) sdk.Coins {
	bcu.AssetCoinsHold = bcu.AssetCoinsHold.Add(coins)
	return bcu.AssetCoinsHold
}
func (bcu *BaseCU) SubAssetCoinsHold(coins sdk.Coins) sdk.Coins {
	bcu.AssetCoinsHold = bcu.AssetCoinsHold.Sub(coins)
	return bcu.AssetCoinsHold
}

func (bcu *BaseCU) GetAssetCoins() sdk.Coins {
	return bcu.AssetCoins
}

func (bcu *BaseCU) AddAssetCoins(coins sdk.Coins) sdk.Coins {
	bcu.AssetCoins = bcu.AssetCoins.Add(coins)
	return bcu.AssetCoins
}

func (bcu *BaseCU) SubAssetCoins(coins sdk.Coins) sdk.Coins {
	bcu.AssetCoins = bcu.AssetCoins.Sub(coins)
	return bcu.AssetCoins
}

func (bcu *BaseCU) GetGasRemained(chain string, addr string) sdk.Int {
	asset := bcu.GetAssetByAddr(chain, addr)
	if asset == sdk.NilAsset {
		return sdk.ZeroInt()
	}
	return asset.GasRemained
}

func (bcu *BaseCU) AddGasRemained(chain string, addr string, amt sdk.Int) {
	for i, asset := range bcu.Assets {
		if asset.Denom == chain && asset.Address == addr {
			bcu.Assets[i].GasRemained = asset.GasRemained.Add(amt)
			return
		}
	}
}

func (bcu *BaseCU) SubGasRemained(chain string, addr string, amt sdk.Int) {
	for i, asset := range bcu.Assets {
		if asset.Denom == chain && asset.Address == addr {
			bcu.Assets[i].GasRemained = asset.GasRemained.Sub(amt)
			return
		}
	}
}

func (bcu *BaseCU) GetGasUsed() sdk.Coins {
	return bcu.GasUsed
}

func (bcu *BaseCU) AddGasUsed(coins sdk.Coins) sdk.Coins {
	bcu.GasUsed = bcu.GasUsed.Add(coins)
	return bcu.GasUsed
}

func (bcu *BaseCU) SubGasUsed(coins sdk.Coins) sdk.Coins {
	bcu.GasUsed = bcu.GasUsed.Sub(coins)
	return bcu.GasUsed
}

func (bcu *BaseCU) GetGasReceived() sdk.Coins {
	return bcu.GasReceived
}

func (bcu *BaseCU) AddGasReceived(coins sdk.Coins) sdk.Coins {
	bcu.GasReceived = bcu.GasReceived.Add(coins)
	return bcu.GasReceived
}

func (bcu *BaseCU) SubGasReceived(coins sdk.Coins) sdk.Coins {
	bcu.GasReceived = bcu.GasReceived.Sub(coins)
	return bcu.GasReceived
}

func (bcu *BaseCU) SetMigrationStatus(status sdk.MigrationStatus) {
	bcu.MigrationStatus = status
}

func (bcu *BaseCU) GetMigrationStatus() sdk.MigrationStatus {
	return bcu.MigrationStatus
}

func (bcu *BaseCU) addBalanceFlow(symbol string, previousBalance, balanceChange, previousBalanceOnHold, balanceOnHoldChange sdk.Int) []sdk.BalanceFlow {
	if balanceChange.IsZero() && balanceOnHoldChange.IsZero() {
		return bcu.balanceFlows
	}
	// merge balanceflows with same symbol
	for i, bFlow := range bcu.balanceFlows {
		if bFlow.Symbol.String() == symbol {
			bFlow.BalanceChange = bFlow.BalanceChange.Add(balanceChange)
			bFlow.BalanceOnHoldChange = bFlow.BalanceOnHoldChange.Add(balanceOnHoldChange)
			bcu.balanceFlows[i] = bFlow
			if bFlow.BalanceChange.IsZero() && bFlow.BalanceOnHoldChange.IsZero() {
				// remove zero flow
				bcu.balanceFlows = append((bcu.balanceFlows)[:i], (bcu.balanceFlows)[i+1:]...)
			}
			return bcu.balanceFlows
		}
	}
	bcu.balanceFlows = append(bcu.balanceFlows, sdk.BalanceFlow{CUAddress: bcu.Address, Symbol: sdk.Symbol(symbol), PreviousBalance: previousBalance,
		BalanceChange: balanceChange, PreviousBalanceOnHold: previousBalanceOnHold, BalanceOnHoldChange: balanceOnHoldChange})

	return bcu.balanceFlows
}

func (bcu *BaseCU) GetBalanceFlows() []sdk.BalanceFlow {
	return bcu.balanceFlows
}

func (bcu *BaseCU) ResetBalanceFlows() {
	bcu.balanceFlows = nil
}

// String implements fmt.Stringer
func (bcu *BaseCU) String() string {
	pk := bcu.GetPubKey()
	if pk == nil {
		return fmt.Sprintf(`Account:
  Address:       %s
  Sequence:      %d`,
			(sdk.CUAddress)(bcu.GetAddress()).String(), bcu.GetSequence(),
		)
	}

	return fmt.Sprintf(`Account:
  Address:       %s
  Pubkey:        %s
  Sequence:      %d`,
		(sdk.CUAddress)(bcu.GetAddress()).String(), sdk.PubkeyToString(pk), bcu.GetSequence(),
	)
}

// MarshalYAML returns the YAML representation of an custodianunit.
func (bcu *BaseCU) MarshalYAML() (interface{}, error) {
	var bs []byte
	var err error
	var pubkey string
	var assetPubkey string
	if bcu.PubKey != nil {
		pubkey = sdk.PubkeyToString(bcu.PubKey)
	}
	if bcu.AssetPubkey != nil {
		assetPubkey = base58.Encode(bcu.AssetPubkey)
	}
	bs, err = yaml.Marshal(struct {
		Type             sdk.CUType
		Address          sdk.CUAddress
		PubKey           string
		Sequence         uint64
		Coins            sdk.Coins
		CoinsHold        sdk.Coins
		Assets           []sdk.Asset
		AssetCoins       sdk.Coins
		AssetCoinsHold   sdk.Coins
		AssetPubkey      string
		AssetPubkeyEpoch uint64
		GasUsed          sdk.Coins
		GasReceived      sdk.Coins
	}{
		Type:             bcu.Type,
		Address:          bcu.Address,
		Coins:            bcu.Coins,
		CoinsHold:        bcu.CoinsHold,
		PubKey:           pubkey,
		Sequence:         bcu.Sequence,
		AssetCoins:       bcu.AssetCoins,
		AssetCoinsHold:   bcu.AssetCoinsHold,
		GasUsed:          bcu.GasUsed,
		GasReceived:      bcu.GasReceived,
		Assets:           bcu.Assets,
		AssetPubkey:      assetPubkey,
		AssetPubkeyEpoch: bcu.AssetPubkeyEpoch,
	})
	if err != nil {
		return nil, err
	}

	return string(bs), err
}

// Validate checks for errors on the account fields
func (bcu *BaseCU) Validate() error {
	if bcu.PubKey != nil && bcu.Address != nil &&
		!bytes.Equal(bcu.PubKey.Address().Bytes(), bcu.Address.Bytes()) {
		return errors.New("pubkey and address pair is invalid")
	}

	return nil
}

type BaseCUs []BaseCU

func (bs BaseCUs) String() string {
	bsb := strings.Builder{}
	for _, b := range bs {
		bsb.WriteString(b.String())
	}
	return bsb.String()
}

type CUCoin struct {
	Address    sdk.CUAddress `json:"address" yaml:"address"`
	Denom      string        `json:"denom" yaml:"denom"`
	Amount     sdk.Int       `json:"amount" yaml:"amount"`
	AmountHold sdk.Int       `json:"amount_hold" yaml:"amount_hold"`

	ExtAddress    string  `json:"ext_address,omitempty" yaml:"ext_address"`
	ExtAmount     sdk.Int `json:"ext_amount,omitempty" yaml:"ext_amount"`
	ExtAmountHold sdk.Int `json:"ext_amount_hold,omitempty" yaml:"ext_amount_hold"`
}

// String implements fmt.Stringer
func (cc CUCoin) String() string {
	return fmt.Sprintf(`CUCoin:
  Address:       %v
  Denom:       %s
  Amount:        %v
  AmountHold:      %v
  ExtAddress:       %s
  ExtAmount:       %v
  ExtAmountHold:       %v`,
		cc.Address,
		cc.Denom,
		cc.Amount,
		cc.AmountHold,
		cc.ExtAddress,
		cc.ExtAmount,
		cc.ExtAmountHold,
	)
}
