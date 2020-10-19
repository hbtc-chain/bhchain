package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/version"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	govtype "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

// GetCmdAddTokenProposal implements the command to submit a AddToken proposal
func GetCmdAddTokenProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-token [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an add token proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a new token along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal add-token <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "add token",
  "description": "add a new token in",
  "token_info":{
  	"symbol": "usdt",
  	"issuer": "0xC9476A4919a7E5c7e1760b68F945971769D5c1D8",
  	"chain": "eth",
  	"type": "2",
  	"is_send_enabled": true,
  	"is_deposit_enabled": true,
  	"is_withdrawal_enabled": true,
  	"decimals": "6",
  	"total_supply": "30000000000000000",
  	"collect_threshold": "200000000",
  	"deposit_threshold": "200000000",
  	"open_fee": "28000000000000000000",
  	"sys_open_fee": "28000000000000000000",
  	"withdrawal_fee": "8000000000000000",
  	"max_op_cu_number": "10",
  	"sys_transfer_num": "5",
  	"op_cu_systransfer_num": "30",
  	"gas_limit": "80000",
  	"gas_price": "1000",
	"confirmations": "2",
	"is_nonce_based": true
  }
  "deposit": [
    {
      "denom": "hbc",
      "amount": "10000"
    }
  ]
}
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			proposal, err := ParseAddTokenProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := types.NewAddTokenProposal(proposal.Title, proposal.Description, proposal.TokenInfo)

			msg := govtype.NewMsgSubmitProposal(content, proposal.Deposit, from, proposal.VoteTime)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdTokenParamsChangeProposal implements the command to submit a TokenParamsChange proposal
func GetCmdTokenParamsChangeProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-params-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a token params change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a token params change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal token-params-change <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Token Param Change",
  "description": "token param change proposal",
  "changes": [
    {
      "key": "is_send_enabled",
      "value": true
    },
    {
      "key": "is_deposit_enabled",
      "value": false
    },
    {
      "key": "is_withdrawal_enabled",
      "value": false
    },
    {
      "key": "collect_threshold",
      "value": 10000000000
    },
    {
      "key": "deposit_threshold",
      "value": 20000000000
    },
    {
      "key": "open_fee",
      "value": 30000000000
    },
    {
      "key": "sys_open_fee",
      "value": 40000000000
    },
    {
      "key": "withdrawal_fee",
      "value": 50000000000
    },
    {
      "key": "max_op_cu_number",
      "value": 6
    },
    {
      "key": "systransfer_num",
      "value": 3
    },
    {
      "key": "op_cu_systransfer_num",
      "value": 30
    },
    {
      "key": "gas_limit",
      "value": 90000000000
    }
  ],
  "deposit": [
    {
      "denom": "hbc",
      "amount": "10000"
    }
  ]
}
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			proposal, err := ParseTokenParamsChangeProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			changes := proposal.Changes.ToParamChanges()
			from := cliCtx.GetFromAddress()
			content := types.NewTokenParamsChangeProposal(proposal.Title, proposal.Description, proposal.Symbol, changes)

			msg := govtype.NewMsgSubmitProposal(content, proposal.Deposit, from, proposal.VoteTime)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdDisableTokenProposal implements the command to submit a DisableToken proposal
func GetCmdDisableTokenProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable-token [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a disable token proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a disable token proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal disable-token <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Disable Token",
  "description": "disable token proposal",
  "symbol": "testtoken",
  "deposit": [
    {
      "denom": "hbc",
      "amount": "100000"
    }
  ]
}
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := custodianunit.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			proposal, err := ParseDisableTokenProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := types.NewDisableTokenProposal(proposal.Title, proposal.Description, proposal.Symbol)

			msg := govtype.NewMsgSubmitProposal(content, proposal.Deposit, from, proposal.VoteTime)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
