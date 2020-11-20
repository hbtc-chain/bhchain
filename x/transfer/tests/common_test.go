package tests

// DONTCOVER

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/chainnode"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
	distr "github.com/hbtc-chain/bhchain/x/distribution"
	"github.com/hbtc-chain/bhchain/x/evidence"
	evitypes "github.com/hbtc-chain/bhchain/x/evidence/types"
	"github.com/hbtc-chain/bhchain/x/gov"
	"github.com/hbtc-chain/bhchain/x/ibcasset"
	ibcexported "github.com/hbtc-chain/bhchain/x/ibcasset/exported"
	"github.com/hbtc-chain/bhchain/x/order"
	"github.com/hbtc-chain/bhchain/x/receipt"
	"github.com/hbtc-chain/bhchain/x/slashing"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/token"
	"github.com/hbtc-chain/bhchain/x/transfer"
	"github.com/hbtc-chain/bhchain/x/transfer/keeper"

	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

var (
	Addrs = createTestAddrs(100)
	PKs   = createTestPubKeys(100)

	addrDels = []sdk.CUAddress{
		Addrs[0],
		Addrs[1],
	}
	addrVals = []sdk.ValAddress{
		sdk.ValAddress(Addrs[2]),
		sdk.ValAddress(Addrs[3]),
		sdk.ValAddress(Addrs[4]),
		sdk.ValAddress(Addrs[5]),
		sdk.ValAddress(Addrs[6]),
		sdk.ValAddress(Addrs[7]),
		sdk.ValAddress(Addrs[8]),
		sdk.ValAddress(Addrs[9]),
		sdk.ValAddress(Addrs[10]),
		sdk.ValAddress(Addrs[11]),
	}
)

type testInput struct {
	cdc            *codec.Codec
	ctx            sdk.Context
	k              keeper.BaseKeeper
	ck             custodianunit.CUKeeper
	tk             token.Keeper
	ok             order.Keeper
	rk             receipt.Keeper
	stakingkeeper  staking.Keeper
	supplyKeeper   supply.Keeper
	cn             types.Chainnode
	pk             params.Keeper
	validators     []staking.Validator
	ik             ibcasset.Keeper
	opcu           *testCU
	evidenceKeeper evidence.Keeper
	trk            transfer.Keeper
}

type testCU struct {
	custodianunit.CU
	ctx sdk.Context
	trk transfer.Keeper
	ik  ibcasset.Keeper
}

func newTestCU(ctx sdk.Context, trk transfer.Keeper, ik ibcasset.Keeper, cu custodianunit.CU) *testCU {
	return &testCU{CU: cu, ctx: ctx, ik: ik, trk: trk}
}

func (t *testCU) SetCoins(coins sdk.Coins) error {
	curCoins := t.trk.GetAllBalance(t.ctx, t.CU.GetAddress())
	t.trk.SubCoins(t.ctx, t.CU.GetAddress(), curCoins)
	t.trk.AddCoins(t.ctx, t.CU.GetAddress(), coins)
	return nil
}

func (t *testCU) SetCoinsHold(coins sdk.Coins) error {
	curCoins := t.trk.GetAllHoldBalance(t.ctx, t.CU.GetAddress())
	t.trk.SubCoinsHold(t.ctx, t.CU.GetAddress(), curCoins)
	t.trk.AddCoinsHold(t.ctx, t.CU.GetAddress(), coins)
	return nil
}

func (t *testCU) GetCoins() sdk.Coins {
	return t.trk.GetAllBalance(t.ctx, t.CU.GetAddress())
}

func (t *testCU) AddCoins(coins sdk.Coins) {
	t.trk.AddCoins(t.ctx, t.CU.GetAddress(), coins)
}

func (t *testCU) GetMigrationStatus() sdk.MigrationStatus {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetMigrationStatus()
}

func (t *testCU) IsEnabledSendTx(chain string, addr string) bool {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.IsEnabledSendTx(chain, addr)
}

func (t *testCU) SetEnableSendTx(enabled bool, chain string, addr string) {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	asset.SetEnableSendTx(enabled, chain, addr)
	t.ik.SetCUIBCAsset(t.ctx, asset)
}

func (t *testCU) GetAssetAddress(denom string, epoch uint64) string {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetAssetAddress(denom, epoch)
}

func (t *testCU) SetAssetAddress(denom, address string, epoch uint64) error {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	err := asset.SetAssetAddress(denom, address, epoch)
	if err != nil {
		return err
	}
	t.ik.SetCUIBCAsset(t.ctx, asset)
	return nil
}

