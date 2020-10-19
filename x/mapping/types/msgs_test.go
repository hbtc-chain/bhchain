package types

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

var from, _ = sdk.CUAddressFromBase58("HBCZSkjCGQggAT28GcQednHbpJyfxHhmeTCH")

var errMappingNew = []struct {
	from        sdk.CUAddress
	mappingInfo MappingInfo
}{
	{},
	{from: sdk.CUAddress{}},
	{from: sdk.CUAddress([]byte{01, 02})},
	{from: from},
	{from: from, mappingInfo: MappingInfo{
		IssueSymbol: sdk.Symbol("TBTC"), // Invalid issue symbol
	}},
	{from: from, mappingInfo: MappingInfo{
		IssueSymbol: sdk.Symbol("tbtc"),
	}},
	{from: from, mappingInfo: MappingInfo{
		IssueSymbol:  sdk.Symbol("tbtc"),
		TargetSymbol: sdk.Symbol("TTC"), // Invalid target symbol
	}},
	{from: from, mappingInfo: MappingInfo{
		IssueSymbol:  sdk.Symbol("tbtc"),
		TargetSymbol: sdk.Symbol("btc"),
		TotalSupply:  sdk.NewInt(0),
	}},
	{from: from, mappingInfo: MappingInfo{
		IssueSymbol:  sdk.Symbol("tbtc"),
		TargetSymbol: sdk.Symbol("btc"),
		TotalSupply:  sdk.NewInt(2100),
		IssuePool:    sdk.NewInt(2101),
	}},
	{from: from, mappingInfo: MappingInfo{
		IssueSymbol:  sdk.Symbol("tbtc"),
		TargetSymbol: sdk.Symbol("btc"),
		TotalSupply:  sdk.NewInt(2100),
		IssuePool:    sdk.NewInt(2100),
		Enabled:      false,
	}},
}

var errMappingSwap = []struct {
	from        sdk.CUAddress
	issueSymbol sdk.Symbol
	coins       sdk.Coins
}{
	{},
	{from: sdk.CUAddress{}},
	{from: sdk.CUAddress([]byte{01, 02})},
	{from: from},
	{from: from, issueSymbol: sdk.Symbol("TBTC")}, // Invalid issue symbol
	{from: from, issueSymbol: sdk.Symbol("tbtc")},
	{from: from, issueSymbol: sdk.Symbol("tbtc"), coins: sdk.NewCoins(
		sdk.NewCoin("tbtc", sdk.NewInt(10)),
		sdk.NewCoin("btc", sdk.NewInt(10)),
	)},
}

func TestMappingSwap(t *testing.T) {
	for _, m := range errMappingSwap {
		msg := NewMsgMappingSwap(m.from, m.issueSymbol, m.coins)
		assert.NotNil(t, msg.ValidateBasic())
	}

	msg := NewMsgMappingSwap(
		from,
		sdk.Symbol("tbtc"),
		sdk.NewCoins(
			sdk.NewCoin("tbtc", sdk.NewInt(10)),
		))
	assert.Nil(t, msg.ValidateBasic())
}
