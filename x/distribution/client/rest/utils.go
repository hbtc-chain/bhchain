package rest

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
)

type (
	// CommunityPoolSpendProposalReq defines a community pool spend proposal request body.
	CommunityPoolSpendProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title       string        `json:"title" yaml:"title"`
		Description string        `json:"description" yaml:"description"`
		Recipient   sdk.CUAddress `json:"recipient" yaml:"recipient"`
		Amount      sdk.Coins     `json:"amount" yaml:"amount"`
		Proposer    sdk.CUAddress `json:"proposer" yaml:"proposer"`
		Deposit     sdk.Coins     `json:"deposit" yaml:"deposit"`
		VoteTime    uint32        `json:"votetime" yaml:"votetime"`
	}
)
