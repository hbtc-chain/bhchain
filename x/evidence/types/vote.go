package types

import (
	"bytes"
	"encoding/json"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
)

const (
	BehaviourCountDelayBlock uint64 = 10
)

var _ exported.VoteBox = (*VoteBox)(nil)

type BoolVote bool

// VoteBox is the default implement of exported.VoteBox interface.
// It's confirmed when over 2/3 validators vote to the same results.
type VoteBox struct {
	ConfirmThreshold int                  `json:"confirm_threshold"`
	VoteItems        []*exported.VoteItem `json:"vote_items"`
	Confirmed        bool                 `json:"confirmed"`
	Result           exported.Vote        `json:"result"`
}

func NewVoteBox(confirmThreshold int) exported.VoteBox {
	return &VoteBox{
		ConfirmThreshold: confirmThreshold,
	}
}

func (v *VoteBox) AddVote(voter sdk.CUAddress, vote exported.Vote) bool {
	var hasVoted bool
	for i, item := range v.VoteItems {
		if item.Voter.Equals(voter) {
			hasVoted = true
			// 没有 confirm 之前可以更改投票
			if !v.Confirmed {
				v.VoteItems[i].Vote = vote
			}
			break
		}
	}
	if !hasVoted {
		v.VoteItems = append(v.VoteItems, &exported.VoteItem{Vote: vote, Voter: voter})
	}
	if v.Confirmed || len(v.VoteItems) < v.ConfirmThreshold {
		return false
	}
	counter := make(map[string]int)
	for _, item := range v.VoteItems {
		bz, _ := json.Marshal(item.Vote)
		str := string(bz)
		counter[str] = counter[str] + 1
		if counter[str] >= v.ConfirmThreshold {
			v.Confirmed = true
			v.Result = item.Vote
			return true
		}
	}
	return false
}

func (v *VoteBox) ValidVotes() []*exported.VoteItem {
	if !v.Confirmed {
		return nil
	}
	result, _ := json.Marshal(v.Result)
	var validVotes []*exported.VoteItem
	for _, item := range v.VoteItems {
		bz, _ := json.Marshal(item.Vote)
		if bytes.Equal(result, bz) {
			validVotes = append(validVotes, item)
		}
	}
	return validVotes
}

func (v *VoteBox) HasConfirmed() bool {
	return v.Confirmed
}
