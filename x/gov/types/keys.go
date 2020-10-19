package types

import (
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/hbtc-chain/bhchain/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

// Keys for governance store
// Items are stored with the following key: values
//
// - 0x00<proposalID_Bytes>: Proposal
//
// - 0x01<endTime_Bytes><proposalID_Bytes>: activeProposalID
//
// - 0x02<endTime_Bytes><proposalID_Bytes>: inactiveProposalID
//
// - 0x03: nextProposalID
//
// - 0x10<proposalID_Bytes><depositorAddr_Bytes>: Deposit
//
// - 0x20<proposalID_Bytes><voterAddr_Bytes>: Voter
var (
	ProposalsKeyPrefix          = []byte{0x00}
	ActiveProposalQueuePrefix   = []byte{0x01}
	InactiveProposalQueuePrefix = []byte{0x02}
	ProposalIDKey               = []byte{0x03}

	DepositsKeyPrefix = []byte{0x10}

	VotesKeyPrefix = []byte{0x20}

	DaoVotesKeyPrefix = []byte{0x30}
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

// ProposalKey gets a specific proposal from the store
func ProposalKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(ProposalsKeyPrefix, bz...)
}

// ActiveProposalByTimeKey gets the active proposal queue key by endTime
func ActiveProposalByTimeKey(endTime time.Time) []byte {
	return append(ActiveProposalQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

// ActiveProposalQueueKey returns the key for a proposalID in the activeProposalQueue
func ActiveProposalQueueKey(proposalID uint64, endTime time.Time) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)

	return append(ActiveProposalByTimeKey(endTime), bz...)
}

// InactiveProposalByTimeKey gets the inactive proposal queue key by endTime
func InactiveProposalByTimeKey(endTime time.Time) []byte {
	return append(InactiveProposalQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

// InactiveProposalQueueKey returns the key for a proposalID in the inactiveProposalQueue
func InactiveProposalQueueKey(proposalID uint64, endTime time.Time) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)

	return append(InactiveProposalByTimeKey(endTime), bz...)
}

// DepositsKey gets the first part of the deposits key based on the proposalID
func DepositsKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(DepositsKeyPrefix, bz...)
}

// DepositKey key of a specific deposit from the store
func DepositKey(proposalID uint64, depositorAddr sdk.CUAddress) []byte {
	return append(DepositsKey(proposalID), depositorAddr.Bytes()...)
}

// VotesKey gets the first part of the votes key based on the proposalID
func VotesKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(VotesKeyPrefix, bz...)
}

// VoteKey key of a specific vote from the store
func VoteKey(proposalID uint64, voterAddr sdk.CUAddress) []byte {
	return append(VotesKey(proposalID), voterAddr.Bytes()...)
}

// VotesKey gets the first part of the votes key based on the proposalID
func DaoVotesKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(DaoVotesKeyPrefix, bz...)
}

// VoteKey key of a specific vote from the store
func DaoVoteKey(proposalID uint64, voterAddr sdk.CUAddress) []byte {
	return append(DaoVotesKey(proposalID), voterAddr.Bytes()...)
}

// Split keys function; used for iterators

// SplitProposalKey split the proposal key and returns the proposal id
func SplitProposalKey(key []byte) (proposalID uint64) {
	if len(key[1:]) != 8 {
		panic(fmt.Sprintf("unexpected key length (%d ≠ 8)", len(key[1:])))
	}

	return binary.LittleEndian.Uint64(key[1:])
}

// SplitActiveProposalQueueKey split the active proposal key and returns the proposal id and endTime
func SplitActiveProposalQueueKey(key []byte) (proposalID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

// SplitInactiveProposalQueueKey split the inactive proposal key and returns the proposal id and endTime
func SplitInactiveProposalQueueKey(key []byte) (proposalID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

// SplitKeyDeposit split the deposits key and returns the proposal id and depositor address
func SplitKeyDeposit(key []byte) (proposalID uint64, depositorAddr sdk.CUAddress) {
	return splitKeyWithAddress(key)
}

// SplitKeyVote split the votes key and returns the proposal id and voter address
func SplitKeyVote(key []byte) (proposalID uint64, voterAddr sdk.CUAddress) {
	return splitKeyWithAddress(key)
}

// private functions

func splitKeyWithTime(key []byte) (proposalID uint64, endTime time.Time) {
	if len(key[1:]) != 8+lenTime {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(key[1:]), lenTime+8))
	}

	endTime, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		panic(err)
	}
	proposalID = binary.LittleEndian.Uint64(key[1+lenTime:])
	return
}

func splitKeyWithAddress(key []byte) (proposalID uint64, addr sdk.CUAddress) {
	if len(key[1:]) != 8+sdk.AddrLen {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(key), 8+sdk.AddrLen))
	}

	proposalID = binary.LittleEndian.Uint64(key[1:9])
	addr = sdk.CUAddress(key[9:])
	return
}