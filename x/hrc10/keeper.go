package hrc10

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc10/internal"
	"github.com/hbtc-chain/bhchain/x/hrc10/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

type Keeper struct {
	storeKey      sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc           *codec.Codec // The wire codec for binary encoding/decoding
	tk            internal.TokenKeeper
	dk            internal.DistributionKeeper
	sk            internal.SupplyKeeper
	rk            internal.ReceiptKeeper
	trk           internal.TransferKeeper
	paramSubSpace params.Subspace
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSubSpace params.Subspace, tk internal.TokenKeeper, dk internal.DistributionKeeper, sk internal.SupplyKeeper, rk internal.ReceiptKeeper, trk internal.TransferKeeper) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		tk:            tk,
		dk:            dk,
		sk:            sk,
		rk:            rk,
		trk:           trk,
		paramSubSpace: paramSubSpace.WithKeyTable(types.ParamKeyTable()),
	}
}

// SetParams sets the token module's parameters.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubSpace.SetParamSet(ctx, &params)
}

// GetParams gets the token module's parameters.
func (k *Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSubSpace.GetParamSet(ctx, &params)
	return
}
