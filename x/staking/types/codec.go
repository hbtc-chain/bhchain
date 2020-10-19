package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateValidator{}, "hbtcchain/MsgCreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "hbtcchain/MsgEditValidator", nil)
	cdc.RegisterConcrete(MsgKeyNodeHeartbeat{}, "hbtcchain/MsgKeyNodeHeartbeat", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "hbtcchain/MsgDelegate", nil)
	cdc.RegisterConcrete(MsgUndelegate{}, "hbtcchain/MsgUndelegate", nil)
	cdc.RegisterConcrete(MsgBeginRedelegate{}, "hbtcchain/MsgBeginRedelegate", nil)
}

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
