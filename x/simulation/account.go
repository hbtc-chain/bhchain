package simulation

import (
	"math/rand"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// CU contains a privkey, pubkey, address tuple
// eventually more useful data can be placed in here.
// (e.g. number of coins)
type CU struct {
	PrivKey crypto.PrivKey
	PubKey  crypto.PubKey
	Address sdk.CUAddress
}

// are two accounts equal
func (c CU) Equals(cu2 CU) bool {
	return c.Address.Equals(cu2.Address)
}

// RandomAcc pick a random CU from an array
func RandomAcc(r *rand.Rand, cus []CU) CU {
	return cus[r.Intn(
		len(cus),
	)]
}

// RandomCUs generates n random accounts
func RandomCUs(r *rand.Rand, n int) []CU {
	accs := make([]CU, n)
	for i := 0; i < n; i++ {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)
		useSecp := r.Int63()%2 == 0
		if useSecp {
			accs[i].PrivKey = secp256k1.GenPrivKeySecp256k1(privkeySeed)
		} else {
			accs[i].PrivKey = ed25519.GenPrivKeyFromSecret(privkeySeed)
		}

		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.CUAddress(accs[i].PubKey.Address())
	}

	return accs
}
