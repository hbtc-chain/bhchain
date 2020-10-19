package rest

import (
	"github.com/gorilla/mux"

	"github.com/hbtc-chain/bhchain/client/context"
)

// RegisterRoutes registers the auth module REST routes.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(
		"/cu/cus/{address}", QueryCURequestHandlerFn(storeName, cliCtx),
	).Methods("GET")
	r.HandleFunc(
		"/cu/pending_deposit/{address}", QueryPendingDepositRequestHandlerFn(storeName, cliCtx),
	).Methods("GET")
}

// RegisterTxRoutes registers all transaction routes on the provided router.
func RegisterTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(cliCtx)).Methods("GET")
	// r.HandleFunc("/txs", QueryTxsRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/txs", BroadcastTxRequest(cliCtx)).Methods("POST")
	r.HandleFunc("/txs/encode", EncodeTxRequestHandlerFn(cliCtx)).Methods("POST")
}
