package types

const (
	EventTypeExecuteAddMappingProposal    = "execute_add_mapping_proposal"
	EventTypeExecuteSwitchMappingProposal = "execute_switch_mapping_proposal"
	EventTypeCreateFreeSwap               = "create_free_swap"
	EventTypeCreateDirectSwap             = "create_direct_swap"
	EventTypeCancelSwap                   = "cancel_swap"
	EventTypeSwapSymbol                   = "swap_symbol"

	AttributeKeyFrom        = "from"
	AttributeKeyIssueToken  = "issue_token"
	AttributeKeyTargetToken = "target_token"
	AttributeKeyTotalSupply = "total_supply"
	AttributeKeyEnable      = "enable"
	AttributeKeyOrderID     = "order_id"
	AttributeKeyAmount      = "amount"
	AttributeKeySwapType    = "swap_type"
)
