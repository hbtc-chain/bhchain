package gov

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/staking"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(input.keeper)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg := NewMsgSubmitProposal(
		ContentFromProposalType("test", "test", ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		input.addrs[0], 0,
	)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	EndBlocker(ctx, input.keeper)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	require.Equal(t, sdk.DecCoins{sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 5)}, input.dk.GetFeePool(ctx).CommunityPool)

}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(input.keeper)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg := NewMsgSubmitProposal(
		ContentFromProposalType("test", "test", ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		input.addrs[0], 0,
	)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg2 := NewMsgSubmitProposal(
		ContentFromProposalType("test2", "test2", ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		input.addrs[0], 0,
	)

	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	EndBlocker(ctx, input.keeper)
	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	EndBlocker(ctx, input.keeper)
	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickPassedDepositPeriod(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(input.keeper)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg := NewMsgSubmitProposal(
		ContentFromProposalType("test2", "test2", ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		input.addrs[0], 0,
	)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64

	for _, event := range res.Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == "proposal_id" {
					proposalID, _ = strconv.ParseInt(string(attribute.Value), 10, 64)
				}
			}
		}
	}

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := NewMsgDeposit(input.addrs[1], uint64(proposalID), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)
	SortAddresses(input.addrs)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(input.keeper)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(50000))}
	newProposalMsg := NewMsgSubmitProposal(testProposal(), proposalCoins, input.addrs[0], 0)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64

	for _, event := range res.Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == "proposal_id" {
					proposalID, _ = strconv.ParseInt(string(attribute.Value), 10, 64)
				}
			}
		}
	}

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := NewMsgDeposit(input.addrs[1], uint64(proposalID), proposalCoins)
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(input.keeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())

	var activeProposalID uint64

	require.NoError(t, input.keeper.cdc.UnmarshalBinaryLengthPrefixed(activeQueue.Value(), &activeProposalID))
	proposal, ok := input.keeper.GetProposal(ctx, activeProposalID)
	require.True(t, ok)
	require.Equal(t, StatusVotingPeriod, proposal.Status)
	depositsIterator := input.keeper.GetDepositsIterator(ctx, uint64(proposalID))
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	activeQueue.Close()

	EndBlocker(ctx, input.keeper)

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestProposalPassedEndblocker(t *testing.T) {
	input := getMockApp(t, 1, GenesisState{}, nil)
	SortAddresses(input.addrs)

	handler := NewHandler(input.keeper)
	stakingHandler := staking.NewHandler(input.sk)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	valAddr := sdk.ValAddress(input.addrs[0])

	createValidators(t, stakingHandler, ctx, []sdk.ValAddress{valAddr}, []int64{100000})
	staking.EndBlocker(ctx, input.sk)

	macc := input.keeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	initialModuleAccCoins := input.tk.GetAllBalance(ctx, macc.GetAddress())

	proposal, err := input.keeper.SubmitProposal(ctx, testProposal(), 0)
	require.NoError(t, err)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(100000))}
	newDepositMsg := NewMsgDeposit(input.addrs[0], proposal.ProposalID, proposalCoins)
	res := handler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	macc = input.keeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	moduleAccCoins := input.tk.GetAllBalance(ctx, macc.GetAddress())

	deposits := initialModuleAccCoins.Add(proposal.TotalDeposit).Add(proposalCoins)
	require.True(t, moduleAccCoins.IsEqual(deposits))

	err = input.keeper.AddVote(ctx, proposal.ProposalID, input.addrs[0], OptionYes)
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(input.keeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, input.keeper)

	macc = input.keeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	require.True(t, input.tk.GetAllBalance(ctx, macc.GetAddress()).IsEqual(initialModuleAccCoins))
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	input := getMockApp(t, 1, GenesisState{}, nil)
	SortAddresses(input.addrs)

	// hijack the router to one that will fail in a proposal's handler
	input.keeper.router = NewRouter().AddRoute(RouterKey, badProposalHandler)

	handler := NewHandler(input.keeper)
	stakingHandler := staking.NewHandler(input.sk)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	valAddr := sdk.ValAddress(input.addrs[0])

	createValidators(t, stakingHandler, ctx, []sdk.ValAddress{valAddr}, []int64{100000})
	staking.EndBlocker(ctx, input.sk)

	// Create a proposal where the handler will pass for the test proposal
	// because the value of contextKeyBadProposal is true.
	ctx = ctx.WithValue(contextKeyBadProposal, true)
	proposal, err := input.keeper.SubmitProposal(ctx, testProposal(), 0)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(100000)))
	newDepositMsg := NewMsgDeposit(input.addrs[0], proposal.ProposalID, proposalCoins)
	res := handler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	err = input.keeper.AddVote(ctx, proposal.ProposalID, input.addrs[0], OptionYes)
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(input.keeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// Set the contextKeyBadProposal value to false so that the handler will fail
	// during the processing of the proposal in the EndBlocker.
	ctx = ctx.WithValue(contextKeyBadProposal, false)

	// validate that the proposal fails/has been rejected
	EndBlocker(ctx, input.keeper)
}

