package asset

import (
	"github.com/barkisnet/barkis/x/asset/internal/keeper"
	"github.com/barkisnet/barkis/x/asset/internal/types"
)

const (
	DefaultCodespace  = types.DefaultCodespace
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = keeper.DefaultParamspace
)

var (
	// functions aliases
	RegisterCodec = types.RegisterCodec
	NewKeeper     = keeper.NewKeeper
	NewQuerier    = keeper.NewQuerier
	NewParams     = types.NewParams

	// variable aliases
	ModuleCdc = types.ModuleCdc
	StoreKey  = types.StoreKey
)

type (
	Keeper = keeper.Keeper

	IssueMsg = types.IssueMsg
	MintMsg  = types.MintMsg
)
