package rest

import (
	"net/http"
	"time"

	"github.com/hbtc-chain/bhchain/client/context"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/rest"
	"github.com/hbtc-chain/bhchain/x/custodianunit/client/utils"
	govrest "github.com/hbtc-chain/bhchain/x/gov/client/rest"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/upgrade/types"
)

// PlanRequest defines a proposal for a new upgrade plan.
type PlanRequest struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title         string       `json:"title" yaml:"title"`
	Description   string       `json:"description" yaml:"description"`
	Deposit       sdk.Coins    `json:"deposit" yaml:"deposit"`
	UpgradeName   string       `json:"upgrade_name" yaml:"upgrade_name"`
	UpgradeHeight int64        `json:"upgrade_height" yaml:"upgrade_height"`
	UpgradeTime   string       `json:"upgrade_time" yaml:"upgrade_time"`
	UpgradeInfo   string       `json:"upgrade_info" yaml:"upgrade_info"`
}

// CancelRequest defines a proposal to cancel a current plan.
type CancelRequest struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title       string       `json:"title" yaml:"title"`
	Description string       `json:"description" yaml:"description"`
	Deposit     sdk.Coins    `json:"deposit" yaml:"deposit"`
}

func PostPlanProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "upgrade",
		Handler:  postPlanHandler(cliCtx),
	}
}

func CancelPlanProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "cancel_upgrade",
		Handler:  cancelPlanHandler(cliCtx),
	}
}

func postPlanHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlanRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.CUAddressFromBase58(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		var t time.Time
		if req.UpgradeTime != "" {
			t, err = time.Parse(time.RFC3339, req.UpgradeTime)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		plan := types.Plan{Name: req.UpgradeName, Time: t, Height: req.UpgradeHeight, Info: req.UpgradeInfo}
		content := types.NewSoftwareUpgradeProposal(req.Title, req.Description, plan)
		msg := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr, 0)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func cancelPlanHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CancelRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.CUAddressFromBase58(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		content := types.NewCancelSoftwareUpgradeProposal(req.Title, req.Description)
		msg := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr, 0)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
