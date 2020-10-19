/*
 * *******************************************************************
 * @项目名称: types
 * @文件名称: codec.go
 * @Date: 2019/06/05
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

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
