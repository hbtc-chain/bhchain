package server

// DONTCOVER

import (
	"fmt"
	"net"
	"os"
	"runtime/pprof"
	"time"

	"github.com/hbtc-chain/bhchain/services/p2p/pb"
	bhreactor "github.com/hbtc-chain/bhchain/services/p2p/reactor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/abci/server"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"google.golang.org/grpc"
)

// Tendermint full-node start flags
const (
	flagWithTendermint     = "with-tendermint"
	flagAddress            = "address"
	flagTraceStore         = "trace-store"
	flagPruning            = "pruning"
	flagCPUProfile         = "cpu-profile"
	FlagMinGasPrices       = "minimum-gas-prices"
	flagP2PServer          = "p2p-server"
	FlagHaltHeight         = "halt-height"
	FlagHaltTime           = "halt-time"
	FlagChainnodeNetwork   = "chainnode-network"
	FlagUnsafeSkipUpgrades = "unsafe-skip-upgrades"
)

// StartCmd runs the service passed in, either stand-alone or in-process with
// Tendermint.
func StartCmd(ctx *Context, appCreator AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with Tendermint in or out of process. By
default, the application will run with Tendermint in process.

Pruning options can be provided via the '--pruning' flag. The options are as follows:

syncable: only those states not needed for state syncing will be deleted (keeps last 100 + every 10000th)
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current state

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !viper.GetBool(flagWithTendermint) {
				ctx.Logger.Info("starting ABCI without Tendermint")
				return startStandAlone(ctx, appCreator)
			}

			ctx.Logger.Info("starting ABCI with Tendermint")

			_, err := startInProcess(ctx, appCreator)
			return err
		},
	}

	// core flags for the ABCI application
	cmd.Flags().Bool(flagWithTendermint, true, "Run abci app embedded in-process with tendermint")
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(flagPruning, "syncable", "Pruning strategy: syncable, nothing, everything")
	cmd.Flags().String(
		FlagMinGasPrices, "",
		"Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)",
	)
	cmd.Flags().String(
		FlagChainnodeNetwork, "testnet",
		"Network type of chainnode; available value: mainnet, testnet, regtest",
	)
	cmd.Flags().String(flagP2PServer, ":26659", "P2P server address for settle")
	cmd.Flags().Uint64(FlagHaltHeight, 0, "Height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().IntSlice(FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")

	// add support for all Tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func startStandAlone(ctx *Context, appCreator AppCreator) error {
	addr := viper.GetString(flagAddress)
	home := viper.GetString("home")
	traceWriterFile := viper.GetString(flagTraceStore)

	db, err := openDB(home)
	if err != nil {
		return err
	}
	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		return err
	}

	app := appCreator(ctx.Logger, db, traceWriter)

	svr, err := server.NewServer(addr, "socket", app)
	if err != nil {
		return fmt.Errorf("error creating listener: %v", err)
	}

	svr.SetLogger(ctx.Logger.With("module", "abci-server"))

	err = svr.Start()
	if err != nil {
		cmn.Exit(err.Error())
	}

	cmn.TrapSignal(ctx.Logger, func() {
		// cleanup
		err = svr.Stop()
		if err != nil {
			cmn.Exit(err.Error())
		}
	})

	// run forever (the node will not be returned)
	select {}
}

func startInProcess(ctx *Context, appCreator AppCreator) (*node.Node, error) {
	cfg := ctx.Config
	home := cfg.RootDir
	traceWriterFile := viper.GetString(flagTraceStore)

	db, err := openDB(home)
	if err != nil {
		return nil, err
	}
	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		return nil, err
	}

	app := appCreator(ctx.Logger, db, traceWriter)

	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	UpgradeOldPrivValFile(cfg)

	// TODO(kai.wen): panic if it is not started in process
	// TODO(kai.wen): stop p2p server if possible
	p2pServerAddress := viper.GetString(flagP2PServer)
	p2pReactor := bhreactor.NewBHReactor(ctx.Logger.With("module", "bhreactor"))
	p2pLis, err := net.Listen("tcp", p2pServerAddress)
	if err != nil {
		return nil, err
	}
	p2pServer := grpc.NewServer()
	pb.RegisterP2PServiceServer(p2pServer, p2pReactor)
	go func() {
		ctx.Logger.Info("Start p2p server", "address", p2pLis.Addr())
		p2pServer.Serve(p2pLis)
	}()

	customChannels := []byte{bhreactor.BHMsgChannel}
	// create & start tendermint node
	tmNode, err := node.NewNode(
		cfg,
		pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation),
		ctx.Logger.With("module", "node"),
		customChannels,
		node.CustomReactors(map[string]p2p.Reactor{"bhreactor": p2pReactor}),
	)
	if err != nil {
		return nil, err
	}

	if err := tmNode.Start(); err != nil {
		return nil, err
	}

	var cpuProfileCleanup func()

	if cpuProfile := viper.GetString(flagCPUProfile); cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return nil, err
		}

		ctx.Logger.Info("starting CPU profiler", "profile", cpuProfile)
		if err := pprof.StartCPUProfile(f); err != nil {
			return nil, err
		}

		cpuProfileCleanup = func() {
			ctx.Logger.Info("stopping CPU profiler", "profile", cpuProfile)
			pprof.StopCPUProfile()
			f.Close()
		}
	}

	TrapSignal(func() {
		if tmNode.IsRunning() {
			_ = tmNode.Stop()
		}

		if cpuProfileCleanup != nil {
			cpuProfileCleanup()
		}

		ctx.Logger.Info("exiting..., wait 5 seconds")
		time.Sleep(5 * time.Second)
		ctx.Logger.Info("exited")
	})

	// run forever (the node will not be returned)
	select {}
}
