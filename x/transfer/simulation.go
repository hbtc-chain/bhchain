package transfer

import (
	"fmt"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/hbtc-chain/bhchain/baseapp"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/mock"
	"github.com/hbtc-chain/bhchain/x/simulation"
	"github.com/hbtc-chain/bhchain/x/transfer/keeper"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

// SendTx tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSend(mapper types.CUKeeper, bk keeper.Keeper) simulation.Operation {
	handler := NewHandler(bk)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.CU) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		fromAcc, comment, msg, ok := createMsgSend(r, ctx, accs, mapper)
		opMsg = simulation.NewOperationMsg(msg, ok, comment)
		if !ok {
			return opMsg, nil, nil
		}
		err = sendAndVerifyMsgSend(app, mapper, msg, ctx, []crypto.PrivKey{fromAcc.PrivKey}, handler)
		if err != nil {
			return opMsg, nil, err
		}
		return opMsg, nil, nil
	}
}

func createMsgSend(r *rand.Rand, ctx sdk.Context, accs []simulation.CU, mapper types.CUKeeper) (
	fromAcc simulation.CU, comment string, msg types.MsgSend, ok bool) {

	fromAcc = simulation.RandomAcc(r, accs)
	toAcc := simulation.RandomAcc(r, accs)
	// Disallow sending money to yourself
	for {
		if !fromAcc.PubKey.Equals(toAcc.PubKey) {
			break
		}
		toAcc = simulation.RandomAcc(r, accs)
	}
	initFromCoins := mapper.GetCU(ctx, fromAcc.Address).GetCoins()

	if len(initFromCoins) == 0 {
		return fromAcc, "skipping, no coins at all", msg, false
	}

	denomIndex := r.Intn(len(initFromCoins))
	amt, goErr := simulation.RandPositiveInt(r, initFromCoins[denomIndex].Amount)
	if goErr != nil {
		return fromAcc, "skipping bank send due to CU having no coins of denomination " + initFromCoins[denomIndex].Denom, msg, false
	}

	coins := sdk.Coins{sdk.NewCoin(initFromCoins[denomIndex].Denom, amt)}
	msg = types.NewMsgSend(fromAcc.Address, toAcc.Address, coins)
	return fromAcc, "", msg, true
}

// Sends and verifies the transition of a msg send.
func sendAndVerifyMsgSend(app *baseapp.BaseApp, mapper types.CUKeeper, msg types.MsgSend, ctx sdk.Context, privkeys []crypto.PrivKey, handler sdk.Handler) error {
	fromAcc := mapper.GetCU(ctx, msg.FromAddress)
	SequenceNumbers := []uint64{fromAcc.GetSequence()}
	initialFromAddrCoins := fromAcc.GetCoins()

	toAcc := mapper.GetCU(ctx, msg.ToAddress)
	initialToAddrCoins := toAcc.GetCoins()

	if handler != nil {
		res := handler(ctx, msg)
		if !res.IsOK() {
			if res.Code == types.CodeSendDisabled {
				return nil
			}
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("handling msg failed %v", res)
		}
	} else {
		tx := mock.GenTx([]sdk.Msg{msg},
			SequenceNumbers,
			privkeys...)
		res := app.Deliver(tx)
		if !res.IsOK() {
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("Deliver failed %v", res)
		}
	}

	fromAcc = mapper.GetCU(ctx, msg.FromAddress)
	toAcc = mapper.GetCU(ctx, msg.ToAddress)

	if !initialFromAddrCoins.Sub(msg.Amount).IsEqual(fromAcc.GetCoins()) {
		return fmt.Errorf("fromAddress %s had an incorrect amount of coins", fromAcc.GetAddress())
	}

	if !initialToAddrCoins.Add(msg.Amount).IsEqual(toAcc.GetCoins()) {
		return fmt.Errorf("toAddress %s had an incorrect amount of coins", toAcc.GetAddress())
	}

	return nil
}

