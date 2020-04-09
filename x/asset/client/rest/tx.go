package rest

import (
	"net/http"

	"github.com/shinecloudfoundation/shinecloudnet/client/context"
	sdk "github.com/shinecloudfoundation/shinecloudnet/types"
	"github.com/shinecloudfoundation/shinecloudnet/types/rest"
	"github.com/shinecloudfoundation/shinecloudnet/x/asset/internal/types"
	"github.com/shinecloudfoundation/shinecloudnet/x/auth/client/utils"
)

// IssueReq defines the properties of a send request's body.
type IssueReq struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name        string       `json:"name"`
	Symbol      string       `json:"symbol"`
	TotalSupply int64        `json:"total_supply"`
	Mintable    bool         `json:"mintable"`
	Decimal     int8         `json:"decimal"`
	Description string       `json:"description"`
}

// IssueRequestHandlerFn - http request handler to send coins to a address.
func IssueRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req IssueReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		var fromAddress sdk.AccAddress
		var fromName string
		var err error
		if req.BaseReq.GenerateOnly {
			fromAddress, err = sdk.AccAddressFromBech32(req.BaseReq.From)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			fromName=""
		} else {
			fromAddress, fromName, err = context.GetFromFieldsFromAddr(req.BaseReq.From)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress).WithBroadcastMode(req.BaseReq.BroadcastMode)
		msg := types.NewIssueMsg(fromAddress, req.Name, req.Symbol, req.TotalSupply, req.Mintable, req.Decimal, req.Description)
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// MintReq defines the properties of a send request's body.
type MintReq struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Symbol      string       `json:"symbol"`
	Amount      int64        `json:"amount"`
}

// IssueRequestHandlerFn - http request handler to send coins to a address.
func MintRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MintReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		var fromAddress sdk.AccAddress
		var fromName string
		var err error
		if req.BaseReq.GenerateOnly {
			fromAddress, err = sdk.AccAddressFromBech32(req.BaseReq.From)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			fromName=""
		} else {
			fromAddress, fromName, err = context.GetFromFieldsFromAddr(req.BaseReq.From)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress).WithBroadcastMode(req.BaseReq.BroadcastMode)
		msg := types.NewMintMsg(fromAddress, req.Symbol, req.Amount)
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
