package staking

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking/keeper"
	"github.com/hbtc-chain/bhchain/x/staking/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateValidator:
			return handleMsgCreateValidator(ctx, msg, k)

		case types.MsgEditValidator:
			return handleMsgEditValidator(ctx, msg, k)

		case types.MsgKeyNodeHeartbeat:
			return handleMsgKeyNodeHeartbeat(ctx, msg, k)

		case types.MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)

		case types.MsgBeginRedelegate:
			return handleMsgBeginRedelegate(ctx, msg, k)

		case types.MsgUndelegate:
			return handleMsgUndelegate(ctx, msg, k)

		default:
			errMsg := fmt.Sprintf("unrecognized staking message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgCreateValidator(ctx sdk.Context, msg types.MsgCreateValidator, k keeper.Keeper) sdk.Result {
	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, msg.ValidatorAddress); found {
		return ErrValidatorOwnerExists(k.Codespace()).Result()
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(msg.PubKey)); found {
		return ErrValidatorPubKeyExists(k.Codespace()).Result()
	}

	if msg.Value.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
		return ErrSelfDelegationBelowMinimum(k.Codespace()).Result()
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return err.Result()
	}

	if ctx.ConsensusParams() != nil {
		tmPubKey := tmtypes.TM2PB.PubKey(msg.PubKey)
		if !common.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
			return ErrValidatorPubKeyTypeNotSupported(k.Codespace(),
				tmPubKey.Type,
				ctx.ConsensusParams().Validator.PubKeyTypes).Result()
		}
	}

	if msg.IsKeyNode && ctx.BlockHeight() != 0 {
		return ErrKeyNodeDeny(k.Codespace()).Result()
	}

	validator := NewValidator(msg.ValidatorAddress, msg.PubKey, msg.Description, msg.IsKeyNode)
	commission := NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)
	validator, err := validator.SetInitialCommission(commission)
	if err != nil {
		return err.Result()
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.OperatorAddress)

	// move coins from the msg.Address CU to a (self-delegation) delegator CU
	// the validator CU and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	result, _, err := k.Delegate(ctx, msg.DelegatorAddress, msg.Value.Amount, sdk.Unbonded, validator, true)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgEditValidator(ctx sdk.Context, msg types.MsgEditValidator, k keeper.Keeper) sdk.Result {
	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return err.Result()
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return err.Result()
		}

		// call the before-modification hook since we're about to update the commission
		k.BeforeValidatorModified(ctx, msg.ValidatorAddress)

		validator.Commission = commission
	}
	if msg.MinSelfDelegation != nil {
		if !(*msg.MinSelfDelegation).GT(validator.MinSelfDelegation) {
			return ErrMinSelfDelegationDecreased(k.Codespace()).Result()
		}
		if (*msg.MinSelfDelegation).GT(validator.Tokens) {
			return ErrSelfDelegationBelowMinimum(k.Codespace()).Result()
		}
		validator.MinSelfDelegation = (*msg.MinSelfDelegation)
	}
	if msg.IsKeyNode != nil {
		if *msg.IsKeyNode {
			minKeyNodeDelegation := k.GetParams(ctx).MinKeyNodeDelegation
			if validator.Tokens.LT(minKeyNodeDelegation) {
				return ErrDelegationBelowKeyNodeThreshold(k.Codespace()).Result()
			}
			if validator.LastKeyNodeHeartbeatHeight == 0 || uint64(ctx.BlockHeight())-validator.LastKeyNodeHeartbeatHeight >= k.MaxCandidateKeyNodeHeartbeatInterval(ctx) {
				return ErrKeyNodeNoHeartbeat(k.Codespace()).Result()
			}
		}
		validator.IsKeyNode = *msg.IsKeyNode
	}

	k.SetValidator(ctx, validator)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
			// sdk.NewAttribute(types.AttributeKeyMinSelfDelegation, validator.MinSelfDelegation.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgKeyNodeHeartbeat(ctx sdk.Context, msg types.MsgKeyNodeHeartbeat, k keeper.Keeper) sdk.Result {
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	if msg.Height >= uint64(ctx.BlockHeight()) {
		return ErrKeyNodeInvalidHeartbeat(k.Codespace()).Result()
	}
	validator.LastKeyNodeHeartbeatHeight = msg.Height
	k.SetValidator(ctx, validator)
	return sdk.Result{}
}

func handleMsgDelegate(ctx sdk.Context, msg types.MsgDelegate, k keeper.Keeper) sdk.Result {
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	// NOTE: source funds are always unbonded
	result, _, err := k.Delegate(ctx, msg.DelegatorAddress, msg.Amount.Amount, sdk.Unbonded, validator, true)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func handleMsgUndelegate(ctx sdk.Context, msg types.MsgUndelegate, k keeper.Keeper) sdk.Result {
	shares, err := k.ValidateUnbondAmount(
		ctx, msg.DelegatorAddress, msg.ValidatorAddress, msg.Amount.Amount,
	)
	if err != nil {
		return err.Result()
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	completionTime, err := k.Undelegate(ctx, msg.DelegatorAddress, msg.ValidatorAddress, shares)
	if err != nil {
		return err.Result()
	}

	completionTimeBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(completionTime)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return sdk.Result{Data: completionTimeBz, Events: ctx.EventManager().Events()}
}

func handleMsgBeginRedelegate(ctx sdk.Context, msg types.MsgBeginRedelegate, k keeper.Keeper) sdk.Result {
	shares, err := k.ValidateUnbondAmount(
		ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.Amount.Amount,
	)
	if err != nil {
		return err.Result()
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	completionTime, err := k.BeginRedelegation(
		ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.ValidatorDstAddress, shares,
	)
	if err != nil {
		return err.Result()
	}

	completionTimeBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(completionTime)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress.String()),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return sdk.Result{Data: completionTimeBz, Events: ctx.EventManager().Events()}
}
