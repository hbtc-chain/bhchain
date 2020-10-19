package cli

import (
	"fmt"
	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/mapping/types"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	mappingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the mapping module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		//	RunE:                       utils.ValidateCmd,
	}
	mappingQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryMapping(storeKey, cdc),
		GetCmdQueryMappingList(storeKey, cdc),
		GetFreeSwapInfo(cdc),
		GetDirectSwapInfo(cdc),
	)...)
	return mappingQueryCmd
}

func GetCmdQueryMapping(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "info [issue-symbol]",
		Short: "info issue-symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			issueSymbol := args[0]

			out, err := GetMapping(cliCtx, issueSymbol)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdQueryMappingList(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)

			params := types.NewQueryMappingListParams(1, 0) // No pagination
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryList)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var out types.QueryResMappingList
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetMapping(cliCtx client.CLIContext, issueSymbol string) (out types.QueryResMappingInfo, err error) {
	bz, err := cliCtx.Codec.MarshalJSON(types.QueryMappingParams{issueSymbol})
	if err != nil {
		return
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInfo)
	res, _, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return
	}

	cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return
}

func GetFreeSwapInfo(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeswap [orderid]",
		Short: "get a free swapinfo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			orderid, err := uuid.FromString(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.QueryFreeSwapOrderParams{OrderID: orderid.String()})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryFreeSwapInfo)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var order types.FreeSwapOrder
			if err = cdc.UnmarshalJSON(res, &order); err != nil {
				return err
			}

			return cliCtx.PrintOutput(order)
		},
	}
	return cmd
}

func GetDirectSwapInfo(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "directswap [orderid]",
		Short: "get a direct swapinfo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			orderid, err := uuid.FromString(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.QueryDirectSwapOrderParams{OrderID: orderid.String()})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDirectSwapInfo)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var order types.DirectSwapOrder
			if err = cdc.UnmarshalJSON(res, &order); err != nil {
				return err
			}

			return cliCtx.PrintOutput(order)
		},
	}
	return cmd
}
