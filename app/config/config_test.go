package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/barkisnet/barkis/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultAppConfig()
	require.True(t, cfg.GetMinGasPrices().IsZero())
}

func TestSetMinimumFees(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.SetMinGasPrices(sdk.DecCoins{sdk.NewInt64DecCoin("foo", 5)})
	require.Equal(t, "5.000000000000000000foo", cfg.MinGasPrices)
}
