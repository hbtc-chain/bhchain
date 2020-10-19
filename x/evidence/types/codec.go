package types

import (
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.Evidence)(nil), nil)
	cdc.RegisterInterface((*exported.Vote)(nil), nil)
	cdc.RegisterInterface((*exported.VoteBox)(nil), nil)
	cdc.RegisterConcrete(MsgSubmitEvidence{}, "hbtcchain/MsgSubmitEvidence", nil)
	cdc.RegisterConcrete(BoolVote(false), "hbtcchain/evidence/BoolVote", nil)
	cdc.RegisterConcrete(&VoteBox{}, "hbtcchain/evidence/VoteBox", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
