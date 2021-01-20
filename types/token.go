package types

import (
	"fmt"
	"regexp"
)

const (
	NativeToken        = "hbc"
	NativeTokenDecimal = 18

	NativeDefiToken        = "hdt"
	NativeDefiTokenDecimal = 8

	GasPriceBucketWindow uint64 = 10
)

var (
	KeySendEnabled        = "send_enabled"
	KeyDepositEnabled     = "deposit_enabled"
	KeyWithdrawalEnabled  = "withdrawal_enabled"
	KeyCollectThreshold   = "collect_threshold"
	KeyDepositThreshold   = "deposit_threshold"
	KeyOpenFee            = "open_fee"
	KeySysOpenFee         = "sys_open_fee"
	KeyWithdrawalFeeRate  = "withdrawal_fee_rate"
	KeyMaxOpCUNumber      = "max_op_cu_number"
	KeySysTransferNum     = "systransfer_num"
	KeyOpCUSysTransferNum = "op_cu_systransfer_num"
	KeyGasLimit           = "gas_limit"
	KeyConfirmations      = "confirmations"
	KeyNeedCollectFee     = "need_collect_fee"
)

var (
	reTokenNameString = `[a-z][a-z0-9]{1,15}`
	reTokenName       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reTokenNameString))
)

//TokenType
type TokenType uint64

const (
	UtxoBased          TokenType = 0x1 //UtxoBased   i.e. token likes BTC
	AccountBased       TokenType = 0x2 //AccountBased i.e. token likes ETH, every user has its own account
	AccountSharedBased TokenType = 0x3 //Memo Based, i.e. token  like EOS/XRP, users share a account with differen memos/tags
)

func IsTokenTypeValid(tokenType TokenType) bool {
	return tokenType >= UtxoBased && tokenType <= AccountSharedBased
}

// IsTokenNameValid check token name.
func IsTokenNameValid(s string) bool {
	return reTokenName.MatchString(s)
}

type Symbol string

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) IsValid() bool {
	return reDnm.MatchString(s.String())
}

type Token interface {
	GetName() string
	GetSymbol() Symbol
	GetIssuer() string
	GetChain() Symbol
	IsSendEnabled() bool
	GetDecimals() uint64
	GetTotalSupply() Int
	GetWeight() int
	IsIBCToken() bool
	IsValid() bool
	String() string
}

var _ Token = (*BaseToken)(nil)

type BaseToken struct {
	Name        string `json:"name"`
	Symbol      Symbol `json:"symbol" yaml:"symbol"`
	Issuer      string `json:"issuer" yaml:"issuer"`              //token's issuer
	Chain       Symbol `json:"chain" yaml:"chain"`                //related mainnet token, e.g. ERC20 token's Chain is ETH
	SendEnabled bool   `json:"send_enabled" yaml:"send_enabled"`  //whether send enabled or not
	Decimals    uint64 `json:"decimals" yaml:"decimals"`          //token's decimals, represents by the decimals's
	TotalSupply Int    `json:"total_supply" yaml:"total_supply" ` //token's total supply
	Weight      int    `json:"weight" yaml:"weight"`
}

func (t *BaseToken) String() string {
	return fmt.Sprintf(`
	Name:%s
	Symbol:%s
	Issuer:%v
	Chain:%v
	SendEnabled:%v
	Decimals:%v
	TotalSupply:%v
	Weight:%v
	`, t.Name, t.Symbol, t.Issuer, t.Chain, t.SendEnabled, t.Decimals, t.TotalSupply, t.Weight)
}

func (t *BaseToken) IsValid() bool {
	if !t.Symbol.IsValid() || !t.Chain.IsValid() {
		return false
	}
	if !IsTokenNameValid(t.Name) {
		return false
	}

	if t.Decimals > Precision {
		return false
	}

	return true
}

func (t *BaseToken) GetName() string {
	return t.Name
}

func (t *BaseToken) GetSymbol() Symbol {
	return t.Symbol
}

func (t *BaseToken) GetIssuer() string {
	return t.Issuer
}

func (t *BaseToken) GetChain() Symbol {
	return t.Chain
}

func (t *BaseToken) IsSendEnabled() bool {
	return t.SendEnabled
}

func (t *BaseToken) GetDecimals() uint64 {
	return t.Decimals
}

func (t *BaseToken) GetTotalSupply() Int {
	return t.TotalSupply
}

func (t *BaseToken) GetWeight() int {
	return t.Weight
}

func (t *BaseToken) IsIBCToken() bool {
	return t.Chain != NativeToken
}

