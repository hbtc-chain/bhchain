package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/chainnode/chaindispatcher"
	"github.com/hbtc-chain/chainnode/config"
	"github.com/hbtc-chain/chainnode/proto"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	errInvalidBHSymbol       = errors.New("Invalid symbol for BH")
	errInvalidNumberString   = errors.New("Invalid number string")
	errTxSignatureNotMatched = errors.New("Transaction signature not matched")
)

var _ chainnode.Chainnode = (*Chainnode)(nil)

// Chainnode implements chainnode.Chainnode
type Chainnode struct {
	client *chaindispatcher.ChainDispatcher
	logger log.Logger
}

type grpcCommonReply interface {
	GetCode() proto.ReturnCode
	GetMsg() string
}

func validateReply(reply grpcCommonReply) error {
	if reply.GetCode() != proto.ReturnCode_SUCCESS {
		return fmt.Errorf("Reply fails for %v, msg: %v",
			reflect.TypeOf(reply), reply.GetMsg())
	}
	return nil
}

// New creates a new instance
func New(logger log.Logger) *Chainnode {
	return &Chainnode{
		logger: logger,
	}
}

// Connect to Chainnode server via grpc
func (ch *Chainnode) Init(network string) error {
	var networkType config.NetWorkType
	if network == "mainnet" {
		networkType = config.MainNet
	} else if network == "testnet" {
		networkType = config.TestNet
	} else if network == "regtest" {
		networkType = config.RegTest
	} else {
		panic("unsupported chainnode network type: " + network)
	}
	ch.logger.Info("Init Chainnode", "networktype", networkType)
	ch.client = chaindispatcher.NewLocal(networkType)
	return nil
}

// SupportAsset checks if a symbol is supported
func (ch *Chainnode) SupportChain(chain string) bool {
	reply, err := ch.client.SupportChain(context.TODO(), &proto.SupportChainRequest{
		Chain: chain,
	})
	if err != nil {
		ch.logger.Error("SupportChain fails", "err", err, "chain", chain)
		return false
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("SupportAsset fails", "err", err, "chain", chain)
		return false
	}
	return reply.Support
}

// IsValidAddress checks the validity of an address and to be withdrawed to
func (ch *Chainnode) ValidAddress(chain, symbol, address string) (bool, string) {
	reply, err := ch.client.ValidAddress(context.TODO(), &proto.ValidAddressRequest{
		Chain:   chain,
		Symbol:  symbol,
		Address: address,
	})
	if err != nil {
		ch.logger.Error("IsValidAddress fails", "chain", chain, "symbol", symbol, "err", err)
		return false, ""
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("IsValidAddress fails", "chain", chain, "symbol", symbol, "err", err)
		return false, ""
	}
	return reply.Valid, reply.CanonicalAddress
}

// ConvertAddress converts publicKey to address
func (ch *Chainnode) ConvertAddress(chain string, publicKey []byte) (string, error) {
	reply, err := ch.client.ConvertAddress(context.TODO(), &proto.ConvertAddressRequest{
		Chain:     chain,
		PublicKey: publicKey,
	})
	if err != nil {
		ch.logger.Error("ConvertAddress fails", "err", err, "chain", chain)
		return "", err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("ConvertAddress fails", "err", err, "chain", chain)
		return "", err
	}
	return reply.Address, nil
}

func (ch *Chainnode) QueryBalance(chain, symbol, address, contractAddress string, blockHeight uint64) (
	sdk.Int, error) {
	reply, err := ch.client.QueryBalance(context.TODO(), &proto.QueryBalanceRequest{
		Chain:           chain,
		Symbol:          symbol,
		Address:         address,
		ContractAddress: contractAddress,
		BlockHeight:     blockHeight,
	})

	if err != nil {
		ch.logger.Error("Failed to query balance ", "err", err, "chain", chain, "symbol", symbol, "Address", address, "ContactAddress", contractAddress, "blockHeight", blockHeight)
		return sdk.ZeroInt(), err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("Failed to query balance ", "err", err, "chain", chain, "symbol", symbol, "Address", address, "ContactAddress", contractAddress, "blockHeight", blockHeight)
		return sdk.ZeroInt(), err
	}
	amt, ok := sdk.NewIntFromString(reply.Balance)
	if !ok {
		return sdk.ZeroInt(), errors.New("Fail to parse the int")
	}
	return amt, nil
}

