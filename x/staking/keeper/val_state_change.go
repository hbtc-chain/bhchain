package keeper

import (
	"bytes"
	"fmt"
	"sort"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Apply and return accumulated updates to the bonded validator set. Also,
// * Updates the active valset as keyed by LastValidatorPowerKey.
// * Updates the total power as keyed by LastTotalPowerKey.
// * Updates validator status' according to updated powers.
// * Updates the fee pool bonded vs not-bonded tokens.
// * Updates relevant indices.
// It gets called once after genesis, another time maybe after genesis transactions,
// then once at every EndBlock.
//
// CONTRACT: Only validators with non-zero power or zero-power that were bonded
// at the previous block height or were removed from the validator set entirely
// are returned to Tendermint.
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) (updates []abci.ValidatorUpdate) {

	store := ctx.KVStore(k.storeKey)
	maxValidators := k.GetParams(ctx).MaxValidators
	maxKeyNodes := k.GetParams(ctx).MaxKeyNodes
	minValidatorDelegation := k.GetParams(ctx).MinValidatorDelegation
	minKeyNodeDelegation := k.GetParams(ctx).MinKeyNodeDelegation
	maxCandidateKeyNodeHeartbeatInterval := k.GetParams(ctx).MaxCandidateKeyNodeHeartbeatInterval
	blkHeight := uint64(ctx.BlockHeight())
	totalPower := sdk.ZeroInt()
	amtFromBondedToNotBonded, amtFromNotBondedToBonded := sdk.ZeroInt(), sdk.ZeroInt()

	// Retrieve the last validator set.
	// The persistent set is updated later in this function.
	// (see LastValidatorPowerKey).
	last := k.getLastValidatorsByAddr(ctx)
	lastKeyNodeSet := k.GetCurrentEpoch(ctx).KeyNodeSet

	var retainKeyNodes []sdk.ValAddress
	if ctx.BlockHeight() > 0 {
		maxUpdateKeyNodeNum := sdk.OneSixthCeil(uint16(len(lastKeyNodeSet)))
		k.jailQueuedValidatorNow(ctx, maxUpdateKeyNodeNum)
		// 计算必须连任的 keynodes
		retainKeyNodes = k.getRetainKeyNodes(ctx, lastKeyNodeSet, len(lastKeyNodeSet)-int(maxUpdateKeyNodeNum), minKeyNodeDelegation)
	}

	newKeyNodes := make([]sdk.CUAddress, 0)
	for _, valAddr := range retainKeyNodes {
		validator := k.mustGetValidator(ctx, valAddr)

		var valAddrBytes [sdk.AddrLen]byte
		copy(valAddrBytes[:], valAddr[:])
		oldPowerBytes := last[valAddrBytes]

		newPower := validator.ConsensusPower()
		newPowerBytes := k.cdc.MustMarshalBinaryLengthPrefixed(newPower)

		// update the validator set if power has changed
		if !bytes.Equal(oldPowerBytes, newPowerBytes) {
			updates = append(updates, validator.ABCIValidatorUpdate())

			// set validator power on lookup index
			k.SetLastValidatorPower(ctx, valAddr, newPower)
		}
		delete(last, valAddrBytes)

		// keep count
		totalPower = totalPower.Add(sdk.NewInt(newPower))

		newKeyNodes = append(newKeyNodes, sdk.CUAddress(valAddr))
	}

	// Iterate over validators, highest power to lowest.
	iterator := sdk.KVStoreReversePrefixIterator(store, types.ValidatorsByPowerIndexKey)
	defer iterator.Close()
	for count := len(retainKeyNodes); iterator.Valid() && count < int(maxValidators); iterator.Next() {

		// everything that is iterated in this loop is becoming or already a
		// part of the bonded validator set

		// fetch the validator
		valAddr := sdk.ValAddress(iterator.Value())
		if isInValAddresses(valAddr, retainKeyNodes) {
			continue
		}
		validator := k.mustGetValidator(ctx, valAddr)

		// if we get to a zero-power validator (which we don't bond),
		// there are no more possible bonded validators
		if validator.PotentialConsensusPower() == 0 {
			break
		}
		if validator.Tokens.LT(minValidatorDelegation) {
			break
		}

		// apply the appropriate state change if necessary
		switch {
		case validator.IsUnbonded():
			validator = k.unbondedToBonded(ctx, validator)
			amtFromNotBondedToBonded = amtFromNotBondedToBonded.Add(validator.GetTokens())
		case validator.IsUnbonding():
			validator = k.unbondingToBonded(ctx, validator)
			amtFromNotBondedToBonded = amtFromNotBondedToBonded.Add(validator.GetTokens())
		case validator.IsBonded():
			// no state change
		default:
			panic("unexpected validator status")
		}

		// fetch the old power bytes
		var valAddrBytes [sdk.AddrLen]byte
		copy(valAddrBytes[:], valAddr[:])
		oldPowerBytes, found := last[valAddrBytes]

		// calculate the new power bytes
		newPower := validator.ConsensusPower()
		newPowerBytes := k.cdc.MustMarshalBinaryLengthPrefixed(newPower)

		// update the validator set if power has changed
		if !found || !bytes.Equal(oldPowerBytes, newPowerBytes) {
			updates = append(updates, validator.ABCIValidatorUpdate())

			// set validator power on lookup index
			k.SetLastValidatorPower(ctx, valAddr, newPower)
		}

		// validator still in the validator set, so delete from the copy
		delete(last, valAddrBytes)

		// keep count
		count++
		totalPower = totalPower.Add(sdk.NewInt(newPower))

		if validator.CanBeKeyNode(minKeyNodeDelegation) && blkHeight-validator.LastKeyNodeHeartbeatHeight < maxCandidateKeyNodeHeartbeatInterval &&
			len(newKeyNodes) < int(maxKeyNodes) {
			newKeyNodes = append(newKeyNodes, sdk.CUAddress(validator.OperatorAddress))
		}
	}

	if isCUAddressSetDiff(newKeyNodes, lastKeyNodeSet) {
		// emit event and update epoch
		epoch := k.StartNewEpoch(ctx, newKeyNodes)
		if ctx.BlockHeight() > 0 {
			//set allopcu in migration // delete prekeygen order
			k.AfterNewEpoch(ctx, epoch)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeMigrationBegin,
					sdk.NewAttribute(types.AttributeMigrationNewEpochIndex, fmt.Sprintf("%d", epoch.Index)),
				),
			)
		}
	}

	// sort the no-longer-bonded validators
	noLongerBonded := sortNoLongerBonded(last)

	// iterate through the sorted no-longer-bonded validators
	for _, valAddrBytes := range noLongerBonded {

		// fetch the validator
		validator := k.mustGetValidator(ctx, valAddrBytes)

		// bonded to unbonding
		validator = k.bondedToUnbonding(ctx, validator)
		amtFromBondedToNotBonded = amtFromBondedToNotBonded.Add(validator.GetTokens())

		// delete from the bonded validator index
		k.DeleteLastValidatorPower(ctx, validator.GetOperator())

		// update the validator set
		updates = append(updates, validator.ABCIValidatorUpdateZero())
	}

	// Update the pools based on the recent updates in the validator set:
	// - The tokens from the non-bonded candidates that enter the new validator set need to be transferred
	// to the Bonded pool.
	// - The tokens from the bonded validators that are being kicked out from the validator set
	// need to be transferred to the NotBonded pool.
	switch {
	// Compare and subtract the respective amounts to only perform one transfer.
	// This is done in order to avoid doing multiple updates inside each iterator/loop.
	case amtFromNotBondedToBonded.GT(amtFromBondedToNotBonded):
		k.notBondedTokensToBonded(ctx, amtFromNotBondedToBonded.Sub(amtFromBondedToNotBonded))
	case amtFromNotBondedToBonded.LT(amtFromBondedToNotBonded):
		k.bondedTokensToNotBonded(ctx, amtFromBondedToNotBonded.Sub(amtFromNotBondedToBonded))
	default:
		// equal amounts of tokens; no update required
	}

	// set total power on lookup index if there are any updates
	if len(updates) > 0 {
		k.SetLastTotalPower(ctx, totalPower)
	}

	return updates
}

