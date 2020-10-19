package upgrade

import (
	"fmt"
	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/upgrade/keeper"
	"github.com/hbtc-chain/bhchain/x/upgrade/types"
)

// NewSoftwareUpgradeProposalHandler creates a governance handler to manage new proposal types.
// It enables SoftwareUpgradeProposal to propose an Upgrade, and CancelSoftwareUpgradeProposal
// to abort a previously voted upgrade.
func NewSoftwareUpgradeProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Result {
		switch c := content.(type) {
		case *types.SoftwareUpgradeProposal:
			return handleSoftwareUpgradeProposal(ctx, k, c)

		case *types.CancelSoftwareUpgradeProposal:
			return handleCancelSoftwareUpgradeProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized upgrade proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleSoftwareUpgradeProposal(ctx sdk.Context, k keeper.Keeper, p *types.SoftwareUpgradeProposal) sdk.Result {
	return k.ScheduleUpgrade(ctx, p.Plan)
}

func handleCancelSoftwareUpgradeProposal(ctx sdk.Context, k keeper.Keeper, _ *types.CancelSoftwareUpgradeProposal) sdk.Result {
	k.ClearUpgradePlan(ctx)
	return sdk.Result{}
}
