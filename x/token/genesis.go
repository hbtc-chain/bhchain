/*
 * *******************************************************************
 * @项目名称: token
 * @文件名称: genesis.go
 * @Date: 2019/06/14
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

package token

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
)

var (
	EthToken  = "eth"
	BtcToken  = "btc"
	UsdtToken = "usdt"

	TrxToken  = "trx"
	Trx20USDT = "trxusdt"
)

var emptyInt = sdk.Int{}

type GenesisTokenInfo struct {
	Symbol        string `json:"symbol"`
	sdk.TokenInfo `json:"token_info"`
}

type GenesisState struct {
	GenesisTokenInfos []GenesisTokenInfo `json:"genesis_token_info"`
	Params            Params             `json:"params"`
}

func NewGenesisState(tokenInfo map[sdk.Symbol]sdk.TokenInfo, params Params) GenesisState {
	genInfo := make([]GenesisTokenInfo, 0, len(tokenInfo))
	for symbol, info := range tokenInfo {
		genInfo = append(genInfo, GenesisTokenInfo{Symbol: symbol.String(), TokenInfo: info})
	}
	sort.Sort(sortGenesisTokenInfo(genInfo))

	return GenesisState{
		GenesisTokenInfos: genInfo,
		Params:            params,
	}
}

func ValidateGenesis(data GenesisState) (map[string]struct{}, error) {
	tokenMap := make(map[string]struct{}, len(data.GenesisTokenInfos))

	for _, genInfo := range data.GenesisTokenInfos {
		symbol := sdk.Symbol(genInfo.Symbol)
		if !symbol.IsValidTokenName() {
			return nil, fmt.Errorf("Invalid Symbol:%v", symbol)
		}

		if sdk.NativeToken == symbol {
			continue
		}

		if !IsTokenTypeLegal(genInfo.TokenType) {
			return nil, fmt.Errorf("Invalid TokenInfoMap: Value: %v. Error: Missing type", genInfo)
		}

		if genInfo.TotalSupply == emptyInt || genInfo.TotalSupply.IsNegative() {
			return nil, fmt.Errorf("Invalid TokenInfoMap: Value: %v. Error: TotalSupply is negative", genInfo)
		}

		if genInfo.CollectThreshold == emptyInt || genInfo.CollectThreshold.IsNegative() {
			return nil, fmt.Errorf("Invalid TokenInfoMap: Value: %v. Error: Collect Threshold is negative", genInfo)
		}

		if genInfo.OpenFee == emptyInt || genInfo.OpenFee.IsNegative() {
			return nil, fmt.Errorf("Invalid TokenInfoMap: Value: %v. Error: Open Fee is negative", genInfo)
		}

		if genInfo.SysOpenFee == emptyInt || genInfo.SysOpenFee.IsNegative() {
			return nil, fmt.Errorf("Invalid TokenInfoMap: Value: %v. Error: Open Fee is negative", genInfo)
		}

		if genInfo.WithdrawalFeeRate == sdk.ZeroDec() || genInfo.WithdrawalFeeRate.IsNegative() {
			return nil, fmt.Errorf("Invalid TokenInfoMap: Value: %v. Error: Withdrawal Fee is negative", genInfo)
		}

		if _, ok := tokenMap[symbol.String()]; ok {
			return nil, fmt.Errorf("invalid TokenInfoMap: Value: %v. Error: Duplicated token", genInfo)
		}

		tokenMap[symbol.String()] = struct{}{}

	}

	return tokenMap, nil
}

func DefaultGenesisState() GenesisState {
	genInfos := []GenesisTokenInfo{
		{
			Symbol: sdk.NativeToken,
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol(sdk.NativeToken),
				Issuer:              "",
				Chain:               sdk.Symbol(sdk.NativeToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    false,
				IsWithdrawalEnabled: false,
				Decimals:            sdk.NativeTokenDecimal,
				TotalSupply:         sdk.NewIntWithDecimal(21, 24),
				CollectThreshold:    sdk.ZeroInt(),
				OpenFee:             sdk.ZeroInt(),
				SysOpenFee:          sdk.ZeroInt(),
				WithdrawalFeeRate:   sdk.ZeroDec(),
				MaxOpCUNumber:       0,
				SysTransferNum:      sdk.NewInt(0),
				OpCUSysTransferNum:  sdk.NewInt(0),
				GasLimit:            sdk.NewInt(1000000),
				GasPrice:            sdk.NewInt(1),
				DepositThreshold:    sdk.NewInt(0),
				Confirmations:       0,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: sdk.NativeDefiToken,
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol(sdk.NativeDefiToken),
				Issuer:              "",
				Chain:               sdk.Symbol(sdk.NativeToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    false,
				IsWithdrawalEnabled: false,
				Decimals:            sdk.NativeDefiTokenDecimal,
				TotalSupply:         sdk.NewIntWithDecimal(1, 16),
				CollectThreshold:    sdk.ZeroInt(),
				OpenFee:             sdk.ZeroInt(),
				SysOpenFee:          sdk.ZeroInt(),
				WithdrawalFeeRate:   sdk.ZeroDec(),
				MaxOpCUNumber:       0,
				SysTransferNum:      sdk.NewInt(0),
				OpCUSysTransferNum:  sdk.NewInt(0),
				GasLimit:            sdk.NewInt(1000000),
				GasPrice:            sdk.NewInt(1),
				DepositThreshold:    sdk.NewInt(0),
				Confirmations:       0,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: "btc",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("btc"),
				Issuer:              "",
				Chain:               sdk.Symbol("btc"),
				TokenType:           sdk.UtxoBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            8,
				TotalSupply:         sdk.NewIntWithDecimal(21, 14),
				CollectThreshold:    sdk.NewIntWithDecimal(1, 6),  // btc
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.ZeroInt(),                // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(1, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1), // gas * 3
				OpCUSysTransferNum:  sdk.NewInt(1), // SysTransferAmount * 10
				GasLimit:            sdk.NewInt(1),
				GasPrice:            sdk.NewInt(10000),
				DepositThreshold:    sdk.NewIntWithDecimal(1, 6),
				Confirmations:       1,
				IsNonceBased:        false,
			},
		},
		{
			Symbol: "eth",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol(EthToken),
				Issuer:              "",
				Chain:               sdk.Symbol(EthToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            18,
				TotalSupply:         sdk.NewIntWithDecimal(1, 27),
				CollectThreshold:    sdk.NewIntWithDecimal(1, 17), // 0.1eth
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(1, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(1),
				GasLimit:            sdk.NewInt(21000),
				GasPrice:            sdk.NewInt(10000),
				DepositThreshold:    sdk.NewIntWithDecimal(1, 17), // 0.1eth
				Confirmations:       2,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: "usdt",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("usdt"),
				Issuer:              "0xdac17f958d2ee523a2206206994597c13d831ec7", // TODO (diff testnet & mainnet) (0xdAC17F958D2ee523a2206206994597C13D831ec7)
				Chain:               sdk.Symbol(EthToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            6,
				TotalSupply:         sdk.NewIntWithDecimal(1, 17),
				CollectThreshold:    sdk.NewIntWithDecimal(10, 6), // 10 usdt
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),     //
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(2),
				GasLimit:            sdk.NewInt(80000), //  eth
				GasPrice:            sdk.NewIntWithDecimal(5, 9),
				DepositThreshold:    sdk.NewIntWithDecimal(10, 6), //10 usdt
				Confirmations:       2,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: "link",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("link"),
				Issuer:              "0x514910771af9ca656af840dff83e8264ecf986ca",
				Chain:               sdk.Symbol(EthToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            18,
				TotalSupply:         sdk.NewIntWithDecimal(1, 27),
				CollectThreshold:    sdk.NewIntWithDecimal(5, 18),
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(2),
				GasLimit:            sdk.NewInt(80000), //  eth
				GasPrice:            sdk.NewInt(1000),
				DepositThreshold:    sdk.NewIntWithDecimal(5, 18),
				Confirmations:       2,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: "ht",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("ht"),
				Issuer:              "0x6f259637dcd74c767781e37bc6133cd6a68aa161",
				Chain:               sdk.Symbol(EthToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            18,
				TotalSupply:         sdk.NewIntWithDecimal(5, 26),
				CollectThreshold:    sdk.NewIntWithDecimal(10, 18),
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(2),
				GasLimit:            sdk.NewInt(80000), //  eth
				GasPrice:            sdk.NewInt(1000),
				DepositThreshold:    sdk.NewIntWithDecimal(10, 18),
				Confirmations:       2,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: "ebtc",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("ebtc"),
				Issuer:              "0xd401551d50e33bc8a3424748df3d4b9a2afefe0e",
				Chain:               sdk.Symbol(EthToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            8,
				TotalSupply:         sdk.NewIntWithDecimal(21, 14),
				CollectThreshold:    sdk.NewIntWithDecimal(1, 6),  // btc
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(2),
				GasLimit:            sdk.NewInt(80000), //  eth
				GasPrice:            sdk.NewInt(1000),
				DepositThreshold:    sdk.NewIntWithDecimal(1, 6),
				Confirmations:       2,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: "etrx",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("etrx"),
				Issuer:              "0x6e9053b329ad6045eb2973ded766d922e8ebec69",
				Chain:               sdk.Symbol(EthToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            6,
				TotalSupply:         sdk.NewIntWithDecimal(1, 17),
				CollectThreshold:    sdk.NewIntWithDecimal(100, 6),
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(2),
				GasLimit:            sdk.NewInt(80000), //  eth
				GasPrice:            sdk.NewInt(1000),
				DepositThreshold:    sdk.NewIntWithDecimal(100, 6),
				Confirmations:       2,
				IsNonceBased:        true,
			},
		},
		{
			Symbol: TrxToken,
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol(TrxToken),
				Issuer:              "",
				Chain:               sdk.Symbol(TrxToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            6,
				TotalSupply:         sdk.NewIntWithDecimal(1, 17),
				CollectThreshold:    sdk.NewIntWithDecimal(100, 6), // 100 trx
				OpenFee:             sdk.NewIntWithDecimal(1, 16),  // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17),  // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(1, 0),      // 1 trx
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),       //1x gas
				OpCUSysTransferNum:  sdk.NewInt(5),       //5x gas
				GasLimit:            sdk.NewInt(1000000), //  1tron
				GasPrice:            sdk.NewInt(1),
				DepositThreshold:    sdk.NewIntWithDecimal(100, 6), // same as btc
				Confirmations:       20,
				IsNonceBased:        false,
			},
		},
		{
			Symbol: Trx20USDT,
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol(Trx20USDT),
				Issuer:              "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", // TODO (diff testnet & mainnet)
				Chain:               sdk.Symbol(TrxToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            6,
				TotalSupply:         sdk.NewIntWithDecimal(1, 17),
				CollectThreshold:    sdk.NewIntWithDecimal(10, 6), // 10 usdt
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),     // 1 tron
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),       //1x gas
				OpCUSysTransferNum:  sdk.NewInt(5),       //5x gas
				GasLimit:            sdk.NewInt(1000000), //  1trx
				GasPrice:            sdk.NewInt(1),
				DepositThreshold:    sdk.NewIntWithDecimal(10, 6), // 10 TRXUSDT
				Confirmations:       20,
				IsNonceBased:        false,
			},
		},
		{
			Symbol: "tbtc",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("tbtc"),
				Issuer:              "TK3Hr4ZVhjYH5AdxtR9fh6THeKh6WTnRAY",
				Chain:               sdk.Symbol(TrxToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            8,
				TotalSupply:         sdk.NewIntWithDecimal(21, 14),
				CollectThreshold:    sdk.NewIntWithDecimal(1, 6),  // btc
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(5),
				GasLimit:            sdk.NewInt(1000000), //  eth
				GasPrice:            sdk.NewInt(1),
				DepositThreshold:    sdk.NewIntWithDecimal(1, 6),
				Confirmations:       20,
				IsNonceBased:        false,
			},
		},
		{
			Symbol: "teth",
			TokenInfo: sdk.TokenInfo{
				Symbol:              sdk.Symbol("teth"),
				Issuer:              "TLkFhyYhRAK8AU8TxvFh6SkGQHZUk2joXk",
				Chain:               sdk.Symbol(TrxToken),
				TokenType:           sdk.AccountBased,
				IsSendEnabled:       true,
				IsDepositEnabled:    true,
				IsWithdrawalEnabled: true,
				Decimals:            18,
				TotalSupply:         sdk.NewIntWithDecimal(1, 27),
				CollectThreshold:    sdk.NewIntWithDecimal(1, 17),
				OpenFee:             sdk.NewIntWithDecimal(1, 16), // nativeToken
				SysOpenFee:          sdk.NewIntWithDecimal(1, 17), // nativeToken
				WithdrawalFeeRate:   sdk.NewDecWithPrec(2, 0),
				MaxOpCUNumber:       4,
				SysTransferNum:      sdk.NewInt(1),
				OpCUSysTransferNum:  sdk.NewInt(5),
				GasLimit:            sdk.NewInt(1000000), //  eth
				GasPrice:            sdk.NewInt(1),
				DepositThreshold:    sdk.NewIntWithDecimal(1, 17),
				Confirmations:       20,
				IsNonceBased:        false,
			},
		},
	}
	sort.Sort(sortGenesisTokenInfo(genInfos))

	return GenesisState{
		GenesisTokenInfos: genInfos,
		Params:            DefaultParams(),
	}
}

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) []abci.ValidatorUpdate {

	for _, genInfo := range data.GenesisTokenInfos {
		k.SetTokenInfo(ctx, &genInfo.TokenInfo)
	}
	k.SetParams(ctx, data.Params)
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	genTokenInfos := make([]GenesisTokenInfo, 0)
	iter := k.GetSymbolIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		symbol := sdk.Symbol(bytes.TrimPrefix(iter.Key(), TokenStoreKeyPrefix))
		info := k.GetTokenInfo(ctx, symbol)
		genTokenInfos = append(genTokenInfos, GenesisTokenInfo{Symbol: symbol.String(), TokenInfo: *info})
	}
	sort.Sort(sortGenesisTokenInfo(genTokenInfos))

	params := k.GetParams(ctx)
	return GenesisState{GenesisTokenInfos: genTokenInfos, Params: params}
}

func (g *GenesisState) AddTokenInfoIntoGenesis(new GenesisTokenInfo) error {
	genInfos := g.GenesisTokenInfos

	for _, info := range genInfos {
		if info.Symbol == new.Symbol {
			return fmt.Errorf("Token:%v already exist", new.Symbol)
		}
	}

	genInfos = append(genInfos, new)
	sort.Sort(sortGenesisTokenInfo(genInfos))
	g.GenesisTokenInfos = genInfos
	return nil
}

// Checks whether 2 GenesisState structs are equivalent.
func (g GenesisState) Equal(g2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(g)
	b2 := ModuleCdc.MustMarshalBinaryBare(g2)
	return bytes.Equal(b1, b2)
}

// Returns if a GenesisState is empty or has data in it
func (g GenesisState) IsEmpty() bool {
	emptyGenState := GenesisState{}
	return g.Equal(emptyGenState)
}

func (g GenesisState) String() string {
	var b strings.Builder

	for _, info := range g.GenesisTokenInfos {
		b.WriteString(fmt.Sprintf(`Symbol:%v`, info.Symbol))
		b.WriteString(info.TokenInfo.String())
		b.WriteString("\n")
	}
	b.WriteString(g.Params.String())

	return b.String()
}

type sortGenesisTokenInfo []GenesisTokenInfo

func (s sortGenesisTokenInfo) Len() int      { return len(s) }
func (s sortGenesisTokenInfo) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortGenesisTokenInfo) Less(i, j int) bool {
	return s[i].Symbol < s[j].Symbol
}
