package types

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	cuexported "github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

// StakingKeeper defines the expected staking keeper (noalias)
type StakingKeeper interface {
	ApplyAndReturnValidatorSetUpdates(sdk.Context) (updates []abci.ValidatorUpdate)
}

// CUKeeper defines the expected CustodianUnit keeper (noalias)
type CUKeeper interface {
	NewCU(sdk.Context, cuexported.CustodianUnit) cuexported.CustodianUnit
	SetCU(sdk.Context, cuexported.CustodianUnit)
	IterateCUs(ctx sdk.Context, process func(cuexported.CustodianUnit) (stop bool))
}

// GenesisCUsIterator defines the expected iterating genesis accounts object (noalias)
type GenesisCUsIterator interface {
	IterateGenesisCUs(
		cdc *codec.Codec,
		appGenesis map[string]json.RawMessage,
		iterateFn func(cuexported.CustodianUnit) (stop bool),
	)
}
