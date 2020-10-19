package types

const (
	EventTypeSyncGasPrice                     = "sync_gas_price"
	EventTypeExecuteAddTokenProposal          = "execute_add_token_proposal"
	EventTypeExecuteTokenParamsChangeProposal = "execute_token_params_change_proposal"
	EventTypeExecuteDisableTokenProposal      = "execute_disable_token_proposal"

	AttributeKeyFrom            = "from"
	AttributeKeyTokeninfo       = "token_info"
	AttributeKeyGasPrice        = "gas_price"
	AttributeKeyToken           = "token"
	AttributeKeyTokenParam      = "param"
	AttributeKeyTokenParamValue = "value"
	AttributeKeyHeight          = "height"

	AttributeValueCategory = ModuleName
)
