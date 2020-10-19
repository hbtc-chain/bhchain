package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/hbtc-chain/bhchain/types"
	authtypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
)

var _ exported.ModuleAccountI = (*ModuleAccount)(nil)

// ModuleAccount defines an CustodianUnit for modules that holds coins on a pool
type ModuleAccount struct {
	*authtypes.BaseCU
	Name        string   `json:"name" yaml:"name"`               // name of the module
	Permissions []string `json:"permissions" yaml:"permissions"` // permissions of module CustodianUnit
}

// NewModuleAddress creates an CUAddress from the hash of the module's name
func NewModuleAddress(name string) sdk.CUAddress {
	return sdk.CUAddress(crypto.AddressHash([]byte(name)))
}

func NewEmptyModuleAccount(name string, permissions ...string) *ModuleAccount {
	moduleAddress := NewModuleAddress(name)
	baseAcc := authtypes.NewBaseCUWithAddress(moduleAddress, sdk.CUTypeUser)

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseCU:      &baseAcc,
		Name:        name,
		Permissions: permissions,
	}
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(ba *authtypes.BaseCU,
	name string, permissions ...string) *ModuleAccount {

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseCU:      ba,
		Name:        name,
		Permissions: permissions,
	}
}

// HasPermission returns whether or not the module CustodianUnit has permission.
func (ma ModuleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// GetPermissions returns permissions granted to the module CustodianUnit
func (ma ModuleAccount) GetPermissions() []string {
	return ma.Permissions
}

// SetPubKey - Implements CustodianUnit
func (ma ModuleAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for module accounts")
}

// SetSequence - Implements CustodianUnit
func (ma ModuleAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for module accounts")
}

// String follows stringer interface
func (ma ModuleAccount) String() string {
	b, err := yaml.Marshal(ma)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// MarshalYAML returns the YAML representation of a ModuleAccount.
func (ma ModuleAccount) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(struct {
		Address     sdk.CUAddress
		Coins       sdk.Coins
		PubKey      string
		Sequence    uint64
		Name        string
		Permissions []string
	}{
		Address:     ma.Address,
		Coins:       ma.Coins,
		PubKey:      "",
		Sequence:    ma.Sequence,
		Name:        ma.Name,
		Permissions: ma.Permissions,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}
