package types

import (
	"github.com/hbtc-chain/bhchain/x/evidence"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type TxVote struct {
	CostFee int64 `json:"cost_fee"`
	Valid   bool  `json:"valid"`
}

func NewTxVote(costFee int64, valid bool) *TxVote {
	return &TxVote{
		CostFee: costFee,
		Valid:   valid,
	}
}

type OrderRetryVoteItem struct {
	Evidences []EvidenceValidator `json:"evidences"`
	Voter     sdk.CUAddress       `json:"voter"`
}

var _ evidence.VoteBox = (*OrderRetryVoteBox)(nil)

type OrderRetryVoteBox struct {
	ConfirmThreshold int                   `json:"confirm_threshold"`
	VoteItems        []*OrderRetryVoteItem `json:"vote_items"`
	Confirmed        bool                  `json:"confirmed"`
}

func NewOrderRetryVoteBox(confirmThreshold int) evidence.VoteBox {
	return &OrderRetryVoteBox{
		ConfirmThreshold: confirmThreshold,
	}
}

func (v *OrderRetryVoteBox) AddVote(voter sdk.CUAddress, vote evidence.Vote) bool {
	evidences, ok := vote.([]EvidenceValidator)
	if !ok {
		return false
	}
	var hasVoted bool
	for i, item := range v.VoteItems {
		if item.Voter.Equals(voter) {
			hasVoted = true
			// 没有 confirm 之前可以更改投票
			if !v.Confirmed {
				v.VoteItems[i].Evidences = evidences
			}
			break
		}
	}
	if !hasVoted {
		v.VoteItems = append(v.VoteItems, &OrderRetryVoteItem{Evidences: evidences, Voter: voter})
	}
	if v.Confirmed || len(v.VoteItems) < v.ConfirmThreshold {
		return false
	}

	v.Confirmed = true
	return true
}

func (v *OrderRetryVoteBox) ValidVotes() []*evidence.VoteItem {
	if !v.Confirmed {
		return nil
	}
	var validVotes []*evidence.VoteItem
	for _, item := range v.VoteItems {
		validVotes = append(validVotes, &evidence.VoteItem{Voter: item.Voter, Vote: item.Evidences})
	}
	return validVotes
}

func (v *OrderRetryVoteBox) HasConfirmed() bool {
	return v.Confirmed
}
