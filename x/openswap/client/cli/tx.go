package cli

import (
	"fmt"
	"strings"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/version"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	"github.com/hbtc-chain/bhchain/x/openswap/types"

	"github.com/spf13/cobra"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	openswapTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Openswap transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	openswapTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateDex(cdc),
		GetCmdEditDex(cdc),
		GetCmdCreateTradingPair(cdc),
		GetCmdEditTradingPair(cdc),
		GetCmdAddLiquidity(cdc),
		GetCmdRemoveLiquidity(cdc),
		GetCmdSwapExactIn(cdc),
		GetCmdSwapExactOut(cdc),
		GetCmdLimitSwap(cdc),
		GetCmdCancelLimitSwap(cdc),
		GetCmdClaimEarning(cdc),
	)...)
	return openswapTxCmd
}

func GetCmdCreateDex(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-dex",
		Short: "create a new dex",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildCreateDexMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(FlagDexName, "", "The name of dex")
	cmd.Flags().String(FlagDexIncomeReceiver, "", "The income receiver of dex")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagDexName)

	return cmd
}

func GetCmdEditDex(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-dex",
		Short: "edit a dex",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildEditDexMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagDexName, "", "The name of dex")
	cmd.Flags().String(FlagDexIncomeReceiver, "", "The income receiver of dex")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagDexID)

	return cmd
}

func GetCmdCreateTradingPair(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pair",
		Short: "create a trading pair in a dex",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildCreateTradingPairMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagTokenA, "", "The first token of the pair")
	cmd.Flags().String(FlagTokenB, "", "The second token of the pair")
	cmd.Flags().String(FlagIsPublic, "fales", "Whether is public")
	cmd.Flags().String(FlagLpRewardRate, "", "LP reward rate")
	cmd.Flags().String(FlagRefererRewardRate, "", "Referer reward rate")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagDexID)
	cmd.MarkFlagRequired(FlagTokenA)
	cmd.MarkFlagRequired(FlagTokenB)
	cmd.MarkFlagRequired(FlagLpRewardRate)
	cmd.MarkFlagRequired(FlagRefererRewardRate)

	return cmd
}

func GetCmdEditTradingPair(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-pair",
		Short: "edit a trading pair in a dex",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildEditTradingPairMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagTokenA, "", "The first token of the pair")
	cmd.Flags().String(FlagTokenB, "", "The second token of the pair")
	cmd.Flags().String(FlagIsPublic, "", "Whether is public")
	cmd.Flags().String(FlagLpRewardRate, "", "LP reward rate")
	cmd.Flags().String(FlagRefererRewardRate, "", "Referer reward rate")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagDexID)
	cmd.MarkFlagRequired(FlagTokenA)
	cmd.MarkFlagRequired(FlagTokenB)

	return cmd
}

func GetCmdAddLiquidity(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity",
		Short: "add liquidity to a trading pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildAddLiquidityMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagTokenA, "", "The first token of the pair")
	cmd.Flags().String(FlagTokenB, "", "The second token of the pair")
	cmd.Flags().String(FlagTokenAAmount, "", "The amount of the first token")
	cmd.Flags().String(FlagTokenBAmount, "", "The amount of the second token")
	cmd.Flags().String(FlagExpiredTime, "-1", "The expired timestamp of the transaction")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagTokenA)
	cmd.MarkFlagRequired(FlagTokenB)
	cmd.MarkFlagRequired(FlagTokenAAmount)
	cmd.MarkFlagRequired(FlagTokenBAmount)

	return cmd
}

func GetCmdRemoveLiquidity(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity",
		Short: "remove liquidity from a trading pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildRemoveLiquidityMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagTokenA, "", "The first token of the pair")
	cmd.Flags().String(FlagTokenB, "", "The second token of the pair")
	cmd.Flags().String(FlagLiquidity, "", "The liquidity you want to remove")
	cmd.Flags().String(FlagExpiredTime, "-1", "The expired timestamp of the transaction")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagTokenA)
	cmd.MarkFlagRequired(FlagTokenB)
	cmd.MarkFlagRequired(FlagLiquidity)

	return cmd
}

func GetCmdSwapExactIn(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exact-in",
		Short: "swap tokens with exact input amount",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildSwapExactInMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagReferer, "", "The referer of you")
	cmd.Flags().String(FlagReceiver, "", "The receiver of this swap")
	cmd.Flags().String(FlagAmountIn, "", "The exact amount of input token")
	cmd.Flags().String(FlagMinAmountOut, "", "The minimum amount of output token")
	cmd.Flags().String(FlagSwapPath, "", "The swap path")
	cmd.Flags().String(FlagExpiredTime, "-1", "The expired timestamp of the transaction")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagAmountIn)
	cmd.MarkFlagRequired(FlagMinAmountOut)
	cmd.MarkFlagRequired(FlagSwapPath)

	return cmd
}

func GetCmdSwapExactOut(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exact-out",
		Short: "swap tokens with exact output amount",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildSwapExactOutMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagReferer, "", "The referer of you")
	cmd.Flags().String(FlagReceiver, "", "The receiver of this swap")
	cmd.Flags().String(FlagMaxAmountIn, "", "The maximum amount of input token")
	cmd.Flags().String(FlagAmountOut, "", "The exact amount of output token")
	cmd.Flags().String(FlagSwapPath, "", "The swap path")
	cmd.Flags().String(FlagExpiredTime, "-1", "The expired timestamp of the transaction")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagMaxAmountIn)
	cmd.MarkFlagRequired(FlagAmountOut)
	cmd.MarkFlagRequired(FlagSwapPath)

	return cmd
}

func GetCmdLimitSwap(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "limit",
		Short: "create a limit-price swap order",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildLimitSwapMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagReferer, "", "The referer of you")
	cmd.Flags().String(FlagReceiver, "", "The receiver of this swap")
	cmd.Flags().String(FlagAmountIn, "", "The amount of input token")
	cmd.Flags().String(FlagPrice, "", "The price of the order")
	cmd.Flags().String(FlagSide, "", "The side of the order, 0-buy, 1-sell")
	cmd.Flags().String(FlagBaseSymbol, "", "The base symbol of the order")
	cmd.Flags().String(FlagQuoteSymbol, "", "The quote symbol of the order")
	cmd.Flags().String(FlagExpiredTime, "-1", "The expired timestamp of the transaction")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagAmountIn)
	cmd.MarkFlagRequired(FlagBaseSymbol)
	cmd.MarkFlagRequired(FlagQuoteSymbol)
	cmd.MarkFlagRequired(FlagSide)
	cmd.MarkFlagRequired(FlagPrice)

	return cmd
}

func GetCmdCancelLimitSwap(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "cancel a limit-price swap order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel a batch of orders.

Example:
$ %s tx openswap cancel 99466110-708d-47b4-8276-390bf115d675,27eca534-7cd8-4c78-abec-823ffff78afb
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			from := cliCtx.GetFromAddress()
			orderIDs := strings.Split(args[0], ",")
			msg := types.NewMsgCancelLimitSwap(from, orderIDs)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.MarkFlagRequired(client.FlagFrom)

	return cmd
}

func GetCmdClaimEarning(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "claim earning of a trading pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := buildClaimEarningMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Uint32(FlagDexID, 0, "The dex id")
	cmd.Flags().String(FlagTokenA, "", "The first token of the pair")
	cmd.Flags().String(FlagTokenB, "", "The second token of the pair")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagTokenA)
	cmd.MarkFlagRequired(FlagTokenB)

	return cmd
}
