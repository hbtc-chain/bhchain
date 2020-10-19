/*
 * *******************************************************************
 * @项目名称: types
 * @文件名称: key.go
 * @Date: 2019/06/05
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

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
