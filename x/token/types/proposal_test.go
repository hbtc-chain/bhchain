package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestAddTokenProposal(t *testing.T) {
	ti := sdk.TokenInfo{
		Symbol:              sdk.Symbol("btc"),
		Issuer:              "",
		Chain:               sdk.Symbol("btc"),
		TokenType:           sdk.UtxoBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    true,
		IsWithdrawalEnabled: true,
		Decimals:            8,
		TotalSupply:         sdk.NewIntWithDecimal(21, 15),
		CollectThreshold:    sdk.NewIntWithDecimal(2, 5),   // btc
		OpenFee:             sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:          sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),      // gas * 10  btc
		MaxOpCUNumber:       10,
		SysTransferNum:      sdk.NewInt(3000),
		OpCUSysTransferNum:  sdk.NewInt(30000),
		GasLimit:            sdk.NewInt(1),
		GasPrice:            sdk.NewInt(1000),
		DepositThreshold:    sdk.NewIntWithDecimal(2, 4),
		Confirmations:       1,
		IsNonceBased:        false,
	}

	atp := NewAddTokenProposal("Test", "Description", ti)
	require.Equal(t, "Test", atp.GetTitle())
	require.Equal(t, "Description", atp.GetDescription())
	require.Equal(t, RouterKey, atp.ProposalRoute())
	require.Equal(t, ProposalTypeAddToken, atp.ProposalType())
	require.Nil(t, atp.ValidateBasic())
	//	require.Equal(t, expectedStr, atp.String())

}

func TestTokenParamsChangeProposal(t *testing.T) {
	expectedStr := "Change Token Param Proposal:\n Title:       Test\n Description: Description\n Symbol:      btc\n Changes:\nis_deposit_enabled: true\tmax_op_cu_number: \"0\"\tcollect_threshold: \"21000000000000000\"\t"

	changes := []ParamChange{}
	changes = append(changes, NewParamChange(sdk.KeyIsDepositEnabled, "true"))
	changes = append(changes, NewParamChange(sdk.KeyMaxOpCUNumber, `"0"`))
	changes = append(changes, NewParamChange(sdk.KeyCollectThreshold, `"21000000000000000"`))
	tpcp := NewTokenParamsChangeProposal("Test", "Description", "btc", changes)

	require.Equal(t, "Test", tpcp.GetTitle())
	require.Equal(t, "Description", tpcp.GetDescription())
	require.Equal(t, RouterKey, tpcp.ProposalRoute())
	require.Equal(t, ProposalTypeTokenParamsChange, tpcp.ProposalType())
	require.Nil(t, tpcp.ValidateBasic())
	require.Equal(t, expectedStr, tpcp.String())

	//symbol is illegal
	tpcp = NewTokenParamsChangeProposal("Test", "Description", "Btc", changes)
	err := tpcp.ValidateBasic()
	require.NotNil(t, err)
	require.Equal(t, sdk.CodeInvalidSymbol, err.Code())

	//duplicated keys
	changes = append(changes, NewParamChange(sdk.KeyCollectThreshold, `"1000"`))
	tpcp = NewTokenParamsChangeProposal("Test", "Description", "btc", changes)
	err = tpcp.ValidateBasic()
	require.NotNil(t, err)
	require.Equal(t, CodeDuplicatedKey, err.Code())

}

func TestDisableTokenProposal(t *testing.T) {
	expectedStr := "Disable Token Proposal:\n Title:       Test\n Description: Description\n Symbol:      btc\n"
	dtp := NewDisableTokenProposal("Test", "Description", "btc")

	require.Equal(t, "Test", dtp.GetTitle())
	require.Equal(t, "Description", dtp.GetDescription())
	require.Equal(t, RouterKey, dtp.ProposalRoute())
	require.Equal(t, ProposalTypeDisableToken, dtp.ProposalType())
	require.Nil(t, dtp.ValidateBasic())
	require.Equal(t, expectedStr, dtp.String())

	dtp = NewDisableTokenProposal("Test", "Description", "Btc")
	require.NotNil(t, dtp.ValidateBasic())
}
