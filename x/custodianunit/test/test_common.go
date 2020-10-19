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
	"github.com/hbtc-chain/bhchain/x/params/subspace"
	"github.com/hbtc-chain/bhchain/x/supply/exported"
	supplyexported "github.com/hbtc-chain/bhchain/x/supply/exported"
	"github.com/hbtc-chain/bhchain/x/token"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  cu.CUKeeper
	sk  internal.SupplyKeeper
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
	cdc.RegisterConcrete(&moduleAccount{}, "hbtcchain/ModuleAccount", nil)
	// remove this sentence,after sdk.OpCUInfo move to settle
	cdc.RegisterConcrete(&sdk.OpCUInfo{}, "cu/sdk.OpCUInfo", nil)
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
	tk := token.NewKeeper(tokenKey, cdc, subspace.NewSubspace(cdc, keyParams, tkeyParams, token.DefaultParamspace))

	ak := cu.NewCUKeeper(cdc, authCapKey, &tk, ps, types.ProtoBaseCU)
	sk := NewDummySupplyKeeper(ak)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, types.DefaultParams())

	//init token info
	for _, tokenInfo := range token.TestTokenData {
		token := token.NewTokenInfo(tokenInfo.Symbol, tokenInfo.Chain, tokenInfo.Issuer, tokenInfo.TokenType,
			tokenInfo.IsSendEnabled, tokenInfo.IsDepositEnabled, tokenInfo.IsWithdrawalEnabled, tokenInfo.Decimals,
			tokenInfo.TotalSupply, tokenInfo.CollectThreshold, tokenInfo.DepositThreshold, tokenInfo.OpenFee,
			tokenInfo.SysOpenFee, tokenInfo.WithdrawalFeeRate, tokenInfo.SysTransferNum, tokenInfo.OpCUSysTransferNum,
			tokenInfo.GasLimit, tokenInfo.GasPrice, tokenInfo.MaxOpCUNumber, tokenInfo.Confirmations, tokenInfo.IsNonceBased) //WithdrawalAddress and depositAddress will be added later.
		tk.SetTokenInfo(ctx, token)
	}

	return testInput{cdc: cdc, ctx: ctx, ak: ak, sk: sk}
}

// DummySupplyKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type DummySupplyKeeper struct {
	ak cu.CUKeeper
}

// NewDummySupplyKeeper creates a DummySupplyKeeper instance
func NewDummySupplyKeeper(ak cu.CUKeeper) DummySupplyKeeper {
	return DummySupplyKeeper{ak}
}

// SendCoinsFromAccountToModule for the dummy supply keeper
func (sk DummySupplyKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, fromAddr sdk.CUAddress, recipientModule string, amt sdk.Coins) (sdk.Result, sdk.Error) {

	fromAcc := sk.ak.GetCU(ctx, fromAddr)
	moduleAcc := sk.GetModuleAccount(ctx, recipientModule)

	newFromCoins, hasNeg := fromAcc.GetCoins().SafeSub(amt)
	if hasNeg {
		return sdk.Result{}, sdk.ErrInsufficientCoins(fromAcc.GetCoins().String())
	}

	newToCoins := moduleAcc.GetCoins().Add(amt)

	if err := fromAcc.SetCoins(newFromCoins); err != nil {
		return sdk.Result{}, sdk.ErrInternal(err.Error())
	}

	if err := moduleAcc.SetCoins(newToCoins); err != nil {
		return sdk.Result{}, sdk.ErrInternal(err.Error())
	}

	sk.ak.SetCU(ctx, fromAcc)
	sk.ak.SetCU(ctx, moduleAcc)

	return sdk.Result{}, nil
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
	tk := token.NewKeeper(tokenKey, cdc, subspace.NewSubspace(cdc, keyParams, tkeyParams, token.DefaultParamspace))
	ak := cu.NewCUKeeper(cdc, authCapKey, &tk, ps, types.ProtoBaseCU)
	sk := DummySupplyKeeper{}

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, types.DefaultParams())
	//init token info
	for _, tokenInfo := range token.TestTokenData {
		token := token.NewTokenInfo(tokenInfo.Symbol, tokenInfo.Chain, tokenInfo.Issuer, tokenInfo.TokenType,
			tokenInfo.IsSendEnabled, tokenInfo.IsDepositEnabled, tokenInfo.IsWithdrawalEnabled, tokenInfo.Decimals,
			tokenInfo.TotalSupply, tokenInfo.CollectThreshold, tokenInfo.DepositThreshold, tokenInfo.OpenFee,
			tokenInfo.SysOpenFee, tokenInfo.WithdrawalFeeRate, tokenInfo.SysTransferNum, tokenInfo.OpCUSysTransferNum,
			tokenInfo.GasLimit, tokenInfo.GasPrice, tokenInfo.MaxOpCUNumber, tokenInfo.Confirmations, tokenInfo.IsNonceBased) //WithdrawalAddress and depositAddress will be added later.
		tk.SetTokenInfo(ctx, token)
	}

	return testInputForCUKeeper{Cdc: cdc, Ctx: ctx, Ck: ak, Sk: sk}
}

type testInputForCUKeeper struct {
	Cdc *codec.Codec
	Ctx sdk.Context
	Ck  cu.CUKeeperI
	Sk  internal.SupplyKeeper // use by test cases out of this package avoid cycle import. current is nil
}
