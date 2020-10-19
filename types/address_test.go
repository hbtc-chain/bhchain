package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hbtc-chain/bhchain/base58"
	"github.com/hbtc-chain/bhchain/codec"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var invalidStrs = []string{
	"hello, world!",
	"0xAA",
	"AAA",
	Bech32PrefixAccAddr + "AB0C",
	Bech32PrefixAccPub + "1234",
	Bech32PrefixValAddr + "5678",
	Bech32PrefixValPub + "BBAB",
	Bech32PrefixConsAddr + "FF04",
	Bech32PrefixConsPub + "6789",
}

func testMarshal(t *testing.T, original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	require.Nil(t, err)
	err = unmarshal(bz)
	require.Nil(t, err)
	require.Equal(t, original, res)
}

var addrStringTests = []struct {
	in  string
	out string
}{
	{"0656D0a3Ee4cEa51d7Fd281Ccadf593F612B2b73", "HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy"},
	{"00Cb32D3C9c0040E117158AaBBa7ACEE6f7Be307", "HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"},
	{"27b6a0b2dd5aafA8455504d9822A4216487e698c", "HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo"},
}

// guard the dependent lib base58
func TestBaseCU_GetAddress(t *testing.T) {
	for x, test := range addrStringTests {
		// test encoding
		ethaddbyte, err := hex.DecodeString(test.in)
		if err != nil {
			t.Error("hex address decode fail")
		}
		res2 := base58.EthAddrToHBCAddr(ethaddbyte)
		if res2 != test.out {
			t.Errorf("EthAddrToHBCAddr test #%d failed: got %s, want: %s", x, res2, test.out)
		}

		// test decoding
		res, err := base58.HBCAddrToEthAddr(test.out)
		if err != nil {
			t.Errorf("HBCAddrToEthAddr test #%d failed with err: %v", x, err)
		} else if bytes.Equal(res, []byte(test.in)) {
			t.Errorf("HBCAddrToEthAddr test #%d failed: got: %s want: %s", x, res, test.in)
		}
	}
	// test the two decoding failure cases
	// case 1: checksum error
	_, err := base58.HBCAddrToEthAddr("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2n")
	if err != base58.ErrChecksum {
		t.Error("HBCAddrToEthAddr test failed, expected ErrChecksum")
	}
	// case 2: invalid formats (string lengths below 6 mean the version byte and/or the checksum
	// bytes are missing).
	testString := "Invalid"
	_, err = base58.HBCAddrToEthAddr(testString)
	if err != base58.ErrInvalidFormat {
		t.Error("HBCAddrToEthAddr test failed, expected ErrInvalidFormat")
	}
}

func TestEmptyAddresses(t *testing.T) {
	require.Equal(t, (CUAddress{}).String(), "")
	require.Equal(t, (ValAddress{}).String(), "")
	require.Equal(t, (ConsAddress{}).String(), "")

	accAddr, err := CUAddressFromBase58("")
	require.True(t, accAddr.Empty())
	require.Nil(t, err)

	valAddr, err := ValAddressFromBech32("")
	require.True(t, valAddr.Empty())
	require.Nil(t, err)

	consAddr, err := ConsAddressFromBech32("")
	require.True(t, consAddr.Empty())
	require.Nil(t, err)
}

