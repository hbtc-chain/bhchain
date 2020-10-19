package upgrade

import (
	"github.com/hbtc-chain/bhchain/x/upgrade/keeper"
	"github.com/hbtc-chain/bhchain/x/upgrade/types"
)

const (
	ModuleName                        = types.ModuleName
	RouterKey                         = types.RouterKey
	StoreKey                          = types.StoreKey
	QuerierKey                        = types.QuerierKey
	PlanByte                          = types.PlanByte
	DoneByte                          = types.DoneByte
	ProposalTypeSoftwareUpgrade       = types.ProposalTypeSoftwareUpgrade
	ProposalTypeCancelSoftwareUpgrade = types.ProposalTypeCancelSoftwareUpgrade
	QueryCurrent                      = types.QueryCurrent
	QueryApplied                      = types.QueryApplied
)

var (
	// functions aliases
	RegisterCodec                    = types.RegisterCodec
	PlanKey                          = types.PlanKey
	NewSoftwareUpgradeProposal       = types.NewSoftwareUpgradeProposal
	NewCancelSoftwareUpgradeProposal = types.NewCancelSoftwareUpgradeProposal
	NewQueryAppliedParams            = types.NewQueryAppliedParams
	NewKeeper                        = keeper.NewKeeper
	NewQuerier                       = keeper.NewQuerier
)

type (
	UpgradeHandler                = types.UpgradeHandler // nolint
	Plan                          = types.Plan
	SoftwareUpgradeProposal       = types.SoftwareUpgradeProposal
	CancelSoftwareUpgradeProposal = types.CancelSoftwareUpgradeProposal
	UpgradeInfo                   = types.UpgradeInfo
	QueryAppliedParams            = types.QueryAppliedParams
	Keeper                        = keeper.Keeper
)
