package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hbtc-chain/bhchain/base58"

	"github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/bech32"
	"gopkg.in/yaml.v2"
)

const (
	// Constants defined here are the defaults value for address.
	// You can use the specific values for your project.
	// Add the follow lines to the `main()` of your server.
	//
	//	config := sdk.GetConfig()
	//	config.SetBech32PrefixForAccount(yourBech32PrefixAccAddr, yourBech32PrefixAccPub)
	//	config.SetBech32PrefixForValidator(yourBech32PrefixValAddr, yourBech32PrefixValPub)
	//	config.SetBech32PrefixForConsensusNode(yourBech32PrefixConsAddr, yourBech32PrefixConsPub)
	//	config.SetCoinType(yourCoinType)
	//	config.SetFullFundraiserPath(yourFullFundraiserPath)
	//	config.Seal()

	// AddrLen defines a valid address length
	AddrLen = 20
	// Bech32PrefixAccAddr defines the Bech32 prefix of an CU's address
	Bech32MainPrefix = "hbc"

	// bht in https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	CoinType = 496

	// BIP44Prefix is the parts of the BIP44 HD path that are fixed by
	// what we used during the fundraiser.
	FullFundraiserPath = "44'/496'/0'/0/0"

	// PrefixAccount is the prefix for CU keys
	PrefixAccount = "acc"
	// PrefixValidator is the prefix for validator keys
	PrefixValidator = "val"
	// PrefixConsensus is the prefix for consensus keys
	PrefixConsensus = "cons"
	// PrefixPublic is the prefix for public keys
	PrefixPublic = "pub"
	// PrefixOperator is the prefix for operator keys
	PrefixOperator = "oper"

	// PrefixAddress is the prefix for addresses
	PrefixAddress = "addr"

	// Bech32PrefixAccAddr defines the Bech32 prefix of an CU's address
	Bech32PrefixAccAddr = Bech32MainPrefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an CU's public key
	Bech32PrefixAccPub = Bech32MainPrefix + PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32MainPrefix + PrefixValidator + PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32MainPrefix + PrefixValidator + PrefixOperator + PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32MainPrefix + PrefixValidator + PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32MainPrefix + PrefixValidator + PrefixConsensus + PrefixPublic
)

type CUType int

const (
	CUTypeUser CUType = 0x1 //用户地址
	CUTypeOp   CUType = 0x2 //运营地址
	CUTypeORG  CUType = 0x3 //机构地址
)

// Address is a common interface for different types of addresses used by the SDK
type Address interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}

// Ensure that different address types implement the interface
var _ Address = CUAddress{}
var _ Address = ValAddress{}
var _ Address = ConsAddress{}

var _ yaml.Marshaler = CUAddress{}
var _ yaml.Marshaler = ValAddress{}
var _ yaml.Marshaler = ConsAddress{}

// ----------------------------------------------------------------------------
// CU
// ----------------------------------------------------------------------------

// CUAddress a wrapper around bytes meant to represent an CU address.
// When marshaled to a string or JSON, it uses Base58.
// Implement address interface
type CUAddress []byte

// CUAddressFromHex creates an CUAddress from a hex string.
func CUAddressFromHex(address string) (addr CUAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding hex address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return CUAddress(bz), nil
}

// CUAddressFromBase58 creates an AccAddress from a base58 string prefixed with "HBT".
func CUAddressFromBase58(address string) (addr CUAddress, err error) {
	// blank input get CUAddress{} without error
	if len(strings.TrimSpace(address)) == 0 {
		return CUAddress{}, nil
	}

	// TODO uncomment
	if len(strings.TrimSpace(address)) < base58.AddrPrefixLen || address[:base58.AddrPrefixLen] != base58.AddrStrPrefix {
		return nil, errors.New(fmt.Sprintf("invalid cuaddress:%v with prefixed !=HBC", address))
	}

	bz, version, err := base58.CheckDecode(address)

	if err != nil {
		return CUAddress{}, err
	}

	if len(bz) != (AddrLen + base58.AddrPrefixLen - 1) { //?
		return nil, errors.New("Incorrect address length")
	}

	prefix := make([]byte, 0, base58.AddrPrefixLen)
	prefix = append(prefix, version)
	prefix = append(prefix, bz[:base58.AddrPrefixLen-1]...)

	if !bytes.Equal(prefix, base58.AddrBytePrefix) {
		return CUAddress{}, errors.New("string is not prefixed with `HBC`")
	}
	return CUAddress(bz[2:]), nil
}

