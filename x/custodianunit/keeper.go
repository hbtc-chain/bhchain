package custodianunit

import (
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/hbtc-chain/bhchain/x/custodianunit/internal"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/params/subspace"
)

type CUKeeperI interface {
	// CU functions
	NewCUWithPubkey(ctx sdk.Context, cuType sdk.CUType, pub crypto.PubKey) exported.CustodianUnit

	NewCUWithAddress(ctx sdk.Context, cuType sdk.CUType, cuaddr sdk.CUAddress) exported.CustodianUnit

	NewCU(ctx sdk.Context, cu exported.CustodianUnit) exported.CustodianUnit

	NewOpCUWithAddress(ctx sdk.Context, symbol string, addr sdk.CUAddress) exported.CustodianUnit

	GetCU(context sdk.Context, addresses sdk.CUAddress) exported.CustodianUnit

	GetOrNewCU(context sdk.Context, cuType sdk.CUType, addresses sdk.CUAddress) exported.CustodianUnit

	SetCU(ctx sdk.Context, cu exported.CustodianUnit)

	GetAllCUs(ctx sdk.Context) []exported.CustodianUnit

	GetOpCUs(ctx sdk.Context, symbol string) []exported.CustodianUnit

	GetOpCUsInfo(ctx sdk.Context, symbol string) []sdk.OpCUInfo

	RemoveCU(ctx sdk.Context, cu exported.CustodianUnit)

	IterateCUs(ctx sdk.Context, process func(exported.CustodianUnit) (stop bool))

	GetNextCUNumber(ctx sdk.Context) uint64

	// param funcsions
	SetParams(ctx sdk.Context, params types.Params)

	GetParams(ctx sdk.Context) (params types.Params)

	Logger(ctx sdk.Context) log.Logger

	// deposit functions
	GetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress) sdk.DepositList

	GetDepositListByHash(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string) sdk.DepositList

	SetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress, list sdk.DepositList)

	SaveDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, deposit sdk.DepositItem) error

	DelDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64)

	SetDepositStatus(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64, stautus sdk.DepositItemStatus) error

	GetDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) sdk.DepositItem

	IsDepositExist(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) bool

	SetExtAddresseWithCU(ctx sdk.Context, symbol, extAddress string, cuAddress sdk.CUAddress)

	GetCUFromExtAddress(ctx sdk.Context, symbol, extAddress string) (sdk.CUAddress, error)
}

var _ CUKeeperI = (*CUKeeper)(nil)

// CUKeeper encodes/decodes CUs using the go-amino (binary)
// encoding/decoding library.
type CUKeeper struct {
	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical CU constructor.
	proto func() exported.CustodianUnit

	tk internal.TokenKeeper
	sk internal.StakingKeeper

	// The codec codec for binary encoding/decoding of CUs.
	cdc *codec.Codec

	ParamSubspace subspace.Subspace
}

func (ck *CUKeeper) SetStakingKeeper(sk internal.StakingKeeper) {
	ck.sk = sk
}

