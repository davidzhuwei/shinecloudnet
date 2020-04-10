package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shinecloudfoundation/shinecloudnet/client"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/shinecloudfoundation/shinecloudnet/app"

	clientkeys "github.com/shinecloudfoundation/shinecloudnet/client/keys"
	"github.com/shinecloudfoundation/shinecloudnet/codec"
	"github.com/shinecloudfoundation/shinecloudnet/crypto/keys"
	"github.com/shinecloudfoundation/shinecloudnet/server"
	"github.com/shinecloudfoundation/shinecloudnet/simapp"
	"github.com/shinecloudfoundation/shinecloudnet/tests"
	sdk "github.com/shinecloudfoundation/shinecloudnet/types"
	"github.com/shinecloudfoundation/shinecloudnet/x/auth"
	"github.com/shinecloudfoundation/shinecloudnet/x/distribution"
	"github.com/shinecloudfoundation/shinecloudnet/x/gov"
	"github.com/shinecloudfoundation/shinecloudnet/x/slashing"
	"github.com/shinecloudfoundation/shinecloudnet/x/staking"
)

const (
	denom        = "stake"
	keyFoo       = "foo"
	keyBar       = "bar"
	fooDenom     = "footoken"
	feeDenom     = "feetoken"
	fee2Denom    = "fee2token"
	keyBaz       = "baz"
	keyVesting   = "vesting"
	keyFooBarBaz = "foobarbaz"
)

var (
	totalCoins = sdk.NewCoins(
		sdk.NewCoin(fee2Denom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(fooDenom, sdk.TokensFromConsensusPower(2000)),
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(300).Add(sdk.NewInt(12))), // add coins from inflation
	)

	startCoins = sdk.NewCoins(
		sdk.NewCoin(fee2Denom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(fooDenom, sdk.TokensFromConsensusPower(1000)),
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(150)),
	)

	vestingCoins = sdk.NewCoins(
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(500000)),
	)
)

//___________________________________________________________________________________
// Fixtures

// Fixtures is used to setup the testing environment
type Fixtures struct {
	BuildDir        string
	RootDir         string
	ScloudBinary    string
	ScloudcliBinary string
	ChainID         string
	RPCAddr         string
	Port            string
	ScloudHome      string
	ScloudcliHome   string
	P2PAddr         string
	T               *testing.T
}

// NewFixtures creates a new instance of Fixtures with many vars set
func NewFixtures(t *testing.T) *Fixtures {
	tmpDir, err := ioutil.TempDir("", "scloud_integration_"+t.Name()+"_")
	require.NoError(t, err)

	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)

	buildDir := os.Getenv("BUILDDIR")
	if buildDir == "" {
		buildDir, err = filepath.Abs("../build/")
		require.NoError(t, err)
	}

	return &Fixtures{
		T:               t,
		BuildDir:        buildDir,
		RootDir:         tmpDir,
		ScloudBinary:    filepath.Join(buildDir, "scloud"),
		ScloudcliBinary: filepath.Join(buildDir, "scloudcli"),
		ScloudHome:      filepath.Join(tmpDir, ".scloud"),
		ScloudcliHome:   filepath.Join(tmpDir, ".scloudcli"),
		RPCAddr:         servAddr,
		P2PAddr:         p2pAddr,
		Port:            port,
	}
}

// GenesisFile returns the path of the genesis file
func (f Fixtures) GenesisFile() string {
	return filepath.Join(f.ScloudHome, "config", "genesis.json")
}

// GenesisFile returns the application's genesis state
func (f Fixtures) GenesisState() simapp.GenesisState {
	cdc := codec.New()
	genDoc, err := tmtypes.GenesisDocFromFile(f.GenesisFile())
	require.NoError(f.T, err)

	var appState simapp.GenesisState
	require.NoError(f.T, cdc.UnmarshalJSON(genDoc.AppState, &appState))
	return appState
}

