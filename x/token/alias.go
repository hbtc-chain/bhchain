package token

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/client"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

const (
	ModuleName            = types.ModuleName
	RouterKey             = types.RouterKey
	StoreKey              = types.StoreKey
	QuerierRoute          = types.QuerierRoute
	QueryTokens           = types.QueryTokens
	QueryToken            = types.QueryToken
	QuerySymbols          = types.QuerySymbols
	QueryParameters       = types.QueryParameters
	QueryDecimal          = types.QueryDecimal
	QuerierKey            = types.QuerierRoute
	DefaultParamspace     = types.DefaultParamspace
	DefaultTokenCacheSize = types.DefaultTokenCacheSize
)

type (
	QueryDecimals    = types.QueryDecimals
	QueryTokenInfo   = types.QueryTokenInfo
	QueryResToken    = types.QueryResToken
	QueryResDecimals = types.QueryResDecimals
	QueryResSymbols  = types.QueryResSymbols
	QueryResTokens   = types.QueryResTokens
	Params           = types.Params
)

var (
	NewTokenInfo                     = sdk.NewTokenInfo
	ModuleCdc                        = types.ModuleCdc
	RegisterCodec                    = types.RegisterCodec
	IsTokenTypeLegal                 = sdk.IsTokenTypeLegal
	DefaultParams                    = types.DefaultParams
	KeyTokenCacheSize                = types.KeyTokenCacheSize
	KeyReservedSymbols               = types.KeyReservedSymbols
	DisableTokenProposalHandler      = client.DisableTokenProposalHandler
	AddTokenProposalHandler          = client.AddTokenProposalHandler
	TokenParamsChangeProposalHandler = client.TokenParamsChangeProposalHandler
	NewAddTokenProposal              = types.NewAddTokenProposal
	NewTokenParamsChangeProposal     = types.NewTokenParamsChangeProposal
	NewDisableTokenProposal          = types.NewDisableTokenProposal
)
