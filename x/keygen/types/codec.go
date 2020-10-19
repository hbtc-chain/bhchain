package types

import (
	"github.com/hbtc-chain/bhchain/codec"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
)

var ModuleCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(cutypes.StdSignature{}, "hbtcchain/keygen/StdSignature", nil)
	cdc.RegisterConcrete(MsgKeyGen{}, "hbtcchain/keygen/MsgKeyGen", nil)
	cdc.RegisterConcrete(MsgKeyGenWaitSign{}, "hbtcchain/keygen/MsgKeyGenWaitSign", nil)
	cdc.RegisterConcrete(MsgKeyGenFinish{}, "hbtcchain/keygen/MsgKeyGenFinish", nil)
	cdc.RegisterConcrete(MsgPreKeyGen{}, "hbtcchain/keygen/MsgPreKeyGen", nil)
	cdc.RegisterConcrete(MsgOpcuMigrationKeyGen{}, "hbtcchain/keygen/MsgOpcuMigrationKeyGen", nil)
	cdc.RegisterConcrete(MsgNewOpCU{}, "hbtcchain/keygen/MsgNewOpCU", nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
