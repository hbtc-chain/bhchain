package types

const (
	// module name
	ModuleName = "token"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// Parameter store default parameter store
	DefaultParamspace = ModuleName

	// query endpoints supported by the nameservice Querier
	QueryToken      = "token"
	QuerySymbols    = "symbols"
	QueryDecimal    = "decimal"
	QueryTokens     = "tokens"
	QueryParameters = "parameters"

	TypeMsgSynGasPrice = "token-syngasprice"
)
