package types

import (
	"bytes"

	"github.com/hbtc-chain/bhchain/types"
	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	// module name
	ModuleName = "ibcasset"

	// StoreKey is string representation of the store key for auth
	StoreKey = "ibcasset"

	// RouterKey is the message route for Module
	RouterKey = ModuleName

	// FeeCollectorName the root string for the fee collector CU address
	FeeCollectorName = "fee_collector"

	// QuerierRoute is the querier route for cu
	QuerierRoute = StoreKey

	DefaultParamspace = ModuleName
)

var (
	// AddressStoreKeyPrefix prefix for CU-by-address store
	AddressStoreKeyPrefix = []byte{0x01}

	depositListPrefix = []byte{0x03}
	depositListSep    = []byte{0x04}

	Sep = []byte{0x01}

	ExtAddressPrefix = []byte("extAddress")

	// param key for global CU number
	GlobalCUNumberKey = []byte("globalCUNumber")
)

// AddressStoreKey turn an address to key used to get it from the CU store
// key = prefix + cuaddress
func AddressStoreKey(addr sdk.CUAddress) []byte {
	return append(AddressStoreKeyPrefix, addr.Bytes()...)
}

// key = prefix + symbol + cuaddress
func DepositStoreKey(symbol string, addr sdk.CUAddress, hash string, index uint64) []byte {
	key := DepositStorePrefixKey(symbol, addr)
	key = append(key, []byte(hash)...)
	key = append(key, sdk.Uint64ToBigEndian(index)...)
	return key
}

func DepositStorePrefixKey(symbol string, addr sdk.CUAddress) []byte {
	key := DepositStorePrefixKeyWithAddr(addr)
	key = append(key, symbol...)
	key = append(key, depositListSep...)
	return key
}

func DepositStorePrefixKeyWithAddr(addr sdk.CUAddress) []byte {
	return append(depositListPrefix, addr...)
}

func DecodeSymbolFromDepositListKey(key []byte) string {
	prefixLen := len(depositListPrefix) + types.AddrLen
	key = key[prefixLen:]
	res := bytes.Split(key, depositListSep)
	return string(res[0])
}
