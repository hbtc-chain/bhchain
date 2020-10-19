package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tendermint/go-amino"

	"github.com/hbtc-chain/bhchain/codec"

	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// CategoryType indicates the category type that causes the receipt in the flow.
type CategoryType uint64

type DepositType uint64

const (
	// TODO(kai.wen): Add more category type
	CategoryTypeTransfer          CategoryType = 0x1
	CategoryTypeMultiTransfer     CategoryType = 0x2
	CategoryTypeKeyGen            CategoryType = 0x3
	CategoryTypeDeposit           CategoryType = 0x4
	CategoryTypeWithdrawal        CategoryType = 0x5
	CategoryTypeCollect           CategoryType = 0x6
	CategoryTypeSysTransfer       CategoryType = 0x7
	CategoryTypeOrderRetry        CategoryType = 0x8
	CategoryTypeMapping           CategoryType = 0x9
	CategoryTypeOpcuAssetTransfer CategoryType = 0xA
	CategoryTypeOpenswap          CategoryType = 0xB
	CategoryTypeQuickSwap         CategoryType = 0xC
	CategoryTypeHrc20             CategoryType = 0xD
)

const (
	DepositTypeCU   DepositType = 0x1
	DepositTypeOPCU DepositType = 0x2
)

// Receipt defines basic interface for all kind of receipts
type Receipt struct {
	// Category for the transaction that causes the receipt.
	Category CategoryType

	// Flows list of flows.
	Flows []Flow
}

// Flow defines the interface of the flow in the receipt
type Flow interface{}

// OrderFlow for order change
type OrderFlow struct {
	Symbol Symbol

	// CUAddress the address for the custodian unit for the order change
	CUAddress CUAddress

	// OrderID for the order change
	OrderID string

	// OrderType for the order
	OrderType OrderType

	OrderStatus OrderStatus
}

// BalanceFlow for asset balance change
type BalanceFlow struct {
	// CUAddress the address for the custodian unit for the balance change
	CUAddress CUAddress

	// Symbol token symbol for the asset
	//Symbol Symbol
	Symbol Symbol //FIXME(liyong.zhang): temp change

	// PreviousBalance previous balance for the balance change
	PreviousBalance Int

	// BalanceChange the actual balance change
	BalanceChange Int

	// PreviousBalanceOnHold previous balance on hold
	PreviousBalanceOnHold Int

	// BalanceOnHoldChange the actual balance on hold change
	BalanceOnHoldChange Int
}

type DepositFlow struct {
	CuAddress         string
	Multisignedadress string
	Symbol            string
	Index             uint64
	Txhash            string
	Amount            Int
	OrderID           string
	DepositType       DepositType
	Memo              string
	Epoch             uint64
}

type OrderRetryFlow struct {
	OrderIDs []string
}

type DepositConfirmedFlow struct {
	ValidOrderIDs   []string
	InValidOrderIDs []string
}

type CollectWaitSignFlow struct {
	OrderIDs []string
	RawData  []byte
}

type CollectSignFinishFlow struct {
	OrderIDs []string
	SignedTx []byte
	TxHash   string
}

type CollectFinishFlow struct {
	OrderIDs []string
	CostFee  Int
}

type WithdrawalFlow struct {
	OrderID        string
	FromCu         string
	ToAddr         string
	Symbol         string
	Amount         Int
	GasFee         Int
	WithdrawStatus WithdrawStatus
}

type WithdrawalConfirmFlow struct {
	OrderID        string
	WithdrawStatus WithdrawStatus
}

type WithdrawalWaitSignFlow struct {
	OrderIDs []string
	OpCU     string
	FromAddr string
	RawData  []byte
}

type WithdrawalSignFinishFlow struct {
	OrderIDs []string
	SignedTx []byte
	TxHash   string
}

type WithdrawalFinishFlow struct {
	OrderIDs []string
	CostFee  Int
	Valid    bool
}

type OpcuAssetTransferFlow struct {
	OrderID       string
	Opcu          string
	FromAddr      string
	ToAddr        string
	Symbol        string
	TransferItems []TransferItem
}

type OpcuAssetTransferWaitSignFlow struct {
	OrderID string
	RawData []byte
}

type OpcuAssetTransferSignFinishFlow struct {
	OrderID  string
	SignedTx []byte
	TxHash   string
}

type OpcuAssetTransferFinishFlow struct {
	OrderID string
	CostFee Int
}

