package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/hbtc-chain/bhchain/types"
	. "github.com/hbtc-chain/bhchain/x/custodianunit"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

//-- cases for CUKeeperI functions ----

func TestCUKeeper_NewCU(t *testing.T) {
	input := setupTestInputForCUKeeper()
	ctx := input.Ctx
	cuKeeper := input.Ck

	// test NewCU do not change argument type to keeper.proto'S return type
	// should assign cunumber
	// with other fileds no changes
	cuTest := cuTypeForTest{}
	cuKeeper.NewCU(ctx, &BaseCU{})
	cuTestGot := cuKeeper.NewCU(ctx, &cuTest).(*cuTypeForTest)
	assert.EqualValues(t, cuTypeForTest{CUNumber: 0}, *cuTestGot)

	// test NewCUWithAddress
	// should create CU with given CUType
	// with given address
	// with sequnce==0
	// with other fields empty
	addr := sdk.CUAddress([]byte("some-address"))
	cu := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, addr)
	assert.NotNil(t, cu)
	assert.EqualValues(t, sdk.CUTypeUser, cu.GetCUType())
	assert.EqualValues(t, 0, cu.GetSequence())
	assert.EqualValues(t, addr, cu.GetAddress())
	assert.EqualValues(t, 0, len(cu.GetAssets()))
	assert.EqualValues(t, nil, cu.GetPubKey())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetCoins())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetGasUsed())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetGasReceived())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetAssetCoins())

	// test NewCUWithPubkey
	// should create CU with given CUType
	// with publicKey == given pubKey
	// with a address == pubKey.Address()
	// with sequnce==0
	// with other fields empty
	pubKey := ed25519.GenPrivKey().PubKey()
	addr2 := sdk.CUAddressFromPubKey(pubKey)
	cu = cuKeeper.NewCUWithPubkey(ctx, sdk.CUTypeUser, pubKey)
	assert.NotNil(t, cu)
	assert.EqualValues(t, sdk.CUTypeUser, cu.GetCUType())
	assert.EqualValues(t, 0, cu.GetSequence())
	assert.EqualValues(t, addr2, cu.GetAddress())
	assert.EqualValues(t, addr2, cu.GetAddress())
	assert.EqualValues(t, pubKey, cu.GetPubKey())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetCoins())
	assert.EqualValues(t, 0, len(cu.GetAssets()))
	assert.EqualValues(t, sdk.Coins(nil), cu.GetGasUsed())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetGasReceived())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetAssetCoins())

	// test SetCU & GetCU
	cuKeeper.SetCU(ctx, cu)
	cuGot := cuKeeper.GetCU(ctx, sdk.CUAddressFromPubKey(pubKey))
	assert.EqualValues(t, cu, cuGot)
	// get a none exist CU should return exported.CustodianUnit(nil)
	cuGotNoExist := cuKeeper.GetCU(ctx, sdk.CUAddress(ed25519.GenPrivKey().PubKey().Address()))
	assert.EqualValues(t, exported.CustodianUnit(nil), cuGotNoExist)

	// test GetOrNewCU
	// if CU with given address not exist , create CU with given address & CUType
	// with publicKey == given pubKey
	// with a address == pubKey.Address()
	// with sequnce==0
	// with other fields empty
	pubKey = ed25519.GenPrivKey().PubKey()
	addrNew := sdk.NewCUAddress()
	// get a none exist CU will create it
	cu = cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, addrNew)
	assert.NotNil(t, cu)
	assert.EqualValues(t, sdk.CUTypeUser, cu.GetCUType())
	assert.EqualValues(t, 0, cu.GetSequence())
	assert.EqualValues(t, addrNew, cu.GetAddress())
	assert.EqualValues(t, addrNew, cu.GetAddress())
	assert.EqualValues(t, nil, cu.GetPubKey())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetCoins())
	// if cutype != CUTypeUser never create new one
	cu = cuKeeper.GetOrNewCU(ctx, sdk.CUTypeOp, sdk.NewCUAddress())
	assert.Nil(t, cu)
	// invalid cu type
	cu = cuKeeper.GetOrNewCU(ctx, 8, sdk.NewCUAddress())
	assert.Nil(t, cu)
	// address is nil
	cu = cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, nil)
	assert.Nil(t, cu)
	// get a exist CU just return it (the cu created by above test case for NewCUWithPubkey)
	cu = cuKeeper.GetOrNewCU(ctx, sdk.CUTypeUser, sdk.CUAddressFromByte(addr2))
	assert.NotNil(t, cu)
	assert.EqualValues(t, sdk.CUAddressFromByte(addr2), cu.GetAddress())

}