func (ch *Chainnode) QueryUtxo(chain, symbol string, vin *sdk.UtxoIn) (
	bool, error) {
	reply, err := ch.client.QueryUtxo(context.TODO(), &proto.QueryUtxoRequest{
		Chain:  chain,
		Symbol: symbol,
		Vin: &proto.Vin{
			Hash:    vin.Hash,
			Index:   uint32(vin.Index),
			Amount:  vin.Amount.Int64(),
			Address: vin.Address,
		},
	})

	if err != nil {
		ch.logger.Error("Failed to query utxo ", "err", err, "chain", chain, "symbol", symbol, "utxo", vin)
		return false, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("Failed to query utxo ", "err", err, "chain", chain, "symbol", symbol, "utxo", vin)
		return false, err
	}

	return reply.Unspent, nil
}

func (ch *Chainnode) QueryNonce(chain, address string) (uint64, error) {
	reply, err := ch.client.QueryNonce(context.TODO(), &proto.QueryNonceRequest{
		Chain:   chain,
		Address: address,
	})
	if err != nil {
		ch.logger.Error("Failed to QueryNonce ", "err", err, "chain", chain, "Address", address)
		return 0, err
	}

	if err := validateReply(reply); err != nil {
		ch.logger.Error("Failed to QueryNonce ", "err", err, "chain", chain, "Address", address)
		return 0, err
	}

	return reply.Nonce, nil

}
func (ch *Chainnode) QueryGasPrice(chain string) (sdk.Int, error) {
	reply, err := ch.client.QueryGasPrice(context.TODO(), &proto.QueryGasPriceRequest{
		Chain: chain,
	})
	if err != nil {
		ch.logger.Error("Failed to QueryNonce ", "err", err, "chain", chain)
		return sdk.ZeroInt(), err
	}

	if err := validateReply(reply); err != nil {
		ch.logger.Error("Failed to QueryNonce ", "err", err, "chain", chain)
		return sdk.ZeroInt(), err
	}

	price, ok := sdk.NewIntFromString(reply.GasPrice)
	if !ok {
		ch.logger.Error("Failed to QueryNonce ", "err", err, "chain", chain)
		return sdk.ZeroInt(), err
	}

	return price, nil
}

func convertProtoTxStatusToChainnodeStatus(status proto.TxStatus) uint64 {
	var s uint64
	switch status {
	case proto.TxStatus_NotFound:
		s = chainnode.StatusNotFound
	case proto.TxStatus_Pending:
		s = chainnode.StatusPending
	case proto.TxStatus_Failed:
		s = chainnode.StatusFailed
	case proto.TxStatus_Success:
		s = chainnode.StatusSuccess
	case proto.TxStatus_ContractExecuteFailed:
		s = chainnode.StatusContractExecuteFailed
	case proto.TxStatus_Other:
		s = chainnode.StatusOther
	}
	return s
}

func loadExtAccountTransaction(reply *proto.QueryAccountTransactionReply) (*chainnode.ExtAccountTransaction, []byte, error) {
	tx := &chainnode.ExtAccountTransaction{
		Hash:            reply.TxHash,
		From:            reply.From,
		To:              reply.To,
		BlockHeight:     reply.BlockHeight,
		BlockTime:       reply.BlockTime,
		Memo:            reply.Memo,
		Nonce:           reply.Nonce,
		ContractAddress: reply.ContractAddress,
	}

	tx.Status = convertProtoTxStatusToChainnodeStatus(reply.TxStatus)
	// TODO(kai.wen): In eth, when tx status is Pending(1), the reply only has TxStatus,
	// not other field. So it will return errInvalidNumberString. Fix it.
	//var ok bool
	//if tx.Amount, ok = sdk.NewIntFromString(reply.Amount); !ok {
	//	return nil, nil, errInvalidNumberString
	//}
	//if tx.GasLimit, ok = sdk.NewIntFromString(reply.GasLimit); !ok {
	//	return nil, nil, errInvalidNumberString
	//}
	//if tx.GasPrice, ok = sdk.NewIntFromString(reply.GasPrice); !ok {
	//	return nil, nil, errInvalidNumberString
	//}
	//if tx.CostFee, ok = sdk.NewIntFromString(reply.CostFee); !ok {
	//	return nil, nil, errInvalidNumberString
	//}

	tx.Amount = stringToInt(reply.Amount)
	tx.GasLimit = stringToInt(reply.GasLimit)
	tx.GasPrice = stringToInt(reply.GasPrice)
	tx.CostFee = stringToInt(reply.CostFee)

	return tx, reply.SignHash, nil
}