type SysTransferFlow struct {
	OrderID  string
	FromCU   string
	ToCU     string
	FromAddr string
	ToAddr   string
	Symbol   string
	Amount   Int
}

type SysTransferWaitSignFlow struct {
	OrderID string
	RawData []byte
}

type SysTransferSignFinishFlow struct {
	OrderID  string
	SignedTx []byte
	TxHash   string
}

type SysTransferFinishFlow struct {
	OrderID string
	CostFee Int
}

type KeyGenFlow struct {
	OrderID     string
	Symbol      Symbol
	From        CUAddress
	To          CUAddress
	IsPreKeyGen bool
}

type KeyGenWaitSignFlow struct {
	OrderID string
	PubKey  []byte
}

type KeyGenFinishFlow struct {
	OrderID     string
	ToAddr      string
	IsPreKeyGen bool
}

// MappingBalanceFlow for mapping balance change
type MappingBalanceFlow struct {
	// Symbol token symbol for the mapping asset
	IssueSymbol       Symbol
	PreviousIssuePool Int
	IssuePoolChange   Int
}

// GetReceiptFromData decode receipts from tx result. A tx may have more than one receipts.
func GetReceiptFromData(cdc *codec.Codec, data []byte) ([]Receipt, error, bool) {
	if data == nil {
		return nil, errors.New("invalid data"), false
	}
	rcs := make([]Receipt, 0)
	buffer := bytes.NewBuffer(data)
	l := int64(buffer.Len())
	consumed := int64(0)
	skipped := false
	for consumed < l {
		var rc Receipt
		n, err := unmarshalBinaryLengthPrefixedReader(cdc, buffer, &rc, 0)
		if err != nil {
			if _, ok := err.(unmarshalBinaryBareErr); ok {
				// skip other type
				consumed += n
				skipped = true
			} else {
				return nil, err, false
			}
		} else {
			consumed += n
			rcs = append(rcs, rc)
		}
	}

	return rcs, nil, skipped
}

// copied from Codec.UnmarshalBinaryLengthPrefixedReader, return custom err type for UnmarshalBinaryBare err
func unmarshalBinaryLengthPrefixedReader(cdc *amino.Codec, r io.Reader, ptr interface{}, maxSize int64) (n int64, err error) {
	if maxSize < 0 {
		panic("maxSize cannot be negative.")
	}

	// Read byte-length prefix.
	var l int64
	var buf [binary.MaxVarintLen64]byte
	for i := 0; i < len(buf); i++ {
		_, err = r.Read(buf[i : i+1])
		if err != nil {
			return
		}
		n++
		if buf[i]&0x80 == 0 {
			break
		}
		if n >= maxSize {
			err = fmt.Errorf("read overflow, maxSize is %v but uvarint(length-prefix) is itself greater than maxSize", maxSize)
		}
	}
	u64, _ := binary.Uvarint(buf[:])
	if err != nil {
		return
	}
	if maxSize > 0 {
		if uint64(maxSize) < u64 {
			err = fmt.Errorf("read overflow, maxSize is %v but this amino binary object is %v bytes", maxSize, u64)
			return
		}
		if (maxSize - n) < int64(u64) {
			err = fmt.Errorf("read overflow, maxSize is %v but this length-prefixed amino binary object is %v+%v bytes", maxSize, n, u64)
			return
		}
	}
	l = int64(u64)
	if l < 0 {
		err = fmt.Errorf("read overflow, this implementation can't read this because, why would anyone have this much data? Hello from 2018")
	}

	// Read that many bytes.
	var bz = make([]byte, l, l)
	_, err = io.ReadFull(r, bz)
	if err != nil {
		return
	}
	n += l

	// Decode.
	err = cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		err = unmarshalBinaryBareErr(err.Error())
	}
	return
}

type unmarshalBinaryBareErr string

func (u unmarshalBinaryBareErr) Error() string {
	return string(u)
}

func NewResultFromResultTx(res *ctypes.ResultTx) Result {
	return NewResultFromDeliverTx(&res.TxResult)
}

func NewResultFromDeliverTx(res *abci.ResponseDeliverTx) Result {
	return Result{
		Code:      CodeType(res.Code),
		Codespace: CodespaceType(res.Codespace),
		Data:      res.Data,
		Log:       res.Log,
		GasWanted: uint64(res.GasWanted),
		GasUsed:   uint64(res.GasUsed),
		//Tags:      res.Tags, FIXME(liyong.zhang): temp change
	}
}
