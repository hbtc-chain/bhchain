package types

import (
	"github.com/stretchr/testify/require"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestValidateParams(t *testing.T) {
	params := DefaultParams()

	err := ValidateParams(params)
	require.Nil(t, err)

	params.Inflation = sdk.NewDec(0)
	err = ValidateParams(params)
	require.Nil(t, err)

	params.Inflation = sdk.NewDec(-1)
	err = ValidateParams(params)
	require.NotNil(t, err)

	params.Inflation = sdk.NewDec(1)
	err = ValidateParams(params)
	require.Nil(t, err)
}