func CUAddressFromPubKey(pubKey crypto.PubKey) CUAddress {
	return CUAddress(pubKey.Address().Bytes())
}

func CUAddressFromByte(b []byte) CUAddress {
	if len(b) != AddrLen {
		return CUAddress{}
	}
	return CUAddress(b)
}

// Returns boolean for whether CUAddress equal to another address
func (ca CUAddress) Equals(ca2 Address) bool {
	if ca.Empty() && ca2.Empty() {
		return true
	}

	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Returns boolean for whether an CUAddress is empty
func (ca CUAddress) Empty() bool {
	if ca == nil {
		return true
	}

	ca2 := CUAddress{}
	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ca CUAddress) Marshal() ([]byte, error) {
	return ca, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ca *CUAddress) Unmarshal(data []byte) error {
	*ca = data
	return nil
}

// MarshalJSON marshals to JSON using Base58.
func (ca CUAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ca.String())
}

// UnmarshalJSON unmarshals from JSON assuming Base58 encoding.
func (ca *CUAddress) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	ca2, err := CUAddressFromBase58(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// Bytes returns the raw address bytes.
func (ca CUAddress) Bytes() []byte {
	return ca
}

// String implements the Stringer interface.

func (ca CUAddress) String() string {
	if len(ca) != AddrLen {
		return ""
	}
	b := make([]byte, 0, base58.AddrPrefixLen+len(ca)+4)
	b = append(b, base58.AddrBytePrefix...) //add 'HBC' prefix
	b = append(b, ca[:]...)
	cksum := base58.Checksum(b)
	b = append(b, cksum[:]...)
	return base58.Encode(b)
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (ca CUAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(ca.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ca)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(ca))))
	}
}

func (ca CUAddress) IsValidAddr() bool {
	if len(ca) != AddrLen {
		return false
	}
	return IsValidAddr(ca.String())
}

func CosmosAddressToCUAddress(cosmosAddr Address) CUAddress {
	switch cosmosAddr.(type) {
	case ConsAddress:
		return CUAddress(cosmosAddr.(ConsAddress))
	case ValAddress:
		return CUAddress(cosmosAddr.(ValAddress))
	case CUAddress:
		return cosmosAddr.(CUAddress)
	default:
		return nil
	}
}

func NewCUAddress() CUAddress {
	pubKey := secp256k1.GenPrivKey().PubKey()
	return CUAddress(pubKey.Address())
}

func PubkeyToString(pubkey crypto.PubKey) string {
	return "BHPubKey:" + base58.Encode(pubkey.Bytes())
}

func IsValidAddr(addr string) bool {
	if len(addr) <= base58.AddrPrefixLen {
		return false
	}
	if addr[:3] != base58.AddrStrPrefix {
		return false
	}
	_, _, err := base58.CheckDecode(addr)
	return err == nil
}

type CUAddressList []CUAddress

func (l CUAddressList) Len() int           { return len(l) }
func (l CUAddressList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l CUAddressList) Less(i, j int) bool { return bytes.Compare(l[i], l[j]) == -1 }

func (l CUAddressList) Join() string {
	l2 := make([]string, len(l))
	for i, t := range l {
		l2[i] = t.String()
	}
	return strings.Join(l2, ",")
}

// Any is a method on CUAddressList that returns true if at least one member of the list satisfies a function. It returns false if the list is empty.
func (l CUAddressList) Any(f func(CUAddress) bool) bool {
	for _, t := range l {
		if f(t) {
			return true
		}
	}
	return false
}

func (l CUAddressList) Contains(target Address) bool {
	return l.Any(func(address CUAddress) bool {
		return address.Equals(target)
	})
}

// VerifyAddressFormat verifies that the provided bytes form a valid address
// according to the default address rules or a custom address verifier set by
// GetConfig().SetAddressVerifier()
func VerifyAddressFormat(bz []byte) error {
	verifier := GetConfig().GetAddressVerifier()
	if verifier != nil {
		return verifier(bz)
	}
	if len(bz) != AddrLen {
		return errors.New("Incorrect address length")
	}
	return nil
}

//// CUAddressFromBech32 creates an CUAddress from a Bech32 string.
//func CUAddressFromBech32(address string) (addr CUAddress, err error) {
//	if len(strings.TrimSpace(address)) == 0 {
//		return CUAddress{}, nil
//	}
//
//	bech32PrefixAccAddr := GetConfig().GetBech32AccountAddrPrefix()
//
//	bz, err := GetFromBech32(address, bech32PrefixAccAddr)
//	if err != nil {
//		return nil, err
//	}
//
//	err = VerifyAddressFormat(bz)
//	if err != nil {
//		return nil, err
//	}
//
//	return CUAddress(bz), nil
//}