func (t *testCU) GetIBCAsset() ibcexported.CUIBCAsset {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset
}

func (t *testCU) GetAssetPubkey(epoch uint64) []byte {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetAssetPubkey(epoch)
}

func (t *testCU) SetAssetPubkey(pubkey []byte, epoch uint64) error {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	err := asset.SetAssetPubkey(pubkey, epoch)
	if err != nil {
		return err
	}
	t.ik.SetCUIBCAsset(t.ctx, asset)
	return nil
}

func (t *testCU) AddAsset(denom, address string, epoch uint64) error {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	err := asset.AddAsset(denom, address, epoch)
	if err != nil {
		return err
	}
	t.ik.SetCUIBCAsset(t.ctx, asset)
	return nil
}

func (t *testCU) GetAssetCoinsHold() sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetAssetCoinsHold()
}

func (t *testCU) GetAssetCoins() sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetAssetCoins()
}

func (t *testCU) GetCoinsHold() sdk.Coins {
	return t.trk.GetAllHoldBalance(t.ctx, t.CU.GetAddress())
}

func (t *testCU) AddGasReceived(coins sdk.Coins) sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	ret := asset.AddGasReceived(coins)
	t.ik.SetCUIBCAsset(t.ctx, asset)
	return ret
}

func (t *testCU) GetGasReceived() sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetGasReceived()
}

func (t *testCU) GetGasUsed() sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	return asset.GetGasUsed()
}

func (t *testCU) AddAssetCoins(coins sdk.Coins) sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	ret := asset.AddAssetCoins(coins)
	t.ik.SetCUIBCAsset(t.ctx, asset)
	return ret
}

func (t *testCU) SubAssetCoins(coins sdk.Coins) sdk.Coins {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	ret := asset.SubAssetCoins(coins)
	t.ik.SetCUIBCAsset(t.ctx, asset)
	return ret
}

func (t *testCU) AddGasRemained(chain string, addr string, amt sdk.Int) {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	asset.AddGasRemained(chain, addr, amt)
	t.ik.SetCUIBCAsset(t.ctx, asset)
}

func (t *testCU) SubGasRemained(chain string, addr string, amt sdk.Int) {
	asset := t.ik.GetOrNewCUIBCAsset(t.ctx, t.CU.GetCUType(), t.CU.GetAddress())
	asset.SubGasRemained(chain, addr, amt)
	t.ik.SetCUIBCAsset(t.ctx, asset)
}

var mockCN chainnode.MockChainnode
var ethAddr = "0x12Db85318582809C733A14f48279ea9f21B9c6B9"

