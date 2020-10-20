package receipt

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
	cdc.RegisterInterface((*Flow)(nil), nil)
	cdc.RegisterConcrete(BalanceFlow{}, "hbtcchain/receipt/BalanceFlow", nil)
	cdc.RegisterConcrete(OrderFlow{}, "hbtcchain/receipt/OrderFlow", nil)
	cdc.RegisterConcrete(DepositFlow{}, "hbtcchain/receipt/DepositFlow", nil)
	cdc.RegisterConcrete(MappingBalanceFlow{}, "hbtcchain/receipt/MappingBalanceFlow", nil)
	cdc.RegisterConcrete(sdk.DepositConfirmedFlow{}, "hbtcchain/receipt/DepositConfrimedFlow", nil)
	cdc.RegisterConcrete(sdk.CollectWaitSignFlow{}, "hbtcchain/receipt/CollectWaitSignFlow", nil)
	cdc.RegisterConcrete(sdk.CollectSignFinishFlow{}, "hbtcchain/receipt/CollectSignFinishFlow", nil)
	cdc.RegisterConcrete(sdk.CollectFinishFlow{}, "hbtcchain/receipt/CollectFinishFlow", nil)
	cdc.RegisterConcrete(sdk.WithdrawalFlow{}, "hbtcchain/receipt/WithdrawalFlow", nil)
	cdc.RegisterConcrete(sdk.WithdrawalConfirmFlow{}, "hbtcchain/receipt/WithdrawalConfirmFlow", nil)
	cdc.RegisterConcrete(sdk.WithdrawalWaitSignFlow{}, "hbtcchain/receipt/WithdrawalWaitSignFlow", nil)
	cdc.RegisterConcrete(sdk.WithdrawalSignFinishFlow{}, "hbtcchain/receipt/WithdrawalSignFinishFlow", nil)
	cdc.RegisterConcrete(sdk.WithdrawalFinishFlow{}, "hbtcchain/receipt/WithdrawalFinishFlow", nil)
	cdc.RegisterConcrete(sdk.SysTransferFlow{}, "hbtcchain/receipt/SysTransferFlow", nil)
	cdc.RegisterConcrete(sdk.SysTransferWaitSignFlow{}, "hbtcchain/receipt/SysTransferWaitSignFlow", nil)
	cdc.RegisterConcrete(sdk.SysTransferSignFinishFlow{}, "hbtcchain/receipt/SysTransferSignFinishFlow", nil)
	cdc.RegisterConcrete(sdk.SysTransferFinishFlow{}, "hbtcchain/receipt/SysTransferFinishFlow", nil)
	cdc.RegisterConcrete(sdk.OpcuAssetTransferFlow{}, "hbtcchain/receipt/OpcuAssetTransferFlow", nil)
	cdc.RegisterConcrete(sdk.OpcuAssetTransferWaitSignFlow{}, "hbtcchain/receipt/OpcuAssetTransferWaitSignFlow", nil)
	cdc.RegisterConcrete(sdk.OpcuAssetTransferSignFinishFlow{}, "hbtcchain/receipt/OpcuAssetTransferSignFinishFlow", nil)
	cdc.RegisterConcrete(sdk.OpcuAssetTransferFinishFlow{}, "hbtcchain/receipt/OpcuAssetTransferFinishFlow", nil)

	cdc.RegisterConcrete(sdk.KeyGenFlow{}, "hbtcchain/receipt/KeyGenFlow", nil)
	cdc.RegisterConcrete(sdk.KeyGenWaitSignFlow{}, "hbtcchain/receipt/KeyGenWaitSignFlow", nil)
	cdc.RegisterConcrete(sdk.KeyGenFinishFlow{}, "hbtcchain/receipt/KeyGenFinishFlow", nil)
	cdc.RegisterConcrete(sdk.OrderRetryFlow{}, "hbtcchain/receipt/OrderRetryFlow", nil)
}
