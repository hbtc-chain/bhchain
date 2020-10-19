package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence"
)

var (
	priceUpLimitRatio  = sdk.NewDecWithPrec(12, 1) //gas price uplimit 1.2x
	priceLowLimitRatio = sdk.NewDecWithPrec(9, 1)  //gas price lowlimit 0.9x
)

type GasPriceVoteItem struct {
	Price sdk.Int       `json:"price"`
	Voter sdk.CUAddress `json:"voter"`
}

var _ evidence.VoteBox = (*GasPriceVoteBox)(nil)

type GasPriceVoteBox struct {
	ConfirmThreshold int                 `json:"confirm_threshold"`
	VoteItems        []*GasPriceVoteItem `json:"vote_items"`
	Confirmed        bool                `json:"confirmed"`
	ConfirmedMedian  sdk.Int             `json:"confirmed_median"`
}

func NewGasPriceVoteBox(confirmThreshold int) evidence.VoteBox {
	return &GasPriceVoteBox{
		ConfirmThreshold: confirmThreshold,
	}
}

func (v *GasPriceVoteBox) AddVote(voter sdk.CUAddress, vote evidence.Vote) bool {
	price, ok := vote.(sdk.Int)
	if !ok {
		return false
	}
	var hasVoted bool
	for i, item := range v.VoteItems {
		if item.Voter.Equals(voter) {
			hasVoted = true
			// 没有 confirm 之前可以更改投票
			if !v.Confirmed {
				v.VoteItems[i].Price = price
			}
			break
		}
	}
	if !hasVoted {
		v.VoteItems = append(v.VoteItems, &GasPriceVoteItem{Price: price, Voter: voter})
	}
	if v.Confirmed || len(v.VoteItems) < v.ConfirmThreshold {
		return false
	}
	gasPrices := make([]sdk.Int, 0)
	for _, item := range v.VoteItems {
		gasPrices = append(gasPrices, item.Price)
	}
	medianGasPrice := sdk.IntMedian(gasPrices)
	priceUpLimit := sdk.NewDecFromInt(medianGasPrice).Mul(priceUpLimitRatio)
	priceLowLimit := sdk.NewDecFromInt(medianGasPrice).Mul(priceLowLimitRatio)

	newGasPrices := gasPrices[:0]
	for _, price := range gasPrices {
		gasPrice := sdk.NewDecFromInt(price)
		if gasPrice.LTE(priceUpLimit) && gasPrice.GTE(priceLowLimit) {
			newGasPrices = append(newGasPrices, price)
		}
	}
	if len(newGasPrices) < v.ConfirmThreshold {
		return false
	}
	v.Confirmed = true
	v.ConfirmedMedian = medianGasPrice
	return true
}

func (v *GasPriceVoteBox) ValidVotes() []*evidence.VoteItem {
	if !v.Confirmed {
		return nil
	}
	priceUpLimit := sdk.NewDecFromInt(v.ConfirmedMedian).Mul(priceUpLimitRatio)
	priceLowLimit := sdk.NewDecFromInt(v.ConfirmedMedian).Mul(priceLowLimitRatio)
	var validVotes []*evidence.VoteItem
	for _, item := range v.VoteItems {
		gasPrice := sdk.NewDecFromInt(item.Price)
		if gasPrice.LTE(priceUpLimit) && gasPrice.GTE(priceLowLimit) {
			validVotes = append(validVotes, &evidence.VoteItem{Voter: item.Voter, Vote: item.Price})
		}
	}
	return validVotes
}

func (v *GasPriceVoteBox) HasConfirmed() bool {
	return v.Confirmed
}
