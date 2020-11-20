package token

import (
	"github.com/hbtc-chain/bhchain/x/token/client"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

const (
	ModuleName       = types.ModuleName
	RouterKey        = types.RouterKey
	StoreKey         = types.StoreKey
	QuerierRoute     = types.QuerierRoute
	QuerierKey       = types.QuerierRoute
	DefaultCodespace = types.DefaultCodespace

	QueryToken     = types.QueryToken
	QueryIBCTokens = types.QueryIBCTokens
)

type (
	QueryTokenInfoParams = types.QueryTokenInfoParams
	ResToken             = types.ResToken
)

var (
	ModuleCdc                        = types.ModuleCdc
	RegisterCodec                    = types.RegisterCodec
	AddTokenProposalHandler          = client.AddTokenProposalHandler
	TokenParamsChangeProposalHandler = client.TokenParamsChangeProposalHandler
	NewAddTokenProposal              = types.NewAddTokenProposal
	NewTokenParamsChangeProposal     = types.NewTokenParamsChangeProposal
)
