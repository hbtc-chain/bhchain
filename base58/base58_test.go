// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"
)

var stringTests = []struct {
	in  string
	out string
}{
	{"", ""},
	{" ", "Z"},
	{"-", "n"},
	{"0", "q"},
	{"1", "r"},
	{"-1", "4SU"},
	{"11", "4k8"},
	{"abc", "ZiCa"},
	{"1234598760", "3mJr7AoUXx2Wqd"},
	{"abcdefghijklmnopqrstuvwxyz", "3yxU3u1igY8WkgtjK92fbJQCd4BZiiT1v25f"},
	{"00000000000000000000000000000000000000000000000000000000000000", "3sN2THZeE9Eh9eYrwkvZqNstbHGvrxSAM7gXUXvyFQP8XvQLUqNCS27icwUeDT7ckHm4FUHM2mTVh1vbLmk7y"},
}

var invalidStringTests = []struct {
	in  string
	out string
}{
	{"0", ""},
	{"O", ""},
	{"I", ""},
	{"l", ""},
	{"3mJr0", ""},
	{"O3yxU", ""},
	{"3sNI", ""},
	{"4kl8", ""},
	{"0OIl", ""},
	{"!@#$%^&*()-_=+~`", ""},
}

var hexTests = []struct {
	in  string
	out string
}{
	{"61", "2g"},
	{"626262", "a3gV"},
	{"636363", "aPEr"},
	{"73696d706c792061206c6f6e6720737472696e67", "2cFupjhnEsSn59qHXstmK2ffpLv2"},
	{"00eb15231dfceb60925886b67d065299925915aeb172c06647", "1NS17iag9jJgTHD1VXjvLCEnZuQ3rJDE9L"},
	{"516b6fcd0f", "ABnLTmg"},
	{"bf4f89001e670274dd", "3SEo3LWLoPntC"},
	{"572e4794", "3EFU7m"},
	{"ecac89cad93923c02321", "EJDM8drfXA6uyA"},
	{"10c8511e", "Rt5zm"},
	{"00000000000000000000", "1111111111"},
}

func TestBase58(t *testing.T) {
	// Encode tests
	for x, test := range stringTests {
		tmp := []byte(test.in)
		if res := Encode(tmp); res != test.out {
			t.Errorf("Encode test #%d failed: got: %s want: %s",
				x, res, test.out)
			continue
		}
	}

	// Decode tests
	for x, test := range hexTests {
		b, err := hex.DecodeString(test.in)
		if err != nil {
			t.Errorf("hex.DecodeString failed failed #%d: got: %s", x, test.in)
			continue
		}
		s := Encode(b)

		//fmt.Printf("in:%v b:%v s:%v\n", test.in, b, s)

		if test.out != s {
			t.Fatal(1)
		}

		if res := Decode(test.out); !bytes.Equal(res, b) {
			t.Errorf("Decode test #%d failed: got: %q want: %q",
				x, res, test.in)
			continue
		}
	}

	//Decode with invalid input
	for x, test := range invalidStringTests {
		if res := Decode(test.in); string(res) != test.out {
			t.Errorf("Decode invalidString test #%d failed: got: %q want: %q",
				x, res, test.out)
			continue
		}
	}
}

func TestPrependedBytes(t *testing.T) {
	//var res []byte

	cases := []struct {
		prefixs   string
		b256count int64
		res       string
	}{
		{"xprv", 78, "0488ADE3"},
		{"xpub", 78, "0488B21E"},
		{"tprv", 78, "04358394"},
		{"tpub", 78, "043587CE"},
		{"Bhex", 24, "4F2CE9"},
		{"Bh", 24, "0605"},
		{"111111", 78, "000000000000"},
		{"11xprv", 78, "00000488ADE3"},
		{"HBT", 24, "021067"},
		{"BH", 24, "05ca"},
		{"HBC", 24, "021042"},
	}

	for i, c := range cases {
		res := PrependedBytes(c.prefixs, c.b256count)
		//fmt.Printf("res:%x\n", res)
		expected, err := hex.DecodeString(c.res)
		if err != nil {
			t.Fatal("not a hex string")
		}

		if !bytes.Equal(expected, res) {
			t.Logf("i:%x, expected %v, got %v\n", i, expected, res)
			t.Fatal(1)
		}
	}
}

func TestMinMaxBase58String(t *testing.T) {
	cases := []struct {
		prefixs   string
		b256count int64
		res       string
	}{
		{"xprv", 78, "0488ADE3"},
		{"xpub", 78, "0488B21E"},
		{"tprv", 78, "04358394"},
		{"tpub", 78, "043587CE"},
	}

	for _, c := range cases {
		v, err := hex.DecodeString(c.res)
		if err != nil {
			t.Fatal("not a hex string")
		}

		// origV
		big256Exp := new(big.Int).Exp(big.NewInt(256), big.NewInt(c.b256count), nil)
		origV := new(big.Int).SetBytes(v)
		val := new(big.Int).Mul(origV, big256Exp)

		minStr := EncodeFromBigInt(val)
		//	fmt.Printf("min orig string:%v\n", minStr)
		if !strings.HasPrefix(minStr, c.prefixs) {
			t.Errorf("expected %v, Got %v\n", c.prefixs, minStr)
		}

		val.Add(val, big256Exp)
		val.Sub(val, big.NewInt(1))

		maxStr := EncodeFromBigInt(val)
		//	fmt.Printf("max orig string:%v\n", maxStr)
		if !strings.HasPrefix(maxStr, c.prefixs) {
			t.Errorf("expected %v, Got %v\n", c.prefixs, maxStr)
		}

	}
}
