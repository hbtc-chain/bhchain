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
	cuOp := NewBaseCUWithAddress(addr1, sdk.CUTypeOp)
	cuOp.AddAsset("eth", "addressa", 1)
	s = cuOp.GetSymbol()
	assert.Equal(t, "eth", s)
}

func TestBaseCUCoins(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)

	someCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.NativeToken, 200), sdk.NewInt64Coin("eth", 300))

	// SetCoins
	err := cu.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.Equal(t, someCoins, cu.GetCoins())
	// SetCoinsHold
	err = cu.SetCoinsHold(someCoins)
	assert.Nil(t, err)
	assert.Equal(t, someCoins, cu.GetCoinsHold())

	// SubCoins
	otherCoins := sdk.NewCoins(sdk.NewInt64Coin("eth", 100))
	coinsGot := cu.SubCoins(otherCoins)
	assert.EqualValues(t, someCoins.Sub(otherCoins), coinsGot, cu.Coins)
	// SubCoins
	otherCoins = sdk.NewCoins(sdk.NewInt64Coin("eth", 100))
	coinsGot = cu.SubCoinsHold(otherCoins)
	assert.EqualValues(t, someCoins.Sub(otherCoins), coinsGot, cu.CoinsHold)

	// AddCoins
	coinsGot = cu.AddCoins(otherCoins)
	assert.EqualValues(t, someCoins, coinsGot, cu.Coins)
	// AddCoinsHold
	coinsGot = cu.AddCoinsHold(otherCoins)
	assert.EqualValues(t, someCoins, coinsGot, cu.CoinsHold)

	// should panic
	otherCoins = sdk.NewCoins(sdk.NewInt64Coin("eth", 400))
	assert.Panics(t, func() {
		coinsGot = cu.SubCoins(otherCoins)
	})
	otherCoins = sdk.NewCoins(sdk.NewInt64Coin("btc", 1))
	assert.Panics(t, func() {
		coinsGot = cu.SubCoinsHold(otherCoins)
	})

	// AddAssetCoins
	cu.AddAsset(EthToken, "", 0)
	oneEth := sdk.NewCoins(sdk.NewInt64Coin(EthToken, 1))
	cu.AddAssetCoins(oneEth)
	cu.AddAssetCoinsHold(oneEth)
	assert.EqualValues(t, sdk.NewInt(1), cu.GetAssetCoins().AmountOf(EthToken))
	assert.EqualValues(t, sdk.NewInt(1), cu.GetAssetCoins().AmountOf(EthToken))
	cu.SubAssetCoins(oneEth)
	cu.SubAssetCoinsHold(oneEth)
	assert.EqualValues(t, sdk.ZeroInt(), cu.GetAssetCoinsHold().AmountOf(EthToken))
	assert.EqualValues(t, sdk.ZeroInt(), cu.GetAssetCoins().AmountOf(EthToken))
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

	someCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.NativeToken, 123), sdk.NewInt64Coin("eth", 246))
	seq := uint64(7)

	// set everything on the CU
	err := cu.SetPubKey(pub)
	assert.Nil(t, err)
	err = cu.SetSequence(seq)
	assert.Nil(t, err)
	err = cu.SetCoins(someCoins)
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
	cuGot.balanceFlows = cu.balanceFlows
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

func TestBaseCU_SetEnabaleSendTx(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)
	cu.AddAsset(EthToken, addr.String(), 1)
	enable := cu.IsEnabledSendTx(EthToken, addr.String())
	assert.True(t, enable)

	// disable send eth tx
	cu.SetEnableSendTx(false, EthToken, addr.String())
	enable = cu.IsEnabledSendTx(EthToken, addr.String())
	assert.False(t, enable)

	// enable send eth tx
	cu.SetEnableSendTx(true, EthToken, addr.String())
	enable = cu.IsEnabledSendTx(EthToken, addr.String())
	assert.True(t, enable)

}

