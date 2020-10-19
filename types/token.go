package types

import (
	"fmt"
)

//TokenType
type TokenType uint64

const (
	UtxoBased          TokenType = 0x1 //UtxoBased   i.e. token likes BTC
	AccountBased       TokenType = 0x2 //AccountBased i.e. token likes ETH, every user has its own account
	AccountSharedBased TokenType = 0x3 //Memo Based, i.e. token  like EOS/XRP, users share a account with differen memos/tags
)

const (
	NativeToken        = "hbc"
	NativeTokenDecimal = 18

	NativeDefiToken        = "hdt"
	NativeDefiTokenDecimal = 8

	GasPriceBucketWindow uint64 = 10
)

var (
	KeyIsSendEnabled       = "is_send_enabled"
	KeyIsDepositEnabled    = "is_deposit_enabled"
	KeyIsWithdrawalEnabled = "is_withdrawal_enabled"
	KeyCollectThreshold    = "collect_threshold"
	KeyDepositThreshold    = "deposit_threshold"
	KeyOpenFee             = "open_fee"
	KeySysOpenFee          = "sys_open_fee"
	KeyWithdrawalFeeRate   = "withdrawal_fee_rate"
	KeyMaxOpCUNumber       = "max_op_cu_number"
	KeySysTransferNum      = "systransfer_num"
	KeyOpCUSysTransferNum  = "op_cu_systransfer_num"
	KeyGasLimit            = "gas_limit"
	KeyConfirmations       = "confirmations"
)

type TokensGasPrice struct {
	Chain    string `json:"chain" yaml:"chain"`
	GasPrice Int    `json:"gas_price" yaml:"gas_price"`
}

//TokenInfo defines information in token module
type TokenInfo struct {
	Symbol              Symbol    `json:"symbol" yaml:"symbol"`
	Issuer              string    `json:"issuer" yaml:"issuer"`                                 //token's issuer
	Chain               Symbol    `json:"chain" yaml:"chain"`                                   //related mainnet token, e.g. ERC20 token's Chain is ETH
	TokenType           TokenType `json:"type" yaml:"type"`                                     //token's type
	IsSendEnabled       bool      `json:"is_send_enabled" yaml:"is_send_enabled"`               //whether send enabled or not
	IsDepositEnabled    bool      `json:"is_deposit_enabled" yaml:"is_deposit_enabled"`         //whether send enabled or not
	IsWithdrawalEnabled bool      `json:"is_withdrawal_enabled" yaml:"is_withdrawal_enabled"`   //whether withdrawal enabled or not
	Decimals            uint64    `json:"decimals" yaml:"decimals"`                             //token's decimals, represents by the decimals's
	TotalSupply         Int       `json:"total_supply" yaml:"total_supply" `                    //token's total supply
	CollectThreshold    Int       `json:"collect_threshold" yaml:"collect_threshold" `          // token's collect threshold == account threshold
	DepositThreshold    Int       `json:"deposit_threshold" yaml:"deposit_threshold"`           // token's deposit threshold
	OpenFee             Int       `json:"open_fee" yaml:"open_fee"`                             // token's open fee for custodianunit address
	SysOpenFee          Int       `json:"sys_open_fee" yaml:"sys_open_fee"`                     // token's open fee for external address
	WithdrawalFeeRate   Dec       `json:"withdrawal_fee_rate" yaml:"withdrawal_fee_rate"`       // token's WithdrawalFeeRate
	MaxOpCUNumber       uint64    `json:"max_op_cu_number" yaml:"max_op_cu_number"`             // token's opcu num
	SysTransferNum      Int       `json:"sys_transfer_num" yaml:"sys_transfer_num"`             // 给user反向打币每次限额
	OpCUSysTransferNum  Int       `json:"op_cu_sys_transfer_num" yaml:"op_cu_sys_transfer_num"` // 给 opcu之间转gas的每次限额
	GasLimit            Int       `json:"gas_limit" yaml:"gas_limit"`
	GasPrice            Int       `json:"gas_price" yaml:"gas_price"`
	Confirmations       uint64    `json:"confirmations" yaml:"confirmations"` //confirmation of chain
	IsNonceBased        bool      `json:"is_nonce_based" yaml:"is_nonce_based"`
}

