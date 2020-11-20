package types

import (
	"encoding/hex"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
)

type OrderType int
type OrderStatus int

const (
	OrderTypeKeyGen            OrderType = 0x1
	OrderTypeWithdrawal        OrderType = 0x2
	OrderTypeCollect           OrderType = 0x3
	OrderTypeDeposit           OrderType = 0x4
	OrderTypeSysTransfer       OrderType = 0x5
	OrderTypeOpcuAssetTransfer OrderType = 0x6
)

const (
	OrderStatusBegin      OrderStatus = 0x1
	OrderStatusWaitSign   OrderStatus = 0x2
	OrderStatusSignFinish OrderStatus = 0x3
	OrderStatusFinish     OrderStatus = 0x4
	OrderStatusCancel     OrderStatus = 0x5 // order cancelled without broadcasting to chain
	OrderStatusFailed     OrderStatus = 0x6 // order failed after broadcasting to chain
)

// Match check if status equals
// or `s` == SignFinish and `other` is at terminate status
func (s OrderStatus) Match(other OrderStatus) bool {
	return s == other || (s == OrderStatusSignFinish && (other == OrderStatusFinish || other == OrderStatusFailed))
}

// Terminated check if status is at terminate status (OrderStatusFinish, OrderStatusCancel, OrderStatusFailed)
func (s OrderStatus) Terminated() bool {
	return s >= OrderStatusFinish
}

const (
	OrderIDLen = 36
)

const (
	DepositUnconfirm   = 0
	DepositWaitConfirm = 1
	DepositConfirmed   = 2
)

type WithdrawStatus int

const (
	WithdrawStatusUnconfirmed WithdrawStatus = 0
	// withdraw order confirmed valid by majority of settle
	WithdrawStatusValid = 1
	// withdraw order confirmed invalid (eg. withdraw to contract address) by majority of settle
	WithdrawStatusInvalid = 2
)

type TransferItem struct {
	Hash   string `json:"hash"`
	Index  uint64 `json:"index"`
	Amount Int    `json:"amount"`
}

type Order interface {
	GetOrderType() OrderType
	SetOrderType(orderType OrderType)
	GetOrderStatus() OrderStatus
	SetOrderStatus(status OrderStatus)
	GetID() string
	SetID(string) Error
	GetSymbol() string
	GetCUAddress() CUAddress
	GetHeight() uint64
	//DeepCopy() Order
	String() string
}

var _ Order = (*OrderBase)(nil)

type OrderBase struct {
	CUAddress CUAddress   `json:"cu_address"`
	ID        string      `json:"id"`
	OrderType OrderType   `json:"order_type"`
	Symbol    string      `json:"symbol"`
	Status    OrderStatus `json:"status"`
	Height    uint64      `json:"height"`
}

func (o *OrderBase) GetOrderType() OrderType {
	return o.OrderType
}

func (o *OrderBase) SetOrderType(orderType OrderType) {
	o.OrderType = orderType
}
func (o *OrderBase) GetOrderStatus() OrderStatus {
	return o.Status
}

func (o *OrderBase) SetOrderStatus(status OrderStatus) {
	o.Status = status
}

func (o *OrderBase) GetID() string {
	return o.ID
}

func (o *OrderBase) GetHeight() uint64 {
	return o.Height
}

func (o *OrderBase) SetID(id string) Error {
	if o.ID != "" {
		return ErrInternal("order id already exist")
	}
	o.ID = id
	return nil
}

func (o *OrderBase) GetSymbol() string {
	return o.Symbol
}

func (o *OrderBase) GetCUAddress() CUAddress {
	return o.CUAddress
}

// DeepCopy OrderBase
func (o *OrderBase) DeepCopy() Order {
	newOrder := &OrderBase{
		ID:        o.ID,
		OrderType: o.OrderType,
		Symbol:    o.Symbol,
		Status:    o.Status,
		CUAddress: make(CUAddress, len(o.CUAddress)),
	}
	copy(newOrder.CUAddress, o.CUAddress)
	return newOrder
}

func (o *OrderBase) String() string {
	return fmt.Sprintf(`
		CUAddress:%v
		ID:%s
		OrderType:%v
		Symbol:%v
		Status:%v`, o.CUAddress.String(), o.ID, o.OrderType, o.Symbol, o.Status)
}

//______________________________
var _ Order = (*OrderKeyGen)(nil)