func setupTestInput(t *testing.T) testInput {
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(custodianunit.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyToken := sdk.NewKVStoreKey(token.StoreKey)
	keyReceipt := sdk.NewKVStoreKey(receipt.StoreKey)
	keyOrder := sdk.NewKVStoreKey(order.StoreKey)
	keyGov := sdk.NewKVStoreKey(gov.StoreKey)
	keySlash := sdk.NewKVStoreKey(slashing.StoreKey)
	keyTransfer := sdk.NewKVStoreKey(types.StoreKey)
	keyDistr := sdk.NewKVStoreKey(distr.StoreKey)
	keyEvid := sdk.NewKVStoreKey(evidence.StoreKey)
	keyIbcAsset := sdk.NewKVStoreKey(ibcasset.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)

	ms.MountStoreWithDB(keyToken, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyReceipt, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyOrder, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySlash, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyTransfer, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyEvid, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIbcAsset, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	//register cdc
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	custodianunit.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	params.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	token.RegisterCodec(cdc)
	receipt.RegisterCodec(cdc)
	order.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	slashing.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	evidence.RegisterCodec(cdc)
	ibcasset.RegisterCodec(cdc)
	cdc.RegisterConcrete(&testCU{}, "hbtcchain/test/testCU", nil)

	feeCollectorAcc := supply.NewEmptyModuleAccount(custodianunit.FeeCollectorName)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[sdk.CUAddress([]byte("moduleAcc")).String()] = true
	blacklistedAddrs[feeCollectorAcc.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	tk := token.NewKeeper(keyToken, cdc)
	rk := receipt.NewKeeper(cdc)
	ok := order.NewKeeper(cdc, keyOrder, pk.Subspace(order.DefaultParamspace))
	ck := custodianunit.NewCUKeeper(
		cdc, keyAcc, pk.Subspace(custodianunit.DefaultParamspace), custodianunit.ProtoBaseCU,
	)
	ck.SetParams(ctx, custodianunit.DefaultParams())
	ik := ibcasset.NewKeeper(cdc, keyIbcAsset, ck, &tk, ibcasset.ProtoBaseCUIBCAsset)

	bankKeeper := keeper.NewBaseKeeper(cdc, keyTransfer, ck, ik, &tk, &ok, rk, nil, &mockCN, pk.Subspace(types.DefaultParamspace), types.DefaultCodespace, blacklistedAddrs)
	transfer.InitGenesis(ctx, *bankKeeper, transfer.DefaultGenesisState())

	maccPerms := map[string][]string{
		custodianunit.FeeCollectorName: nil,
		staking.NotBondedPoolName:      []string{supply.Burner, supply.Staking},
		staking.BondedPoolName:         []string{supply.Burner, supply.Staking},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, ck, bankKeeper, maccPerms)
	initPower := int64(100000)
	numValidators := 4
	initTokens := sdk.TokensFromConsensusPower(initPower)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(int64(numValidators))))
	testSetCUCoins(ctx, bankKeeper, notBondedPool.Address, totalSupply)
	//notBondedPool.SetCoins(totalSupply)

	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))
	supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	supplyKeeper.SetModuleAccount(ctx, bondPool)
	supplyKeeper.SetModuleAccount(ctx, notBondedPool)

	stakingKeeper := staking.NewKeeper(cdc, keyStaking, tkeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), types.DefaultCodespace)
	params := staking.DefaultParams()
	stakingKeeper.SetParams(ctx, params)

	evidenceKeeper := evidence.NewKeeper(cdc, keyEvid, pk.Subspace(evidence.DefaultParamspace), stakingKeeper)

	evidence.InitGenesis(ctx, evidenceKeeper, evidence.DefaultGenesisState())
	eviParams := evitypes.BehaviourParams{
		MaxMisbehaviourCount:   10,
		BehaviourWindow:        100,
		BehaviourSlashFraction: sdk.NewDecFromIntWithPrec(sdk.NewInt(1), 1),
	}

	evidenceKeeper.SetBehaviourParams(ctx, string(evitypes.VoteBehaviourKey), eviParams)
	evidenceKeeper.SetBehaviourParams(ctx, string(evitypes.DsignBehaviourKey), eviParams)
	//fill all the addresses with some coins, set the loose pool tokens simultaneously
	//for _, addr := range Addrs {
	//	_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	bankKeeper.SetStakingKeeper(stakingKeeper)
	tk.SetStakingKeeper(stakingKeeper)

	//create 4 validators

	for i := 0; i < numValidators; i++ {
		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		_, _, err := bankKeeper.AddCoins(ctx, sdk.CUAddress(valAddr), initCoins)
		require.Nil(t, err)
		require.Equal(t, initTokens, bankKeeper.GetAllBalance(ctx, sdk.CUAddress(valAddr)).AmountOf(sdk.DefaultBondDenom))

		valTokens := sdk.TokensFromConsensusPower(initPower)
		validator := staking.NewValidator(valAddr, valPubKey, staking.Description{})
		validator, _ = validator.AddTokensFromDel(valTokens)
		require.Equal(t, sdk.Unbonded, validator.Status)
		require.Equal(t, valTokens, validator.Tokens)
		require.Equal(t, valTokens, validator.DelegatorShares.RoundInt())
		stakingKeeper.SetValidator(ctx, validator)
		stakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	}
	updates := stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, numValidators, len(updates))

	vals := stakingKeeper.GetAllValidators(ctx)
	require.Equal(t, numValidators, len(vals))

	valsAddr := []sdk.CUAddress{}
	for _, val := range vals {
		valsAddr = append(valsAddr, sdk.CUAddress(val.OperatorAddress))
	}
	ctx = ctx.WithBlockHeight(0)
	stakingKeeper.StartNewEpoch(ctx, valsAddr)
	ctx = ctx.WithBlockHeight(1)

	//Setup TokenInfo
	setupTokenInfo(ctx, tk)
	opcu := setupAccounts(ctx, ck, ik, bankKeeper)
	opcuAsset := ik.GetOrNewCUIBCAsset(ctx, sdk.CUTypeOp, opcu.GetAddress())
	ik.SetCUIBCAsset(ctx, opcuAsset)
	opTestcu := newTestCU(ctx, bankKeeper, ik, opcu)

	bankKeeper.SetEvidenceKeeper(evidenceKeeper)

	return testInput{cdc: cdc, ctx: ctx, k: *bankKeeper, ck: ck, tk: tk, ok: ok, ik: ik, trk: bankKeeper, rk: *rk, cn: &mockCN, pk: pk, validators: vals, opcu: opTestcu,
		evidenceKeeper: evidenceKeeper, stakingkeeper: stakingKeeper,
	}
}