func TestRandBech32PubkeyConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		mustBech32AccPub := MustBech32ifyAccPub(pub)
		bech32AccPub, err := Bech32ifyAccPub(pub)
		require.Nil(t, err)
		require.Equal(t, bech32AccPub, mustBech32AccPub)

		mustBech32ValPub := MustBech32ifyValPub(pub)
		bech32ValPub, err := Bech32ifyValPub(pub)
		require.Nil(t, err)
		require.Equal(t, bech32ValPub, mustBech32ValPub)

		mustBech32ConsPub := MustBech32ifyConsPub(pub)
		bech32ConsPub, err := Bech32ifyConsPub(pub)
		require.Nil(t, err)
		require.Equal(t, bech32ConsPub, mustBech32ConsPub)

		mustAccPub := MustGetAccPubKeyBech32(bech32AccPub)
		accPub, err := GetAccPubKeyBech32(bech32AccPub)
		require.Nil(t, err)
		require.Equal(t, accPub, mustAccPub)

		mustValPub := MustGetValPubKeyBech32(bech32ValPub)
		valPub, err := GetValPubKeyBech32(bech32ValPub)
		require.Nil(t, err)
		require.Equal(t, valPub, mustValPub)

		mustConsPub := MustGetConsPubKeyBech32(bech32ConsPub)
		consPub, err := GetConsPubKeyBech32(bech32ConsPub)
		require.Nil(t, err)
		require.Equal(t, consPub, mustConsPub)

		require.Equal(t, valPub, accPub)
		require.Equal(t, valPub, consPub)
	}
}

func TestYAMLMarshalers(t *testing.T) {
	addr := secp256k1.GenPrivKey().PubKey().Address()

	acc := CUAddress(addr)
	val := ValAddress(addr)
	cons := ConsAddress(addr)

	got, _ := yaml.Marshal(&acc)
	require.Equal(t, acc.String()+"\n", string(got))

	got, _ = yaml.Marshal(&val)
	require.Equal(t, val.String()+"\n", string(got))

	got, _ = yaml.Marshal(&cons)
	require.Equal(t, cons.String()+"\n", string(got))
}

