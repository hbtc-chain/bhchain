package types

// GenesisState - all evidence state that must be provided at genesis
type GenesisState struct {
	BehaviourParams map[string]BehaviourParams
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(behaviourParams map[string]BehaviourParams) GenesisState {
	return GenesisState{BehaviourParams: behaviourParams}
}

func DefaultGenesisState() GenesisState {
	params := map[string]BehaviourParams{}
	for _, key := range AllBehaviourKeys {
		params[key] = DefaultBehaviourParams()
	}
	return GenesisState{
		BehaviourParams: params,
	}
}

// ValidateGenesis validates the evidence genesis parameters
func ValidateGenesis(GenesisState) error {
	return nil
}