// Validator state transitions

func (k Keeper) bondedToUnbonding(ctx sdk.Context, validator types.Validator) types.Validator {
	if !validator.IsBonded() {
		panic(fmt.Sprintf("bad state transition bondedToUnbonding, validator: %v\n", validator))
	}
	return k.beginUnbondingValidator(ctx, validator)
}

func (k Keeper) unbondingToBonded(ctx sdk.Context, validator types.Validator) types.Validator {
	if !validator.IsUnbonding() {
		panic(fmt.Sprintf("bad state transition unbondingToBonded, validator: %v\n", validator))
	}
	return k.bondValidator(ctx, validator)
}

func (k Keeper) unbondedToBonded(ctx sdk.Context, validator types.Validator) types.Validator {
	if !validator.IsUnbonded() {
		panic(fmt.Sprintf("bad state transition unbondedToBonded, validator: %v\n", validator))
	}
	return k.bondValidator(ctx, validator)
}

// switches a validator from unbonding state to unbonded state
func (k Keeper) unbondingToUnbonded(ctx sdk.Context, validator types.Validator) types.Validator {
	if !validator.IsUnbonding() {
		panic(fmt.Sprintf("bad state transition unbondingToBonded, validator: %v\n", validator))
	}
	return k.completeUnbondingValidator(ctx, validator)
}

