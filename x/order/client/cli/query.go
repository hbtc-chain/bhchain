package cli

import (
	"fmt"
	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/version"
	"github.com/hbtc-chain/bhchain/x/order/types"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"strings"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	QueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the order module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	QueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryOrder(cdc),
	)...)
	return QueryCmd

}

// GetCmdQueryOrder implements the validator query command.
func GetCmdQueryOrder(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "order [orderID]",
		Short: "Query a order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual order.

Example:
$ %s query order order	123e4567-e89b-12d3-a456-426655440000`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			orderid, err := uuid.FromString(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.QueryOrderParams{OrderID: orderid.String()})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryOrder)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var order sdk.Order
			if err = cdc.UnmarshalJSON(res, &order); err != nil {
				return err
			}

			return cliCtx.PrintOutput(order)
		},
	}
}
