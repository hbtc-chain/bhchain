package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

var ModuleCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgNewToken{}, "hbtcchain/hrc20/MsgNewToken", nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
