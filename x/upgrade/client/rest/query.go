package rest

import (
	"fmt"
	"net/http"

	"github.com/hbtc-chain/bhchain/client/context"
	"github.com/hbtc-chain/bhchain/types/rest"
	"github.com/hbtc-chain/bhchain/x/upgrade/types"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers REST routes for the upgrade module under the path specified by routeName.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/upgrade/current", getCurrentPlanHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/upgrade/applied/{name}", getDonePlanHandler(cliCtx)).Methods("GET")
}

func getCurrentPlanHandler(cliCtx context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		// ignore height for now
		res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryCurrent))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getDonePlanHandler(cliCtx context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryAppliedParams(name)
		bz, err := cliCtx.Codec.MarshalJSON(params)

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryApplied), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
