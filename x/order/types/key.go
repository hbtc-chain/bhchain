package types

const (
	// module name
	ModuleName = "order"

	RouterKey = ModuleName
	// StoreKey to be used when creating the KVStore
	// use same store with module CustodianUnit
	StoreKey = ModuleName

	QuerierRoute = ModuleName

	DefaultParamspace = ModuleName

	// query endpoints supported by the nameservice Querier
	QueryOrder       = "order"
	QueryProcessList = "processList"
)

var (
	TagKeyOrderer = "orderer"

	// OrderStoreKeyPrefix prefix for order store
	// order key : OrderStoreKeyPrefix + bhaddress + OrderStoreKeyPrefix + orderID
	OrderStoreKeyPrefix = []byte{0x01, 0x02}

	ProcessOrderListKey = []byte("processOrderList")

	ProcessOrderStoreKeyPrefix = []byte{0x03, 0x04}
	// OrderNumber Key : OrderIDKey + OrderStoreKeyPrefix + cuaddress
	OrderIDKey = []byte("orderIdKey")
)

type QueryOrderParams struct {
	OrderID string
}

func NewQueryOrderParams(orderID string) QueryOrderParams {
	return QueryOrderParams{OrderID: orderID}
}
