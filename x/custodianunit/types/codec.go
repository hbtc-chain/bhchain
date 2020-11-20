package types

import (
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.CustodianUnit)(nil), nil)
	cdc.RegisterConcrete(&BaseCU{}, "hbtcchain/CustodianUnit", nil)
	cdc.RegisterConcrete(&BaseCUs{}, "hbtcchain/CustodianUnits", nil)
	cdc.RegisterInterface((*exported.VestingCU)(nil), nil)
	cdc.RegisterConcrete(StdTx{}, "hbtcchain/StdTx", nil)
}

// module wide codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
