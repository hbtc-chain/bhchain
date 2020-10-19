package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

var ModuleCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&TradingPair{}, "hbtcchain/openswap/TradingPair", nil)
	cdc.RegisterConcrete(&AddrLiquidity{}, "hbtcchain/openswap/AddrLiquidity", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "hbtcchain/openswap/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "hbtcchain/openswap/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgSwapExactIn{}, "hbtcchain/openswap/MsgSwapExactIn", nil)
	cdc.RegisterConcrete(MsgSwapExactOut{}, "hbtcchain/openswap/MsgSwapExactOut", nil)
	cdc.RegisterConcrete(MsgLimitSwap{}, "hbtcchain/openswap/MsgLimitSwap", nil)
	cdc.RegisterConcrete(MsgCancelLimitSwap{}, "hbtcchain/openswap/MsgCancelLimitSwap", nil)
	cdc.RegisterConcrete(MsgClaimEarning{}, "hbtcchain/openswap/MsgClaimEarning", nil)
	cdc.RegisterConcrete(&Order{}, "hbtcchain/openswap/Order", nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
