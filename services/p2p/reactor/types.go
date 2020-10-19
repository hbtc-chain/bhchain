/*
 * *******************************************************************
 * @项目名称: p2p
 * @文件名称: reactor.go
 * @Date: 2019/04/19
 * @Author: kai.wen
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

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
