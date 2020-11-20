// nolint
package test

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	cu "github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/internal"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/params/subspace"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  cu.CUKeeper
	sk  internal.SupplyKeeper
	ik  ibcasset.IBCAssetKeeperI
	tk  transfer.Keeper
}

// moduleAccount defines an CustodianUnit for modules that holds coins on a pool
type moduleAccount struct {
	*types.BaseCU
	name        string   `json:"name" yaml:"name"`              // name of the module
	permissions []string `json:"permissions" yaml"permissions"` // permissions of module CustodianUnit
}

// HasPermission returns whether or not the module CustodianUnit has permission.
func (ma moduleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the the name of the holder's module
func (ma moduleAccount) GetName() string {
	return ma.name
}

// GetPermissions returns permissions granted to the module CustodianUnit
func (ma moduleAccount) GetPermissions() []string {
	return ma.permissions
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	cdc.RegisterInterface((*exported.ModuleAccountI)(nil), nil)
	// remove this sentence,after sdk.OpCUInfo move to settle
	cdc.RegisterConcrete(&sdk.OpCUAstInfo{}, "cu/sdk.OpCUAstInfo", nil)
	codec.RegisterCrypto(cdc)
	receipt.RegisterCodec(cdc)
	transfer.RegisterCodec(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyReceipt := sdk.NewKVStoreKey(receipt.StoreKey)
	keyParams := sdk.NewKVStoreKey("subspace")
	keyTransfer := sdk.NewKVStoreKey(transfer.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")
	keyIbcasset := sdk.NewKVStoreKey(ibcasset.StoreKey)
	tokenKey := sdk.NewKVStoreKey(token.ModuleName)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyReceipt, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyTransfer, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tokenKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIbcasset, sdk.StoreTypeIAVL, db)

	ms.LoadLatestVersion()

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	rk := receipt.NewKeeper(cdc)
	tk := token.NewKeeper(tokenKey, cdc)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := cu.NewCUKeeper(cdc, authCapKey, ps, types.ProtoBaseCU)
	transferKeeper := transfer.NewBaseKeeper(cdc, keyTransfer, ak, nil, &tk, nil, rk, nil, nil, pk.Subspace(transfer.DefaultParamspace), transfer.DefaultCodespace, nil)
	sk := NewDummySupplyKeeper(cdc, ak, transferKeeper)
	ik := ibcasset.NewKeeper(cdc, keyIbcasset, ak, &tk, ibcasset.ProtoBaseCUIBCAsset)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, types.DefaultParams())

	//init token info
	for _, tokenInfo := range token.TestTokenData {
		tk.CreateToken(ctx, tokenInfo)
	}

	return testInput{cdc: cdc, ctx: ctx, ak: ak, sk: sk, ik: ik, tk: transferKeeper}
}

// DummySupplyKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type DummySupplyKeeper struct {
	ak cu.CUKeeper
	tk internal.TransferKeeper
}

// NewDummySupplyKeeper creates a DummySupplyKeeper instance
func NewDummySupplyKeeper(cdc *codec.Codec, ak cu.CUKeeper, tk internal.TransferKeeper) DummySupplyKeeper {
	cdc.RegisterConcrete(&moduleAccount{}, "hbtcchain/test/ModuleAccount", nil)
	return DummySupplyKeeper{ak, tk}
}

// SendCoinsFromAccountToModule for the dummy supply keeper
func (sk DummySupplyKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, fromAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error) {

	moduleAcc := sk.GetModuleAccount(ctx, recipientModule)
	res, _, err := sk.tk.SendCoins(ctx, fromAddr, moduleAcc.GetAddress(), amt)
	return res, err
}

func (sk DummySupplyKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.CUAddress, amt sdk.Coins) (sdk.Result, sdk.Error) {

	moduleAcc := sk.GetModuleAccount(ctx, senderModule)
	res, _, err := sk.tk.SendCoins(ctx, moduleAcc.GetAddress(), recipientAddr, amt)
	return res, err
}

// GetModuleAccount for dummy supply keeper
func (sk DummySupplyKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI {
	addr := sk.GetModuleAddress(moduleName)

	cu := sk.ak.GetCU(ctx, addr)
	if cu != nil {
		macc, ok := cu.(exported.ModuleAccountI)
		if ok {
			return macc
		}
	}

	moduleAddress := sk.GetModuleAddress(moduleName)
	baseAcc := types.NewBaseCUWithAddress(moduleAddress, sdk.CUTypeUser)

	// create a new module CustodianUnit
	macc := &moduleAccount{
		BaseCU:      &baseAcc,
		name:        moduleName,
		permissions: []string{"basic"},
	}

	maccI := (sk.ak.NewCU(ctx, macc)).(exported.ModuleAccountI)
	sk.ak.SetCU(ctx, maccI)
	return maccI
}

// GetModuleAddress for dummy supply keeper
func (sk DummySupplyKeeper) GetModuleAddress(moduleName string) sdk.CUAddress {
	return sdk.CUAddress(crypto.AddressHash([]byte(moduleName)))
}

var (
	ethToken  = "eth"
	btcToken  = "btc"
	usdtToken = "usdt"
)

func setupTestInputForCUKeeper() testInputForCUKeeper {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	cdc.RegisterInterface((*supplyexported.ModuleAccountI)(nil), nil)
	cdc.RegisterConcrete(&moduleAccount{}, "hbtcchain/ModuleAccount", nil)
	codec.RegisterCrypto(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")
	tokenKey := sdk.NewKVStoreKey(token.ModuleName)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(tokenKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	tk := token.NewKeeper(tokenKey, cdc)
	ak := cu.NewCUKeeper(cdc, authCapKey, ps, types.ProtoBaseCU)
	sk := DummySupplyKeeper{}

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, types.DefaultParams())
	//init token info
	for _, tokenInfo := range token.TestTokenData {
		tk.CreateToken(ctx, tokenInfo)
	}

	return testInputForCUKeeper{Cdc: cdc, Ctx: ctx, Ck: ak, Sk: sk}
}

type testInputForCUKeeper struct {
	Cdc *codec.Codec
	Ctx sdk.Context
	Ck  cu.CUKeeperI
	Sk  internal.SupplyKeeper // use by test cases out of this package avoid cycle import. current is nil
}
