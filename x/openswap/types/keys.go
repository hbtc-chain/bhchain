package types

import (
	"encoding/binary"
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	// ModuleName is the name of this module
	ModuleName = "openswap"

	// RouterKey is used to route governance proposals
	RouterKey = ModuleName

	// StoreKey is the prefix under which we store this module's data
	StoreKey = ModuleName

	// QuerierKey is used to handle abci_query requests
	QuerierKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

var (
	ModuleCUAddress = sdk.CUAddress(crypto.AddressHash([]byte(ModuleName)))
)

var (
	DexKeyPrefix                = []byte{0x00}
	DexIDKey                    = []byte{0x01}
	TradingPairKeyPrefix        = []byte{0x02}
	LiquidityKeyPrefix          = []byte{0x03}
	OrderKeyPrefix              = []byte{0x04}
	UnfinishedOrderKeyPrefix    = []byte{0x05}
	WaitToInsertMatchingKey     = []byte{0x06}
	WaitToRemoveFromMatchingKey = []byte{0x07}
	RefererKeyPrefix            = []byte{0x08}
	TotalShareKeyPrefix         = []byte{0x09}
	GlobalMaskKeyPrefix         = []byte{0x0a}
	AddrMaskKeyPrefix           = []byte{0x0b}
	RepurchaseFundKeyPrefix     = []byte{0x0c}
)

func DexKey(dexID uint32) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	return append(DexKeyPrefix, bz...)
}

func TradingPairKeyPrefixWithDexID(dexID uint32) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	return append(TradingPairKeyPrefix, bz...)
}

func TradingPairKey(dexID uint32, tokenA, tokenB sdk.Symbol) []byte {
	prefix := TradingPairKeyPrefixWithDexID(dexID)
	return append(prefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func AddrLiquidityKeyPrefix(addr sdk.CUAddress) []byte {
	return append(LiquidityKeyPrefix, addr...)
}

func AddrLiquidityKeyPrefixWithDexID(addr sdk.CUAddress, dexID uint32) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	return append(AddrLiquidityKeyPrefix(addr), bz...)
}

func LiquidityKey(addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol) []byte {
	prefix := AddrLiquidityKeyPrefixWithDexID(addr, dexID)
	return append(prefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func DecodeLiquidityKey(key []byte) (uint32, sdk.Symbol, sdk.Symbol) {
	addrPrefixLen := len(LiquidityKeyPrefix) + sdk.AddrLen
	dexID := binary.BigEndian.Uint32(key[addrPrefixLen : addrPrefixLen+4])
	tokens := strings.Split(string(key[addrPrefixLen+4:]), "-")
	if len(tokens) != 2 {
		panic("invalid key and prefix")
	}
	return dexID, sdk.Symbol(tokens[0]), sdk.Symbol(tokens[1])
}

func OrderKey(orderID string) []byte {
	return append(OrderKeyPrefix, orderID...)
}

func UnfinishedOrderKeyPrefixWithAddr(addr sdk.CUAddress) []byte {
	return append(UnfinishedOrderKeyPrefix, addr...)
}

func UnfinishedOrderKeyPrefixWithPair(addr sdk.CUAddress, dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	prefix := append(UnfinishedOrderKeyPrefixWithAddr(addr), bz...)
	return append(prefix, fmt.Sprintf("%s-%s:", baseSymbol.String(), quoteSymbol.String())...)
}

func UnfinishedOrderKey(order *Order) []byte {
	return append(UnfinishedOrderKeyPrefixWithPair(order.From, order.DexID, order.BaseSymbol, order.QuoteSymbol), order.OrderID...)
}

func GetOrderIDFromUnfinishedOrderKey(key []byte) string {
	prefixLen := len(UnfinishedOrderKeyPrefix) + sdk.AddrLen + 4
	strs := strings.Split(string(key[prefixLen:]), ":")
	if len(strs) != 2 {
		panic("invalid key")
	}
	return strs[1]
}

func RefererKey(addr sdk.CUAddress) []byte {
	return append(RefererKeyPrefix, addr...)
}

func TotalShareKey(dexID uint32, tokenA, tokenB sdk.Symbol) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	prefix := append(TotalShareKeyPrefix, bz...)
	return append(prefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func GlobalMaskKey(dexID uint32, tokenA, tokenB sdk.Symbol) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	prefix := append(GlobalMaskKeyPrefix, bz...)
	return append(prefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func AddrMaskKey(addr sdk.CUAddress, dexID uint32, tokenA, tokenB sdk.Symbol) []byte {
	bz := sdk.Uint32ToBigEndian(dexID)
	prefix := append(AddrMaskKeyPrefix, addr...)
	return append(append(prefix, bz...), fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func RepurchaseFundKey(symbol string) []byte {
	return append(RepurchaseFundKeyPrefix, symbol...)
}

func GetSymbolFromRepurchaseFundKey(key []byte) string {
	return string(key[len(RepurchaseFundKeyPrefix):])
}
