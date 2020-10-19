package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	tmcrypto "github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/hbtc-chain/bhchain/crypto/keys/hd"
	"github.com/hbtc-chain/bhchain/tests"
	sdk "github.com/hbtc-chain/bhchain/types"
)

func TestLedgerErrorHandling(t *testing.T) {
	// first, try to generate a key, must return an error
	// (no panic)
	path := *hd.NewParams(44, 555, 0, false, 0)
	_, err := NewPrivKeyLedgerSecp256k1Unsafe(path)
	require.Error(t, err)
}

func TestPublicKeyUnsafe(t *testing.T) {
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := NewPrivKeyLedgerSecp256k1Unsafe(path)
	require.Nil(t, err, "%s", err)
	require.NotNil(t, priv)

	require.Equal(t, "eb5ae9872103db6a5d05fff3c211e8694b3e5fb202ab0bbb818d9f68ad5924e84786e8a3a364",
		fmt.Sprintf("%x", priv.PubKey().Bytes()),
		"Is your device using test mnemonic: %s ?", tests.TestMnemonic)

	pubKeyAddr, err := sdk.Bech32ifyAccPub(priv.PubKey())
	require.NoError(t, err)
	require.Equal(t, "hbcpub1addwnpepq0dk5hg9lleuyy0gd99nuhajq24shwup3k0k3t2eyn5y0phg5w3kgeeygy4",
		pubKeyAddr, "Is your device using test mnemonic: %s ?", tests.TestMnemonic)

	addr := sdk.CUAddress(priv.PubKey().Address()).String()
	require.Equal(t, "HBCd5RfRgxsRveWVWM2hko5hVP3dyGpEtTa6",
		addr, "Is your device using test mnemonic: %s ?", tests.TestMnemonic)
}

func TestPublicKeyUnsafeHDPath(t *testing.T) {
	expectedAnswers := []string{
		"hbcpub1addwnpepq0dk5hg9lleuyy0gd99nuhajq24shwup3k0k3t2eyn5y0phg5w3kgeeygy4",
		"hbcpub1addwnpepqgrk57k7w8y6u43crtxqq4zmrxgy8nxz3gsku6j3p5ssh0dul96p7c7q3kw",
		"hbcpub1addwnpepqgqvheac6wud0dwsmewf4vdm68we88xt8zkzcc63f39l3aqfvjpmqvursdd",
		"hbcpub1addwnpepq2ux9a4llpcz0mn4hs69p46axxe0r2txj5dzys4f8cqukjpyyrw0zkf82h2",
		"hbcpub1addwnpepqg20pvw7m9707nyfg0sjfhd2lemqxfppwzda06v6snlqjpqn6ceqwt89dvf",
		"hbcpub1addwnpepqgmv0ks8cvryqgv4uprm6nfecxvy3fmrfewwgd2uwpq0uncxgxq0y9r8h7c",
		"hbcpub1addwnpepqvhhtlrwt578fxc6av0dpv39yw304aqpjvjykueg3dtrq8hq08e7s7ag80n",
		"hbcpub1addwnpepq200srf45h3nn94tu8xn7ek0wqzrfqmtf2nv9l67kska4y9svwcdz3pz2yj",
		"hbcpub1addwnpepqvpt9k80545zjahnlktwhvvu5a85rq8nhu960nwx9endlgywupjrkcvph7f",
		"hbcpub1addwnpepqvn3guvd6nmxzu8afmcug48zhk25kkh7sdlm9mykml5qcekhts376zatl0r",
	}

	const numIters = 10

	privKeys := make([]tmcrypto.PrivKey, numIters)

	// Check with device
	for i := uint32(0); i < 10; i++ {
		path := *hd.NewFundraiserParams(0, sdk.CoinType, i)
		fmt.Printf("Checking keys at %v\n", path)

		priv, err := NewPrivKeyLedgerSecp256k1Unsafe(path)
		require.Nil(t, err, "%s", err)
		require.NotNil(t, priv)

		// Check other methods
		require.NoError(t, priv.(PrivKeyLedgerSecp256k1).ValidateKey())
		tmp := priv.(PrivKeyLedgerSecp256k1)
		(&tmp).AssertIsPrivKeyInner()

		pubKeyAddr, err := sdk.Bech32ifyAccPub(priv.PubKey())
		require.NoError(t, err)
		require.Equal(t,
			expectedAnswers[i], pubKeyAddr,
			"Is your device using test mnemonic: %s ?", tests.TestMnemonic)

		// Store and restore
		serializedPk := priv.Bytes()
		require.NotNil(t, serializedPk)
		require.True(t, len(serializedPk) >= 50)

		privKeys[i] = priv
	}

	// Now check equality
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			require.Equal(t, i == j, privKeys[i].Equals(privKeys[j]))
			require.Equal(t, i == j, privKeys[j].Equals(privKeys[i]))
		}
	}
}

