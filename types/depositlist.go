package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type DepositItemStatus uint16

// state transform flow:
//               CollectWaitSign                CollectFinish
// UnCollected -------------------> InProcess -----------------> Confirmed
const (
	DepositItemStatusUnCollected DepositItemStatus = 0x0
	DepositItemStatusWaitCollect DepositItemStatus = 0x1
	DepositItemStatusConfirmed   DepositItemStatus = 0x2
	DepositItemStatusInProcess   DepositItemStatus = 0x3
)

// In check if current status is in the list
func (s DepositItemStatus) In(l []DepositItemStatus) bool {
	for _, status := range l {
		if status == s {
			return true
		}
	}
	return false
}

// ---------------------------------------------------
// DepositList
type DepositList []DepositItem
type depositItemSortByAmount DepositList

func NewDepositList(deposits ...DepositItem) (DepositList, error) {
	newdeposits := DepositList(deposits)
	for _, d := range newdeposits {
		if !d.IsValid() {
			return nil, errors.New(fmt.Sprintf("invalid deposit %v", d))
		}
	}
	if i := findDupDeposit(newdeposits); i != -1 {
		return nil, errors.New(fmt.Sprintf("duplicate deposit %v", newdeposits[i]))
	}
	return newdeposits, nil
}

func (dts *DepositList) AddDepositItem(d DepositItem) error {
	if !d.IsValid() {
		return errors.New("invalid Deposit")
	}

	if _, i := dts.GetDepositItem(d.Hash, d.Index); i > -1 {
		return errors.New("deposit already exist")
	}
	*dts = append(*dts, d)
	return nil
}

// DelDeposit remove Deposit from list &  list len() -1
func (dts *DepositList) RemoveDepositItem(hash string, index uint64) error {
	_, i := dts.GetDepositItem(hash, index)
	if i < 0 {
		return errors.New(fmt.Sprintf("deposit not exist %v %v", hash, index))
	}
	*dts = append((*dts)[:i], (*dts)[i+1:]...)
	return nil
}

//// UpdateDeposit  update the deposit in depositList which have same hash & index
func (dts *DepositList) UpdateDepositItem(d DepositItem) error {
	_, i := dts.GetDepositItem(d.Hash, d.Index)
	if i < 0 {
		return errors.New(fmt.Sprintf("deposit not exist %v %v", d.Hash, d.Index))
	}
	(*dts)[i] = d
	return nil
}

// GetDeposit 返回 Deposit 和在数组中的下标，不存在返回空和 -1
func (dts *DepositList) GetDepositItem(hash string, index uint64) (DepositItem, int) {
	for i, dt := range *dts {
		if dt.Hash == hash && dt.Index == index {
			return dt, i
		}
	}
	return DepositItem{}, -1
}

func (dts *DepositList) Sum() Int {
	sum := ZeroInt()
	for _, dt := range *dts {
		sum = sum.Add(dt.Amount)
	}
	return sum
}

func (dts *DepositList) SumByStatus(status DepositItemStatus) Int {
	dlsGot := dts.Filter(func(d DepositItem) bool {
		return d.Status == status
	})
	return dlsGot.Sum()
}

func (dts *DepositList) Filter(filterFunc func(DepositItem) bool) DepositList {
	dls := make([]DepositItem, 0, len(*dts))
	for _, dl := range *dts {
		if filterFunc(dl) {
			dls = append(dls, dl)
		}
	}
	return dls
}

func (dts *DepositList) String() string {
	var b strings.Builder
	for _, s := range *dts {
		b.WriteString(s.String())
	}
	return b.String()
}

// MarshalJSON implements a custom JSON marshaller for the DepositList type to allow
// nil DepositList to be encoded as an empty array.
/* func (dts *DepositList) MarshalJSON() ([]byte, error) { */
// if dts == nil {
// return json.Marshal(DepositList{})
// }

// return json.Marshal(dts)
/* } */

// ---------------------------------------------------
// Deposit

type DepositItem struct {
	Hash       string            `json:"hash"`
	Index      uint64            `json:"index"`
	Amount     Int               `json:"amount"`
	ExtAddress string            `json:"ext_address"`
	Memo       string            `json:"memo"`
	Status     DepositItemStatus `json:"status"`
}

var DepositNil = DepositItem{}

func NewDepositItem(hash string, index uint64, amount Int, extAddress, memo string, status DepositItemStatus) (DepositItem, error) {
	if hash == "" {
		return DepositItem{}, errors.New("hash is empty")
	}

	return DepositItem{Hash: hash, Index: index, Amount: amount, ExtAddress: extAddress, Memo: memo, Status: status}, nil
}