type OrderKeyGen struct {
	OrderBase
	KeyNodes         []CUAddress `json:"key_nodes"`
	SignThreshold    uint64      `json:"sign_threshold"`
	To               CUAddress   `json:"to"`
	OpenFee          Coin        `json:"open_fee"`
	MultiSignAddress string      `json:"multi_sign_address"`
	Pubkey           []byte      `json:"pubkey"`
	Epoch            uint64      `json:"epoch"`
}

// DeepCopy OrderKeygen
func (o *OrderKeyGen) DeepCopy() Order {
	ob := o.OrderBase.DeepCopy().(*OrderBase)
	newOrder := &OrderKeyGen{
		OrderBase:        *ob,
		KeyNodes:         make([]CUAddress, len(o.KeyNodes)),
		SignThreshold:    o.SignThreshold,
		To:               o.To,
		OpenFee:          o.OpenFee,
		MultiSignAddress: o.MultiSignAddress,
	}
	copy(newOrder.KeyNodes, o.KeyNodes)
	return newOrder
}

func (o *OrderKeyGen) String() string {
	var build strings.Builder

	//TODO(Keep), keygen的code不全
	build.WriteString(o.OrderBase.String())
	build.WriteString(fmt.Sprintf(`
		To:%v
		MultiSignAddress:%v
		KeyNodes:%v
		SignThreshold:%v
		OpenFee:%v`, o.To.String(), o.MultiSignAddress, o.KeyNodes, o.SignThreshold, o.OpenFee))

	return build.String()
}

//____________________________________
var _ Order = (*OrderCollect)(nil)

type OrderCollect struct {
	OrderBase
	CollectFromCU      CUAddress `json:"collect_from_cu"`
	CollectFromAddress string    `json:"collect_from_address"`
	CollectToCU        CUAddress `json:"collect_to_cu"`
	Amount             Int       `json:"amount"`
	GasPrice           Int       `json:"gas_price"`
	GasLimit           Int       `json:"gas_limit"`
	CostFee            Int       `json:"cost_fee"`
	// external chain deposit tx hash
	Txhash string `json:"tx_hash"`
	// external chain deposit tx index
	Index    uint64 `json:"index"`
	Memo     string `json:"memo"`
	RawData  []byte `json:"raw_data"`
	SignedTx []byte `json:"signed_Tx"`
	// external chain collect tx hash
	ExtTxHash     string `json:"ext_txhash"`
	DepositStatus uint16 `json:"deposit_status"`
}

func (o *OrderCollect) GetRawdata() []byte {
	return o.RawData
}

func (o *OrderCollect) GetSignedTx() []byte {
	return o.SignedTx
}

// DeepCopy OrderCollect
func (o *OrderCollect) DeepCopy() Order {
	ob := o.OrderBase.DeepCopy().(*OrderBase)
	newOrder := &OrderCollect{
		OrderBase:          *ob,
		CollectFromCU:      o.CollectFromCU,
		CollectFromAddress: o.CollectFromAddress,
		CollectToCU:        o.CollectToCU,
		Amount:             o.Amount,
		GasPrice:           o.GasPrice,
		GasLimit:           o.GasLimit,
		Txhash:             o.Txhash,
		Index:              o.Index,
		RawData:            make([]byte, len(o.RawData)),
		SignedTx:           make([]byte, len(o.SignedTx)),
		DepositStatus:      o.DepositStatus,
		ExtTxHash:          o.ExtTxHash,
	}
	copy(newOrder.RawData, o.RawData)
	copy(newOrder.SignedTx, o.SignedTx)
	return newOrder
}

func (o *OrderCollect) String() string {
	var build strings.Builder
	build.WriteString(o.OrderBase.String())
	build.WriteString(fmt.Sprintf(`
        CollectFromCU:%v
        CollectFromAddress:%v
        CollectToCU:%v
		Amount:%v
		GasPrice:%v
        GasLimit:%v
        TxHash:%v
	    Index:%v
        RawData:%x
		SignedTx:%x
        DepositStatus:%x
        ExtTxHash:%v`,
		o.CollectFromCU, o.CollectFromAddress, o.CollectToCU, o.Amount,
		o.GasPrice, o.GasLimit, o.Txhash, o.Index, o.RawData, o.SignedTx,
		o.DepositStatus, o.ExtTxHash))
	return build.String()
}

