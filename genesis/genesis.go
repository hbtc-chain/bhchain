package genesis

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/token"
	tmtypes "github.com/tendermint/tendermint/types"

	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenesisState represents chain state at the start of the chain. Any initial state (custodianunit balances) are stored here.
type GenesisState struct {
	CustodianUnitData custodianunit.GenesisState `json:"custodian_units"`
	TokenData         token.GenesisState         `json:"token"`
	GenTxs            []json.RawMessage          `json:"gen_txs"`
}

// NewGenesisState generates state according inputs
func NewGenesisState(cuData custodianunit.GenesisState,
	tokenData token.GenesisState) GenesisState {
	return GenesisState{
		CustodianUnitData: cuData,
		TokenData:         tokenData,
	}
}

// NewDefaultGenesisState generates the default state for bhclear.
func NewDefaultGenesisState() GenesisState {
	return GenesisState{
		CustodianUnitData: custodianunit.DefaultGenesisState(),
		TokenData:         token.DefaultGenesisState(),
		GenTxs:            nil,
	}

}

//Create the finial Genesis file after including genTxs
func BhclearAppGenState(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	genesisState GenesisState, err error) {

	if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		return genesisState, err
	}

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		return genesisState, errors.New("there must be at least one genesis tx")
	}

	//check genTxs, each genTx should only contains one MsgCreateValidator
	for i, genTx := range appGenTxs {
		var tx custodianunit.StdTx
		if err := cdc.UnmarshalJSON(genTx, &tx); err != nil {
			return genesisState, err
		}

		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return genesisState, errors.New(
				"must provide genesis StdTx with exactly 1 CreateValidator message")
		}

		if _, ok := msgs[0].(staking.MsgCreateValidator); !ok {
			return genesisState, fmt.Errorf(
				"Genesis transaction %v does not contain a MsgCreateValidator", i)
		}
	}

	//setup GenTxs
	genesisState.GenTxs = appGenTxs

	return genesisState, nil
}

// BhclearValidateGenesisState ensures that the genesis state obeys the expected invariants
func BhclearValidateGenesisState(genesisState GenesisState) error {
	// skip stakingData validation as genesis is created from txs
	if err := custodianunit.ValidateGenesis(genesisState.CustodianUnitData); err != nil {
		return err
	}

	return token.ValidateGenesis(genesisState.TokenData)
}

// Marshal BhclearAppGenState reuslt
func BhclearAppGenStateJSON(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	appState json.RawMessage, err error) {
	// create the final app state
	genesisState, err := BhclearAppGenState(cdc, genDoc, appGenTxs)
	if err != nil {
		return nil, err
	}
	return codec.MarshalJSONIndent(cdc, genesisState)
}

// CollectStdTxs processes and validates application's genesis StdTxs and returns
// the list of appGenTxs, and persistent peers required to generate genesis.json.
func CollectStdTxs(cdc *codec.Codec, moniker string, genTxsDir string, genDoc tmtypes.GenesisDoc) (
	appGenTxs []custodianunit.StdTx, persistentPeers string, err error) {

	var fos []os.FileInfo
	fos, err = ioutil.ReadDir(genTxsDir)
	if err != nil {
		return appGenTxs, persistentPeers, err
	}

	// prepare a map of all CustodianUnits in genesis state to then validate
	// against the validators addresses
	var appState GenesisState
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		return appGenTxs, persistentPeers, err
	}

	addrMap := make(map[string]exported.CustodianUnit, len(appState.CustodianUnitData.Cus))
	for i := 0; i < len(appState.CustodianUnitData.Cus); i++ {
		cu := appState.CustodianUnitData.Cus[i]
		addrMap[cu.GetAddress().String()] = cu
		//	fmt.Printf("addrMap[%v]= %v\n", cu.GetCUAddress().String(), cu)
	}

	// addresses and IPs (and port) validator server info
	var addressesIPs []string

	for _, fo := range fos {
		filename := filepath.Join(genTxsDir, fo.Name())
		if !fo.IsDir() && (filepath.Ext(filename) != ".json") {
			continue
		}

		// get the genStdTx
		var jsonRawTx []byte
		if jsonRawTx, err = ioutil.ReadFile(filename); err != nil {
			return appGenTxs, persistentPeers, err
		}
		var genStdTx custodianunit.StdTx
		if err = cdc.UnmarshalJSON(jsonRawTx, &genStdTx); err != nil {
			return appGenTxs, persistentPeers, err
		}
		appGenTxs = append(appGenTxs, genStdTx)

		// the memo flag is used to store
		// the ip and node-id, for example this may be:
		// "528fd3df22b31f4969b05652bfe8f0fe921321d5@192.168.2.37:26656"
		nodeAddrIP := genStdTx.GetMemo()
		if len(nodeAddrIP) == 0 {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"couldn't find node's address and IP in %s", fo.Name())
		}

		// genesis transactions must be single-message
		msgs := genStdTx.GetMsgs()
		if len(msgs) != 1 {

			return appGenTxs, persistentPeers, errors.New(
				"each genesis transaction must provide a single genesis message")
		}

		msg, ok := msgs[0].(staking.MsgCreateValidator)
		if !ok {
			return appGenTxs, persistentPeers, errors.New(
				"not a create-validator msg")
		}
		// validate delegator and validator addresses and funds against the CustodianUnits in the state
		//delAddr := msg.DelegatorAddress.String()
		//valAddr := sdk.AccAddress(msg.ValidatorAddress).String()
		delAddr := sdk.CosmosAddressToCUAddress(msg.DelegatorAddress).String()
		valAddr := sdk.CosmosAddressToCUAddress(msg.ValidatorAddress).String()

		//fmt.Printf("delAddr:%v, valAddr:%v\n", delAddr, valAddr)

		_, delOk := addrMap[delAddr]
		_, valOk := addrMap[valAddr]

		cusNotInGenesis := []string{}
		if !delOk {
			cusNotInGenesis = append(cusNotInGenesis, delAddr)
		}
		if !valOk {
			cusNotInGenesis = append(cusNotInGenesis, valAddr)
		}
		if len(cusNotInGenesis) != 0 {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"custodianunit(s) %v not in genesis.json: %+v", strings.Join(cusNotInGenesis, " "), addrMap)
		}

		// exclude itself from persistent peers
		if msg.Description.Moniker != moniker {
			addressesIPs = append(addressesIPs, nodeAddrIP)
		}
	}

	sort.Strings(addressesIPs)
	persistentPeers = strings.Join(addressesIPs, ",")

	return appGenTxs, persistentPeers, nil
}
