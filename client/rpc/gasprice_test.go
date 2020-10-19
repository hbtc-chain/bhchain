package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestMedian(t *testing.T) {
	desc := sdk.Decs{
		sdk.NewDec(2),
		sdk.NewDec(1),
	}
	m := sdk.DecMedian(desc)
	assert.True(t, sdk.NewDec(3).Quo(sdk.NewDec(2)).Equal(m))

	desc = sdk.Decs{
		sdk.NewDec(2),
		sdk.NewDec(3),
		sdk.NewDec(1),
	}
	m = sdk.DecMedian(desc)
	assert.True(t, sdk.NewDec(2).Equal(m))

	desc = sdk.Decs{
		sdk.NewDec(3),
	}
	m = sdk.DecMedian(desc)
	assert.True(t, sdk.NewDec(3).Equal(m))

	desc = sdk.Decs{
		sdk.NewDec(1),
		sdk.NewDec(1),
		sdk.NewDec(1),
		sdk.NewDec(1),
		sdk.NewDec(5),
		sdk.NewDec(6),
	}
	m = sdk.DecMedian(desc)
	assert.True(t, sdk.NewDec(1).Equal(m))
}
