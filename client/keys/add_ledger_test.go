//+build ledger test_ledger_mock

package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/hbtc-chain/bhchain/client/flags"
	"github.com/hbtc-chain/bhchain/crypto/keys"
	"github.com/hbtc-chain/bhchain/tests"
	sdk "github.com/hbtc-chain/bhchain/types"
)

func Test_runAddCmdLedger(t *testing.T) {
	cmd := addKeyCommand()
	assert.NotNil(t, cmd)

	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(flags.FlagUseLedger, true)

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("test1234\ntest1234\n")
	assert.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	// Now check that it has been stored properly
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	assert.NotNil(t, kb)
	key1, err := kb.Get("keyname1")
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	assert.Equal(t, "keyname1", key1.GetName())
	assert.Equal(t, keys.TypeLedger, key1.GetType())
	assert.Equal(t,
		"hbcpub1addwnpepq0dk5hg9lleuyy0gd99nuhajq24shwup3k0k3t2eyn5y0phg5w3kgeeygy4",
		sdk.MustBech32ifyAccPub(key1.GetPubKey()))
}
