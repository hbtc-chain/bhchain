package tests

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/hbtc-chain/bhchain/chainnode"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/transfer/types"
)

func TestBehaviourOPCUAssetTransferNormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)
	input.opcu.SetAssetAddress("eth", "eth", 0)
	input.opcu.SetEnableSendTx(false, "eth", ethAddr)
	input.ck.SetCU(ctx, input.opcu)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareOPCUSystransferOrder(input, ctx)
		costFee := sdk.NewInt(0)
		for _, validator := range validators {
			result := keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validator.OperatorAddress), orderID, costFee)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}
}

func TestBehaviourOPCUAssetTransferAbnormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)
	input.opcu.SetAssetAddress("eth", "eth", 0)
	input.opcu.SetEnableSendTx(false, "eth", ethAddr)
	input.ck.SetCU(ctx, input.opcu)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareOPCUSystransferOrder(input, ctx)
		costFee := sdk.NewInt(0)
		for j, validator := range validators {
			if j == 0 {
				continue
			}
			result := keeper.OpcuAssetTransferFinish(ctx, sdk.CUAddress(validator.OperatorAddress), orderID, costFee)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestBehaviourSysTransferNormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	input.opcu.SetEnableSendTx(false, "eth", ethAddr)
	input.ck.SetCU(ctx, input.opcu)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareSystransferOrder(input, ctx)
		costFee := sdk.NewInt(0)
		for _, validator := range validators {
			result := keeper.SysTransferFinish(ctx, sdk.CUAddress(validator.OperatorAddress), orderID, costFee)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}
}

func TestBehaviourSysTransferAbnormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	input.opcu.SetEnableSendTx(false, "eth", ethAddr)
	input.ck.SetCU(ctx, input.opcu)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareSystransferOrder(input, ctx)
		costFee := sdk.NewInt(0)
		for j, validator := range validators {
			if j == 0 {
				continue
			}
			result := keeper.SysTransferFinish(ctx, sdk.CUAddress(validator.OperatorAddress), orderID, costFee)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestBehaviourWithdrawNormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	err := input.opcu.SetAssetAddress("eth", ethAddr, 1)
	require.NoError(t, err)
	input.opcu.SetEnableSendTx(false, "eth", ethAddr)
	input.ck.SetCU(ctx, input.opcu)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareWithdrawOrder(t, input, ctx)
		costFee := sdk.NewInt(0)
		for _, validator := range validators {
			result := keeper.WithdrawalFinish(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, costFee, true)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}
}

func TestBehaviourWithdrawAbnormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	err := input.opcu.SetAssetAddress("eth", ethAddr, 1)
	require.NoError(t, err)
	input.ck.SetCU(ctx, input.opcu)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareWithdrawOrder(t, input, ctx)
		costFee := sdk.NewInt(0)
		for j, validator := range validators {
			if j == 0 {
				continue
			}
			result := keeper.WithdrawalFinish(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, costFee, true)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestBehaviourCollectNormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareCollectOrder(t, input, ctx, sdk.DepositConfirmed)
		costFee := sdk.NewInt(160000)
		for _, validator := range validators {
			result := keeper.CollectFinish(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, costFee)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}
}

func TestBehaviourCollectAbnormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareCollectOrder(t, input, ctx, sdk.DepositConfirmed)
		costFee := sdk.NewInt(160000)
		for j, validator := range validators {
			if j == 0 {
				continue
			}
			result := keeper.CollectFinish(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, costFee)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestBehaviourConfirmDepositNormal(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareDepositOrder(t, input, ctx, sdk.DepositUnconfirm)
		for _, validator := range validators {
			result := keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, []string{})
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}
}

