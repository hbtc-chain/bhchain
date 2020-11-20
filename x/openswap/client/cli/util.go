package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/hbtc-chain/bhchain/client/context"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

func buildCreateDexMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	name := viper.GetString(FlagDexName)
	incomeReceiver := from
	incomeReceiverStr := viper.GetString(FlagDexIncomeReceiver)
	var err error
	if incomeReceiverStr != "" {
		incomeReceiver, err = sdk.CUAddressFromBase58(incomeReceiverStr)
		if err != nil {
			return nil, err
		}
	}
	msg := types.NewMsgCreateDex(from, name, incomeReceiver)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildEditDexMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	dexID := viper.GetUint32(FlagDexID)
	name := viper.GetString(FlagDexName)
	var incomeReceiver *sdk.CUAddress
	incomeReceiverStr := viper.GetString(FlagDexIncomeReceiver)
	if incomeReceiverStr != "" {
		addr, err := sdk.CUAddressFromBase58(incomeReceiverStr)
		if err != nil {
			return nil, err
		}
		incomeReceiver = &addr
	}
	msg := types.NewMsgEditDex(from, dexID, name, incomeReceiver)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildCreateTradingPairMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	dexID := viper.GetUint32(FlagDexID)
	tokenA := sdk.Symbol(viper.GetString(FlagTokenA))
	tokenB := sdk.Symbol(viper.GetString(FlagTokenB))
	isPublic, _ := strconv.ParseBool(viper.GetString(FlagIsPublic))
	lpReward, err := sdk.NewDecFromStr(viper.GetString(FlagLpRewardRate))
	if err != nil {
		return nil, err
	}
	refererReward, err := sdk.NewDecFromStr(viper.GetString(FlagRefererRewardRate))
	if err != nil {
		return nil, err
	}
	msg := types.NewMsgCreateTradingPair(from, dexID, tokenA, tokenB, isPublic, lpReward, refererReward)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildEditTradingPairMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	dexID := viper.GetUint32(FlagDexID)
	tokenA := sdk.Symbol(viper.GetString(FlagTokenA))
	tokenB := sdk.Symbol(viper.GetString(FlagTokenB))
	var isPublic *bool
	boolStr := viper.GetString(FlagIsPublic)
	if boolStr != "" {
		b, err := strconv.ParseBool(boolStr)
		if err != nil {
			return nil, err
		}
		isPublic = &b
	}
	var lpReward, refererReward *sdk.Dec
	decStr := viper.GetString(FlagLpRewardRate)
	if decStr != "" {
		d, err := sdk.NewDecFromStr(decStr)
		if err != nil {
			return nil, err
		}
		lpReward = &d
	}
	decStr = viper.GetString(FlagRefererRewardRate)
	if decStr != "" {
		d, err := sdk.NewDecFromStr(decStr)
		if err != nil {
			return nil, err
		}
		refererReward = &d
	}

	msg := types.NewMsgEditTradingPair(from, dexID, tokenA, tokenB, isPublic, lpReward, refererReward)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildAddLiquidityMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	tokenA := sdk.Symbol(viper.GetString(FlagTokenA))
	tokenB := sdk.Symbol(viper.GetString(FlagTokenB))
	tokenAAmt, ok := sdk.NewIntFromString(viper.GetString(FlagTokenAAmount))
	if !ok {
		return nil, errors.New("invalid token a amount")
	}
	tokenBAmt, ok := sdk.NewIntFromString(viper.GetString(FlagTokenBAmount))
	if !ok {
		return nil, errors.New("invalid token b amount")
	}
	expiredAt := viper.GetInt64(FlagExpiredTime)
	msg := types.NewMsgAddLiquidity(from, viper.GetUint32(FlagDexID), tokenA, tokenB, tokenAAmt, tokenBAmt, expiredAt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildRemoveLiquidityMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	tokenA := sdk.Symbol(viper.GetString(FlagTokenA))
	tokenB := sdk.Symbol(viper.GetString(FlagTokenB))
	liquidity, ok := sdk.NewIntFromString(viper.GetString(FlagLiquidity))
	if !ok {
		return nil, errors.New("invalid liquidity amount")
	}
	expiredAt := viper.GetInt64(FlagExpiredTime)
	msg := types.NewMsgRemoveLiquidity(from, viper.GetUint32(FlagDexID), tokenA, tokenB, liquidity, expiredAt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildSwapExactInMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()

	var err error
	referer := from
	refererStr := viper.GetString(FlagReferer)
	if refererStr != "" {
		referer, err = sdk.CUAddressFromBase58(refererStr)
		if err != nil {
			return nil, errors.New("invalid referer address")
		}
	}
	receiver := from
	receiverStr := viper.GetString(FlagReceiver)
	if receiverStr != "" {
		receiver, err = sdk.CUAddressFromBase58(receiverStr)
		if err != nil {
			return nil, errors.New("invalid receiver address")
		}
	}

	amtIn, ok := sdk.NewIntFromString(viper.GetString(FlagAmountIn))
	if !ok {
		return nil, errors.New("invalid amount in")
	}
	minAmtOut, ok := sdk.NewIntFromString(viper.GetString(FlagMinAmountOut))
	if !ok {
		return nil, errors.New("invalid min amount out")
	}

	var path []sdk.Symbol
	for _, token := range strings.Split(viper.GetString(FlagSwapPath), ",") {
		path = append(path, sdk.Symbol(token))
	}

	expiredAt := viper.GetInt64(FlagExpiredTime)
	msg := types.NewMsgSwapExactIn(viper.GetUint32(FlagDexID), from, referer, receiver, amtIn, minAmtOut, path, expiredAt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildSwapExactOutMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()

	var err error
	referer := from
	refererStr := viper.GetString(FlagReferer)
	if refererStr != "" {
		referer, err = sdk.CUAddressFromBase58(refererStr)
		if err != nil {
			return nil, errors.New("invalid referer address")
		}
	}
	receiver := from
	receiverStr := viper.GetString(FlagReceiver)
	if receiverStr != "" {
		receiver, err = sdk.CUAddressFromBase58(receiverStr)
		if err != nil {
			return nil, errors.New("invalid receiver address")
		}
	}

	maxAmtIn, ok := sdk.NewIntFromString(viper.GetString(FlagMaxAmountIn))
	if !ok {
		return nil, errors.New("invalid max amount in")
	}
	amtOut, ok := sdk.NewIntFromString(viper.GetString(FlagAmountOut))
	if !ok {
		return nil, errors.New("invalid amount out")
	}

	var path []sdk.Symbol
	for _, token := range strings.Split(viper.GetString(FlagSwapPath), ",") {
		path = append(path, sdk.Symbol(token))
	}

	expiredAt := viper.GetInt64(FlagExpiredTime)
	msg := types.NewMsgSwapExactOut(viper.GetUint32(FlagDexID), from, referer, receiver, amtOut, maxAmtIn, path, expiredAt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildLimitSwapMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()

	var err error
	referer := from
	refererStr := viper.GetString(FlagReferer)
	if refererStr != "" {
		referer, err = sdk.CUAddressFromBase58(refererStr)
		if err != nil {
			return nil, errors.New("invalid referer address")
		}
	}
	receiver := from
	receiverStr := viper.GetString(FlagReceiver)
	if receiverStr != "" {
		receiver, err = sdk.CUAddressFromBase58(receiverStr)
		if err != nil {
			return nil, errors.New("invalid receiver address")
		}
	}

	amtIn, ok := sdk.NewIntFromString(viper.GetString(FlagAmountIn))
	if !ok {
		return nil, errors.New("invalid amountIn")
	}
	price, err := sdk.NewDecFromStr(viper.GetString(FlagPrice))
	if err != nil {
		return nil, err
	}

	baseSymbol := sdk.Symbol(viper.GetString(FlagBaseSymbol))
	quoteSymbol := sdk.Symbol(viper.GetString(FlagQuoteSymbol))
	side := viper.GetInt(FlagSide)

	expiredAt := viper.GetInt64(FlagExpiredTime)
	msg := types.NewMsgLimitSwap(uuid.NewV4().String(), viper.GetUint32(FlagDexID), from, referer, receiver, amtIn, price, baseSymbol, quoteSymbol, side, expiredAt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildClaimEarningMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	tokenA := sdk.Symbol(viper.GetString(FlagTokenA))
	tokenB := sdk.Symbol(viper.GetString(FlagTokenB))
	msg := types.NewMsgClaimEarning(from, viper.GetUint32(FlagDexID), tokenA, tokenB)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}
