package grpcclient

import (
	"github.com/hbtc-chain/chainnode/proto"
	"github.com/stretchr/testify/require"
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/stretchr/testify/assert"
)

func TestStringToInt(t *testing.T) {
	val := stringToInt("")
	assert.Equal(t, sdk.ZeroInt(), val)

	val = stringToInt("12")
	assert.Equal(t, sdk.NewInt(12), val)

	val = stringToInt("12345")
	assert.Equal(t, sdk.NewInt(12345), val)
}

func TestLoadExtAccountTransaction(t *testing.T) {
	signHash := []byte{01, 02}
	reply := &proto.QueryAccountTransactionReply{
		TxHash:   "hash",
		TxStatus: proto.TxStatus_Pending,
		From:     "from",
		To:       "to",
		SignHash: signHash,
	}

	tx, hash, err := loadExtAccountTransaction(reply)
	assert.NoError(t, err)
	assert.Equal(t, signHash, hash)
	assert.Equal(t, sdk.ZeroInt(), tx.GasPrice)
	assert.Equal(t, sdk.ZeroInt(), tx.CostFee)
	assert.Equal(t, sdk.ZeroInt(), tx.GasLimit)
	assert.Equal(t, sdk.ZeroInt(), tx.Amount)

	reply = &proto.QueryAccountTransactionReply{
		TxHash:   "hash",
		TxStatus: proto.TxStatus_Success,
		From:     "from",
		To:       "to",
		SignHash: signHash,
		GasPrice: "1",
		GasLimit: "2",
		CostFee:  "3",
		Amount:   "4",
		Nonce:    10,
	}
	tx, hash, err = loadExtAccountTransaction(reply)
	assert.NoError(t, err)
	assert.Equal(t, signHash, hash)
	assert.Equal(t, sdk.NewInt(1), tx.GasPrice)
	assert.Equal(t, sdk.NewInt(3), tx.CostFee)
	assert.Equal(t, sdk.NewInt(2), tx.GasLimit)
	assert.Equal(t, sdk.NewInt(4), tx.Amount)
	assert.Equal(t, uint64(10), tx.Nonce)
	assert.Equal(t, uint64(3), tx.Status)

}

