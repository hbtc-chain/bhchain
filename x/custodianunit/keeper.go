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

	RemoveCU(ctx sdk.Context, cu exported.CustodianUnit)

	IterateCUs(ctx sdk.Context, process func(exported.CustodianUnit) (stop bool))

	// param funcsions
	SetParams(ctx sdk.Context, params types.Params)

	GetParams(ctx sdk.Context) (params types.Params)

	Logger(ctx sdk.Context) log.Logger

	SetExtAddressWithCU(ctx sdk.Context, symbol, extAddress string, cuAddress sdk.CUAddress)

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
	cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, proto func() exported.CustodianUnit,
) CUKeeper {
	return CUKeeper{
		key:           key,
		proto:         proto,
		cdc:           cdc,
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
	if symbol == "" || addr == nil {
		return nil
	}
	if c := ck.GetCU(ctx, addr); c != nil {
		ctx.Logger().Error("NewOpCUWithAddress cu already used", "cu", c.GetAddress().String())
		return nil
	}
	opcu := ck.newCUWithAddress(ctx, addr, sdk.CUTypeOp)

	_ = opcu.SetSymbol(symbol)
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

	var iter sdk.Iterator
	if symbol == "" {
		iter = sdk.KVStorePrefixIterator(store, types.OpCUPrefix)
	} else {
		iter = sdk.KVStorePrefixIterator(store, types.OpCUKeyPrefix(symbol))
	}
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
func (ck CUKeeper) SetExtAddressWithCU(ctx sdk.Context, chain, extAddress string, cuAddress sdk.CUAddress) {
	store := ctx.KVStore(ck.key)
	bz, err := ck.cdc.MarshalBinaryBare(cuAddress)
	if err != nil {
		panic(err)
	}
	store.Set(types.ExtAddressKey(chain, extAddress), bz)
}