func TestBehaviourConfirmDepositMissing(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareDepositOrder(t, input, ctx, sdk.DepositUnconfirm)
		for j, validator := range validators {
			if j == 0 {
				continue
			}
			result := keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, []string{})
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestBehaviourConfirmDepositInvalid(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	for i := 1; i <= 11; i++ {
		orderID, _ := prepareDepositOrder(t, input, ctx, sdk.DepositUnconfirm)
		for j, validator := range validators {
			var validOrderIds []string
			var invalidOrderIDs []string
			if j == 0 {
				validOrderIds = []string{orderID}
				invalidOrderIDs = []string{}
			} else {
				validOrderIds = []string{}
				invalidOrderIDs = []string{orderID}
			}
			result := keeper.ConfirmedDeposit(ctx, sdk.CUAddress(validator.OperatorAddress), validOrderIds, invalidOrderIDs)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}

	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestOrderRetryPunishment(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	evidences := []types.EvidenceValidator{
		{
			EvidenceType: 1,
			Validator:    sdk.CUAddress(validators[0].OperatorAddress).String(),
		},
	}
	for i := 1; i <= 11; i++ {
		orderID := prepareKeyGenOrder(input, ctx)
		for _, validator := range validators {
			result := keeper.OrderRetry(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, 1, evidences)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}
	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for i, val := range vals {
		if i == 0 {
			require.True(t, val.Jailed)
		} else {
			require.False(t, val.Jailed)
		}
	}
}

func TestOrderRetryNoPunishment(t *testing.T) {
	input := setupTestInput(t)
	validators := input.validators
	keeper := input.k
	ctx := input.ctx.WithBlockHeight(100)

	evidences := []types.EvidenceValidator{
		{
			EvidenceType: 1,
			Validator:    sdk.CUAddress(validators[0].OperatorAddress).String(),
		},
	}
	for i := 1; i <= 10; i++ {
		orderID := prepareKeyGenOrder(input, ctx)
		for _, validator := range validators {
			result := keeper.OrderRetry(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, 1, evidences)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}
	ctx = ctx.WithBlockHeight(200)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals := input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}

	for i := 1; i <= 90; i++ {
		orderID := prepareKeyGenOrder(input, ctx)
		for _, validator := range validators {
			result := keeper.OrderRetry(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, 1, nil)
			require.Equal(t, sdk.CodeOK, result.Code, result)
		}
	}
	orderID := prepareKeyGenOrder(input, ctx)
	for _, validator := range validators {
		result := keeper.OrderRetry(ctx, sdk.CUAddress(validator.OperatorAddress), []string{orderID}, 1, evidences)
		require.Equal(t, sdk.CodeOK, result.Code, result)
	}
	ctx = ctx.WithBlockHeight(300)
	input.evidenceKeeper.RecordMisbehaviourVoter(ctx)
	vals = input.stakingkeeper.GetAllValidators(ctx)
	for _, val := range vals {
		require.False(t, val.Jailed)
	}
}

func prepareCollectOrder(t *testing.T, input testInput, ctx sdk.Context, depositStatus uint16) (string, string) {
	orderID := uuid.NewV4().String()
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	symbol := "eth"
	chain := "eth"
	depositTxHash := "depositTxHash"
	depositTxIndex := uint64(0)
	collectTxHash := orderID
	amount := sdk.ZeroInt()

	asset := input.ik.GetOrNewCUIBCAsset(input.ctx, sdk.CUTypeUser, user1CUAddr)
	input.ik.SetCUIBCAsset(input.ctx, asset)
	require.Nil(t, err)

	depositItem, err := sdk.NewDepositItem(depositTxHash, depositTxIndex, amount, user1EthAddr, "", sdk.DepositItemStatusInProcess)
	require.NoError(t, err)

	input.ik.SaveDeposit(ctx, symbol, user1CUAddr, depositItem)

	signedData := []byte("Collect")
	input.ok.SetOrder(ctx, &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        orderID,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
			Height:    uint64(ctx.BlockHeight()),
			Status:    sdk.OrderStatusSignFinish,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1EthAddr,
		CollectToCU:        input.opcu.GetAddress(),
		Amount:             amount,
		ExtTxHash:          collectTxHash,
		Txhash:             depositTxHash,
		Index:              depositTxIndex,
		Memo:               "",
		DepositStatus:      depositStatus,
		RawData:            []byte("raw"),
		SignedTx:           signedData,
	})

	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:   collectTxHash,
		To:     input.opcu.GetAddress().String(),
		Amount: amount,
		Nonce:  0,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()
	return orderID, collectTxHash
}

func prepareDepositOrder(t *testing.T, input testInput, ctx sdk.Context, depositStatus uint16) (string, string) {
	orderID := uuid.NewV4().String()
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	symbol := "eth"
	chain := "eth"
	depositTxHash := "depositTxHash"
	depositTxIndex := uint64(0)
	collectTxHash := "collectTxHash"
	amount := sdk.NewInt(80000000000)

	asset := input.ik.GetOrNewCUIBCAsset(input.ctx, sdk.CUTypeUser, user1CUAddr)
	input.ik.SetCUIBCAsset(input.ctx, asset)

	require.Nil(t, err)

	signedData := []byte("Deposit")
	input.ok.SetOrder(ctx, &sdk.OrderCollect{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        orderID,
			OrderType: sdk.OrderTypeCollect,
			Symbol:    symbol,
			Height:    uint64(ctx.BlockHeight()),
			Status:    sdk.OrderStatusSignFinish,
		},
		CollectFromCU:      user1CUAddr,
		CollectFromAddress: user1EthAddr,
		CollectToCU:        input.opcu.GetAddress(),
		Amount:             amount,
		ExtTxHash:          collectTxHash,
		Txhash:             depositTxHash,
		Index:              depositTxIndex,
		Memo:               "",
		DepositStatus:      depositStatus,
		RawData:            []byte("raw"),
		SignedTx:           signedData,
	})

	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:   collectTxHash,
		To:     input.opcu.GetAddress().String(),
		Amount: amount,
		Nonce:  0,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()
	return orderID, collectTxHash
}

func prepareWithdrawOrder(t *testing.T, input testInput, ctx sdk.Context) (string, string) {
	orderID := uuid.NewV4().String()
	user1CUAddr, err := sdk.CUAddressFromBase58("HBCLmQcskpdQivEkRrh1gNPm7c9aVB8hh1fy")
	user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	symbol := "eth"
	chain := "eth"
	txHash := orderID
	amount := sdk.ZeroInt()

	require.Nil(t, err)

	signedData := []byte("Withdraw")
	input.ok.SetOrder(ctx, &sdk.OrderWithdrawal{
		OrderBase: sdk.OrderBase{
			CUAddress: user1CUAddr,
			ID:        orderID,
			OrderType: sdk.OrderTypeWithdrawal,
			Symbol:    symbol,
			Height:    uint64(ctx.BlockHeight()),
			Status:    sdk.OrderStatusSignFinish,
		},
		Txhash:            txHash,
		RawData:           []byte{1},
		SignedTx:          signedData,
		WithdrawToAddress: user1EthAddr,
		OpCUaddress:       input.opcu.GetAddress().String(),
		WithdrawStatus:    sdk.WithdrawStatusValid,
	})

	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:   txHash,
		To:     input.opcu.GetAddress().String(),
		Amount: amount,
		Nonce:  0,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()
	mockCN.On("ValidAddress", chain, symbol, "").Return(true, "").Once()

	return orderID, txHash
}

func prepareSystransferOrder(input testInput, ctx sdk.Context) (string, string) {
	orderID := uuid.NewV4().String()
	//user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	symbol := "eth"
	chain := "eth"
	txHash := orderID
	amount := sdk.ZeroInt()

	signedData := []byte("Systransfer")
	input.ok.SetOrder(ctx, &sdk.OrderSysTransfer{
		OrderBase: sdk.OrderBase{
			CUAddress: input.opcu.GetAddress(),
			ID:        orderID,
			OrderType: sdk.OrderTypeSysTransfer,
			Symbol:    symbol,
			Height:    uint64(ctx.BlockHeight()),
			Status:    sdk.OrderStatusSignFinish,
		},
		TxHash:      txHash,
		RawData:     []byte{1},
		SignedTx:    signedData,
		OpCUaddress: input.opcu.GetAddress().String(),
		ToCU:        input.opcu.GetAddress().String(),
	})

	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:   txHash,
		To:     input.opcu.GetAddress().String(),
		Amount: amount,
		Nonce:  0,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()
	mockCN.On("ValidAddress", chain, symbol, "").Return(true, "").Once()

	return orderID, txHash
}

func prepareOPCUSystransferOrder(input testInput, ctx sdk.Context) (string, string) {
	orderID := uuid.NewV4().String()
	//user1EthAddr := "0xc96d141c9110a8E61eD62caaD8A7c858dB15B82c"
	symbol := "eth"
	chain := "eth"
	txHash := orderID
	amount := sdk.ZeroInt()

	signedData := []byte("OPCUSystransfer")
	input.ok.SetOrder(ctx, &sdk.OrderOpcuAssetTransfer{
		OrderBase: sdk.OrderBase{
			CUAddress: input.opcu.GetAddress(),
			ID:        orderID,
			OrderType: sdk.OrderTypeSysTransfer,
			Symbol:    symbol,
			Height:    uint64(ctx.BlockHeight()),
			Status:    sdk.OrderStatusSignFinish,
		},
		Txhash:   txHash,
		RawData:  []byte{1},
		SignedTx: signedData,
	})

	chainnodeCollectTx := chainnode.ExtAccountTransaction{
		Hash:   txHash,
		To:     input.opcu.GetAddress().String(),
		Amount: amount,
		Nonce:  0,
	}
	mockCN.On("QueryAccountTransactionFromSignedData", chain, symbol, signedData).Return(&chainnodeCollectTx, nil).Once()
	mockCN.On("ValidAddress", chain, symbol, "").Return(true, "").Once()

	return orderID, txHash
}

func prepareKeyGenOrder(input testInput, ctx sdk.Context) string {
	orderID := uuid.NewV4().String()
	symbol := "eth"

	input.ok.SetOrder(ctx, &sdk.OrderKeyGen{
		OrderBase: sdk.OrderBase{
			CUAddress: input.opcu.GetAddress(),
			ID:        orderID,
			OrderType: sdk.OrderTypeKeyGen,
			Symbol:    symbol,
			Height:    uint64(ctx.BlockHeight()),
			Status:    sdk.OrderStatusBegin,
		},
	})

	return orderID
}
