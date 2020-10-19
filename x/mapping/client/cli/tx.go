package cli

import (
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strconv"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	"github.com/hbtc-chain/bhchain/x/gov/client/cli"
	gov "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/mapping/types"
	"github.com/spf13/cobra"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	mappingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "mapping transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		//	RunE:                       utils.ValidateCmd,
	}

	mappingTxCmd.AddCommand(client.PostCommands(
		GetCmdMappingSwap(cdc),
		CreateDirectSwapCmd(cdc),
		CreateFreeSwapCmd(cdc),
		SwapSymbolCmd(cdc),
		CancelSwapCmd(cdc),
	)...)

	return mappingTxCmd
}

func NewCmdSubmitAddMappingProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-mapping [issue-symbol] [target-symbol] [total-supply]",
		Short: "Create a add-mapping proposal with issue symbol, target symbol, and total supply",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			from := cliCtx.GetFromAddress()

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}
			totalSupply, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return errors.New("invalid total supply")
			}
			content := types.NewAddMappingProposal(from.String(), title, description, sdk.Symbol(args[0]), sdk.Symbol(args[1]), totalSupply)
			err = content.ValidateBasic()
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoins(depositStr)
			if err != nil {
				return err
			}

			voteTime, err := cmd.Flags().GetUint32(cli.FlagVoteTime)
			if err != nil {
				return err
			}

			msg := gov.NewMsgSubmitProposal(content, deposit, from, voteTime)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Uint32(cli.FlagVoteTime, 0, "votetime of proposal")

	return cmd
}

func NewCmdSubmitSwitchMappingProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch-mapping [issue-symbol] [enable]",
		Short: "Create a switch-mapping proposal with issue symbol and enable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			from := cliCtx.GetFromAddress()

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}
			enable, err := strconv.ParseBool(args[1])
			if err != nil {
				return err
			}
			content := types.NewSwitchMappingProposal(title, description, sdk.Symbol(args[0]), enable)
			err = content.ValidateBasic()
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoins(depositStr)
			if err != nil {
				return err
			}
			voteTime, err := cmd.Flags().GetUint32(cli.FlagVoteTime)
			if err != nil {
				return err
			}
			msg := gov.NewMsgSubmitProposal(content, deposit, from, voteTime)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Uint32(cli.FlagVoteTime, 0, "votetime of proposal")

	return cmd
}

func GetCmdMappingSwap(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [issue-symbol] [amount]",
		Short: "Swap amount by a issue symbol",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			if coins.Len() != 1 {
				return fmt.Errorf("invalid amount coins, only 1 coin is allowed: %s",
					coins)
			}

			msg := types.NewMsgMappingSwap(
				cliCtx.GetFromAddress(),
				sdk.Symbol(args[0]),
				coins)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func CancelSwapCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelswap [swap-type] [orderid]",
		Short: "cancel a swap by orderid",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			swaptype, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelSwap(cliCtx.GetFromAddress().String(), args[1], swaptype)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func CreateFreeSwapCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createfreeswap [srcsymbol] [dstsymbol] [totalamt] [maxswapamt] [minswapamt] [swapprice] [expiredtime] [desc]",
		Short: "create a freeswap",
		Args:  cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			var swapInfo types.FreeSwapInfo
			swapInfo.SrcSymbol = sdk.Symbol(args[0])
			swapInfo.TargetSymbol = sdk.Symbol(args[1])
			swapInfo.TotalAmount, _ = sdk.NewIntFromString(args[2])
			swapInfo.MaxSwapAmount, _ = sdk.NewIntFromString(args[3])
			swapInfo.MinSwapAmount, _ = sdk.NewIntFromString(args[4])
			swapInfo.SwapPrice, _ = sdk.NewIntFromString(args[5])
			time, _ := sdk.NewIntFromString(args[6])
			swapInfo.ExpiredTime = time.Int64()
			swapInfo.Desc = args[7]
			orderID := uuid.NewV4().String()

			msg := types.NewMsgCreateFreeSwap(cliCtx.GetFromAddress().String(), orderID, swapInfo)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func CreateDirectSwapCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createdirectswap [srcsymbol] [dstsymbol] [totalamt] [swapamt] [receiveaddr] [expiredtime] [desc]",
		Short: "create a directswap",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			var swapInfo types.DirectSwapInfo
			swapInfo.SrcSymbol = sdk.Symbol(args[0])
			swapInfo.TargetSymbol = sdk.Symbol(args[1])
			swapInfo.Amount, _ = sdk.NewIntFromString(args[2])
			swapInfo.SwapAmount, _ = sdk.NewIntFromString(args[3])
			swapInfo.ReceiveAddr = args[4]
			time, _ := sdk.NewIntFromString(args[5])
			swapInfo.ExpiredTime = time.Int64()
			swapInfo.Desc = args[6]
			orderID := uuid.NewV4().String()

			msg := types.NewMsgCreateDirectSwap(cliCtx.GetFromAddress().String(), orderID, swapInfo)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func SwapSymbolCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swapsymbol [swaptype] [swapamt] [orderid]",
		Short: "swap a symbol",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.NewCLIContext().WithCodec(cdc)
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			swaptype, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			swapAmt, _ := sdk.NewIntFromString(args[1])
			msg := types.NewMsgSwapSymbol(cliCtx.GetFromAddress().String(), args[2], swapAmt, swaptype)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
