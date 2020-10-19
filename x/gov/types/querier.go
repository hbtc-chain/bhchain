package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

// query endpoints supported by the governance Querier
const (
	QueryParams    = "params"
	QueryProposals = "proposals"
	QueryProposal  = "proposal"
	QueryDeposits  = "deposits"
	QueryDeposit   = "deposit"
	QueryVotes     = "votes"
	QueryVote      = "vote"
	QueryTally     = "tally"

	ParamDeposit  = "deposit"
	ParamVoting   = "voting"
	ParamTallying = "tallying"
)

// Params for queries:
// - 'custom/gov/proposal'
// - 'custom/gov/deposits'
// - 'custom/gov/tally'
// - 'custom/gov/votes'
type QueryProposalParams struct {
	ProposalID uint64
}

// creates a new instance of QueryProposalParams
func NewQueryProposalParams(proposalID uint64) QueryProposalParams {
	return QueryProposalParams{
		ProposalID: proposalID,
	}
}

// Params for query 'custom/gov/deposit'
type QueryDepositParams struct {
	ProposalID uint64
	Depositor  sdk.CUAddress
}

// creates a new instance of QueryDepositParams
func NewQueryDepositParams(proposalID uint64, depositor sdk.CUAddress) QueryDepositParams {
	return QueryDepositParams{
		ProposalID: proposalID,
		Depositor:  depositor,
	}
}

// Params for query 'custom/gov/vote'
type QueryVoteParams struct {
	ProposalID uint64
	Voter      sdk.CUAddress
}

// creates a new instance of QueryVoteParams
func NewQueryVoteParams(proposalID uint64, voter sdk.CUAddress) QueryVoteParams {
	return QueryVoteParams{
		ProposalID: proposalID,
		Voter:      voter,
	}
}

// Params for query 'custom/gov/proposals'
type QueryProposalsParams struct {
	Voter          sdk.CUAddress
	Depositor      sdk.CUAddress
	ProposalStatus ProposalStatus
	Limit          uint64
}

// creates a new instance of QueryProposalsParams
func NewQueryProposalsParams(status ProposalStatus, limit uint64, voter, depositor sdk.CUAddress) QueryProposalsParams {
	return QueryProposalsParams{
		Voter:          voter,
		Depositor:      depositor,
		ProposalStatus: status,
		Limit:          limit,
	}
}
