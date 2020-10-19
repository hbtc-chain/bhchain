package order

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*sdk.Order)(nil), nil)
	cdc.RegisterConcrete(&sdk.OrderBase{}, "hbtcchain/order/OrderBase", nil)
	cdc.RegisterConcrete(&sdk.OrderCollect{}, "hbtcchain/order/OrderCollect", nil)
	cdc.RegisterConcrete(&sdk.OrderWithdrawal{}, "hbtcchain/order/OrderWithdrawal", nil)
	cdc.RegisterConcrete(&sdk.OrderKeyGen{}, "hbtcchain/order/OrderKeyGen", nil)
	cdc.RegisterConcrete(&sdk.OrderSysTransfer{}, "hbtcchain/order/OrderSysTransfer", nil)
	cdc.RegisterConcrete(&sdk.OrderOpcuAssetTransfer{}, "hbtcchain/order/OrderOpcuAssetTransfer", nil)
	cdc.RegisterConcrete(&sdk.TxFinishNodeData{}, "hbtcchain/order/TxFinishNodeData", nil)
}
