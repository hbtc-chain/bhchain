package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/supply/internal/types"
	"github.com/hbtc-chain/bhchain/x/transfer"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// nolint: deadcode unused
var (
	multiPerm  = "multiple permissions CU"
	randomPerm = "random permission"
	holder     = "holder"
)

// nolint: deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.New()

	transfer.RegisterCodec(cdc)
	custodianunit.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	receipt.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

// nolint: deadcode unused
func createTestInput(t *testing.T, isCheckTx bool, initPower int64, nAccs int64) (sdk.Context, custodianunit.CUKeeper, Keeper) {

	keyAcc := sdk.NewKVStoreKey(custodianunit.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(types.StoreKey)
	keyTransfer := sdk.NewKVStoreKey(transfer.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "supply-chain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	cdc := makeTestCodec()

	blacklistedAddrs := make(map[string]bool)
	rk := receipt.NewKeeper(cdc)
	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := custodianunit.NewCUKeeper(cdc, keyAcc, nil, pk.Subspace(custodianunit.DefaultParamspace), custodianunit.ProtoBaseCU)
	bk := transfer.NewBaseKeeper(cdc, keyTransfer, ak, nil, nil, rk, nil, nil,
		pk.Subspace(transfer.DefaultParamspace), transfer.DefaultCodespace, blacklistedAddrs)

	valTokens := sdk.TokensFromConsensusPower(initPower)

	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	createTestAccs(ctx, int(nAccs), initialCoins, &ak)

	maccPerms := map[string][]string{
		holder:       nil,
		types.Minter: []string{types.Minter},
		types.Burner: []string{types.Burner},
		multiPerm:    []string{types.Minter, types.Burner, types.Staking},
		randomPerm:   []string{"random"},
	}
	keeper := NewKeeper(cdc, keySupply, ak, bk, maccPerms)
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens.MulRaw(nAccs)))
	keeper.SetSupply(ctx, types.NewSupply(totalSupply))

	return ctx, ak, *keeper
}

// nolint: unparam deadcode unused
func createTestAccs(ctx sdk.Context, numAccs int, initialCoins sdk.Coins, ak *custodianunit.CUKeeper) (accs []custodianunit.CU) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.CUAddress(pubKey.Address())
		acc := custodianunit.NewBaseCUWithAddress(addr, sdk.CUTypeUser)
		acc.Coins = initialCoins
		acc.PubKey = pubKey
		ak.SetCU(ctx, &acc)
	}
	return
}
