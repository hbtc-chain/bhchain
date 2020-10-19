package types

import "encoding/binary"

const (
	// ModuleName is the name of the transfer module
	ModuleName = "transfer"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName
)

var (
	OrderRetryTimesPrefix           = []byte{0x01}
	OrderRetryEvidenceHandledPrefix = []byte{0x02}
)

func GetOrderRetryEvidenceHandledKey(txID string, retryTimes uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, retryTimes)
	return append(append(OrderRetryEvidenceHandledPrefix, txID...), buf...)
}
