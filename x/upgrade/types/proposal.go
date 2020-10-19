package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	gov "github.com/hbtc-chain/bhchain/x/gov/types"
)

const (
	DefaultCodespace                  sdk.CodespaceType = "upgrade"
	ProposalTypeSoftwareUpgrade       string            = "SoftwareUpgrade"
	ProposalTypeCancelSoftwareUpgrade string            = "CancelSoftwareUpgrade"
)

type SoftwareUpgradeProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Plan        Plan   `json:"plan"`
}

func NewSoftwareUpgradeProposal(title, description string, plan Plan) gov.Content {
	return &SoftwareUpgradeProposal{title, description, plan}
}

// Implements Proposal Interface
var _ gov.Content = &SoftwareUpgradeProposal{}

func init() {
	gov.RegisterProposalType(ProposalTypeSoftwareUpgrade)
	gov.RegisterProposalTypeCodec(&SoftwareUpgradeProposal{}, "hbtcchain/SoftwareUpgradeProposal")
	gov.RegisterProposalType(ProposalTypeCancelSoftwareUpgrade)
	gov.RegisterProposalTypeCodec(&CancelSoftwareUpgradeProposal{}, "hbtcchain/CancelSoftwareUpgradeProposal")
}

func (sup *SoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup *SoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup *SoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup *SoftwareUpgradeProposal) ProposalToken() string  { return sdk.NativeToken }
func (sup *SoftwareUpgradeProposal) ProposalType() string   { return ProposalTypeSoftwareUpgrade }
func (sup *SoftwareUpgradeProposal) ValidateBasic() sdk.Error {
	if err := sup.Plan.ValidateBasic(); err != nil {
		return err
	}
	return gov.ValidateAbstract(DefaultCodespace, sup)
}

func (sup SoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, sup.Title, sup.Description)
}

type CancelSoftwareUpgradeProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewCancelSoftwareUpgradeProposal(title, description string) gov.Content {
	return &CancelSoftwareUpgradeProposal{title, description}
}

// Implements Proposal Interface
var _ gov.Content = &CancelSoftwareUpgradeProposal{}

func (sup *CancelSoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup *CancelSoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup *CancelSoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup *CancelSoftwareUpgradeProposal) ProposalToken() string  { return sdk.NativeToken }
func (sup *CancelSoftwareUpgradeProposal) ProposalType() string {
	return ProposalTypeCancelSoftwareUpgrade
}
func (sup *CancelSoftwareUpgradeProposal) ValidateBasic() sdk.Error {
	return gov.ValidateAbstract(DefaultCodespace, sup)
}

func (sup CancelSoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Cancel Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, sup.Title, sup.Description)
}