func TestPublicKeySafe(t *testing.T) {
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, addr, err := NewPrivKeyLedgerSecp256k1(path, "cosmos")

	require.Nil(t, err, "%s", err)
	require.NotNil(t, priv)

	require.Equal(t, "eb5ae9872103db6a5d05fff3c211e8694b3e5fb202ab0bbb818d9f68ad5924e84786e8a3a364",
		fmt.Sprintf("%x", priv.PubKey().Bytes()),
		"Is your device using test mnemonic: %s ?", tests.TestMnemonic)

	pubKeyAddr, err := sdk.Bech32ifyAccPub(priv.PubKey())
	require.NoError(t, err)
	require.Equal(t, "hbcpub1addwnpepq0dk5hg9lleuyy0gd99nuhajq24shwup3k0k3t2eyn5y0phg5w3kgeeygy4",
		pubKeyAddr, "Is your device using test mnemonic: %s ?", tests.TestMnemonic)

	require.Equal(t, "HBCd5RfRgxsRveWVWM2hko5hVP3dyGpEtTa6",
		addr, "Is your device using test mnemonic: %s ?", tests.TestMnemonic)

	addr2 := sdk.CUAddress(priv.PubKey().Address()).String()
	require.Equal(t, addr, addr2)
}

func TestPublicKeyHDPath(t *testing.T) {
	expectedPubKeys := []string{
		"hbcpub1addwnpepq0dk5hg9lleuyy0gd99nuhajq24shwup3k0k3t2eyn5y0phg5w3kgeeygy4",
		"hbcpub1addwnpepqgrk57k7w8y6u43crtxqq4zmrxgy8nxz3gsku6j3p5ssh0dul96p7c7q3kw",
		"hbcpub1addwnpepqgqvheac6wud0dwsmewf4vdm68we88xt8zkzcc63f39l3aqfvjpmqvursdd",
		"hbcpub1addwnpepq2ux9a4llpcz0mn4hs69p46axxe0r2txj5dzys4f8cqukjpyyrw0zkf82h2",
		"hbcpub1addwnpepqg20pvw7m9707nyfg0sjfhd2lemqxfppwzda06v6snlqjpqn6ceqwt89dvf",
		"hbcpub1addwnpepqgmv0ks8cvryqgv4uprm6nfecxvy3fmrfewwgd2uwpq0uncxgxq0y9r8h7c",
		"hbcpub1addwnpepqvhhtlrwt578fxc6av0dpv39yw304aqpjvjykueg3dtrq8hq08e7s7ag80n",
		"hbcpub1addwnpepq200srf45h3nn94tu8xn7ek0wqzrfqmtf2nv9l67kska4y9svwcdz3pz2yj",
		"hbcpub1addwnpepqvpt9k80545zjahnlktwhvvu5a85rq8nhu960nwx9endlgywupjrkcvph7f",
		"hbcpub1addwnpepqvn3guvd6nmxzu8afmcug48zhk25kkh7sdlm9mykml5qcekhts376zatl0r",
	}

	expectedAddrs := []string{
		"HBCd5RfRgxsRveWVWM2hko5hVP3dyGpEtTa6",
		"HBCbHM2zZk2PpzPoUXSL1DJRjB4tczAwdune",
		"HBCfTUh7cUKQJtQD8AHYRGQZGzukByN7uuCZ",
		"HBCjUYDutpT23kpWpusccVQ99TY3D92o61yZ",
		"HBCVHPsspvjWoyPaZz7vK3DYMWqhfkX2RpeV",
		"HBCgoaGABZNSyTvUZ8MbBym1iidTSNHmaJW1",
		"HBCLip9MBLhD7kQcKe6fcpDp2NsDYyZYnsvM",
		"HBCSw5uwfZb7ZTxvZGG8fvKETg2XP2X9ovP7",
		"HBCaY7Q1GfUfP9CWxPZrjyfayRefsGsjQhX8",
		"HBCg3yd7mLSjiERoaCaeVjAhRXePjXvhZTvz",
	}

	const numIters = 10

	privKeys := make([]tmcrypto.PrivKey, numIters)

	// Check with device
	for i := uint32(0); i < 10; i++ {
		path := *hd.NewFundraiserParams(0, sdk.CoinType, i)
		fmt.Printf("Checking keys at %v\n", path)

		priv, addr, err := NewPrivKeyLedgerSecp256k1(path, "cosmos")
		require.Nil(t, err, "%s", err)
		require.NotNil(t, addr)
		require.NotNil(t, priv)

		addr2 := sdk.CUAddress(priv.PubKey().Address()).String()
		require.Equal(t, addr2, addr)
		require.Equal(t,
			expectedAddrs[i], addr,
			"Is your device using test mnemonic: %s ?", tests.TestMnemonic)

		// Check other methods
		require.NoError(t, priv.(PrivKeyLedgerSecp256k1).ValidateKey())
		tmp := priv.(PrivKeyLedgerSecp256k1)
		(&tmp).AssertIsPrivKeyInner()

		pubKeyAddr, err := sdk.Bech32ifyAccPub(priv.PubKey())
		require.NoError(t, err)
		require.Equal(t,
			expectedPubKeys[i], pubKeyAddr,
			"Is your device using test mnemonic: %s ?", tests.TestMnemonic)

		// Store and restore
		serializedPk := priv.Bytes()
		require.NotNil(t, serializedPk)
		require.True(t, len(serializedPk) >= 50)

		privKeys[i] = priv
	}

	// Now check equality
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			require.Equal(t, i == j, privKeys[i].Equals(privKeys[j]))
			require.Equal(t, i == j, privKeys[j].Equals(privKeys[i]))
		}
	}
}