// NewCUKeeper returns a new CUKeeper that uses go-amino to
// (binary) encode and decode concrete sdk.CustodianUnits.
// nolint
func NewCUKeeper(
	cdc *codec.Codec, key sdk.StoreKey, tk internal.TokenKeeper, paramstore subspace.Subspace, proto func() exported.CustodianUnit,
) CUKeeper {
	return CUKeeper{
		key:           key,
		proto:         proto,
		cdc:           cdc,
		tk:            tk,
		ParamSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
}

// Logger returns a module-specific logger.
func (ck CUKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

//------implement CUKeeperI---
func (ck CUKeeper) NewCUWithPubkey(ctx sdk.Context, cuType sdk.CUType, pub crypto.PubKey) exported.CustodianUnit {
	cu := ck.newCUWithAddress(ctx, sdk.CUAddressFromPubKey(pub), cuType)
	cu.SetPubKey(pub)
	return cu
}

func (ck CUKeeper) NewCUWithAddress(ctx sdk.Context, cuType sdk.CUType, addr sdk.CUAddress) exported.CustodianUnit {
	return ck.newCUWithAddress(ctx, addr, cuType)
}

// NewCU set CUNumber for the CustodianUnit
// the CustodianUnit may not created by the proto func, for example: ModuleAccountI
func (ck CUKeeper) NewCU(ctx sdk.Context, cu exported.CustodianUnit) exported.CustodianUnit {
	return cu
}

// NewOpCUWithAddress create OP CU
func (ck CUKeeper) NewOpCUWithAddress(ctx sdk.Context, symbol string, addr sdk.CUAddress) exported.CustodianUnit {
	if symbol == "" || addr == nil || !ck.tk.IsTokenSupported(ctx, sdk.Symbol(symbol)) {
		return nil
	}
	if c := ck.GetCU(ctx, addr); c != nil {
		ctx.Logger().Error("NewOpCUWithAddress cu already used", "cu", c.GetAddress().String())
		return nil
	}
	opcu := ck.newCUWithAddress(ctx, addr, sdk.CUTypeOp)

	opcu.AddAsset(symbol, "", 0)
	return opcu
}

// GetOrNewCU get or new a CustodianUnit.
// If CustodianUnit is not found, and CU type is user, new one but not yet save it.
// If CustodianUnit is not found, and CU type not user, return CustodianUnit(nil).
func (ck CUKeeper) GetOrNewCU(context sdk.Context, cuType sdk.CUType, addr sdk.CUAddress) exported.CustodianUnit {
	if addr == nil {
		return nil
	}
	cu := ck.getCU(context, addr)
	if cu == nil && cuType == sdk.CUTypeUser {
		newCU := ck.newCUWithAddress(context, addr, cuType)
		return newCU
	}
	return cu
}

// SetCU implements sdk.CUKeeper.
func (ck CUKeeper) SetCU(ctx sdk.Context, cu exported.CustodianUnit) {
	addr := cu.GetAddress()
	store := ctx.KVStore(ck.key)

	bz, err := ck.cdc.MarshalBinaryBare(cu)
	if err != nil {
		panic(err)
	}
	if cu.GetCUType() == sdk.CUTypeOp {
		if err = ck.opCUCheck(ctx, cu); err != nil {
			panic(err)
		}
		ck.setOpAddress(ctx, cu.GetSymbol(), cu.GetAddress())
	}
	store.Set(types.AddressStoreKey(addr), bz)
}

func (ck CUKeeper) opCUCheck(ctx sdk.Context, cu exported.CustodianUnit) error {
	if cu == nil || cu.GetCUType() != sdk.CUTypeOp {
		return errors.New("not a Op CU")
	}
	if cu.GetSymbol() == "" || cu.GetAddress() == nil {
		return errors.New("symbol or address of custodianunit is nil")
	}

	return nil
}

func (ck CUKeeper) GetCU(context sdk.Context, addresses sdk.CUAddress) exported.CustodianUnit {
	cu := ck.getCU(context, sdk.CUAddress(addresses))
	//  none exist CU return exported.CustodianUnit(nil)
	if cu == nil {
		return nil
	}
	return cu
}

// GetOpCUs returns all operation custodian units of the symbol .
// if symbol is empty return all operation custodian units of the symbol.
func (ck CUKeeper) GetOpCUs(ctx sdk.Context, symbol string) []exported.CustodianUnit {
	addresses := ck.getOpAddresses(ctx, symbol)
	CUs := []exported.CustodianUnit{}
	for _, a := range addresses {
		CUs = append(CUs, ck.getCU(ctx, a))
	}
	return CUs
}

// GetOpCUsInfo returns all operation custodian units and depositList of the symbol.
// if symbol empty ,return all operation custodian units and depositList.
func (ck CUKeeper) GetOpCUsInfo(ctx sdk.Context, symbol string) []sdk.OpCUInfo {
	cus := ck.GetOpCUs(ctx, symbol)
	cusInfo := make([]sdk.OpCUInfo, len(cus))

	for i, cu := range cus {
		cusymbol := cu.GetSymbol()
		chain := ck.tk.GetChain(ctx, sdk.Symbol(cusymbol)).String()
		cusInfo[i].Symbol = cusymbol
		cusInfo[i].Amount = cu.GetAssetCoins().AmountOf(cusymbol)
		cusInfo[i].CuAddress = cu.GetAddress().String()
		cusInfo[i].MultisignAddress = cu.GetAssetAddress(cusymbol, cu.GetAssetPubkeyEpoch())
		cusInfo[i].LastEpochMultisignAddress = cu.GetAssetAddress(cusymbol, cu.GetAssetPubkeyEpoch()-1)
		sendEnable := cu.IsEnabledSendTx(chain, cusInfo[i].MultisignAddress)
		cusInfo[i].Locked = !sendEnable
		cusInfo[i].GasUsed = cu.GetGasUsed().AmountOf(chain)
		cusInfo[i].GasReceived = cu.GetGasReceived().AmountOf(chain)
		cusInfo[i].MainNetAmount = cu.GetAssetCoins().AmountOf(chain)
		cusInfo[i].MigrationStatus = cu.GetMigrationStatus()

		if ck.tk.IsUtxoBased(ctx, sdk.Symbol(cusymbol)) {
			cusInfo[i].DepositList = ck.GetDepositList(ctx, cusymbol, cu.GetAddress())
		}
	}
	return cusInfo
}

func (ck CUKeeper) startMigrationForAllOpcus(ctx sdk.Context, epoch sdk.Epoch) {
	opCUs := ck.GetOpCUs(ctx, "")
	if len(opCUs) == 0 {
		epoch.MigrationFinished = true
		ck.sk.SetEpoch(ctx, epoch)
		return
	}
	for _, opCU := range opCUs {
		opCU.SetMigrationStatus(sdk.MigrationBegin)
		ck.SetCU(ctx, opCU)
	}
}

// GetAllCUs returns all custodian units in the CUKeeper.
func (ck CUKeeper) GetAllCUs(ctx sdk.Context) []exported.CustodianUnit {
	CUs := []exported.CustodianUnit{}
	appendCU := func(cu exported.CustodianUnit) (stop bool) {
		CUs = append(CUs, cu)
		return false
	}
	ck.IterateCUs(ctx, appendCU)
	return CUs
}

// RemoveCU removes an cu for the cu mapper store.
// NOTE: this will cause supply invariant violation if called
func (ck CUKeeper) RemoveCU(ctx sdk.Context, cu exported.CustodianUnit) {
	addr := cu.GetAddress()
	store := ctx.KVStore(ck.key)
	store.Delete(types.AddressStoreKey(addr))
}

// IterateCUs implements sdk.CUKeeper.
func (ck CUKeeper) IterateCUs(ctx sdk.Context, process func(exported.CustodianUnit) (stop bool)) {
	store := ctx.KVStore(ck.key)
	iter := sdk.KVStorePrefixIterator(store, types.AddressStoreKeyPrefix)
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		cu := ck.decodeCU(val)
		if process(cu) {
			return
		}
		iter.Next()
	}
}

// GetNextCUNumber Returns and increments the global CU number counter
func (ck CUKeeper) GetNextCUNumber(ctx sdk.Context) uint64 {
	return 0
}

func (ck CUKeeper) GetPendingDepositList(ctx sdk.Context, address sdk.CUAddress) map[string]sdk.DepositList {
	store := ctx.KVStore(ck.key)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositStorePrefixKeyWithAddr(address))
	defer iterator.Close()

	deposits := make(map[string]sdk.DepositList)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dl sdk.DepositItem
		ck.cdc.UnmarshalBinaryBare(bz, &dl)
		if dl.Status == sdk.DepositItemStatusUnCollected {
			token := types.DecodeSymbolFromDepositListKey(iterator.Key())
			deposits[token] = append(deposits[token], dl)
		}
	}
	return deposits

}