func stringToInt(s string) sdk.Int {
	if s == "" {
		return sdk.ZeroInt()
	}

	value, ok := sdk.NewIntFromString(s)
	if !ok {
		return sdk.ZeroInt()
	}

	return value
}

func loadExtUtxoTransaction(reply *proto.QueryUtxoTransactionReply) (*chainnode.ExtUtxoTransaction, [][]byte, error) {
	tx := &chainnode.ExtUtxoTransaction{
		Hash:        reply.TxHash,
		BlockHeight: reply.BlockHeight,
		BlockTime:   reply.BlockTime,
	}
	tx.Status = convertProtoTxStatusToChainnodeStatus(reply.TxStatus)

	tx.CostFee = stringToInt(reply.CostFee)
	vin := make([]*sdk.UtxoIn, len(reply.Vins))
	for i, v := range reply.Vins {
		vin[i] = &sdk.UtxoIn{
			Hash:    v.Hash,
			Index:   uint64(v.Index),
			Amount:  sdk.NewInt(v.Amount),
			Address: v.Address,
		}
	}
	tx.Vins = vin

	vout := make([]*sdk.UtxoOut, len(reply.Vouts))
	for i, v := range reply.Vouts {
		vout[i] = &sdk.UtxoOut{
			Address: v.Address,
			Amount:  sdk.NewInt(v.Amount),
		}
	}
	tx.Vouts = vout

	signHashes := make([][]byte, 0, len(reply.SignHashes))

	for _, hash := range reply.SignHashes {
		signHashes = append(signHashes, hash)
	}

	return tx, signHashes, nil
}

func (ch *Chainnode) QueryUtxoTransaction(chain, symbol, hash string, asynMode bool) (*chainnode.ExtUtxoTransaction, error) {
	reply, err := ch.client.QueryUtxoTransaction(context.TODO(), &proto.QueryTransactionRequest{
		Chain:     chain,
		Symbol:    symbol,
		TxHash:    hash,
		AsyncMode: asynMode,
	})
	if err != nil {
		ch.logger.Error("QueryUtxoTransaction fails", "err", err, "chain", chain, "symbol", symbol, "Hash", hash)
		return nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryUtxoTransaction fails", "err", err, "chain", chain, "symbol", symbol, "Hash", hash)
		return nil, err
	}

	tx, _, err := loadExtUtxoTransaction(reply)
	return tx, err
}

func (ch *Chainnode) QueryAccountTransaction(chain, symbol, hash string, asynMode bool) (*chainnode.ExtAccountTransaction, error) {
	reply, err := ch.client.QueryAccountTransaction(context.TODO(), &proto.QueryTransactionRequest{
		Chain:     chain,
		Symbol:    symbol,
		TxHash:    hash,
		AsyncMode: asynMode,
	})
	if err != nil {
		ch.logger.Error("QueryAccountTransaction fails", "err", err, "chain", chain, "symbol", symbol, "Hash", hash)
		return nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryAccountTransaction fails", "err", err, "chain", chain, "symbol", symbol, "Hash", hash)
		return nil, err
	}
	tx, _, err := loadExtAccountTransaction(reply)
	return tx, err

}

