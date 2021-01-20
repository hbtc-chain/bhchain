package types

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

type QueryTokenInfoParams struct {
	Symbol string `json:"symbol"`
}

func NewQueryTokenInfoParams(symbol string) QueryTokenInfoParams {
	return QueryTokenInfoParams{Symbol: symbol}
}

type ResToken struct {
	Name                  string        `json:"name"`
	Symbol                string        `json:"symbol"`
	Issuer                string        `json:"issuer"`       //token's issuer
	Chain                 string        `json:"chain"`        //related mainnet token, e.g. ERC20 token's Chain is ETH
	TotalSupply           sdk.Int       `json:"total_supply"` //token's total supply
	Weight                int           `json:"weight"`
	Decimals              uint64        `json:"decimals"`            //token's decimals, represents by the decimals's
	SendEnabled           bool          `json:"send_enabled"`        //whether send enabled or not
	DepositEnabled        bool          `json:"deposit_enabled"`     //whether send enabled or not
	WithdrawalEnabled     bool          `json:"withdrawal_enabled"`  //whether withdrawal enabled or not
	TokenType             sdk.TokenType `json:"type"`                //token's type
	CollectThreshold      sdk.Int       `json:"collect_threshold"`   // token's collect threshold == account threshold
	DepositThreshold      sdk.Int       `json:"deposit_threshold"`   // token's deposit threshold
	OpenFee               sdk.Int       `json:"open_fee"`            // token's open fee for custodianunit address
	SysOpenFee            sdk.Int       `json:"sys_open_fee"`        // token's open fee for external address
	WithdrawalFeeRate     sdk.Dec       `json:"withdrawal_fee_rate"` // token's WithdrawalFeeRate
	MaxOpCUNumber         uint64        `json:"max_op_cu_number"`
	SysTransferNum        sdk.Int       `json:"sys_transfer_num"`       // 给user反向打币每次限额
	OpCUSysTransferNum    sdk.Int       `json:"op_cu_sys_transfer_num"` // 给 opcu之间转gas的每次限额
	GasLimit              sdk.Int       `json:"gas_limit"`
	GasPrice              sdk.Int       `json:"gas_price"`
	Confirmations         uint64        `json:"confirmations"` //confirmation of chain
	IsNonceBased          bool          `json:"is_nonce_based"`
	NeedCollectFee        bool          `json:"need_collect_fee"`
	SysTransferAmount     sdk.Int       `json:"sys_transfer_amount"`
	OpCUSysTransferAmount sdk.Int       `json:"op_cu_sys_transfer_amount"`
	WithdrawalFee         sdk.Int       `json:"withdrawal_fee"`
	CollectFee            sdk.Int       `json:"collect_fee"`
}

func NewResToken(token sdk.Token) *ResToken {
	ret := &ResToken{
		Name:        token.GetName(),
		Symbol:      token.GetSymbol().String(),
		Issuer:      token.GetIssuer(),
		Chain:       token.GetChain().String(),
		Decimals:    token.GetDecimals(),
		TotalSupply: token.GetTotalSupply(),
		Weight:      token.GetWeight(),
		SendEnabled: token.IsSendEnabled(),
	}
	if token.IsIBCToken() {
		ibcToken := token.(*sdk.IBCToken)
		ret.TokenType = ibcToken.TokenType
		ret.DepositEnabled = ibcToken.DepositEnabled
		ret.WithdrawalEnabled = ibcToken.WithdrawalEnabled
		ret.CollectThreshold = ibcToken.CollectThreshold
		ret.OpenFee = ibcToken.OpenFee
		ret.SysOpenFee = ibcToken.SysOpenFee
		ret.WithdrawalFeeRate = ibcToken.WithdrawalFeeRate
		ret.DepositThreshold = ibcToken.DepositThreshold
		ret.MaxOpCUNumber = ibcToken.MaxOpCUNumber
		ret.SysTransferNum = ibcToken.SysTransferNum
		ret.OpCUSysTransferNum = ibcToken.OpCUSysTransferNum
		ret.GasLimit = ibcToken.GasLimit
		ret.GasPrice = ibcToken.GasPrice
		ret.Confirmations = ibcToken.Confirmations
		ret.IsNonceBased = ibcToken.IsNonceBased
		ret.NeedCollectFee = ibcToken.NeedCollectFee
		ret.SysTransferAmount = ibcToken.SysTransferAmount()
		ret.OpCUSysTransferAmount = ibcToken.OpCUSysTransferAmount()
		ret.WithdrawalFee = ibcToken.WithdrawalFee().Amount
		ret.CollectFee = ibcToken.CollectFee().Amount
	}
	return ret
}
