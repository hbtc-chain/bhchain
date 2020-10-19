package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/version"
	"github.com/hbtc-chain/bhchain/x/openswap/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	openswapQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the openswap module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	openswapQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryTradingPair(queryRoute, cdc),
		GetCmdQueryAddrLiquidity(queryRoute, cdc),
		GetCmdQueryOrderbook(queryRoute, cdc),
		GetCmdQueryOrder(queryRoute, cdc),
		GetCmdQueryUnfinishedOrders(queryRoute, cdc),
		GetCmdQueryEarnings(queryRoute, cdc),
		GetCmdQueryParams(queryRoute, cdc),
	)...)
	return openswapQueryCmd
}

func GetCmdQueryTradingPair(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pair [tokenA] [tokenB]",
		Short: "Query a trading pair",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about a trading pair.

Example:
$ %s query openswap pair usdt hbc 
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			tokenA, tokenB := sdk.Symbol(args[0]), sdk.Symbol(args[1])

			params := types.NewQueryTradingPairParams(tokenA, tokenB)
			bz := cliCtx.Codec.MustMarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryTradingPair), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("trading pair %s-%s not found", args[0], args[1])
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryAddrLiquidity(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "liquidity [addr]",
		Short: "Query all liquidity of an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all liquidity of an address.

Example:
$ %s query openswap liquidity HBCWn2fXDbRPjyrzPyjYLsXYcAcUjE1PJDq9
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryAddrLiquidityParams(addr)
			bz := cliCtx.Codec.MustMarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryAddrLiquidity), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("trading pair %s-%s not found", args[0], args[1])
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryOrderbook(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orderbook [pair]",
		Short: "Query the orderbook of a trading pair",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the orderbook of a trading pair.

Example:
$ %s query openswap orderbook btc-usdt
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			symbols := strings.Split(args[0], "-")
			if len(symbols) != 2 {
				return errors.New("invalid trading pair")
			}
			merge, _ := strconv.ParseBool(viper.GetString(FlagMergeOrderbook))
			params := types.NewQueryOrderbookParams(sdk.Symbol(symbols[0]), sdk.Symbol(symbols[1]), merge)
			bz := cliCtx.Codec.MustMarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryOrderbook), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("trading pair %s not found", args[0])
			}

			fmt.Println(string(res))
			return nil
		},
	}

	cmd.Flags().String(FlagMergeOrderbook, "false", "Whether to merge orderbook")

	return cmd
}

func GetCmdQueryOrder(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "order [orderID]",
		Short: "Query an order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query an order.

Example:
$ %s query openswap order 99466110-708d-47b4-8276-390bf115d675
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			orderID := args[0]
			params := types.NewQueryOrderParams(orderID)
			bz := cliCtx.Codec.MustMarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryOrder), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("order %s not found", orderID)
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryUnfinishedOrders(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pending-orders [pair] [addr]",
		Short: "Query the pending orders of an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the pending orders of an address.

Example:
$ %s query openswap pending-orders btc-usdt HBCWn2fXDbRPjyrzPyjYLsXYcAcUjE1PJDq9
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			symbols := strings.Split(args[0], "-")
			if len(symbols) != 2 {
				return errors.New("invalid trading pair")
			}
			addr, err := sdk.CUAddressFromBase58(args[1])
			if err != nil {
				return err
			}
			params := types.NewQueryUnfinishedOrderParams(sdk.Symbol(symbols[0]), sdk.Symbol(symbols[1]), addr)
			bz := cliCtx.Codec.MustMarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryUnfinishedOrder), bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryEarnings(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "earnings [addr]",
		Short: "Query all unclaimed earnings of an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all unclaimed earnings of an address.

Example:
$ %s query openswap earnings HBCWn2fXDbRPjyrzPyjYLsXYcAcUjE1PJDq9
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryUnclaimedEarningParams(addr)
			bz := cliCtx.Codec.MustMarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryUnclaimedEarnings), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("trading pair %s-%s not found", args[0], args[1])
			}

			fmt.Println(string(res))
			return nil
		},
	}
}

func GetCmdQueryParams(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current openswap parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as openswap parameters.

Example:
$ %s query openswap params
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", storeName, types.QueryParameters)
			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			cdc.MustUnmarshalJSON(bz, &params)
			return cliCtx.PrintOutput(params)
		},
	}
}
