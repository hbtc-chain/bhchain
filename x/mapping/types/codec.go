package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

var ModuleCdc = codec.New()

func init() {
	codec.RegisterCrypto(ModuleCdc)
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MappingInfo{}, "hbtcchain/mapping/MappingInfo", nil)
	cdc.RegisterConcrete(AddMappingProposal{}, "hbtcchain/mapping/AddMappingProposal", nil)
	cdc.RegisterConcrete(SwitchMappingProposal{}, "hbtcchain/mapping/SwitchMappingProposal", nil)
	cdc.RegisterConcrete(MsgMappingSwap{}, "hbtcchain/mapping/MsgMappingSwap", nil)
	cdc.RegisterConcrete(MsgCreateDirectSwap{}, "hbtcchain/mapping/MsgCreateDirectSwap", nil)
	cdc.RegisterConcrete(MsgCreateFreeSwap{}, "hbtcchain/mapping/MsgCreateFreeSwap", nil)
	cdc.RegisterConcrete(MsgSwapSymbol{}, "hbtcchain/mapping/MsgSwapSymbol", nil)
	cdc.RegisterConcrete(MsgCancelSwap{}, "hbtcchain/mapping/MsgCancelSwap", nil)
	cdc.RegisterConcrete(FreeSwapInfo{}, "hbtcchain/mapping/FreeSwapInfo", nil)
	cdc.RegisterConcrete(FreeSwapOrder{}, "hbtcchain/mapping/FreeSwapOrder", nil)
	cdc.RegisterConcrete(DirectSwapInfo{}, "hbtcchain/mapping/DirectSwapInfo", nil)
	cdc.RegisterConcrete(DirectSwapOrder{}, "hbtcchain/mapping/DirectSwapOrder", nil)
}
