package client

import (
	govclient "github.com/hbtc-chain/bhchain/x/gov/client"
	"github.com/hbtc-chain/bhchain/x/upgrade/client/cli"
	"github.com/hbtc-chain/bhchain/x/upgrade/client/rest"
)

var PostProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpgradeProposal, rest.PostPlanProposalRESTHandler)
var CancelProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitCancelUpgradeProposal, rest.CancelPlanProposalRESTHandler)
