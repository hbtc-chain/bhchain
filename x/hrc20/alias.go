package hrc20

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc20/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = types.DefaultParamspace
	DefaultCodespace  = types.DefaultCodespace
	QueryParameters   = types.QueryParameters
)

type (
	Params      = types.Params
	MsgNewToken = types.MsgNewToken
)

var (
	NewTokenInfo         = sdk.NewTokenInfo
	ModuleCdc            = types.ModuleCdc
	RegisterCodec        = types.RegisterCodec
	DefaultParams        = types.DefaultParams
	DefaultIssueTokenFee = types.DefaultIssueTokenFee
	KeyIssueTokenFee     = types.KeyIssueTokenFee
)
