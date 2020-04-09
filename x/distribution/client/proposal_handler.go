package client

import (
	"github.com/shinecloudfoundation/shinecloudnet/x/distribution/client/cli"
	"github.com/shinecloudfoundation/shinecloudnet/x/distribution/client/rest"
	govclient "github.com/shinecloudfoundation/shinecloudnet/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
