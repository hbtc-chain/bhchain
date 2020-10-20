package types

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
)

var ModuleCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(sdk.TokenInfo{}, "hbtcchain/types/TokenInfo", nil)
	cdc.RegisterConcrete(sdk.TokensGasPrice{}, "hbtcchain/types/TokensGasPrice", nil)
	cdc.RegisterConcrete(MsgSynGasPrice{}, "hbtcchain/token/MsgSynGasPrice", nil)
	cdc.RegisterConcrete(AddTokenProposal{}, "hbtcchain/AddTokenProposal", nil)
	cdc.RegisterConcrete(TokenParamsChangeProposal{}, "hbtcchain/TokenParamsChangeProposal", nil)
	cdc.RegisterConcrete(DisableTokenProposal{}, "hbtcchain/DisableTokenProposal", nil)
	cdc.RegisterConcrete(&GasPriceVoteBox{}, "hbtcchain/token/GasPriceVoteBox", nil)
	cdc.RegisterConcrete(&GasPriceVoteItem{}, "hbtcchain/token/GasPriceVoteItem", nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