func TestRandBech32AccAddrConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		acc := CUAddress(pub.Address())
		res := CUAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := CUAddressFromBase58(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = CUAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := CUAddressFromHex(str)
		require.NotNil(t, err)

		_, err = CUAddressFromBase58(str)
		require.NotNil(t, err)

		err = (*CUAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}

func TestValAddr(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := ValAddress(pub.Address())
		res := ValAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := ValAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = ValAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := ValAddressFromHex(str)
		require.NotNil(t, err)

		_, err = ValAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*ValAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}

func TestConsAddress(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := ConsAddress(pub.Address())
		res := ConsAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := ConsAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = ConsAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := ConsAddressFromHex(str)
		require.NotNil(t, err)

		_, err = ConsAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*ConsAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestAddressInterface(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	rand.Read(pub[:])

	addrs := []Address{
		ConsAddress(pub.Address()),
		ValAddress(pub.Address()),
		CUAddress(pub.Address()),
	}

	for _, addr := range addrs {
		switch addr := addr.(type) {
		case CUAddress:
			res, err := CUAddressFromBase58(addr.String())
			require.Equal(t, res.Bytes(), addr.Bytes())
			require.Nil(t, err)
		case ValAddress:
			res, err := ValAddressFromBech32(addr.String())
			require.Equal(t, res.Bytes(), addr.Bytes())
			require.Nil(t, err)
		case ConsAddress:
			res, err := ConsAddressFromBech32(addr.String())
			require.Equal(t, res.Bytes(), addr.Bytes())
			require.Nil(t, err)
		default:
			t.Fail()
		}
	}

}

func TestCustomAddressVerifier(t *testing.T) {
	// Create a 10 byte address
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	accBech := CUAddress(addr).String()
	valBech := ValAddress(addr).String()
	consBech := ConsAddress(addr).String()
	// Verifiy that the default logic rejects this 10 byte address
	err := VerifyAddressFormat(addr)
	require.NotNil(t, err)
	//_, err = CUAddressFromBase58(accBech)
	//require.NotNil(t, err)
	_, err = ValAddressFromBech32(valBech)
	require.NotNil(t, err)
	_, err = ConsAddressFromBech32(consBech)
	require.NotNil(t, err)

	// Set a custom address verifier that accepts 10 or 20 byte addresses
	GetConfig().SetAddressVerifier(func(bz []byte) error {
		n := len(bz)
		if n == 10 || n == AddrLen {
			return nil
		}
		return fmt.Errorf("incorrect address length %d", n)
	})

	// Verifiy that the custom logic accepts this 10 byte address
	err = VerifyAddressFormat(addr)
	require.Nil(t, err)
	_, err = CUAddressFromBase58(accBech)
	require.Nil(t, err)
	_, err = ValAddressFromBech32(valBech)
	require.Nil(t, err)
	_, err = ConsAddressFromBech32(consBech)
	require.Nil(t, err)
}

//---------- CUAddress cases -------
var IsValidAddrTests = []struct {
	in  bool
	out string
}{
	{true, "HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy"},
	{true, "HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"},
	{true, "HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo"},
	{false, "BHPWuMc6MDhA7Rgt8ejeSXWCG6Sn4BuzXWbw"},
	{false, "PWuMc6MDhA7Rgt8ejeSXWCG6Sn4BzXWbw"},
	{false, "BHPWuMc6MDhA7Rgt8ejeSXWCG6Sn4BzXWb"},
	{false, ""},
	{false, "BH"},
	{false, "B"},
	{false, "1"},
}

var addressStringTests = []struct {
	in  string
	out string
}{
	{"0656D0a3Ee4cEa51d7Fd281Ccadf593F612B2b73", "HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy"},
	{"00Cb32D3C9c0040E117158AaBBa7ACEE6f7Be307", "HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q"},
	{"27b6a0b2dd5aafA8455504d9822A4216487e698c", "HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo"},
}

func TestIsValidAddr(t *testing.T) {
	for _, test := range IsValidAddrTests {
		//fmt.Println(len(test.out))
		in := IsValidAddr(test.out)
		assert.Equal(t, in, test.in, "should equal")
	}
}

func TestCUAddressFromByte(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	cuAddr := CUAddressFromByte(b)
	assert.False(t, cuAddr.Empty())

	b = []byte{1, 2, 3}
	cuAddr = CUAddressFromByte(b)
	assert.True(t, cuAddr.Empty())

	b = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4}
	cuAddr = CUAddressFromByte(b)
	assert.True(t, cuAddr.Empty())

	b = []byte{}
	cuAddr = CUAddressFromByte(b)
	assert.True(t, cuAddr.Empty())

}

type CUAddrAndString struct {
	hexStr string
	b58Str string
}

var testData = []CUAddrAndString{
	{"B5AD24DD9E5D60E1F0734AF2D819FF9A198A2A38", "HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe"},
	{"549C15831315AD56F89C0EDDF9D852B6CB7605E3", "HBCTuGJjbfN7aA13QdhvZR45bQb8UXBMxFNY"},
	{"7FDE0B354783BC8607222CD994C805E303A80BCD", "HBCXqzNwvZGR88XvkccrXEz6c33Za2Da3bBo"},
	{"C48C652CD2A293D7F57BB36FD3B30AD87C857923", "HBCe79C9Cf1ixV82SfAG78WBTHNGmGsN42uy"},
	{"FEBF0CA4CB4897C9A27A54275E612FEF275752AE", "HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463"},
	{"3BDA7843C6CE02FB1B274DF18F58E04354750EB8", "HBCReN9aTJTErpnT11ZfJxAK6oXPnBgP7d4b"},
	{"CCE353A7008DD9E838691E5921D935848A0410F8", "HBCesEjnbc7wu2m6dTL8ekd45VrVAQzqYD7J"},
	{"4DBC8579C8A7453E7547A496AB07FB48F435B1F0", "HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx"},
	{"98488C3C1BDF59D448A52EC97A9410F164834CF3", "HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE"},
	{"2D44FDEFC054FD718550B60210C07982012C8D11", "HBCQKFczL6oYG39rAnvPaobge1xBA2qQPsxX"},
	{"935FCC04364A3C68CFD4E014CC21BBC1A29845EB", "HBCZd8ezxCzPgkyV8u1AkBizJvAnPRc7sSvE"},
	{"03A62A1613A098DF0C607CEBFEEC7475118E02E8", "HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"},
	{"F81E3C014E639ACA87BB7A7F8A38C24F355DB3DB", "HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg"},
}

func TestCUAddressFromBase58AndFromHex(t *testing.T) {
	for _, d := range testData {
		cuAddr1, _ := CUAddressFromHex(d.hexStr)
		//t.Logf("addr:%s", cuAddr1)
		cuAddr2, _ := CUAddressFromBase58(d.b58Str)
		assert.True(t, cuAddr1.Equals(cuAddr2))
		//	t.Log(cuAddr1.Bytes())
	}
}

func TestCUAddressFromBase58Error(t *testing.T) {
	b58Data := []struct {
		b58Str string
		res    bool
	}{
		{"HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe", true},
		{"     ", true}, // blank input should get CUAddress{} without error
		{"", true},
		{"!", false},
		{"$C!", false},
		{"BH $ C ! ", false},
		{"HbcckWHh1gtoiWXtyALegeudFPhSwnrwoYhe", false},
		{"HbCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe", false},
		{"hBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe", false},
		{"hbCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe", false},
		{"hbcckWHh1gtoiWXtyALegeudFPhSwnrwoYhe", false},
		{"bhcWFpii3MAUrvnw8NFB1SrGKVz7UhCxEpSF", false},
	}

	for _, d := range b58Data {
		_, err := CUAddressFromBase58(d.b58Str)
		assert.Equal(t, d.res, err == nil)
	}
}

func TestCUAddressFromHexError(t *testing.T) {
	hexData := []struct {
		hexStr string
		res    bool
	}{
		{"B5AD24DD9E5D60E1F0734AF2D819FF9A198A2A38", true},
		{" ", false},
		{"B5AD24DD9E5D60E1F0734AF2D819FF9A198A2A3#", false},
		{"%%%###@@@@@", false},
	}
	for _, d := range hexData {
		_, err := CUAddressFromHex(d.hexStr)
		assert.Equal(t, d.res, err == nil)
	}

}

func TestSortCUAddress(t *testing.T) {
	addrs := []string{
		("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe"),
		("HBCTuGJjbfN7aA13QdhvZR45bQb8UXBMxFNY"),
		("HBCXqzNwvZGR88XvkccrXEz6c33Za2Da3bBo"),
		("HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463"),
		("HBCReN9aTJTErpnT11ZfJxAK6oXPnBgP7d4b"),
		("HBCesEjnbc7wu2m6dTL8ekd45VrVAQzqYD7J"),
		("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE"),
		("HBCQKFczL6oYG39rAnvPaobge1xBA2qQPsxX"),
		("HBCZd8ezxCzPgkyV8u1AkBizJvAnPRc7sSvE"),
		("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"),
		("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg"),
	}

	expected := []string{
		("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP"),
		("HBCQKFczL6oYG39rAnvPaobge1xBA2qQPsxX"),
		("HBCReN9aTJTErpnT11ZfJxAK6oXPnBgP7d4b"),
		("HBCTuGJjbfN7aA13QdhvZR45bQb8UXBMxFNY"),
		("HBCXqzNwvZGR88XvkccrXEz6c33Za2Da3bBo"),
		("HBCZd8ezxCzPgkyV8u1AkBizJvAnPRc7sSvE"),
		("HBCa56BabeA4hNamfgY2xPp914kBe4rMPctE"),
		("HBCckWHh1gtoiWXtyALegeudFPhSwnrwoYhe"),
		("HBCesEjnbc7wu2m6dTL8ekd45VrVAQzqYD7J"),
		("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg"),
		("HBCjQs4nKHHfu5LrznRm3vBbbaeW1ghVw463"),
	}

	cuAddrs := make([]CUAddress, 0, len(addrs))

	for _, addr := range addrs {
		cuAddr, err := CUAddressFromBase58(addr)
		assert.Nil(t, err)
		cuAddrs = append(cuAddrs, cuAddr)
	}

	sort.Sort(CUAddressList(cuAddrs))

	for i, addr := range cuAddrs {
		assert.Equal(t, expected[i], addr.String())
		//t.Log(i, addr.String())
	}

}

func TestCUAddressAccAddressConsAddressValAddressEqual(t *testing.T) {

	for i := 0; i < 100; i++ {
		privKey := ed25519.GenPrivKey()
		pubKey := privKey.PubKey()

		cuAddr := CUAddress(pubKey.Address())
		valAddr := ValAddress(pubKey.Address())
		consAddr := ConsAddress(pubKey.Address())

		assert.True(t, cuAddr.Equals(valAddr))
		assert.True(t, cuAddr.Equals(consAddr))
		assert.True(t, valAddr.Equals(consAddr))

		msg := crypto.CRandBytes(128)
		signature, _ := privKey.Sign(msg)
		assert.True(t, pubKey.VerifyBytes(msg, signature))

		//t.Logf("cuAddr Byte:%v, Str:%v", cuAddr, cuAddr.String())
	}
}

func testMarshalCUAddr(t *testing.T, original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	assert.Nil(t, err)
	err = unmarshal(bz)
	assert.Nil(t, err)
	assert.Equal(t, original, res)
}

func TestCUAddressMarshal(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		cuAddr := CUAddress(pub.Address())
		res := CUAddress{}

		testMarshalCUAddr(t, &cuAddr, &res, cuAddr.MarshalJSON, (&res).UnmarshalJSON)
		testMarshalCUAddr(t, &cuAddr, &res, cuAddr.Marshal, (&res).Unmarshal)

		str := cuAddr.String()
		res, err := CUAddressFromBase58(str)
		assert.Nil(t, err)
		assert.Equal(t, cuAddr, res)

		hexStr := hex.EncodeToString(cuAddr)
		res, err = CUAddressFromHex(hexStr)
		assert.Nil(t, err)
		assert.Equal(t, cuAddr, res)

	}
}

