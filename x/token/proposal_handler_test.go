package token

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

func addProposal(symbol, chain string, tokenType sdk.TokenType) types.AddTokenProposal {
	return types.AddTokenProposal{
		Title:       "Test",
		Description: "description",
		TokenInfo: &sdk.IBCToken{
			BaseToken: sdk.BaseToken{
				Symbol:      sdk.Symbol(symbol),
				Issuer:      "",
				Chain:       sdk.Symbol(chain),
				SendEnabled: true,
				Decimals:    8,
				TotalSupply: sdk.NewIntWithDecimal(21, 15),
			},
			TokenType:          tokenType,
			DepositEnabled:     true,
			WithdrawalEnabled:  true,
			CollectThreshold:   sdk.NewIntWithDecimal(2, 4),   // btc
			OpenFee:            sdk.NewIntWithDecimal(28, 18), // nativeToken
			SysOpenFee:         sdk.NewIntWithDecimal(28, 18), // nativeToken
			WithdrawalFeeRate:  sdk.NewDecWithPrec(2, 0),      // gas * 10  btc
			MaxOpCUNumber:      10,
			SysTransferNum:     sdk.NewInt(3),
			OpCUSysTransferNum: sdk.NewInt(30),
			GasLimit:           sdk.NewInt(1),
			GasPrice:           sdk.NewInt(1000),
			DepositThreshold:   sdk.NewIntWithDecimal(2, 3),
		},
	}
}

func changeProposal(symbol string, changes []types.ParamChange) types.TokenParamsChangeProposal {
	return types.TokenParamsChangeProposal{
		Title:       "Test",
		Description: "description",
		Symbol:      symbol,
		Changes:     changes,
	}
}

func TestAddTokenProposalPassed(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk
	ctx.WithBlockHeight(10)

	tp := addProposal("ebtc", "eth", sdk.AccountBased)
	hdlr := NewTokenProposalHandler(keeper)
	res := hdlr(ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)
	require.Equal(t, 1, len(res.Events))
	require.Equal(t, types.EventTypeExecuteAddTokenProposal, res.Events[0].Type)

	tokenInfo := keeper.GetToken(ctx, parseSymbolFromProposalResp(res))
	require.Equal(t, tp.TokenInfo, tokenInfo)
}

func TestAddTokenProposalFailed(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk
	ctx.WithBlockHeight(10)

	//symbol != chain, chain does not exist
	tp := addProposal("obtc", "eos", sdk.AccountBased)
	hdlr := NewTokenProposalHandler(keeper)
	res := hdlr(ctx, tp)

	require.Equal(t, sdk.CodeInvalidSymbol, res.Code)
	require.Equal(t, 0, len(res.Events))
	require.Contains(t, res.Log, "chain eos does not exist")

	//duplicated adding
	tp = addProposal("xbtc", "eth", sdk.AccountBased)
	hdlr = NewTokenProposalHandler(keeper)
	res = hdlr(ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)
	require.Equal(t, 1, len(res.Events))
	require.Equal(t, types.EventTypeExecuteAddTokenProposal, res.Events[0].Type)
	sb := parseSymbolFromProposalResp(res)

	tokenInfo := keeper.GetToken(ctx, sb)
	require.Equal(t, tp.TokenInfo, tokenInfo)

	ctx.WithBlockHeight(20)
	res = hdlr(ctx, tp)
	require.Equal(t, sdk.CodeSymbolAlreadyExist, res.Code)
}

