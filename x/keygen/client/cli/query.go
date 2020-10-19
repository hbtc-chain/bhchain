package cli

import (
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{}
}
