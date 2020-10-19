// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var checkEncodingStringTests = []struct {
	version byte
	in      string
	out     string
}{
	{20, "", "3MNQE1X"},
	{20, " ", "B2Kr6dBE"},
	{20, "-", "B3jv1Aft"},
	{20, "0", "B482yuaX"},
	{20, "1", "B4CmeGAC"},
	{20, "-1", "mM7eUf6kB"},
	{20, "11", "mP7BMTDVH"},
	{20, "abc", "4QiVtDjUdeq"},
	{20, "1234598760", "ZmNb8uQn5zvnUohNCEPP"},
	{20, "abcdefghijklmnopqrstuvwxyz", "K2RYDcKfupxwXdWhSAxQPCeiULntKm63UXyx5MvEH2"},
	{20, "00000000000000000000000000000000000000000000000000000000000000", "bi1EWXwJay2udZVxLJozuTb8Meg4W9c6xnmJaRDjg6pri5MBAxb9XwrpQXbtnqEoRV5U2pixnFfwyXC8tRAVC8XxnjK"},
}

var addressStringTests = []struct {
	in  string
	out string
}{
	{"0656D0a3Ee4cEa51d7Fd281Ccadf593F612B2b73", "HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy"},
	{"00Cb32D3C9c0040E117158AaBBa7ACEE6f7Be307", "HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"},
	{"27b6a0b2dd5aafA8455504d9822A4216487e698c", "HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo"},
}

func TestBase58Check(t *testing.T) {
	for x, test := range checkEncodingStringTests {
		// test encoding
		res2 := CheckEncode([]byte(test.in), test.version)
		if res2 != test.out {
			t.Errorf("CheckEncode test #%d failed: got %s, want: %s", x, res2, test.out)
		}
		//fmt.Println(res2)

		// test decoding
		res, version, err := CheckDecode(test.out)
		if err != nil {
			t.Errorf("CheckDecode test #%d failed with err: %v", x, err)
		} else if version != test.version {
			t.Errorf("CheckDecode test #%d failed: got version: %d want: %d", x, version, test.version)
		} else if string(res) != test.in {
			t.Errorf("CheckDecode test #%d failed: got: %s want: %s", x, res, test.in)
		}
		//fmt.Println(string(res))
	}

	// test the two decoding failure cases
	// case 1: checksum error
	_, _, err := CheckDecode("3MNQE1Y")
	if err != ErrChecksum {
		t.Error("Checkdecode test failed, expected ErrChecksum")
	}
	// case 2: invalid formats (string lengths below 5 mean the version byte and/or the checksum
	// bytes are missing).
	testString := ""
	for len := 0; len < 4; len++ {
		// make a string of length `len`
		_, _, err = CheckDecode(testString)
		if err != ErrInvalidFormat {
			t.Error("Checkdecode test failed, expected ErrInvalidFormat")
		}
	}

}

func TestEthAddrToHBTAddr(t *testing.T) {
	for x, test := range addressStringTests {
		// test encoding
		ethaddbyte, err := hex.DecodeString(test.in)
		if err != nil {
			t.Error("hex address decode fail")
		}
		res2 := EthAddrToHBCAddr(ethaddbyte)
		if res2 != test.out {
			t.Errorf("EthAddrToHBCAddr test #%d failed: got %s, want: %s", x, res2, test.out)
		}
		//fmt.Printf("length:[%d],%s\n", len(res2), res2)

		// test decoding
		res, err := HBCAddrToEthAddr(test.out)
		if err != nil {
			t.Errorf("HBCAddrToEthAddr test #%d failed with err: %v", x, err)
		} else if bytes.Equal(res, []byte(test.in)) {
			t.Errorf("HBCAddrToEthAddr test #%d failed: got: %s want: %s", x, res, test.in)
		}
		//	fmt.Println(hex.EncodeToString(res))
	}

	// test the two decoding failure cases
	// case 1: checksum error
	_, err := HBCAddrToEthAddr("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2n")
	if err != ErrChecksum {
		t.Error("HBCAddrToEthAddr test failed, expected ErrChecksum")
	}
	// case 2: invalid formats (string lengths below 6 mean the version byte and/or the checksum
	// bytes are missing).
	testString := "Invala"
	_, err = HBCAddrToEthAddr(testString)
	if err != ErrInvalidFormat {
		t.Error("HBCAddrToEthAddr test failed, expected ErrInvalidFormat")
	}
}
