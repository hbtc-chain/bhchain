package types

import (
	"encoding/json"
	"fmt"
	"github.com/hbtc-chain/bhchain/codec"
	"sort"
)

// State to Unmarshal
type GenesisState GenesisCUs

// get the genesis state from the expected app state
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

// set the genesis state within the expected app state
func SetGenesisStateInAppState(cdc *codec.Codec,
	appState map[string]json.RawMessage, genesisState GenesisState) map[string]json.RawMessage {

	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz
	return appState
}

// Sanitize sorts accounts and coin sets.
// TODO sort not use CUNmuber
func (gs GenesisState) Sanitize() {
	sort.Slice(gs, func(i, j int) bool {
		return gs[i].Address.String() < gs[j].Address.String()
	})

	for _, acc := range gs {
		acc.Coins = acc.Coins.Sort()
	}
}

// ValidateGenesis performs validation of genesis accounts. It
// ensures that there are no duplicate accounts in the genesis state and any
// provided vesting accounts are valid.
func ValidateGenesis(genesisState GenesisState) error {
	addrMap := make(map[string]bool, len(genesisState))
	for _, acc := range genesisState {
		addrStr := acc.Address.String()

		// disallow any duplicate accounts
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate CU found in genesis state; address: %s", addrStr)
		}

		addrMap[addrStr] = true
	}
	return nil
}