// MarshalYAML marshals to YAML using Bech32.
func (ca CUAddress) MarshalYAML() (interface{}, error) {
	return ca.String(), nil
}

// UnmarshalYAML unmarshals from JSON assuming Bech32 encoding.
func (ca *CUAddress) UnmarshalYAML(data []byte) error {
	var s string
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	ca2, err := CUAddressFromBase58(s)
	if err != nil {
		return err
	}
	*ca = ca2
	return nil
}

// ----------------------------------------------------------------------------
// validator operator
// ----------------------------------------------------------------------------

// ValAddress defines a wrapper around bytes meant to present a validator's
// operator. When marshaled to a string or JSON, it uses Bech32.
type ValAddress []byte

// ValAddressFromHex creates a ValAddress from a hex string.
func ValAddressFromHex(address string) (addr ValAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding Bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// ValAddressFromBech32 creates a ValAddress from a Bech32 string.
func ValAddressFromBech32(address string) (addr ValAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ValAddress{}, nil
	}

	bech32PrefixValAddr := GetConfig().GetBech32ValidatorAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}

	err = VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// Returns boolean for whether two ValAddresses are Equal
func (va ValAddress) Equals(va2 Address) bool {
	if va.Empty() && va2.Empty() {
		return true
	}

	return bytes.Equal(va.Bytes(), va2.Bytes())
}

