package types

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createUtxoIns(num uint64) UtxoIns {
	var utxoIn = []UtxoIn{}
	for i := uint64(0); i < num; i++ {
		utxo := NewUtxoIn(fmt.Sprintf("hash%v", i), i, NewInt(int64(10+i)), "")
		utxoIn = append(utxoIn, utxo)
	}
	return UtxoIns(utxoIn)
}

func TestUtxoTotal(t *testing.T) {
	utxoIn := createUtxoIns(10)
	assert.Equal(t, NewInt(145), utxoIn.UtxoInTotalAmount())
}

func TestUtxoInEqual(t *testing.T) {
	utxo5 := NewUtxoIn("hash5", 5, NewInt(5), "")
	utxo50 := NewUtxoIn("hash6", 5, NewInt(5), "")
	utxo51 := NewUtxoIn("hash5", 6, NewInt(5), "")
	utxo52 := NewUtxoIn("hash5", 5, NewInt(9), "")
	utxo53 := NewUtxoIn("hash5", 5, NewInt(5), "")
	utxo54 := NewUtxoIn("hash5", 5, NewInt(5), "address1")
	utxo55 := NewUtxoIn("hash5", 5, NewInt(5), "address1")
	utxo56 := NewUtxoIn("hash5", 5, NewInt(5), "address2")

	assert.False(t, utxo5.Equal(utxo50))
	assert.False(t, utxo5.Equal(utxo51))
	assert.False(t, utxo5.Equal(utxo52))
	assert.True(t, utxo5.Equal(utxo53))
	assert.True(t, utxo54.Equal(utxo55))
	assert.False(t, utxo54.Equal(utxo56))
}

func TestDuplicatedUtxoItems(t *testing.T) {
	var utxoIn = []UtxoIn{}
	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(9), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash5", 5, NewInt(5), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash4", 4, NewInt(4), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash1", 4, NewInt(10), ""))
	assert.Equal(t, false, UtxoIns(utxoIn).HasDuplicatedItem())

	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(8), ""))
	assert.Equal(t, false, UtxoIns(utxoIn).HasDuplicatedItem())

	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(9), ""))
	assert.Equal(t, true, UtxoIns(utxoIn).HasDuplicatedItem())
}

func TestDuplicatedUtxoItems2(t *testing.T) {
	var utxoIn = []UtxoIn{}
	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(9), "address1"))
	utxoIn = append(utxoIn, NewUtxoIn("hash5", 5, NewInt(5), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash4", 4, NewInt(4), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash1", 4, NewInt(10), ""))
	assert.Equal(t, false, UtxoIns(utxoIn).HasDuplicatedItem())

	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(8), ""))
	assert.Equal(t, false, UtxoIns(utxoIn).HasDuplicatedItem())

	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(9), ""))
	assert.False(t, UtxoIns(utxoIn).HasDuplicatedItem())

	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(9), "address2"))
	assert.False(t, UtxoIns(utxoIn).HasDuplicatedItem())

	utxoIn = append(utxoIn, NewUtxoIn("hash", 9, NewInt(9), "address1"))
	assert.True(t, UtxoIns(utxoIn).HasDuplicatedItem())
}

