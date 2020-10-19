package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hbtc-chain/bhchain/client/context"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
	"github.com/hbtc-chain/bhchain/x/openswap/types"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers order-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
}

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/openswap/pair/{tokenA}/{tokenB}", getTradingPairHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/pairs", getAllTradingPairHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/liquidity/{addr}", getAddrLiquidityHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/orderbook/{pair}", getOrderbookHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/order/{orderID}", getOrderHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/pending_orders/{pair}/{addr}", getUnfinishedOrdersHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/earnings/{addr}", getEarningsHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/openswap/parameters", paramsHandlerFn(cliCtx)).Methods("GET")
}

func getTradingPairHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenA := mux.Vars(r)["tokenA"]
		tokenB := mux.Vars(r)["tokenB"]
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryTradingPairParams(sdk.Symbol(tokenA), sdk.Symbol(tokenB))
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryTradingPair), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getAllTradingPairHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryAllTradingPair), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getAddrLiquidityHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr, err := sdk.CUAddressFromBase58(mux.Vars(r)["addr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryAddrLiquidityParams(addr)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryAddrLiquidity), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getOrderbookHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbols := strings.Split(mux.Vars(r)["pair"], "-")
		if len(symbols) != 2 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid trading pair")
			return
		}
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		merge, _ := strconv.ParseBool(r.FormValue("merge"))
		params := types.NewQueryOrderbookParams(sdk.Symbol(symbols[0]), sdk.Symbol(symbols[1]), merge)
		bz := cliCtx.Codec.MustMarshalJSON(params)

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryOrderbook), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getOrderHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := mux.Vars(r)["orderID"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryOrderParams(orderID)
		bz := cliCtx.Codec.MustMarshalJSON(params)

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryOrder), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getUnfinishedOrdersHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbols := strings.Split(mux.Vars(r)["pair"], "-")
		if len(symbols) != 2 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid trading pair")
			return
		}
		addr, err := sdk.CUAddressFromBase58(mux.Vars(r)["addr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryUnfinishedOrderParams(sdk.Symbol(symbols[0]), sdk.Symbol(symbols[1]), addr)
		bz := cliCtx.Codec.MustMarshalJSON(params)

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryUnfinishedOrder), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getEarningsHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr, err := sdk.CUAddressFromBase58(mux.Vars(r)["addr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryUnclaimedEarningParams(addr)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryUnclaimedEarnings), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func paramsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryParameters), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
