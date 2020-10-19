package types

// bank module event types
const (
	EventTypeTransfer               = "transfer"
	EventTypeMultiTransfer          = "multi_transfer"
	EventTypeDeposit                = "deposit"
	EventTypeDepositConfirm         = "deposit_confirm"
	EventTypeCollectWaitSign        = "collect_wait_sign"
	EventTypeCollectSignFinish      = "collect_sign_finish"
	EventTypeCollectFinish          = "collect_finish"
	EventTypeWithdrawal             = "withdrawal"
	EventTypeWithdrawalConfirm      = "withdrawal_confirm"
	EventTypeWithdrawalWaitSign     = "withdrawal_wait_sign"
	EventTypeWithdrawalSignFinish   = "withdrawal_sign_finish"
	EventTypeWithdrawalFinish       = "withdrawal_finish"
	EventTypeCancelWithdrawal       = "cancel_withdrawal"
	EventTypeSysTransfer            = "sys_transfer"
	EventTypeSysTransferWaitSign    = "sys_transfer_wait_sign"
	EventTypeSysTransferSignFinish  = "sys_transfer_sign_finish"
	EventTypeSysTransferFinish      = "sys_transfer_finish"
	EventTypeOpcuTransfer           = "opcu_transfer"
	EventTypeOpcuTransferWaitSign   = "opcu_transfer_wait_sign"
	EventTypeOpcuTransferSignFinish = "opcu_transfer_sign_finish"
	EventTypeOpcuTransferFinish     = "opcu_transfer_finish"
	EventTypeOrderRetry             = "order_retry"

	AttributeKeyRecipient       = "recipient"
	AttributeKeySender          = "sender"
	AttributeKeySymbol          = "symbol"
	AttributeKeyAmount          = "amount"
	AttributeKeyHash            = "hash"
	AttributeKeyIndex           = "index"
	AttributeKeyMemo            = "memo"
	AttributeKeyOrderIDs        = "order_ids"
	AttributeKeyOrderID         = "order_id"
	AttributeKeyValidOrderIDs   = "valid_order_ids"
	AttributeKeyInvalidOrderIDs = "invalid_order_ids"

	AttributeValueCategory = ModuleName
)