// ----------------------------------------------------------------------
// DepositList funcs
func (ck CUKeeper) GetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress) sdk.DepositList {
	store := ctx.KVStore(ck.key)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositStorePrefixKey(symbol, address))
	defer iterator.Close()

	var dls sdk.DepositList
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dl sdk.DepositItem
		ck.cdc.UnmarshalBinaryBare(bz, &dl)
		dls.AddDepositItem(dl)
	}
	return dls
}

func (ck CUKeeper) GetDepositListByHash(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string) sdk.DepositList {
	dls := ck.GetDepositList(ctx, symbol, address)
	dlsGot := dls.Filter(func(d sdk.DepositItem) bool {
		return d.GetHash() == hash
	})
	return dlsGot
}

func (ck CUKeeper) SetDepositList(ctx sdk.Context, symbol string, address sdk.CUAddress, list sdk.DepositList) {
	for _, item := range list {
		ck.SaveDeposit(ctx, symbol, address, item)
	}
}

// GetDeposit get deposit item from store
func (ck CUKeeper) GetDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) sdk.DepositItem {
	store := ctx.KVStore(ck.key)
	bz := store.Get(types.DepositStoreKey(symbol, address, hash, index))
	if bz == nil {
		return sdk.DepositNil
	}
	var item sdk.DepositItem
	ck.cdc.MustUnmarshalBinaryBare(bz, &item)

	return item
}

func (ck CUKeeper) IsDepositExist(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) bool {
	store := ctx.KVStore(ck.key)
	return store.Has(types.DepositStoreKey(symbol, address, hash, index))
}

// SaveDeposit save deposit item to store
func (ck CUKeeper) SaveDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, depositItem sdk.DepositItem) error {
	store := ctx.KVStore(ck.key)

	bz, err := ck.cdc.MarshalBinaryBare(depositItem)
	if err != nil {
		return err
	}

	store.Set(types.DepositStoreKey(symbol, address, depositItem.GetHash(), depositItem.GetIndex()), bz)

	return nil
}

