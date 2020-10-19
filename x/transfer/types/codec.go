package types

import (
	"github.com/hbtc-chain/bhchain/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSend{}, "hbtcchain/transfer/MsgSend", nil)
	cdc.RegisterConcrete(MsgMultiSend{}, "hbtcchain/transfer/MsgMultiSend", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "hbtcchain/transfer/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgConfirmedDeposit{}, "hbtcchain/transfer/MsgConfirmedDeposit", nil)
	cdc.RegisterConcrete(MsgCollectWaitSign{}, "hbtcchain/transfer/MsgCollectWaitSign", nil)
	cdc.RegisterConcrete(MsgCollectSignFinish{}, "hbtcchain/transfer/MsgCollectSignFinish", nil)
	cdc.RegisterConcrete(MsgCollectFinish{}, "hbtcchain/transfer/MsgCollectFinish", nil)
	cdc.RegisterConcrete(MsgWithdrawal{}, "hbtcchain/transfer/MsgWithdrawal", nil)
	cdc.RegisterConcrete(MsgWithdrawalConfirm{}, "hbtcchain/transfer/MsgWithdrawalConfirm", nil)
	cdc.RegisterConcrete(MsgWithdrawalWaitSign{}, "hbtcchain/transfer/MsgWithdrawalWaitSign", nil)
	cdc.RegisterConcrete(MsgWithdrawalSignFinish{}, "hbtcchain/transfer/MsgWithdrawalSignFinish", nil)
	cdc.RegisterConcrete(MsgWithdrawalFinish{}, "hbtcchain/transfer/MsgWithdrawalFinish", nil)
	cdc.RegisterConcrete(MsgSysTransfer{}, "hbtcchain/transfer/MsgSysTransfer", nil)
	cdc.RegisterConcrete(MsgSysTransferWaitSign{}, "hbtcchain/transfer/MsgSysTransferWaitSign", nil)
	cdc.RegisterConcrete(MsgSysTransferSignFinish{}, "hbtcchain/transfer/MsgSysTransferSignFinish", nil)
	cdc.RegisterConcrete(MsgSysTransferFinish{}, "hbtcchain/transfer/MsgSysTransferFinish", nil)
	cdc.RegisterConcrete(MsgOpcuAssetTransfer{}, "hbtcchain/transfer/MsgOpcuAssetTransfer", nil)
	cdc.RegisterConcrete(MsgOpcuAssetTransferWaitSign{}, "hbtcchain/transfer/MsgOpcuAssetTransferWaitSign", nil)
	cdc.RegisterConcrete(MsgOpcuAssetTransferSignFinish{}, "hbtcchain/transfer/MsgOpcuAssetTransferSignFinish", nil)
	cdc.RegisterConcrete(MsgOpcuAssetTransferFinish{}, "hbtcchain/transfer/MsgOpcuAssetTransferFinish", nil)
	cdc.RegisterConcrete(MsgOrderRetry{}, "hbtcchain/transfer/MsgOrderRetry", nil)
	cdc.RegisterConcrete(MsgCancelWithdrawal{}, "hbtcchain/transfer/MsgCancelWithdrawal", nil)
	cdc.RegisterConcrete(&TxVote{}, "hbtcchain/transfer/FinishTxVote", nil)
	cdc.RegisterConcrete(&OrderRetryVoteBox{}, "hbtcchain/transfer/OrderRetryVoteBox", nil)
	cdc.RegisterConcrete(&OrderRetryVoteItem{}, "hbtcchain/transfer/OrderRetryVoteItem", nil)
}

// ModuleCdc - module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