func getFakeTx(cuNumber uint32) []byte {
	tmp := fmt.Sprintf(
		`{"cu_number":"%d","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"5000"},"memo":"memo","msgs":[[""]],"sequence":"6"}`,
		cuNumber)

	return []byte(tmp)
}

func TestSignaturesHD(t *testing.T) {
	for CU := uint32(0); CU < 100; CU += 30 {
		msg := getFakeTx(CU)

		path := *hd.NewFundraiserParams(CU, sdk.CoinType, CU/5)
		fmt.Printf("Checking signature at %v    ---   PLEASE REVIEW AND ACCEPT IN THE DEVICE\n", path)

		priv, err := NewPrivKeyLedgerSecp256k1Unsafe(path)
		require.Nil(t, err, "%s", err)

		pub := priv.PubKey()
		sig, err := priv.Sign(msg)
		require.Nil(t, err)

		valid := pub.VerifyBytes(msg, sig)
		require.True(t, valid, "Is your device using test mnemonic: %s ?", tests.TestMnemonic)
	}
}

func TestRealLedgerSecp256k1(t *testing.T) {
	msg := getFakeTx(50)
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := NewPrivKeyLedgerSecp256k1Unsafe(path)
	require.Nil(t, err, "%s", err)

	pub := priv.PubKey()
	sig, err := priv.Sign(msg)
	require.Nil(t, err)

	valid := pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// now, let's serialize the public key and make sure it still works
	bs := priv.PubKey().Bytes()
	pub2, err := cryptoAmino.PubKeyFromBytes(bs)
	require.Nil(t, err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig, err = priv.Sign(msg)
	require.Nil(t, err)
	valid = pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = pub.Bytes()
	bpub, err := cryptoAmino.PubKeyFromBytes(bs)
	require.NoError(t, err)
	require.Equal(t, pub, bpub)
}
