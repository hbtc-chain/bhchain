package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

var from, _ = sdk.CUAddressFromBase58("HBCTGvSBNT8oEA4xEqCZGERWkzvpNXCTDGMx")
var to, _ = sdk.CUAddressFromBase58("HBCiopN1Vw38QyjEnfJ7nKeVgMSi9sFjQfkg")
var btc, eth = sdk.Symbol("btc"), sdk.Symbol("eth")
var pubKey = []byte("testpubkey1")

var keygenTC = []struct {
	OrderID  string
	Symbol   sdk.Symbol
	From     sdk.CUAddress
	To       sdk.CUAddress
	Expected bool
}{
	{Symbol: "", Expected: false},
	{OrderID: "", Symbol: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: sdk.NativeToken, From: nil, To: to, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: sdk.NativeToken, From: []byte("1234567890"), To: to, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: sdk.NativeToken, From: from, To: nil, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: sdk.NativeToken, From: from, To: []byte("1234567890"), Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: sdk.NativeToken, From: from, To: to, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: from, To: to, Expected: true},
	{OrderID: uuid.NewV4().String(), Symbol: "bheos", From: from, To: to, Expected: true},
}

var privKeyBytes = [32]byte{'1', '2', '3', '4'}
var prikey = secp256k1.PrivKeySecp256k1(privKeyBytes)
var pubkey = prikey.PubKey()
var cu1 = sdk.NewCUAddress()
var cu2 = sdk.NewCUAddress()
var cu3 = sdk.NewCUAddress()
var cu4 = sdk.NewCUAddress()

var sign1, _ = prikey.Sign([]byte("sign msgs1"))
var sign2, _ = prikey.Sign([]byte("sign msgs2"))
var sign3, _ = prikey.Sign([]byte("sign msgs3"))
var sign4, _ = prikey.Sign([]byte("sign msgs4"))
var sig1 = cutypes.StdSignature{prikey.PubKey(), sign1}
var sig2 = cutypes.StdSignature{prikey.PubKey(), sign2}
var sig3 = cutypes.StdSignature{prikey.PubKey(), sign3}
var sig4 = cutypes.StdSignature{prikey.PubKey(), sign4}

func TestTokenKeyGen(t *testing.T) {
	for _, tc := range keygenTC {
		msg := NewMsgKeyGen(tc.OrderID, tc.Symbol, tc.From, tc.To)
		if tc.Expected {
			assert.Nil(t, msg.ValidateBasic())
		} else {
			assert.NotNil(t, msg.ValidateBasic())
		}
	}
}

var keygenFinishTC = []struct {
	From     sdk.CUAddress
	OrderID  string
	Symbol   sdk.Symbol
	PubKey   []byte
	Address  string
	KeyNodes []sdk.CUAddress
	KeySigs  []cutypes.StdSignature

	Expected bool
}{
	{Symbol: "", Expected: false},
	{OrderID: "", Symbol: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: sdk.NativeToken, From: nil, Address: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: []byte("1234567890"), Address: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "eth", From: from, Address: "", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: from, Address: "1234567890", Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: from, Address: "1234567890", PubKey: pubKey, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "tusdt", From: from, Address: "1234567890", KeyNodes: []sdk.CUAddress{}, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: from, Address: "1234567890", PubKey: pubKey, KeyNodes: []sdk.CUAddress{cu1, cu2, cu3, cu4}, Expected: false},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: from, Address: "1234567890", PubKey: pubKey, KeyNodes: []sdk.CUAddress{cu1, cu2, cu3, cu4}, KeySigs: []cutypes.StdSignature{}, Expected: true},
	{OrderID: uuid.NewV4().String(), Symbol: "btc", From: from, Address: "1234567890", PubKey: pubKey, KeyNodes: []sdk.CUAddress{cu1, cu2, cu3, cu4}, KeySigs: []cutypes.StdSignature{sig1, sig2, sig3, sig4}, Expected: true},
}

func TestMsgKeyGenEncode(t *testing.T) {
	cu1 := sdk.NewCUAddress()
	cu2 := sdk.NewCUAddress()
	msg := NewMsgKeyGen(uuid.NewV4().String(), "eth", cu1, cu2)

	err := msg.ValidateBasic()
	assert.Nil(t, err)
	bz := msg.GetSignBytes()
	assert.NotNil(t, bz)
	var p MsgKeyGen
	ModuleCdc.UnmarshalJSON(bz, &p)
	assert.Equal(t, msg, p)
}
