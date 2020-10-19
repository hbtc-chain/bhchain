// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"math"
	"math/big"
	"strings"
)

//go:generate go run genalphabet.go

var (
	bigRadix  = big.NewInt(58)
	bigZero   = big.NewInt(0)
	big256    = big.NewInt(256)
	big58     = big.NewInt(58)
	indexZero = int64(0)
)

// Decode decodes a modified base58 string to a byte slice.
func Decode(b string) []byte {
	answer := big.NewInt(0)
	j := big.NewInt(1)

	scratch := new(big.Int)
	for i := len(b) - 1; i >= 0; i-- {
		tmp := b58[b[i]]
		if tmp == 255 {
			return []byte("")
		}
		scratch.SetInt64(int64(tmp))
		scratch.Mul(j, scratch)
		answer.Add(answer, scratch)
		j.Mul(j, bigRadix)
	}

	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ {
		if b[numZeros] != alphabetIdx0 {
			break
		}
	}
	flen := numZeros + len(tmpval)
	val := make([]byte, flen)
	copy(val[numZeros:], tmpval)

	return val
}

// Encode encodes a byte slice to a modified base58 string.
func Encode(b []byte) string {
	x := new(big.Int)
	x.SetBytes(b)

	answer := make([]byte, 0, len(b)*136/100)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, alphabet[mod.Int64()])
	}

	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, alphabetIdx0)
	}

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}

func EncodeFromBigInt(y *big.Int) string {

	x := new(big.Int).Set(y)

	answer := make([]byte, 0)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, big58, mod)
		answer = append(answer, alphabet[mod.Int64()])
	}
	//no need to considering the leading zero bytes

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}
	return string(answer)
}

/*  PrependedBytes: Get the minimum int for base58 specified characters
prefix: base58 symbols as prefix, i.e. "xprv" in BIP32 private key
b256Count: number of bytes exluded prefix and included checksum,  i.e. 78 in BIP32's private key
*/

func PrependedBytes(prefixs string, b256Count int64) []byte {

	numLeadingZeros := int(0)
	needOverFlowCheck := bool(true)
	stripPrefixs := make([]byte, 0)

	//strip the leading zeros
	for _, prefix := range []byte(prefixs) {
		if prefix == (byte)(alphabet[indexZero]) {
			numLeadingZeros++
		} else {
			// break at the first non alphabet[0]
			stripPrefixs = []byte(prefixs)[numLeadingZeros:]
			break
		}
	}

	//create leadingzero string
	s0 := make([]byte, numLeadingZeros)
	for i := 0; i < numLeadingZeros; i++ {
		s0[i] = byte(0)
	}

	if numLeadingZeros == len(prefixs) {
		//	fmt.Printf("res: %v\n", s0)
		return s0
	}

	b58Count := (int64)(math.Ceil(float64(b256Count) * 136 / 100))

	for {
		min := big.NewInt(0)
		for _, prefix := range stripPrefixs {
			min.Mul(min, big58)
			min.Add(min, big.NewInt((int64)(b58[prefix])))
			//	fmt.Printf("prefix:%v, val:%v, min:%v\n", prefix, (int64)(b58[prefix]), min)
		}

		//left-shift(mult. by 58 ** b58Count),
		min.Mul(min, new(big.Int).Exp(big58, big.NewInt(b58Count), nil))

		//right-shift(div. by 256** b256Count),
		big256Exp := new(big.Int).Exp(big256, big.NewInt(b256Count), nil)
		min.Div(min, big256Exp)
		min.Add(min, big.NewInt(1))

		if needOverFlowCheck == true {
			//max = min * 256^k +256^k - 1
			max := new(big.Int)
			max.Mul(min, big256Exp)
			max.Add(max, big256Exp)
			max.Sub(max, big.NewInt(1))
			s := EncodeFromBigInt(max)

			//add a alphabet[indexZero] if overflow
			if !strings.HasPrefix(s, string(stripPrefixs[:])) {
				stripPrefixs = append(stripPrefixs, alphabet[indexZero])
				needOverFlowCheck = false //just check overflow once
				continue
			}
		}

		res := append(s0, min.Bytes()...)
		//	fmt.Printf("res %v\n", res)
		return res
	}
}
