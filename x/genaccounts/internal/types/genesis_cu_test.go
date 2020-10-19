package types

import (
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/supply"
)

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	pubkey2 := secp256k1.GenPrivKey().PubKey()
	addr := sdk.CUAddress(pubkey.Address())
	tests := []struct {
		name   string
		acc    GenesisCU
		expErr error
	}{
		{
			"valid CU",
			NewGenesisCURaw(sdk.CUTypeUser, pubkey, pubkey2.Bytes(), addr, sdk.NewCoins(), sdk.NewCoins(),
				sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), []sdk.Asset{}, "", ""),
			nil,
		},
		{
			"valid module CU",
			NewGenesisCURaw(sdk.CUTypeUser, pubkey, pubkey2.Bytes(), addr, sdk.NewCoins(), sdk.NewCoins(),
				sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), []sdk.Asset{}, "mint", supply.Minter),
			nil,
		},

		{
			"invalid module CU name",
			NewGenesisCURaw(sdk.CUTypeUser, pubkey, pubkey2.Bytes(), addr, sdk.NewCoins(), sdk.NewCoins(),
				sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), []sdk.Asset{}, " ", supply.Minter),
			errors.New("module CustodianUnit name cannot be blank"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			require.Equal(t, tt.expErr, err)
		})
	}
}