// send a validator to jail
func (k Keeper) jailValidator(ctx sdk.Context, validator types.Validator) {
	if validator.Jailed {
		return
	}

	validator.Jailed = true
	if validator.IsUnbonded() {
		k.DeleteValidatorByPowerIndex(ctx, validator)
	} else {
		index := k.insertJailedQueue(ctx, validator)
		validator.JailedIndex = index
	}

	k.SetValidator(ctx, validator)
}

// remove a validator from jail
func (k Keeper) unjailValidator(ctx sdk.Context, validator types.Validator) {
	if !validator.Jailed {
		panic(fmt.Sprintf("cannot unjail already unjailed validator, validator: %v\n", validator))
	}

	validator.Jailed = false

	// remove from jailed queue or readd to power index if not in queue
	if !k.deleteFromJailedQueue(ctx, validator) {
		k.SetValidatorByPowerIndex(ctx, validator)
	}

	validator.JailedIndex = 0
	k.SetValidator(ctx, validator)
}

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	// delete the validator by power index, as the key will change
	k.DeleteValidatorByPowerIndex(ctx, validator)

	// set the status
	validator = validator.UpdateStatus(sdk.Bonded)

	// save the now bonded validator record to the two referenced stores
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator)

	// delete from queue if present
	k.DeleteValidatorQueue(ctx, validator)

	// trigger hook
	k.AfterValidatorBonded(ctx, validator.ConsAddress(), validator.OperatorAddress)

	return validator
}

// perform all the store operations for when a validator begins unbonding
func (k Keeper) beginUnbondingValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	params := k.GetParams(ctx)

	// delete the validator by power index, as the key will change
	k.DeleteValidatorByPowerIndex(ctx, validator)

	// sanity check
	if validator.Status != sdk.Bonded {
		panic(fmt.Sprintf("should not already be unbonded or unbonding, validator: %v\n", validator))
	}

	// set the status
	validator = validator.UpdateStatus(sdk.Unbonding)

	// set the unbonding completion time and completion height appropriately
	validator.UnbondingCompletionTime = ctx.BlockHeader().Time.Add(params.UnbondingTime)
	validator.UnbondingHeight = ctx.BlockHeader().Height

	// save the now unbonded validator record and power index
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator)

	// Adds to unbonding validator queue
	k.InsertValidatorQueue(ctx, validator)

	// trigger hook
	k.AfterValidatorBeginUnbonding(ctx, validator.ConsAddress(), validator.OperatorAddress)

	return validator
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) completeUnbondingValidator(ctx sdk.Context, validator types.Validator) types.Validator {
	validator = validator.UpdateStatus(sdk.Unbonded)
	k.SetValidator(ctx, validator)
	return validator
}

