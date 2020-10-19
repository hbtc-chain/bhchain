package cli

import (
	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "CU transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		//GetCmdNewOpCU(cdc),
		GetMultiSignCommand(cdc),
		GetSignCommand(cdc),
	)
	return txCmd
}
