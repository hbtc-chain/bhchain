package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/hbtc-chain/bhchain/client/keys"
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/server"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/genaccounts"
	"github.com/hbtc-chain/bhchain/x/genutil"
)

const (
	flagClientHome   = "home-client"
	flagVestingStart = "vesting-start-time"
	flagVestingEnd   = "vesting-end-time"
	flagVestingAmt   = "vesting-amount"
)

// AddGenesisAccountCmd returns add-genesis-CU cobra Command.
func AddGenesisAccountCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-CU [address_or_key_name] [coin][,[coin]]",
		Short: "Add genesis CU to genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			addr, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				kb, err := keys.NewKeyBaseFromDir(viper.GetString(flagClientHome))
				if err != nil {
					return err
				}

				info, err := kb.Get(args[0])
				if err != nil {
					return err
				}

				addr = info.GetAddress()
			}

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			genAcc := genaccounts.NewGenesisCURaw(sdk.CUTypeUser, nil, nil, addr, coins, sdk.NewCoins(),
				sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), []sdk.Asset{}, "", "")

			if err := genAcc.Validate(); err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// add genesis CU to the app state
			var genesisAccounts genaccounts.GenesisCUs

			cdc.MustUnmarshalJSON(appState[genaccounts.ModuleName], &genesisAccounts)

			if genesisAccounts.Contains(addr) {
				return fmt.Errorf("cannot add CU at existing address %v", addr)
			}

			genesisAccounts = append(genesisAccounts, genAcc)

			genesisStateBz := cdc.MustMarshalJSON(genaccounts.GenesisState(genesisAccounts))
			appState[genaccounts.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			// export app state
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")
	cmd.Flags().String(flagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Uint64(flagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Uint64(flagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	return cmd
}