// map of operator addresses to serialized power
type validatorsByAddr map[[sdk.AddrLen]byte][]byte

// 计算下周期必须连任的核心节点
func (k Keeper) getRetainKeyNodes(ctx sdk.Context, keyNodeSet []sdk.CUAddress, num int, minKeyNodeDelegation sdk.Int) []sdk.ValAddress {
	var validators []types.Validator
	for _, addr := range keyNodeSet {
		validator := k.mustGetValidator(ctx, sdk.ValAddress(addr))
		validators = append(validators, validator)
	}
	// 按票数排序
	sort.SliceStable(validators, func(i, j int) bool {
		if validators[i].Tokens.Equal(validators[j].Tokens) {
			return bytes.Compare(validators[i].OperatorAddress, validators[j].OperatorAddress) == 1
		}
		return validators[i].Tokens.GT(validators[j].Tokens)
	})
	// 过滤 jail 的节点，并按是否有资格成为 keynode 拆分
	var canBe, cannotBe []sdk.ValAddress
	maxKeyNodeHeartbeatInterval := k.GetParams(ctx).MaxKeyNodeHeartbeatInterval
	blkHeight := uint64(ctx.BlockHeight())
	for _, v := range validators {
		if v.Jailed {
			continue
		}
		if v.CanBeKeyNode(minKeyNodeDelegation) && blkHeight-v.LastKeyNodeHeartbeatHeight < maxKeyNodeHeartbeatInterval {
			canBe = append(canBe, v.OperatorAddress)
		} else {
			cannotBe = append(cannotBe, v.OperatorAddress)
		}
	}
	// 数量足够，直接截取
	if len(canBe) >= num {
		return canBe[:num]
	}
	// 数量不够，拿没有资格连任的拼凑
	return append(canBe, cannotBe[:(num-len(canBe))]...)
}

// get the last validator set
func (k Keeper) getLastValidatorsByAddr(ctx sdk.Context) validatorsByAddr {
	last := make(validatorsByAddr)
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.LastValidatorPowerKey)
	defer iterator.Close()
	// iterate over the last validator set index
	for ; iterator.Valid(); iterator.Next() {
		var valAddr [sdk.AddrLen]byte
		// extract the validator address from the key (prefix is 1-byte)
		copy(valAddr[:], iterator.Key()[1:])
		// power bytes is just the value
		powerBytes := iterator.Value()
		last[valAddr] = make([]byte, len(powerBytes))
		copy(last[valAddr][:], powerBytes[:])
	}
	return last
}

// given a map of remaining validators to previous bonded power
// returns the list of validators to be unbonded, sorted by operator address
func sortNoLongerBonded(last validatorsByAddr) [][]byte {
	// sort the map keys for determinism
	noLongerBonded := make([][]byte, len(last))
	index := 0
	for valAddrBytes := range last {
		valAddr := make([]byte, sdk.AddrLen)
		copy(valAddr[:], valAddrBytes[:])
		noLongerBonded[index] = valAddr
		index++
	}
	// sorted by address - order doesn't matter
	sort.SliceStable(noLongerBonded, func(i, j int) bool {
		// -1 means strictly less than
		return bytes.Compare(noLongerBonded[i], noLongerBonded[j]) == -1
	})
	return noLongerBonded
}

func isInValAddresses(addr sdk.ValAddress, addrs []sdk.ValAddress) bool {
	for _, a := range addrs {
		if addr.Equals(a) {
			return true
		}
	}
	return false
}

func isCUAddressSetDiff(a, b []sdk.CUAddress) bool {
	if len(a) != len(b) {
		return true
	}
	for _, cuA := range a {
		var found bool
		for _, cuB := range b {
			if cuA.Equals(cuB) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}
