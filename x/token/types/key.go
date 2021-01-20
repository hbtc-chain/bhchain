package types

import sdk "github.com/hbtc-chain/bhchain/types"

const (
	// module name
	ModuleName = "token"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// query endpoints supported by the nameservice Querier
	QueryToken     = "token"
	QueryIBCTokens = "ibc_tokens"

	TypeMsgSynGasPrice = "token-syngasprice"

	DefaultStableCoinWeight  = 0
	DefaultNativeTokenWeight = 10
	DefaultIBCTokenWeight    = 20
	DefaultHrc10TokenWeight  = 30
)

var (
	TokenStoreKeyPrefix = []byte{0x01}
	IBCTokenListKey     = []byte{0x02}
)

func TokenStoreKey(symbol sdk.Symbol) []byte {
	return append(TokenStoreKeyPrefix, []byte(symbol)...)
}
