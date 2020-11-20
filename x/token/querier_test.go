package token

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

func TestQuery(t *testing.T) {
	input := setupUnitTestEnv(true)
	ctx := input.ctx
	keeper := input.tk
	cdc := input.cdc

	for s := range TestTokenData {
		symbol := s
		keeper.SetToken(ctx, newTokenInfo(symbol))
	}

	req := abci.RequestQuery{
		Path: fmt.Sprintf("token/%s/%s", types.QuerierRoute, types.QueryToken),
		Data: nil,
	}

	for s := range TestIBCTokens {
		symbol := string(s)

		bz, err := cdc.MarshalJSON(QueryTokenInfoParams{symbol})
		assert.NoError(t, err)

		req.Data = bz
		bz, err = queryToken(ctx, req, keeper)
		assert.Nil(t, err)

		var res ResToken
		keeper.cdc.MustUnmarshalJSON(bz, &res)
		compareQueryRes(t, symbol, res)
	}

	bz, err := cdc.MarshalJSON(QueryTokenInfoParams{"BHEMN"})
	assert.NoError(t, err)
	req.Data = bz
	_, err = queryToken(ctx, req, keeper)
	assert.NotNil(t, err)

}

func TestQueryIBCTokens(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk
	// set token info
	for _, s := range TestIBCTokens {
		keeper.SetToken(ctx, newTokenInfo(s.Symbol))
	}

	//queryTokens
	bz, err := queryIBCTokens(ctx, keeper)
	assert.Nil(t, err)

	// Unmarshal tokeninfos
	var res []ResToken
	keeper.cdc.MustUnmarshalJSON(bz, &res)
	// check tokeninfos
	for _, s := range res {
		compareQueryRes(t, s.Symbol, s)
	}
}

func compareQueryRes(t *testing.T, s string, res ResToken) {
	symbol := sdk.Symbol(s)
	testToken := TestIBCTokens[symbol]
	assert.Equal(t, testToken.Issuer, res.Issuer)
	assert.Equal(t, testToken.Chain.String(), res.Chain)
	assert.Equal(t, testToken.TokenType, res.TokenType)
	assert.Equal(t, testToken.SendEnabled, res.SendEnabled)
	assert.Equal(t, testToken.DepositEnabled, res.DepositEnabled)
	assert.Equal(t, testToken.WithdrawalEnabled, res.WithdrawalEnabled)
	assert.Equal(t, testToken.Decimals, res.Decimals)
	assert.True(t, testToken.TotalSupply.Equal(res.TotalSupply))
	assert.True(t, testToken.CollectThreshold.Equal(res.CollectThreshold))
	assert.True(t, testToken.OpenFee.Equal(res.OpenFee))
	assert.True(t, testToken.SysOpenFee.Equal(res.SysOpenFee))
	assert.True(t, testToken.WithdrawalFeeRate.Equal(res.WithdrawalFeeRate))
	assert.Equal(t, testToken.Symbol.String(), res.Symbol)
	assert.Equal(t, testToken.DepositThreshold, res.DepositThreshold)
	assert.Equal(t, testToken.MaxOpCUNumber, res.MaxOpCUNumber)
	assert.Equal(t, testToken.GasLimit, res.GasLimit)
	assert.Equal(t, testToken.OpCUSysTransferNum, res.OpCUSysTransferNum)
	assert.Equal(t, testToken.SysTransferNum, res.SysTransferNum)
}
