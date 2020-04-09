package rest

import (
	"net/http"

	"github.com/shinecloudfoundation/shinecloudnet/client/context"
	sdk "github.com/shinecloudfoundation/shinecloudnet/types"
	"github.com/shinecloudfoundation/shinecloudnet/types/rest"
	"github.com/shinecloudfoundation/shinecloudnet/x/auth/client/utils"
	"github.com/shinecloudfoundation/shinecloudnet/x/gov"
	govrest "github.com/shinecloudfoundation/shinecloudnet/x/gov/client/rest"
	"github.com/shinecloudfoundation/shinecloudnet/x/params"
	paramscutils "github.com/shinecloudfoundation/shinecloudnet/x/params/client/utils"
)

// ProposalRESTHandler returns a ProposalRESTHandler that exposes the param
// change REST handler with a given sub-route.
func ProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "param_change",
		Handler:  postProposalHandlerFn(cliCtx),
	}
}

func postProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req paramscutils.ParamChangeProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		content := params.NewParameterChangeProposal(req.Title, req.Description, req.Changes.ToParamChanges())

		msg := gov.NewMsgSubmitProposal(content, req.Deposit, req.Proposer)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// derive the from account address and name from the Keybase
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
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
