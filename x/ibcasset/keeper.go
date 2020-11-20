package ibcasset

import (
	"fmt"
	"github.com/hbtc-chain/bhchain/x/ibcasset/internal"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/ibcasset/exported"
	"github.com/hbtc-chain/bhchain/x/ibcasset/types"
)

type IBCAssetKeeperI interface {
	// CU functions
	NewCUIBCAssetWithAddress(ctx sdk.Context, cuType sdk.CUType, cuaddr sdk.CUAddress) exported.CUIBCAsset

	GetCUIBCAsset(context sdk.Context, addresses sdk.CUAddress) exported.CUIBCAsset

	GetOrNewCUIBCAsset(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CUIBCAsset

	SetCUIBCAsset(ctx sdk.Context, cu exported.CUIBCAsset)

	Logger(ctx sdk.Context) log.Logger

	GetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress) sdk.DepositList
	GetDepositListByHash(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string) sdk.DepositList
	SetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress, list sdk.DepositList)
	SaveDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, deposit sdk.DepositItem) error
	DelDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64)
	SetDepositStatus(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64, status sdk.DepositItemStatus) error
	GetDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) sdk.DepositItem
	IsDepositExist(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) bool
}

var _ IBCAssetKeeperI = (*Keeper)(nil)

// Keeper encodes/decodes CUs using the go-amino (binary)
// encoding/decoding library.
type Keeper struct {
	// The (unexposed) key used to access the store from the Context.
	key   sdk.StoreKey
	proto func() exported.CUIBCAsset

	ck internal.CUKeeper
	tk internal.TokenKeeper

	cdc *codec.Codec
}

// NewKeeper returns a new Keeper that uses go-amino to
// (binary) encode and decode concrete sdk.CustodianUnits.
// nolint
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, ck internal.CUKeeper, tk internal.TokenKeeper, proto func() exported.CUIBCAsset) Keeper {
	return Keeper{
		key:   key,
		cdc:   cdc,
		proto: proto,
		ck:    ck,
		tk:    tk,
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (keeper Keeper) GetCUIBCAsset(context sdk.Context, addr sdk.CUAddress) exported.CUIBCAsset {
	if addr == nil {
		return nil
	}

	astCU := keeper.getCUAsset(context, addr)
	return astCU
}

func (keeper Keeper) GetOrNewCUIBCAsset(context sdk.Context, cuType sdk.CUType, addr sdk.CUAddress) exported.CUIBCAsset {
	if addr == nil {
		return nil
	}

	astCU := keeper.getCUAsset(context, addr)
	if astCU == nil {
		newCU := keeper.NewCUIBCAssetWithAddress(context, cuType, addr)
		return newCU
	}
	return astCU
}

// SetCU implements sdk.Keeper.
func (keeper Keeper) SetCUIBCAsset(ctx sdk.Context, cuAst exported.CUIBCAsset) {
	addr := cuAst.GetAddress()
	store := ctx.KVStore(keeper.key)

	bz, err := keeper.cdc.MarshalBinaryBare(cuAst)
	if err != nil {
		panic(err)
	}

	store.Set(types.AddressStoreKey(addr), bz)
}

// IterateCUs implements sdk.Keeper.
func (keeper Keeper) IterateCUAssets(ctx sdk.Context, process func(asset exported.CUIBCAsset) (stop bool)) {
	store := ctx.KVStore(keeper.key)
	iter := sdk.KVStorePrefixIterator(store, types.AddressStoreKeyPrefix)
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		cu := keeper.decodeCUAst(val)
		if process(cu) {
			return
		}
		iter.Next()
	}
}

func (keeper Keeper) GetPendingDepositList(ctx sdk.Context, address sdk.CUAddress) map[string]sdk.DepositList {
	store := ctx.KVStore(keeper.key)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositStorePrefixKeyWithAddr(address))
	defer iterator.Close()

	deposits := make(map[string]sdk.DepositList)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dl sdk.DepositItem
		keeper.cdc.UnmarshalBinaryBare(bz, &dl)
		if dl.Status == sdk.DepositItemStatusUnCollected {
			token := types.DecodeSymbolFromDepositListKey(iterator.Key())
			deposits[token] = append(deposits[token], dl)
		}
	}
	return deposits

}

// ----------------------------------------------------------------------
// DepositList funcs
func (keeper Keeper) GetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress) sdk.DepositList {
	store := ctx.KVStore(keeper.key)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositStorePrefixKey(symbol, address))
	defer iterator.Close()

	var dls sdk.DepositList
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dl sdk.DepositItem
		keeper.cdc.UnmarshalBinaryBare(bz, &dl)
		dls.AddDepositItem(dl)
	}
	return dls
}

func (keeper Keeper) GetDepositListByHash(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string) sdk.DepositList {
	dls := keeper.GetDepositList(ctx, symbol, address)
	dlsGot := dls.Filter(func(d sdk.DepositItem) bool {
		return d.GetHash() == hash
	})
	return dlsGot
}

func (keeper Keeper) SetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress, list sdk.DepositList) {
	for _, item := range list {
		keeper.SaveDeposit(ctx, symbol, address, item)
	}
}

