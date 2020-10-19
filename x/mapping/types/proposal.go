package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
)

const (
	ProposalTypeAddMapping    = "AddMapping"
	ProposalTypeSwitchMapping = "SwitchMapping"
)

var _ govtypes.Content = AddMappingProposal{}
var _ govtypes.Content = SwitchMappingProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMapping)
	govtypes.RegisterProposalTypeCodec(AddMappingProposal{}, "hbtcchain/AddMappingProposal")
	govtypes.RegisterProposalType(ProposalTypeSwitchMapping)
	govtypes.RegisterProposalTypeCodec(SwitchMappingProposal{}, "hbtcchain/SwitchMappingProposal")
}

type AddMappingProposal struct {
	From         string     `json:"from"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	IssueSymbol  sdk.Symbol `json:"issue_symbol"`
	TargetSymbol sdk.Symbol `json:"target_symbol"`
	TotalSupply  sdk.Int    `json:"total_supply"`
}

func NewAddMappingProposal(from, title, desc string, issueSymbol, targetSymbol sdk.Symbol, totalSupply sdk.Int) AddMappingProposal {
	return AddMappingProposal{
		From:         from,
		Title:        title,
		Description:  desc,
		IssueSymbol:  issueSymbol,
		TargetSymbol: targetSymbol,
		TotalSupply:  totalSupply,
	}
}
func (amp AddMappingProposal) GetTitle() string { return amp.Title }

func (amp AddMappingProposal) GetDescription() string { return amp.Description }

func (amp AddMappingProposal) ProposalRoute() string { return RouterKey }

func (amp AddMappingProposal) ProposalToken() string { return sdk.NativeToken }

func (amp AddMappingProposal) ProposalType() string { return ProposalTypeAddMapping }

func (amp AddMappingProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, amp)
	if err != nil {
		return err
	}
	if amp.From == "" || !sdk.IsValidAddr(amp.From) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("from address can not be empty or invalid:%v", amp.From))
	}
	if !amp.IssueSymbol.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid issue symbol")
	}
	if !amp.TargetSymbol.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid target symbol")
	}
	if !amp.TotalSupply.IsPositive() {
		return sdk.ErrInvalidAmount("total supply should be positive")
	}
	return nil
}

// String implements the Stringer interface.
func (amp AddMappingProposal) String() string {
	return fmt.Sprintf(`Add Mapping Proposal:
  From:		   %s
  Title:       %s
  Description: %s
  IssueSymbol: %s
  TargetSymbol:%s
  TotalSupply: %s
`,
		amp.From, amp.Title, amp.Description, amp.IssueSymbol.String(), amp.TargetSymbol.String(), amp.TotalSupply.String())
}

type SwitchMappingProposal struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IssueSymbol sdk.Symbol `json:"issue_symbol"`
	Enable      bool       `json:"enable"`
}

func NewSwitchMappingProposal(title, desc string, issueSymbol sdk.Symbol, enable bool) SwitchMappingProposal {
	return SwitchMappingProposal{
		Title:       title,
		Description: desc,
		IssueSymbol: issueSymbol,
		Enable:      enable,
	}
}
func (smp SwitchMappingProposal) GetTitle() string { return smp.Title }

func (smp SwitchMappingProposal) GetDescription() string { return smp.Description }

func (smp SwitchMappingProposal) ProposalRoute() string { return RouterKey }

func (smp SwitchMappingProposal) ProposalToken() string { return sdk.NativeToken }

func (smp SwitchMappingProposal) ProposalType() string { return ProposalTypeSwitchMapping }

// ValidateBasic runs basic stateless validity checks
func (smp SwitchMappingProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, smp)
	if err != nil {
		return err
	}
	if !smp.IssueSymbol.IsValidTokenName() {
		return sdk.ErrInvalidSymbol("invalid issue symbol")
	}
	return nil
}

// String implements the Stringer interface.
func (smp SwitchMappingProposal) String() string {
	return fmt.Sprintf(`Switch Mapping Proposal:
  Title:       %s
  Description: %s
  IssueSymbol: %s
  Enable: 	   %v
`,
		smp.Title, smp.Description, smp.IssueSymbol.String(), smp.Enable)
}