func TestGetSetCU(t *testing.T) {
	input := setupTestInput()
	cuKeeper := input.ak
	ctx := input.ctx
	addr := sdk.NewCUAddress()

	// no CU before its created
	cu := cuKeeper.GetCU(ctx, addr)
	assert.Nil(t, cu)

	// create CU and check default values
	cu = cuKeeper.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
	assert.NotNil(t, cu)
	assert.Equal(t, addr, cu.GetAddress())
	assert.EqualValues(t, nil, cu.GetPubKey())
	assert.EqualValues(t, 0, cu.GetSequence())
	assert.EqualValues(t, 0, len(cu.GetAssets()))
	assert.EqualValues(t, nil, cu.GetPubKey())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetCoins())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetGasUsed())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetGasReceived())
	assert.EqualValues(t, sdk.Coins(nil), cu.GetAssetCoins())

	// NewCU doesn't call Set, so it's still nil
	assert.Nil(t, cuKeeper.GetCU(ctx, addr))

	// set some values on the CU and save it
	newSequence := uint64(20)
	cu.SetSequence(newSequence)
	cuKeeper.SetCU(ctx, cu)

	// check the new values
	cu = cuKeeper.GetCU(ctx, addr)
	assert.NotNil(t, cu)
	assert.Equal(t, newSequence, cu.GetSequence())

	// set opcu with used address
	opcu := BaseCU{Type: sdk.CUTypeOp, Address: addr}
	assert.Panics(t, func() {
		cuKeeper.SetCU(ctx, &opcu)
	})
	// set opcu without symbol
	opcuaddr := sdk.NewCUAddress()
	opcu = BaseCU{Type: sdk.CUTypeOp, Address: opcuaddr}
	assert.Panics(t, func() {
		cuKeeper.SetCU(ctx, &opcu)
	})
	// set opcu without address
	opcu = BaseCU{Type: sdk.CUTypeOp, Address: nil}
	assert.Panics(t, func() {
		cuKeeper.SetCU(ctx, &opcu)
	})

	// should ok, set opcu with 1 assets
	opcu = BaseCU{Type: sdk.CUTypeOp, Address: opcuaddr}
	opcu.SetSymbol(btcToken, 0)
	cuKeeper.SetCU(ctx, &opcu)
	opcuGot := cuKeeper.GetOpCUs(ctx, btcToken)
	assert.EqualValues(t, opcu, *(opcuGot[0]).(*BaseCU))
	assert.EqualValues(t, btcToken, opcuGot[0].GetSymbol())

	// should ok set opcu with 2 assets in same chain, subtoken first, mainnet token second
	opcu = BaseCU{Type: sdk.CUTypeOp, Address: opcuaddr}
	opcu.AddAsset(usdtToken, "", 0)
	opcu.AddAsset(ethToken, "", 0)
	cuKeeper.SetCU(ctx, &opcu)
	opcuGot = cuKeeper.GetOpCUs(ctx, btcToken)
	assert.EqualValues(t, opcu, *(opcuGot[0]).(*BaseCU))
	assert.EqualValues(t, usdtToken, opcuGot[0].GetSymbol())

	// GetOpCUs
	cuKeeper.SetCU(ctx, cuKeeper.NewOpCUWithAddress(ctx, btcToken, sdk.NewCUAddress()))
	opcuGot = cuKeeper.GetOpCUs(ctx, btcToken)
	assert.EqualValues(t, 2, len(opcuGot))
}

