package token

import (
	"testing"

	"github.com/hbtc-chain/bhchain/x/token/types"
	"github.com/stretchr/testify/assert"
)

func TestParamsEqual(t *testing.T) {
	p1 := types.DefaultParams()
	p2 := types.DefaultParams()
	assert.Equal(t, p1, p2)

	p1.TokenCacheSize += 10
	assert.NotEqual(t, p1, p2)
}

func TestGetSetParams(t *testing.T) {
	input := setupUnitTestEnv()
	ctx := input.ctx
	keeper := input.tk

	InitGenesis(ctx, keeper, DefaultGenesisState())
	genState1 := ExportGenesis(ctx, keeper)
	assert.Equal(t, DefaultGenesisState(), genState1)

	param1 := keeper.GetParams(ctx)
	assert.Equal(t, types.DefaultParams(), param1)

	//change tokenCacheSize
	param2 := param1
	param2.TokenCacheSize += 10

	keeper.SetParams(ctx, param2)
	param2Get := keeper.GetParams(ctx)
	assert.Equal(t, param2, param2Get)
	assert.True(t, param2.Equal(param2Get))
	assert.NotEqual(t, param1, param2Get)

	//append a new symbol
	param3 := param2
	param3.ReservedSymbols = append(param3.ReservedSymbols, "mytoken")
	keeper.SetParams(ctx, param3)
	param3Get := keeper.GetParams(ctx)
	assert.Equal(t, param3, param3Get)
	assert.True(t, param3.Equal(param3Get))
	assert.NotEqual(t, param2, param3Get)
	t.Logf("%v", param3.ReservedSymbols)

}
func TestParamString(t *testing.T) {
	expected := "Params:TokenCacheSize:32\tReservedSymbols:eos,usdt,bch,bsv,ltc,bnb,xrp,okb,ht,dash,etc,neo,atom,zec,ont,doge,tusd,bat,qtum,vsys,iost,dcr,zrx,beam,grin\t"
	param := DefaultParams()
	assert.Equal(t, expected, param.String())
}
