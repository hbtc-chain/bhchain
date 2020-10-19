package types

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

var (
	testCU1, _ = sdk.CUAddressFromBase58("HBCXaiZNDZ6gsz165Rzbo9qRVpANFoKySYLh")
	testCU2, _ = sdk.CUAddressFromBase58("HBCeZXknux28DScaavB4eoC5CaVQcfnPwpHg")
	testCU3, _ = sdk.CUAddressFromBase58("HBCcqd2fE98XqdJMDuFvVzFJd5Tjoegi5Cfo")
	testCU4, _ = sdk.CUAddressFromBase58("HBCR86y741nFRA1waZvDH5eLDPWxswXEEsx5")
	allCUs     = []sdk.CUAddress{testCU1, testCU2, testCU3, testCU4}
)

func TestGasPriceVoteBox(t *testing.T) {
	// 测试最简单情形
	voteBox := NewGasPriceVoteBox(3)
	assert.False(t, voteBox.AddVote(testCU1, sdk.NewInt(90)))
	assert.False(t, voteBox.AddVote(testCU2, sdk.NewInt(100)))
	assert.False(t, voteBox.HasConfirmed())
	assert.True(t, voteBox.AddVote(testCU3, sdk.NewInt(110)))
	assert.False(t, voteBox.AddVote(testCU4, sdk.NewInt(120)))
	assert.True(t, voteBox.HasConfirmed())
	validVotes := voteBox.ValidVotes()
	assert.Equal(t, 4, len(validVotes))
	for i, item := range validVotes {
		assert.True(t, allCUs[i].Equals(item.Voter))
		assert.EqualValues(t, sdk.NewInt(int64(90+i*10)), item.Vote)
	}

	// 测试修改投票
	voteBox = NewGasPriceVoteBox(3)
	assert.False(t, voteBox.AddVote(testCU1, sdk.NewInt(90)))
	assert.False(t, voteBox.AddVote(testCU2, sdk.NewInt(20000)))
	assert.False(t, voteBox.AddVote(testCU2, sdk.NewInt(1000)))
	assert.False(t, voteBox.AddVote(testCU3, sdk.NewInt(100)))
	assert.False(t, voteBox.HasConfirmed())
	assert.True(t, voteBox.AddVote(testCU4, sdk.NewInt(100)))
	assert.False(t, voteBox.AddVote(testCU2, sdk.NewInt(100)))
	validVotes = voteBox.ValidVotes()
	assert.Equal(t, 3, len(validVotes))
}
