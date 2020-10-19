package cli

import (
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{}
}
