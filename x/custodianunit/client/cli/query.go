package cli

import (
	"fmt"
	"strings"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/client/flags"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	flagTags  = "tags"
	flagPage  = "page"
	flagLimit = "limit"
)

// GetTxCmd returns the transaction commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the custodian unit module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(GetCUCmd(cdc))
	cmd.AddCommand(GetOpCUCmd(cdc))

	return cmd
}

// GetCUCmd returns a query CU that will display the state of the
// CU at a given address.
func GetCUCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cuinfo [address]",
		Aliases: []string{"ci"},
		Short:   "Query custodian unit info",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			cuGetter := types.NewCURetriever(cliCtx)

			key, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}

			if err := cuGetter.EnsureExists(key); err != nil {
				return err
			}

			cu, err := cuGetter.GetCU(key)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(cu)
		},
	}

	return flags.GetCommands(cmd)[0]
}

// QueryTxsByEventsCmd returns a command to search through transactions by events.
func QueryTxsByEventsCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Query for paginated transactions that match a set of tags",
		Long: strings.TrimSpace(`
Search for transactions that match the exact given tags where results are paginated.

Example:
$ <appcli> query txs --tags '<tag1>:<value1>&<tag2>:<value2>' --page 1 --limit 30
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			tagsStr := viper.GetString(flagTags)
			tagsStr = strings.Trim(tagsStr, "'")

			var tags []string
			if strings.Contains(tagsStr, "&") {
				tags = strings.Split(tagsStr, "&")
			} else {
				tags = append(tags, tagsStr)
			}

			var tmTags []string
			for _, tag := range tags {
				if !strings.Contains(tag, ":") {
					return fmt.Errorf("%s should be of the format <key>:<value>", tagsStr)
				} else if strings.Count(tag, ":") > 1 {
					return fmt.Errorf("%s should only contain one <key>:<value> pair", tagsStr)
				}

				keyValue := strings.Split(tag, ":")
				if keyValue[0] == tmtypes.TxHeightKey {
					tag = fmt.Sprintf("%s=%s", keyValue[0], keyValue[1])
				} else {
					tag = fmt.Sprintf("%s='%s'", keyValue[0], keyValue[1])
				}

				tmTags = append(tmTags, tag)
			}

			page := viper.GetInt(flagPage)
			limit := viper.GetInt(flagLimit)

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txs, err := utils.QueryTxsByEvents(cliCtx, tmTags, page, limit)
			if err != nil {
				return err
			}

			var output []byte
			if cliCtx.Indent {
				output, err = cdc.MarshalJSONIndent(txs, "", "  ")
			} else {
				output, err = cdc.MarshalJSON(txs)
			}

			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	cmd.Flags().Bool(flags.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(flags.FlagTrustNode, cmd.Flags().Lookup(flags.FlagTrustNode))

	cmd.Flags().String(flagTags, "", "tag:value list of tags that must match")
	cmd.Flags().Uint32(flagPage, rest.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Uint32(flagLimit, rest.DefaultLimit, "Query number of transactions results per page returned")
	cmd.MarkFlagRequired(flagTags)

	return cmd
}

// QueryTxCmd implements the default command for a tx query.
func QueryTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "Query for a transaction by hash in a committed block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			output, err := utils.QueryTx(cliCtx, args[0])
			if err != nil {
				return err
			}

			if output.Empty() {
				return fmt.Errorf("No transaction found with hash %s", args[0])
			}

			return cliCtx.PrintOutput(output)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	cmd.Flags().Bool(flags.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(flags.FlagTrustNode, cmd.Flags().Lookup(flags.FlagTrustNode))

	return cmd
}

// GetCUNonceCmd returns the sequence number of CU at a given address.
func GetCUNonceCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nonce [address]",
		Short: "Query custodian unit nonce",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			cuGetter := types.NewCURetriever(cliCtx)

			key, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}

			if err := cuGetter.EnsureExists(key); err != nil {
				return err
			}

			cu, err := cuGetter.GetCU(key)
			if err != nil {
				return err
			}
			println(fmt.Sprintf("nonce%v", cu.GetSequence()))
			return nil
		},
	}

	return flags.GetCommands(cmd)[0]
}

// GetOpCUCmd returns the operations CU list of the given symbol
func GetOpCUCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "op [symbol]",
		Short: "Query operation custodian units",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			cuGetter := types.NewCURetriever(cliCtx)

			cus, _, err := cuGetter.GetOpCUWithHeight(args[0])
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(cus)
		},
	}

	return flags.GetCommands(cmd)[0]
}
