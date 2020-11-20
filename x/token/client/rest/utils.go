package rest

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
	"github.com/hbtc-chain/bhchain/x/token/client/cli"
)

type (

	// AddTokenProposalReq defines a add token proposal request body.
	AddTokenProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title       string        `json:"title" yaml:"title"`
		Description string        `json:"description" yaml:"description"`
		TokenInfo   *sdk.IBCToken `json:"token_info" yaml:"token_info"`
		Deposit     sdk.Coins     `json:"deposit" yaml:"deposit"`
		Proposer    sdk.CUAddress `json:"proposer" yaml:"proposer"`
	}

	// TokenParamsChangeProposalReq defines a token params change request body.
	TokenParamsChangeProposalReq struct {
		BaseReq     rest.BaseReq         `json:"base_req" yaml:"base_req"`
		Title       string               `json:"title" yaml:"title"`
		Description string               `json:"description" yaml:"description"`
		Symbol      string               `json:"symbol" yaml:"symbol"`
		Changes     cli.ParamChangesJSON `json:"changes" yaml:"changes"`
		Deposit     sdk.Coins            `json:"deposit" yaml:"deposit"`
		Proposer    sdk.CUAddress        `json:"proposer" yaml:"proposer"`
	}
)
