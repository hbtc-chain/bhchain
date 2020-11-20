package types

import (
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
)

const (
	ProposalUpdateKeyNodes = "UpdateKeyNodes"
)

var _ govtypes.Content = &UpdateKeyNodesProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalUpdateKeyNodes)
	govtypes.RegisterProposalTypeCodec(&UpdateKeyNodesProposal{}, "hbtcchain/UpdateKeyNodesProposal")
}

type UpdateKeyNodesProposal struct {
	Title          string          `json:"title" yaml:"title"`
	Description    string          `json:"description" yaml:"description"`
	AddKeyNodes    []sdk.CUAddress `json:"add_key_nodes,omitempty" yaml:"add_key_nodes"`
	RemoveKeyNodes []sdk.CUAddress `json:"remove_key_nodes,omitempty" yaml:"remove_key_nodes"`
}

func NewUpdateKeyNodesProposal(title, description string, addKeyNodes, removeKeyNodes []sdk.CUAddress) *UpdateKeyNodesProposal {
	return &UpdateKeyNodesProposal{
		Title:          title,
		Description:    description,
		AddKeyNodes:    addKeyNodes,
		RemoveKeyNodes: removeKeyNodes,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (p *UpdateKeyNodesProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of a community pool spend proposal.
func (p *UpdateKeyNodesProposal) GetDescription() string { return p.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (p *UpdateKeyNodesProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (p *UpdateKeyNodesProposal) ProposalType() string { return ProposalUpdateKeyNodes }

func (p *UpdateKeyNodesProposal) ProposalToken() string { return sdk.NativeToken }

// ValidateBasic runs basic stateless validity checks
func (p *UpdateKeyNodesProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, p)
	if err != nil {
		return err
	}
	if len(p.AddKeyNodes) == 0 && len(p.RemoveKeyNodes) == 0 {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	exists := make(map[string]bool)
	for _, cu := range p.AddKeyNodes {
		if !cu.IsValidAddr() {
			return ErrBadValidatorAddr(DefaultCodespace)
		}
		if exists[cu.String()] {
			return ErrDuplicatedValidatorAddr(DefaultCodespace)
		}
		exists[cu.String()] = true
	}
	for _, cu := range p.RemoveKeyNodes {
		if !cu.IsValidAddr() {
			return ErrBadValidatorAddr(DefaultCodespace)
		}
		if exists[cu.String()] {
			return ErrDuplicatedValidatorAddr(DefaultCodespace)
		}
		exists[cu.String()] = true
	}
	return nil
}

// String implements the Stringer interface.
func (p *UpdateKeyNodesProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update KeyNodes Proposal:
  Title:       %s
  Description: %s
  AddKeyNodes: %v
  RemoveKeyNodes: %v
`, p.Title, p.Description, p.AddKeyNodes, p.RemoveKeyNodes))
	return b.String()
}
