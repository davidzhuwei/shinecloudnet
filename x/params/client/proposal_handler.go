package client

import (
	govclient "github.com/shinecloudfoundation/shinecloudnet/x/gov/client"
	"github.com/shinecloudfoundation/shinecloudnet/x/params/client/cli"
	"github.com/shinecloudfoundation/shinecloudnet/x/params/client/rest"
)

// param change proposal handler
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
