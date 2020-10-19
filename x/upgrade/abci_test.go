package upgrade_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hbtc-chain/bhchain/x/gov"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/module"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"github.com/hbtc-chain/bhchain/x/upgrade"
	"github.com/hbtc-chain/bhchain/x/upgrade/keeper"
	"github.com/hbtc-chain/bhchain/x/upgrade/types"
)

type TestSuite struct {
	module  module.AppModule
	keeper  keeper.Keeper
	querier sdk.Querier
	handler govtypes.Handler
	ctx     sdk.Context
}

var s TestSuite

func setupTest(height int64, skip map[int64]bool) TestSuite {
	keyUpgrade := sdk.NewKVStoreKey(types.StoreKey)
	keyGov := sdk.NewKVStoreKey(gov.StoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyUpgrade, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	cdc := codec.New()
	types.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)

	s.keeper = keeper.NewKeeper(skip, keyUpgrade, cdc, "")
	s.ctx = sdk.NewContext(ms, abci.Header{
		Height: height,
	}, false, log.NewNopLogger())

	s.module = upgrade.NewAppModule(s.keeper)
	s.querier = s.module.NewQuerierHandler()
	s.handler = upgrade.NewSoftwareUpgradeProposalHandler(s.keeper)
	return s
}

func TestRequireName(t *testing.T) {
	s := setupTest(10, map[int64]bool{})

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{}})
	require.NotNil(t, err)
	//	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestRequireFutureTime(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: s.ctx.BlockHeader().Time}})
	require.NotNil(t, err)
	//	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestRequireFutureBlock(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight()}})
	require.NotNil(t, err)
	//	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestCantSetBothTimeAndHeight(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now(), Height: s.ctx.BlockHeight() + 1}})
	require.NotNil(t, err)
	//require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestDoTimeUpgrade(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify can schedule an upgrade")
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.True(t, result.IsOK())

	VerifyDoUpgrade(t)
}

func TestDoHeightUpgrade(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify can schedule an upgrade")
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.True(t, result.IsOK())

	VerifyDoUpgrade(t)
}

func TestCanOverwriteScheduleUpgrade(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Can overwrite plan")
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "bad_test", Height: s.ctx.BlockHeight() + 10}})
	require.True(t, result.IsOK())
	result = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.True(t, result.IsOK())

	VerifyDoUpgrade(t)
}

func VerifyDoUpgrade(t *testing.T) {
	t.Log("Verify that a panic happens at the upgrade time/height")
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx sdk.Context, plan types.Plan) {})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)
}

func VerifyDoUpgradeWithCtx(t *testing.T, newCtx sdk.Context, proposalName string) {
	t.Log("Verify that a panic happens at the upgrade time/height")
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler(proposalName, func(ctx sdk.Context, plan types.Plan) {})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)
}

func TestHaltIfTooNew(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify that we don't panic with registered plan not in database at all")
	var called int
	s.keeper.SetUpgradeHandler("future", func(ctx sdk.Context, plan types.Plan) { called++ })

	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})
	require.Equal(t, 0, called)

	t.Log("Verify we panic if we have a registered handler ahead of time")
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "future", Height: s.ctx.BlockHeight() + 3}})
	require.True(t, result.IsOK())
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})
	require.Equal(t, 0, called)

	t.Log("Verify we no longer panic if the plan is on time")

	futCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 3).WithBlockTime(time.Now())
	req = abci.RequestBeginBlock{Header: futCtx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(futCtx, req)
	})
	require.Equal(t, 1, called)

	VerifyCleared(t, futCtx)
}

func VerifyCleared(t *testing.T, newCtx sdk.Context) {
	t.Log("Verify that the upgrade plan has been cleared")
	bz, err := s.querier(newCtx, []string{types.QueryCurrent}, abci.RequestQuery{})
	require.NoError(t, err)
	require.Nil(t, bz)
}

func TestCanClear(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify upgrade is scheduled")
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.True(t, result.IsOK())

	result = s.handler(s.ctx, &types.CancelSoftwareUpgradeProposal{Title: "cancel"})
	require.True(t, result.IsOK())

	VerifyCleared(t, s.ctx)
}

func TestCantApplySameUpgradeTwice(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.True(t, result.IsOK())
	VerifyDoUpgrade(t)
	t.Log("Verify an executed upgrade \"test\" can't be rescheduled")
	result = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.Equal(t, sdk.CodeInvalidTx, result.Code)
	require.True(t, strings.Contains(result.Log, "already been completed"))
}