func TestCUKeeper_NewOpCUWithAddress(t *testing.T) {
	input := setupTestInputForCUKeeper()
	ctx := input.Ctx
	cuKeeper := input.Ck

	addr1 := sdk.NewCUAddress()
	opcu := cuKeeper.NewOpCUWithAddress(ctx, ethToken, addr1)
	assert.NotNil(t, sdk.CUTypeOp, opcu.GetCUType())
	assert.Equal(t, ethToken, opcu.GetSymbol())
	assert.Equal(t, "", opcu.GetAssetAddress(ethToken, 0))
	assert.True(t, opcu.GetAddress().IsValidAddr())
	cuKeeper.SetCU(ctx, opcu)
	// address has been used
	opcu2 := cuKeeper.NewOpCUWithAddress(ctx, ethToken, addr1)
	assert.EqualValues(t, nil, opcu2)

	// symbol == ""
	opcu2 = cuKeeper.NewOpCUWithAddress(ctx, "", addr1)
	assert.EqualValues(t, nil, opcu2)
	// address == nil
	opcu2 = cuKeeper.NewOpCUWithAddress(ctx, ethToken, nil)
	assert.EqualValues(t, nil, opcu2)
	// address == nil && symbol == ""
	opcu2 = cuKeeper.NewOpCUWithAddress(ctx, ethToken, nil)
	assert.EqualValues(t, nil, opcu2)
	// symbol not supported
	opcu2 = cuKeeper.NewOpCUWithAddress(ctx, "notsupport", addr1)
	assert.EqualValues(t, nil, opcu2)

}

func TestCUKeeper_DepositList_OpCU(t *testing.T) {
	input := setupTestInputForCUKeeper()
	ctx := input.Ctx
	cuKeeper := input.Ck

	cu := cuKeeper.NewOpCUWithAddress(ctx, btcToken, sdk.NewCUAddress())
	// depositList not exist
	dlsGot := cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.Nil(t, dlsGot)

	// SetDepositList
	d1, _ := sdk.NewDepositItem("hash1", 1, sdk.NewInt(10), "", "memo1", 0)
	d2, _ := sdk.NewDepositItem("hash2", 2, sdk.NewInt(20), "", "memo2", 0)
	dls, _ := sdk.NewDepositList(d1, d2)
	cuKeeper.SetCU(ctx, cu)
	cuKeeper.SetDepositList(ctx, btcToken, cu.GetAddress(), dls)
	dlsGot = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.EqualValues(t, dls, dlsGot)

	// AddDeposit
	d3, _ := sdk.NewDepositItem("hash3", 3, sdk.NewInt(30), "", "memo3", 0)
	err := cuKeeper.SaveDeposit(ctx, btcToken, cu.GetAddress(), d3)
	assert.Nil(t, err)
	dlsGot = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.Equal(t, 3, dlsGot.Len())
	// add duplicate deposit
	err = cuKeeper.SaveDeposit(ctx, btcToken, cu.GetAddress(), d3)
	assert.Nil(t, err)
	dlsGot = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.Equal(t, 3, dlsGot.Len())

	// GetDepositListByHash
	d4, _ := sdk.NewDepositItem("hash2", 4, sdk.NewInt(40), "", "memo4", 0)
	err = cuKeeper.SaveDeposit(ctx, btcToken, cu.GetAddress(), d4)
	assert.Nil(t, err)
	dlsGot = cuKeeper.GetDepositListByHash(ctx, btcToken, cu.GetAddress(), "hash2")
	assert.Equal(t, 2, dlsGot.Len())
	// GetDepositByHashIndex
	dGot := cuKeeper.GetDeposit(ctx, btcToken, cu.GetAddress(), "hash2", 2)
	require.NoError(t, err)
	assert.EqualValues(t, d2, dGot)

	// should ok
	cuKeeper.DelDeposit(ctx, btcToken, cu.GetAddress(), "hash2", 4)

	// IsDepositExist d4
	b := cuKeeper.IsDepositExist(ctx, btcToken, cu.GetAddress(), "hash2", 4)
	assert.False(t, b)
	b = cuKeeper.IsDepositExist(ctx, btcToken, cu.GetAddress(), "hash2", 2)
	assert.True(t, b)
	// wrong symbol
	b = cuKeeper.IsDepositExist(ctx, "eth", cu.GetAddress(), "hash2", 2)
	assert.False(t, b)
}

