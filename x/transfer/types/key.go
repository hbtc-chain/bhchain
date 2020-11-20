package types

import (
	"encoding/binary"

	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	// ModuleName is the name of the transfer module
	ModuleName = "transfer"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName

	MaxKeyNodeHeartbeat = 1000
)

var (
	OrderRetryTimesPrefix           = []byte{0x01}
	OrderRetryEvidenceHandledPrefix = []byte{0x02}

	balanceKeyPrefix     = []byte{0x03}
	holdBalanceKeyPrefix = []byte{0x04}
)

func GetOrderRetryEvidenceHandledKey(txID string, retryTimes uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, retryTimes)
	return append(append(OrderRetryEvidenceHandledPrefix, txID...), buf...)
}

func BalanceKey(addr sdk.CUAddress, symbol string) []byte {
	return append(BalanceKeyPrefix(addr), []byte(symbol)...)
}

func BalanceKeyPrefix(addr sdk.CUAddress) []byte {
	return append(balanceKeyPrefix, addr...)
}

func HoldBalanceKey(addr sdk.CUAddress, symbol string) []byte {
	return append(HoldBalanceKeyPrefix(addr), []byte(symbol)...)
}

func HoldBalanceKeyPrefix(addr sdk.CUAddress) []byte {
	return append(holdBalanceKeyPrefix, addr...)
}

func GetSymbolFromBalanceKey(key []byte) string {
	return string(key[len(balanceKeyPrefix)+sdk.AddrLen:])
}

func GetSymbolFromHoldBalanceKey(key []byte) string {
	return string(key[len(holdBalanceKeyPrefix)+sdk.AddrLen:])
}
