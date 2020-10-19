package types

import (
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
	TradingPairKeyPrefix     = []byte{0x00}
	LiquidityKeyPrefix       = []byte{0x01}
	OrderKeyPrefix           = []byte{0x02}
	UnfinishedOrderKeyPrefix = []byte{0x03}
	RefererKeyPrefix         = []byte{0x04}
	TotalShareKeyPrefix      = []byte{0x05}
	GlobalMaskKeyPrefix      = []byte{0x06}
	AddrMaskKeyPrefix        = []byte{0x07}
)

func TradingPairKey(tokenA, tokenB sdk.Symbol) []byte {
	return append(TradingPairKeyPrefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func AddrLiquidityKeyPrefix(addr sdk.CUAddress) []byte {
	return append(LiquidityKeyPrefix, addr...)
}

func LiquidityKey(tokenA, tokenB sdk.Symbol, addr sdk.CUAddress) []byte {
	prefix := AddrLiquidityKeyPrefix(addr)
	return append(prefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func DecodeTokensFromLiquidityKey(key, prefix []byte) (sdk.Symbol, sdk.Symbol) {
	tokens := strings.Split(string(key[len(prefix):]), "-")
	if len(tokens) != 2 {
		panic("invalid key and prefix")
	}
	return sdk.Symbol(tokens[0]), sdk.Symbol(tokens[1])
}

func OrderKey(orderID string) []byte {
	return append(OrderKeyPrefix, orderID...)
}

func GetUnfinishedOrderKeyPrefix(baseSymbol, quoteSymbol sdk.Symbol, from sdk.CUAddress) []byte {
	return append(UnfinishedOrderKeyPrefix, fmt.Sprintf("%s-%s-%s:", baseSymbol, quoteSymbol, from.String())...)
}

func UnfinishedOrderKey(order *Order) []byte {
	return append(GetUnfinishedOrderKeyPrefix(order.BaseSymbol, order.QuoteSymbol, order.From), order.OrderID...)
}

func GetOrderIDFromUnfinishedOrderKey(key []byte) string {
	strs := strings.Split(string(key), ":")
	if len(strs) != 2 {
		panic("invalid key")
	}
	return strs[1]
}

func RefererKey(addr sdk.CUAddress) []byte {
	return append(RefererKeyPrefix, addr...)
}

func TotalShareKey(tokenA, tokenB sdk.Symbol) []byte {
	return append(TotalShareKeyPrefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func GlobalMaskKey(tokenA, tokenB sdk.Symbol) []byte {
	return append(GlobalMaskKeyPrefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}

func AddrMaskKey(tokenA, tokenB sdk.Symbol, addr sdk.CUAddress) []byte {
	prefix := append(AddrMaskKeyPrefix, addr...)
	return append(prefix, fmt.Sprintf("%s-%s", tokenA.String(), tokenB.String())...)
}
