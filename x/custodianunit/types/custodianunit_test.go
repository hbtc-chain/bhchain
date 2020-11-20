package types

import (
	"encoding/json"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

const EthToken = "eth"

func TestBaseAddressPubKey(t *testing.T) {
	_, pub1, addr1 := KeyTestPubAddr()
	_, pub2, addr2 := KeyTestPubAddr()
	cu := NewBaseCUWithPubkey(pub1, sdk.CUTypeUser)
	// check the address (set) and pubkey (set) and cutype (set)
	assert.EqualValues(t, pub1, cu.GetPubKey())
	assert.EqualValues(t, addr1, cu.GetAddress())
	assert.EqualValues(t, sdk.CUTypeUser, cu.GetCUType())

	cu = NewBaseCUWithAddress(addr1, sdk.CUTypeUser)

	// check the address (set) and pubkey (not set)
	assert.EqualValues(t, addr1, cu.GetAddress())
	assert.EqualValues(t, nil, cu.GetPubKey())
	assert.EqualValues(t, sdk.CUTypeUser, cu.GetCUType())

	// can't override cutype
	err := cu.SetCUType(sdk.CUTypeOp)
	assert.NotNil(t, err)
	assert.EqualValues(t, sdk.CUTypeUser, cu.GetCUType())
	// empty BaseCU can set cutype
	cuEmpty := BaseCU{}
	err = cuEmpty.SetCUType(sdk.CUTypeOp)
	assert.Nil(t, err)
	assert.EqualValues(t, sdk.CUTypeOp, cuEmpty.GetCUType())

	// can't override address
	err = cu.SetAddress(addr2)
	assert.NotNil(t, err)
	assert.EqualValues(t, addr1, cu.GetAddress())
	// can set address on empty CU
	cuEmpty = BaseCU{}
	err = cuEmpty.SetAddress(addr2)
	assert.Nil(t, err)
	assert.EqualValues(t, addr2, cuEmpty.GetAddress())

	// set the pubkey
	err = cu.SetPubKey(pub1)
	assert.Nil(t, err)
	assert.Equal(t, pub1, cu.GetPubKey())

	// cosmos can override pubkey // ???
	// bhchain can not override pubkey
	err = cu.SetPubKey(pub2)
	assert.NotNil(t, err)
	//assert.Equal(t, pub2, cu.GetPubKey())
	// can set pubkey on empty CU
	cuEmpty = BaseCU{}
	err = cuEmpty.SetPubKey(pub1)
	assert.Nil(t, err)

	assert.EqualValues(t, pub1, cuEmpty.GetPubKey())

	// GetSymbol from a CUTypeUser should get ""
	s := cu.GetSymbol()
	assert.Equal(t, s, "")
	//  GetSymbol from a CUTypeOp should get the symbol
}

func TestBaseCUSequence(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)

	seq := uint64(7)

	err := cu.SetSequence(seq)
	assert.Nil(t, err)
	assert.Equal(t, seq, cu.GetSequence())
}

func TestBaseCUMarshal(t *testing.T) {
	_, pub, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)

	seq := uint64(7)
	// set everything on the CU
	err := cu.SetPubKey(pub)
	assert.Nil(t, err)
	err = cu.SetSequence(seq)
	assert.Nil(t, err)
	// need a codec for marshaling
	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	b, err := cdc.MarshalBinaryLengthPrefixed(cu)
	assert.Nil(t, err)
	var cuGot BaseCU
	err = cdc.UnmarshalBinaryLengthPrefixed(b, &cuGot)
	assert.Nil(t, err)
	// the low case field balanceFlows can't be Marshaled by amino
	assert.EqualValues(t, cu, cuGot)

	// error on bad bytes
	cuGot = BaseCU{}
	err = cdc.UnmarshalBinaryLengthPrefixed(b[:len(b)/2], &cuGot)
	assert.NotNil(t, err)
}

func TestCUMarshal(t *testing.T) {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	cu := NewBaseCUWithPubkey(pubKey, sdk.CUTypeOp)

	bz, err := json.Marshal(cu)
	assert.Nil(t, err)

	var dec BaseCU
	err = json.Unmarshal(bz, &dec)
	assert.NotNil(t, err)

	bz, err = cdc.MarshalJSON(cu)
	assert.Nil(t, err)

	var dec1 BaseCU
	err = cdc.UnmarshalJSON(bz, &dec1)
	assert.Nil(t, err)

}