func TestTickExpiriedDepositPeriodWithMultiDeposits(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(input.keeper)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg := NewMsgSubmitProposal(
		ContentFromProposalType("test2", "test2", ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		input.addrs[0], 0,
	)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID uint64

	for _, event := range res.Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == "proposal_id" {
					id, _ := strconv.ParseInt(string(attribute.Value), 10, 64)
					proposalID = uint64(id)
				}
			}
		}
	}

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := NewMsgDeposit(input.addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	newDepositMsg = NewMsgDeposit(input.addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 200000)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	newDepositMsg = NewMsgDeposit(input.addrs[2], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 34560)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 334565)}, proposal.TotalDeposit)
	require.Equal(t, StatusDepositPeriod, proposal.Status)

	deposits := input.keeper.GetDeposits(ctx, proposalID)
	require.Equal(t, 3, len(deposits))

	deposit, found := input.keeper.GetDeposit(ctx, proposalID, input.addrs[2])
	require.True(t, found)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 34560)}, deposit.Amount)

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	EndBlocker(ctx, input.keeper)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	require.Equal(t, sdk.DecCoins(nil), input.dk.GetFeePool(ctx).CommunityPool)

	EndBlocker(ctx, input.keeper)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	require.Equal(t, sdk.DecCoins{sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 334565)}, input.dk.GetFeePool(ctx).CommunityPool)
}

func TestTickExpiredVotingPeriod(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)
	SortAddresses(input.addrs)

	handler := NewHandler(input.keeper)
	stakingHandler := staking.NewHandler(input.sk)
	govHandler := NewHandler(input.keeper)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	valAddr0 := sdk.ValAddress(input.addrs[0])
	valAddr1 := sdk.ValAddress(input.addrs[1])
	valAddr2 := sdk.ValAddress(input.addrs[2])
	valAddr3 := sdk.ValAddress(input.addrs[3])

	createValidators(t, stakingHandler, ctx, []sdk.ValAddress{valAddr0, valAddr1, valAddr2, valAddr3}, []int64{100000, 100000, 100000, 100000})
	staking.EndBlocker(ctx, input.sk)

	macc := input.keeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	initialModuleAccCoins := input.tk.GetAllBalance(ctx, macc.GetAddress())
	require.Equal(t, sdk.Coins(nil), initialModuleAccCoins)

	inactiveQueue := input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(50000))}
	newProposalMsg := NewMsgSubmitProposal(testProposal(), proposalCoins, input.addrs[0], 0)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID uint64

	for _, event := range res.Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == "proposal_id" {
					id, _ := strconv.ParseInt(string(attribute.Value), 10, 64)
					proposalID = uint64(id)
				}
			}
		}
	}

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, input.keeper)

	newDepositMsg := NewMsgDeposit(input.addrs[1], proposalID, proposalCoins)
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(input.keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(input.keeper.GetVotingParams(ctx).VotingPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = input.keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())

	newvoteMsg0 := NewMsgVote(input.addrs[0], proposalID, OptionYes)
	res = handler(ctx, newvoteMsg0)
	newvoteMsg1 := NewMsgVote(input.addrs[1], proposalID, OptionNoWithVeto)
	res = handler(ctx, newvoteMsg1)
	newvoteMsg2 := NewMsgVote(input.addrs[2], proposalID, OptionNoWithVeto)
	res = handler(ctx, newvoteMsg2)
	newvoteMsg3 := NewMsgVote(input.addrs[3], proposalID, OptionAbstain)
	res = handler(ctx, newvoteMsg3)

	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())

	var activeProposalID uint64

	require.NoError(t, input.keeper.cdc.UnmarshalBinaryLengthPrefixed(activeQueue.Value(), &activeProposalID))
	proposal, ok := input.keeper.GetProposal(ctx, activeProposalID)
	require.True(t, ok)
	require.Equal(t, StatusVotingPeriod, proposal.Status)
	depositsIterator := input.keeper.GetDepositsIterator(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	activeQueue.Close()

	require.Equal(t, sdk.DecCoins(nil), input.dk.GetFeePool(ctx).CommunityPool)

	EndBlocker(ctx, input.keeper)

	activeQueue = input.keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	require.Equal(t, sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(100000))}, input.dk.GetFeePool(ctx).CommunityPool)
}