func TestBaseCUBalanceFlows(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)

	someCoins := sdk.NewCoins(sdk.NewInt64Coin("coin2", 246), sdk.NewInt64Coin("coin1", 123), sdk.NewInt64Coin("coin3", 345), sdk.NewInt64Coin("coin4", 0))

	// 3 balance flow
	err := cu.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(cu.balanceFlows))

	bf := cu.balanceFlows[0]
	assert.Equal(t, "coin1", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(123), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin2", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(246), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[2]
	assert.Equal(t, "coin3", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(345), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin5", 500), sdk.NewInt64Coin("coin2", 246), sdk.NewInt64Coin("coin3", 100))
	err = cu.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin1", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(123), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(123).Neg(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin3", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(345), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(245).Neg(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[2]
	assert.Equal(t, "coin5", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(500), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin3", 123), sdk.NewInt64Coin("coin4", 246), sdk.NewInt64Coin("coin0", 0))
	// 2 balance flow
	cu.AddCoins(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin3", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(100), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(123), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin4", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(246), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	// -2 balance flow, symbol balanceFlows's change is 0
	cu.SubCoins(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin3", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(223), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(123).Neg(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin4", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(246), bf.PreviousBalance)
	assert.Equal(t, sdk.NewInt(246).Neg(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin5", 123), sdk.NewInt64Coin("coin6", 246), sdk.NewInt64Coin("coin0", 0))
	// 2 balance flow
	cu.AddCoinsHold(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin5", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(500), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(123), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin6", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(246), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	// 0 balance flow, symbol is already exist in flows
	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin5", 23), sdk.NewInt64Coin("coin6", 46), sdk.NewInt64Coin("coin0", 0))
	cu.SubCoinsHold(someCoins)

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin5", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(500), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(123), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(23).Neg(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin6", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(246), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(46).Neg(), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	// 0 balance flow, symbol is already exist in flows
	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin1", 23), sdk.NewInt64Coin("coin2", 46), sdk.NewInt64Coin("coin0", 0))
	cu.AddCoinsHold(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))
	bf = cu.balanceFlows[0]
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(23), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin2", bf.Symbol.String())
	assert.Equal(t, sdk.NewInt(246), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(46), bf.BalanceOnHoldChange)

}

func TestBaseCUBalanceFlows2(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)

	someCoins := sdk.NewCoins(sdk.NewInt64Coin("coin1", 123), sdk.NewInt64Coin("coin2", 246), sdk.NewInt64Coin("coin3", 345), sdk.NewInt64Coin("coin4", 0))
	// 3 balance flow
	err := cu.SetCoinsHold(someCoins)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(cu.balanceFlows))

	bf := cu.balanceFlows[0]
	assert.Equal(t, "coin1", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(123), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin2", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(246), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[2]
	assert.Equal(t, "coin3", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(345), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin3", 200), sdk.NewInt64Coin("coin5", 500), sdk.NewInt64Coin("coin2", 246), sdk.NewInt64Coin("coin8", 800), sdk.NewInt64Coin("coin9", 0))
	err = cu.SetCoinsHold(someCoins)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin1", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(123), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(123).Neg(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin3", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(345), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(145).Neg(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[2]
	assert.Equal(t, "coin5", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(500), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[3]
	assert.Equal(t, "coin8", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(800), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin5", 123), sdk.NewInt64Coin("coin6", 246), sdk.NewInt64Coin("coin0", 0))
	// 2 balance flow
	cu.AddCoinsHold(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin5", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(500), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(123), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin6", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(246), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin5", 23), sdk.NewInt64Coin("coin6", 46), sdk.NewInt64Coin("coin0", 0))
	cu.SubCoinsHold(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))
	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin5", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(623), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(23).Neg(), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin6", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(246), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(46).Neg(), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

	// 0 balance flow, symbol is already exist in flows
	someCoins = sdk.NewCoins(sdk.NewInt64Coin("coin1", 23), sdk.NewInt64Coin("coin2", 46), sdk.NewInt64Coin("coin0", 0))
	cu.AddCoinsHold(someCoins)
	assert.EqualValues(t, 2, len(cu.balanceFlows))

	bf = cu.balanceFlows[0]
	assert.Equal(t, "coin1", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(0), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(23), bf.BalanceOnHoldChange)

	bf = cu.balanceFlows[1]
	assert.Equal(t, "coin2", bf.Symbol.String())
	assert.Equal(t, sdk.ZeroInt(), bf.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), bf.BalanceChange)
	assert.Equal(t, sdk.NewInt(246), bf.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(46), bf.BalanceOnHoldChange)
	cu.ResetBalanceFlows()

}

func TestBaseCU_AddCoins(t *testing.T) {
	_, _, addr := KeyTestPubAddr()
	cu := NewBaseCUWithAddress(addr, sdk.CUTypeUser)

	someCoins := sdk.NewCoins(sdk.NewInt64Coin("coin3", 30), sdk.NewInt64Coin("coin1", 10), sdk.NewInt64Coin("coin2", 20))
	err := cu.SetCoins(someCoins)
	assert.Nil(t, err)
	cu.ResetBalanceFlows()

	otherCoins := sdk.NewCoins(sdk.NewInt64Coin("coin4", 40), sdk.NewInt64Coin("coin3", 130), sdk.NewInt64Coin("coin2", 120), sdk.NewInt64Coin("coin5", 50))

	cu.AddCoins(otherCoins)
	t.Logf("before sort:%v\n", cu.GetCoins())

	cuCoins2 := cu.GetCoins()
	cuCoins2.Sort()

	t.Logf("after sort, get Coins again:%v\n", cu.GetCoins())
	t.Logf("after sort, cuCoins:%v\n", cuCoins2)

}
