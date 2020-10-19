/*
 * *******************************************************************
 * @项目名称: chainnode
 * @文件名称: chainnode.go
 * @Date: 2019/03/22
 * @Author: kai.wen
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

package chainnode

import (
	"errors"
	sdk "github.com/hbtc-chain/bhchain/types"
)

// Transaction status
const (
	StatusNotFound              = 0
	StatusPending               = 1
	StatusFailed                = 2
	StatusSuccess               = 3
	StatusContractExecuteFailed = 4
	StatusOther                 = 5
)

// Errors
var (
	ErrorNotSupported     = errors.New("Chain is not supported")
	ErrorInvalidSignature = errors.New("Signature is invalid")
	ErrorInvalidInput     = errors.New("invalid input")
)

// Chainnode interface
type Chainnode interface {
	SupportChain(chain string) bool
	ConvertAddress(chain string, pubKey []byte) (string, error)
	ValidAddress(chain, symbol, address string) (bool, string)
	QueryBalance(chain, symbol, address, contractAddress string, blockHeight uint64) (sdk.Int, error) //will be obsoleted later, please use ValidateAsset
	QueryUtxo(chain, symbol string, vin *sdk.UtxoIn) (bool, error)
	QueryNonce(chain, address string) (uint64, error)
	QueryGasPrice(chain string) (sdk.Int, error)
	QueryUtxoTransaction(chain, symbol, hash string, asynMode bool) (*ExtUtxoTransaction, error)
	QueryAccountTransaction(chain, symbol, hash string, asynMode bool) (*ExtAccountTransaction, error)
	CreateUtxoTransaction(chain, symbol string, transaction *ExtUtxoTransaction) ([]byte, [][]byte, error)
	CreateAccountTransaction(chain, symbol, contractAddress string, transaction *ExtAccountTransaction) ([]byte, []byte, error)
	CreateUtxoSignedTransaction(chain, symbol string, raw []byte, signatures, pubKeys [][]byte) ([]byte, []byte, error)
	CreateAccountSignedTransaction(chain, symbol string, raw []byte, signatures, pubKeys []byte) ([]byte, []byte, error)
	VerifyUtxoSignedTransaction(chain, symbol string, address []string, signedTxData []byte, vins []*sdk.UtxoIn) (bool, error)
	VerifyAccountSignedTransaction(chain, symbol string, address string, signedTxData []byte) (bool, error)
	QueryAccountTransactionFromSignedData(chain, symbol string, signedTxData []byte) (*ExtAccountTransaction, error)
	QueryUtxoTransactionFromSignedData(chain, symbol string, signedTxData []byte, vins []*sdk.UtxoIn) (*ExtUtxoTransaction, error)
	QueryAccountTransactionFromData(chain, symbol string, rawData []byte) (*ExtAccountTransaction, []byte, error)
	QueryUtxoTransactionFromData(chain, symbol string, rawData []byte, vins []*sdk.UtxoIn) (*ExtUtxoTransaction, [][]byte, error)
	BroadcastTransaction(chain, symbol string, signedTxData []byte) (transactionHash string, err error)
	QueryUtxoInsFromData(chain, symbol string, data []byte) ([]*sdk.UtxoIn, error)
}

// ExtTransaction for external chain connected by the chainnode
type ExtUtxoTransaction struct {
	Hash        string
	Status      uint64
	Vins        []*sdk.UtxoIn
	Vouts       []*sdk.UtxoOut
	CostFee     sdk.Int
	BlockHeight uint64
	BlockTime   uint64
}

type ExtAccountTransaction struct {
	Hash            string
	Status          uint64
	From            string
	To              string
	Amount          sdk.Int
	Memo            string
	Nonce           uint64
	GasLimit        sdk.Int
	GasPrice        sdk.Int
	CostFee         sdk.Int
	BlockHeight     uint64
	BlockTime       uint64
	ContractAddress string
}
