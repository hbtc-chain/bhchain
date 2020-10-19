package token

import (
	sdk "github.com/hbtc-chain/bhchain/types"
)

var TestTokenData = map[sdk.Symbol]sdk.TokenInfo{
	"btc": {
		Symbol:              sdk.Symbol("btc"),
		Issuer:              "",
		Chain:               sdk.Symbol("btc"),
		TokenType:           sdk.UtxoBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    true,
		IsWithdrawalEnabled: true,
		Decimals:            8,
		TotalSupply:         sdk.NewIntWithDecimal(21, 15),
		CollectThreshold:    sdk.NewIntWithDecimal(2, 5),   // btc
		OpenFee:             sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:          sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),      // gas * 10  btc
		MaxOpCUNumber:       10,
		SysTransferNum:      sdk.NewInt(3),  // gas * 3
		OpCUSysTransferNum:  sdk.NewInt(30), // SysTransferAmount * 10
		GasLimit:            sdk.NewInt(1),
		GasPrice:            sdk.NewInt(1000),
		DepositThreshold:    sdk.NewIntWithDecimal(2, 4),
		Confirmations:       1,
		IsNonceBased:        false,
	},
	"eth": {
		Symbol:              sdk.Symbol(EthToken),
		Issuer:              "",
		Chain:               sdk.Symbol(EthToken),
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    true,
		IsWithdrawalEnabled: true,
		Decimals:            18,
		TotalSupply:         sdk.NewInt(0),
		CollectThreshold:    sdk.NewIntWithDecimal(2, 16),  // 0.02eth
		OpenFee:             sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:          sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),      // gas * 10, eth
		MaxOpCUNumber:       10,
		SysTransferNum:      sdk.NewInt(3),
		OpCUSysTransferNum:  sdk.NewInt(30),
		GasLimit:            sdk.NewInt(21000),
		GasPrice:            sdk.NewInt(1000),
		DepositThreshold:    sdk.NewIntWithDecimal(2, 15), // 0.002eth
		Confirmations:       2,
		IsNonceBased:        true,
	},

	//a ERC20
	"usdt": {
		Symbol:              sdk.Symbol(UsdtToken),
		Issuer:              "0xFF760fcB0fa4Ba68d9DD2e28fc7A3c593b5d2106", // TODO (diff testnet & mainnet) (0xdAC17F958D2ee523a2206206994597C13D831ec7)
		Chain:               sdk.Symbol(EthToken),
		TokenType:           sdk.AccountBased,
		IsSendEnabled:       true,
		IsDepositEnabled:    true,
		IsWithdrawalEnabled: true,
		Decimals:            18,
		TotalSupply:         sdk.NewIntWithDecimal(1, 28),
		CollectThreshold:    sdk.NewIntWithDecimal(2, 19),  // 20, tusdt
		OpenFee:             sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:          sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),      // gas * 10, eth
		MaxOpCUNumber:       10,
		SysTransferNum:      sdk.NewInt(1),
		OpCUSysTransferNum:  sdk.NewIntWithDecimal(1, 2),
		GasLimit:            sdk.NewInt(80000), //  eth
		GasPrice:            sdk.NewInt(1000),
		DepositThreshold:    sdk.NewIntWithDecimal(1, 19), // 10tusdt
		Confirmations:       1,
		IsNonceBased:        true,
	},
}
