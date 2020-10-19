//nolint
package bhexapp

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/hbtc-chain/bhchain/baseapp"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/distribution"
	"github.com/hbtc-chain/bhchain/x/genaccounts"
	"github.com/hbtc-chain/bhchain/x/gov"
	"github.com/hbtc-chain/bhchain/x/mint"
	"github.com/hbtc-chain/bhchain/x/simulation"
	"github.com/hbtc-chain/bhchain/x/slashing"
	"github.com/hbtc-chain/bhchain/x/staking"
	"github.com/hbtc-chain/bhchain/x/supply"
	"github.com/hbtc-chain/bhchain/x/transfer"
)

// List of available flags for the simulator
var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// NewbhexappUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewbhexappUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, skipUpgradeHeights map[int64]bool, home string, baseAppOptions ...func(*baseapp.BaseApp),
) (gapp *bhexapp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = Newbhexapp(logger, db, traceStore, loadLatest, invCheckPeriod, nil, skipUpgradeHeights, home, baseAppOptions...)
	return gapp, gapp.keys[baseapp.MainStoreKey], gapp.keys[staking.StoreKey], gapp.stakingKeeper
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file
func AppStateFromGenesisFileFn(r *rand.Rand) (tmtypes.GenesisDoc, []simulation.CU) {
	var genesis tmtypes.GenesisDoc
	cdc := MakeCodec()

	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	cdc.MustUnmarshalJSON(bytes, &genesis)

	var appState GenesisState
	cdc.MustUnmarshalJSON(genesis.AppState, &appState)

	accounts := genaccounts.GetGenesisStateFromAppState(cdc, appState)

	newAccs := make([]simulation.CU, len(accounts))
	for i, acc := range accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)

		privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)
		newAccs[i] = simulation.CU{privKey, privKey.PubKey(), acc.Address}
	}

	return genesis, newAccs
}

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	authGenesis := custodianunit.NewGenesisState(
		custodianunit.NewParams(
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.MaxMemoChars, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxMemoChars](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.TxSigLimit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSigLimit](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.TxSizeCostPerByte, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSizeCostPerByte](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.SigVerifyCostED25519, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostED25519](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.SigVerifyCostSECP256K1, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostSECP256K1](r).(uint64)
					})
				return v
			}(r),
		),
		[]exported.CustodianUnit{},
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, authGenesis.Params))
	genesisState[custodianunit.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}

// GenBankGenesisState generates a random GenesisState for bank
func GenBankGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	bankGenesis := transfer.NewGenesisState(
		func(r *rand.Rand) bool {
			var v bool
			ap.GetOrGenerate(cdc, simulation.SendEnabled, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.SendEnabled](r).(bool)
				})
			return v
		}(r),
	)

	fmt.Printf("Selected randomly generated bank parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, bankGenesis))
	genesisState[transfer.ModuleName] = cdc.MustMarshalJSON(bankGenesis)
}

// GenSupplyGenesisState generates a random GenesisState for supply
func GenSupplyGenesisState(cdc *codec.Codec, amount, numInitiallyBonded, numAccs int64, genesisState map[string]json.RawMessage) {
	totalSupply := sdk.NewInt(amount * (numAccs + numInitiallyBonded))
	supplyGenesis := supply.NewGenesisState(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)),
	)

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, supplyGenesis))
	genesisState[supply.ModuleName] = cdc.MustMarshalJSON(supplyGenesis)
}

// GenGenesisAccounts generates a random GenesisState for the genesis accounts
func GenGenesisAccounts(
	cdc *codec.Codec, r *rand.Rand, accs []simulation.CU,
	genesisTimestamp time.Time, amount, numInitiallyBonded int64,
	genesisState map[string]json.RawMessage,
) {

	var genesisAccounts []genaccounts.GenesisCU

	// randomly generate some genesis accounts
	for _, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(amount))}
		bacc := custodianunit.NewBaseCUWithAddress(acc.Address, sdk.CUTypeUser)
		bacc.SetCoins(coins)

		var gacc genaccounts.GenesisCU

		gacc = genaccounts.NewGenesisCU(&bacc)

		genesisAccounts = append(genesisAccounts, gacc)
	}

	genesisState[genaccounts.ModuleName] = cdc.MustMarshalJSON(genesisAccounts)
}

