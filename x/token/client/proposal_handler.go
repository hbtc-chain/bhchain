package client

import (
	govclient "github.com/hbtc-chain/bhchain/x/gov/client"
	"github.com/hbtc-chain/bhchain/x/token/client/cli"
	"github.com/hbtc-chain/bhchain/x/token/client/rest"
)

// param change proposal handler
var (
	AddTokenProposalHandler          = govclient.NewProposalHandler(cli.GetCmdAddTokenProposal, rest.AddTokenProposalRESTHandler)
	TokenParamsChangeProposalHandler = govclient.NewProposalHandler(cli.GetCmdTokenParamsChangeProposal, rest.TokenParamsChangeProposalRESTHandler)
)