func (d DepositItem) GetHash() string {
	return d.Hash
}

func (d DepositItem) GetIndex() uint64 {
	return d.Index
}

func (d DepositItem) GetStatus() DepositItemStatus {
	return d.Status
}

func (d DepositItem) IsValid() bool {
	if d.Hash == "" {
		return false
	}
	return true
}

func (d DepositItem) String() string {
	return fmt.Sprintf("%s %v %v %s %s %v\n", d.Hash, d.Index, d.Amount, d.ExtAddress, d.Memo, d.Status)
}

//-----------------------------------------------------------------------------
// Sort interface

//nolint
func (dts DepositList) Len() int { return len(dts) }
func (dts DepositList) Less(i, j int) bool {
	return dts[i].Hash+string(dts[i].Index) < dts[j].Hash+string(dts[i].Index)
}
func (dts DepositList) Swap(i, j int) {
	dts[i], dts[j] = dts[j], dts[i]
}

func (dts depositItemSortByAmount) Len() int { return len(dts) }
func (dts depositItemSortByAmount) Less(i, j int) bool {
	return dts[i].Amount.LT(dts[j].Amount)
}
func (dts depositItemSortByAmount) Swap(i, j int) {
	dts[i], dts[j] = dts[j], dts[i]
}

var _ sort.Interface = DepositList{}
var _ sort.Interface = depositItemSortByAmount{}

// Sort is a helper function to sort the set of depositList inplace
// sort by hash & index
func (dts DepositList) Sort() DepositList {
	sort.Sort(dts)
	return dts
}

func (dts DepositList) SortByAmount() {
	sort.Sort(depositItemSortByAmount(dts))
}

func (dts DepositList) SortByAmountDesc() {
	sort.Sort(sort.Reverse(depositItemSortByAmount(dts)))
}

// -----------------------------------------------------------------------------
// Utils

// findDupDeposit 查找hash & index相等的数组成员
// 返回第一个重复对象下标，没有重复返回 -1
func findDupDeposit(dls DepositList) int {
	// sort by hash
	sort.Sort(dls)
	if len(dls) <= 1 {
		return -1
	}

	prevHash := dls[0].Hash
	prevIndex := dls[0].Index
	for i := 1; i < len(dls); i++ {
		if dls[i].Hash == prevHash && dls[i].Index == prevIndex {

			return i
		}
		prevHash = dls[i].Hash
		prevIndex = dls[i].Index
	}

	return -1
}

// ----------------------------------------------------------------------
type OpCUInfo struct {
	Symbol                    string          `json:"symbol"`
	CuAddress                 string          `json:"cu_address"`
	Locked                    bool            `json:"locked"`
	Amount                    Int             `json:"amount"` // TODO 计算当前余额
	MultisignAddress          string          `json:"multisign_address"`
	LastEpochMultisignAddress string          `json:"last_epoch_multisign_address"`
	DepositList               DepositList     `json:"deposit_list"`
	MainNetAmount             Int             `json:"main_net_amount"` // erc20's amount of eth
	GasUsed                   Int             `json:"gas_used"`
	GasReceived               Int             `json:"gas_received"`
	MigrationStatus           MigrationStatus `json:"migration_status"`
}

// String implements fmt.Stringer
func (oc *OpCUInfo) String() string {
	return fmt.Sprintf(`Account:
  Address:   %s
  ExtAddress: %s
  Locked:    %v
  GasReceived: %v
  GasUsed:   %v`,
		oc.CuAddress, oc.MultisignAddress, oc.Locked, oc.GasReceived, oc.GasUsed)
}

type OpCUsInfo []OpCUInfo

func (cs OpCUsInfo) String() string {
	bsb := strings.Builder{}
	for _, c := range cs {
		bsb.WriteString(c.String())
	}
	return bsb.String()
}

type ChainCUInfo struct {
	Chain       string `json:"chain"`
	Addr        string `json:"addr"`
	CuAddress   string `json:"cu_address"`
	IsChainAddr bool   `json:"ischainaddr"`
	IsOPCU      bool   `json:"isopcu"`
}

func (info *ChainCUInfo) String() string {
	return fmt.Sprintf(`ChainInfo:
  Chain:     %s
  Address:   %s
  Cuaddress: %s
  IsChainAddr: %v
  IsOPCU:    %v`,
		info.Chain, info.Addr, info.CuAddress, info.IsChainAddr, info.IsOPCU)
}