func TestLoadExtUtxoTransaction(t *testing.T) {
	signHash := []byte{01, 02}
	extVins := []*proto.Vin{{Hash: "859e7d5fe215fb1305fc34cc144f4cf7ddf0f2676780d9e55c629613bd041a17", Index: 1, Amount: 197422083446, Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"}}
	extVouts := []*proto.Vout{
		{Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg", Amount: 197336494895},
		{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: 85475551},
	}

	vins := []*sdk.UtxoIn{{Hash: "859e7d5fe215fb1305fc34cc144f4cf7ddf0f2676780d9e55c629613bd041a17", Index: 1, Amount: sdk.NewInt(197422083446), Address: "n3buST7Uz99E3ERQ1kMCQ848JbTVNQVUeP"}}
	vouts := []*sdk.UtxoOut{
		{Address: "mh1DurxerNqH3nf9p3ivyn7yjgit1ep2Gg", Amount: sdk.NewInt(197336494895)},
		{Address: "mhoGjKn5xegDXL6u5LFSUQdm5ozdM6xao9", Amount: sdk.NewInt(85475551)},
	}

	reply := &proto.QueryUtxoTransactionReply{
		TxHash:     "hash",
		TxStatus:   proto.TxStatus_Pending,
		Vins:       extVins,
		Vouts:      extVouts,
		CostFee:    (sdk.NewInt(197422083446).SubRaw(19733649489).SubRaw(85475551)).String(),
		SignHashes: [][]byte{signHash},
	}

	tx, signHashes, err := loadExtUtxoTransaction(reply)
	assert.Equal(t, uint64(1), tx.Status)
	assert.NoError(t, err)
	assert.Equal(t, vins, tx.Vins)
	assert.Equal(t, vouts, tx.Vouts)
	assert.Equal(t, sdk.NewInt(177602958406), tx.CostFee)
	assert.Equal(t, [][]byte{signHash}, signHashes)
}

func TestLoadExtUtxoTransaction1(t *testing.T) {

	signHash := []byte{01, 02}
	extVins := []*proto.Vin{{Hash: "9f96e84aabb2e31432334220bd314738a1a437fdf29c8091dd9386537d350183", Index: 1, Amount: 500000, Address: "n28anUvZ4RvHsUchWETX7MjbwVYziFy94C"}}
	extVouts := []*proto.Vout{
		{Address: "n2w6xNHy5vAh9gZY5tvpHo2cywrPPR9c3m", Amount: 499000},
	}

	vins := []*sdk.UtxoIn{{Hash: "9f96e84aabb2e31432334220bd314738a1a437fdf29c8091dd9386537d350183", Index: 1, Amount: sdk.NewInt(500000), Address: "n28anUvZ4RvHsUchWETX7MjbwVYziFy94C"}}
	vouts := []*sdk.UtxoOut{
		{Address: "n2w6xNHy5vAh9gZY5tvpHo2cywrPPR9c3m", Amount: sdk.NewInt(499000)},
	}

	reply := &proto.QueryUtxoTransactionReply{
		TxHash:     "hash",
		TxStatus:   proto.TxStatus_Other,
		Vins:       extVins,
		Vouts:      extVouts,
		CostFee:    "1000",
		SignHashes: [][]byte{signHash},
	}

	tx, signHashes, err := loadExtUtxoTransaction(reply)
	assert.Equal(t, uint64(5), tx.Status)
	assert.NoError(t, err)
	assert.Equal(t, vins, tx.Vins)
	assert.Equal(t, vouts, tx.Vouts)
	assert.Equal(t, sdk.NewInt(1000), tx.CostFee)
	assert.Equal(t, [][]byte{signHash}, signHashes)
}

func TestConvertSdkUtxoInsToProtoUtxoIns(t *testing.T) {
	sdkUtxoIns := []*sdk.UtxoIn{
		&sdk.UtxoIn{Hash: "hash0", Index: 0, Address: "address0", Amount: sdk.NewInt(100)},
		&sdk.UtxoIn{Hash: "hash1", Index: 1, Address: "address1", Amount: sdk.NewInt(101)},
		&sdk.UtxoIn{Hash: "hash2", Index: 2, Address: "address2", Amount: sdk.NewInt(102)},
		&sdk.UtxoIn{Hash: "hash3", Index: 3, Address: "address3", Amount: sdk.NewInt(103)},
	}

	protoVins := convertSdkUtxoInsToProtoVins(sdkUtxoIns)
	for i, sdkUtxoIn := range sdkUtxoIns {
		require.Equal(t, sdkUtxoIn.Hash, protoVins[i].Hash)
		require.Equal(t, sdkUtxoIn.Index, uint64(protoVins[i].Index))
		require.Equal(t, sdkUtxoIn.Amount.Int64(), protoVins[i].Amount)
		require.Equal(t, sdkUtxoIn.Address, protoVins[i].Address)
	}
}

func TestConvertProtoUtxoInsToSdkUtxoIns(t *testing.T) {
	protoVins := []*proto.Vin{
		&proto.Vin{Hash: "hash0", Index: 0, Address: "address0", Amount: 100},
		&proto.Vin{Hash: "hash1", Index: 1, Address: "address1", Amount: 101},
		&proto.Vin{Hash: "hash2", Index: 2, Address: "address2", Amount: 102},
		&proto.Vin{Hash: "hash3", Index: 3, Address: "address3", Amount: 103},
	}

	sdkUtxoIns := convertProtoVinsToSdkUtxoIns(protoVins)
	for i, sdkUtxoIn := range sdkUtxoIns {
		require.Equal(t, sdkUtxoIn.Hash, protoVins[i].Hash)
		require.Equal(t, sdkUtxoIn.Index, uint64(protoVins[i].Index))
		require.Equal(t, sdkUtxoIn.Amount.Int64(), protoVins[i].Amount)
		require.Equal(t, sdkUtxoIn.Address, protoVins[i].Address)
	}
}