func TestCUKeeper_DepositList(t *testing.T) {
	input := setupTestInputForCUKeeper()
	ctx := input.Ctx
	cuKeeper := input.Ck

	cu := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, sdk.NewCUAddress())
	// depositList not exist
	dlsGot := cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.Nil(t, dlsGot)

	// SetDepositList
	d1, _ := sdk.NewDepositItem("hash1", 1, sdk.NewInt(10), "", "memo1", 0)
	d2, _ := sdk.NewDepositItem("hash2", 2, sdk.NewInt(20), "", "memo2", 0)
	dls, _ := sdk.NewDepositList(d1, d2)
	cuKeeper.SetCU(ctx, cu)
	cuKeeper.SetDepositList(ctx, btcToken, cu.GetAddress(), dls)
	dlsGot = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.EqualValues(t, dls, dlsGot)
	// second depositList in cu,no use
	cuKeeper.SetDepositList(ctx, "eth", cu.GetAddress(), dls)

	// AddDeposit
	d3, _ := sdk.NewDepositItem("hash3", 3, sdk.NewInt(30), "", "memo3", 0)
	err := cuKeeper.SaveDeposit(ctx, btcToken, cu.GetAddress(), d3)
	assert.Nil(t, err)
	dlsGot = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.Equal(t, 3, dlsGot.Len())
	dlsGot = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	assert.Equal(t, 3, dlsGot.Len())

	// GetDepositListByHash
	d4, _ := sdk.NewDepositItem("hash2", 4, sdk.NewInt(40), "", "memo4", 0)
	err = cuKeeper.SaveDeposit(ctx, btcToken, cu.GetAddress(), d4)
	assert.Nil(t, err)
	dlsGot = cuKeeper.GetDepositListByHash(ctx, btcToken, cu.GetAddress(), "hash2")
	assert.Equal(t, 2, dlsGot.Len())
	// GetDepositByHashIndex
	dGot := cuKeeper.GetDeposit(ctx, btcToken, cu.GetAddress(), "hash2", 2)
	require.NoError(t, err)
	assert.EqualValues(t, d2, dGot)

	// should ok
	cuKeeper.DelDeposit(ctx, btcToken, cu.GetAddress(), "hash2", 4)

	// IsDepositExist d4
	b := cuKeeper.IsDepositExist(ctx, btcToken, cu.GetAddress(), "hash2", 4)
	assert.False(t, b)
	b = cuKeeper.IsDepositExist(ctx, btcToken, cu.GetAddress(), "hash2", 2)
	assert.True(t, b)
	// wrong symbol
	b = cuKeeper.IsDepositExist(ctx, "btt", cu.GetAddress(), "hash2", 2)
	assert.False(t, b)

	// del all item in list
	dls = cuKeeper.GetDepositList(ctx, btcToken, cu.GetAddress())
	for _, d := range dls {
		cuKeeper.DelDeposit(ctx, btcToken, cu.GetAddress(), d.Hash, d.Index)
	}
}

func TestRemoveCU(t *testing.T) {
	input := setupTestInput()
	addr1 := sdk.CUAddress([]byte("addr1"))
	addr2 := sdk.CUAddress([]byte("addr2"))

	// create CUs
	cu1 := input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr1)
	cu2 := input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr2)

	cuSeq1 := uint64(20)
	cuSeq2 := uint64(40)

	cu1.SetSequence(cuSeq1)
	cu2.SetSequence(cuSeq2)
	input.ak.SetCU(input.ctx, cu1)
	input.ak.SetCU(input.ctx, cu2)

	// GetAllCUs
	assert.Equal(t, 2, len(input.ak.GetAllCUs(input.ctx)))

	cu1 = input.ak.GetCU(input.ctx, addr1)
	assert.NotNil(t, cu1)
	assert.Equal(t, cuSeq1, cu1.GetSequence())

	// remove one CU
	input.ak.RemoveCU(input.ctx, cu1)
	assert.Equal(t, 1, len(input.ak.GetAllCUs(input.ctx)))
	cu1 = input.ak.GetCU(input.ctx, addr1)
	assert.Nil(t, cu1)

	cu2 = input.ak.GetCU(input.ctx, addr2)
	assert.NotNil(t, cu2)
	assert.Equal(t, cuSeq2, cu2.GetSequence())
}

