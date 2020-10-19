package types

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestSanitize(t *testing.T) {

	addr1 := sdk.CUAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc1 := custodianunit.NewBaseCUWithAddress(addr1, sdk.CUTypeUser)
	authAcc1.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("bcoin", 150),
		sdk.NewInt64Coin("acoin", 150),
	})

	genAcc1 := NewGenesisCU(&authAcc1)

	addr2 := sdk.CUAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc2 := custodianunit.NewBaseCUWithAddress(addr2, sdk.CUTypeUser)
	authAcc2.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("acoin", 150),
		sdk.NewInt64Coin("bcoin", 150),
	})
	genAcc2 := NewGenesisCU(&authAcc2)

	genesisState := GenesisState([]GenesisCU{genAcc1, genAcc2})
	require.NoError(t, ValidateGenesis(genesisState))
	require.Equal(t, genesisState[0].Coins[0].Denom, "bcoin")
	require.Equal(t, genesisState[0].Coins[1].Denom, "acoin")
	require.Equal(t, genesisState[1].Address, addr2)
	genesisState.Sanitize()
	// guard genesisState is sorted by cuaddress not cuNumber
	if addr1.String() > addr2.String() {
		require.Equal(t, genesisState[1].Address, addr1)

	} else {
		require.Equal(t, genesisState[1].Address, addr2)
	}
	require.Equal(t, genesisState[1].Coins[0].Denom, "acoin")
	require.Equal(t, genesisState[1].Coins[1].Denom, "bcoin")
}

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require duplicate accounts fails validation
func TestValidateGenesisDuplicateCUs(t *testing.T) {
	acc1 := custodianunit.NewBaseCUWithAddress(sdk.CUAddress(addr1), sdk.CUTypeUser)
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]GenesisCU, 2)
	genAccs[0] = NewGenesisCU(&acc1)
	genAccs[1] = NewGenesisCU(&acc1)

	genesisState := GenesisState(genAccs)
	err := ValidateGenesis(genesisState)
	require.Error(t, err)
}

// require invalid vesting CU fails validation (invalid end time)
func TestValidateGenesisInvalidCUs(t *testing.T) {
	acc1 := custodianunit.NewBaseCUWithAddress(sdk.CUAddress(addr1), sdk.CUTypeUser)
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))
	acc2 := custodianunit.NewBaseCUWithAddress(sdk.CUAddress(addr2), sdk.CUTypeUser)
	acc2.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]GenesisCU, 2)
	genAccs[0] = NewGenesisCU(&acc1)
	genAccs[1] = NewGenesisCU(&acc2)

	genesisState := GenesisState(genAccs)
	err := ValidateGenesis(genesisState)
	require.NoError(t, err)

}