// GenGovGenesisState generates a random GenesisState for gov
func GenGovGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	var vp time.Duration
	ap.GetOrGenerate(cdc, simulation.VotingParamsVotingPeriod, &vp, r,
		func(r *rand.Rand) {
			vp = simulation.ModuleParamSimulator[simulation.VotingParamsVotingPeriod](r).(time.Duration)
		})

	govGenesis := gov.NewGenesisState(
		uint64(r.Intn(100)),
		gov.NewDepositParams(
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(cdc, simulation.DepositParamsMinInitDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinInitDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(cdc, simulation.DepositParamsMinDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(cdc, simulation.DepositParamsMinDaoInitDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinDaoInitDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(cdc, simulation.DepositParamsMinDaoDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinDaoDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			vp,
		),
		gov.NewVotingParams(vp),
		gov.NewTallyParams(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsQuorum, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsQuorum](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsThreshold, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsThreshold](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsVeto, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsVeto](r).(sdk.Dec)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, govGenesis))
	genesisState[gov.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}

/*

func NewParams(mintDenom string, interval sdk.Int, mintNum sdk.Int, blocksPerYear uint64) Params {
	return Params{
		MintDenom:                mintDenom,
		SubsidyReductionInterval: interval,
		BaseMintNum:              mintNum,
		BlocksPerYear:            blocksPerYear,
	}
}
*/

// GenMintGenesisState generates a random GenesisState for mint
func GenMintGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	mintGenesis := mint.NewGenesisState(
		mint.InitialMinter(),
		mint.NewParams(
			sdk.DefaultBondDenom,
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.Inflation, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.Inflation](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Int {
				var v sdk.Int
				ap.GetOrGenerate(cdc, simulation.InitialMintPerBlock, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InitialMintPerBlock](r).(sdk.Int)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, mintGenesis.Params))
	genesisState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
}

// GenDistrGenesisState generates a random GenesisState for distribution
func GenDistrGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	distrGenesis := distribution.GenesisState{
		FeePool: distribution.InitialFeePool(),
		CommunityTax: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.CommunityTax, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.CommunityTax](r).(sdk.Dec)
				})
			return v
		}(r),
		BaseProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.BaseProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BaseProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
		BonusProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.BonusProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BonusProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, distrGenesis))
	genesisState[distribution.ModuleName] = cdc.MustMarshalJSON(distrGenesis)
}

// GenSlashingGenesisState generates a random GenesisState for slashing
func GenSlashingGenesisState(
	cdc *codec.Codec, r *rand.Rand, stakingGen staking.GenesisState,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) {
	slashingGenesis := slashing.NewGenesisState(
		slashing.NewParams(
			stakingGen.Params.UnbondingTime,
			func(r *rand.Rand) int64 {
				var v int64
				ap.GetOrGenerate(cdc, simulation.SignedBlocksWindow, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SignedBlocksWindow](r).(int64)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.MinSignedPerWindow, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MinSignedPerWindow](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(cdc, simulation.DowntimeJailDuration, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DowntimeJailDuration](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.SlashFractionDoubleSign, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SlashFractionDoubleSign](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.SlashFractionDowntime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SlashFractionDowntime](r).(sdk.Dec)
					})
				return v
			}(r),
		),
		nil,
		nil,
	)

	fmt.Printf("Selected randomly generated slashing parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, slashingGenesis.Params))
	genesisState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingGenesis)
}