func TestRemoveOneUtxoIn(t *testing.T) {

	var utxoIn = []UtxoIn{}
	utxoIn = append(utxoIn, NewUtxoIn("hash9", 9, NewInt(9), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash5", 5, NewInt(5), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash4", 4, NewInt(4), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash10", 4, NewInt(10), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash20", 4, NewInt(20), ""))

	//delete NewUtxoIn("hash5", 5, NewInt(5))
	assert.True(t, UtxoIns(utxoIn).Has(NewUtxoIn("hash5", 5, NewInt(5), "")))
	found, utxoIn := UtxoIns(utxoIn).RemoveOneUtxoIn(NewUtxoIn("hash5", 5, NewInt(5), ""))
	assert.Equal(t, true, found)
	assert.False(t, UtxoIns(utxoIn).Has(NewUtxoIn("hash5", 5, NewInt(5), "")))

	found, utxoIn = UtxoIns(utxoIn).RemoveOneUtxoIn(NewUtxoIn("hash5", 5, NewInt(5), ""))
	assert.Equal(t, false, found)
	assert.False(t, UtxoIns(utxoIn).Has(NewUtxoIn("hash5", 5, NewInt(5), "")))

	found, utxoIn = UtxoIns(utxoIn).RemoveOneUtxoIn(NewUtxoIn("hash20", 5, NewInt(20), ""))
	assert.Equal(t, false, found)

	assert.True(t, UtxoIns(utxoIn).Has(NewUtxoIn("hash20", 4, NewInt(20), "")))
	found, utxoIn = UtxoIns(utxoIn).RemoveOneUtxoIn(NewUtxoIn("hash20", 4, NewInt(20), ""))
	assert.Equal(t, true, found)
	assert.False(t, UtxoIns(utxoIn).Has(NewUtxoIn("hash20", 4, NewInt(20), "")))

	// fmt.Printf("delete hash4\n")
	assert.True(t, UtxoIns(utxoIn).Has(NewUtxoIn("hash4", 4, NewInt(4), "")))
	found, utxoIn = UtxoIns(utxoIn).RemoveOneUtxoIn(NewUtxoIn("hash4", 4, NewInt(4), ""))
	assert.Equal(t, true, found)
	assert.Equal(t, 2, len(utxoIn))
	assert.NotContains(t, utxoIn, NewUtxoIn("hash4", 4, NewInt(4), ""))

}

func TestUtxosEqualOrSubset(t *testing.T) {
	var utxoIn1, utxoIn2 []UtxoIn
	utxo5 := NewUtxoIn("hash5", 5, NewInt(5), "")
	utxo9 := NewUtxoIn("hash9", 9, NewInt(9), "")
	utxo4 := NewUtxoIn("hash4", 4, NewInt(4), "")

	utxoIn1 = append(utxoIn1, utxo4, utxo9, utxo5)
	utxoIn2 = append(utxoIn2, utxo9, utxo5, utxo4)

	assert.True(t, UtxoIns(utxoIn1).IsSubsetOf(UtxoIns(utxoIn2)))
	assert.True(t, UtxoIns(utxoIn2).IsSubsetOf(UtxoIns(utxoIn1)))
	assert.True(t, UtxoIns(utxoIn2).Equal(UtxoIns(utxoIn1)))

	utxoIn2 = append(utxoIn2, NewUtxoIn("hash10", 4, NewInt(10), ""))
	assert.True(t, UtxoIns(utxoIn1).IsSubsetOf(UtxoIns(utxoIn2)))
	assert.False(t, UtxoIns(utxoIn2).IsSubsetOf(UtxoIns(utxoIn1)))
	assert.False(t, UtxoIns(utxoIn2).Equal(UtxoIns(utxoIn1)))

}

func TestRemoveUtxoIns(t *testing.T) {
	var utxoIn = []UtxoIn{}
	utxo5 := NewUtxoIn("hash5", 5, NewInt(5), "")
	utxo9 := NewUtxoIn("hash9", 9, NewInt(9), "")
	utxo4 := NewUtxoIn("hash4", 4, NewInt(4), "")

	utxoIn = append(utxoIn, utxo9)
	utxoIn = append(utxoIn, utxo5)
	utxoIn = append(utxoIn, utxo4)
	utxoIn = append(utxoIn, NewUtxoIn("hash10", 4, NewInt(10), ""))
	utxoIn = append(utxoIn, NewUtxoIn("hash20", 4, NewInt(20), ""))

	utxoIn = UtxoIns(utxoIn).RemoveUtxoIns(append([]UtxoIn{}, utxo5))

	assert.NotContains(t, utxoIn, utxo5)

	utxoIn = UtxoIns(utxoIn).RemoveUtxoIns(append([]UtxoIn{}, utxo9, utxo4))
	assert.Equal(t, 2, len(utxoIn))

	// not found, after that uxtoItems is emoty
	utxoIn = UtxoIns(utxoIn).RemoveUtxoIns(append([]UtxoIn{}, NewUtxoIn("hash22", 22, NewInt(22), "")))
	assert.Equal(t, 2, len(utxoIn))

	utxoIn = UtxoIns(utxoIn).RemoveUtxoIns(append([]UtxoIn{}, NewUtxoIn("hash10", 4, NewInt(10), "")))
	assert.Equal(t, 1, len(utxoIn))

	utxoIn = UtxoIns(utxoIn).RemoveUtxoIns(append([]UtxoIn{}, NewUtxoIn("hash20", 4, NewInt(20), "")))
	assert.Equal(t, 0, len(utxoIn))

	assert.True(t, UtxoIns(utxoIn).Empty())
}

