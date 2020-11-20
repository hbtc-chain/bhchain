package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTokenTypeValid(t *testing.T) {
	assert.True(t, IsTokenTypeValid(UtxoBased))
	assert.True(t, IsTokenTypeValid(AccountBased))
	assert.True(t, IsTokenTypeValid(AccountSharedBased))
	assert.False(t, IsTokenTypeValid(0))
	assert.False(t, IsTokenTypeValid(4))
	assert.False(t, IsTokenTypeValid(5))
}

func TestIsValidTokenName(t *testing.T) {
	testdata := []struct {
		name  string
		valid bool
	}{
		{NativeToken, true},
		{"bh124", true},
		{"bh12345678901234", true},
		{"bhabc", true},
		{"bhabc123", true},
		{"bh123456789012345", false}, // length limit
		{"bhCABC", false},
		{" bh124", false},
		{"_bh124", false},
		{"BhT", false},
		{"bHT", false},
		{"bh 123", false},
		{"bh123 ", false},
		{"bh#123 ", false},
		{"bh123% ", false},
		{"HBC124", false},
		{"bh^124", false},
		{"bh 125", false},
		{"bh 125*", false},
		{"bh125 ", false},
		{"1bhC", false},
		{"B1HC", false},
		{"BTC", false},
		{"bhABCDEFGHIGKLMNOP", false},
	}

	for _, d := range testdata {

		assert.Equal(t, d.valid, IsTokenNameValid(d.name), d.name)

		if d.valid {
			assert.Equal(t, strings.ToLower(d.name), d.name)
			// must be valid denom
			assert.Nil(t, validateDenom(d.name))
		}
	}
}

func TestIBCTokenInfoIsValid(t *testing.T) {
	tokenInfo := IBCToken{
		BaseToken: BaseToken{
			Name:        "btc",
			Symbol:      "btc",
			Issuer:      "",
			Chain:       "btc",
			SendEnabled: true,
			Decimals:    8,
			TotalSupply: NewIntWithDecimal(21, 15),
		},
		TokenType:          UtxoBased,
		DepositEnabled:     true,
		WithdrawalEnabled:  true,
		CollectThreshold:   NewIntWithDecimal(2, 4),   // btc
		OpenFee:            NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:         NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:  NewDecWithPrec(2, 0),      // gas * 10  btc
		MaxOpCUNumber:      10,
		SysTransferNum:     NewInt(3),
		OpCUSysTransferNum: NewInt(30),
		GasLimit:           NewInt(1),
		GasPrice:           NewInt(1000),
		DepositThreshold:   NewIntWithDecimal(2, 3),
	}

	assert.True(t, tokenInfo.IsValid())

	// name is illegal
	tokenInfo.Name = "Btc"
	assert.False(t, tokenInfo.IsValid())

	// symbol is illegal
	tokenInfo.Name = "btc"
	tokenInfo.Symbol = "b tc"
	assert.False(t, tokenInfo.IsValid())

	//chain is illegal
	tokenInfo.Symbol = "btc"
	tokenInfo.Chain = "b tc"
	assert.False(t, tokenInfo.IsValid())

	//tokentype is illegal
	tokenInfo.Chain = "btc"
	tokenInfo.TokenType = 9
	assert.False(t, tokenInfo.IsValid())

	//tokeType is legal
	tokenInfo.TokenType = UtxoBased
	assert.True(t, tokenInfo.IsValid())

	tokenInfo.TokenType = AccountBased
	assert.True(t, tokenInfo.IsValid())

	tokenInfo.TokenType = AccountSharedBased
	assert.True(t, tokenInfo.IsValid())

	//Decimal is illegal
	for i := 1; i <= 18; i++ {
		tokenInfo.Decimals = uint64(i)
		assert.True(t, tokenInfo.IsValid())
	}
	tokenInfo.Decimals = 19
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.Decimals = 18
	assert.True(t, tokenInfo.IsValid())

	//TotalSupply is -1
	tokenInfo.TotalSupply = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	//TotalSupply is 0
	tokenInfo.TotalSupply = NewInt(0)
	assert.True(t, tokenInfo.IsValid())

	//CollectThreshold is not positive
	tokenInfo.CollectThreshold = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.CollectThreshold = NewInt(0)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.CollectThreshold = NewInt(1)
	assert.True(t, tokenInfo.IsValid())

	//DepositThreshold is not positive
	tokenInfo.DepositThreshold = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.DepositThreshold = NewInt(0)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.DepositThreshold = NewInt(1)
	assert.True(t, tokenInfo.IsValid())

	//OpenFee is not positive
	tokenInfo.OpenFee = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.OpenFee = NewInt(0)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.OpenFee = NewInt(1)
	assert.True(t, tokenInfo.IsValid())

	//SysOpenFee is not positive
	tokenInfo.SysOpenFee = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.SysOpenFee = NewInt(0)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.SysOpenFee = NewInt(1)
	assert.True(t, tokenInfo.IsValid())

	//SysOpenFee is not positive
	tokenInfo.WithdrawalFeeRate = NewDecWithPrec(2, 0)
	assert.True(t, tokenInfo.IsValid())

	//GasLimit is negative
	tokenInfo.GasLimit = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.GasLimit = NewInt(1)
	assert.True(t, tokenInfo.IsValid())

	//GasLimit is negative
	tokenInfo.GasPrice = NewInt(-1)
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.GasPrice = NewInt(1)
	assert.True(t, tokenInfo.IsValid())

	//GasLimit is not positive
	tokenInfo.MaxOpCUNumber = 0
	assert.False(t, tokenInfo.IsValid())

	tokenInfo.MaxOpCUNumber = 1
	assert.True(t, tokenInfo.IsValid())
}
