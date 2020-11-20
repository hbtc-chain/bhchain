package types

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/params"
)

var DefaultIssueTokenFee = sdk.NewIntWithDecimal(1, 18) //1hbc as open fee

// Parameter keys
var (
	KeyIssueTokenFee = []byte("IssueTokenFee")
)

var _ params.ParamSet = &Params{}

// Params defines the parameters for the auth module.
type Params struct {
	IssueTokenFee sdk.Int `json:"issue_token_fee"`
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
		{KeyIssueTokenFee, &p.IssueTokenFee},
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
		IssueTokenFee: DefaultIssueTokenFee,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("OpenTokenFee: %v\n", p.IssueTokenFee))
	return sb.String()
}

func (p Params) Validate() error {
	if p.IssueTokenFee.IsPositive() {
		return nil
	}

	return fmt.Errorf("IssueTokenFee %v is not positive", p.IssueTokenFee)
}