func TestUtxoInsTotalAmount(t *testing.T) {
	var utxos = UtxoIns{}
	utxo5 := NewUtxoIn("hash5", 5, NewInt(5), "")
	utxo9 := NewUtxoIn("hash9", 9, NewInt(9), "")
	utxo4 := NewUtxoIn("hash4", 4, NewInt(4), "")
	utxo10 := NewUtxoIn("hash10", 4, NewInt(10), "")
	utxo20 := NewUtxoIn("hash20", 4, NewInt(20), "")

	utxos = append(utxos, utxo9)
	assert.Equal(t, NewInt(9), utxos.UtxoInTotalAmount())
	utxos = append(utxos, utxo5)
	assert.Equal(t, NewInt(14), utxos.UtxoInTotalAmount())
	utxos = append(utxos, utxo4)
	assert.Equal(t, NewInt(18), utxos.UtxoInTotalAmount())
	utxos = append(utxos, utxo10)
	assert.Equal(t, NewInt(28), utxos.UtxoInTotalAmount())
	utxos = append(utxos, utxo20)
	assert.Equal(t, NewInt(48), utxos.UtxoInTotalAmount())
}

func TestUtxoInsSort(t *testing.T) {
	var utxos = UtxoIns{}
	utxo5 := NewUtxoIn("hash5", 5, NewInt(5), "")
	utxo9 := NewUtxoIn("hash9", 9, NewInt(9), "")
	utxo4 := NewUtxoIn("hash4", 4, NewInt(4), "")
	utxo10 := NewUtxoIn("hash10", 4, NewInt(10), "")
	utxo20 := NewUtxoIn("hash20", 4, NewInt(20), "")

	utxos = append(utxos, utxo9, utxo20, utxo5, utxo4, utxo10)
	utxos.Sort()

	assert.Equal(t, utxo10, utxos[0])
	assert.Equal(t, utxo20, utxos[1])
	assert.Equal(t, utxo4, utxos[2])
	assert.Equal(t, utxo5, utxos[3])
	assert.Equal(t, utxo9, utxos[4])
}

func TestUtxoInsString(t *testing.T) {
	var utxos = UtxoIns{}
	utxo5 := NewUtxoIn("hash5", 5, NewInt(5), "address5")
	utxo9 := NewUtxoIn("hash9", 9, NewInt(9), "address9")
	utxo4 := NewUtxoIn("hash4", 4, NewInt(4), "address4")
	utxo10 := NewUtxoIn("hash10", 4, NewInt(10), "address10")
	utxo20 := NewUtxoIn("hash20", 4, NewInt(20), "address20")

	utxos = append(utxos, utxo9, utxo4, utxo20, utxo10, utxo5)
	utxos.Sort()

	out := ""
	for _, utxo := range utxos {
		out += utxo.String()
	}

	out1 := utxos.String()
	assert.Equal(t, out, out1)
}

func TestUtxoOutsTotalAmount(t *testing.T) {
	var utxos = UtxoOuts{}
	utxo5 := NewUtxoOut("outaddress5", NewInt(5))
	utxo9 := NewUtxoOut("outaddress9", NewInt(9))
	utxo4 := NewUtxoOut("outaddress4", NewInt(4))
	utxo10 := NewUtxoOut("outaddress10", NewInt(10))
	utxo20 := NewUtxoOut("outaddress20", NewInt(20))

	utxos = append(utxos, utxo9)
	assert.Equal(t, NewInt(9), utxos.UtxoOutTotalAmount())
	utxos = append(utxos, utxo5)
	assert.Equal(t, NewInt(14), utxos.UtxoOutTotalAmount())
	utxos = append(utxos, utxo4)
	assert.Equal(t, NewInt(18), utxos.UtxoOutTotalAmount())
	utxos = append(utxos, utxo10)
	assert.Equal(t, NewInt(28), utxos.UtxoOutTotalAmount())
	utxos = append(utxos, utxo20)
	assert.Equal(t, NewInt(48), utxos.UtxoOutTotalAmount())

}

