package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"sort"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Cus    []exported.CustodianUnit `json:"custodianunits"`
	Params Params                   `json:"params" yaml:"params"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, cus []exported.CustodianUnit) GenesisState {
	return GenesisState{Params: params, Cus: cus}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams(), []exported.CustodianUnit{})
}

func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if data.Params.TxSigLimit == 0 {
		return fmt.Errorf("invalid tx signature limit: %d", data.Params.TxSigLimit)
	}
	if data.Params.SigVerifyCostED25519 == 0 {
		return fmt.Errorf("invalid ED25519 signature verification cost: %d", data.Params.SigVerifyCostED25519)
	}
	if data.Params.SigVerifyCostSecp256k1 == 0 {
		return fmt.Errorf("invalid SECK256k1 signature verification cost: %d", data.Params.SigVerifyCostSecp256k1)
	}
	if data.Params.MaxMemoCharacters == 0 {
		return fmt.Errorf("invalid max memo characters: %d", data.Params.MaxMemoCharacters)
	}
	if data.Params.TxSizeCostPerByte == 0 {
		return fmt.Errorf("invalid tx size cost per byte: %d", data.Params.TxSizeCostPerByte)
	}
	return nil
}

func (g *GenesisState) AddCUIntoGenesisState(new exported.CustodianUnit) error {
	cus := g.Cus
	for _, cu := range cus {
		if cu.GetAddress().Equals(new.GetAddress()) {
			return fmt.Errorf("CU %v already exist", new)
		}
	}
	cus = append(cus, new)
	sort.Sort(sortCustodianUnit(cus))
	g.Cus = cus
	return nil
}

type sortCustodianUnit []exported.CustodianUnit

func (s sortCustodianUnit) Len() int      { return len(s) }
func (s sortCustodianUnit) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortCustodianUnit) Less(i, j int) bool {
	return bytes.Compare(s[i].GetAddress(), s[j].GetAddress()) == -1
}

// GenesisAccountIterator implements genesis account iteration.
type GenesisCUIterator struct{}

// IterateGenesisCUs iterates over all the genesis accounts found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisCUIterator) IterateGenesisAccounts(
	cdc *codec.Codec, appGenesis map[string]json.RawMessage, cb func(unit exported.CustodianUnit) (stop bool),
) {

	for _, genAcc := range GetGenesisStateFromAppState(cdc, appGenesis).Cus {
		if cb(genAcc) {
			break
		}
	}
}
