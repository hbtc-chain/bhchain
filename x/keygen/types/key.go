package types

const (
	// module name
	ModuleName = "keygen"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// Parameter store default parameter store
	DefaultParamspace = ModuleName

	QueryWaitAssignKeys = "waitAssign"

	EventTypeKeyGen              = "key_gen"
	EventTypeKeyGenWaitSign      = "key_gen_waitsign"
	EventTypeKeyGenFinish        = "key_gen_finish"
	EventTypePreKeyGen           = "pre_key_gen"
	EventTypeOpcuMigrationKeyGen = "opcu_migration_key_gen"
	EventTypeKeyNewOPCU          = "new_opcu"

	// in keygenfinish 'sender' is the validator, which send the keygenfinish tx.
	AttributeKeySender   = "sender"
	AttributeKeyFrom     = "from"
	AttributeKeyTo       = "to"
	AttributeKeySymbol   = "symbol"
	AttributeKeyOrderID  = "order_id"
	AttributeKeyOrderIDs = "order_ids"

	MaxWaitAssignKeyOrders = 32
	MaxPreKeyGenOrders     = 5
	MaxKeyNodeHeartbeat    = 1000
)

var (
	WaitAssignKey = []byte("waitAssign")
)
