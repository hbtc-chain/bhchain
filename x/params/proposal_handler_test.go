package params_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/params/subspace"
	"github.com/hbtc-chain/bhchain/x/params/types"
)

type testInput struct {
	ctx    sdk.Context
	cdc    *codec.Codec
	keeper params.Keeper
}

var (
	_ subspace.ParamSet = (*testParams)(nil)

	keyMaxValidators = "MaxValidators"
	keySlashingRate  = "SlashingRate"
	testSubspace     = "TestSubspace"
	keyBondName      = "BondName"
	keyRatio         = "Ratio"
)

type testParamsSlashingRate struct {
	DoubleSign uint16 `json:"double_sign,omitempty" yaml:"double_sign,omitempty"`
	Downtime   uint16 `json:"downtime,omitempty" yaml:"downtime,omitempty"`
}

type testParams struct {
	MaxValidators uint16                 `json:"max_validators" yaml:"max_validators"` // maximum number of validators (max uint16 = 65535)
	SlashingRate  testParamsSlashingRate `json:"slashing_rate" yaml:"slashing_rate"`
	BondDenom     string                 `json:"bond_demom" yaml:"bond_denom"`
	Ratio         sdk.Dec                `json:"ratio" yaml:"ratio"`
}

func (tp *testParams) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{[]byte(keyMaxValidators), &tp.MaxValidators},
		{[]byte(keySlashingRate), &tp.SlashingRate},
		{[]byte(keyBondName), &tp.BondDenom},
		{[]byte(keyRatio), &tp.Ratio},
	}
}

func testProposal(changes ...params.ParamChange) params.ParameterChangeProposal {
	return params.NewParameterChangeProposal(
		"Test",
		"description",
		changes,
	)
}

func newTestInput(t *testing.T) testInput {
	cdc := codec.New()
	types.RegisterCodec(cdc)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)

	keyParams := sdk.NewKVStoreKey("params")
	tKeyParams := sdk.NewTransientStoreKey("transient_params")

	cms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)

	err := cms.LoadLatestVersion()
	require.Nil(t, err)

	keeper := params.NewKeeper(cdc, keyParams, tKeyParams, params.DefaultCodespace)
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())

	return testInput{ctx, cdc, keeper}
}

func TestProposalHandlerPassed(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		params.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	tp := testProposal(params.NewParamChange(testSubspace, keyMaxValidators, "1"))
	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	require.Equal(t, sdk.CodeOK, hdlr(input.ctx, tp).Code)

	var param uint16
	ss.Get(input.ctx, []byte(keyMaxValidators), &param)
	require.Equal(t, param, uint16(1))
}

func TestProposalHandlerPassed1(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		params.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	tp := testProposal(params.NewParamChange(testSubspace, keyRatio, "\"0.040000000000000000\""))
	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	res := hdlr(input.ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)

	var param sdk.Dec
	ss.Get(input.ctx, []byte(keyRatio), &param)
	require.Equal(t, param, sdk.NewDecWithPrec(4, 2))
}

func TestProposalHandlerPassed2(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		params.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	tp := testProposal(params.NewParamChange(testSubspace, keyBondName, `"btc"`))
	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	res := hdlr(input.ctx, tp)
	require.Equal(t, sdk.CodeOK, res.Code)

	var param string
	ss.Get(input.ctx, []byte(keyBondName), &param)
	require.Equal(t, param, "btc")
}

func TestProposalHandlerFailed(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		params.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	tp := testProposal(params.NewParamChange(testSubspace, keyMaxValidators, "invalidType"))
	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	require.NotEqual(t, sdk.CodeOK, hdlr(input.ctx, tp))

	require.False(t, ss.Has(input.ctx, []byte(keyMaxValidators)))
}

func TestProposalHandlerUpdateOmitempty(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		params.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	var param testParamsSlashingRate

	tp := testProposal(params.NewParamChange(testSubspace, keySlashingRate, `{"downtime": 7}`))
	require.Equal(t, sdk.CodeOK, hdlr(input.ctx, tp).Code)

	ss.Get(input.ctx, []byte(keySlashingRate), &param)
	require.Equal(t, testParamsSlashingRate{0, 7}, param)

	tp = testProposal(params.NewParamChange(testSubspace, keySlashingRate, `{"double_sign": 10}`))
	require.Equal(t, sdk.CodeOK, hdlr(input.ctx, tp).Code)

	ss.Get(input.ctx, []byte(keySlashingRate), &param)
	require.Equal(t, testParamsSlashingRate{10, 7}, param)
}
