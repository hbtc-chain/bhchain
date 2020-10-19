package token

import (
	"fmt"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestQuery(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk
	cdc := input.cdc

	for s := range TestTokenData {
		symbol := s
		keeper.SetTokenInfo(ctx, newTokenInfo(symbol))
	}

	req := abci.RequestQuery{
		Path: fmt.Sprintf("token/%s/%s", types.QuerierRoute, types.QueryToken),
		Data: nil,
	}

	for s := range TestTokenData {
		symbol := string(s)

		bz, err := cdc.MarshalJSON(QueryTokenInfo{symbol})
		assert.NoError(t, err)

		req.Data = bz
		bz, err = queryToken(ctx, req, keeper)
		assert.Nil(t, err)

		var res QueryResToken
		keeper.cdc.MustUnmarshalJSON(bz, &res)
		compareQueryRes(t, symbol, res)
	}

	for s := range TestTokenData {
		symbol := string(s)

		bz, err := cdc.MarshalJSON(QueryDecimals{symbol})
		assert.NoError(t, err)

		req.Data = bz
		bz, err = queryDecimal(ctx, req, keeper)
		assert.Nil(t, err)

		var res QueryResDecimals
		keeper.cdc.MustUnmarshalJSON(bz, &res)
		assert.Equal(t, TestTokenData[s].Decimals, res.Decimals)

	}

	bz, err := cdc.MarshalJSON(QueryTokenInfo{"BHEMN"})
	assert.NoError(t, err)
	req.Data = bz
	_, err = queryToken(ctx, req, keeper)
	assert.NotNil(t, err)

	//err
	bz, err = querySymbols(ctx, keeper)
	assert.Nil(t, err)
	var symbols QueryResSymbols
	keeper.cdc.MustUnmarshalJSON(bz, &symbols)
	for symbol := range TestTokenData {
		assert.Contains(t, symbols, symbol.String())
	}
}

func TestQueryTokens(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk
	// set token info
	for _, s := range TestTokenData {
		keeper.SetTokenInfo(ctx, newTokenInfo(s.Symbol))
	}

	//queryTokens
	bz, err := queryTokens(ctx, keeper)
	assert.Nil(t, err)

	// Unmarshal tokeninfos
	var res []sdk.TokenInfo
	keeper.cdc.MustUnmarshalJSON(bz, &res)
	//// check tokeninfos
	//for _, s := range res {
	//	compareQueryRes(t, s.Symbol.String(), s)
	//}
}

func compareQueryRes(t *testing.T, s string, res QueryResToken) {
	symbol := sdk.Symbol(s)
	testToken := TestTokenData[symbol]
	assert.Equal(t, testToken.Issuer, res.Issuer)
	assert.Equal(t, testToken.Chain.String(), res.Chain)
	assert.Equal(t, uint64(testToken.TokenType), res.TokenType)
	assert.Equal(t, testToken.IsSendEnabled, res.IsSendEnabled)
	assert.Equal(t, testToken.IsDepositEnabled, res.IsDepositEnabled)
	assert.Equal(t, testToken.IsWithdrawalEnabled, res.IsWithdrawalEnabled)
	assert.Equal(t, testToken.Decimals, res.Decimals)
	assert.True(t, testToken.TotalSupply.Equal(res.TotalSupply))
	assert.True(t, testToken.CollectThreshold.Equal(res.CollectThreshold))
	assert.True(t, testToken.OpenFee.Equal(res.OpenFee))
	assert.True(t, testToken.SysOpenFee.Equal(res.SysOpenFee))
	assert.True(t, testToken.WithdrawalFeeRate.Equal(res.WithdrawalFeeRate))
	assert.Equal(t, testToken.Symbol, res.Symbol)
	assert.Equal(t, testToken.DepositThreshold, res.DepositThreshold)
	assert.Equal(t, testToken.MaxOpCUNumber, res.MaxOpCUNumber)
	assert.Equal(t, testToken.GasLimit, res.GasLimit)
	assert.Equal(t, testToken.OpCUSysTransferNum, res.OpCUSysTransferNum)
	assert.Equal(t, testToken.SysTransferNum, res.SysTransferNum)
}
