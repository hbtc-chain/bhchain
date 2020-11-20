package mock

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/supply"
)

const msgRoute = "testMsg"

var (
	numAccts                       = 2
	genCoins                       = sdk.Coins{sdk.NewInt64Coin("foocoin", 77)}
	accs, addrs, pubKeys, privKeys = CreateGenAccounts(numAccts, genCoins)
)

// testMsg is a mock transaction that has a validation which can fail.
type testMsg struct {
	signers     []sdk.CUAddress
	positiveNum int64
}

func (tx testMsg) Route() string                               { return msgRoute }
func (tx testMsg) Type() string                                { return "test" }
func (tx testMsg) GetMsg() sdk.Msg                             { return tx }
func (tx testMsg) GetMemo() string                             { return "" }
func (tx testMsg) GetSignBytes() []byte                        { return nil }
func (tx testMsg) GetSigners() []sdk.CUAddress                 { return tx.signers }
func (tx testMsg) GetSignatures() []custodianunit.StdSignature { return nil }
func (tx testMsg) ValidateBasic() sdk.Error {
	if tx.positiveNum >= 0 {
		return nil
	}
	return sdk.ErrTxDecode("positiveNum should be a non-negative integer.")
}

// getMockApp returns an initialized mock application.
func getMockApp(t *testing.T) *App {
	mApp := NewApp()

	mApp.Router().AddRoute(msgRoute, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	require.NoError(t, mApp.CompleteSetup(mApp.KeyTransfer))

	return mApp
}

func TestCheckAndDeliverGenTx(t *testing.T) {
	mApp := getMockApp(t)
	mApp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)

	SetGenesis(mApp, accs)
	ctxCheck := mApp.BaseApp.NewContext(true, abci.Header{})

	msg := testMsg{signers: []sdk.CUAddress{addrs[0]}, positiveNum: 1}

	acct := mApp.CUKeeper.GetCU(ctxCheck, addrs[0])
	require.Equal(t, custodianunit.NewBaseCU(accs[0].Type, accs[0].Address, accs[0].PubKey, accs[0].Sequence), acct.(*custodianunit.BaseCU))

	header := abci.Header{Height: mApp.LastBlockHeight() + 1}

	SignCheckDeliver(
		t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{msg},
		[]uint64{0},
		true, true, privKeys[0],
	)

	// Signing a tx with the wrong privKey should result in an auth error
	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	res, _, _ := SignCheckDeliver(
		t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{msg},
		[]uint64{1},
		true, false, privKeys[1],
	)

	require.Equal(t, sdk.CodeUnauthorized, res.Code, res.Log)
	require.Equal(t, sdk.CodespaceRoot, res.Codespace)

	// Resigning the tx with the correct privKey should result in an OK result
	header = abci.Header{Height: mApp.LastBlockHeight() + 1}

	SignCheckDeliver(
		t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{msg},
		[]uint64{1},
		true, true, privKeys[0],
	)
}

func TestCheckGenTx(t *testing.T) {
	mApp := getMockApp(t)
	mApp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)
	supply.RegisterCodec(mApp.Cdc)

	SetGenesis(mApp, accs)

	msg1 := testMsg{signers: []sdk.CUAddress{addrs[0]}, positiveNum: 1}
	CheckGenTx(
		t, mApp.BaseApp, []sdk.Msg{msg1},
		[]uint64{0},
		true, privKeys[0],
	)

	msg2 := testMsg{signers: []sdk.CUAddress{addrs[0]}, positiveNum: -1}
	CheckGenTx(
		t, mApp.BaseApp, []sdk.Msg{msg2},
		[]uint64{0},
		false, privKeys[0],
	)
}
