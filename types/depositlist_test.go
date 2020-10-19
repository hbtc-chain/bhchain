package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDepositItem(t *testing.T) {
	// error
	d, err := NewDepositItem("", 0, NewInt(0), "", "memo", 0)
	assert.NotNil(t, err)
	assert.Equal(t, DepositItem{}, d)

	// ok
	d, err = NewDepositItem("hashtest", 2, NewInt(3), "", "memo", 0)
	assert.Nil(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, "hashtest", d.Hash)
	assert.EqualValues(t, 2, d.Index)
	assert.EqualValues(t, NewInt(3), d.Amount)
	assert.EqualValues(t, "memo", d.Memo)
}

func TestNewDepositItemList(t *testing.T) {
	errTestCases := [][]DepositItem{
		[]DepositItem{
			{},
			{},
		},
		[]DepositItem{
			{},
			{"hash1", 1, NewInt(100), "", "", 0},
		},
		[]DepositItem{
			{"", 1, NewInt(100), "", "", 0},
			{"hash1", 1, NewInt(100), "", "", 0},
		},
		// hash1==hash2 && index1== index2 is not ok
		[]DepositItem{
			{"hash1", 1, NewInt(100), "", "", 0},
			{"hash1", 1, NewInt(200), "", "", 0},
		},
	}
	for _, errCase := range errTestCases {
		depositList, err := NewDepositList(errCase...)
		assert.NotNil(t, err)
		assert.Nil(t, depositList)
	}

	okTestCases := [][]DepositItem{
		[]DepositItem{
			{Hash: "hash1", Index: 1},
			{"hash1", 2, NewInt(200), "", "", 0},
		},
		// hash1==hash2 && index1!= index2 is ok
		[]DepositItem{
			{"hash1", (1), NewInt(100), "", "", 0},
			{"hash1", 2, NewInt(200), "", "", 0},
		},
		[]DepositItem{
			{"hash1", 0, NewInt(200), "", "", 0},
			{"hash2", 2, NewInt(200), "", "", 0},
		},
		[]DepositItem{
			{"hash1", 1, NewInt(100), "", "", 0},
			{"hash2", 2, NewInt(0), "", "", 0},
		},
	}
	for _, okCase := range okTestCases {
		depositList, err := NewDepositList(okCase...)
		assert.Nil(t, err)
		assert.NotNil(t, depositList)
	}

}

func TestDepositList_CRUD(t *testing.T) {
	// create
	dls, err := NewDepositList()
	assert.Nil(t, err)
	assert.Nil(t, dls)
	d1, _ := NewDepositItem("hash1", 1, NewInt(100), "", "", 0)
	dls, err = NewDepositList(d1)
	assert.Nil(t, err)
	assert.NotNil(t, dls)

	// add
	d2, _ := NewDepositItem("hash2", 2, NewInt(200), "", "", 0)
	err = dls.AddDepositItem(d2)
	assert.Nil(t, err)
	assert.Equal(t, 2, dls.Len())
	// add duplicate deposit
	err = dls.AddDepositItem(d2)
	assert.NotNil(t, err)
	assert.Equal(t, 2, dls.Len())

	d3, _ := NewDepositItem("hash2", 3, NewInt(200), "", "", 0)
	err = dls.AddDepositItem(d3)
	assert.Nil(t, err)
	assert.Equal(t, 3, dls.Len())

	// get
	d2Got, i := dls.GetDepositItem(d2.Hash, d2.Index)
	assert.Equal(t, 1, i)
	assert.EqualValues(t, d2, d2Got, dls[1])

	// del,remove d3 ,should d1,d2, in dls
	err = dls.RemoveDepositItem(d2.GetHash(), d2.GetIndex())
	assert.Nil(t, err)
	assert.Equal(t, 2, dls.Len())

}

func TestFilter(t *testing.T) {
	d1, _ := NewDepositItem("hash1", 1, NewInt(100), "", "", 0)
	d2, _ := NewDepositItem("hash2", 2, NewInt(20), "", "", 0)
	d3, _ := NewDepositItem("hash3", 3, NewInt(200), "", "", 0)
	depositList, _ := NewDepositList(d1, d2, d3)

	dlsGot := depositList.Filter(func(d DepositItem) bool {
		return d.Hash == "hash1"
	})
	assert.Equal(t, 1, dlsGot.Len())

	dlsGot = depositList.Filter(func(d DepositItem) bool {
		return d.Amount.Equal(NewInt(100))
	})
	assert.Equal(t, 1, dlsGot.Len())

	dlsGot = depositList.Filter(func(d DepositItem) bool {
		return d.Memo == ""
	})
	assert.Equal(t, 3, dlsGot.Len())
	dlsGot = depositList.Filter(func(d DepositItem) bool {
		return d.Index == 1 && d.Hash == "hash1"
	})
	assert.Equal(t, 1, dlsGot.Len())

}

func TestDepositList_SumByStatus(t *testing.T) {
	// empty depositList
	depositList, _ := NewDepositList()
	assert.EqualValues(t, NewInt(0), depositList.Sum())
	assert.EqualValues(t, NewInt(0), depositList.SumByStatus(0))
	assert.EqualValues(t, NewInt(0), depositList.SumByStatus(1))

	// 1 item in depositList
	d1, _ := NewDepositItem("hash1", 1, NewInt(100), "", "", 0)
	depositList, _ = NewDepositList(d1)
	assert.EqualValues(t, d1.Amount, depositList.Sum())
	assert.EqualValues(t, d1.Amount, depositList.SumByStatus(0))
	assert.EqualValues(t, NewInt(0), depositList.SumByStatus(1))

	// 3 item in depositList
	d2, _ := NewDepositItem("hash2", 2, NewInt(20), "", "", 1)
	d3, _ := NewDepositItem("hash3", 3, NewInt(200), "", "", 0)
	depositList, _ = NewDepositList(d1, d2, d3)

	assert.EqualValues(t, d1.Amount.Add(d2.Amount).Add(d3.Amount), depositList.Sum())
	assert.EqualValues(t, d1.Amount.Add(d3.Amount), depositList.SumByStatus(0))
	assert.EqualValues(t, d2.Amount, depositList.SumByStatus(1))
	assert.EqualValues(t, NewInt(0), depositList.SumByStatus(2))
}

func TestDepositListSortByAmount(t *testing.T) {
	d1, _ := NewDepositItem("hash1", 1, NewInt(100), "", "", 0)
	d2, _ := NewDepositItem("hash2", 2, NewInt(20), "", "", 0)
	d3, _ := NewDepositItem("hash3", 3, NewInt(200), "", "", 0)
	depositList, _ := NewDepositList(d1, d2, d3)

	depositList.SortByAmount()
	assert.Equal(t, d2, depositList[0])
	assert.Equal(t, d1, depositList[1])
	assert.Equal(t, d3, depositList[2])

	depositList.SortByAmountDesc()
	assert.Equal(t, d2, depositList[2])
	assert.Equal(t, d1, depositList[1])
	assert.Equal(t, d3, depositList[0])
}
