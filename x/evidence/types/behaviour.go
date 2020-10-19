package types

var (
	VoteBehaviourKey  = "Vote"
	DsignBehaviourKey = "dsign"
)

var AllBehaviourKeys = []string{
	VoteBehaviourKey,
	DsignBehaviourKey,
}

// Signing info for a validator
type ValidatorBehaviour struct {
	IndexOffset         int64 `json:"index_offset" yaml:"index_offset"`                   // index offset into signed block bit array
	MisbehaviourCounter int64 `json:"missed_blocks_counter" yaml:"missed_blocks_counter"` // missed blocks counter (to avoid scanning the array every time)
}
