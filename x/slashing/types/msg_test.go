package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestMsgUnjailGetSignBytes(t *testing.T) {
	addr := sdk.CUAddress("abcd")
	msg := NewMsgUnjail(sdk.ValAddress(addr))
	bytes := msg.GetSignBytes()
	require.Equal(
		t,
		`{"type":"hbtcchain/MsgUnjail","value":{"address":"hbcvaloper1v93xxeq4ttg2c"}}`,
		string(bytes),
	)
}
