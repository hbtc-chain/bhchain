package client

import (
	govclient "github.com/hbtc-chain/bhchain/x/gov/client"
	"github.com/hbtc-chain/bhchain/x/staking/client/cli"
)

var UpdateKeyNodesProposalHandler = govclient.NewProposalHandler(cli.NewCmdUpdateKeyNodesProposal, nil)
