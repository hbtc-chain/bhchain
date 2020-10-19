package bhexapp

import (
	"io"
	"os"

	"github.com/hbtc-chain/bhchain/chainnode"
	"github.com/hbtc-chain/bhchain/x/evidence"
	"github.com/hbtc-chain/bhchain/x/hrc20"
	"github.com/hbtc-chain/bhchain/x/mapping"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/token"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/hbtc-chain/bhchain/baseapp"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/types/module"
	"github.com/hbtc-chain/bhchain/version"
	"github.com/hbtc-chain/bhchain/x/crisis"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	distr "github.com/hbtc-chain/bhchain/x/distribution"
	"github.com/hbtc-chain/bhchain/x/genaccounts"
	"github.com/hbtc-chain/bhchain/x/genutil"
	"github.com/hbtc-chain/bhchain/x/gov"
	"github.com/hbtc-chain/bhchain/x/keygen"
	mappingclient "github.com/hbtc-chain/bhchain/x/mapping/client"
	"github.com/hbtc-chain/bhchain/x/mint"
	"github.com/hbtc-chain/bhchain/x/openswap"
	"github.com/hbtc-chain/bhchain/x/order"
	otypes "github.com/hbtc-chain/bhchain/x/order/types"
	"github.com/hbtc-chain/bhchain/x/params"
	paramsclient "github.com/hbtc-chain/bhchain/x/params/client"
	"github.com/hbtc-chain/bhchain/x/slashing"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/transfer"
	"github.com/hbtc-chain/bhchain/x/upgrade"
	upgradeclient "github.com/hbtc-chain/bhchain/x/upgrade/client"
)

const appName = "HBTCApp"

var (
	// default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/.hbtccli")

	// default home directories for the application daemon
	DefaultNodeHome = os.ExpandEnv("$HOME/.hbtcchain")

	// The module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		custodianunit.AppModuleBasic{},
		transfer.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(paramsclient.ProposalHandler, distr.ProposalHandler,
			token.AddTokenProposalHandler, token.TokenParamsChangeProposalHandler,
			token.DisableTokenProposalHandler, upgradeclient.PostProposalHandler,
			upgradeclient.CancelProposalHandler, mappingclient.AddMappingProposalHandler, mappingclient.SwitchMappingProposalHandler),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
		token.AppModuleBasic{},
		receipt.AppModuleBasic{},
		order.AppModuleBasic{},
		keygen.AppModuleBasic{},
		hrc20.AppModuleBasic{},
		mapping.AppModuleBasic{},
		evidence.AppModuleBasic{},
		openswap.AppModuleBasic{},
		upgrade.AppModuleBasic{},
	)

	// module CU permissions
	maccPerms = map[string][]string{
		custodianunit.FeeCollectorName: nil,
		distr.ModuleName:               nil,
		mint.ModuleName:                {supply.Minter},
		staking.BondedPoolName:         {supply.Burner, supply.Staking},
		staking.NotBondedPoolName:      {supply.Burner, supply.Staking},
		gov.ModuleName:                 {supply.Burner},
		hrc20.ModuleName:               {supply.Minter, supply.Burner},
		openswap.ModuleName:            {supply.Minter, supply.Burner},
	}
)

// custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// Extended ABCI application
type bhexapp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// keepers
	cuKeeper       custodianunit.CUKeeper
	transferKeeper transfer.Keeper
	supplyKeeper   *supply.Keeper
	stakingKeeper  staking.Keeper
	slashingKeeper slashing.Keeper
	mintKeeper     mint.Keeper
	distrKeeper    distr.Keeper
	govKeeper      gov.Keeper
	crisisKeeper   crisis.Keeper
	paramsKeeper   params.Keeper
	tokenKeeper    token.Keeper
	receiptKeeper  receipt.Keeper
	orderKeeper    order.Keeper
	keygenKeeper   keygen.Keeper
	hrc20Keeper    hrc20.Keeper
	openswapKeeper openswap.Keeper
	mappingKeeper  mapping.Keeper
	evidenceKeeper evidence.Keeper
	upgradeKeeper  upgrade.Keeper

	// the module manager
	mm *module.Manager
}

