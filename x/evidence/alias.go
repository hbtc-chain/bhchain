package evidence

import (
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
	"github.com/hbtc-chain/bhchain/x/evidence/keeper"
	"github.com/hbtc-chain/bhchain/x/evidence/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
	QuerierRoute      = types.QuerierRoute
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	RegisterCodec       = types.RegisterCodec
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState

	// variable aliases
	ModuleCdc         = types.ModuleCdc
	DsignBehaviourKey = types.DsignBehaviourKey
)

type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
	Vote         = exported.Vote
	VoteItem     = exported.VoteItem
	VoteBox      = exported.VoteBox
)
