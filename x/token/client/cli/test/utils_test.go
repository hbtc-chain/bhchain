package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/client/cli"
)

func TestParseAddTokenProposalJSON(t *testing.T) {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	ti := sdk.TokenInfo{
		Symbol:              sdk.Symbol("usdt"),
		Issuer:              "0xC9476A4919a7E5c7e1760b68F945971769D5c1D8",
		Chain:               sdk.Symbol("eth"),
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    true,
		IsWithdrawalEnabled: true,
		Decimals:            6,
		TotalSupply:         sdk.NewIntWithDecimal(3, 16),
		CollectThreshold:    sdk.NewIntWithDecimal(2, 8),   // 200, tusdt
		OpenFee:             sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:          sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),      // gas * 10, eth
		MaxOpCUNumber:       10,
		SysTransferNum:      sdk.NewInt(5),     // gas * 5
		OpCUSysTransferNum:  sdk.NewInt(50),    // SysTransferAmount * 10
		GasLimit:            sdk.NewInt(80000), //  eth
		GasPrice:            sdk.NewInt(1000),
		DepositThreshold:    sdk.NewIntWithDecimal(2, 8), // 200tusdt
		Confirmations:       1,
		IsNonceBased:        false,
	}
	ap, err := cli.ParseAddTokenProposalJSON(cdc, "add_token_proposal.json")
	assert.Nil(t, err)
	assert.Equal(t, "Add Token", ap.Title)
	assert.Equal(t, "add token proposal", ap.Description)
	assert.Equal(t, ti, ap.TokenInfo)
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("hbc", sdk.NewInt(10000))), ap.Deposit)
}
func TestParseTokenParamsChangeProposalJSON(t *testing.T) {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	cp, err := cli.ParseTokenParamsChangeProposalJSON(cdc, "token_params_change_proposal.json")
	assert.Nil(t, err)
	assert.Equal(t, "Token Parameters Change", cp.Title)
	assert.Equal(t, "token parameter change proposal", cp.Description)
	assert.Equal(t, "testtoken", cp.Symbol)
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("hbc", sdk.NewInt(10000))), cp.Deposit)
	changes := cp.Changes.ToParamChanges()
	assert.Equal(t, sdk.KeyIsSendEnabled, changes[0].Key)
	assert.Equal(t, "true", changes[0].Value)
	assert.Equal(t, sdk.KeyIsDepositEnabled, changes[1].Key)
	assert.Equal(t, "false", changes[1].Value)
	assert.Equal(t, sdk.KeyIsWithdrawalEnabled, changes[2].Key)
	assert.Equal(t, "false", changes[2].Value)
	assert.Equal(t, sdk.KeyCollectThreshold, changes[3].Key)
	assert.Equal(t, "10000000000", changes[3].Value)
	assert.Equal(t, sdk.KeyDepositThreshold, changes[4].Key)
	assert.Equal(t, "20000000000", changes[4].Value)
	assert.Equal(t, sdk.KeyOpenFee, changes[5].Key)
	assert.Equal(t, "30000000000", changes[5].Value)
	assert.Equal(t, sdk.KeySysOpenFee, changes[6].Key)
	assert.Equal(t, "40000000000", changes[6].Value)
	assert.Equal(t, sdk.KeyWithdrawalFeeRate, changes[7].Key)
	assert.Equal(t, "2", changes[7].Value)
	assert.Equal(t, sdk.KeyMaxOpCUNumber, changes[8].Key)
	assert.Equal(t, "6", changes[8].Value)
	assert.Equal(t, sdk.KeySysTransferNum, changes[9].Key)
	assert.Equal(t, "500000", changes[9].Value)
	assert.Equal(t, sdk.KeyOpCUSysTransferNum, changes[10].Key)
	assert.Equal(t, "10000000", changes[10].Value)
	assert.Equal(t, sdk.KeyGasLimit, changes[11].Key)
	assert.Equal(t, "90000000000", changes[11].Value)
}

func TestParseDisableTokenProposalJSON(t *testing.T) {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	dp, err := cli.ParseDisableTokenProposalJSON(cdc, "disable_token_proposal.json")
	assert.Nil(t, err)
	assert.Equal(t, "Disable Token", dp.Title)
	assert.Equal(t, "disable token proposal", dp.Description)
	assert.Equal(t, "usdt", dp.Symbol)
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("hbc", sdk.NewInt(10000))), dp.Deposit)
}
