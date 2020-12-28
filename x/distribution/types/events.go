package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress                = "set_withdraw_address"
	EventTypeRewards                           = "rewards"
	EventTypeCommission                        = "commission"
	EventTypeWithdrawRewards                   = "withdraw_rewards"
	EventTypeWithdrawCommission                = "withdraw_commission"
	EventTypeProposerReward                    = "proposer_reward"
	EventTypeExecuteCommunityPoolSpendProposal = "execute_communitypool_spend_proposal"

	AttributeKeyWithdrawAddress = "withdraw_address"
	AttributeKeyValidator       = "validator"
	AttributeKeyDelegator       = "delegator"
	AttributeKeyRecipient       = "recipient"
	AttributeKeyAmount          = "amount"

	AttributeValueCategory = ModuleName
)