func TestCUAddressFromPubKey(t *testing.T) {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	pubKeyBytes, _ := hex.DecodeString("1624de64207009997bfedafb02f6648d653bcb2488961b65b8707dafe664fcf80e72512cee")
	var pubKey crypto.PubKey
	cdc.MustUnmarshalBinaryBare(pubKeyBytes, &pubKey)

	assert.Equal(t, "HBCYQVQifcq9xVvp31swV3JtztCGqPwgzenk",
		CUAddressFromPubKey(pubKey).String())
}

func TestCUAddress_FromHexEquals(t *testing.T) {
	hexAddrStr := "B5AD24DD9E5D60E1F0734AF2D819FF9A198A2A38"
	CUAddr, err := CUAddressFromHex(hexAddrStr)
	assert.Nil(t, err)
	ValAddr, err := ValAddressFromHex(hexAddrStr)
	assert.Nil(t, err)
	assert.True(t, CUAddr.Equals(ValAddr))
	ConAddr, err := ConsAddressFromHex(hexAddrStr)
	assert.Nil(t, err)
	assert.True(t, CUAddr.Equals(ConAddr))
}

func TestCUAddressToValiadatorAddress(t *testing.T) {
	for i := 0; i < 10; i++ {
		pubKey := secp256k1.GenPrivKey().PubKey()
		cuAddr := CUAddressFromPubKey(pubKey)

		valOperAddrStr := ValAddress(pubKey.Address().Bytes()).String()
		valOperAddr, err := ValAddressFromBech32(valOperAddrStr)
		assert.Nil(t, err)
		assert.Equal(t, cuAddr, CosmosAddressToCUAddress(valOperAddr))

		valConsAddrStr := ConsAddress(pubKey.Address().Bytes()).String()
		valConsAddr, err := ConsAddressFromBech32(valConsAddrStr)
		assert.Nil(t, err)
		assert.Equal(t, cuAddr, CosmosAddressToCUAddress(valConsAddr))
	}
}
