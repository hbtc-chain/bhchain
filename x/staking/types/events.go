package types

// staking module event types
const (
	EventTypeCompleteUnbonding    = "complete_unbonding"
	EventTypeCompleteRedelegation = "complete_redelegation"
	EventTypeCreateValidator      = "create_validator"
	EventTypeEditValidator        = "edit_validator"
	EventTypeDelegate             = "delegate"
	EventTypeUnbond               = "unbond"
	EventTypeRedelegate           = "redelegate"
	EventTypeMigrationBegin       = "migration_beign"

	AttributeKeyValidator           = "validator"
	AttributeKeyCommissionRate      = "commission_rate"
	AttributeKeySrcValidator        = "source_validator"
	AttributeKeyDstValidator        = "destination_validator"
	AttributeKeyDelegator           = "delegator"
	AttributeKeyCompletionTime      = "completion_time"
	AttributeValueCategory          = ModuleName
	AttributeMigrationNewEpochIndex = "migration_begin_new_epoch_index"
)
