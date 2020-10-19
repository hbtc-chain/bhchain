package hrc20

import (
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc20/internal"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

type Keeper struct {
	storeKey      sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc           *codec.Codec // The wire codec for binary encoding/decoding
	tk            internal.TokenKeeper
	ck            internal.CustodianUnitKeeper
	dk            internal.DistributionKeeper
	sk            internal.SupplyKeeper
	rk            internal.ReceiptKeeper
	paramSubSpace params.Subspace
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSubSpace params.Subspace, tk internal.TokenKeeper,
	ck internal.CustodianUnitKeeper, dk internal.DistributionKeeper, sk internal.SupplyKeeper, rk internal.ReceiptKeeper) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		tk:            tk,
		ck:            ck,
		dk:            dk,
		sk:            sk,
		rk:            rk,
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
