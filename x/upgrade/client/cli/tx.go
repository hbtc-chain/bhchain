package cli

import (
	"fmt"
	"time"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	gov "github.com/hbtc-chain/bhchain/x/gov/types"

	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/x/gov/client/cli"

	"github.com/spf13/cobra"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/upgrade/types"
)

const (
	// TimeFormat specifies ISO UTC format for submitting the time for a new upgrade proposal
	TimeFormat = "2006-01-02T15:04:05Z"

	FlagUpgradeHeight = "upgrade-height"
	FlagUpgradeTime   = "upgrade-time"
	FlagUpgradeInfo   = "upgrade-info"
)

// NewCmdSubmitUpgradeProposal implements a command handler for submitting a software upgrade proposal transaction.
func NewCmdSubmitUpgradeProposal(cdc *codec.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "software-upgrade [name] (--upgrade-height [height] | --upgrade-time [time]) (--upgrade-info [info]) [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a software upgrade proposal",
		Long: "Submit a software upgrade along with an initial deposit.\n" +
			"Please specify a unique name and height OR time for the upgrade to take effect.\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			content, err := parseArgsToContent(cmd, name)
			if err != nil {
				return err
			}

			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			from := cliCtx.GetFromAddress()

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
	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen (not to be used together with --upgrade-time)")
	cmd.Flags().String(FlagUpgradeTime, "", fmt.Sprintf("The time at which the upgrade must happen (ex. %s) (not to be used together with --upgrade-height)", TimeFormat))
	cmd.Flags().String(FlagUpgradeInfo, "", "Optional info for the planned upgrade such as commit hash, etc.")
	cmd.Flags().Uint32(cli.FlagVoteTime, 0, "votetime of proposal")

	return cmd
}

// NewCmdSubmitCancelUpgradeProposal implements a command handler for submitting a software upgrade cancel proposal transaction.
func NewCmdSubmitCancelUpgradeProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-software-upgrade [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a software upgrade proposal",
		Long:  "Cancel a software upgrade along with an initial deposit.",
		RunE: func(cmd *cobra.Command, args []string) error {

			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			from := cliCtx.GetFromAddress()
			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoins(depositStr)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			voteTime, err := cmd.Flags().GetUint32(cli.FlagVoteTime)
			if err != nil {
				return err
			}

			content := types.NewCancelSoftwareUpgradeProposal(title, description)

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

func parseArgsToContent(cmd *cobra.Command, name string) (gov.Content, error) {
	title, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(cli.FlagDescription)
	if err != nil {
		return nil, err
	}

	height, err := cmd.Flags().GetInt64(FlagUpgradeHeight)
	if err != nil {
		return nil, err
	}

	timeStr, err := cmd.Flags().GetString(FlagUpgradeTime)
	if err != nil {
		return nil, err
	}

	if height != 0 && len(timeStr) != 0 {
		return nil, fmt.Errorf("only one of --upgrade-time or --upgrade-height should be specified")
	}

	var upgradeTime time.Time
	if len(timeStr) != 0 {
		upgradeTime, err = time.Parse(TimeFormat, timeStr)
		if err != nil {
			return nil, err
		}
	}

	info, err := cmd.Flags().GetString(FlagUpgradeInfo)
	if err != nil {
		return nil, err
	}

	plan := types.Plan{Name: name, Time: upgradeTime, Height: height, Info: info}
	content := types.NewSoftwareUpgradeProposal(title, description, plan)
	return content, nil
}
