package cli

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	ctypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/keygen/types"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var flagOrderID = "order-id"

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "KeyGen transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(client.PostCommands(
		GetCmdKeyGen(cdc),
		GetCmdNewOpCU(cdc),
	)...)

	return txCmd
}

func GetCmdKeyGen(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen [from_key_or_address] [symbol] [to]",
		Short: "keygen",
		Long:  ` Example: keygen alice btc HBCao7JRxCAUc89DkUjSf8r4nVRURyLzPb6b`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := ctypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			symbol := sdk.Symbol(args[1])
			if !symbol.IsValid() {
				return fmt.Errorf("Invalid symbol:%v", args[1])
			}

			var to sdk.CUAddress
			var err error
			if len(args) > 2 {
				to, err = sdk.CUAddressFromBase58(args[2])
				if err != nil {
					return err
				}
			} else {
				to = sdk.CUAddress(cliCtx.GetFromAddress())
			}

			orderID := viper.GetString(flagOrderID)
			if len(orderID) == 0 {
				orderID = uuid.NewV4().String()
			}

			msg := types.NewMsgKeyGen(orderID, symbol, sdk.CUAddress(cliCtx.GetFromAddress()), to)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagOrderID, "", "order ID of keygen is a uuid string. e.g. 'fc9ffd98-c99f-4a7c-b3ab-a517fed807c4'")
	return cmd
}

func GetCmdNewOpCU(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "newopcu [from_key_or_address] [symbol] [Op_CU_address]",
		Short: "create Op CU with address",
		Long:  ` Example: newopcu btc HBCao7JRxCAUc89DkUjSf8r4nVRURyLzPb6b`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := ctypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			symbol := sdk.Symbol(args[1])
			if !symbol.IsValid() {
				return fmt.Errorf("Invalid symbol:%v", args[1])
			}

			to, err := sdk.CUAddressFromBase58(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgNewOpCU(symbol.String(), to, sdk.CUAddress(cliCtx.GetFromAddress()))
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
