// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"bytes"
	"crypto/sha256"
	"errors"
)

// ErrChecksum indicates that the Checksum of a check-encoded string does not verify against
// the Checksum.
var ErrChecksum = errors.New("Checksum error")

// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
var ErrInvalidFormat = errors.New("invalid format: version and/or Checksum bytes missing")

// Checksum: first four bytes of sha256^2
func Checksum(input []byte) (cksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(cksum[:], h2[:4])
	return
}

//func Checksum(input []byte) (cksum [4]byte) {
//	return Checksum(input)
//}

// CheckEncode prepends a version byte and appends a four byte Checksum.
func CheckEncode(input []byte, version byte) string {
	b := make([]byte, 0, 1+len(input)+4)
	b = append(b, version)
	b = append(b, input[:]...)
	cksum := Checksum(b)
	b = append(b, cksum[:]...)
	return Encode(b)
}

// CheckDecode decodes a string that was encoded with CheckEncode and verifies the Checksum.
func CheckDecode(input string) (result []byte, version byte, err error) {
	decoded := Decode(input)
	if len(decoded) < 5 {
		return nil, 0, ErrInvalidFormat
	}
	version = decoded[0]
	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])
	if Checksum(decoded[:len(decoded)-4]) != cksum {
		return nil, 0, ErrChecksum
	}
	payload := decoded[1 : len(decoded)-4]
	result = append(result, payload...)
	return
}

// String implements the Stringer interface.
var AddrBytePrefix = []byte{2, 16, 66}
var AddrStrPrefix = "HBC"
var AddrPrefixLen = len(AddrBytePrefix)

// EthAddrToHBCAddr prepends 'HBC' and appends a four byte Checksum.
func EthAddrToHBCAddr(input []byte) string {
	b := make([]byte, 0, AddrPrefixLen+len(input)+4)
	b = append(b, AddrBytePrefix...) //add 'HBC' prefix
	b = append(b, input[:]...)
	cksum := Checksum(b)
	b = append(b, cksum[:]...)
	return Encode(b)
}

// HBCAddrToEthAddr decodes a BH Address that was encoded with EthAddrToHBCAddr and verifies the Checksum.
func HBCAddrToEthAddr(input string) (result []byte, err error) {
	decoded := Decode(input)
	if len(decoded) < AddrPrefixLen+4 {
		return nil, ErrInvalidFormat
	}
	prefix := []byte(input)[0:AddrPrefixLen]
	if !bytes.Equal(prefix, []byte(AddrStrPrefix)) {
		return nil, ErrInvalidFormat
	}
	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])
	if Checksum(decoded[:len(decoded)-4]) != cksum {
		return nil, ErrChecksum
	}
	payload := decoded[AddrPrefixLen : len(decoded)-4]
	result = append(result, payload...)
	return result, nil
}