// Newbhexapp returns a reference to an initialized bhexapp.
func Newbhexapp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, invCheckPeriod uint,
	cn chainnode.Chainnode, skipUpgradeHeights map[int64]bool, home string, baseAppOptions ...func(*bam.BaseApp),
) *bhexapp {

	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, custodianunit.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(bam.MainStoreKey, custodianunit.StoreKey, staking.StoreKey,
		supply.StoreKey, mint.StoreKey, distr.StoreKey, slashing.StoreKey, transfer.StoreKey,
		gov.StoreKey, params.StoreKey, token.StoreKey, receipt.StoreKey, otypes.StoreKey, keygen.StoreKey,
		hrc20.StoreKey, openswap.StoreKey, mapping.StoreKey, evidence.StoreKey, upgrade.StoreKey)
	tkeys := sdk.NewTransientStoreKeys(staking.TStoreKey, params.TStoreKey)

	app := &bhexapp{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, keys[params.StoreKey], tkeys[params.TStoreKey], params.DefaultCodespace)
	authSubspace := app.paramsKeeper.Subspace(custodianunit.DefaultParamspace)
	transferSubspace := app.paramsKeeper.Subspace(transfer.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace)
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	tokenSubspace := app.paramsKeeper.Subspace(token.DefaultParamspace)
	//receiptSubspace := app.paramsKeeper.Subspace(receipt.DefaultParamspace)
	orderSubspace := app.paramsKeeper.Subspace(otypes.DefaultParamspace)
	//keygenSubspace := app.paramsKeeper.Subspace(keygen.DefaultParamspace)
	hrc20Subspace := app.paramsKeeper.Subspace(hrc20.DefaultParamspace)
	mappingSubspace := app.paramsKeeper.Subspace(mapping.DefaultParamspace)
	openswapSubspace := app.paramsKeeper.Subspace(openswap.DefaultParamspace)
	evidenceSubspace := app.paramsKeeper.Subspace(evidence.DefaultParamspace)

	// add keepers
	app.tokenKeeper = token.NewKeeper(keys[token.StoreKey], app.cdc, tokenSubspace)
	app.receiptKeeper = *receipt.NewKeeper(app.cdc)
	app.orderKeeper = order.NewKeeper(app.cdc, keys[otypes.StoreKey], orderSubspace)
	app.cuKeeper = custodianunit.NewCUKeeper(app.cdc, keys[custodianunit.StoreKey], &app.tokenKeeper, authSubspace, custodianunit.ProtoBaseCU)
	app.supplyKeeper = supply.NewKeeper(app.cdc, keys[supply.StoreKey], app.cuKeeper, nil, maccPerms)
	stakingKeeper := staking.NewKeeper(app.cdc, keys[staking.StoreKey], tkeys[staking.TStoreKey],
		app.supplyKeeper, stakingSubspace, staking.DefaultCodespace)
	app.mintKeeper = mint.NewKeeper(app.cdc, keys[mint.StoreKey], mintSubspace, &stakingKeeper, app.supplyKeeper, custodianunit.FeeCollectorName)
	app.distrKeeper = distr.NewKeeper(app.cdc, keys[distr.StoreKey], distrSubspace, &stakingKeeper,
		app.supplyKeeper, distr.DefaultCodespace, custodianunit.FeeCollectorName, app.ModuleAccountAddrs())
	app.slashingKeeper = slashing.NewKeeper(app.cdc, keys[slashing.StoreKey], &stakingKeeper,
		slashingSubspace, slashing.DefaultCodespace)
	app.crisisKeeper = crisis.NewKeeper(crisisSubspace, invCheckPeriod, app.supplyKeeper, custodianunit.FeeCollectorName)
	app.keygenKeeper = keygen.NewKeeper(keys[keygen.StoreKey], app.cdc, &app.tokenKeeper, app.cuKeeper, &app.orderKeeper, &app.receiptKeeper, &stakingKeeper, app.distrKeeper, cn)
	app.transferKeeper = transfer.NewBaseKeeper(app.cdc, keys[transfer.StoreKey], app.cuKeeper, &app.tokenKeeper, &app.orderKeeper, &app.receiptKeeper, &stakingKeeper, cn, transferSubspace, transfer.DefaultCodespace, app.ModuleAccountAddrs())
	app.tokenKeeper.SetStakingKeeper(&stakingKeeper)

	app.hrc20Keeper = hrc20.NewKeeper(app.cdc, keys[hrc20.StoreKey], hrc20Subspace, &app.tokenKeeper, app.cuKeeper, app.distrKeeper, app.supplyKeeper, &app.receiptKeeper)
	app.mappingKeeper = mapping.NewKeeper(keys[mapping.StoreKey], app.cdc, &app.tokenKeeper, app.cuKeeper, &app.receiptKeeper, mappingSubspace)
	app.evidenceKeeper = evidence.NewKeeper(app.cdc, keys[evidence.StoreKey], evidenceSubspace, &stakingKeeper)

	app.transferKeeper.SetEvidenceKeeper(app.evidenceKeeper)
	app.tokenKeeper.SetEvidenceKeeper(app.evidenceKeeper)

	app.upgradeKeeper = upgrade.NewKeeper(skipUpgradeHeights, keys[upgrade.StoreKey], app.cdc, home)
	app.openswapKeeper = openswap.NewKeeper(app.cdc, keys[openswap.StoreKey], &app.tokenKeeper, app.cuKeeper, &app.receiptKeeper, app.supplyKeeper, app.transferKeeper, openswapSubspace)
	// register the proposal types
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(token.RouterKey, token.NewTokenProposalHandler(app.tokenKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper)).
		AddRoute(mapping.RouterKey, mapping.NewMappingProposalHandler(app.mappingKeeper))
	app.govKeeper = gov.NewKeeper(app.cdc, keys[gov.StoreKey], app.paramsKeeper, govSubspace,
		app.supplyKeeper, &stakingKeeper, app.distrKeeper, gov.DefaultCodespace, govRouter)

	app.cuKeeper.SetStakingKeeper(stakingKeeper)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks(),
			app.keygenKeeper.Hooks(),
			app.cuKeeper.Hooks(),
		),
	)

	app.supplyKeeper.SetTransferKeeper(app.transferKeeper)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genaccounts.NewAppModule(app.cuKeeper),
		genutil.NewAppModule(app.cuKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		custodianunit.NewAppModule(app.cuKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		supply.NewAppModule(*app.supplyKeeper, app.cuKeeper),
		distr.NewAppModule(app.distrKeeper, app.supplyKeeper),
		gov.NewAppModule(app.govKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.distrKeeper, app.cuKeeper, app.supplyKeeper),
		token.NewAppModule(app.tokenKeeper),
		receipt.NewAppModule(app.receiptKeeper),
		order.NewAppModule(app.orderKeeper),
		keygen.NewAppModule(app.keygenKeeper, &app.tokenKeeper, app.cuKeeper, &app.orderKeeper, &app.receiptKeeper, cn),
		transfer.NewAppModule(app.transferKeeper, app.cuKeeper, &app.tokenKeeper, &app.orderKeeper, &app.receiptKeeper, cn),
		hrc20.NewAppModule(app.hrc20Keeper),
		openswap.NewAppModule(app.openswapKeeper),
		mapping.NewAppModule(app.mappingKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(upgrade.ModuleName, mint.ModuleName, distr.ModuleName, openswap.ModuleName, slashing.ModuleName)
	app.mm.SetOrderEndBlockers(crisis.ModuleName, gov.ModuleName, staking.ModuleName, openswap.ModuleName, evidence.ModuleName)

	// NOTE: The genutils moodule must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		genaccounts.ModuleName, otypes.ModuleName, receipt.ModuleName, token.ModuleName, keygen.ModuleName, distr.ModuleName, staking.ModuleName,
		custodianunit.ModuleName, transfer.ModuleName, slashing.ModuleName, gov.ModuleName,
		mint.ModuleName, supply.ModuleName, crisis.ModuleName, genutil.ModuleName, hrc20.ModuleName, mapping.ModuleName, openswap.ModuleName,
		evidence.ModuleName)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(custodianunit.NewAnteHandler(app.cuKeeper, app.supplyKeeper, app.stakingKeeper, custodianunit.DefaultSigVerificationGasConsumer))
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			cmn.Exit(err.Error())
		}
	}
	return app
}

// application updates every begin block
func (app *bhexapp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// application updates every end block
func (app *bhexapp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// application update at chain initialization
func (app *bhexapp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, genesisState)
}

// load a particular height
func (app *bhexapp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// ModuleAccountAddrs returns all the app's module CU addresses.
func (app *bhexapp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
