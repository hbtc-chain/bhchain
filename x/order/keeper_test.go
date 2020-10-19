package order

import (
	"github.com/hbtc-chain/bhchain/codec"
	"github.com/hbtc-chain/bhchain/store"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/order/types"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

func TestKeeper_NewOrders(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	ok := input.ook
	fromAddr, toAddr := sdk.NewCUAddress(), sdk.NewCUAddress()
	keynodes := []sdk.CUAddress{sdk.NewCUAddress(), sdk.NewCUAddress()}

	// NewOrderKeyGen
	orderID := uuid.NewV4().String()
	order := ok.NewOrderKeyGen(input.ctx, fromAddr, orderID, "eth", keynodes, 3, toAddr, sdk.NewCoins(sdk.Coin{sdk.NativeToken, sdk.NewInt(10000)}))
	ok.SetOrder(input.ctx, order)
	orderGot := ok.GetOrder(ctx, orderID)
	assert.NotNil(t, order)
	assert.Equal(t, sdk.OrderStatusBegin, orderGot.GetOrderStatus())
	assert.Equal(t, sdk.OrderTypeKeyGen, orderGot.GetOrderType())

	// NewOrderCollect
	orderID = uuid.NewV4().String()
	order2 := ok.NewOrderCollect(input.ctx, fromAddr, orderID, "eth",
		sdk.NewCUAddress(), "0x12b3c42a12fe9", sdk.NewInt(3), sdk.NewInt(3), sdk.NewInt(3), "txhash", 0, "")
	ok.SetOrder(input.ctx, order2)
	orderGot = ok.GetOrder(ctx, orderID)
	assert.NotNil(t, order)
	assert.Equal(t, sdk.OrderStatusBegin, orderGot.GetOrderStatus())
	assert.Equal(t, sdk.OrderTypeCollect, orderGot.GetOrderType())

	// NewOrderWithdrawal
	orderID = uuid.NewV4().String()
	order3 := ok.NewOrderWithdrawal(input.ctx, fromAddr, orderID, "eth",
		sdk.NewInt(3), sdk.NewInt(3), sdk.NewInt(3), "0x1322cabc2334098", "BHfty2766vghvvx", "txhash")
	ok.SetOrder(input.ctx, order3)
	orderGot = ok.GetOrder(ctx, orderID)
	assert.NotNil(t, order)
	assert.Equal(t, sdk.OrderStatusBegin, orderGot.GetOrderStatus())
	assert.Equal(t, sdk.OrderTypeWithdrawal, orderGot.GetOrderType())

	// NewOrderSysTransfer
	orderID = uuid.NewV4().String()
	order4 := ok.NewOrderSysTransfer(input.ctx, fromAddr, orderID, "eth",
		sdk.NewInt(3), sdk.NewInt(3), "BHy5766fghdhxw34", "0x1322cabc2334098", "BHfty2766vghvvx", "BHfty2766vghvvx")
	ok.SetOrder(input.ctx, order4)
	orderGot = ok.GetOrder(ctx, orderID)
	assert.NotNil(t, order)
	assert.Equal(t, sdk.OrderStatusBegin, orderGot.GetOrderStatus())
	assert.Equal(t, sdk.OrderTypeSysTransfer, orderGot.GetOrderType())
}

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ook Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()
	cdc := codec.New()
	RegisterCodec(cdc)
	storeKey := sdk.NewKVStoreKey(StoreKey)
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, sdk.CodespaceType("order"))
	ook := NewKeeper(cdc, storeKey, pk.Subspace(types.DefaultParamspace))

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())

	return testInput{cdc: cdc, ctx: ctx, ook: ook}
}
