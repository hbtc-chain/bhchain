package bhexapp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/codec"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestbhexappExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := Newbhexapp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0, nil, nil, "")

	genesisState := NewDefaultGenesisState()
	stateBytes, err := codec.MarshalJSONIndent(app.cdc, genesisState)
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := Newbhexapp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0, nil, nil, "")
	_, _, err = app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// ensure that black listed addresses are properly set in bank keeper
func TestBlackListedAddrs(t *testing.T) {
	db := dbm.NewMemDB()
	app := Newbhexapp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0, nil, nil, "")

	for acc := range maccPerms {
		require.True(t, app.transferKeeper.BlacklistedAddr(app.supplyKeeper.GetModuleAddress(acc)))
	}
}