func TestSetParams(t *testing.T) {
	input := setupTestInput()
	params := DefaultParams()

	input.ak.SetParams(input.ctx, params)

	newParams := Params{}
	input.ak.ParamSubspace.Get(input.ctx, KeyTxSigLimit, &newParams.TxSigLimit)
	assert.Equal(t, newParams.TxSigLimit, DefaultTxSigLimit)
}

func TestGetParams(t *testing.T) {
	input := setupTestInput()
	params := DefaultParams()

	input.ak.SetParams(input.ctx, params)

	newParams := input.ak.GetParams(input.ctx)
	assert.Equal(t, params, newParams)
}

func TestCUKeeper_GetOpCUsInfo(t *testing.T) {
	input := setupTestInputForCUKeeper()
	ctx := input.Ctx
	cuKeeper := input.Ck
	opcu := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeOp, sdk.NewCUAddress())
	opcu.AddAsset("btc", "btcaddress1", 0)

	d1, _ := sdk.NewDepositItem("hash1", 1, sdk.NewInt(10), "", "memo1", 0)
	d2, _ := sdk.NewDepositItem("hash2", 2, sdk.NewInt(20), "", "memo2", 0)
	dls, _ := sdk.NewDepositList(d1, d2)
	cuKeeper.SetCU(ctx, opcu)
	cuKeeper.SetDepositList(ctx, "btc", opcu.GetAddress(), dls)

	curecordGot := cuKeeper.GetOpCUsInfo(ctx, "btc")
	assert.Equal(t, 1, len(curecordGot))
	assert.Equal(t, 2, len(curecordGot[0].DepositList))

}

func TestCUKeeper_BalanceFlows(t *testing.T) {
	input := setupTestInput()
	cuKeeper := input.ak
	ctx := input.ctx
	addr := sdk.NewCUAddress()
	cu := cuKeeper.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("coin1", 123), sdk.NewInt64Coin("coin2", 246), sdk.NewInt64Coin("coin0", 0)}

	cu.AddCoins(someCoins)
	assert.EqualValues(t, 2, len(cu.GetBalanceFlows()))

	cuKeeper.SetCU(ctx, cu)

	cuGot := cuKeeper.GetCU(ctx, addr)
	assert.NotNil(t, cuGot)
	// GetCU must return cu without balanceFlows
	assert.EqualValues(t, 0, len(cuGot.GetBalanceFlows()))

}

func TestCUKeeper_BalanceFlows1(t *testing.T) {
	input := setupTestInput()
	cuKeeper := input.ak
	ctx := input.ctx
	addr := sdk.NewCUAddress()
	cu := cuKeeper.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)

	initCoins := sdk.Coins{sdk.NewInt64Coin("btc", 10000)}

	cu.AddCoins(initCoins)
	assert.EqualValues(t, 1, len(cu.GetBalanceFlows()))
	assert.Equal(t, 1, len(cu.GetBalanceFlows()))
	balanceFlow := cu.GetBalanceFlows()[0]
	assert.Equal(t, sdk.ZeroInt(), balanceFlow.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), balanceFlow.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(10000), balanceFlow.BalanceChange)
	assert.Equal(t, sdk.ZeroInt(), balanceFlow.BalanceOnHoldChange)

	cuKeeper.SetCU(ctx, cu)
	assert.Equal(t, sdk.NewInt(10000), cu.GetCoins().AmountOf("btc"))

	//withdrawal
	cu = cuKeeper.GetCU(ctx, addr)
	withdrawalCoins := sdk.Coins{sdk.NewInt64Coin("btc", 3000)}
	cu.SubCoins(withdrawalCoins)
	cu.AddCoinsHold(withdrawalCoins)
	assert.Equal(t, 1, len(cu.GetBalanceFlows()))
	balanceFlow = cu.GetBalanceFlows()[0]
	assert.Equal(t, sdk.NewInt(10000), balanceFlow.PreviousBalance)
	assert.Equal(t, sdk.ZeroInt(), balanceFlow.PreviousBalanceOnHold)
	assert.Equal(t, sdk.NewInt(-3000), balanceFlow.BalanceChange)
	assert.Equal(t, sdk.NewInt(3000), balanceFlow.BalanceOnHoldChange)
	cuKeeper.SetCU(ctx, cu)

	//withdrawal finish
	cu = cuKeeper.GetCU(ctx, addr)
	cu.SubCoinsHold(withdrawalCoins)
	assert.Equal(t, 1, len(cu.GetBalanceFlows()))
	balanceFlow = cu.GetBalanceFlows()[0]
	assert.Equal(t, sdk.NewInt(7000), balanceFlow.PreviousBalance)
	assert.Equal(t, sdk.NewInt(3000), balanceFlow.PreviousBalanceOnHold)
	assert.Equal(t, sdk.ZeroInt(), balanceFlow.BalanceChange)
	assert.Equal(t, sdk.NewInt(-3000), balanceFlow.BalanceOnHoldChange)

}

