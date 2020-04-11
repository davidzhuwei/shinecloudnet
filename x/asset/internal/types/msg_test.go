package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/shinecloudfoundation/shinecloudnet/types"
	"github.com/tendermint/tendermint/crypto"
)

func TestIssueMsgValidation(t *testing.T) {
	var emptyAddr sdk.AccAddress
	issuer := sdk.AccAddress(crypto.AddressHash([]byte("issuer")))

	cases := []struct {
		valid   bool
		errCode CodeType
		tx      IssueMsg
	}{
		{true, 0, NewIssueMsg(issuer, "bitcoin", "btc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{true, 0, NewIssueMsg(issuer, "bitcoin", "Btc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{true, 0, NewIssueMsg(issuer, "bitcoin", "BTC", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, sdk.CodeInvalidAddress, NewIssueMsg(emptyAddr, "bitcoin", "btc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},

		{false, CodeInvalidTokenSymbol, NewIssueMsg(issuer, "bitcoin", "uscds", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenSymbol, NewIssueMsg(issuer, "bitcoin", "scds", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenSymbol, NewIssueMsg(issuer, "bitcoin", "bt1", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenSymbol, NewIssueMsg(issuer, "bitcoin", "btc_", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenSymbol, NewIssueMsg(issuer, "bitcoin", "btcbtcbtcbtcbtc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},

		{false, CodeInvalidTokenName, NewIssueMsg(issuer, "uscds", "btc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenName, NewIssueMsg(issuer, "scds", "btc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenName, NewIssueMsg(issuer, "bitcoinbitcoinbitcoinbitcoinbitcoin", "btc", 21000000000000, false, 6, "bitcoin on shinecloudnet")},

		{false, CodeInvalidTotalSupply, NewIssueMsg(issuer, "bitcoin", "btc", 9000000000000000001, false, 6, "bitcoin on shinecloudnet")},
		{false, CodeInvalidDecimal, NewIssueMsg(issuer, "bitcoin", "btc", 21000000000000, false, -1, "bitcoin on shinecloudnet")},
		{false, CodeInvalidTokenDescription, NewIssueMsg(issuer, "bitcoin", "btc", 21000000000000, false, 6, "bitcoin on shinecloudnetbitcoin on shinecloudnetbitcoin on shinecloudnetbitcoin on shinecloudnetbitcoin on shinecloudnetbitcoin on shinecloudnetbitcoin on shinecloudnetbitcoin on shinecloudnet")},
	}

	for index, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			require.Nil(t, err)
		} else {
			require.NotNil(t, err)
			require.Equal(t, tc.errCode, err.Code(), fmt.Sprintf("index: %d, errMsg: %s", index, err.Error()))
		}
	}
}

func TestMintMsgValidation(t *testing.T) {
	var emptyAddr sdk.AccAddress
	minter := sdk.AccAddress(crypto.AddressHash([]byte("minter")))

	cases := []struct {
		valid   bool
		errCode CodeType
		tx      MintMsg
	}{
		{true, 0, NewMintMsg(minter, "btc", 10000)},

		{false, sdk.CodeInvalidAddress, NewMintMsg(emptyAddr, "btc", 10000)},

		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "Btc", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "BTC", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "btc_", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "btc_123", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "uscds", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "Uscds", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "scds", 10000)},
		{false, CodeInvalidTokenSymbol, NewMintMsg(minter, "SCDS", 10000)},

		{false, CodeInvalidMintAmount, NewMintMsg(minter, "btc", 9000000000000000001)},
	}

	for index, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			require.Nil(t, err)
		} else {
			require.NotNil(t, err)
			require.Equal(t, tc.errCode, err.Code(), fmt.Sprintf("index: %d, errMsg: %s", index, err.Error()))
		}
	}
}
