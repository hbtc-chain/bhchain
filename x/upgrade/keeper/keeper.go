package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/hbtc-chain/bhchain/x/upgrade/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
)

// UpgradeInfoFileName file to store upgrade information
const UpgradeInfoFileName string = "upgrade-info.json"

type Keeper struct {
	homePath           string
	skipUpgradeHeights map[int64]bool
	storeKey           sdk.StoreKey
	cdc                *codec.Codec
	upgradeHandlers    map[string]types.UpgradeHandler
}

// NewKeeper constructs an upgrade Keeper
func NewKeeper(skipUpgradeHeights map[int64]bool, storeKey sdk.StoreKey, cdc *codec.Codec, homePath string) Keeper {
	return Keeper{
		homePath:           homePath,
		skipUpgradeHeights: skipUpgradeHeights,
		storeKey:           storeKey,
		cdc:                cdc,
		upgradeHandlers:    map[string]types.UpgradeHandler{},
	}
}

// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
// must be set even if it is a no-op function.
func (k Keeper) SetUpgradeHandler(name string, upgradeHandler types.UpgradeHandler) {
	k.upgradeHandlers[name] = upgradeHandler
}

// ScheduleUpgrade schedules an upgrade based on the specified plan.
// If there is another Plan already scheduled, it will overwrite it
// (implicitly cancelling the current plan)
func (k Keeper) ScheduleUpgrade(ctx sdk.Context, plan types.Plan) sdk.Result {
	if err := plan.ValidateBasic(); err != nil {
		return sdk.ErrInvalidTx(err.Error()).Result()
	}

	if !plan.Time.IsZero() {
		if !plan.Time.After(ctx.BlockHeader().Time) {
			return sdk.ErrInvalidTx("upgrade cannot be scheduled in the past").Result()
		}
	} else if plan.Height <= ctx.BlockHeight() {
		return sdk.ErrInvalidTx("upgrade cannot be scheduled in the past").Result()
	}

	if k.GetDoneHeight(ctx, plan.Name) != 0 {
		return sdk.ErrInvalidTx(fmt.Sprintf("upgrade with name %s has already been completed", plan.Name)).Result()
	}

	bz := k.cdc.MustMarshalBinaryBare(&plan)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PlanKey(), bz)

	return sdk.Result{}
}

// GetDoneHeight returns the height at which the given upgrade was executed
func (k Keeper) GetDoneHeight(ctx sdk.Context, name string) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(append([]byte{types.DoneByte}, []byte(name)...))
	if len(bz) == 0 {
		return 0
	}

	return int64(binary.BigEndian.Uint64(bz))
}

// ClearUpgradePlan clears any schedule upgrade
func (k Keeper) ClearUpgradePlan(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.PlanKey())
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
// upgrade or false if there is none
func (k Keeper) GetUpgradePlan(ctx sdk.Context) (plan types.Plan, havePlan bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PlanKey())
	if bz == nil {
		return plan, false
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &plan)
	return plan, true
}

// setDone marks this upgrade name as being done so the name can't be reused accidentally
func (k Keeper) setDone(ctx sdk.Context, name string) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	store.Set(append([]byte{types.DoneByte}, []byte(name)...), bz)
}

// HasHandler returns true iff there is a handler registered for this name
func (k Keeper) HasHandler(name string) bool {
	_, ok := k.upgradeHandlers[name]
	return ok
}

// ApplyUpgrade will execute the handler associated with the Plan and mark the plan as done.
func (k Keeper) ApplyUpgrade(ctx sdk.Context, plan types.Plan) {
	handler := k.upgradeHandlers[plan.Name]
	if handler == nil {
		panic("ApplyUpgrade should never be called without first checking HasHandler")
	}

	handler(ctx, plan)

	k.ClearUpgradePlan(ctx)
	k.setDone(ctx, plan.Name)
}

// IsSkipHeight checks if the given height is part of skipUpgradeHeights
func (k Keeper) IsSkipHeight(height int64) bool {
	return k.skipUpgradeHeights[height]
}

// DumpUpgradeInfoToDisk writes upgrade information to UpgradeInfoFileName.
func (k Keeper) DumpUpgradeInfoToDisk(height int64, name string) error {
	upgradeInfoFilePath, err := k.GetUpgradeInfoPath()
	if err != nil {
		return err
	}

	upgradeInfo := types.UpgradeInfo{
		Name:   name,
		Height: height,
	}
	info, err := json.Marshal(upgradeInfo)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(upgradeInfoFilePath, info, 0600)
}

// GetUpgradeInfoPath returns the upgrade info file path
func (k Keeper) GetUpgradeInfoPath() (string, error) {
	upgradeInfoFileDir := path.Join(k.getHomeDir(), "data")
	err := EnsureDir(upgradeInfoFileDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return filepath.Join(upgradeInfoFileDir, UpgradeInfoFileName), nil
}

// getHomeDir returns the height at which the given upgrade was executed
func (k Keeper) getHomeDir() string {
	return k.homePath
}

// ReadUpgradeInfoFromDisk returns the name and height of the upgrade
// which is written to disk by the old binary when panic'ing
// if there's an error in reading the info,
// it assumes that the upgrade info is not available
func (k Keeper) ReadUpgradeInfoFromDisk() (upgradeInfo types.UpgradeInfo) {
	upgradeInfoPath, err := k.GetUpgradeInfoPath()
	// if error in reading the path, assume there are no upgrades
	if err != nil {
		return upgradeInfo
	}

	data, err := ioutil.ReadFile(upgradeInfoPath)
	// if error in reading the file, assume there are no upgrades
	if err != nil {
		return upgradeInfo
	}

	json.Unmarshal(data, &upgradeInfo)
	return
}

func EnsureDir(dir string, mode os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, mode)
		if err != nil {
			return fmt.Errorf("could not create directory %v. %v", dir, err)
		}
	}
	return nil
}
