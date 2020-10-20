package types

import (
	"testing"
	"time"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

var native100 = sdk.NewCoin(sdk.NativeToken, sdk.NewInt(100))
var native159 = sdk.NewCoin(sdk.NativeToken, sdk.NewInt(159))
var native300 = sdk.NewCoin(sdk.NativeToken, sdk.NewInt(300))
var native1000 = sdk.NewCoin(sdk.NativeToken, sdk.NewInt(1000))
var nativehdt100 = sdk.NewCoin(sdk.NativeDefiToken, sdk.NewInt(100))
var nativehdt159 = sdk.NewCoin(sdk.NativeDefiToken, sdk.NewInt(159))
var nativehdt300 = sdk.NewCoin(sdk.NativeDefiToken, sdk.NewInt(300))
var nativehdt1000 = sdk.NewCoin(sdk.NativeDefiToken, sdk.NewInt(1000))
var dur10 = time.Duration(10)
var dur200 = time.Duration(200)

var data = []struct {
	minInitDeposit    sdk.Coin
	minDeposit        sdk.Coin
	minInitDaoDeposit sdk.Coin
	minDaoDeposit     sdk.Coin
	duration          time.Duration
}{
	{native100, native300, nativehdt100, nativehdt300, dur10},
	{native159, native1000, nativehdt159, nativehdt1000, dur200},
}

func TestNewDepositParams(t *testing.T) {

	for _, d := range data {
		dp := NewDepositParams(sdk.NewCoins(d.minInitDeposit), sdk.NewCoins(d.minDeposit), sdk.NewCoins(d.minInitDaoDeposit),
			sdk.NewCoins(d.minDaoDeposit), d.duration)
		assert.Equal(t, sdk.NewCoins(d.minInitDeposit), dp.MinInitDeposit)
		assert.Equal(t, sdk.NewCoins(d.minDeposit), dp.MinDeposit)
		assert.Equal(t, d.duration, dp.MaxDepositPeriod)
	}

}

func TestDepositParamsEqual(t *testing.T) {

	for _, d := range data {
		dp0 := NewDepositParams(sdk.NewCoins(d.minInitDeposit), sdk.NewCoins(d.minDeposit), sdk.NewCoins(d.minInitDaoDeposit),
			sdk.NewCoins(d.minDaoDeposit), d.duration)
		dp1 := NewDepositParams(sdk.NewCoins(d.minInitDeposit), sdk.NewCoins(d.minDeposit), sdk.NewCoins(d.minInitDaoDeposit),
			sdk.NewCoins(d.minDaoDeposit), d.duration)
		dp0.Equal(dp1)
	}

	d0, d1 := data[0], data[1]
	dp0 := NewDepositParams(sdk.NewCoins(d0.minInitDeposit), sdk.NewCoins(d0.minDeposit), sdk.NewCoins(d0.minInitDaoDeposit),
		sdk.NewCoins(d0.minDaoDeposit), d0.duration)
	dp1 := NewDepositParams(sdk.NewCoins(d1.minInitDeposit), sdk.NewCoins(d1.minDeposit), sdk.NewCoins(d0.minInitDaoDeposit),
		sdk.NewCoins(d0.minDaoDeposit), d0.duration)
	assert.False(t, dp0.Equal(dp1))
}