func TestTokenParamsChangeProposalPassed(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk
	ctx.WithBlockHeight(10)

	ap := addProposal("xbtc", "eth", sdk.AccountBased)
	hdlr := NewTokenProposalHandler(keeper)
	res := hdlr(ctx, ap)
	require.Equal(t, sdk.CodeOK, res.Code)
	require.Equal(t, 1, len(res.Events))
	require.Equal(t, types.EventTypeExecuteAddTokenProposal, res.Events[0].Type)
	sb := parseSymbolFromProposalResp(res)

	tokenInfo := keeper.GetIBCToken(ctx, parseSymbolFromProposalResp(res))
	require.Equal(t, ap.TokenInfo, tokenInfo)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	changes := []types.ParamChange{}
	changes = append(changes, types.NewParamChange(sdk.KeySendEnabled, "false"))
	changes = append(changes, types.NewParamChange(sdk.KeyDepositEnabled, "false"))
	changes = append(changes, types.NewParamChange(sdk.KeyWithdrawalEnabled, "false"))
	changes = append(changes, types.NewParamChange(sdk.KeyCollectThreshold, `"21000000000000000"`))
	changes = append(changes, types.NewParamChange(sdk.KeyDepositThreshold, `"11000000000000000"`))
	changes = append(changes, types.NewParamChange(sdk.KeyOpenFee, `"300"`))
	changes = append(changes, types.NewParamChange(sdk.KeySysOpenFee, `"4000"`))
	changes = append(changes, types.NewParamChange(sdk.KeyWithdrawalFeeRate, `"2"`))
	changes = append(changes, types.NewParamChange(sdk.KeyMaxOpCUNumber, `"60"`))
	changes = append(changes, types.NewParamChange(sdk.KeySysTransferNum, `"1"`))
	changes = append(changes, types.NewParamChange(sdk.KeyOpCUSysTransferNum, `"10"`))
	changes = append(changes, types.NewParamChange(sdk.KeyGasLimit, `"90000"`))

	cp := changeProposal(sb.String(), changes)
	res = hdlr(ctx, cp)
	require.Equal(t, sdk.CodeOK, res.Code)
	require.Equal(t, 1, len(res.Events))
	require.Equal(t, types.EventTypeExecuteTokenParamsChangeProposal, res.Events[0].Type)
	require.Equal(t, len(changes)*2, len(res.Events[0].Attributes))
	require.Equal(t, sdk.KeySendEnabled, string(res.Events[0].Attributes[0].Value))

	tokenInfo = keeper.GetIBCToken(ctx, sb)
	require.Equal(t, false, tokenInfo.SendEnabled)
	require.Equal(t, false, tokenInfo.DepositEnabled)
	require.Equal(t, false, tokenInfo.WithdrawalEnabled)
	require.Equal(t, sdk.NewInt(21000000000000000), tokenInfo.CollectThreshold)
	require.Equal(t, sdk.NewInt(11000000000000000), tokenInfo.DepositThreshold)
	require.Equal(t, sdk.NewInt(300), tokenInfo.OpenFee)
	require.Equal(t, sdk.NewInt(4000), tokenInfo.SysOpenFee)
	require.Equal(t, sdk.NewDec(2), tokenInfo.WithdrawalFeeRate)
	require.Equal(t, uint64(60), tokenInfo.MaxOpCUNumber)
	require.Equal(t, sdk.NewInt(90000000), tokenInfo.SysTransferAmount())
	require.Equal(t, sdk.NewInt(900000000), tokenInfo.OpCUSysTransferAmount())
	require.Equal(t, sdk.NewInt(90000), tokenInfo.GasLimit)
}

func TestTokenParamsChangeProposalFailed(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	//cdc := input.cdc
	keeper := input.tk
	ctx.WithBlockHeight(10)

	ap := addProposal("ebtc", "eth", sdk.AccountBased)
	hdlr := NewTokenProposalHandler(keeper)
	res := hdlr(ctx, ap)
	sb := parseSymbolFromProposalResp(res)

	require.Equal(t, sdk.CodeOK, res.Code)
	require.Equal(t, 1, len(res.Events))
	require.Equal(t, types.EventTypeExecuteAddTokenProposal, res.Events[0].Type)
	tokenInfo := keeper.GetToken(ctx, sb)
	require.Equal(t, ap.TokenInfo, tokenInfo)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	changes := []types.ParamChange{}
	changes = append(changes, types.NewParamChange("collect", `"21000000000000000"`))

	cp := changeProposal(sb.String(), changes)
	res = hdlr(ctx, cp)
	require.NotEqual(t, sdk.CodeOK, res.Code)
	require.Equal(t, 0, len(res.Events))

	//tokenInfo unchanged
	tokenInfo = keeper.GetIBCToken(ctx, sb)
	require.Equal(t, ap.TokenInfo, tokenInfo)
}

func parseSymbolFromProposalResp(res sdk.Result) sdk.Symbol {
	for _, event := range res.Events {
		if event.Type == types.EventTypeExecuteAddTokenProposal {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == types.AttributeKeyToken {
					return sdk.Symbol(attribute.Value)
				}
			}
		}
	}
	return sdk.Symbol("")
}