// GenStakingGenesisState generates a random GenesisState for staking
func GenStakingGenesisState(
	cdc *codec.Codec, r *rand.Rand, accs []simulation.CU, amount, numAccs, numInitiallyBonded int64,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) staking.GenesisState {

	stakingGenesis := staking.NewGenesisState(
		staking.NewParams(
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(cdc, simulation.UnbondingTime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.UnbondingTime](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) uint16 {
				var v uint16
				ap.GetOrGenerate(cdc, simulation.MaxValidators, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxValidators](r).(uint16)
					})
				return v
			}(r),
			11,
			7,
			sdk.DefaultBondDenom,
			1,
			sdk.NewIntWithDecimal(10, 22),
			sdk.NewIntWithDecimal(50, 22),
			10000,
			100,
		),
		nil,
		nil,
	)

	var (
		validators  []staking.Validator
		delegations []staking.Delegation
	)

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := staking.NewValidator(valAddr, accs[i].PubKey, staking.Description{}, true)
		validator.Tokens = sdk.NewInt(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := staking.NewDelegation(accs[i].Address, valAddr, sdk.NewDec(amount))
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis.Validators = validators
	stakingGenesis.Delegations = delegations

	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, stakingGenesis.Params))
	genesisState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	return stakingGenesis
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	log = fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)

	if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
		return
	}

	switch storeName {
	case custodianunit.StoreKey:
		return DecodeAccountStore(cdcA, cdcB, kvA, kvB)
	case mint.StoreKey:
		return DecodeMintStore(cdcA, cdcB, kvA, kvB)
	case staking.StoreKey:
		return DecodeStakingStore(cdcA, cdcB, kvA, kvB)
	case slashing.StoreKey:
		return DecodeSlashingStore(cdcA, cdcB, kvA, kvB)
	case gov.StoreKey:
		return DecodeGovStore(cdcA, cdcB, kvA, kvB)
	case distribution.StoreKey:
		return DecodeDistributionStore(cdcA, cdcB, kvA, kvB)
	case supply.StoreKey:
		return DecodeSupplyStore(cdcA, cdcB, kvA, kvB)
	default:
		return
	}
}

// DecodeAccountStore unmarshals the KVPair's Value to the corresponding auth type
func DecodeAccountStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], custodianunit.AddressStoreKeyPrefix):
		var accA, accB custodianunit.CU
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &accA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &accB)
		return fmt.Sprintf("%v\n%v", accA, accB)
	case bytes.Equal(kvA.Key, custodianunit.GlobalCUNumberKey):
		var globalAccNumberA, globalAccNumberB uint64
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &globalAccNumberA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &globalAccNumberB)
		return fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumberA, globalAccNumberB)
	default:
		panic(fmt.Sprintf("invalid CU key %X", kvA.Key))
	}
}

// DecodeMintStore unmarshals the KVPair's Value to the corresponding mint type
func DecodeMintStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key, mint.MinterKey):
		var minterA, minterB mint.Minter
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &minterA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &minterB)
		return fmt.Sprintf("%v\n%v", minterA, minterB)
	default:
		panic(fmt.Sprintf("invalid mint key %X", kvA.Key))
	}
}

// DecodeDistributionStore unmarshals the KVPair's Value to the corresponding distribution type
func DecodeDistributionStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], distribution.FeePoolKey):
		var feePoolA, feePoolB distribution.FeePool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &feePoolA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &feePoolB)
		return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], distribution.ProposerKey):
		return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorOutstandingRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorWithdrawAddrPrefix):
		return fmt.Sprintf("%v\n%v", sdk.CUAddress(kvA.Value), sdk.CUAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], distribution.DelegatorStartingInfoPrefix):
		var infoA, infoB distribution.DelegatorStartingInfo
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorHistoricalRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB distribution.ValidatorCurrentRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB distribution.ValidatorAccumulatedCommission
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &commissionA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &commissionB)
		return fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], distribution.ValidatorSlashEventPrefix):
		var eventA, eventB distribution.ValidatorSlashEvent
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &eventA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &eventB)
		return fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
	}
}

