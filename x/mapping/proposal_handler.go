package mapping

import (
	"fmt"
	"strconv"

	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/mapping/types"
)

func handleAddMappingProposal(ctx sdk.Context, keeper Keeper, proposal types.AddMappingProposal) sdk.Result {
	ctx.Logger().Info("handleAddMappingProposal", "proposal", proposal)

	if keeper.GetMappingInfo(ctx, proposal.IssueSymbol) != nil {
		return types.ErrDuplicatedIssueSymbol(DefaultCodespace, "duplicated issuer symbol").Result()
	}
	if keeper.GetMappingInfo(ctx, proposal.TargetSymbol) != nil {
		return types.ErrTargetSymbolUsedAsIssue(DefaultCodespace,
			"target symbol is used as issue symbol in another mapping").Result()
	}
	if keeper.HasTargetSymbol(ctx, proposal.IssueSymbol) {
		return types.ErrIssueSymbolUsedAsTarget(DefaultCodespace,
			"issue symbol is used as target symbol in another mapping").Result()
	}

	issueTokenInfo := keeper.tk.GetToken(ctx, proposal.IssueSymbol)
	if issueTokenInfo == nil {
		return sdk.ErrInvalidSymbol("issuer symbol does not exist").Result()
	}
	if issueTokenInfo.GetChain() == issueTokenInfo.GetSymbol() {
		return sdk.ErrInvalidSymbol("issuer symbol cannot be chain token").Result()
	}

	targetTokenInfo := keeper.tk.GetToken(ctx, proposal.TargetSymbol)
	if targetTokenInfo == nil {
		return sdk.ErrInvalidSymbol("target symbol does not exist").Result()
	}

	if !issueTokenInfo.GetTotalSupply().Equal(proposal.TotalSupply) {
		return types.ErrInvalidInitialIssuePool(DefaultCodespace,
			"initial issue pool does not match issue total supply").Result()
	}

	if issueTokenInfo.GetDecimals() != targetTokenInfo.GetDecimals() {
		return types.ErrUnmatchedDecimals(DefaultCodespace,
			"issue decimals do not match target decimals").Result()
	}

	fromCUAddr := proposal.From
	have := keeper.trk.GetBalance(ctx, fromCUAddr, proposal.IssueSymbol.String())
	if !have.Equal(proposal.TotalSupply) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("from CU's token balance %s is not equal to total supply of issue symbol %s", have.String(), proposal.TotalSupply.String())).Result()
	}
	pledgeCoin := sdk.NewCoin(proposal.IssueSymbol.String(), proposal.TotalSupply)
	feeCoin := sdk.NewCoin(sdk.NativeToken, keeper.NewMappingFee(ctx))
	need := sdk.NewCoins(pledgeCoin, feeCoin)
	_, flows, err := keeper.trk.SubCoins(ctx, fromCUAddr, need)
	if err != nil {
		return err.Result()
	}

	mappingInfo := &types.MappingInfo{
		IssueSymbol:  proposal.IssueSymbol,
		TargetSymbol: proposal.TargetSymbol,
		TotalSupply:  proposal.TotalSupply,
		IssuePool:    proposal.TotalSupply,
		Enabled:      true,
	}
	keeper.SetMappingInfo(ctx, mappingInfo)

	flows = append(flows, MappingBalanceFlow{
		IssueSymbol:       proposal.IssueSymbol,
		PreviousIssuePool: sdk.ZeroInt(),
		IssuePoolChange:   proposal.TotalSupply,
	})

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeMapping, flows)
	res := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &res)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteAddMappingProposal,
			sdk.NewAttribute(types.AttributeKeyFrom, proposal.From.String()),
			sdk.NewAttribute(types.AttributeKeyIssueToken, proposal.IssueSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyTargetToken, proposal.TargetSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyTotalSupply, proposal.TotalSupply.String()),
		),
	)

	res.Events = append(res.Events, ctx.EventManager().Events()...)
	return res
}

func handleSwitchMappingProposal(ctx sdk.Context, keeper Keeper, proposal types.SwitchMappingProposal) sdk.Result {
	ctx.Logger().Info("handleDisableTokenProposal", "proposal", proposal)

	mappingInfo := keeper.GetMappingInfo(ctx, proposal.IssueSymbol)
	if mappingInfo == nil {
		return types.ErrDuplicatedIssueSymbol(DefaultCodespace, "mapping not found").Result()
	}
	mappingInfo.Enabled = proposal.Enable

	keeper.SetMappingInfo(ctx, mappingInfo)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteSwitchMappingProposal,
			sdk.NewAttribute(types.AttributeKeyIssueToken, proposal.IssueSymbol.String()),
			sdk.NewAttribute(types.AttributeKeyEnable, strconv.FormatBool(proposal.Enable)),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func NewMappingProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Result {
		switch c := content.(type) {
		case types.AddMappingProposal:
			return handleAddMappingProposal(ctx, k, c)
		case types.SwitchMappingProposal:
			return handleSwitchMappingProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized mapping proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
