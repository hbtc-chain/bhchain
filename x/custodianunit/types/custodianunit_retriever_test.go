package types

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/hbtc-chain/bhchain/tests/mocks"
)

var dummyError = errors.New("dummy")

func TestCURetriever(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNodeQuerier := mocks.NewMockNodeQuerier(mockCtrl)
	cuRetr := NewCURetriever(mockNodeQuerier)
	addr := []byte("test")
	bs, err := ModuleCdc.MarshalJSON(NewQueryCUParams(addr))
	require.NoError(t, err)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq("custom/cu/CU"),
		gomock.Eq(bs)).Return(nil, int64(0), dummyError).Times(1)
	_, err = cuRetr.GetCU(addr)
	require.Error(t, err)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq("custom/cu/CU"),
		gomock.Eq(bs)).Return(nil, int64(0), dummyError).Times(1)
	s, err := cuRetr.GetSequence(addr)
	require.Error(t, err)
	require.Equal(t, uint64(0), s)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq("custom/cu/CU"),
		gomock.Eq(bs)).Return(nil, int64(0), dummyError).Times(1)
	require.Error(t, cuRetr.EnsureExists(addr))
}
