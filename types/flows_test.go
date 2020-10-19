package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbtc-chain/bhchain/bhexapp"
	. "github.com/hbtc-chain/bhchain/types"
)

func TestDecodeData(t *testing.T) {
	cdc := bhexapp.MakeCodec()
	marshalled := cdc.MustMarshalBinaryLengthPrefixed(Receipt{
		Category: 0,
		Flows: []Flow{
			OrderFlow{
				Symbol:      "eth",
				CUAddress:   nil,
				OrderID:     "o",
				OrderType:   0,
				OrderStatus: 0,
			},
		},
	})

	var bz []byte
	// nil
	receipts, err, _ := GetReceiptFromData(cdc, bz)
	require.Error(t, err)

	// empty
	bz = make([]byte, 0)
	receipts, err, _ = GetReceiptFromData(cdc, bz)
	require.NoError(t, err)
	require.Len(t, receipts, 0)

	// one
	bz = append(bz, marshalled...)
	receipts, err, _ = GetReceiptFromData(cdc, bz)
	require.NoError(t, err)
	require.Len(t, receipts, 1)

	// two
	bz = append(bz, marshalled...)
	receipts, err, _ = GetReceiptFromData(cdc, bz)
	require.NoError(t, err)
	require.Len(t, receipts, 2)

	// three
	bz = append(bz, marshalled...)
	receipts, err, _ = GetReceiptFromData(cdc, bz)
	require.NoError(t, err)
	require.Len(t, receipts, 3)

	// invalid data
	receipts, err, _ = GetReceiptFromData(cdc, []byte("abc"))
	require.Error(t, err)
	require.Len(t, receipts, 0)

	// skip data of other type
	bz = append(cdc.MustMarshalBinaryLengthPrefixed("id"), bz...)
	receipts, err, _ = GetReceiptFromData(cdc, bz)
	require.NoError(t, err)
	require.Len(t, receipts, 3)

	// mix invalid with valid
	bz = append([]byte("abc"), bz...)
	receipts, err, _ = GetReceiptFromData(cdc, bz)
	require.Error(t, err)
	require.Len(t, receipts, 0)
}