// MarshalYAML returns the YAML representation of an SysTransfer order.
func (o *OrderCollect) MarshalYAML() (interface{}, error) {
	var orderStr []byte
	var err error
	var rawData string
	var signedTx string
	if o.RawData != nil {
		rawData = hex.EncodeToString(o.RawData)
	}
	if o.SignedTx != nil {
		signedTx = hex.EncodeToString(o.SignedTx)
	}
	orderStr, err = yaml.Marshal(struct {
		CUAddress          CUAddress
		ID                 string
		OrderType          OrderType
		Symbol             string
		Status             OrderStatus
		CollectFromCU      CUAddress
		CollectFromAddress string
		CollectToCU        CUAddress
		Amount             Int
		GasPrice           Int
		GasLimit           Int
		Txhash             string
		Index              uint64
		Memo               string
		RawData            string
		SignedTx           string
		ExtTxHash          string
		DepositStatus      uint16
	}{

		CUAddress:          o.CUAddress,
		ID:                 o.ID,
		OrderType:          o.OrderType,
		Symbol:             o.Symbol,
		Status:             o.Status,
		CollectFromCU:      o.CollectFromCU,
		CollectFromAddress: o.CollectFromAddress,
		CollectToCU:        o.CollectToCU,
		Amount:             o.Amount,
		GasPrice:           o.GasPrice,
		GasLimit:           o.GasLimit,
		Txhash:             o.Txhash,
		Index:              o.Index,
		Memo:               o.Memo,
		RawData:            rawData,
		SignedTx:           signedTx,
		DepositStatus:      o.DepositStatus,
		ExtTxHash:          o.ExtTxHash,
	})
	if err != nil {
		return nil, err
	}

	return string(orderStr), err
}

//_____________________________
var _ Order = (*OrderWithdrawal)(nil)

type OrderWithdrawal struct {
	OrderBase
	Amount            Int            `json:"amount"`
	GasFee            Int            `json:"gas_fee"`
	CostFee           Int            `json:"cost_fee"`
	WithdrawToAddress string         `json:"withdraw_to_address"`
	FromAddress       string         `json:"from_address"`
	OpCUaddress       string         `json:"opcu_address"`
	UtxoInNum         int            `json:"utxoin_num"`
	Txhash            string         `json:"tx_hash"`
	RawData           []byte         `json:"raw_data"`
	SignedTx          []byte         `json:"signed_tx"`
	WithdrawStatus    WithdrawStatus `json:"withdraw_status"`
}

func (o *OrderWithdrawal) GetRawdata() []byte {
	return o.RawData
}
func (o *OrderWithdrawal) GetSignedTx() []byte {
	return o.SignedTx
}

// DeepCopy OrderWithdrawal
func (o *OrderWithdrawal) DeepCopy() Order {
	ob := o.OrderBase.DeepCopy().(*OrderBase)
	newOrder := &OrderWithdrawal{
		OrderBase:         *ob,
		Amount:            o.Amount,
		GasFee:            o.GasFee,
		CostFee:           o.CostFee,
		WithdrawToAddress: o.WithdrawToAddress,
		FromAddress:       o.FromAddress,
		OpCUaddress:       o.OpCUaddress,
		Txhash:            o.Txhash,
		UtxoInNum:         o.UtxoInNum,
		RawData:           make([]byte, len(o.RawData)),
		SignedTx:          make([]byte, len(o.SignedTx)),
	}
	copy(newOrder.RawData, o.RawData)
	copy(newOrder.SignedTx, o.SignedTx)
	return newOrder
}

func (o *OrderWithdrawal) String() string {
	var build strings.Builder
	build.WriteString(o.OrderBase.String())
	build.WriteString(fmt.Sprintf(`
		Amount:%v
		GasFee:%v
        CostFee:%v
        WithdrawalToAddress:%v
		FromAddress:%v
        TxHash:%v
        UtxoInNum:%v
        RawData:%x
		SignedTx:%x`,
		o.Amount, o.GasFee, o.CostFee, o.WithdrawToAddress, o.FromAddress, o.Txhash, o.UtxoInNum,
		hex.EncodeToString(o.RawData), hex.EncodeToString(o.SignedTx)))
	return build.String()
}