// IBCToken defines information of inter-blockchain token
type IBCToken struct {
	BaseToken

	TokenType          TokenType `json:"type" yaml:"type"`                                     //token's type
	DepositEnabled     bool      `json:"deposit_enabled" yaml:"deposit_enabled"`               //whether send enabled or not
	WithdrawalEnabled  bool      `json:"withdrawal_enabled" yaml:"withdrawal_enabled"`         //whether withdrawal enabled or not
	CollectThreshold   Int       `json:"collect_threshold" yaml:"collect_threshold" `          // token's collect threshold == account threshold
	DepositThreshold   Int       `json:"deposit_threshold" yaml:"deposit_threshold"`           // token's deposit threshold
	OpenFee            Int       `json:"open_fee" yaml:"open_fee"`                             // token's open fee for custodianunit address
	SysOpenFee         Int       `json:"sys_open_fee" yaml:"sys_open_fee"`                     // token's open fee for external address
	WithdrawalFeeRate  Dec       `json:"withdrawal_fee_rate" yaml:"withdrawal_fee_rate"`       // token's WithdrawalFeeRate
	MaxOpCUNumber      uint64    `json:"max_op_cu_number" yaml:"max_op_cu_number"`             // token's opcu num
	SysTransferNum     Int       `json:"sys_transfer_num" yaml:"sys_transfer_num"`             // 给user反向打币每次限额
	OpCUSysTransferNum Int       `json:"op_cu_sys_transfer_num" yaml:"op_cu_sys_transfer_num"` // 给 opcu之间转gas的每次限额
	GasLimit           Int       `json:"gas_limit" yaml:"gas_limit"`
	GasPrice           Int       `json:"gas_price" yaml:"gas_price"`
	Confirmations      uint64    `json:"confirmations" yaml:"confirmations"` //confirmation of chain
	IsNonceBased       bool      `json:"is_nonce_based" yaml:"is_nonce_based"`
	NeedCollectFee     bool      `json:"need_collect_fee" yaml:"need_collect_fee"`
}

func (t *IBCToken) String() string {
	return fmt.Sprintf(`
	Name:%s
	Symbol:%s
	Issuer:%v
	Chain:%v
	TokenType:%v
	SendEnabled:%v
	DepositEnabled:%v
	WithdrawalEnabled:%v
	Decimals:%v
	TotalSupply:%v
	Weight:%v
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
	NeedCollectFee:%v
	`, t.Name, t.Symbol, t.Issuer, t.Chain, t.TokenType, t.SendEnabled, t.DepositEnabled,
		t.WithdrawalEnabled, t.Decimals, t.TotalSupply, t.Weight, t.CollectThreshold, t.DepositThreshold,
		t.OpenFee, t.SysOpenFee, t.WithdrawalFeeRate, t.MaxOpCUNumber, t.SysTransferNum,
		t.OpCUSysTransferNum, t.GasLimit, t.GasPrice, t.Confirmations, t.IsNonceBased, t.NeedCollectFee)
}

func (t *IBCToken) IsValid() bool {
	if !t.BaseToken.IsValid() {
		return false
	}

	if !IsTokenTypeValid(t.TokenType) {
		return false
	}
	if !(t.CollectThreshold.IsPositive() && t.DepositThreshold.IsPositive() && t.OpenFee.IsPositive() && t.WithdrawalFeeRate.IsPositive()) {
		return false
	}

	if t.GasLimit.IsNegative() || t.GasPrice.IsNegative() || t.TotalSupply.IsNegative() {
		return false
	}

	if t.MaxOpCUNumber == 0 {
		return false
	}

	// sub token must have a issuer
	if t.Symbol != t.Chain && t.Issuer == "" {
		return false
	}

	return true
}

func (t *IBCToken) SysTransferAmount() Int {
	return t.GasPrice.Mul(t.GasLimit).Mul(t.SysTransferNum)
}

func (t *IBCToken) OpCUSysTransferAmount() Int {
	return t.GasPrice.Mul(t.GasLimit).Mul(t.OpCUSysTransferNum)
}

func (t *IBCToken) WithdrawalFee() Coin {
	var baseGasFee Int
	if t.TokenType == UtxoBased {
		baseGasFee = DefaultUtxoWithdrawTxSize().Mul(t.GasPrice).QuoRaw(KiloBytes)
	} else {
		baseGasFee = t.GasPrice.Mul(t.GasLimit)
	}
	withdrawalFeeAmt := t.WithdrawalFeeRate.Mul(NewDecFromInt(baseGasFee)).TruncateInt()
	return NewCoin(t.Chain.String(), withdrawalFeeAmt)
}

func (t *IBCToken) CollectFee() Coin {
	if !t.NeedCollectFee {
		return NewCoin(t.Chain.String(), ZeroInt())
	}

	var baseGasFee Int
	if t.TokenType == UtxoBased {
		baseGasFee = DefaultUtxoCollectTxSize().Mul(t.GasPrice).QuoRaw(KiloBytes)
	} else {
		baseGasFee = t.GasPrice.Mul(t.GasLimit)
	}
	if t.Chain != t.Symbol {
		collectFeeAmount := t.SysTransferAmount().Add(baseGasFee)
		return NewCoin(t.Chain.String(), collectFeeAmount)
	} else {
		return NewCoin(t.Chain.String(), baseGasFee)
	}
}

type TokensGasPrice struct {
	Chain    string `json:"chain" yaml:"chain"`
	GasPrice Int    `json:"gas_price" yaml:"gas_price"`
}

func (gp TokensGasPrice) String() string {
	return fmt.Sprintf(
		`Chain:%s
		"GasPrice:%v`,
		gp.Chain, gp.GasPrice.String())
}
