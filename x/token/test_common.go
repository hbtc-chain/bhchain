package token

import (
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
)

var (
	testBtcSymbol  = sdk.Symbol("btc")  //calSymbol("btc", "btc")
	testEthSymbol  = sdk.Symbol("eth")  //calSymbol("eth", "eth")
	testUsdtSymbol = sdk.Symbol("usdt") //calSymbol("usdt", "eth")
)

var TestIBCTokens = map[sdk.Symbol]*sdk.IBCToken{
	testBtcSymbol: {
		BaseToken: sdk.BaseToken{
			Name:        "btc",
			Symbol:      testBtcSymbol,
			Issuer:      "",
			Chain:       sdk.Symbol("btc"),
			SendEnabled: true,
			Decimals:    8,
			TotalSupply: sdk.NewIntWithDecimal(21, 15),
			Weight:      types.DefaultIBCTokenWeight,
		},
		TokenType:          sdk.UtxoBased,
		DepositEnabled:     true,
		WithdrawalEnabled:  true,
		CollectThreshold:   sdk.NewIntWithDecimal(2, 5),   // btc
		OpenFee:            sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:         sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:  sdk.NewDecWithPrec(2, 0),      // gas * 10  btc
		MaxOpCUNumber:      10,
		SysTransferNum:     sdk.NewInt(3),  // gas * 3
		OpCUSysTransferNum: sdk.NewInt(30), // SysTransferAmount * 10
		GasLimit:           sdk.NewInt(1),
		GasPrice:           sdk.NewInt(1000),
		DepositThreshold:   sdk.NewIntWithDecimal(2, 4),
		Confirmations:      1,
		IsNonceBased:       false,
		NeedCollectFee:     false,
	},
	testEthSymbol: {
		BaseToken: sdk.BaseToken{
			Name:        "eth",
			Symbol:      testEthSymbol,
			Issuer:      "",
			Chain:       sdk.Symbol("eth"),
			SendEnabled: true,
			Decimals:    18,
			TotalSupply: sdk.NewInt(0),
			Weight:      types.DefaultIBCTokenWeight + 1,
		},
		TokenType:          sdk.AccountBased,
		DepositEnabled:     true,
		WithdrawalEnabled:  true,
		CollectThreshold:   sdk.NewIntWithDecimal(2, 16),  // 0.02eth
		OpenFee:            sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:         sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:  sdk.NewDecWithPrec(2, 0),      // gas * 10, eth
		MaxOpCUNumber:      10,
		SysTransferNum:     sdk.NewInt(3),
		OpCUSysTransferNum: sdk.NewInt(30),
		GasLimit:           sdk.NewInt(21000),
		GasPrice:           sdk.NewInt(1000),
		DepositThreshold:   sdk.NewIntWithDecimal(2, 15), // 0.002eth
		Confirmations:      2,
		IsNonceBased:       true,
		NeedCollectFee:     false,
	},

	//a ERC20
	testUsdtSymbol: {
		BaseToken: sdk.BaseToken{
			Name:        "usdt",
			Symbol:      testUsdtSymbol,
			Issuer:      "0xFF760fcB0fa4Ba68d9DD2e28fc7A3c593b5d2106", // TODO (diff testnet & mainnet) (0xdAC17F958D2ee523a2206206994597C13D831ec7)
			Chain:       sdk.Symbol("eth"),
			SendEnabled: true,
			Decimals:    18,
			TotalSupply: sdk.NewIntWithDecimal(1, 28),
			Weight:      types.DefaultStableCoinWeight,
		},
		TokenType:          sdk.AccountBased,
		DepositEnabled:     true,
		WithdrawalEnabled:  true,
		CollectThreshold:   sdk.NewIntWithDecimal(2, 19),  // 20, tusdt
		OpenFee:            sdk.NewIntWithDecimal(28, 18), // nativeToken
		SysOpenFee:         sdk.NewIntWithDecimal(28, 18), // nativeToken
		WithdrawalFeeRate:  sdk.NewDecWithPrec(2, 0),      // gas * 10, eth
		MaxOpCUNumber:      10,
		SysTransferNum:     sdk.NewInt(1),
		OpCUSysTransferNum: sdk.NewIntWithDecimal(1, 2),
		GasLimit:           sdk.NewInt(80000), //  eth
		GasPrice:           sdk.NewInt(1000),
		DepositThreshold:   sdk.NewIntWithDecimal(1, 19), // 10tusdt
		Confirmations:      1,
		IsNonceBased:       true,
		NeedCollectFee:     false,
	},
}

var TestBaseTokens = map[sdk.Symbol]*sdk.BaseToken{
	sdk.NativeToken: {
		Name:        sdk.NativeToken,
		Symbol:      sdk.Symbol(sdk.NativeToken),
		Issuer:      "",
		Chain:       sdk.Symbol(sdk.NativeToken),
		SendEnabled: true,
		Decimals:    sdk.NativeTokenDecimal,
		TotalSupply: sdk.NewIntWithDecimal(21, 24),
		Weight:      types.DefaultNativeTokenWeight,
	},
	sdk.NativeDefiToken: {
		Name:        sdk.NativeDefiToken,
		Symbol:      sdk.Symbol(sdk.NativeDefiToken),
		Issuer:      "",
		Chain:       sdk.Symbol(sdk.NativeToken),
		SendEnabled: true,
		Decimals:    sdk.NativeDefiTokenDecimal,
		TotalSupply: sdk.NewIntWithDecimal(1, 16),
		Weight:      types.DefaultHrc10TokenWeight,
	},
}

var TestTokenData map[sdk.Symbol]sdk.Token

func init() {
	TestTokenData = make(map[sdk.Symbol]sdk.Token)
	for k, v := range TestIBCTokens {
		TestTokenData[k] = v
	}
	for k, v := range TestBaseTokens {
		TestTokenData[k] = v
	}
}
