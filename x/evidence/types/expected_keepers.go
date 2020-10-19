package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
	GetWithSubkey(ctx sdk.Context, key, subkey []byte, ptr interface{})
	GetParamSetWithSubkey(ctx sdk.Context, key []byte, ps params.ParamSet)
	SetParamSetWithSubkey(ctx sdk.Context, key []byte, ps params.ParamSet)
}

type StakingKeeper interface {
	SlashByOperator(sdk.Context, sdk.ValAddress, int64, sdk.Dec)
	JailByOperator(ctx sdk.Context, operator sdk.ValAddress)
	GetEpochByHeight(ctx sdk.Context, height uint64) sdk.Epoch
	GetCurrentEpoch(ctx sdk.Context) sdk.Epoch
}
