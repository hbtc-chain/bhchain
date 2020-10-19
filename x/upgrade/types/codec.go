package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(Plan{}, "hbtcchain/Plan", nil)
	cdc.RegisterConcrete(&SoftwareUpgradeProposal{}, "hbtcchain/SoftwareUpgradeProposal", nil)
	cdc.RegisterConcrete(&CancelSoftwareUpgradeProposal{}, "hbtcchain/CancelSoftwareUpgradeProposal", nil)
}