func TestCUKeeper_BalanceFlowsWhenSetCoins(t *testing.T) {
	input := setupTestInput()
	cuKeeper := input.ak
	ctx := input.ctx
	addr := sdk.NewCUAddress()
	cu := cuKeeper.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)

	someCoins := sdk.NewCoins(sdk.NewInt64Coin("coin2", 246), sdk.NewInt64Coin("coin0", 0), sdk.NewInt64Coin("coin1", 123))

	cu.SetCoins(someCoins)
	assert.EqualValues(t, 2, len(cu.GetBalanceFlows()))

	cuKeeper.SetCU(ctx, cu)

	cuGot := cuKeeper.GetCU(ctx, addr)
	assert.NotNil(t, cuGot)
	// GetCU must return cu without balanceFlows
	assert.EqualValues(t, 0, len(cuGot.GetBalanceFlows()))

}

func TestSetGetExtAddresseWithCU(t *testing.T) {
	input := setupTestInputForCUKeeper()
	ctx := input.Ctx
	cuKeeper := input.Ck

	cu := cuKeeper.NewCUWithAddress(ctx, sdk.CUTypeUser, sdk.NewCUAddress())

	// invalid args
	cuGot, err := cuKeeper.GetCUFromExtAddress(ctx, "", "0x3443aDbe92F0AA15FDf9e63F301F1440b341f053")
	assert.NotNil(t, err)
	assert.Nil(t, cuGot)

	cuGot, err = cuKeeper.GetCUFromExtAddress(ctx, ethToken, "")
	assert.NotNil(t, err)
	assert.Nil(t, cuGot)

	// extAddress not exist
	cuGot, err = cuKeeper.GetCUFromExtAddress(ctx, ethToken, "0x3443aDbe92F0AA15FDf9e63F301F1440b341f053")
	assert.NotNil(t, err)
	assert.Nil(t, cuGot)

	// SetExtAddresseWithCU
	cuKeeper.SetExtAddresseWithCU(ctx, ethToken, "0x3443aDbe92F0AA15FDf9e63F301F1440b341f053", cu.GetAddress())
	cuGot, err = cuKeeper.GetCUFromExtAddress(ctx, ethToken, "0x3443aDbe92F0AA15FDf9e63F301F1440b341f053")
	assert.Nil(t, err)
	assert.NotNil(t, cuGot)
	assert.EqualValues(t, cu.GetAddress(), cuGot)

}

//----------- utils ----------------
var _ exported.CustodianUnit = (*cuTypeForTest)(nil)

type cuTypeForTest struct {
	CUNumber uint64
}

func (ct *cuTypeForTest) GetBalanceFlows() []sdk.BalanceFlow {
	panic("implement me")
}