func TestUtxoOutsString(t *testing.T) {
	var utxos = UtxoOuts{}
	utxo5 := NewUtxoOut("outaddress5", NewInt(5))
	utxo9 := NewUtxoOut("outaddress9", NewInt(9))
	utxo4 := NewUtxoOut("outaddress4", NewInt(4))
	utxo10 := NewUtxoOut("outaddress10", NewInt(10))
	utxo20 := NewUtxoOut("outaddress20", NewInt(20))

	utxos = append(utxos, utxo9, utxo5, utxo4, utxo10, utxo20)
	utxos.Sort()

	out := ""
	for _, utxo := range utxos {
		//	t.Logf("%v", utxo.String())
		out += utxo.String()
	}

	out1 := utxos.String()
	assert.Equal(t, out, out1)
}

func TestUtxoOutEqual(t *testing.T) {
	utxo5 := NewUtxoOut("outaddress5", NewInt(5))
	utxo51 := NewUtxoOut("outaddress5", NewInt(5))
	utxo52 := NewUtxoOut("outaddress52", NewInt(5))
	utxo53 := NewUtxoOut("outaddress5", NewInt(6))

	assert.True(t, utxo5.Equal(utxo51))
	assert.False(t, utxo5.Equal(utxo52))
	assert.False(t, utxo5.Equal(utxo53))
}

func TestUtxoOutsEqual(t *testing.T) {
	utxos := UtxoOuts{}
	utxo5 := NewUtxoOut("outaddress5", NewInt(5))
	utxo9 := NewUtxoOut("outaddress9", NewInt(9))
	utxo4 := NewUtxoOut("outaddress4", NewInt(4))
	utxo10 := NewUtxoOut("outaddress10", NewInt(10))
	utxo20 := NewUtxoOut("outaddress20", NewInt(20))

	utxos = append(utxos, utxo9, utxo5, utxo4, utxo10, utxo20)

	utxos1 := UtxoOuts{}
	utxos1 = append(utxos1, utxo10, utxo5, utxo4, utxo9, utxo20)
	assert.False(t, utxos.Equal(utxos1))

	utxos2 := UtxoOuts{}
	utxos2 = append(utxos2, utxo10, utxo10, utxo10, utxo10)
	utxos3 := UtxoOuts{}
	utxos3 = append(utxos3, utxo10, utxo10, utxo10)
	assert.False(t, utxos2.Equal(utxos3))

	utxos3 = append(utxos3, utxo10)
	assert.True(t, utxos2.Equal(utxos3))

	utxos4 := UtxoOuts{}
	utxos4 = append(utxos4, utxo4, utxo5)
	utxos5 := UtxoOuts{}
	utxos5 = append(utxos5, utxo5, utxo4)

	assert.False(t, utxos4.Equal(utxos5))
	assert.True(t, utxos4.Equal(UtxoOuts{utxo4, utxo5}))
	assert.True(t, utxos5.Equal(UtxoOuts{utxo5, utxo4}))

}

func TestUtxoOutsHas(t *testing.T) {
	utxos := UtxoOuts{}
	utxo5 := NewUtxoOut("outaddress5", NewInt(5))
	utxo9 := NewUtxoOut("outaddress9", NewInt(9))
	utxo4 := NewUtxoOut("outaddress4", NewInt(4))
	utxo10 := NewUtxoOut("outaddress10", NewInt(10))
	utxo20 := NewUtxoOut("outaddress20", NewInt(20))

	utxos = append(utxos, utxo5, utxo4, utxo10, utxo20)

	assert.True(t, utxos.Has(utxo5))
	assert.False(t, utxos.Has(utxo9))

}
