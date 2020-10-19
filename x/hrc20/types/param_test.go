package types

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	p2 := DefaultParams()
	assert.Equal(t, p1, p2)

	p1.IssueTokenFee = p1.IssueTokenFee.SubRaw(10)
	assert.NotEqual(t, p1, p2)
}

func TestParamsString(t *testing.T) {
	expectedStr := "Params: \nOpenTokenFee: 1000000000000000000\n"
	assert.Equal(t, expectedStr, DefaultParams().String())
}

func TestParamValidate(t *testing.T) {
	p := DefaultParams()
	assert.Nil(t, p.Validate())

	p.IssueTokenFee = sdk.ZeroInt()
	assert.NotNil(t, p.Validate())

	p.IssueTokenFee = sdk.NewInt(-1)

	assert.NotNil(t, p.Validate())

}
