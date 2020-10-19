package types

// Minting module event types
const (
	EventTypeMint                                = ModuleName
	EventExecuteInflationParameterChangeProposal = "execute_inflation_parameter_change_proposal"

	AttributeKeyInflation        = "inflation"
	AttributeKeyInitialPrice     = "initial_price"
	AttributeKeyCurrentPrice     = "current_price"
	AttributeKeyNodeCostPerMonth = "node_cost_per_month"
)
