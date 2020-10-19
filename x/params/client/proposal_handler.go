package client

import (
	govclient "github.com/hbtc-chain/bhchain/x/gov/client"
	"github.com/hbtc-chain/bhchain/x/params/client/cli"
	"github.com/hbtc-chain/bhchain/x/params/client/rest"
)

// param change proposal handler
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
