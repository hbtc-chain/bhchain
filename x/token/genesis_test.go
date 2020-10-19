/*
 * *******************************************************************
 * @项目名称: token
 * @文件名称: genesis_test.go
 * @Date: 2019/06/14
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */
package token

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestNewGenesisState(t *testing.T) {
	genState := NewGenesisState(TestTokenData, DefaultParams())
	_, err := ValidateGenesis(genState)
	assert.Nil(t, err)

	for _, info := range genState.GenesisTokenInfos {
		symbol := sdk.Symbol(info.Symbol)
		assert.Equal(t, TestTokenData[symbol], info.TokenInfo)
	}

	assert.Equal(t, DefaultParams(), genState.Params)
}

func TestDefaultGensisState(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	InitGenesis(ctx, keeper, DefaultGenesisState())
	genState1 := ExportGenesis(ctx, keeper)
	assert.Equal(t, DefaultGenesisState(), genState1)
}

func TestIllegalGensisState(t *testing.T) {
	token := map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType: 0,
		},
	}

	genState := NewGenesisState(token, DefaultParams())
	_, err := ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType: 4,
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//TotalSupply is nil
	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//TotalSupply <0
	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			TotalSupply:       sdk.NewInt(-1),
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//CollectThreshold
	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			TotalSupply:       sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			TotalSupply:       sdk.NewInt(1),
			CollectThreshold:  sdk.NewInt(-1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//OpenFee
	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			TotalSupply:       sdk.NewInt(1),
			CollectThreshold:  sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			TotalSupply:       sdk.NewInt(1),
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(-1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//Withdrawal Fee
	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:        1,
			TotalSupply:      sdk.NewInt(1),
			CollectThreshold: sdk.NewInt(1),
			OpenFee:          sdk.NewInt(1),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	token = map[sdk.Symbol]sdk.TokenInfo{
		"BHILLEGAL": {
			TokenType:         1,
			TotalSupply:       sdk.NewInt(1),
			CollectThreshold:  sdk.NewInt(1),
			OpenFee:           sdk.NewInt(1),
			WithdrawalFeeRate: sdk.NewDecWithPrec(2, 0),
		},
	}
	genState = NewGenesisState(token, DefaultParams())
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

	//Duplicated token info
	btcTokenInfo := sdk.TokenInfo{
		Issuer:              "",
		Chain:               "",
		TokenType:           sdk.UtxoBased,
		IsSendEnabled:       false,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            8,
		TotalSupply:         sdk.NewInt(2100),
		CollectThreshold:    sdk.NewInt(100),
		OpenFee:             sdk.NewInt(1000),
		SysOpenFee:          sdk.NewInt(1000),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
	}
	genState = GenesisState{
		GenesisTokenInfos: []GenesisTokenInfo{
			{Symbol: "btc", TokenInfo: btcTokenInfo},
			{Symbol: "btc", TokenInfo: btcTokenInfo},
			{Symbol: "eth", TokenInfo: btcTokenInfo}, // set btcTokenInfo for test only
		},
		Params: DefaultParams(),
	}
	_, err = ValidateGenesis(genState)
	assert.NotNil(t, err)

}

func TestDefaultGenesisStateMarshal(t *testing.T) {
	defaulGenState := DefaultGenesisState()
	bz, err := json.Marshal(defaulGenState)
	assert.Nil(t, err)

	var gotGenState GenesisState
	err = json.Unmarshal(bz, &gotGenState)
	assert.Nil(t, err)
	assert.True(t, defaulGenState.Equal(gotGenState))
}

func TestAddTokenInfoIntoGenesis(t *testing.T) {
	defaulGenState := DefaultGenesisState()
	gs := &defaulGenState

	origLen := len(gs.GenesisTokenInfos)

	symbol := sdk.Symbol("bheos")
	eos := GenesisTokenInfo{Symbol: symbol.String(), TokenInfo: sdk.TokenInfo{
		Symbol:              symbol,
		Issuer:              "",
		Chain:               "",
		TokenType:           sdk.AccountSharedBased,
		IsSendEnabled:       false,
		IsDepositEnabled:    false,
		IsWithdrawalEnabled: false,
		Decimals:            18,
		TotalSupply:         sdk.NewIntWithDecimal(1, 28),
		CollectThreshold:    sdk.NewIntWithDecimal(2, 19),
		OpenFee:             sdk.NewIntWithDecimal(1, 18),
		SysOpenFee:          sdk.NewIntWithDecimal(1, 18),
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
		MaxOpCUNumber:       3,
		SysTransferNum:      sdk.NewInt(3),  // gas * 3
		OpCUSysTransferNum:  sdk.NewInt(30), // SysTransferNum * 10
		GasLimit:            sdk.NewInt(1000000),
		DepositThreshold:    sdk.NewIntWithDecimal(1, 19),
	}}
	err := gs.AddTokenInfoIntoGenesis(eos)
	assert.Nil(t, err)
	assert.Equal(t, origLen+1, len(gs.GenesisTokenInfos))

	symbol = sdk.Symbol("btc")
	btc := GenesisTokenInfo{Symbol: symbol.String(), TokenInfo: TestTokenData[symbol]}
	err = gs.AddTokenInfoIntoGenesis(btc)
	assert.NotNil(t, err)
	assert.Equal(t, origLen+1, len(gs.GenesisTokenInfos))

}
