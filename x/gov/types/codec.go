package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

// module codec
var ModuleCdc = codec.New()

// RegisterCodec registers all the necessary types and interfaces for
// governance.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Content)(nil), nil)

	cdc.RegisterConcrete(MsgSubmitProposal{}, "hbtcchain/gov/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "hbtcchain/gov/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgVote{}, "hbtcchain/gov/MsgVote", nil)
	cdc.RegisterConcrete(MsgDaoVote{}, "hbtcchain/gov/MsgDaoVote", nil)
	cdc.RegisterConcrete(MsgCancelDaoVote{}, "hbtcchain/gov/MsgCancelDaoVote", nil)

	cdc.RegisterConcrete(TextProposal{}, "hbtcchain/gov/TextProposal", nil)
}

// RegisterProposalTypeCodec registers an external proposal content type defined
// in another module for the internal ModuleCdc. This allows the MsgSubmitProposal
// to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

// TODO determine a good place to seal this codec
func init() {
	RegisterCodec(ModuleCdc)
}