//___________________setup tokeninfo_________
func setupTokenInfo(ctx sdk.Context, tk token.Keeper) {
	for _, info := range token.TestTokenData {
		tk.SetToken(ctx, info)
	}
}

//___________________setup accounts___________
func setupAccounts(ctx sdk.Context, ck custodianunit.CUKeeper, ik ibcasset.Keeper, trk transfer.Keeper) exported.CustodianUnit {
	cuAddr, _ := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	cu := ck.NewCUWithAddress(ctx, sdk.CUTypeUser, cuAddr)
	trk.AddCoins(ctx, cu.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(10000000000))))
	//	asset.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(10000000000))))
	ck.SetCU(ctx, cu)

	cuAddr, _ = sdk.CUAddressFromBase58("HBCLG5zCH4FtXi3G6wZps8TNfYYWgzb1Rr2q")
	cu = ck.NewCUWithAddress(ctx, sdk.CUTypeUser, cuAddr)
	asset := ik.GetOrNewCUIBCAsset(ctx, cu.GetCUType(), cu.GetAddress())
	ik.SetCUIBCAsset(ctx, asset)
	trk.AddCoins(ctx, cu.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(20000000000))))
	//cu.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(20000000000))))
	ck.SetCU(ctx, cu)

	cuAddr, _ = sdk.CUAddressFromBase58("HBCPoshPen4yTWCwCvCVuwbfSmrb3EzNbXTo")
	cu = ck.NewOpCUWithAddress(ctx, "btc", cuAddr)
	asset = ik.GetOrNewCUIBCAsset(ctx, cu.GetCUType(), cu.GetAddress())
	ik.SetCUIBCAsset(ctx, asset)
	trk.AddCoins(ctx, cu.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(80000000000))))
	//cu.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(80000000000))))
	ck.SetCU(ctx, cu)

	cuAddr, _ = sdk.CUAddressFromBase58("HBCLXBebMwEWaEZYsqJij7xcpBayzJqdrKJP")
	cu = ck.NewOpCUWithAddress(ctx, "eth", cuAddr)
	//cu.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(80000000000))))
	trk.AddCoins(ctx, cu.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.NativeToken, sdk.NewInt(80000000000))))
	ck.SetCU(ctx, cu)
	return cu
}

//___________________setup validators___________
func setupValidators(ctx sdk.Context, ck custodianunit.CUKeeper) {

}

// nolint: unparam
func createTestPubKeys(numPubKeys int) []crypto.PubKey {
	var publicKeys []crypto.PubKey
	var buffer bytes.Buffer

	//start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") //base pubkey string
		buffer.WriteString(numString)                                                       //adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, newPubKey(buffer.String()))
		buffer.Reset()
	}
	return publicKeys
}

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

func createTestAddrs(numAddrs int) []sdk.CUAddress {
	var addresses []sdk.CUAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, _ := sdk.CUAddressFromHex(buffer.String())
		bech := res.String()
		addresses = append(addresses, testAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}

// for incode address generation
func testAddr(addr string, bech string) sdk.CUAddress {

	res, err := sdk.CUAddressFromHex(addr)
	if err != nil {
		panic(err)
	}
	bechexpected := res.String()
	if bech != bechexpected {
		panic("Bech encoding doesn't match reference")
	}

	//bechres, err := sdk.CUAddressFromBase58(bech)
	bechres, err := sdk.CUAddressFromBase58(bech)

	if err != nil {
		panic(err)
	}
	if !bytes.Equal(bechres, res) {
		panic("Bech decode and hex decode don't match")
	}

	return res
}

func testSetCUCoins(ctx sdk.Context, trk transfer.Keeper, cu sdk.CUAddress, coins sdk.Coins) {
	curCoins := trk.GetAllBalance(ctx, cu)
	trk.SubCoins(ctx, cu, curCoins)
	trk.AddCoins(ctx, cu, coins)
}
