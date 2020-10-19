package types

import (
	"fmt"
	"sort"
	"strings"
)

type AddressType uint64

const (
	CustodianUnitAddressType AddressType = 0x1 //CustodianUnitAddress
	DepositAddressType       AddressType = 0x2 //Deposit Address
	WithdrawalAddressType    AddressType = 0x3 //WithdrawalAddress Address
)

const (
	MaxVinNum  = 6
	MaxVoutNum = 6
)

func IsAddressTypeLegal(addrType AddressType) bool {
	if addrType >= CustodianUnitAddressType && addrType <= WithdrawalAddressType {
		return true
	}
	return false
}

//UtxoIn defines UtxoIn in address module
type UtxoIn struct {
	Hash    string `json:"hash"`
	Index   uint64 `json:"index"`
	Amount  Int    `json:"amount"`
	Address string `json:"address"`
}

//UtxoOut defines UtxoOut ,only valid for utxobased
type UtxoOut struct {
	Address string `json:"address"`
	Amount  Int    `json:"amount"`
}

type UtxoIns []UtxoIn
type UtxoOuts []UtxoOut

// TODO add address
//NewUtxoIn creates a new UtxoIn
func NewUtxoIn(hash string, index uint64, amount Int, address string) UtxoIn {
	return UtxoIn{
		Hash:    hash,
		Index:   index,
		Amount:  amount,
		Address: address,
	}
}

//NewUtxoOut create a UtxoOut
func NewUtxoOut(addr string, amount Int) UtxoOut {
	return UtxoOut{
		Address: addr,
		Amount:  amount,
	}
}

func (u UtxoIn) Equal(other UtxoIn) bool {
	if u.Hash != other.Hash ||
		u.Index != other.Index || !(u.Amount.Equal(other.Amount)) ||
		u.Address != other.Address {
		return false
	}
	return true
}

//String
func (u UtxoIn) String() string {
	out := fmt.Sprintf("hash:%v index:%v amount:%v address:%v\n",
		u.Hash, u.Index, u.Amount, u.Address)
	return strings.TrimSpace(out)
}

//Equal check whether two utxoins same
func (us UtxoIns) Equal(others UtxoIns) bool {
	if us.IsSubsetOf(others) && others.IsSubsetOf(us) {
		return true
	}
	return false
}

//IsSubsetOf check whether us is subset of others
func (us UtxoIns) IsSubsetOf(others UtxoIns) bool {
	for _, u := range us {
		found := false
		for _, o := range others {
			if u.Equal(o) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

//UtxoInTotalAmount return the us's total amount
func (us UtxoIns) UtxoInTotalAmount() Int {
	var total = NewInt(0)
	for _, utxo := range us {
		total = total.Add(utxo.Amount)
	}
	return total
}

//HasDuplicatedItem check whether duplication items exist in utxoins
func (us UtxoIns) HasDuplicatedItem() bool {
	uxtoInMap := map[string]UtxoIn{}
	for _, u := range us {
		h, ok := uxtoInMap[u.Hash]
		if !ok {
			uxtoInMap[u.Hash] = u
		} else {
			//Check index and Amount in case of same hash
			if (u.Index == h.Index) && (u.Amount.Equal(h.Amount)) && (u.Address == h.Address) {
				return true
			}
		}
	}
	return false
}

// Has check whether a utxo in utxoIn
func (us UtxoIns) Has(utxo UtxoIn) bool {
	for _, u := range us {
		if u.Equal(utxo) {
			return true
		}
	}
	return false
}

//RemoveOneUtxoIn remove a utxo from utxoIns, then return the remained utxoIns
func (us UtxoIns) RemoveOneUtxoIn(utxo UtxoIn) (bool, UtxoIns) {
	found := false
	var i int
	var u UtxoIn
	for i, u = range us {
		if u.Equal(utxo) {
			found = true
			break
		}
	}
	if found {
		return found, append(us[:i], us[i+1:]...)
	}
	return false, us
}

func (us UtxoIns) PrintUtxoIn() {
	for i, utxo := range us {
		fmt.Printf("%v utxo:%v\n ", i, utxo)
	}
}

//Strings
func (us UtxoIns) String() string {
	out := ""
	for _, u := range us {
		out += u.String()

	}
	return strings.TrimSpace(out)
}

//RemoveUtxoIns
func (us UtxoIns) RemoveUtxoIns(delete UtxoIns) UtxoIns {
	utxos := make([]UtxoIn, len(us))
	copy(utxos, []UtxoIn(us))
	for _, d := range delete {
		_, utxos = UtxoIns(utxos).RemoveOneUtxoIn(d)
	}
	return UtxoIns(utxos)
}

//Empty check whether the UtxoIns is empty
func (us UtxoIns) Empty() bool {
	if len(us) == 0 {
		return true
	}
	return false
}

func (us UtxoIns) IsValid() bool {
	if us.HasDuplicatedItem() {
		return false
	}
	for _, u := range us {
		if !u.Amount.IsPositive() || u.Address == "" || u.Hash == "" || u.Index < 0 {
			return false
		}

	}
	return true
}

type SortUtxoIns UtxoIns

func (s SortUtxoIns) Len() int           { return len(s) }
func (s SortUtxoIns) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortUtxoIns) Less(i, j int) bool { return (s[i].Hash < s[j].Hash) }

//Sort the utxo by hash
func (us UtxoIns) Sort() {
	sort.Sort(SortUtxoIns(us))
}

func (u UtxoOut) String() string {
	out := fmt.Sprintf("address:%v amount:%v\n",
		u.Address, u.Amount)
	return strings.TrimSpace(out)
}

func (u UtxoOut) Equal(other UtxoOut) bool {
	if u.Address != other.Address || !u.Amount.Equal(other.Amount) {
		return false
	}
	return true
}

func (us UtxoOuts) Has(utxoOut UtxoOut) bool {
	for _, u := range us {
		if u.Equal(utxoOut) {
			return true
		}
	}
	return false
}

func (us UtxoOuts) Equal(others UtxoOuts) bool {
	if len(us) != len(others) {
		return false
	}

	for i := 0; i < len(us); i++ {
		if !us[i].Equal(others[i]) {
			return false
		}
	}

	return true
}

//UtxoOutTotalAmount calculate UtxoOut total amount
func (us UtxoOuts) UtxoOutTotalAmount() Int {
	var total = NewInt(0)
	for _, utxo := range us {
		total = total.Add(utxo.Amount)
	}
	return total
}

func (us UtxoOuts) String() string {
	out := ""
	for _, u := range us {
		out += u.String()
	}
	return strings.TrimSpace(out)
}

type SortUtxoOuts UtxoOuts

func (s SortUtxoOuts) Len() int           { return len(s) }
func (s SortUtxoOuts) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortUtxoOuts) Less(i, j int) bool { return (s[i].Address < s[j].Address) }

func (us UtxoOuts) Sort() {
	sort.Sort(SortUtxoOuts(us))
}
