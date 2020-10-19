package cli

import (
	"errors"
	"strings"

	"github.com/hbtc-chain/bhchain/client/context"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

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
	msg := types.NewMsgAddLiquidity(from, tokenA, tokenB, tokenAAmt, tokenBAmt, expiredAt)
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
	msg := types.NewMsgRemoveLiquidity(from, tokenA, tokenB, liquidity, expiredAt)
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
	msg := types.NewMsgSwapExactIn(from, referer, receiver, amtIn, minAmtOut, path, expiredAt)
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
	msg := types.NewMsgSwapExactOut(from, referer, receiver, amtOut, maxAmtIn, path, expiredAt)
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
	msg := types.NewMsgLimitSwap(uuid.NewV4().String(), from, referer, receiver, amtIn, price, baseSymbol, quoteSymbol, side, expiredAt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

func buildClaimEarningMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	from := cliCtx.GetFromAddress()
	tokenA := sdk.Symbol(viper.GetString(FlagTokenA))
	tokenB := sdk.Symbol(viper.GetString(FlagTokenB))
	msg := types.NewMsgClaimEarning(from, tokenA, tokenB)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}