// DelDeposit delete deposit item from store
func (ck CUKeeper) DelDeposit(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64) {
	store := ctx.KVStore(ck.key)
	store.Delete(types.DepositStoreKey(symbol, address, hash, index))
}

func (ck CUKeeper) SetDepositStatus(ctx sdk.Context, symbol string, address sdk.CUAddress, hash string, index uint64, stautus sdk.DepositItemStatus) error {
	item := ck.GetDeposit(ctx, symbol, address, hash, index)
	if item == sdk.DepositNil {
		return fmt.Errorf("deposit not exist %v%v", hash, index)
	}

	item.Status = stautus
	ck.SaveDeposit(ctx, symbol, address, item)
	return nil
}

func (ck CUKeeper) GetTokenKeeper(ctx sdk.Context) internal.TokenKeeper {
	return ck.tk
}

// -----------------------------------------------------------------------------
// Params

// SetParams sets the auth module's parameters.
func (ck CUKeeper) SetParams(ctx sdk.Context, params types.Params) {
	ck.ParamSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the auth module's parameters.
func (ck CUKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	ck.ParamSubspace.GetParamSet(ctx, &params)
	return
}

// -----------------------------------------------------------------------------
// Misc.
// TODO test case
// getOpAddresses get all OP CU's address of symbol,
// if symbol is empty ,return all OP CU's address
func (ck CUKeeper) getOpAddresses(ctx sdk.Context, symbol string) []sdk.CUAddress {
	store := ctx.KVStore(ck.key)
	var addresses []sdk.CUAddress

	iter := sdk.KVStorePrefixIterator(store, append(types.OpCUPrefix, []byte(symbol)...))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addresses = append(addresses, types.AddressFromOpCUKey(iter.Key()))
	}

	return addresses
}

// setOpAddress
func (ck CUKeeper) setOpAddress(ctx sdk.Context, symbol string, address sdk.CUAddress) {
	store := ctx.KVStore(ck.key)
	bz, err := ck.cdc.MarshalBinaryBare(address)
	if err != nil {
		panic(err)
	}
	store.Set(types.OpCUKey(symbol, address), bz)
}

func (ck CUKeeper) getCU(ctx sdk.Context, addr sdk.CUAddress) exported.CustodianUnit {
	store := ctx.KVStore(ck.key)
	bz := store.Get(types.AddressStoreKey(addr))
	if bz == nil {
		return nil
	}

	cu := ck.decodeCU(bz)
	return cu
}

func (ck CUKeeper) decodeCU(bz []byte) (cu exported.CustodianUnit) {
	err := ck.cdc.UnmarshalBinaryBare(bz, &cu)
	if err != nil {
		panic(err)
	}
	return
}

func (ck CUKeeper) decodeDeposit(bz []byte) (dls sdk.DepositList) {
	err := ck.cdc.UnmarshalBinaryBare(bz, &dls)
	if err != nil {
		panic(err)
	}
	return
}

func (ck CUKeeper) newCUWithAddress(ctx sdk.Context, address sdk.CUAddress, cuType sdk.CUType) exported.CustodianUnit {
	cu := ck.proto()
	err := cu.SetCUType(cuType)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	err = cu.SetAddress(address)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}

	cu.SetMigrationStatus(sdk.MigrationFinish)
	return cu
}

// ------------------------------------------------
// GetCUFromExtAddresse get the CU's address which has the etxAddress,
func (ck CUKeeper) GetCUFromExtAddress(ctx sdk.Context, chain, extAddress string) (sdk.CUAddress, error) {
	store := ctx.KVStore(ck.key)
	var cuAddress sdk.CUAddress

	if len(chain) == 0 || len(extAddress) == 0 {
		return nil, errors.New(fmt.Sprintf("empty chain:%v or extAddress:%v", chain, extAddress))
	}

	key := types.ExtAddressKey(chain, extAddress)
	bz := store.Get(key)
	if bz == nil {
		return nil, errors.New(fmt.Sprintf("not exists chain:%v add extaddr:%v", chain, extAddress))
	}

	_ = ck.cdc.UnmarshalBinaryBare(bz, &cuAddress)

	return cuAddress, nil
}

// SetExtAddresseWithCU
func (ck CUKeeper) SetExtAddresseWithCU(ctx sdk.Context, chain, extAddress string, cuAddress sdk.CUAddress) {
	store := ctx.KVStore(ck.key)
	bz, err := ck.cdc.MarshalBinaryBare(cuAddress)
	if err != nil {
		panic(err)
	}
	store.Set(types.ExtAddressKey(chain, extAddress), bz)
}