// CreateTransaction for a transaction,
func (ch *Chainnode) CreateUtxoTransaction(chain, symbol string, transaction *chainnode.ExtUtxoTransaction) (
	transactionData []byte, signHashs [][]byte, err error) {
	vins := transaction.Vins
	protoVin := make([]*proto.Vin, len(vins))
	fee := big.NewInt(0)
	for i, v := range vins {
		protoVin[i] = &proto.Vin{
			Hash:    v.Hash,
			Index:   uint32(v.Index),
			Amount:  v.Amount.Int64(),
			Address: v.Address,
		}
		fee = fee.Add(fee, v.Amount.BigInt())
	}
	vouts := transaction.Vouts
	protoVout := make([]*proto.Vout, len(vouts))
	for i, v := range vouts {
		protoVout[i] = &proto.Vout{
			Address: v.Address,
			Amount:  v.Amount.Int64(),
		}
		fee = fee.Sub(fee, v.Amount.BigInt())
	}

	if fee.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, errors.New("fee is negative")
	}

	reply, err := ch.client.CreateUtxoTransaction(context.TODO(), &proto.CreateUtxoTransactionRequest{
		Chain:  chain,
		Symbol: symbol,
		Vins:   protoVin,
		Vouts:  protoVout,
		Fee:    fee.String(),
	})
	if err != nil {
		ch.logger.Error("CreateUtxoTransaction fails", "err", err, "chain", chain, "symbol", symbol, "utxoin", vins, "utxoout", vouts)
		return nil, nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("CreateUtxoTransaction fails", "err", err, "chain", chain, "symbol", symbol, "utxoin", vins, "utxoout", vouts)
		return nil, nil, err
	}

	return reply.TxData, reply.SignHashes, nil
}

// CreateTransaction for a transaction,
func (ch *Chainnode) CreateAccountTransaction(chain, symbol, contractAddress string, transaction *chainnode.ExtAccountTransaction) (
	[]byte, []byte, error) {

	from := transaction.From
	to := transaction.To
	nonce := transaction.Nonce
	amount := transaction.Amount.String()
	gasPrice := transaction.GasPrice.String()
	gasLimit := transaction.GasLimit.String()
	memo := transaction.Memo

	reply, err := ch.client.CreateAccountTransaction(context.TODO(), &proto.CreateAccountTransactionRequest{
		Chain:           chain,
		Symbol:          symbol,
		From:            from,
		To:              to,
		Nonce:           nonce,
		Amount:          amount,
		GasLimit:        gasLimit,
		GasPrice:        gasPrice,
		ContractAddress: contractAddress,
		Memo:            memo,
	})
	if err != nil {
		ch.logger.Error("CreateAccountTransaction fails", "err", err, "chain", chain, "symbol", symbol, "from", from, "to", to, "amount", amount, "memo", memo, "contractAddress", contractAddress, "gasLimit", gasLimit, "gasPrice", gasPrice, "nonce", nonce)
		return nil, nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("CreateAccountTransaction fails", "err", err, "chain", chain, "symbol", symbol, "from", from, "to", to, "amount", amount, "memo", memo, "contractAddress", contractAddress, "gasLimit", gasLimit, "gasPrice", gasPrice, "nonce", nonce)
		return nil, nil, err
	}

	return reply.TxData, reply.SignHash, nil
}