//NewTokenInfo create a TokenInfo
func NewTokenInfo(symbol, chain Symbol, issuer string, tokenType TokenType, isSendEnabled, isDepositEnbled, isWithdrawalEnabled bool,
	decimals uint64, totalSupply, collectThreshold, depositThreshold, openFee, sysOpenFee Int, withdrawalFeeRate Dec, sysTransferNum,
	opCUSysTransferNum, gasLimit Int, gasPrice Int, maxOpCUNumber, confirmations uint64, isNonceBased bool) *TokenInfo {
	return &TokenInfo{
		Symbol:              symbol,
		Issuer:              issuer,
		Chain:               chain,
		TokenType:           tokenType,
		IsSendEnabled:       isSendEnabled,
		IsDepositEnabled:    isDepositEnbled,
		IsWithdrawalEnabled: isWithdrawalEnabled,
		Decimals:            decimals,
		TotalSupply:         totalSupply,
		CollectThreshold:    collectThreshold,
		DepositThreshold:    depositThreshold,
		OpenFee:             openFee,
		SysOpenFee:          sysOpenFee,
		WithdrawalFeeRate:   withdrawalFeeRate,
		MaxOpCUNumber:       maxOpCUNumber,
		SysTransferNum:      sysTransferNum,
		OpCUSysTransferNum:  opCUSysTransferNum,
		GasLimit:            gasLimit,
		GasPrice:            gasPrice,
		Confirmations:       confirmations,
		IsNonceBased:        isNonceBased,
	}
}

func IsTokenTypeLegal(tokenType TokenType) bool {
	if tokenType >= UtxoBased && tokenType <= AccountSharedBased {
		return true
	}
	return false
}

func (t TokenInfo) String() string {
	return fmt.Sprintf(`
	Symbol:%s
	Issuer:%v
	Chain:%v
	TokenType:%v
	IsSendEnabled:%v
	IsDepositEnabled:%v
	IsWithdrawalEnabled:%v
	Decimals:%v
	TotalSupply:%v
	CollectThreshold:%v
	DepositThreshold:%v
	OpenFee:%v
	SysOpenFee:%v
	WithdrawalFee:%v
	MaxOpCUNumber:%v
	SysTransferNum:%v
	OpCUSysTransferNum:%v
	GasLimit:%v
	GasPrice:%v
	Confirmations:%v
	IsNonceBased:%v
	`, t.Symbol, t.Issuer, t.Chain, t.TokenType, t.IsSendEnabled, t.IsDepositEnabled,
		t.IsWithdrawalEnabled, t.Decimals, t.TotalSupply, t.CollectThreshold, t.DepositThreshold,
		t.OpenFee, t.SysOpenFee, t.WithdrawalFeeRate, t.MaxOpCUNumber, t.SysTransferNum,
		t.OpCUSysTransferNum, t.GasLimit, t.GasPrice, t.Confirmations, t.IsNonceBased)
}

func (t TokenInfo) IsValid() bool {
	if !(t.Symbol.IsValidTokenName() && t.Chain.IsValidTokenName()) {
		return false
	}

	if !t.TokenType.IsValid() {
		return false
	}

	if t.Decimals > Precision {
		return false
	}

	if !(t.CollectThreshold.IsPositive() && t.DepositThreshold.IsPositive() && t.OpenFee.IsPositive() &&
		t.SysOpenFee.IsPositive() && t.WithdrawalFeeRate.IsPositive()) {
		return false
	}

	if t.GasLimit.IsNegative() || t.GasPrice.IsNegative() || t.TotalSupply.IsNegative() {
		return false
	}

	if t.MaxOpCUNumber == 0 {
		return false
	}

	return true
}

func (t *TokenInfo) SysTransferAmount() Int {
	return t.GasPrice.Mul(t.GasLimit).Mul(t.SysTransferNum)
}

func (t *TokenInfo) OpCUSysTransferAmount() Int {
	return t.GasPrice.Mul(t.GasLimit).Mul(t.OpCUSysTransferNum)
}

func (t *TokenInfo) WithDrawalFee() Coin {
	baseGasFee := t.GasPrice.Mul(t.GasLimit)
	withdrawalFeeAmt := t.WithdrawalFeeRate.Mul(NewDecFromInt(baseGasFee)).TruncateInt()
	return NewCoin(t.Chain.String(), withdrawalFeeAmt)
}

func (t *TokenInfo) CollectFee() Coin {
	baseGasFee := t.GasPrice.Mul(t.GasLimit)
	if t.Chain.String() != t.Symbol.String() {
		collectFeeAmount := t.SysTransferAmount().Add(baseGasFee)
		return NewCoin(t.Chain.String(), collectFeeAmount)
	} else {
		return NewCoin(t.Chain.String(), baseGasFee)
	}
}

type Symbol string

// IsValidTokenName check token name.
// a valid token name must be a valid coin denom
func (s Symbol) IsValidTokenName() bool {
	// same as coin
	if reDnm.MatchString(string(s)) {
		return true
	}
	return false
}

func (s Symbol) ToDenomName() string {
	if s.IsValidTokenName() {
		return string(s)
	}
	return ""
}

func (s Symbol) String() string {
	if s.IsValidTokenName() {
		return string(s)
	}
	return ""
}

func (t TokenType) IsValid() bool {
	if t == UtxoBased || t == AccountBased || t == AccountSharedBased {
		return true
	}
	return false
}

func (gp TokensGasPrice) String() string {
	return fmt.Sprintf(
		`Chain:%s
		"GasPrice:%v`,
		gp.Chain, gp.GasPrice.String())
}
