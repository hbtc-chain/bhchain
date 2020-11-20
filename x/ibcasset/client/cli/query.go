package cli

import (
	"fmt"
	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/client/flags"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the ibcasset unit module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(GetCUAssetCmd(cdc))

	return cmd
}

// GetCUCmd returns a query CU that will display the state of the
// CU at a given address.
func GetCUAssetCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cuasset [address]",
		Aliases: []string{"ci"},
		Short:   "Query cu asset info",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			addr, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.NewQueryCUAssetParams(addr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCUAsset)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return flags.GetCommands(cmd)[0]
}