// MarshalYAML returns the YAML representation of an withdrawal order.
func (o *OrderWithdrawal) MarshalYAML() (interface{}, error) {
	var orderStr []byte
	var err error
	var rawData string
	var signedTx string
	if o.RawData != nil {
		rawData = hex.EncodeToString(o.RawData)
	}
	if o.SignedTx != nil {
		signedTx = hex.EncodeToString(o.SignedTx)
	}
	orderStr, err = yaml.Marshal(struct {
		CUAddress         CUAddress
		ID                string
		OrderType         OrderType
		Symbol            string
		Status            OrderStatus
		Amount            Int
		GasFee            Int
		CostFee           Int
		WithdrawToAddress string
		FromAddress       string
		OpCUaddress       string
		Txhash            string
		UtxoInNum         int
		RawData           string
		SignedTx          string
	}{

		CUAddress:         o.CUAddress,
		ID:                o.ID,
		OrderType:         o.OrderType,
		Symbol:            o.Symbol,
		Status:            o.Status,
		Amount:            o.Amount,
		GasFee:            o.GasFee,
		CostFee:           o.CostFee,
		WithdrawToAddress: o.WithdrawToAddress,
		FromAddress:       o.FromAddress,
		OpCUaddress:       o.OpCUaddress,
		Txhash:            o.Txhash,
		UtxoInNum:         o.UtxoInNum,
		RawData:           rawData,
		SignedTx:          signedTx,
	})
	if err != nil {
		return nil, err
	}

	return string(orderStr), err
}

//_____________________________
var _ Order = (*OrderSysTransfer)(nil)

type OrderSysTransfer struct {
	OrderBase
	Amount      Int    `json:"amount"`
	CostFee     Int    `json:"cost_fee"`
	ToCU        string `json:"to_cu"`
	ToAddress   string `json:"to_address"`
	FromAddress string `json:"from_address"`
	OpCUaddress string `json:"opcu_address"`
	TxHash      string `json:"tx_hash"`
	RawData     []byte `json:"raw_data"`
	SignedTx    []byte `json:"signed_tx"`
}

func (o *OrderSysTransfer) GetRawdata() []byte {
	return o.RawData
}
func (o *OrderSysTransfer) GetSignedTx() []byte {
	return o.SignedTx
}

// DeepCopy OrderWithdrawal
func (o *OrderSysTransfer) DeepCopy() Order {
	ob := o.OrderBase.DeepCopy().(*OrderBase)
	newOrder := &OrderSysTransfer{
		OrderBase:   *ob,
		Amount:      o.Amount,
		CostFee:     o.CostFee,
		ToCU:        o.ToCU,
		ToAddress:   o.ToAddress,
		FromAddress: o.FromAddress,
		OpCUaddress: o.OpCUaddress,
		TxHash:      o.TxHash,
		RawData:     make([]byte, len(o.RawData)),
		SignedTx:    make([]byte, len(o.SignedTx)),
	}
	copy(newOrder.RawData, o.RawData)
	copy(newOrder.SignedTx, o.SignedTx)
	return newOrder
}

func (o *OrderSysTransfer) String() string {
	var build strings.Builder
	build.WriteString(o.OrderBase.String())
	build.WriteString(fmt.Sprintf(`
		Amount:%v
        CostFee:%v
        ToCU:%v
        ToAddress:%v
        FromAddress:%v
        OpCUAddress:%v
        TxHash:%v
        RawData:%x
		SignedTx:%x`,
		o.Amount, o.CostFee, o.ToCU, o.ToAddress, o.FromAddress, o.OpCUaddress, o.TxHash,
		hex.EncodeToString(o.RawData), hex.EncodeToString(o.SignedTx)))
	return build.String()
}

// MarshalYAML returns the YAML representation of an SysTransfer order.
func (o *OrderSysTransfer) MarshalYAML() (interface{}, error) {
	var orderStr []byte
	var err error
	var rawData string
	var signedTx string
	if o.RawData != nil {
		rawData = hex.EncodeToString(o.RawData)
	}
	if o.SignedTx != nil {
		signedTx = hex.EncodeToString(o.SignedTx)
	}
	orderStr, err = yaml.Marshal(struct {
		CUAddress   CUAddress
		ID          string
		OrderType   OrderType
		Symbol      string
		Status      OrderStatus
		Amount      Int
		CostFee     Int
		ToCU        string
		ToAddress   string
		FromAddress string
		OpCUaddress string
		TxHash      string
		RawData     string
		SignedTx    string
	}{
		CUAddress:   o.CUAddress,
		ID:          o.ID,
		OrderType:   o.OrderType,
		Symbol:      o.Symbol,
		Status:      o.Status,
		Amount:      o.Amount,
		CostFee:     o.CostFee,
		ToCU:        o.ToCU,
		ToAddress:   o.ToAddress,
		FromAddress: o.FromAddress,
		OpCUaddress: o.OpCUaddress,
		TxHash:      o.TxHash,
		RawData:     rawData,
		SignedTx:    signedTx,
	})
	if err != nil {
		return nil, err
	}

	return string(orderStr), err
}