func TestNoSpuriousUpgrades(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify that no upgrade panic is triggered in the BeginBlocker when we haven't scheduled an upgrade")
	req := abci.RequestBeginBlock{Header: s.ctx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(s.ctx, req)
	})
}

func TestPlanStringer(t *testing.T) {
	ti, err := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	require.Nil(t, err)
	require.Equal(t, `Upgrade Plan
  Name: test
  Time: 2020-01-01T00:00:00Z
  Info: `, types.Plan{Name: "test", Time: ti}.String())
	require.Equal(t, `Upgrade Plan
  Name: test
  Height: 100
  Info: `, types.Plan{Name: "test", Height: 100}.String())
}

func VerifyNotDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Log("Verify that upgrade was not done")
	height := s.keeper.GetDoneHeight(newCtx, name)
	require.Zero(t, height)
}

func VerifyDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Log("Verify that the upgrade plan has been executed")
	height := s.keeper.GetDoneHeight(newCtx, name)
	require.NotZero(t, height)
}

func VerifySet(t *testing.T, skipUpgradeHeights map[int64]bool) {
	t.Log("Verify if the skip upgrade has been set")

	for k := range skipUpgradeHeights {
		require.True(t, s.keeper.IsSkipHeight(k))
	}
}

func TestContains(t *testing.T) {
	var (
		skipOne int64 = 11
	)
	s := setupTest(10, map[int64]bool{skipOne: true})

	VerifySet(t, map[int64]bool{skipOne: true})
	t.Log("case where array contains the element")
	require.True(t, s.keeper.IsSkipHeight(11))

	t.Log("case where array doesn't contain the element")
	require.False(t, s.keeper.IsSkipHeight(4))
}

func TestSkipUpgradeSkippingAll(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(10, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.True(t, result.IsOK())

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a second proposal also is being cleared")
	result = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.True(t, result.IsOK())

	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	// To ensure verification is being done only after both upgrades are cleared
	t.Log("Verify if both proposals are cleared")
	VerifyCleared(t, s.ctx)
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
}

func TestUpgradeSkippingOne(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(10, map[int64]bool{skipOne: true})

	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.True(t, result.IsOK())

	t.Log("Verify if skip upgrade flag clears upgrade plan in one case and does upgrade on another")
	VerifySet(t, map[int64]bool{skipOne: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify the second proposal is not skipped")
	result = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.True(t, result.IsOK())
	// Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	VerifyDoUpgradeWithCtx(t, newCtx, "test2")

	t.Log("Verify first proposal is cleared and second is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyDone(t, s.ctx, "test2")
}

func TestUpgradeSkippingOnlyTwo(t *testing.T) {
	var (
		skipOne   int64 = 11
		skipTwo   int64 = 20
		skipThree int64 = 25
	)
	s := setupTest(10, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.True(t, result.IsOK())

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases and does third upgrade")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	// A new proposal with height in skipUpgradeHeights
	result = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.True(t, result.IsOK())
	// Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a new proposal is not skipped")
	result = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop3", Plan: types.Plan{Name: "test3", Height: skipThree}})
	require.True(t, result.IsOK())
	newCtx = newCtx.WithBlockHeight(skipThree)
	VerifyDoUpgradeWithCtx(t, newCtx, "test3")

	t.Log("Verify two proposals are cleared and third is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
	VerifyDone(t, s.ctx, "test3")
}

func TestUpgradeWithoutSkip(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	result := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.True(t, result.IsOK())
	t.Log("Verify if upgrade happens without skip upgrade")
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyDoUpgrade(t)
	VerifyDone(t, s.ctx, "test")
}

func TestDumpUpgradeInfoToFile(t *testing.T) {
	s := setupTest(10, map[int64]bool{})

	planHeight := s.ctx.BlockHeight() + 1
	name := "test"
	t.Log("verify if upgrade height is dumped to file")
	err := s.keeper.DumpUpgradeInfoToDisk(planHeight, name)
	require.Nil(t, err)

	upgradeInfoFilePath, err := s.keeper.GetUpgradeInfoPath()
	require.Nil(t, err)

	data, err := ioutil.ReadFile(upgradeInfoFilePath)
	require.NoError(t, err)

	var upgradeInfo types.UpgradeInfo
	err = json.Unmarshal(data, &upgradeInfo)
	require.Nil(t, err)

	t.Log("Verify upgrade height from file matches ")
	require.Equal(t, upgradeInfo.Height, planHeight)

	// clear the test file
	err = os.Remove(upgradeInfoFilePath)
	require.Nil(t, err)
}