func (ct *cuTypeForTest) ResetBalanceFlows() {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAssetPubkey(epoch uint64) []byte {
	panic("implement me")
}

func (ct *cuTypeForTest) SetAssetPubkey(pubkey []byte, epoch uint64) error {
	panic("implement me")
}

func (ct *cuTypeForTest) IsEnabledSendTx(chain string, addr string) bool {
	panic("implement me")
}

func (ct *cuTypeForTest) SetEnableSendTx(enabled bool, chain string, addr string) {
	panic("implement me")
}

func (ct *cuTypeForTest) Validate() error {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAssetCoinsHold() sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) AddAssetCoinsHold(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SubAssetCoinsHold(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAssetCoins() sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) AddAssetCoins(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SubAssetCoins(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) GetGasUsed() sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) AddGasUsed(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SubGasUsed(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) GetGasReceived() sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) AddGasReceived(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SubGasReceived(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) GetGasRemained(chain string, addr string) sdk.Int {
	panic("implement me")
}
func (ct *cuTypeForTest) AddGasRemained(chain string, addr string, amt sdk.Int) {
	panic("implement me")
}
func (ct *cuTypeForTest) SubGasRemained(chain string, addr string, amt sdk.Int) {
	panic("implement me")
}

func (ct *cuTypeForTest) AddCoins(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SubCoins(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) AddCoinsHold(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SubCoinsHold(coins sdk.Coins) sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAssetAddress(denom string, epoch uint64) string {
	panic("implement me")
}

func (ct *cuTypeForTest) GetCoinsHold() sdk.Coins {
	panic("implement me")
}

func (ct *cuTypeForTest) SetCoinsHold(sdk.Coins) error {
	panic("implement me")
}

func (ct *cuTypeForTest) GetSymbol() string {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAssets() []sdk.Asset {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAssetByAddr(denom string, addr string) sdk.Asset {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAsset(denom string, epoch uint64) sdk.Asset {
	panic("implement me")
}

func (ct *cuTypeForTest) AddAsset(denom, address string, epoch uint64) error {
	panic("implement me")
}

func (ct *cuTypeForTest) SetAssetAddress(denom, address string, epoch uint64) error {
	panic("implement me")
}

func (ct *cuTypeForTest) SetAssetNonce(denom string, nonce uint64, epoch uint64) error {
	panic("implement me")
}

func (ct *cuTypeForTest) GetNonce(chain string, addr string) uint64 {
	panic("implement me")
}

func (ct *cuTypeForTest) SetNonce(chain string, nonce uint64, addr string) {
	panic("implement me")
}

func (ct *cuTypeForTest) GetAddress() sdk.CUAddress {
	return nil
}

func (ct *cuTypeForTest) SetAddress(sdk.CUAddress) error {
	return nil
}

func (ct *cuTypeForTest) GetPubKey() crypto.PubKey {
	return nil
}

func (ct *cuTypeForTest) SetPubKey(crypto.PubKey) error {
	return nil
}

func (ct *cuTypeForTest) GetSequence() uint64 {
	return 0
}

func (ct *cuTypeForTest) SetSequence(uint64) error {
	return nil
}

func (ct *cuTypeForTest) GetCoins() sdk.Coins {
	return nil
}

func (ct *cuTypeForTest) SetCoins(sdk.Coins) error {
	return nil
}

func (ct *cuTypeForTest) GetCoinsFrozen() sdk.Coins {
	return nil
}

func (ct *cuTypeForTest) SetCoinsFrozen(sdk.Coins) error {
	return nil
}

func (ct *cuTypeForTest) SpendableCoins(blockTime time.Time) sdk.Coins {
	return nil
}

func (ct *cuTypeForTest) String() string {
	return ""
}

func (ct *cuTypeForTest) GetCUType() sdk.CUType {
	return 0
}

func (ct *cuTypeForTest) SetCUType(sdk.CUType) error {
	return nil
}

func (ct *cuTypeForTest) GetLocked() bool {
	return false
}

func (ct *cuTypeForTest) SetLocked(locked bool) error {

	return nil
}

func (ct *cuTypeForTest) SetMigrationStatus(status sdk.MigrationStatus) {

}

func (ct *cuTypeForTest) GetMigrationStatus() sdk.MigrationStatus {
	return sdk.MigrationFinish
}

func (ct *cuTypeForTest) GetAssetPubkeyEpoch() uint64 {
	return 0
}
