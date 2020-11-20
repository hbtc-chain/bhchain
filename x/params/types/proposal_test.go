package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParameterChangeProposal(t *testing.T) {
	pc1 := NewParamChange("sub", "foo", "baz")
	pc2 := NewParamChangeWithSubkey("sub", "bar", "cat", "dog")
	pcp := NewParameterChangeProposal("test title", "test description", []ParamChange{pc1, pc2})

	require.Equal(t, "test title", pcp.GetTitle())
	require.Equal(t, "test description", pcp.GetDescription())
	require.Equal(t, RouterKey, pcp.ProposalRoute())
	require.Equal(t, ProposalTypeChange, pcp.ProposalType())
	require.Nil(t, pcp.ValidateBasic())

	pc3 := NewParamChangeWithSubkey("", "bar", "cat", "dog")
	pcp = NewParameterChangeProposal("test title", "test description", []ParamChange{pc3})
	require.Error(t, pcp.ValidateBasic())

	pc4 := NewParamChangeWithSubkey("sub", "", "cat", "dog")
	pcp = NewParameterChangeProposal("test title", "test description", []ParamChange{pc4})
	require.Error(t, pcp.ValidateBasic())

	pc5 := NewParamChangeWithSubkey("sub", "foo", "cat", "")
	pcp = NewParameterChangeProposal("test title", "test description", []ParamChange{pc5})
	require.Error(t, pcp.ValidateBasic())
}

type validator struct {
}

func (v *validator) Validate(change ParamChange) error {
	if change.Subspace == "sub" && change.Key == "foo" {
		return fmt.Errorf("foo invalid")
	}
	return nil
}

func TestParameterChangeProposalValidator(t *testing.T) {
	RegisterParameterChangeValidator(&validator{})
	pc1 := NewParamChange("sub", "foo", "baz")
	pc2 := NewParamChange("sub", "cat", "baz")
	pc3 := NewParamChange("baz", "foo", "baz")
	pcp1 := NewParameterChangeProposal("sub", "foo", []ParamChange{pc1})
	pcp2 := NewParameterChangeProposal("sub", "foo", []ParamChange{pc2})
	pcp3 := NewParameterChangeProposal("sub", "foo", []ParamChange{pc3})
	pcp4 := NewParameterChangeProposal("sub", "foo", []ParamChange{pc1, pc2, pc3})

	require.Error(t, pcp1.ValidateBasic())
	require.Nil(t, pcp2.ValidateBasic())
	require.Nil(t, pcp3.ValidateBasic())
	require.Error(t, pcp4.ValidateBasic())
}
