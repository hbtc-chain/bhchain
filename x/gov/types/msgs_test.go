package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
)

var (
	coinsPos         = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinsZero        = sdk.NewCoins()
	coinsPosNotAtoms = sdk.NewCoins(sdk.NewInt64Coin("foo", 10000))
	coinsMulti       = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000), sdk.NewInt64Coin("foo", 10000))
	addrs            = []sdk.CUAddress{
		sdk.CUAddress("test1"),
		sdk.CUAddress("test2"),
	}
)

func init() {
	coinsMulti.Sort()
}

// test ValidateBasic for MsgCreateValidator
func TestMsgSubmitProposal(t *testing.T) {
	tests := []struct {
		title, description string
		proposalType       string
		proposerAddr       sdk.CUAddress
		initialDeposit     sdk.Coins
		expectPass         bool
	}{
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsPos, true},
		{"", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsPos, false},
		{"Test Proposal", "", ProposalTypeText, addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, sdk.CUAddress{}, coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsZero, true},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsMulti, true},
		{strings.Repeat("#", MaxTitleLength*2), "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsMulti, false},
		{"Test Proposal", strings.Repeat("#", MaxDescriptionLength*2), ProposalTypeText, addrs[0], coinsMulti, false},
	}

	for i, tc := range tests {
		msg := NewMsgSubmitProposal(
			ContentFromProposalType(tc.title, tc.description, tc.proposalType),
			tc.initialDeposit,
			tc.proposerAddr,
			0,
		)

		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgDepositGetSignBytes(t *testing.T) {
	addr, _ := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	msg := NewMsgDeposit(addr, 0, coinsPos)
	res := msg.GetSignBytes()

	expected := `{"type":"hbtcchain/gov/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"hbc"}],"depositor":"HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy","proposal_id":"0"}}`
	require.Equal(t, expected, string(res))
}

// test ValidateBasic for MsgDeposit
func TestMsgDeposit(t *testing.T) {
	tests := []struct {
		proposalID    uint64
		depositorAddr sdk.CUAddress
		depositAmount sdk.Coins
		expectPass    bool
	}{
		{0, addrs[0], coinsPos, true},
		{1, sdk.CUAddress{}, coinsPos, false},
		{1, addrs[0], coinsZero, true},
		{1, addrs[0], coinsMulti, true},
	}

	for i, tc := range tests {
		msg := NewMsgDeposit(tc.depositorAddr, tc.proposalID, tc.depositAmount)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// test ValidateBasic for MsgDeposit
func TestMsgVote(t *testing.T) {
	tests := []struct {
		proposalID uint64
		voterAddr  sdk.CUAddress
		option     VoteOption
		expectPass bool
	}{
		{0, addrs[0], OptionYes, true},
		{0, sdk.CUAddress{}, OptionYes, false},
		{0, addrs[0], OptionNo, true},
		{0, addrs[0], OptionNoWithVeto, true},
		{0, addrs[0], OptionAbstain, true},
		{0, addrs[0], VoteOption(0x13), false},
	}

	for i, tc := range tests {
		msg := NewMsgVote(tc.voterAddr, tc.proposalID, tc.option)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}
