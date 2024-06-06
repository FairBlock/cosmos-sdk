package types

import (
	"cosmossdk.io/collections"
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"time"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// Version defines the current version the IBC module supports
	Version = "keyshare-1"

	// PortID is the default port id that module binds to
	PortID = "gov"

	// ChannelID is the default channel id that module will use to transmit IBC packets.
	ChannelID = "channel-0"
)

const (
	ReqQueueKeyPrefix    = "ReqQueue/value/"
	SignalQueueKeyPrefix = "SignalQueue/value/"
)

var (
	ProposalsKeyPrefix            = collections.NewPrefix(0)  // ProposalsKeyPrefix stores the proposals raw bytes.
	ActiveProposalQueuePrefix     = collections.NewPrefix(1)  // ActiveProposalQueuePrefix stores the active proposals.
	InactiveProposalQueuePrefix   = collections.NewPrefix(2)  // InactiveProposalQueuePrefix stores the inactive proposals.
	ProposalIDKey                 = collections.NewPrefix(3)  // ProposalIDKey stores the sequence representing the next proposal ID.
	VotingPeriodProposalKeyPrefix = collections.NewPrefix(4)  // VotingPeriodProposalKeyPrefix stores which proposals are on voting period.
	DepositsKeyPrefix             = collections.NewPrefix(16) // DepositsKeyPrefix stores deposits.
	VotesKeyPrefix                = collections.NewPrefix(32) // VotesKeyPrefix stores the votes of proposals.
	ParamsKey                     = collections.NewPrefix(48) // ParamsKey stores the module's params.
	ConstitutionKey               = collections.NewPrefix(49) // ConstitutionKey stores a chain's constitution.

	PortKey    = []byte("gov-port-")
	ChannelKey = []byte("gov-channel-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

// GetProposalIDBytes returns the byte representation of the proposalID
func GetProposalIDBytes(proposalID uint64) (proposalIDBz []byte) {
	proposalIDBz = make([]byte, 8)
	binary.BigEndian.PutUint64(proposalIDBz, proposalID)
	return
}

// GetProposalIDFromBytes returns proposalID in uint64 format from a byte array
func GetProposalIDFromBytes(bz []byte) (proposalID uint64) {
	return binary.BigEndian.Uint64(bz)
}

// ProposalKey gets a specific proposal from the store
func ProposalKey(proposalID uint64) []byte {
	return append(ProposalsKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// VotingPeriodProposalKey gets if a proposal is in voting period.
func VotingPeriodProposalKey(proposalID uint64) []byte {
	return append(VotingPeriodProposalKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// ActiveProposalByTimeKey gets the active proposal queue key by endTime
func ActiveProposalByTimeKey(endTime time.Time) []byte {
	return append(ActiveProposalQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

// ActiveProposalQueueKey returns the key for a proposalID in the activeProposalQueue
func ActiveProposalQueueKey(proposalID uint64, endTime time.Time) []byte {
	return append(ActiveProposalByTimeKey(endTime), GetProposalIDBytes(proposalID)...)
}

// InactiveProposalByTimeKey gets the inactive proposal queue key by endTime
func InactiveProposalByTimeKey(endTime time.Time) []byte {
	return append(InactiveProposalQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

// InactiveProposalQueueKey returns the key for a proposalID in the inactiveProposalQueue
func InactiveProposalQueueKey(proposalID uint64, endTime time.Time) []byte {
	return append(InactiveProposalByTimeKey(endTime), GetProposalIDBytes(proposalID)...)
}

// DepositsKey gets the first part of the deposits key based on the proposalID
func DepositsKey(proposalID uint64) []byte {
	return append(DepositsKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// DepositKey key of a specific deposit from the store
func DepositKey(proposalID uint64, depositorAddr sdk.AccAddress) []byte {
	return append(DepositsKey(proposalID), address.MustLengthPrefix(depositorAddr.Bytes())...)
}

// VotesKey gets the first part of the votes key based on the proposalID
func VotesKey(proposalID uint64) []byte {
	return append(VotesKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// VoteKey key of a specific vote from the store
func VoteKey(proposalID uint64, voterAddr sdk.AccAddress) []byte {
	return append(VotesKey(proposalID), address.MustLengthPrefix(voterAddr.Bytes())...)
}

// Split keys function; used for iterators

// SplitProposalKey split the proposal key and returns the proposal id
func SplitProposalKey(key []byte) (proposalID uint64) {
	kv.AssertKeyLength(key[1:], 8)

	return GetProposalIDFromBytes(key[1:])
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
func SplitKeyDeposit(key []byte) (proposalID uint64, depositorAddr sdk.AccAddress) {
	return splitKeyWithAddress(key)
}

// SplitKeyVote split the votes key and returns the proposal id and voter address
func SplitKeyVote(key []byte) (proposalID uint64, voterAddr sdk.AccAddress) {
	return splitKeyWithAddress(key)
}

// private functions

func splitKeyWithTime(key []byte) (proposalID uint64, endTime time.Time) {
	kv.AssertKeyLength(key[1:], 8+lenTime)

	endTime, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		panic(err)
	}

	proposalID = GetProposalIDFromBytes(key[1+lenTime:])
	return
}

func splitKeyWithAddress(key []byte) (proposalID uint64, addr sdk.AccAddress) {
	// Both Vote and Deposit store keys are of format:
	// <prefix (1 Byte)><proposalID (8 bytes)><addrLen (1 Byte)><addr_Bytes>
	kv.AssertKeyAtLeastLength(key, 10)
	proposalID = GetProposalIDFromBytes(key[1:9])
	kv.AssertKeyAtLeastLength(key, 11)
	addr = sdk.AccAddress(key[10:])
	return
}

func QueueKey(
	identity string,
) []byte {
	var key []byte

	b := []byte(identity)
	key = append(key, b...)

	return key
}
