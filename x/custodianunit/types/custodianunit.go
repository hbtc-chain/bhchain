package types

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

var _ exported.CustodianUnit = (*BaseCU)(nil)

type BaseCU struct {
	Type     sdk.CUType    `json:"cu_type" yaml:"type"`
	Address  sdk.CUAddress `json:"address" yaml:"address"`
	PubKey   crypto.PubKey `json:"public_key" yaml:"public_key"`
	Sequence uint64        `json:"sequence" yaml:"sequence"`
	Symbol   string        `json:"symbol" yaml:"symbol"`
}

// NewBaseCU creates a new BaseCU object
func NewBaseCU(cutype sdk.CUType, address sdk.CUAddress, pubKey crypto.PubKey, sequence uint64) *BaseCU {
	return &BaseCU{
		Type:     cutype,
		Address:  address,
		PubKey:   pubKey,
		Sequence: sequence,
	}
}

func ProtoBaseCU() exported.CustodianUnit {
	return &BaseCU{}
}

//Keeper independent constructor, only for genesis
func NewBaseCUWithAddress(cuaddr sdk.CUAddress, cuType sdk.CUType) BaseCU {
	return BaseCU{
		Type:    cuType,
		Address: cuaddr,
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
	return bcu.Symbol
}

// SetSymbol for operation CU . set first asset.symbol
func (bcu *BaseCU) SetSymbol(symbol string) error {
	if bcu.Type != sdk.CUTypeOp {
		return errors.New("only opreation custodianunit can set symbol")
	}

	bcu.Symbol = symbol
	return nil
}

func (bcu *BaseCU) GetSequence() uint64 {
	return bcu.Sequence
}

func (bcu *BaseCU) SetSequence(seq uint64) error {
	bcu.Sequence = seq
	return nil
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
	if bcu.PubKey != nil {
		pubkey = sdk.PubkeyToString(bcu.PubKey)
	}

	bs, err = yaml.Marshal(struct {
		Type     sdk.CUType
		Address  sdk.CUAddress
		PubKey   string
		Sequence uint64
		Symbol   string
	}{
		Type:     bcu.Type,
		Address:  bcu.Address,
		PubKey:   pubkey,
		Sequence: bcu.Sequence,
		Symbol:   bcu.Symbol,
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
