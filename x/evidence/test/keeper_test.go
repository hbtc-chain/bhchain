package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVote(t *testing.T) {
	env := setupUnitTestEnv()
	mockEpoch := sdk.Epoch{}
	for _, val := range env.validators {
		mockEpoch.KeyNodeSet = append(mockEpoch.KeyNodeSet, sdk.CUAddress(val.OperatorAddress))
	}
	env.stakingKeeper.On("GetCurrentEpoch", mock.Anything).Return(mockEpoch)

	firstConfirmed, confirmed, validVotes := env.keeper.Vote(env.ctx, "vote1", addr1, types.BoolVote(true), 10)
	assert.False(t, firstConfirmed)
	assert.False(t, confirmed)
	assert.Equal(t, 0, len(validVotes))
	firstConfirmed, confirmed, validVotes = env.keeper.Vote(env.ctx, "vote1", addr2, types.BoolVote(true), 10)
	assert.False(t, firstConfirmed)
	assert.False(t, confirmed)
	assert.Equal(t, 0, len(validVotes))
	firstConfirmed, confirmed, validVotes = env.keeper.Vote(env.ctx, "vote1", addr3, types.BoolVote(false), 10)
	assert.False(t, firstConfirmed)
	assert.False(t, confirmed)
	assert.Equal(t, 0, len(validVotes))
	firstConfirmed, confirmed, validVotes = env.keeper.Vote(env.ctx, "vote1", addr4, types.BoolVote(true), 10)
	assert.True(t, firstConfirmed)
	assert.True(t, confirmed)
	assert.Equal(t, 3, len(validVotes))
}

func TestRecordMisbehaviourVoter(t *testing.T) {
	env := setupUnitTestEnv()
	mockEpoch := sdk.Epoch{}
	for _, val := range env.validators {
		mockEpoch.KeyNodeSet = append(mockEpoch.KeyNodeSet, sdk.CUAddress(val.OperatorAddress))
	}
	env.stakingKeeper.On("GetCurrentEpoch", mock.Anything).Return(mockEpoch)
	env.stakingKeeper.On("GetEpochByHeight", mock.Anything, mock.Anything).Return(mockEpoch)

	for i := 1; i <= 11; i++ {
		env.keeper.Vote(env.ctx, fmt.Sprintf("vote-%d", i), addr1, types.BoolVote(true), uint64(i))
		env.keeper.Vote(env.ctx, fmt.Sprintf("vote-%d", i), addr2, types.BoolVote(true), uint64(i))
		env.keeper.Vote(env.ctx, fmt.Sprintf("vote-%d", i), addr3, types.BoolVote(true), uint64(i))
	}

	env.stakingKeeper.On("JailByOperator", mock.Anything, env.validators[3].OperatorAddress)
	env.stakingKeeper.On("SlashByOperator", mock.Anything, env.validators[3].OperatorAddress, mock.Anything, mock.Anything)
	env.ctx = env.ctx.WithBlockHeight(20)
	env.keeper.RecordMisbehaviourVoter(env.ctx)
	env.stakingKeeper.AssertNotCalled(t, "JailByOperator")
	env.stakingKeeper.AssertNotCalled(t, "SlashByOperator")
	env.ctx = env.ctx.WithBlockHeight(21)
	env.keeper.RecordMisbehaviourVoter(env.ctx)
	env.stakingKeeper.AssertNumberOfCalls(t, "JailByOperator", 1)
	env.stakingKeeper.AssertNumberOfCalls(t, "SlashByOperator", 1)
	env.ctx = env.ctx.WithBlockHeight(31)
	env.keeper.RecordMisbehaviourVoter(env.ctx)
	env.stakingKeeper.AssertNumberOfCalls(t, "JailByOperator", 1)
	env.stakingKeeper.AssertNumberOfCalls(t, "SlashByOperator", 1)
}
