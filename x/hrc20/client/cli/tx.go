package cli

import (
	"fmt"
	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "hrc20 subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(client.PostCommands(
		GetCmdNewToken(cdc),
	)...)

	return txCmd
}

//create a token in bhchain
func GetCmdNewToken(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new-token [to][symbol][decimals][totalSupply]",
		Short: "new a token",
		Long:  ` Example: new-token HBCxxx bhetc 18 1000000000000000000000000000`,

		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			to, err := sdk.CUAddressFromBase58(args[0])
			if err != nil {
				return err
			}

			symbol := sdk.Symbol(args[1])
			if !symbol.IsValidTokenName() {
				return err
			}

			decimals, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("Fail to parse decimals:%v", args[2])
			}

			totalSupply, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("Fail to parse totalSupply:%v", args[3])
			}

			from := cliCtx.GetFromAddress()
			msg := types.NewMsgNewToken(from, to, symbol.String(), uint64(decimals.Int64()), totalSupply)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
