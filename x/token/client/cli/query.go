package cli

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/token/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	tokenQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the token module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		//	RunE:                       utils.ValidateCmd,
	}
	tokenQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryToken(cdc),
		GetCmdQueryIBCTokens(cdc),
	)...)
	return tokenQueryCmd
}

func GetCmdQueryToken(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "token [symbol]",
		Short: "token symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			symbol := args[0]

			bz, err := cdc.MarshalJSON(types.NewQueryTokenInfoParams(symbol))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryToken)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryIBCTokens(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ibc-tokens",
		Short: "ibc-tokens",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryIBCTokens)
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
}