// GetDeposit get deposit item from store
func (keeper Keeper) GetDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) sdk.DepositItem {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(types.DepositStoreKey(symbol, address, hash, index))
	if bz == nil {
		return sdk.DepositNil
	}
	var item sdk.DepositItem
	keeper.cdc.MustUnmarshalBinaryBare(bz, &item)

	return item
}

func (keeper Keeper) IsDepositExist(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) bool {
	store := ctx.KVStore(keeper.key)
	return store.Has(types.DepositStoreKey(symbol, address, hash, index))
}

// SaveDeposit save deposit item to store
func (keeper Keeper) SaveDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, depositItem sdk.DepositItem) error {
	store := ctx.KVStore(keeper.key)

	bz, err := keeper.cdc.MarshalBinaryBare(depositItem)
	if err != nil {
		return err
	}

	store.Set(types.DepositStoreKey(symbol, address, depositItem.GetHash(), depositItem.GetIndex()), bz)

	return nil
}

// DelDeposit delete deposit item from store
func (keeper Keeper) DelDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) {
	store := ctx.KVStore(keeper.key)
	store.Delete(types.DepositStoreKey(symbol, address, hash, index))
}

func (keeper Keeper) SetDepositStatus(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64, stautus sdk.DepositItemStatus) error {
	item := keeper.GetDeposit(ctx, symbol, address, hash, index)
	if item == sdk.DepositNil {
		return fmt.Errorf("deposit not exist %v%v", hash, index)
	}

	item.Status = stautus
	keeper.SaveDeposit(ctx, symbol, address, item)
	return nil
}

func (keeper Keeper) getCUAsset(ctx sdk.Context, addr sdk.CUAddress) exported.CUIBCAsset {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(types.AddressStoreKey(addr))
	if bz == nil {
		return nil
	}

	cu := keeper.decodeCUAst(bz)
	return cu
}

func (keeper Keeper) decodeCUAst(bz []byte) (cu exported.CUIBCAsset) {
	err := keeper.cdc.UnmarshalBinaryBare(bz, &cu)
	if err != nil {
		panic(err)
	}
	return
}

func (keeper Keeper) decodeDeposit(bz []byte) (dls sdk.DepositList) {
	err := keeper.cdc.UnmarshalBinaryBare(bz, &dls)
	if err != nil {
		panic(err)
	}
	return
}

func (keeper Keeper) NewCUIBCAssetWithAddress(ctx sdk.Context, cuType sdk.CUType, address sdk.CUAddress) exported.CUIBCAsset {
	cuAst := keeper.proto()
	err := cuAst.SetAddress(address)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}

	cuAst.SetCUType(cuType)
	cuAst.SetMigrationStatus(sdk.MigrationFinish)

	return cuAst
}

// GetOpCUsAstInfo returns all operation custodian units and depositList of the symbol.
// if symbol empty ,return all operation custodian units and depositList.
func (keeper Keeper) GetOpCUsAstInfo(ctx sdk.Context, symbol string) []sdk.OpCUAstInfo {
	cus := keeper.ck.GetOpCUs(ctx, symbol)
	cusInfo := make([]sdk.OpCUAstInfo, len(cus))

	for i, cu := range cus {
		cusymbol := cu.GetSymbol()
		cuAst := keeper.GetCUIBCAsset(ctx, cu.GetAddress())
		ti := keeper.tk.GetIBCToken(ctx, sdk.Symbol(cusymbol))
		chain := ti.Chain.String()
		cusInfo[i].Symbol = cusymbol
		cusInfo[i].Amount = cuAst.GetAssetCoins().AmountOf(cusymbol)
		cusInfo[i].CuAddress = cu.GetAddress().String()
		cusInfo[i].MultisignAddress = cuAst.GetAssetAddress(cusymbol, cuAst.GetAssetPubkeyEpoch())
		cusInfo[i].LastEpochMultisignAddress = cuAst.GetAssetAddress(cusymbol, cuAst.GetAssetPubkeyEpoch()-1)
		sendEnable := cuAst.IsEnabledSendTx(chain, cusInfo[i].MultisignAddress)
		cusInfo[i].Locked = !sendEnable
		cusInfo[i].GasUsed = cuAst.GetGasUsed().AmountOf(chain)
		cusInfo[i].GasReceived = cuAst.GetGasReceived().AmountOf(chain)
		cusInfo[i].MainNetAmount = cuAst.GetAssetCoins().AmountOf(chain)
		cusInfo[i].MigrationStatus = cuAst.GetMigrationStatus()

		if ti.TokenType == sdk.UtxoBased {
			cusInfo[i].DepositList = keeper.GetDepositList(ctx, cusymbol, cu.GetAddress())
		}
	}
	return cusInfo
}

func (keeper Keeper) startMigrationForAllOpcus(ctx sdk.Context, epoch sdk.Epoch) {
	opCUs := keeper.ck.GetOpCUs(ctx, "")
	if len(opCUs) == 0 {
		epoch.MigrationFinished = true
		return
	}

	for _, opCU := range opCUs {
		opCUAst := keeper.GetCUIBCAsset(ctx, opCU.GetAddress())
		if opCUAst != nil {
			opCUAst.SetMigrationStatus(sdk.MigrationBegin)
			keeper.SetCUIBCAsset(ctx, opCUAst)
		}
	}
}
