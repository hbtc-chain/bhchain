package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTokenTypeLegal(t *testing.T) {
	assert.True(t, IsTokenTypeLegal(UtxoBased))
	assert.True(t, IsTokenTypeLegal(AccountBased))
	assert.True(t, IsTokenTypeLegal(AccountSharedBased))
	assert.False(t, IsTokenTypeLegal(0))
	assert.False(t, IsTokenTypeLegal(4))
	assert.False(t, IsTokenTypeLegal(5))
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
		{"bh123456789012345", false},
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

		assert.Equal(t, d.valid, Symbol(d.name).IsValidTokenName(), d.name)

		if Symbol(d.name).IsValidTokenName() {
			assert.Equal(t, strings.ToLower(d.name), Symbol(d.name).ToDenomName())
			assert.Equal(t, d.name, Symbol(d.name).String())
			// must be valid denom
			assert.Nil(t, validateDenom(d.name))
			// symbol == DenomName // TODO remove symbol ,use DenomName
			assert.Equal(t, d.name, Symbol(d.name).ToDenomName())
		} else {
			assert.Equal(t, "", Symbol(d.name).String())
		}
	}
}

func TestTokenString(t *testing.T) {
	expected := "\n\tSymbol:btc\n\tIssuer:iss\n\tChain:\n\tTokenType:1\n\tIsSendEnabled:false\n\tIsDepositEnabled:false\n\tIsWithdrawalEnabled:false\n\tDecimals:8\n\tTotalSupply:<nil>\n\tCollectThreshold:100\n\tDepositThreshold:<nil>\n\tOpenFee:1000\n\tSysOpenFee:1100\n\tWithdrawalFee:2.000000000000000000\n\tMaxOpCUNumber:102\n\tSysTransferNum:10\n\tOpCUSysTransferNum:100\n\tGasLimit:100000000000000\n\tGasPrice:1\n\tConfirmations:1\n\tIsNonceBased:false\n\t"
	d := TokenInfo{
		Symbol:              "btc",
		Issuer:              "iss",
		Chain:               "mt",
		TokenType:           UtxoBased,
		IsSendEnabled:       false,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		CollectThreshold:    NewInt(100),
		OpenFee:             NewInt(1000),
		SysOpenFee:          NewInt(1100),
		WithdrawalFeeRate:   NewDecWithPrec(2, 0),
		MaxOpCUNumber:       102,
		SysTransferNum:      NewInt(10),
		OpCUSysTransferNum:  NewInt(100),
		GasLimit:            NewIntWithDecimal(1, 14),
		GasPrice:            NewInt(1),
		Confirmations:       1,
		IsNonceBased:        false,
	}
	assert.Equal(t, expected, d.String())
}

func TestTokenInfoIsValid(t *testing.T) {
	tokenInfo := TokenInfo{
		Symbol:              Symbol("btc"),
		Issuer:              "",
		Chain:               Symbol("btc"),
		TokenType:           UtxoBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    true,
		IsWithdrawalEnabled: true,
		Decimals:            8,
		TotalSupply:         NewIntWithDecimal(21, 15),
		CollectThreshold:    NewIntWithDecimal(2, 4),   // btc
		OpenFee:             NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:          NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:   NewDecWithPrec(2, 0),      // gas * 10  btc
		MaxOpCUNumber:       10,
		SysTransferNum:      NewInt(3),
		OpCUSysTransferNum:  NewInt(30),
		GasLimit:            NewInt(1),
		GasPrice:            NewInt(1000),
		DepositThreshold:    NewIntWithDecimal(2, 3),
	}

	assert.True(t, tokenInfo.IsValid())

	//symbol is illegal
	tokenInfo.Symbol = Symbol("Btc")
	assert.False(t, tokenInfo.IsValid())

	//chain is illegal
	tokenInfo.Symbol = Symbol("btc")
	tokenInfo.Chain = Symbol("BTC")
	assert.False(t, tokenInfo.IsValid())

	//tokentype is illegal
	tokenInfo.Chain = Symbol("btc")
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
