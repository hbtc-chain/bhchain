package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter store keys
var (
	KeyBehaviourWindow        = []byte("BehaviourWindow")
	KeyMaxMisbehaviourCount   = []byte("MaxMisbehaviourCount")
	KeyBehaviourSlashFraction = []byte("BehaviourSlashFraction")
)

// ParamKeyTable for {{ .NameLowerCase }} module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&BehaviourParams{})
}

// Params - used for initializing default parameter for {{ .NameLowerCase }} at genesis
type BehaviourParams struct {
	BehaviourWindow        int64
	MaxMisbehaviourCount   int64
	BehaviourSlashFraction sdk.Dec
}

func NewBehaviourParams(behaviourWindow int64, maxMisbehaviourCount int64, behaviourSlashFraction sdk.Dec) BehaviourParams {
	return BehaviourParams{BehaviourWindow: behaviourWindow, MaxMisbehaviourCount: maxMisbehaviourCount, BehaviourSlashFraction: behaviourSlashFraction}
}

// String implements the stringer interface for Params
func (p BehaviourParams) String() string {
	return fmt.Sprintf(`Slashing Params:
  BehaviourWindow:          %d
  MaxMisbehaviourCount:     %d
  BehaviourSlashFraction:   %s`, p.BehaviourWindow, p.MaxMisbehaviourCount, p.BehaviourSlashFraction.String())
}

// ParamSetPairs - Implements params.ParamSet
func (p *BehaviourParams) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyBehaviourWindow, &p.BehaviourWindow},
		{KeyMaxMisbehaviourCount, &p.MaxMisbehaviourCount},
		{KeyBehaviourSlashFraction, &p.BehaviourSlashFraction},
	}
}

// DefaultParams defines the parameters for this module
func DefaultBehaviourParams() BehaviourParams {
	return NewBehaviourParams(30000, 20000, sdk.NewDecFromIntWithPrec(sdk.NewInt(1), 2))
}
