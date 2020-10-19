package exported

import (
	sdk "github.com/hbtc-chain/bhchain/types"

	tmbytes "github.com/tendermint/tendermint/libs/common"
)

// Evidence defines the contract which concrete evidence types of misbehavior
// must implement.
type Evidence interface {
	Route() string
	Type() string
	String() string
	Hash() tmbytes.HexBytes
	ValidateBasic() sdk.Error

	// Height at which the infraction occurred
	GetHeight() int64
}

// ValidatorEvidence extends Evidence interface to define contract
// for evidence against malicious validators
type ValidatorEvidence interface {
	Evidence

	// The consensus address of the malicious validator at time of infraction
	GetConsensusAddress() sdk.ConsAddress

	// The total power of the malicious validator at time of infraction
	GetValidatorPower() int64

	// The total validator set power at time of infraction
	GetTotalPower() int64
}

// MsgSubmitEvidence defines the specific interface a concrete message must
// implement in order to process submitted evidence. The concrete MsgSubmitEvidence
// must be defined at the application-level.
type MsgSubmitEvidence interface {
	sdk.Msg

	GetEvidence() Evidence
	GetSubmitter() sdk.CUAddress
}

type Vote interface{}

type VoteItem struct {
	Vote  Vote
	Voter sdk.CUAddress
}

type VoteBox interface {
	// AddVote adds a vote and returns whether it's confirmed for the first time.
	AddVote(sdk.CUAddress, Vote) bool
	// ValidVotes returns the VoteItem list which confirms this vote.
	ValidVotes() []*VoteItem
	HasConfirmed() bool
}

// NewVoteBox creates a new instance of VoteBox with confirm threshold.
type NewVoteBox func(int) VoteBox