// DecodeStakingStore unmarshals the KVPair's Value to the corresponding staking type
func DecodeStakingStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], staking.LastTotalPowerKey):
		var powerA, powerB sdk.Int
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &powerA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &powerB)
		return fmt.Sprintf("%v\n%v", powerA, powerB)

	case bytes.Equal(kvA.Key[:1], staking.ValidatorsKey):
		var validatorA, validatorB staking.Validator
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &validatorA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &validatorB)
		return fmt.Sprintf("%v\n%v", validatorA, validatorB)

	case bytes.Equal(kvA.Key[:1], staking.LastValidatorPowerKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByConsAddrKey),
		bytes.Equal(kvA.Key[:1], staking.ValidatorsByPowerIndexKey):
		return fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], staking.DelegationKey):
		var delegationA, delegationB staking.Delegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &delegationA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &delegationB)
		return fmt.Sprintf("%v\n%v", delegationA, delegationB)

	case bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationKey),
		bytes.Equal(kvA.Key[:1], staking.UnbondingDelegationByValIndexKey):
		var ubdA, ubdB staking.UnbondingDelegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &ubdA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &ubdB)
		return fmt.Sprintf("%v\n%v", ubdA, ubdB)

	case bytes.Equal(kvA.Key[:1], staking.RedelegationKey),
		bytes.Equal(kvA.Key[:1], staking.RedelegationByValSrcIndexKey):
		var redA, redB staking.Redelegation
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &redA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &redB)
		return fmt.Sprintf("%v\n%v", redA, redB)

	default:
		panic(fmt.Sprintf("invalid staking key prefix %X", kvA.Key[:1]))
	}
}

// DecodeSlashingStore unmarshals the KVPair's Value to the corresponding slashing type
func DecodeSlashingStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], slashing.ValidatorSigningInfoKey):
		var infoA, infoB slashing.ValidatorSigningInfo
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], slashing.ValidatorMissedBlockBitArrayKey):
		var missedA, missedB bool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &missedA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &missedB)
		return fmt.Sprintf("missedA: %v\nmissedB: %v", missedA, missedB)

	case bytes.Equal(kvA.Key[:1], slashing.AddrPubkeyRelationKey):
		var pubKeyA, pubKeyB crypto.PubKey
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &pubKeyA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &pubKeyB)
		bechPKA := sdk.MustBech32ifyAccPub(pubKeyA)
		bechPKB := sdk.MustBech32ifyAccPub(pubKeyB)
		return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPKA, bechPKB)

	default:
		panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
	}
}

// DecodeGovStore unmarshals the KVPair's Value to the corresponding gov type
func DecodeGovStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], gov.ProposalsKeyPrefix):
		var proposalA, proposalB gov.Proposal
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &proposalA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &proposalB)
		return fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], gov.ActiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.InactiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], gov.ProposalIDKey):
		proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
		proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
		return fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

	case bytes.Equal(kvA.Key[:1], gov.DepositsKeyPrefix):
		var depositA, depositB gov.Deposit
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &depositA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &depositB)
		return fmt.Sprintf("%v\n%v", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], gov.VotesKeyPrefix):
		var voteA, voteB gov.Vote
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &voteA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &voteB)
		return fmt.Sprintf("%v\n%v", voteA, voteB)

	default:
		panic(fmt.Sprintf("invalid governance key prefix %X", kvA.Key[:1]))
	}
}

// DecodeSupplyStore unmarshals the KVPair's Value to the corresponding supply type
func DecodeSupplyStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], supply.SupplyKey):
		var supplyA, supplyB supply.Supply
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
		return fmt.Sprintf("%v\n%v", supplyB, supplyB)
	default:
		panic(fmt.Sprintf("invalid supply key %X", kvA.Key))
	}
}
