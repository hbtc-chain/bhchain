package reactor

import (
	"github.com/hbtc-chain/bhchain/services/p2p/pb"

	amino "github.com/tendermint/go-amino"
	"golang.org/x/crypto/sha3"
)

var cdc = amino.NewCodec()

func init() {
	registerBHMessages(cdc)
}

func registerBHMessages(cdc *amino.Codec) {
	cdc.RegisterConcrete(&pb.BHMsg{}, "hbtcchain/BHMsg", nil)
}

type Hash [32]byte

func toHash(bytes []byte) Hash {
	return sha3.Sum256(bytes)
}
