package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/hbtc-chain/bhchain/x/custodianunit/internal"

	"github.com/hbtc-chain/bhchain/baseapp"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
	"github.com/hbtc-chain/bhchain/x/simulation"
)

// SimulateDeductFee
func SimulateDeductFee(ak custodianunit.CUKeeper, supplyKeeper internal.SupplyKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		cus []simulation.CU) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		CU := simulation.RandomAcc(r, cus)
		stored := ak.GetCU(ctx, CU.Address)
		initCoins := stored.GetCoins()
		opMsg = simulation.NewOperationMsgBasic(types.ModuleName, "deduct_fee", "", false, nil)

		feeCollector := ak.GetCU(ctx, supplyKeeper.GetModuleAddress(types.FeeCollectorName))
		if feeCollector == nil {
			panic(fmt.Errorf("fee collector CU hasn't been set"))
		}

		if len(initCoins) == 0 {
			return opMsg, nil, nil
		}

		denomIndex := r.Intn(len(initCoins))
		randCoin := initCoins[denomIndex]

		amt, err := randPositiveInt(r, randCoin.Amount)
		if err != nil {
			return opMsg, nil, nil
		}

		// Create a random fee and verify the fees are within the CU's spendable
		// balance.
		fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))
		spendableCoins := stored.GetCoins()
		if _, hasNeg := spendableCoins.SafeSub(fees); hasNeg {
			return opMsg, nil, nil
		}

		// get the new CU balance
		_, hasNeg := initCoins.SafeSub(fees)
		if hasNeg {
			return opMsg, nil, nil
		}

		_, err = supplyKeeper.SendCoinsFromAccountToModule(ctx, stored.GetAddress(), types.FeeCollectorName, fees)
		if err != nil {
			panic(err)
		}

		opMsg.OK = true
		return opMsg, nil, nil
	}
}

func randPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GT(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}
	max = max.Sub(sdk.OneInt())
	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}