// Returns boolean for whether an CUAddress is empty
func (va ValAddress) Empty() bool {
	if va == nil {
		return true
	}

	va2 := ValAddress{}
	return bytes.Equal(va.Bytes(), va2.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (va ValAddress) Marshal() ([]byte, error) {
	return va, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (va *ValAddress) Unmarshal(data []byte) error {
	*va = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (va ValAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(va.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (va ValAddress) MarshalYAML() (interface{}, error) {
	return va.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (va *ValAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	va2, err := ValAddressFromBech32(s)
	if err != nil {
		return err
	}

	*va = va2
	return nil
}

// UnmarshalYAML unmarshals from YAML assuming Bech32 encoding.
func (va *ValAddress) UnmarshalYAML(data []byte) error {
	var s string

	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	va2, err := ValAddressFromBech32(s)
	if err != nil {
		return err
	}

	*va = va2
	return nil
}

// Bytes returns the raw address bytes.
func (va ValAddress) Bytes() []byte {
	return va
}

// String implements the Stringer interface.
func (va ValAddress) String() string {
	if va.Empty() {
		return ""
	}

	bech32PrefixValAddr := GetConfig().GetBech32ValidatorAddrPrefix()

	bech32Addr, err := bech32.ConvertAndEncode(bech32PrefixValAddr, va.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (va ValAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(va.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", va)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(va))))
	}
}

// ----------------------------------------------------------------------------
// consensus node
// ----------------------------------------------------------------------------

// ConsAddress defines a wrapper around bytes meant to present a consensus node.
// When marshaled to a string or JSON, it uses Bech32.
type ConsAddress []byte

// ConsAddressFromHex creates a ConsAddress from a hex string.
func ConsAddressFromHex(address string) (addr ConsAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding Bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ConsAddress(bz), nil
}

// ConsAddressFromBech32 creates a ConsAddress from a Bech32 string.
func ConsAddressFromBech32(address string) (addr ConsAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ConsAddress{}, nil
	}

	bech32PrefixConsAddr := GetConfig().GetBech32ConsensusAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixConsAddr)
	if err != nil {
		return nil, err
	}

	err = VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return ConsAddress(bz), nil
}

// get ConsAddress from pubkey
func GetConsAddress(pubkey crypto.PubKey) ConsAddress {
	return ConsAddress(pubkey.Address())
}

// Returns boolean for whether two ConsAddress are Equal
func (ca ConsAddress) Equals(ca2 Address) bool {
	if ca.Empty() && ca2.Empty() {
		return true
	}

	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Returns boolean for whether an ConsAddress is empty
func (ca ConsAddress) Empty() bool {
	if ca == nil {
		return true
	}

	ca2 := ConsAddress{}
	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ca ConsAddress) Marshal() ([]byte, error) {
	return ca, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ca *ConsAddress) Unmarshal(data []byte) error {
	*ca = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (ca ConsAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ca.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (ca ConsAddress) MarshalYAML() (interface{}, error) {
	return ca.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (ca *ConsAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	ca2, err := ConsAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// UnmarshalYAML unmarshals from YAML assuming Bech32 encoding.
func (ca *ConsAddress) UnmarshalYAML(data []byte) error {
	var s string

	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	ca2, err := ConsAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// Bytes returns the raw address bytes.
func (ca ConsAddress) Bytes() []byte {
	return ca
}

// String implements the Stringer interface.
func (ca ConsAddress) String() string {
	if ca.Empty() {
		return ""
	}

	bech32PrefixConsAddr := GetConfig().GetBech32ConsensusAddrPrefix()

	bech32Addr, err := bech32.ConvertAndEncode(bech32PrefixConsAddr, ca.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (ca ConsAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(ca.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ca)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(ca))))
	}
}

// ----------------------------------------------------------------------------
// auxiliary
// ----------------------------------------------------------------------------

// Bech32ifyAccPub returns a Bech32 encoded string containing the
// Bech32PrefixAccPub prefix for a given CU PubKey.
func Bech32ifyAccPub(pub crypto.PubKey) (string, error) {
	bech32PrefixAccPub := GetConfig().GetBech32AccountPubPrefix()
	return bech32.ConvertAndEncode(bech32PrefixAccPub, pub.Bytes())
}

// MustBech32ifyAccPub returns the result of Bech32ifyAccPub panicing on failure.
func MustBech32ifyAccPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyAccPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// Bech32ifyValPub returns a Bech32 encoded string containing the
// Bech32PrefixValPub prefix for a given validator operator's PubKey.
func Bech32ifyValPub(pub crypto.PubKey) (string, error) {
	bech32PrefixValPub := GetConfig().GetBech32ValidatorPubPrefix()
	return bech32.ConvertAndEncode(bech32PrefixValPub, pub.Bytes())
}

// MustBech32ifyValPub returns the result of Bech32ifyValPub panicing on failure.
func MustBech32ifyValPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyValPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// Bech32ifyConsPub returns a Bech32 encoded string containing the
// Bech32PrefixConsPub prefixfor a given consensus node's PubKey.
func Bech32ifyConsPub(pub crypto.PubKey) (string, error) {
	bech32PrefixConsPub := GetConfig().GetBech32ConsensusPubPrefix()
	return bech32.ConvertAndEncode(bech32PrefixConsPub, pub.Bytes())
}

// MustBech32ifyConsPub returns the result of Bech32ifyConsPub panicing on
// failure.
func MustBech32ifyConsPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyConsPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// GetAccPubKeyBech32 creates a PubKey for an CU with a given public key
// string using the Bech32 Bech32PrefixAccPub prefix.
func GetAccPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	bech32PrefixAccPub := GetConfig().GetBech32AccountPubPrefix()
	bz, err := GetFromBech32(pubkey, bech32PrefixAccPub)
	if err != nil {
		return nil, err
	}

	pk, err = cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// MustGetAccPubKeyBech32 returns the result of GetAccPubKeyBech32 panicing on
// failure.
func MustGetAccPubKeyBech32(pubkey string) (pk crypto.PubKey) {
	pk, err := GetAccPubKeyBech32(pubkey)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetValPubKeyBech32 creates a PubKey for a validator's operator with a given
// public key string using the Bech32 Bech32PrefixValPub prefix.
func GetValPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	bech32PrefixValPub := GetConfig().GetBech32ValidatorPubPrefix()
	bz, err := GetFromBech32(pubkey, bech32PrefixValPub)
	if err != nil {
		return nil, err
	}

	pk, err = cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// MustGetValPubKeyBech32 returns the result of GetValPubKeyBech32 panicing on
// failure.
func MustGetValPubKeyBech32(pubkey string) (pk crypto.PubKey) {
	pk, err := GetValPubKeyBech32(pubkey)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetConsPubKeyBech32 creates a PubKey for a consensus node with a given public
// key string using the Bech32 Bech32PrefixConsPub prefix.
func GetConsPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	bech32PrefixConsPub := GetConfig().GetBech32ConsensusPubPrefix()
	bz, err := GetFromBech32(pubkey, bech32PrefixConsPub)
	if err != nil {
		return nil, err
	}

	pk, err = cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// MustGetConsPubKeyBech32 returns the result of GetConsPubKeyBech32 panicing on
// failure.
func MustGetConsPubKeyBech32(pubkey string) (pk crypto.PubKey) {
	pk, err := GetConsPubKeyBech32(pubkey)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding Bech32 address failed: must provide an address")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}
