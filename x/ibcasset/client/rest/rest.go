package rest

import (
	"github.com/gorilla/mux"

	"github.com/hbtc-chain/bhchain/client/context"
)

// RegisterRoutes registers the auth module REST routes.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(
		"/ibcasset/{address}", QueryCUAssetRequestHandlerFn(storeName, cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/ibcasset/pending_deposit/{address}", QueryPendingDepositRequestHandlerFn(storeName, cliCtx),
	).Methods("GET")
}
