package types

type Asset struct {
	Denom string `json:"denom" yaml:"denom"`

	Address string `json:"address" yaml:"address"`

	Nonce uint64 `json:"nonce" yaml:"nonce"`

	Epoch uint64 `json:"epoch" yaml:"epoch"`

	EnableSendTx bool `json:"enable_sendtx" yaml:"enable_sendtx"`

	GasRemained Int `json:"gas_remained"`
}

var NilAsset = Asset{}

func NewAsset(denom, address string, epoch uint64, enbalesendtx bool) Asset {
	return Asset{Denom: denom, Address: address, EnableSendTx: enbalesendtx, Epoch: epoch, Nonce: 0, GasRemained: ZeroInt()}
}
