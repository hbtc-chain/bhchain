package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/golang/protobuf/proto"
	dbm "github.com/tendermint/tm-db"
	"golang.org/x/crypto/sha3"
)

var (
	// This is set at compile time. Could be cleveldb, defaults is goleveldb.
	DBBackend = ""

	// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
	// values.
	IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
)

const (
	KiloBytes                 = 1024
	BytesPerUtxoVin           = 150
	BytesPerUtxoVout          = 40
	LimitAccountBasedOrderNum = 1
)

const HashLength = 32

// Hash to identify uniqueness
type Hash [HashLength]byte

func (h Hash) Bytes() []byte {
	return h[:]
}

func BytesToHash(bytes []byte) Hash {
	return sha3.Sum256(bytes)
}

func ProtoToHash(msg proto.Message) Hash {
	bytes, _ := proto.Marshal(msg)
	return BytesToHash(bytes)
}

// SortedJSON takes any JSON and returns it sorted by keys. Also, all white-spaces
// are removed.
// This method can be used to canonicalize JSON to be returned by GetSignBytes,
// e.g. for the ledger integration.
// If the passed JSON isn't valid it will return an error.
func SortJSON(toSortJSON []byte) ([]byte, error) {
	var c interface{}
	err := json.Unmarshal(toSortJSON, &c)
	if err != nil {
		return nil, err
	}
	js, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return js, nil
}

// MustSortJSON is like SortJSON but panic if an error occurs, e.g., if
// the passed JSON isn't valid.
func MustSortJSON(toSortJSON []byte) []byte {
	js, err := SortJSON(toSortJSON)
	if err != nil {
		panic(err)
	}
	return js
}

// Uint64ToBigEndian - marshals uint64 to a bigendian byte slice so it can be sorted
func Uint64ToBigEndian(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func Uint32ToBigEndian(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

// Slight modification of the RFC3339Nano but it right pads all zeros and drops the time zone info
const SortableTimeFormat = "2006-01-02T15:04:05.000000000"

// Formats a time.Time into a []byte that can be sorted
func FormatTimeBytes(t time.Time) []byte {
	return []byte(t.UTC().Round(0).Format(SortableTimeFormat))
}

// Parses a []byte encoded using FormatTimeKey back into a time.Time
func ParseTimeBytes(bz []byte) (time.Time, error) {
	str := string(bz)
	t, err := time.Parse(SortableTimeFormat, str)
	if err != nil {
		return t, err
	}
	return t.UTC().Round(0), nil
}

// NewLevelDB instantiate a new LevelDB instance according to DBBackend.
func NewLevelDB(name, dir string) (db dbm.DB, err error) {
	backend := dbm.GoLevelDBBackend
	if DBBackend == string(dbm.CLevelDBBackend) {
		backend = dbm.CLevelDBBackend
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("couldn't create db: %v", r)
		}
	}()
	return dbm.NewDB(name, backend, dir), err
}

func SizeInKB(size int) Int {
	return NewDec(int64(size)).Quo(NewDec(KiloBytes)).Ceil().TruncateInt()
}

func Majority23(num int) int {
	return 2*num/3 + 1
}

func Majority34(num int) int {
	return 3*num/4 + 1
}

func OneSixthCeil(num int) int {
	return (num + 5) / 6
}

func MaxUint16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

type TxFinishNodeData struct {
	ValidatorAddr string
	CostFee       Int
}

/*EstimateSignedUtxoTxSize, the number of changeback should be counted into numVout.
  for example, if 1 withdrawal to and 1 changeback, numVout =2
*/
func EstimateSignedUtxoTxSize(numVin, numVout int) Int {
	return NewInt(int64(BytesPerUtxoVin*numVin + BytesPerUtxoVout*numVout))
}

func DefaultUtxoWithdrawTxSize() Int {
	return NewInt(int64(BytesPerUtxoVin + BytesPerUtxoVout*2))
}

func DefaultUtxoCollectTxSize() Int {
	return NewInt(int64(BytesPerUtxoVin + BytesPerUtxoVout))
}

// StringsIndex returns the index of the first instance of string `want` in string slice `s`, or -1 if `want` is not present in `s`.
func StringsIndex(s []string, want string) int {
	for i, str := range s {
		if str == want {
			return i
		}
	}
	return -1
}
