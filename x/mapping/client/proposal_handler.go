package client

import (
	govclient "github.com/hbtc-chain/bhchain/x/gov/client"
	"github.com/hbtc-chain/bhchain/x/mapping/client/cli"
)

var AddMappingProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitAddMappingProposal, nil)
var SwitchMappingProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitSwitchMappingProposal, nil)