var _ Order = (*OrderOpcuAssetTransfer)(nil)

type OrderOpcuAssetTransfer struct {
	OrderBase
	TransfertItems []TransferItem `json:"transfer_items"`
	ToAddr         string         `json:"to_address"`
	RawData        []byte         `json:"raw_data"`
	SignedTx       []byte         `json:"signed_tx"`
	Txhash         string         `json:"tx_hash"`
	CostFee        Int            `json:"cost_fee"`
}

func (o *OrderOpcuAssetTransfer) GetRawdata() []byte {
	return o.RawData
}

func (o *OrderOpcuAssetTransfer) GetSignedTx() []byte {
	return o.SignedTx
}

func (o *OrderOpcuAssetTransfer) GetTxHash() string {
	return o.Txhash
}

// DeepCopy OrderWithdrawal
func (o *OrderOpcuAssetTransfer) DeepCopy() Order {
	ob := o.OrderBase.DeepCopy().(*OrderBase)
	newOrder := &OrderOpcuAssetTransfer{
		OrderBase:      *ob,
		TransfertItems: make([]TransferItem, len(o.TransfertItems)),
		CostFee:        o.CostFee,
		ToAddr:         o.ToAddr,
		RawData:        make([]byte, len(o.RawData)),
		SignedTx:       make([]byte, len(o.SignedTx)),
		Txhash:         o.Txhash,
	}
	copy(newOrder.RawData, o.RawData)
	copy(newOrder.SignedTx, o.SignedTx)
	copy(newOrder.TransfertItems, o.TransfertItems)
	return newOrder
}

func (o *OrderOpcuAssetTransfer) String() string {
	var build strings.Builder
	build.WriteString(o.OrderBase.String())
	build.WriteString(fmt.Sprintf(`
		TransferItems:%x
        CostFee:%v
        ToAddress:%v
        RawData:%x
		SignedTx:%x
        Txhash:%x`,
		o.TransfertItems, o.CostFee, o.ToAddr,
		hex.EncodeToString(o.RawData), hex.EncodeToString(o.SignedTx), o.Txhash))
	return build.String()
}

// MarshalYAML returns the YAML representation of an withdrawal order.
func (o *OrderOpcuAssetTransfer) MarshalYAML() (interface{}, error) {
	var orderStr []byte
	var err error
	var rawData string
	var signedTx string
	if o.RawData != nil {
		rawData = hex.EncodeToString(o.RawData)
	}
	if o.SignedTx != nil {
		signedTx = hex.EncodeToString(o.SignedTx)
	}
	orderStr, err = yaml.Marshal(struct {
		CUAddress     CUAddress
		ID            string
		OrderType     OrderType
		Symbol        string
		Status        OrderStatus
		TransferItems []TransferItem
		CostFee       Int
		ToAddr        string
		RawData       string
		SignedTx      string
		Txhash        string
	}{

		CUAddress:     o.CUAddress,
		ID:            o.ID,
		OrderType:     o.OrderType,
		Symbol:        o.Symbol,
		Status:        o.Status,
		TransferItems: o.TransfertItems,
		CostFee:       o.CostFee,
		ToAddr:        o.ToAddr,
		RawData:       rawData,
		SignedTx:      signedTx,
		Txhash:        o.Txhash,
	})
	if err != nil {
		return nil, err
	}

	return string(orderStr), err
}

func IsIllegalOrderID(orderID string) bool {
	_, err := uuid.FromString(orderID)
	return err != nil
}

// IsIllegalOrderIDList checks whether a list of order ID is valid.
// A valid order id list cannot contain duplicated order IDs and every order id must be legal .
func IsIllegalOrderIDList(orderIDs []string) bool {
	existent := make(map[string]bool)
	for _, id := range orderIDs {
		if IsIllegalOrderID(id) || existent[id] {
			return true
		}
		existent[id] = true
	}
	return false
}
