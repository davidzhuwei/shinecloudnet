package types

import (
	sdk "github.com/shinecloudfoundation/shinecloudnet/types"
	authexported "github.com/shinecloudfoundation/shinecloudnet/x/auth/exported"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	NewAccount(sdk.Context, authexported.Account) authexported.Account
	SetAccount(sdk.Context, authexported.Account)
	IterateAccounts(ctx sdk.Context, process func(authexported.Account) (stop bool))
}
