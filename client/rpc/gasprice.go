package rpc

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/client/flags"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
	cutypes "github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	defaultCount  = 12
	maxCount      = 100
	maxFetchTimes = 25
)

func GasPriceCommand(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-price [count]",
		Short: "Suggest for a reasonable gas price from the last [count] blocks",
		RunE:  GasPriceFactory(cdc),
		Args:  cobra.MaximumNArgs(1),
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	return cmd
}

func GasPriceFactory(cdc *codec.Codec) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		var count int
		// optional count
		if len(args) > 0 {
			h, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			if h <= 0 || h > maxCount {
				return fmt.Errorf("invalid count %s", args[0])
			} else {
				count = h
			}
		} else {
			count = defaultCount
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)
		suggestedGasPrice, err := suggestGasPrice(cliCtx, count)
		if err != nil {
			return err
		}

		fmt.Println(suggestedGasPrice.String())
		return nil
	}

}

func GasPriceRequestHandlerFn(cliCtx context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var count int
		countStrs, ok := r.URL.Query()["count"]
		if !ok || len(countStrs[0]) < 1 {
			count = defaultCount
		} else {
			var err error
			count, err = strconv.Atoi(countStrs[0])
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest,
					"couldn't parse count.")
				return
			}
		}

		output, err := suggestGasPrice(cliCtx, count)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, output)
	}
}

func suggestGasPrice(cliCtx context.CLIContext, count int) (sdk.DecCoins, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	blockGasPrices, err := fetchBlockGasPrices(node, cliCtx.Codec, count)
	if err != nil {
		return nil, err
	}

	if len(blockGasPrices) == 0 {
		res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", cutypes.QuerierRoute, cutypes.QueryMinimumGasPrice))
		if err != nil {
			return nil, err
		}
		var result sdk.DecCoins
		err = cliCtx.Codec.UnmarshalJSON(res, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	suggestedGasPrice := sdk.DecMedian(blockGasPrices)
	return []sdk.DecCoin{sdk.NewDecCoinFromDec(sdk.NativeToken, suggestedGasPrice)}, nil
}

func fetchBlockGasPrices(node client.Client, cdc *codec.Codec, expectedCount int) ([]sdk.Dec, error) {
	blockGasPrices := make([]sdk.Dec, 0)
	actualCount := 0       // ignore empty blocks
	lastHeight := int64(0) // height of last handled block

	info, err := node.BlockchainInfo(0, 0) // fetch latest 20 blocks
	if err != nil {
		return nil, err
	}
	fetchTimes := 1
	for {
		for _, meta := range info.BlockMetas {
			lastHeight = meta.Header.Height
			if meta.Header.NumTxs > 0 {
				block, err := node.Block(&meta.Header.Height)
				if err != nil {
					return nil, err
				}
				txGasPrices, err := txGasPricesInOneBlock(block, cdc)
				if err != nil {
					return nil, err
				}

				if len(txGasPrices) > 0 {
					actualCount++
					blockGasPrices = append(blockGasPrices, sdk.DecMedian(txGasPrices))
				}
			}
			if actualCount == expectedCount {
				break
			}
		}
		if actualCount == expectedCount || lastHeight <= 1 || fetchTimes == maxFetchTimes {
			break
		}
		info, err = node.BlockchainInfo(0, lastHeight-1) // fetch 20 blocks to lastHeight-1
		if err != nil {
			return nil, err
		}
		fetchTimes++
	}
	return blockGasPrices, nil
}

func txGasPricesInOneBlock(block *ctypes.ResultBlock, cdc *codec.Codec) ([]sdk.Dec, error) {
	txGasPrices := make([]sdk.Dec, 0)
	for _, tx := range block.Block.Txs {
		var stdTx cutypes.StdTx
		err := cdc.UnmarshalBinaryLengthPrefixed(tx, &stdTx)
		if err != nil {
			return nil, err
		}

		// todo use other method to filter out settle tx after implement zero fee for settle
		if !stdTx.Fee.Amount.IsZero() {
			gasPriceCoins := stdTx.Fee.GasPrices()
			// consider only tx with native token as fee
			if len(gasPriceCoins) != 1 || gasPriceCoins[0].Denom != sdk.NativeToken {
				continue
			}
			txGasPrices = append(txGasPrices, gasPriceCoins[0].Amount)
		}
	}
	return txGasPrices, nil
}
