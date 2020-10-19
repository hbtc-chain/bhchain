package receipt

const (
	// module name
	ModuleName = "receipt"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName

	// Parameter store default parameter store
	DefaultParamspace = ModuleName
)

var (
	TagKeyReceipt = "receipt"

	// OrderStoreKeyPrefix prefix for order store
	// order key : OrderStoreKeyPrefix + bhaddress + OrderStoreKeyPrefix + orderID
	ReceiptStoreKeyPrefix = []byte{0x01}

	// OrderNumber Key : OrderIDKey + OrderStoreKeyPrefix + cuaddress
	HeightHashKeyPrefix = []byte("HeightHashKey")
)
