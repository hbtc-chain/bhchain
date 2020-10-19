package types

import (
	"errors"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/tendermint/tendermint/crypto"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	cuexported "github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
)

// GenesisCU is a struct for CustodianUnit initialization used exclusively during genesis
type GenesisCU struct {
	Type           sdk.CUType    `json:"cu_type" yaml:"cu_type"`
	PubKey         crypto.PubKey `json:"public_key" yaml:"public_key"`
	CoinsHold      sdk.Coins     `json:"coins_hold" yaml:"coins_hold"`
	Assets         []sdk.Asset   `json:"assets" yaml:"assets"`
	AssetCoins     sdk.Coins     `json:"asset_coins" yaml:"asset_coins"`
	AssetCoinsHold sdk.Coins     `json:"asset_coins_hold" yaml:"asset_coins_hold"`
	DisableSendTx  []sdk.Symbol  `json:"disable_send_tx" yaml:"disable_send_tx"`
	AssetPubkey    []byte        `json:"asset_pubkey" yaml:"asset_pubkey"`
	GasUsed        sdk.Coins     `json:"gas_used" yaml:"gas_used"`
	GasReceived    sdk.Coins     `json:"gas_received" yaml:"gas_received"`
	Address        sdk.CUAddress `json:"address" yaml:"address"`
	Coins          sdk.Coins     `json:"coins" yaml:"coins"`
	Sequence       uint64        `json:"sequence_number" yaml:"sequence_number"`

	// module CustodianUnit fields
	ModuleName        string   `json:"module_name" yaml:"module_name"`               // name of the module CustodianUnit
	ModulePermissions []string `json:"module_permissions" yaml:"module_permissions"` // permissions of module CustodianUnit
}

// Validate checks for errors on the vesting and module CustodianUnit parameters
func (ga GenesisCU) Validate() error {
	// don't allow blank (i.e just whitespaces) on the module name
	if ga.ModuleName != "" && strings.TrimSpace(ga.ModuleName) == "" {
		return errors.New("module CustodianUnit name cannot be blank")
	}

	return nil
}

// NewGenesisCURaw creates a new GenesisCU object
func NewGenesisCURaw(cutype sdk.CUType, pubkey crypto.PubKey, assetpubkey []byte, address sdk.CUAddress,
	coins, coinshold, assetcoins, assetcoinshold, gasreceived, gasused sdk.Coins, assets []sdk.Asset,
	module string, permissions ...string) GenesisCU {

	return GenesisCU{
		Type:           cutype,
		PubKey:         pubkey,
		CoinsHold:      coinshold,
		Assets:         assets,
		AssetCoins:     assetcoins,
		AssetCoinsHold: assetcoinshold,
		DisableSendTx:  []sdk.Symbol{},
		AssetPubkey:    assetpubkey,
		GasUsed:        gasused,
		GasReceived:    gasreceived,
		Address:        address,
		Coins:          coins,
		Sequence:       0,

		ModuleName:        module,
		ModulePermissions: permissions,
	}
}

// NewGenesisCU creates a GenesisCU instance from a BaseCU.
func NewGenesisCU(cu *custodianunit.BaseCU) GenesisCU {
	return GenesisCU{
		Address:  cu.Address,
		Coins:    cu.Coins,
		Sequence: cu.Sequence,
	}
}

// NewGenesisCUI creates a GenesisCU instance from an CustodianUnit interface.
func NewGenesisCUI(cu cuexported.CustodianUnit) (GenesisCU, error) {
	gcu := GenesisCU{
		Address:  cu.GetAddress(),
		Coins:    cu.GetCoins(),
		Sequence: cu.GetSequence(),
	}

	if err := gcu.Validate(); err != nil {
		return gcu, err
	}

	switch acc := cu.(type) {

	case supplyexported.ModuleAccountI:
		gcu.ModuleName = acc.GetName()
		gcu.ModulePermissions = acc.GetPermissions()
	}

	return gcu, nil
}

// ToCU converts a GenesisCU to an CustodianUnit interface
func (ga *GenesisCU) ToCU() custodianunit.CU {
	bcu := custodianunit.NewBaseCU(sdk.CUTypeUser, ga.Address, ga.Coins.Sort(), ga.AssetCoins.Sort(), ga.AssetCoinsHold.Sort(), ga.GasUsed.Sort(),
		ga.GasReceived.Sort(), ga.CoinsHold.Sort(), ga.PubKey, ga.Sequence, ga.Assets, ga.AssetPubkey)

	// module accounts
	if ga.ModuleName != "" {
		return supply.NewModuleAccount(bcu, ga.ModuleName, ga.ModulePermissions...)
	}

	return bcu
}

//___________________________________
type GenesisCUs []GenesisCU

// genesis accounts contain an address
func (gcus GenesisCUs) Contains(cu sdk.CUAddress) bool {
	for _, gcu := range gcus {
		if gcu.Address.Equals(cu) {
			return true
		}
	}
	return false
}