// InitFixtures is called at the beginning of a test  and initializes a chain
// with 1 validator.
func InitFixtures(t *testing.T) (f *Fixtures) {
	f = NewFixtures(t)

	// reset test state
	f.UnsafeResetAll()

	// ensure keystore has foo and bar keys
	f.KeysDelete(keyFoo)
	f.KeysDelete(keyBar)
	f.KeysDelete(keyBar)
	f.KeysDelete(keyFooBarBaz)
	f.KeysAdd(keyFoo)
	f.KeysAdd(keyBar)
	f.KeysAdd(keyBaz)
	f.KeysAdd(keyVesting)
	f.KeysAdd(keyFooBarBaz, "--multisig-threshold=2", fmt.Sprintf(
		"--multisig=%s,%s,%s", keyFoo, keyBar, keyBaz))

	// ensure that CLI output is in JSON format
	f.CLIConfig("output", "json")

	// NOTE: GDInit sets the ChainID
	f.GDInit(keyFoo)

	f.CLIConfig("chain-id", f.ChainID)
	f.CLIConfig("broadcast-mode", "block")
	f.CLIConfig("trust-node", "true")

	// start an account with tokens
	f.AddGenesisAccount(f.KeyAddress(keyFoo), startCoins)
	f.AddGenesisAccount(
		f.KeyAddress(keyVesting), startCoins,
		fmt.Sprintf("--vesting-amount=%s", vestingCoins),
		fmt.Sprintf("--vesting-start-time=%d", time.Now().UTC().UnixNano()),
		fmt.Sprintf("--vesting-end-time=%d", time.Now().Add(60*time.Second).UTC().UnixNano()),
	)

	f.GenTx(keyFoo)
	f.CollectGenTxs()

	return
}

// Cleanup is meant to be run at the end of a test to clean up an remaining test state
func (f *Fixtures) Cleanup(dirs ...string) {
	clean := append(dirs, f.RootDir)
	for _, d := range clean {
		require.NoError(f.T, os.RemoveAll(d))
	}
}

// Flags returns the flags necessary for making most CLI calls
func (f *Fixtures) Flags() string {
	return fmt.Sprintf("--home=%s --node=%s", f.ScloudcliHome, f.RPCAddr)
}

//___________________________________________________________________________________
// scloud

// UnsafeResetAll is scloud unsafe-reset-all
func (f *Fixtures) UnsafeResetAll(flags ...string) {
	cmd := fmt.Sprintf("%s --home=%s unsafe-reset-all", f.ScloudBinary, f.ScloudHome)
	executeWrite(f.T, addFlags(cmd, flags))
	err := os.RemoveAll(filepath.Join(f.ScloudHome, "config", "gentx"))
	require.NoError(f.T, err)
}

// GDInit is scloud init
// NOTE: GDInit sets the ChainID for the Fixtures instance
func (f *Fixtures) GDInit(moniker string, flags ...string) {
	cmd := fmt.Sprintf("%s init -o --home=%s %s", f.ScloudBinary, f.ScloudHome, moniker)
	_, stderr := tests.ExecuteT(f.T, addFlags(cmd, flags), client.DefaultKeyPass)

	var chainID string
	var initRes map[string]json.RawMessage

	err := json.Unmarshal([]byte(stderr), &initRes)
	require.NoError(f.T, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)

	f.ChainID = chainID
}

