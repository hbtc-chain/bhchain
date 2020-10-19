/*
 * *******************************************************************
 * @项目名称: types
 * @文件名称: querier.go
 * @Date: 2019/06/05
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */
package types

import (
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type QueryTokenInfo struct {
	Symbol string `json:"symbol"`
}

func NewQueryTokenInfo(symbol string) QueryTokenInfo {
	return QueryTokenInfo{Symbol: symbol}
}

//type QueryResToken types.TokenInfo
type QueryResToken struct {
	Symbol              sdk.Symbol `json:"symbol"`
	Issuer              string     `json:"issuer"`                //token's issuer
	Chain               string     `json:"chain"`                 //related mainnet token, e.g. ERC20 token's Chain is ETH
	TokenType           uint64     `json:"type"`                  //token's type
	IsSendEnabled       bool       `json:"is_send_enabled"`       //whether send enabled or not
	IsDepositEnabled    bool       `json:"is_deposit_enabled"`    //whether send enabled or not
	IsWithdrawalEnabled bool       `json:"is_withdrawal_enabled"` //whether withdrawal enabled or not
	Decimals            uint64     `json:"decimals"`              //token's decimals, represents by the decimals's
	TotalSupply         sdk.Int    `json:"total_supply"`          //token's total supply
	CollectThreshold    sdk.Int    `json:"collect_threshold"`     // token's collect threshold == account threshold
	DepositThreshold    sdk.Int    `json:"deposit_threshold"`     // token's deposit threshold
	OpenFee             sdk.Int    `json:"open_fee"`              // token's open fee for custodianunit address
	SysOpenFee          sdk.Int    `json:"sys_open_fee"`          // token's open fee for external address
	WithdrawalFeeRate   sdk.Dec    `json:"withdrawal_fee_rate"`   // token's WithdrawalFeeRate
	MaxOpCUNumber       uint64     `json:"max_op_cu_number"`
	SysTransferNum      sdk.Int    `json:"sys_transfer_num"`       // 给user反向打币每次限额
	OpCUSysTransferNum  sdk.Int    `json:"op_cu_sys_transfer_num"` // 给 opcu之间转gas的每次限额
	GasLimit            sdk.Int    `json:"gas_limit"`
	GasPrice            sdk.Int    `json:"gas_price"`
	Confirmations       uint64     `json:"confirmations" yaml:"confirmations"` //confirmation of chain
	IsNonceBased        bool       `json:"is_nonce_based" yaml:"is_nonce_based"`
}

func (r QueryResToken) String() string {
	return fmt.Sprintf(`
	Symbol:%s
	Issuer:%v
	Chain:%v
	TokenType:%v
	IsSendEnabled:%v
	IsDepositEnabled:%v
	IsWithdrawalEnabled:%v
	Decimals:%v
	CollectThreshold:%v
	DepositThreshold:%v
	OpenFee:%v
	SysOpenFee:%v
	WithdrawalFeeRate:%v
	MaxOpCUNumber:%v,
	SysTransferNum:%v,
	OpCUSysTransferNum:%v,
	GasLimit:%v,
    GasPrice:%v,
    Confirmations:%v,
    IsNonceBased:%v,
	`, r.Symbol, r.Issuer, r.Chain, r.TokenType, r.IsSendEnabled, r.IsDepositEnabled,
		r.IsWithdrawalEnabled, r.Decimals, r.CollectThreshold, r.DepositThreshold,
		r.OpenFee, r.SysOpenFee, r.WithdrawalFeeRate, r.MaxOpCUNumber, r.SysTransferNum,
		r.OpCUSysTransferNum, r.GasLimit, r.GasPrice, r.Confirmations, r.IsNonceBased)
}

type QueryResTokens []QueryResToken

func (qs QueryResTokens) String() string {
	if len(qs) == 0 {
		return ""
	}

	out := ""
	for _, token := range qs {
		out += fmt.Sprintf("%v,", token.String())
	}
	return out[:len(out)-1]
}

type QueryResSymbols []string

func (r QueryResSymbols) String() string {
	return strings.Join(r[:], ",")
}

type QueryDecimals struct {
	Symbol string `json:"symbol"`
}

//type QueryResToken types.TokenInfo
type QueryResDecimals struct {
	Decimals uint64 `json:"decimals"` //token's decimals, represents by the decimals's
}

func (r QueryResDecimals) String() string {
	return fmt.Sprintf("decimal:%v", r.Decimals)
}
