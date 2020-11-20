package staking

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

func isInCUList(addr sdk.CUAddress, cuList []sdk.CUAddress) bool {
	for _, cu := range cuList {
		if cu.Equals(addr) {
			return true
		}
	}
	return false
}

func handleUpdateKeyNodesProposal(ctx sdk.Context, keeper Keeper, proposal *types.UpdateKeyNodesProposal) sdk.Result {
	ctx.Logger().Info("handleUpdateKeyNodesProposal", "proposal", proposal)

	if !keeper.IsMigrationFinished(ctx) {
		return ErrBadUpdateKeyNodesTime(keeper.Codespace()).Result()
	}
	curKeyNodes := keeper.GetCurrentEpoch(ctx).KeyNodeSet
	maxRemoveNum := sdk.OneSixthCeil(len(curKeyNodes))
	if len(proposal.RemoveKeyNodes) > maxRemoveNum || len(proposal.RemoveKeyNodes) >= len(curKeyNodes) {
		return ErrRemoveTooManyKeyNodes(keeper.Codespace()).Result()
	}

	for _, addr := range proposal.RemoveKeyNodes {
		if !isInCUList(addr, curKeyNodes) {
			return ErrRemoveNotKeyNode(keeper.Codespace()).Result()
		}
		for i, curKeyNode := range curKeyNodes {
			if curKeyNode.Equals(addr) {
				curKeyNodes = append(curKeyNodes[:i], curKeyNodes[i+1:]...)
				break
			}
		}
	}

	stakingParams := keeper.GetParams(ctx)
	blkHeight := uint64(ctx.BlockHeight())
	for _, addr := range proposal.AddKeyNodes {
		if isInCUList(addr, curKeyNodes) {
			return ErrAddDuplicatedKeyNode(keeper.Codespace()).Result()
		}
		validator, found := keeper.GetValidator(ctx, sdk.ValAddress(addr))
		if !found {
			return ErrNoValidatorFound(keeper.Codespace()).Result()
		}
		if !validator.CanBeKeyNode(stakingParams.MinKeyNodeDelegation) ||
			validator.LastKeyNodeHeartbeatHeight == 0 ||
			blkHeight-validator.LastKeyNodeHeartbeatHeight > stakingParams.MaxCandidateKeyNodeHeartbeatInterval {
			return ErrNoQualification(keeper.Codespace()).Result()
		}
		if len(curKeyNodes) >= int(stakingParams.MaxKeyNodes) {
			return ErrKeyNodeNumExceeds(keeper.Codespace()).Result()
		}
		curKeyNodes = append(curKeyNodes, addr)
	}
	epoch := keeper.StartNewEpoch(ctx, curKeyNodes)
	//set allopcu in migration // delete prekeygen order
	keeper.AfterNewEpoch(ctx, epoch)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMigrationBegin,
			sdk.NewAttribute(types.AttributeMigrationNewEpochIndex, fmt.Sprintf("%d", epoch.Index)),
		))

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func NewStakingProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Result {
		switch c := content.(type) {
		case *types.UpdateKeyNodesProposal:
			return handleUpdateKeyNodesProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized staking proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
