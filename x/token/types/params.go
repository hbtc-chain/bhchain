/*
 * *******************************************************************
 * @项目名称: token
 * @文件名称: params.go
 * @Date: 2019/08/07
 * @Author: Keep
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

package types

import (
	"bytes"
	"fmt"
	"github.com/hbtc-chain/bhchain/x/params"
	"strings"
)

// Default parameter values
const (
	DefaultTokenCacheSize uint64 = 32 //cache size for token
)

var DefaultReservedSymbols = []string{"eos", "usdt", "bch", "bsv", "ltc", "bnb", "xrp", "okb", "ht", "dash", "etc", "neo", "atom", "zec", "ont", "doge", "tusd", "bat", "qtum", "vsys", "iost", "dcr", "zrx", "beam", "grin"}

// Parameter keys
var (
	KeyTokenCacheSize  = []byte("TokenCacheSize")
	KeyReservedSymbols = []byte("ReservedSymbols")
)

var _ params.ParamSet = &Params{}

// Params defines the parameters for the auth module.
type Params struct {
	TokenCacheSize  uint64   `json:"token_cache_size"`
	ReservedSymbols []string `json:"reserved_symbols"`
}

// ParamKeyTable for auth module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyTokenCacheSize, &p.TokenCacheSize},
		{KeyReservedSymbols, &p.ReservedSymbols},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		TokenCacheSize:  DefaultTokenCacheSize,
		ReservedSymbols: DefaultReservedSymbols,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params:")
	sb.WriteString(fmt.Sprintf("TokenCacheSize:%v\t", p.TokenCacheSize))
	sb.WriteString(fmt.Sprintf("ReservedSymbols:%s\t", strings.Join(p.ReservedSymbols, ",")))

	return sb.String()
}
