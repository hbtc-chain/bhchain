package rest

import (
	"fmt"
	"net/http"

	"github.com/hbtc-chain/bhchain/client/context"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	govrest "github.com/hbtc-chain/bhchain/x/gov/client/rest"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/token/types"
	"github.com/gorilla/mux"
)

//TODO(liyong)
// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
}

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {

	// Query the information of a single denom
	r.HandleFunc(
		"/token/info/{denom}",
		tokenInfoHandlerFn(cliCtx),
	).Methods("GET")

	// Query the information of all tokens
	r.HandleFunc(
		"/token/tokens",
		allTokenInfosHandlerFn(cliCtx),
	).Methods("GET")
}

// HTTP request handler to query the supply of a single denom
func tokenInfoHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryTokenInfo(denom)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryToken), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// HTTP request handler to query the information of all tokens.
func allTokenInfosHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTokens), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ProposalRESTHandler returns a ProposalRESTHandler that exposes the community pool spend REST handler with a given sub-route.
func DisableTokenProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "disable_token",
		Handler:  disableTokenProposalHandlerFn(cliCtx),
	}
}

func AddTokenProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "add_token",
		Handler:  addTokenProposalHandlerFn(cliCtx),
	}
}

func TokenParamsChangeProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "token_params_change",
		Handler:  tokenParamsChangeProposalHandlerFn(cliCtx),
	}
}

func addTokenProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddTokenProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		content := types.NewAddTokenProposal(req.Title, req.Description, req.TokenInfo)
		msg := govtypes.NewMsgSubmitProposal(content, req.Deposit, req.Proposer, 0)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func tokenParamsChangeProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TokenParamsChangeProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		changes := req.Changes.ToParamChanges()
		content := types.NewTokenParamsChangeProposal(req.Title, req.Description, req.Symbol, changes)
		msg := govtypes.NewMsgSubmitProposal(content, req.Deposit, req.Proposer, 0)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func disableTokenProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DisableTokenProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		content := types.NewDisableTokenProposal(req.Title, req.Description, req.Symbol)
		msg := govtypes.NewMsgSubmitProposal(content, req.Deposit, req.Proposer, 0)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
