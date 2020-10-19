package chainnode

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/mock"
)

type MockChainnode struct {
	mock.Mock
}

var _ Chainnode = (*MockChainnode)(nil)

func (m *MockChainnode) SupportChain(chain string) bool {
	args := m.Called(chain)
	return args.Bool(0)
}

func (m *MockChainnode) ValidAddress(chain, symbol, address string) (bool, string) {
	args := m.Called(chain, symbol, address)
	return args.Bool(0), args.String(1)
}

func (m *MockChainnode) ConvertAddress(chain string, publicKey []byte) (string, error) {
	return m.ConvertAddressFromSerializedPubKey(chain, publicKey)
}

func (m *MockChainnode) ConvertAddressFromSerializedPubKey(chain string, publicKey []byte) (string, error) {
	args := m.Called(chain, publicKey)
	return args.String(0), args.Error(1)
}

func (m *MockChainnode) QueryBalance(chain, symbol string, address, contractAddress string, blockHeight uint64) (sdk.Int, error) {
	args := m.Called(chain, symbol, address, contractAddress, blockHeight)
	return args.Get(0).(sdk.Int), args.Error(1)
}

func (m *MockChainnode) QueryUtxo(chain, symbol string, vin *sdk.UtxoIn) (bool, error) {
	args := m.Called(chain, symbol, vin)
	return args.Bool(0), args.Error(1)

}

func (m *MockChainnode) QueryNonce(chain, address string) (uint64, error) {
	args := m.Called(chain, address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockChainnode) QueryGasPrice(chain string) (sdk.Int, error) {
	args := m.Called(chain)
	return args.Get(0).(sdk.Int), args.Error(1)
}

func (m *MockChainnode) QueryUtxoTransaction(chain, symbol, hash string, asynMode bool) (*ExtUtxoTransaction, error) {
	args := m.Called(chain, symbol, hash, asynMode)
	return args.Get(0).(*ExtUtxoTransaction), args.Error(1)
}

func (m *MockChainnode) QueryAccountTransaction(chain, symbol, hash string, asynMode bool) (*ExtAccountTransaction, error) {
	args := m.Called(chain, symbol, hash, asynMode)
	return args.Get(0).(*ExtAccountTransaction), args.Error(1)
}

func (m *MockChainnode) CreateUtxoTransaction(chain, symbol string, transaction *ExtUtxoTransaction) ([]byte, [][]byte, error) {
	args := m.Called(chain, symbol, transaction)
	return args.Get(0).([]byte), args.Get(1).([][]byte), args.Error(2)
}

func (m *MockChainnode) CreateAccountTransaction(chain, symbol, contractAddress string, transaction *ExtAccountTransaction) (
	[]byte, []byte, error) {
	args := m.Called(chain, symbol, contractAddress, transaction)
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}
func (m *MockChainnode) CreateUtxoSignedTransaction(chain, symbol string, raw []byte, signatures, pubKeys [][]byte) ([]byte, []byte, error) {
	panic("not implemented")
}

func (m *MockChainnode) CreateAccountSignedTransaction(chain, symbol string, raw []byte, signatures, pubKeys []byte) ([]byte, []byte, error) {
	args := m.Called(chain, symbol, raw, signatures, pubKeys)
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockChainnode) VerifyUtxoSignedTransaction(chain, symbol string, address []string, signedTxData []byte, vins []*sdk.UtxoIn) (bool, error) {
	args := m.Called(chain, symbol, address, signedTxData, vins)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockChainnode) VerifyAccountSignedTransaction(chain, symbol, address string, signedTxData []byte) (bool, error) {
	args := m.Called(chain, symbol, address, signedTxData)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockChainnode) QueryAccountTransactionFromSignedData(chain, symbol string, signedTxData []byte) (*ExtAccountTransaction, error) {
	args := m.Called(chain, symbol, signedTxData)
	return args.Get(0).(*ExtAccountTransaction), args.Error(1)
}

func (m *MockChainnode) QueryUtxoTransactionFromSignedData(chain, symbol string, signedTxData []byte, vins []*sdk.UtxoIn) (*ExtUtxoTransaction, error) {
	args := m.Called(chain, symbol, signedTxData, vins)
	return args.Get(0).(*ExtUtxoTransaction), args.Error(1)
}

func (m *MockChainnode) QueryAccountTransactionFromData(chain, symbol string, rawData []byte) (*ExtAccountTransaction, []byte, error) {
	args := m.Called(chain, symbol, rawData)
	return args.Get(0).(*ExtAccountTransaction), args.Get(1).([]byte), args.Error(2)
}

func (m *MockChainnode) QueryUtxoTransactionFromData(chain, symbol string, rawData []byte, vins []*sdk.UtxoIn) (*ExtUtxoTransaction, [][]byte, error) {
	args := m.Called(chain, symbol, rawData, vins)
	return args.Get(0).(*ExtUtxoTransaction), args.Get(1).([][]byte), args.Error(2)
}

func (m *MockChainnode) BroadcastTransaction(chain, symbol string, signedTxData []byte) (string, error) {
	args := m.Called(symbol, signedTxData)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockChainnode) QueryUtxoInsFromData(chain, symbol string, data []byte) ([]*sdk.UtxoIn, error) {
	args := m.Called(chain, symbol, data)
	return args.Get(0).([]*sdk.UtxoIn), args.Error(1)
}
