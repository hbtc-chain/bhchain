package server

import (
	"fmt"

	"github.com/hbtc-chain/bhchain/crypto/keys"

	clkeys "github.com/hbtc-chain/bhchain/client/keys"
	sdk "github.com/hbtc-chain/bhchain/types"
)

// GenerateCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateCoinKey() (sdk.CUAddress, string, error) {

	// generate a private key, with recovery phrase
	info, secret, err := clkeys.NewInMemoryKeyBase().CreateMnemonic(
		"name", keys.English, "pass", keys.Secp256k1)
	if err != nil {
		return sdk.CUAddress([]byte{}), "", err
	}
	addr := info.GetPubKey().Address()
	return sdk.CUAddress(addr), secret, nil
}

// GenerateSaveCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateSaveCoinKey(clientRoot, keyName, keyPass string,
	overwrite bool) (sdk.CUAddress, string, error) {

	// get the keystore from the client
	keybase, err := clkeys.NewKeyBaseFromDir(clientRoot)
	if err != nil {
		return sdk.CUAddress([]byte{}), "", err
	}

	// ensure no overwrite
	if !overwrite {
		_, err := keybase.Get(keyName)
		if err == nil {
			return sdk.CUAddress([]byte{}), "", fmt.Errorf(
				"key already exists, overwrite is disabled (clientRoot: %s)", clientRoot)
		}
	}

	// generate a private key, with recovery phrase
	info, secret, err := keybase.CreateMnemonic(keyName, keys.English, keyPass, keys.Secp256k1)
	if err != nil {
		return sdk.CUAddress([]byte{}), "", err
	}

	return sdk.CUAddress(info.GetPubKey().Address()), secret, nil
}