// AddGenesisAccount is scloud add-genesis-account
func (f *Fixtures) AddGenesisAccount(address sdk.AccAddress, coins sdk.Coins, flags ...string) {
	cmd := fmt.Sprintf("%s add-genesis-account %s %s --home=%s", f.ScloudBinary, address, coins, f.ScloudHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// GenTx is scloud gentx
func (f *Fixtures) GenTx(name string, flags ...string) {
	cmd := fmt.Sprintf("%s gentx --name=%s --home=%s --home-client=%s", f.ScloudBinary, name, f.ScloudHome, f.ScloudcliHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// CollectGenTxs is scloud collect-gentxs
func (f *Fixtures) CollectGenTxs(flags ...string) {
	cmd := fmt.Sprintf("%s collect-gentxs --home=%s", f.ScloudBinary, f.ScloudHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// GDStart runs scloud start with the appropriate flags and returns a process
func (f *Fixtures) GDStart(flags ...string) *tests.Process {
	cmd := fmt.Sprintf("%s start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.ScloudBinary, f.ScloudHome, f.RPCAddr, f.P2PAddr)
	proc := tests.GoExecuteTWithStdout(f.T, addFlags(cmd, flags))
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)
	return proc
}

// GDTendermint returns the results of scloud tendermint [query]
func (f *Fixtures) GDTendermint(query string) string {
	cmd := fmt.Sprintf("%s tendermint %s --home=%s", f.ScloudBinary, query, f.ScloudHome)
	success, stdout, stderr := executeWriteRetStdStreams(f.T, cmd)
	require.Empty(f.T, stderr)
	require.True(f.T, success)
	return strings.TrimSpace(stdout)
}

// ValidateGenesis runs scloud validate-genesis
func (f *Fixtures) ValidateGenesis() {
	cmd := fmt.Sprintf("%s validate-genesis --home=%s", f.ScloudBinary, f.ScloudHome)
	executeWriteCheckErr(f.T, cmd)
}

//___________________________________________________________________________________
// scloudcli keys

// KeysDelete is scloudcli keys delete
func (f *Fixtures) KeysDelete(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys delete --home=%s %s", f.ScloudcliBinary, f.ScloudcliHome, name)
	executeWrite(f.T, addFlags(cmd, append(append(flags, "-y"), "-f")))
}

// KeysAdd is scloudcli keys add
func (f *Fixtures) KeysAdd(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --home=%s %s", f.ScloudcliBinary, f.ScloudcliHome, name)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// KeysAddRecover prepares scloudcli keys add --recover
func (f *Fixtures) KeysAddRecover(name, mnemonic string, flags ...string) (exitSuccess bool, stdout, stderr string) {
	cmd := fmt.Sprintf("%s keys add --home=%s --recover %s", f.ScloudcliBinary, f.ScloudcliHome, name)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass, mnemonic)
}

// KeysAddRecoverHDPath prepares scloudcli keys add --recover --account --index
func (f *Fixtures) KeysAddRecoverHDPath(name, mnemonic string, account uint32, index uint32, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --home=%s --recover %s --account %d --index %d", f.ScloudcliBinary, f.ScloudcliHome, name, account, index)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), client.DefaultKeyPass, mnemonic)
}

// KeysShow is scloudcli keys show
func (f *Fixtures) KeysShow(name string, flags ...string) keys.KeyOutput {
	cmd := fmt.Sprintf("%s keys show --home=%s %s", f.ScloudcliBinary, f.ScloudcliHome, name)
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var ko keys.KeyOutput
	err := clientkeys.UnmarshalJSON([]byte(out), &ko)
	require.NoError(f.T, err)
	return ko
}

// KeyAddress returns the SDK account address from the key
func (f *Fixtures) KeyAddress(name string) sdk.AccAddress {
	ko := f.KeysShow(name)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

//___________________________________________________________________________________
// scloudcli config

// CLIConfig is scloudcli config
func (f *Fixtures) CLIConfig(key, value string, flags ...string) {
	cmd := fmt.Sprintf("%s config --home=%s %s %s", f.ScloudcliBinary, f.ScloudcliHome, key, value)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

//___________________________________________________________________________________
// scloudcli tx send/sign/broadcast

// TxSend is scloudcli tx send
func (f *Fixtures) TxSend(from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send %s %s %s %v", f.ScloudcliBinary, from, to, amount, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxSign is scloudcli tx sign
func (f *Fixtures) TxSign(signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --from=%s %v", f.ScloudcliBinary, f.Flags(), signer, fileName)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxBroadcast is scloudcli tx broadcast
func (f *Fixtures) TxBroadcast(fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.ScloudcliBinary, f.Flags(), fileName)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxEncode is scloudcli tx encode
func (f *Fixtures) TxEncode(fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.ScloudcliBinary, f.Flags(), fileName)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxMultisign is scloudcli tx multisign
func (f *Fixtures) TxMultisign(fileName, name string, signaturesFiles []string,
	flags ...string) (bool, string, string) {

	cmd := fmt.Sprintf("%s tx multisign %v %s %s %s", f.ScloudcliBinary, f.Flags(),
		fileName, name, strings.Join(signaturesFiles, " "),
	)
	return executeWriteRetStdStreams(f.T, cmd)
}

//___________________________________________________________________________________
// scloudcli tx staking

// TxStakingCreateValidator is scloudcli tx staking create-validator
func (f *Fixtures) TxStakingCreateValidator(from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking create-validator %v --from=%s --pubkey=%s", f.ScloudcliBinary, f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	cmd += fmt.Sprintf(" --min-self-delegation=%v", "1")
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxStakingUnbond is scloudcli tx staking unbond
func (f *Fixtures) TxStakingUnbond(from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx staking unbond %s %v --from=%s %v", f.ScloudcliBinary, validator, shares, from, f.Flags())
	return executeWrite(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

//___________________________________________________________________________________
// scloudcli tx gov

// TxGovSubmitProposal is scloudcli tx gov submit-proposal
func (f *Fixtures) TxGovSubmitProposal(from, typ, title, description string, deposit sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov submit-proposal %v --from=%s --type=%s", f.ScloudcliBinary, f.Flags(), from, typ)
	cmd += fmt.Sprintf(" --title=%s --description=%s --deposit=%s", title, description, deposit)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxGovDeposit is scloudcli tx gov deposit
func (f *Fixtures) TxGovDeposit(proposalID int, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov deposit %d %s --from=%s %v", f.ScloudcliBinary, proposalID, amount, from, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxGovVote is scloudcli tx gov vote
func (f *Fixtures) TxGovVote(proposalID int, option gov.VoteOption, from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov vote %d %s --from=%s %v", f.ScloudcliBinary, proposalID, option, from, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxGovSubmitParamChangeProposal executes a CLI parameter change proposal
// submission.
func (f *Fixtures) TxGovSubmitParamChangeProposal(
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {

	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal param-change %s --from=%s %v",
		f.ScloudcliBinary, proposalPath, from, f.Flags(),
	)

	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

// TxGovSubmitCommunityPoolSpendProposal executes a CLI community pool spend proposal
// submission.
func (f *Fixtures) TxGovSubmitCommunityPoolSpendProposal(
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {

	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal community-pool-spend %s --from=%s %v",
		f.ScloudcliBinary, proposalPath, from, f.Flags(),
	)

	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), client.DefaultKeyPass)
}

//___________________________________________________________________________________
// scloudcli query account

// QueryAccount is scloudcli query account
func (f *Fixtures) QueryAccount(address sdk.AccAddress, flags ...string) auth.BaseAccount {
	cmd := fmt.Sprintf("%s query account %s %v", f.ScloudcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)
	return acc
}

//___________________________________________________________________________________
// scloudcli query txs

// QueryTxs is scloudcli query txs
func (f *Fixtures) QueryTxs(page, limit int, tags ...string) *sdk.SearchTxsResult {
	cmd := fmt.Sprintf("%s query txs --page=%d --limit=%d --tags='%s' %v", f.ScloudcliBinary, page, limit, queryTags(tags), f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var result sdk.SearchTxsResult
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

// QueryTxsInvalid query txs with wrong parameters and compare expected error
func (f *Fixtures) QueryTxsInvalid(expectedErr error, page, limit int, tags ...string) {
	cmd := fmt.Sprintf("%s query txs --page=%d --limit=%d --tags='%s' %v", f.ScloudcliBinary, page, limit, queryTags(tags), f.Flags())
	_, err := tests.ExecuteT(f.T, cmd, "")
	require.EqualError(f.T, expectedErr, err)
}

//___________________________________________________________________________________
// scloudcli query staking

// QueryStakingValidator is scloudcli query staking validator
func (f *Fixtures) QueryStakingValidator(valAddr sdk.ValAddress, flags ...string) staking.Validator {
	cmd := fmt.Sprintf("%s query staking validator %s %v", f.ScloudcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var validator staking.Validator
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return validator
}

// QueryStakingUnbondingDelegationsFrom is scloudcli query staking unbonding-delegations-from
func (f *Fixtures) QueryStakingUnbondingDelegationsFrom(valAddr sdk.ValAddress, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations-from %s %v", f.ScloudcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var ubds []staking.UnbondingDelegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return ubds
}

// QueryStakingDelegationsTo is scloudcli query staking delegations-to
func (f *Fixtures) QueryStakingDelegationsTo(valAddr sdk.ValAddress, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations-to %s %v", f.ScloudcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var delegations []staking.Delegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return delegations
}

// QueryStakingPool is scloudcli query staking pool
func (f *Fixtures) QueryStakingPool(flags ...string) staking.Pool {
	cmd := fmt.Sprintf("%s query staking pool %v", f.ScloudcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var pool staking.Pool
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return pool
}

// QueryStakingParameters is scloudcli query staking parameters
func (f *Fixtures) QueryStakingParameters(flags ...string) staking.Params {
	cmd := fmt.Sprintf("%s query staking params %v", f.ScloudcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var params staking.Params
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}

//___________________________________________________________________________________
// scloudcli query gov

// QueryGovParamDeposit is scloudcli query gov param deposit
func (f *Fixtures) QueryGovParamDeposit() gov.DepositParams {
	cmd := fmt.Sprintf("%s query gov param deposit %s", f.ScloudcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var depositParam gov.DepositParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &depositParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return depositParam
}

// QueryGovParamVoting is scloudcli query gov param voting
func (f *Fixtures) QueryGovParamVoting() gov.VotingParams {
	cmd := fmt.Sprintf("%s query gov param voting %s", f.ScloudcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var votingParam gov.VotingParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votingParam
}

// QueryGovParamTallying is scloudcli query gov param tallying
func (f *Fixtures) QueryGovParamTallying() gov.TallyParams {
	cmd := fmt.Sprintf("%s query gov param tallying %s", f.ScloudcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var tallyingParam gov.TallyParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &tallyingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

// QueryGovProposals is scloudcli query gov proposals
func (f *Fixtures) QueryGovProposals(flags ...string) gov.Proposals {
	cmd := fmt.Sprintf("%s query gov proposals %v", f.ScloudcliBinary, f.Flags())
	stdout, stderr := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	if strings.Contains(stderr, "No matching proposals found") {
		return gov.Proposals{}
	}
	require.Empty(f.T, stderr)
	var out gov.Proposals
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(stdout), &out)
	require.NoError(f.T, err)
	return out
}

// QueryGovProposal is scloudcli query gov proposal
func (f *Fixtures) QueryGovProposal(proposalID int, flags ...string) gov.Proposal {
	cmd := fmt.Sprintf("%s query gov proposal %d %v", f.ScloudcliBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var proposal gov.Proposal
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return proposal
}

// QueryGovVote is scloudcli query gov vote
func (f *Fixtures) QueryGovVote(proposalID int, voter sdk.AccAddress, flags ...string) gov.Vote {
	cmd := fmt.Sprintf("%s query gov vote %d %s %v", f.ScloudcliBinary, proposalID, voter, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var vote gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return vote
}

// QueryGovVotes is scloudcli query gov votes
func (f *Fixtures) QueryGovVotes(proposalID int, flags ...string) []gov.Vote {
	cmd := fmt.Sprintf("%s query gov votes %d %v", f.ScloudcliBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var votes []gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votes
}

// QueryGovDeposit is scloudcli query gov deposit
func (f *Fixtures) QueryGovDeposit(proposalID int, depositor sdk.AccAddress, flags ...string) gov.Deposit {
	cmd := fmt.Sprintf("%s query gov deposit %d %s %v", f.ScloudcliBinary, proposalID, depositor, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var deposit gov.Deposit
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &deposit)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposit
}

// QueryGovDeposits is scloudcli query gov deposits
func (f *Fixtures) QueryGovDeposits(propsalID int, flags ...string) []gov.Deposit {
	cmd := fmt.Sprintf("%s query gov deposits %d %v", f.ScloudcliBinary, propsalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var deposits []gov.Deposit
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &deposits)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposits
}

//___________________________________________________________________________________
// query slashing

// QuerySigningInfo returns the signing info for a validator
func (f *Fixtures) QuerySigningInfo(val string) slashing.ValidatorSigningInfo {
	cmd := fmt.Sprintf("%s query slashing signing-info %s %s", f.ScloudcliBinary, val, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)
	cdc := app.MakeCodec()
	var sinfo slashing.ValidatorSigningInfo
	err := cdc.UnmarshalJSON([]byte(res), &sinfo)
	require.NoError(f.T, err)
	return sinfo
}

// QuerySlashingParams is scloudcli query slashing params
func (f *Fixtures) QuerySlashingParams() slashing.Params {
	cmd := fmt.Sprintf("%s query slashing params %s", f.ScloudcliBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)
	cdc := app.MakeCodec()
	var params slashing.Params
	err := cdc.UnmarshalJSON([]byte(res), &params)
	require.NoError(f.T, err)
	return params
}

//___________________________________________________________________________________
// query distribution

// QueryRewards returns the rewards of a delegator
func (f *Fixtures) QueryRewards(delAddr sdk.AccAddress, flags ...string) distribution.QueryDelegatorTotalRewardsResponse {
	cmd := fmt.Sprintf("%s query distribution rewards %s %s", f.ScloudcliBinary, delAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)
	cdc := app.MakeCodec()
	var rewards distribution.QueryDelegatorTotalRewardsResponse
	err := cdc.UnmarshalJSON([]byte(res), &rewards)
	require.NoError(f.T, err)
	return rewards
}

//___________________________________________________________________________________
// query supply

// QueryTotalSupply returns the total supply of coins
func (f *Fixtures) QueryTotalSupply(flags ...string) (totalSupply sdk.Coins) {
	cmd := fmt.Sprintf("%s query supply total %s", f.ScloudcliBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(res), &totalSupply)
	require.NoError(f.T, err)
	return totalSupply
}

// QueryTotalSupplyOf returns the total supply of a given coin denom
func (f *Fixtures) QueryTotalSupplyOf(denom string, flags ...string) sdk.Int {
	cmd := fmt.Sprintf("%s query supply total %s %s", f.ScloudcliBinary, denom, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)
	cdc := app.MakeCodec()
	var supplyOf sdk.Int
	err := cdc.UnmarshalJSON([]byte(res), &supplyOf)
	require.NoError(f.T, err)
	return supplyOf
}

//___________________________________________________________________________________
// executors

func executeWriteCheckErr(t *testing.T, cmdStr string, writes ...string) {
	require.True(t, executeWrite(t, cmdStr, writes...))
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) (exitSuccess bool) {
	exitSuccess, _, _ = executeWriteRetStdStreams(t, cmdStr, writes...)
	return
}

func executeWriteRetStdStreams(t *testing.T, cmdStr string, writes ...string) (bool, string, string) {
	proc := tests.GoExecuteT(t, cmdStr)

	// Enables use of interactive commands
	for _, write := range writes {
		_, err := proc.StdinPipe.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}

	// Read both stdout and stderr from the process
	stdout, stderr, err := proc.ReadAll()
	if err != nil {
		fmt.Println("Err on proc.ReadAll()", err, cmdStr)
	}

	// Log output.
	if len(stdout) > 0 {
		t.Log("Stdout:", string(stdout))
	}
	if len(stderr) > 0 {
		t.Log("Stderr:", string(stderr))
	}

	// Wait for process to exit
	proc.Wait()

	// Return succes, stdout, stderr
	return proc.ExitState.Success(), string(stdout), string(stderr)
}

//___________________________________________________________________________________
// utils

func addFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}

func queryTags(tags []string) (out string) {
	for _, tag := range tags {
		out += tag + "&"
	}
	return strings.TrimSuffix(out, "&")
}

// Write the given string to a new temporary file
func WriteToNewTempFile(t *testing.T, s string) *os.File {
	fp, err := ioutil.TempFile(os.TempDir(), "cosmos_cli_test_")
	require.Nil(t, err)
	_, err = fp.WriteString(s)
	require.Nil(t, err)
	return fp
}

func marshalStdTx(t *testing.T, stdTx auth.StdTx) []byte {
	cdc := app.MakeCodec()
	bz, err := cdc.MarshalBinaryBare(stdTx)
	require.NoError(t, err)
	return bz
}

func unmarshalStdTx(t *testing.T, s string) (stdTx auth.StdTx) {
	cdc := app.MakeCodec()
	require.Nil(t, cdc.UnmarshalJSON([]byte(s), &stdTx))
	return
}
