package main

import (
	"encoding/json"
	"fmt"
	"github.com/hbtc-chain/bhchain/chainnode"
	"github.com/hbtc-chain/bhchain/chainnode/grpcclient"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/baseapp"
	"github.com/hbtc-chain/bhchain/bhexapp"
	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/flags"
	"github.com/hbtc-chain/bhchain/server"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/genaccounts"
	genaccscli "github.com/hbtc-chain/bhchain/x/genaccounts/client/cli"
	genutilcli "github.com/hbtc-chain/bhchain/x/genutil/client/cli"
	"github.com/hbtc-chain/bhchain/x/staking"
)

// bhchain custom flags
const (
	flagInvCheckPeriod = "inv-check-period"
)

var invCheckPeriod uint

func main() {
	cdc := bhexapp.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "hbtcd",
		Short:             "hbtcd Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, bhexapp.ModuleBasics, bhexapp.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, bhexapp.DefaultNodeHome))
	//rootCmd.AddCommand(genutilcli.MigrateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(genutilcli.GenTxCmd(ctx, cdc, bhexapp.ModuleBasics, staking.AppModuleBasic{},
		genaccounts.AppModuleBasic{}, bhexapp.DefaultNodeHome, bhexapp.DefaultCLIHome))
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, bhexapp.ModuleBasics))
	rootCmd.AddCommand(genaccscli.AddGenesisAccountCmd(ctx, cdc, bhexapp.DefaultNodeHome, bhexapp.DefaultCLIHome))
	rootCmd.AddCommand(client.NewCompletionCmd(rootCmd, true))
	rootCmd.AddCommand(testnetCmd(ctx, cdc, bhexapp.ModuleBasics, genaccounts.AppModuleBasic{}))
	rootCmd.AddCommand(replayCmd())

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "BH", bhexapp.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range viper.GetIntSlice(server.FlagUnsafeSkipUpgrades) {
		skipUpgradeHeights[int64(h)] = true
	}
	return bhexapp.Newbhexapp(
		logger, db, traceStore, true, invCheckPeriod, getChainnode(logger, viper.GetString(server.FlagChainnodeNetwork)), skipUpgradeHeights, viper.GetString(flags.FlagHome),
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(uint64(viper.GetInt(server.FlagHaltHeight))),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		gApp := bhexapp.Newbhexapp(logger, db, traceStore, false, uint(1), getChainnode(logger, viper.GetString(server.FlagChainnodeNetwork)), map[int64]bool{}, "")
		err := gApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return gApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	gApp := bhexapp.Newbhexapp(logger, db, traceStore, true, uint(1), getChainnode(logger, viper.GetString(server.FlagChainnodeNetwork)), map[int64]bool{}, "")
	return gApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

func getChainnode(logger log.Logger, network string) chainnode.Chainnode {
	cn := grpcclient.New(logger)
	logger.Info("start init local chainnode", "networktype", network)
	if err := cn.Init(network); err != nil {
		panic(fmt.Sprintf("Failed to init local chainnode err: %v", err))
	}
	return cn
}
