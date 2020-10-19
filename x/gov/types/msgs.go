package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgDaoVote        = "dao_vote"
	TypeMsgCancelDaoVote  = "cancel_dao_vote"
	TypeMsgSubmitProposal = "submit_proposal"
)

var _, _, _, _, _ sdk.Msg = MsgSubmitProposal{}, MsgDeposit{}, MsgVote{}, MsgDaoVote{}, MsgCancelDaoVote{}

// MsgSubmitProposal
type MsgSubmitProposal struct {
	Content        Content       `json:"content" yaml:"content"`
	InitialDeposit sdk.Coins     `json:"initial_deposit" yaml:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive
	Proposer       sdk.CUAddress `json:"proposer" yaml:"proposer"`               //  Address of the proposer
	VoteTime       uint32        `json:"vote_time" yaml:"vote_time"`
}

func NewMsgSubmitProposal(content Content, initialDeposit sdk.Coins, proposer sdk.CUAddress, voteTime uint32) MsgSubmitProposal {
	return MsgSubmitProposal{content, initialDeposit, proposer, voteTime}
}

//nolint
func (msg MsgSubmitProposal) Route() string { return RouterKey }
func (msg MsgSubmitProposal) Type() string  { return TypeMsgSubmitProposal }

// Implements Msg.
func (msg MsgSubmitProposal) ValidateBasic() sdk.Error {
	if msg.Content == nil {
		return ErrInvalidProposalContent(DefaultCodespace, "missing content")
	}
	if msg.Proposer.Empty() {
		return sdk.ErrInvalidAddress(msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if msg.InitialDeposit.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if !IsValidProposalType(msg.Content.ProposalType()) {
		return ErrInvalidProposalType(DefaultCodespace, msg.Content.ProposalType())
	}

	return msg.Content.ValidateBasic()
}

func (msg MsgSubmitProposal) String() string {
	return fmt.Sprintf(`Submit Proposal Message:
  Content:         %s
  Initial Deposit: %s
`, msg.Content.String(), msg.InitialDeposit)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.Proposer}
}

// MsgDeposit
type MsgDeposit struct {
	ProposalID uint64        `json:"proposal_id" yaml:"proposal_id"` // ID of the proposal
	Depositor  sdk.CUAddress `json:"depositor" yaml:"depositor"`     // Address of the depositor
	Amount     sdk.Coins     `json:"amount" yaml:"amount"`           // Coins to add to the proposal's deposit
}

func NewMsgDeposit(depositor sdk.CUAddress, proposalID uint64, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{proposalID, depositor, amount}
}

// Implements Msg.
// nolint
func (msg MsgDeposit) Route() string { return RouterKey }
func (msg MsgDeposit) Type() string  { return TypeMsgDeposit }

// Implements Msg.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.Amount.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

func (msg MsgDeposit) String() string {
	return fmt.Sprintf(`Deposit Message:
  Depositer:   %s
  Proposal ID: %d
  Amount:      %s
`, msg.Depositor, msg.ProposalID, msg.Amount)
}

// Implements Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgDeposit) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.Depositor}
}

// MsgVote
type MsgVote struct {
	ProposalID uint64        `json:"proposal_id" yaml:"proposal_id"` // ID of the proposal
	Voter      sdk.CUAddress `json:"voter" yaml:"voter"`             //  address of the voter
	Option     VoteOption    `json:"option" yaml:"option"`           //  option from OptionSet chosen by the voter
}

func NewMsgVote(voter sdk.CUAddress, proposalID uint64, option VoteOption) MsgVote {
	return MsgVote{proposalID, voter, option}
}

// Implements Msg.
// nolint
func (msg MsgVote) Route() string { return RouterKey }
func (msg MsgVote) Type() string  { return TypeMsgVote }

// Implements Msg.
func (msg MsgVote) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	if !ValidVoteOption(msg.Option) {
		return ErrInvalidVote(DefaultCodespace, msg.Option)
	}

	return nil
}

func (msg MsgVote) String() string {
	return fmt.Sprintf(`Vote Message:
  Proposal ID: %d
  Option:      %s
`, msg.ProposalID, msg.Option)
}

// Implements Msg.
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgVote) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.Voter}
}

// MsgVote
type MsgDaoVote struct {
	ProposalID uint64        `json:"proposal_id" yaml:"proposal_id"` // ID of the proposal
	Voter      sdk.CUAddress `json:"voter" yaml:"voter"`             //  address of the voter
	VoteAmount sdk.Coins     `json:"vote_amount" yaml:"vote_amount"` //  option amount by the voter
	Option     VoteOption    `json:"option" yaml:"option"`           //  option from OptionSet chosen by the voter
}

func NewMsgDaoVote(voter sdk.CUAddress, proposalID uint64, voteAmount sdk.Coins, option VoteOption) MsgDaoVote {
	return MsgDaoVote{proposalID, voter, voteAmount, option}
}

// Implements Msg.
// nolint
func (msg MsgDaoVote) Route() string { return RouterKey }
func (msg MsgDaoVote) Type() string  { return TypeMsgDaoVote }

// Implements Msg.
func (msg MsgDaoVote) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	if msg.VoteAmount.IsAnyNegative() {
		return sdk.ErrInvalidAmount(msg.VoteAmount.String())
	}

	if !ValidVoteOption(msg.Option) {
		return ErrInvalidVote(DefaultCodespace, msg.Option)
	}

	return nil
}

func (msg MsgDaoVote) String() string {
	return fmt.Sprintf(`Vote Message:
  Proposal ID: %d
  VoteAmount:  %s
  Option:      %s
`, msg.ProposalID, msg.VoteAmount, msg.Option)
}

// Implements Msg.
func (msg MsgDaoVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgDaoVote) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.Voter}
}

// MsgCacelDaoVote
type MsgCancelDaoVote struct {
	ProposalID uint64        `json:"proposal_id" yaml:"proposal_id"` // ID of the proposal
	Voter      sdk.CUAddress `json:"voter" yaml:"voter"`             //  address of the voter
}

func NewMsgCancelDaoVote(voter sdk.CUAddress, proposalID uint64) MsgCancelDaoVote {
	return MsgCancelDaoVote{proposalID, voter}
}

// Implements Msg.
// nolint
func (msg MsgCancelDaoVote) Route() string { return RouterKey }
func (msg MsgCancelDaoVote) Type() string  { return TypeMsgCancelDaoVote }

// Implements Msg.
func (msg MsgCancelDaoVote) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}

	return nil
}

func (msg MsgCancelDaoVote) String() string {
	return fmt.Sprintf(`Vote Message:
  Proposal ID: %d
`, msg.ProposalID)
}

// Implements Msg.
func (msg MsgCancelDaoVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgCancelDaoVote) GetSigners() []sdk.CUAddress {
	return []sdk.CUAddress{msg.Voter}
}