// SingleInputSendMsg tests and runs a single msg multisend, with one input and one output, where both
// accounts already exist.
func SimulateSingleInputMsgMultiSend(mapper types.CUKeeper, bk keeper.Keeper) simulation.Operation {
	handler := NewHandler(bk)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.CU) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		fromAcc, comment, msg, ok := createSingleInputMsgMultiSend(r, ctx, accs, mapper)
		opMsg = simulation.NewOperationMsg(msg, ok, comment)
		if !ok {
			return opMsg, nil, nil
		}
		err = sendAndVerifyMsgMultiSend(app, mapper, msg, ctx, []crypto.PrivKey{fromAcc.PrivKey}, handler)
		if err != nil {
			return opMsg, nil, err
		}
		return opMsg, nil, nil
	}
}

func createSingleInputMsgMultiSend(r *rand.Rand, ctx sdk.Context, accs []simulation.CU, mapper types.CUKeeper) (
	fromAcc simulation.CU, comment string, msg types.MsgMultiSend, ok bool) {

	fromAcc = simulation.RandomAcc(r, accs)
	toAcc := simulation.RandomAcc(r, accs)
	// Disallow sending money to yourself
	for {
		if !fromAcc.PubKey.Equals(toAcc.PubKey) {
			break
		}
		toAcc = simulation.RandomAcc(r, accs)
	}
	toAddr := toAcc.Address
	initFromCoins := mapper.GetCU(ctx, fromAcc.Address).GetCoins()

	if len(initFromCoins) == 0 {
		return fromAcc, "skipping, no coins at all", msg, false
	}

	denomIndex := r.Intn(len(initFromCoins))
	amt, goErr := simulation.RandPositiveInt(r, initFromCoins[denomIndex].Amount)
	if goErr != nil {
		return fromAcc, "skipping bank send due to CU having no coins of denomination " + initFromCoins[denomIndex].Denom, msg, false
	}

	coins := sdk.Coins{sdk.NewCoin(initFromCoins[denomIndex].Denom, amt)}
	msg = types.MsgMultiSend{
		Inputs:  []types.Input{types.NewInput(fromAcc.Address, coins)},
		Outputs: []types.Output{types.NewOutput(toAddr, coins)},
	}
	return fromAcc, "", msg, true
}

// Sends and verifies the transition of a msg multisend. This fails if there are repeated inputs or outputs
// pass in handler as nil to handle txs, otherwise handle msgs
func sendAndVerifyMsgMultiSend(app *baseapp.BaseApp, mapper types.CUKeeper, msg types.MsgMultiSend,
	ctx sdk.Context, privkeys []crypto.PrivKey, handler sdk.Handler) error {

	initialInputAddrCoins := make([]sdk.Coins, len(msg.Inputs))
	initialOutputAddrCoins := make([]sdk.Coins, len(msg.Outputs))
	SequenceNumbers := make([]uint64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		acc := mapper.GetCU(ctx, msg.Inputs[i].Address)
		SequenceNumbers[i] = acc.GetSequence()
		initialInputAddrCoins[i] = acc.GetCoins()
	}
	for i := 0; i < len(msg.Outputs); i++ {
		acc := mapper.GetCU(ctx, msg.Outputs[i].Address)
		initialOutputAddrCoins[i] = acc.GetCoins()
	}
	if handler != nil {
		res := handler(ctx, msg)
		if !res.IsOK() {
			if res.Code == types.CodeSendDisabled {
				return nil
			}
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("handling msg failed %v", res)
		}
	} else {
		tx := mock.GenTx([]sdk.Msg{msg},
			SequenceNumbers,
			privkeys...)
		res := app.Deliver(tx)
		if !res.IsOK() {
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("Deliver failed %v", res)
		}
	}

	for i := 0; i < len(msg.Inputs); i++ {
		terminalInputCoins := mapper.GetCU(ctx, msg.Inputs[i].Address).GetCoins()
		if !initialInputAddrCoins[i].Sub(msg.Inputs[i].Coins).IsEqual(terminalInputCoins) {
			return fmt.Errorf("input #%d had an incorrect amount of coins", i)
		}
	}
	for i := 0; i < len(msg.Outputs); i++ {
		terminalOutputCoins := mapper.GetCU(ctx, msg.Outputs[i].Address).GetCoins()
		if !terminalOutputCoins.IsEqual(initialOutputAddrCoins[i].Add(msg.Outputs[i].Coins)) {
			return fmt.Errorf("output #%d had an incorrect amount of coins", i)
		}
	}
	return nil
}
