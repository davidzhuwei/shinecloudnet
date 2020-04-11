package types

import (
	"fmt"

	sdk "github.com/shinecloudfoundation/shinecloudnet/types"
	"github.com/shinecloudfoundation/shinecloudnet/x/params"
)

// Parameter store keys
var (
	KeyMintDenom              = []byte("MintDenom")
	KeyUnfreezeAmountPerBlock = []byte("UnfreezeAmountPerBlock")
)

// mint parameters
type Params struct {
	MintDenom              string `json:"mint_denom" yaml:"mint_denom"`
	UnfreezeAmountPerBlock int64  `json:"unfreeze_amount_per_block" yaml:"unfreeze_amount_per_block"`
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams (mintDenom string, unfreezeAmountPerBlock int64) Params {
	return Params{
		MintDenom:              mintDenom,
		UnfreezeAmountPerBlock: unfreezeAmountPerBlock,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:              sdk.DefaultBondDenom,
		UnfreezeAmountPerBlock: 1000000,
	}
}

// validate params
func ValidateParams(params Params) error {
	if params.MintDenom == "" {
		return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	}
	return nil
}

func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:             %s
  UnfreezeAmountPerBlock: %d
`,
		p.MintDenom, p.UnfreezeAmountPerBlock,
	)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMintDenom, &p.MintDenom},
		{KeyUnfreezeAmountPerBlock, &p.UnfreezeAmountPerBlock},
	}
}
