package token

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func newGenesisState(ibcTokens map[sdk.Symbol]*sdk.IBCToken) GenesisState {
	genTokens := make([]sdk.Token, 0, len(ibcTokens))

	for _, ibcToken := range ibcTokens {
		genTokens = append(genTokens, ibcToken)
	}

	return GenesisState{
		GenesisTokens: genTokens,
	}
}

func TestNewGenesisState(t *testing.T) {
	genState := newGenesisState(TestIBCTokens)
	err := ValidateGenesis(genState)
	assert.Nil(t, err)

	for _, info := range TestTokenData {
		var symbol = info.GetSymbol()
		assert.Equal(t, TestTokenData[symbol], info)
	}
}

func TestDefaultGensisState(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	genState1 := ExportGenesis(ctx, keeper)
	defaultTokens := map[string]sdk.Token{}
	exportTokens := map[string]sdk.Token{}
	for _, t := range TestTokenData {
		defaultTokens[t.GetSymbol().String()] = t
	}
	for _, t := range genState1.GenesisTokens {
		exportTokens[t.GetSymbol().String()] = t
	}
	for k, v := range defaultTokens {
		assert.EqualValues(t, v, exportTokens[k])
	}

	//assert.Equal(t, DefaultGenesisState(), genState1)
}

func TestIllegalGensisState(t *testing.T) {
	token := map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			TokenType: 0,
		},
	}

	genState := newGenesisState(token)
	err := ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			TokenType: 4,
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//TotalSupply is nil
	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//TotalSupply <0
	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(-1),
			},
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//CollectThreshold
	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(1),
			},
			TokenType:         1,
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(1),
			},
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(-1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//OpenFee
	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(1),
			},
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(1),
			},
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(-1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//Withdrawal Fee
	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(1),
			},
			TokenType:        1,
			CollectThreshold: sdk.NewInt(1),
			OpenFee:          sdk.NewInt(1),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]*sdk.IBCToken{
		"BHILLEGAL": {
			BaseToken: sdk.BaseToken{
				TotalSupply: sdk.NewInt(1),
			},
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = newGenesisState(token)
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//Duplicated token info
	btcTokenInfo := &sdk.IBCToken{
		BaseToken: sdk.BaseToken{
			Issuer:      "",
			Chain:       "",
			SendEnabled: false,
			Decimals:    8,
			TotalSupply: sdk.NewInt(2100),
		},
		TokenType:         sdk.UtxoBased,
		DepositEnabled:    false,
		WithdrawalEnabled: false,
		CollectThreshold:  sdk.NewInt(100),
		OpenFee:           sdk.NewInt(1000),
		SysOpenFee:        sdk.NewInt(1000),
		WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
	}
	btcTokenInfo.Symbol = "btc"
	btcTokenInfo2 := *btcTokenInfo
	btcTokenInfo2.Symbol = "aaa"
	genState = GenesisState{
		GenesisTokens: []sdk.Token{
			btcTokenInfo,
			&btcTokenInfo2,
		},
	}
	err = ValidateGenesis(genState)
	assert.NotNil(t, err)

}

func TestDefaultGenesisStateMarshal(t *testing.T) {
	defaulGenState := DefaultGenesisState()

	bz, err := ModuleCdc.MarshalJSON(defaulGenState)
	assert.Nil(t, err)

	var gotGenState GenesisState
	err = ModuleCdc.UnmarshalJSON(bz, &gotGenState)
	assert.Nil(t, err)
	assert.True(t, defaulGenState.Equal(gotGenState))
}
