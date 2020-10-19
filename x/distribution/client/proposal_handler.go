package client

import (
	"github.com/hbtc-chain/bhchain/x/distribution/client/cli"
	"github.com/hbtc-chain/bhchain/x/distribution/client/rest"
	govclient "github.com/hbtc-chain/bhchain/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
