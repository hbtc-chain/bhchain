package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

// GetQueryCmd returns the cli query commands for the minting module.
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	transferQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the hrc10 module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	transferQueryCmd.AddCommand(
		client.GetCommands(
			GetCmdQueryBalance(cdc),
			GetCmdQueryAllBalance(cdc),
		)...,
	)

	return transferQueryCmd
}

func GetCmdQueryBalance(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "balance [address] [symbol]",
		Short: "Query balance of some address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.NewQueryBalanceParams(addr, args[1]))
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryAllBalance(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "balances [address]",
		Short: "Query all balance of some address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.NewQueryAllBalanceParams(addr))
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllBalance)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
}
