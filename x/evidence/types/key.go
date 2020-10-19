package types

import (
	"encoding/binary"

	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "evidence"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)

var (
	EvidencePrefixKey             = []byte{0x00}
	ValidatorBehaviourKey         = []byte{0x01}
	ValidatorBehaviourBitArrayKey = []byte{0x02}
	VoteBoxKey                    = []byte{0x03}
	ConfirmedVoteKey              = []byte{0x04}
)

func GetValidatorBehaviourKey(behaviourName string, v sdk.ValAddress) []byte {
	return append(ValidatorBehaviourKey, append([]byte(behaviourName), v.Bytes()...)...)
}

func GetValidatorBehaviourBitArrayPrefixKey(behaviourName string, v sdk.ValAddress) []byte {
	return append(ValidatorBehaviourBitArrayKey, append([]byte(behaviourName), v.Bytes()...)...)
}

// stored by *Consensus* address (not operator address)
func GetValidatorBehaviourBitArrayKey(behaviourName string, v sdk.ValAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(GetValidatorBehaviourBitArrayPrefixKey(behaviourName, v), b...)
}

func GetVoteBoxKey(voteID string) []byte {
	return append(VoteBoxKey, voteID...)
}

func GetConfirmedVoteKeyPrefix(height uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, height)
	return append(ConfirmedVoteKey, b...)
}

func DecodeConfirmedVoteKey(key []byte) (uint64, string) {
	height := binary.BigEndian.Uint64(key[len(ConfirmedVoteKey) : len(ConfirmedVoteKey)+8])
	voteID := string(key[len(ConfirmedVoteKey)+8:])
	return height, voteID
}