// CreateSignedTransaction for a transaction with signature, pubkey is only for BTC, signatures for multi-Vin in BTC
func (ch *Chainnode) CreateUtxoSignedTransaction(chain, symbol string, raw []byte, signatures, pubKeys [][]byte) (
	[]byte, []byte, error) {
	reply, err := ch.client.CreateUtxoSignedTransaction(context.TODO(), &proto.CreateUtxoSignedTransactionRequest{
		Chain:      chain,
		Symbol:     symbol,
		TxData:     raw,
		PublicKeys: pubKeys,
		Signatures: signatures,
	})
	if err != nil {
		ch.logger.Error("CreateUtxoSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "txdata", raw, "pubkeys", pubKeys, "signatures", signatures)
		return nil, nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("CreateUtxoSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "txdata", raw, "pubkeys", pubKeys, "signatures", signatures)
		return nil, nil, err
	}
	return reply.SignedTxData, reply.Hash, err
}

func (ch *Chainnode) CreateAccountSignedTransaction(chain, symbol string, raw []byte, signatures, pubKeys []byte) (
	[]byte, []byte, error) {
	reply, err := ch.client.CreateAccountSignedTransaction(context.TODO(), &proto.CreateAccountSignedTransactionRequest{
		Chain:     chain,
		Symbol:    symbol,
		TxData:    raw,
		PublicKey: pubKeys,
		Signature: signatures,
	})
	if err != nil {
		ch.logger.Error("CreateAccountSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "txdata", raw, "pubkeys", pubKeys, "signatures", signatures)
		return nil, nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("CreateAccountSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "txdata", raw, "pubkeys", pubKeys, "signatures", signatures)
		return nil, nil, err
	}
	return reply.SignedTxData, reply.Hash, err
}

// VerifyUtxoSignedTransaction for an address and a signed transaction bytes
func (ch *Chainnode) VerifyUtxoSignedTransaction(chain, symbol string, addresses []string, signedTxData []byte, vins []*sdk.UtxoIn) (bool, error) {
	ins := convertSdkUtxoInsToProtoVins(vins)
	reply, err := ch.client.VerifyUtxoSignedTransaction(context.TODO(), &proto.VerifySignedTransactionRequest{
		Chain:        chain,
		Symbol:       symbol,
		Addresses:    addresses,
		SignedTxData: signedTxData,
		Vins:         ins,
	})
	if err != nil {
		ch.logger.Error("VerifyUtxoSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "addresses", addresses)
		return false, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("VerifyUtxoSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "addresses", addresses)
		return false, err
	}
	return reply.Verified, err
}

// VerifySignedTransaction for an address and a signed transaction bytes
func (ch *Chainnode) VerifyAccountSignedTransaction(chain, symbol string, address string, signedTxData []byte) (bool, error) {
	reply, err := ch.client.VerifyAccountSignedTransaction(context.TODO(), &proto.VerifySignedTransactionRequest{
		Chain:        chain,
		Symbol:       symbol,
		Addresses:    []string{address},
		SignedTxData: signedTxData,
	})
	if err != nil {
		ch.logger.Error("VerifyAccountSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "address", address)
		return false, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("VerifyAccountSignedTransaction fails", "err", err, "chain", chain, "symbol", symbol, "address", address)
		return false, err
	}
	return reply.Verified, err
}

// QueryAccountTransactionFromSignedData for an address and a signed transaction bytes
func (ch *Chainnode) QueryAccountTransactionFromSignedData(chain, symbol string, signedTxData []byte) (*chainnode.ExtAccountTransaction, error) {
	reply, err := ch.client.QueryAccountTransactionFromSignedData(context.TODO(), &proto.QueryTransactionFromSignedDataRequest{
		Chain:        chain,
		Symbol:       symbol,
		SignedTxData: signedTxData,
	})
	if err != nil {
		ch.logger.Error("QueryAccountTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "signedTxData", signedTxData)
		return nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryAccountTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "signedTxData", signedTxData)
		return nil, err
	}
	tx, _, err := loadExtAccountTransaction(reply)
	return tx, err
}

// QueryUtxoTransactionFromSignedData for an address and a signed transaction bytes
func (ch *Chainnode) QueryUtxoTransactionFromSignedData(chain, symbol string, signedTxData []byte, vins []*sdk.UtxoIn) (*chainnode.ExtUtxoTransaction, error) {
	ins := convertSdkUtxoInsToProtoVins(vins)
	reply, err := ch.client.QueryUtxoTransactionFromSignedData(context.TODO(), &proto.QueryTransactionFromSignedDataRequest{
		Chain:        chain,
		Symbol:       symbol,
		SignedTxData: signedTxData,
		Vins:         ins,
	})
	if err != nil {
		ch.logger.Error("QueryUtxoTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "signedTxData", signedTxData)
		return nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryUtxoTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "signedTxData", signedTxData)
		return nil, err
	}

	tx, _, err := loadExtUtxoTransaction(reply)
	return tx, err
}

// QueryAccountTransactionFromData for an address and a signed transaction bytes
func (ch *Chainnode) QueryAccountTransactionFromData(chain, symbol string, rawData []byte) (*chainnode.ExtAccountTransaction, []byte, error) {
	reply, err := ch.client.QueryAccountTransactionFromData(context.TODO(), &proto.QueryTransactionFromDataRequest{
		Chain:   chain,
		Symbol:  symbol,
		RawData: rawData,
	})

	if err != nil {
		ch.logger.Error("QueryAccountTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "rawData", rawData)
		return nil, nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryAccountTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "rawData", rawData)
		return nil, nil, err
	}

	return loadExtAccountTransaction(reply)
}

// VerifySignedTransaction for an address and a signed transaction bytes
func (ch *Chainnode) QueryUtxoTransactionFromData(chain, symbol string, rawData []byte, vins []*sdk.UtxoIn) (*chainnode.ExtUtxoTransaction, [][]byte, error) {
	ins := convertSdkUtxoInsToProtoVins(vins)
	reply, err := ch.client.QueryUtxoTransactionFromData(context.TODO(), &proto.QueryTransactionFromDataRequest{
		Chain:   chain,
		Symbol:  symbol,
		RawData: rawData,
		Vins:    ins,
	})
	if err != nil {
		ch.logger.Error("QueryUtxoTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "rawData", rawData)
		return nil, nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryUtxoTransactionFromData fails", "err", err, "chain", chain, "symbol", symbol, "rawData", rawData)
		return nil, nil, err
	}

	tx, bz, err := loadExtUtxoTransaction(reply)
	ch.logger.Info("QueryUtxoTransactionFromData result", "tx", tx, "err", err)
	return tx, bz, err
}

// BroadcastTransaction broadcasts the signedTransaction
func (ch *Chainnode) BroadcastTransaction(chain, symbol string, signedTxData []byte) (
	transactionHash string, err error) {
	reply, err := ch.client.BroadcastTransaction(context.TODO(), &proto.BroadcastTransactionRequest{
		Chain:        chain,
		Symbol:       symbol,
		SignedTxData: signedTxData,
	})
	if err != nil {
		ch.logger.Error("BroadcastTransaction fails", "err", err, "symbol", symbol)
		return "", err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("BroadcastTransaction fails", "err", err, "symbol", symbol)
		return "", err
	}
	return reply.TxHash, err
}

// VerifySignedTransaction for an address and a signed transaction bytes
func (ch *Chainnode) QueryUtxoInsFromData(chain, symbol string, data []byte) ([]*sdk.UtxoIn, error) {
	reply, err := ch.client.QueryUtxoInsFromData(context.TODO(), &proto.QueryUtxoInsFromDataRequest{
		Chain:  chain,
		Symbol: symbol,
		Data:   data,
	})
	if err != nil {
		ch.logger.Error("QueryUtxoInsFromData fails", "err", err, "chain", chain, "symbol", symbol, "data", data)
		return nil, err
	}
	if err := validateReply(reply); err != nil {
		ch.logger.Error("QueryUtxoInsFromData fails", "err", err, "chain", chain, "symbol", symbol, "data", data)
		return nil, err
	}

	ins := convertProtoVinsToSdkUtxoIns(reply.Vins)
	return ins, nil
}

func convertSdkUtxoInsToProtoVins(vins []*sdk.UtxoIn) []*proto.Vin {
	ins := make([]*proto.Vin, len(vins))
	for i, vin := range vins {
		ins[i] = &proto.Vin{
			Hash:    vin.Hash,
			Index:   uint32(vin.Index),
			Address: vin.Address,
			Amount:  vin.Amount.Int64(),
		}
	}
	return ins
}

func convertProtoVinsToSdkUtxoIns(vins []*proto.Vin) []*sdk.UtxoIn {
	ins := make([]*sdk.UtxoIn, len(vins))
	for i, vin := range vins {
		ins[i] = &sdk.UtxoIn{
			Hash:    vin.Hash,
			Index:   uint64(vin.Index),
			Address: vin.Address,
			Amount:  sdk.NewInt(vin.Amount),
		}
	}
	return ins
}
