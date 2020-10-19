package keygen

import (
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/keygen/types"
)

const (
	ModuleName = types.ModuleName

	DefaultParamspace = types.DefaultParamspace
	StoreKey          = types.StoreKey
	RouterKey         = types.RouterKey

	MaxWaitAssignKeyOrders = types.MaxWaitAssignKeyOrders
	MaxPreKeyGenOrders     = types.MaxPreKeyGenOrders
)

var (
	RegisterCodec                  = types.RegisterCodec
	ModuleCdc                      = types.ModuleCdc
	HandleMsgNewOpCUForTest        = handleMsgNewOpCU
	HandleMsgKeyGenForTest         = handleMsgKeyGen
	HandleMsgKeyGenWaitSignForTest = handleMsgKeyGenWaitSign
)

type (
	MsgKeyGen              = types.MsgKeyGen
	MsgKeyGenWaitSign      = types.MsgKeyGenWaitSign
	MsgPreKeyGen           = types.MsgPreKeyGen
	MsgKeyGenFinish        = types.MsgKeyGenFinish
	MsgOpcuMigrationKeyGen = types.MsgOpcuMigrationKeyGen
	MsgNewOpCU             = types.MsgNewOpCU
	CustodianUnit          = exported.CustodianUnit
)
