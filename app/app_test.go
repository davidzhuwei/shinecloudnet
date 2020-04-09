package app

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/shinecloudfoundation/shinecloudnet/codec"
	"github.com/shinecloudfoundation/shinecloudnet/simapp"
	"github.com/shinecloudfoundation/shinecloudnet/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBarkisdExport(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewBarkisApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	setGenesis(gapp)

	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewBarkisApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	_, _, err := newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestTxDecoder(t *testing.T) {
	cdc := MakeCodec()
	//txBytes, _ := hex.DecodeString("a602282816a90a9c01b42d614e0a68b9f831ab0a144368616e6765204d617856616c696461746f7273122c4368616e6765206d6178696d756d2076616c696461746f72207175616e74697479206c696d69746174696f6e1a1e0a077374616b696e67120d4d617856616c696461746f727322042233312212160a07756261726b6973120b31303030303030303030301a1492f72d9567793ec4ec022424c9cd171c146aca6212150a0f0a07756261726b697312043230303010c09a0c1a6a0a26eb5ae9872102b0664b6799d10e12e632eece4f738bf4e285c21201b8297fc760edf8e7579e4e1240e84336da2109cca8f7ecc2357afa5e205cf5e622da16f4e935e1628aa96379df03c0cbb1e8bbcb07cd375912af8ff89e937da17f2b02d83a69196932177da583")
	txBytes, _ := hex.DecodeString("a402282816a90a9a01b42d614e0a66b9f831ab0a144368616e6765204d617856616c696461746f7273122c4368616e6765206d6178696d756d2076616c696461746f72207175616e74697479206c696d69746174696f6e1a1c0a077374616b696e67120d4d617856616c696461746f72732202333112160a07756261726b6973120b31303030303030303030301a149481f41e0d9731182ef93c8147ba5d9f2476cd0212150a0f0a07756261726b697312043230303010c09a0c1a6a0a26eb5ae98721039eb41a732dc22e41d8e216e8606cc8f6eda20539a5f984f901eb364bd8e43e1712402d762edc57cb45858f7b6586c6889820691bd71074ff7badffed02bb01727c5a1375d1117d796d6bdb4436ca7f4a986994c6f1748b89fa78c1bec64a420019bf")
	decoder := auth.DefaultTxDecoder(cdc)
	tx, _ := decoder(txBytes)
	msg := tx.GetMsgs()
	t.Log(len(msg))
}

func setGenesis(gapp *BarkisApp) error {

	genesisState := simapp.NewDefaultGenesisState()
	stateBytes, err := codec.MarshalJSONIndent(gapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	gapp.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	gapp.Commit()
	return nil
}
