package types

import (
	"errors"
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type Dex struct {
	ID             uint32        `json:"id"`
	Name           string        `json:"name"`
	Owner          sdk.CUAddress `json:"owner"`
	IncomeReceiver sdk.CUAddress `json:"income_receiver"`
}

func (d *Dex) Validate() error {
	if d.Name == "" || len(d.Name) > maxDexNameLength {
		return errors.New("invalid dex name")
	}

	if !d.Owner.IsValidAddr() {
		return fmt.Errorf("Owner address: %s is invalid", d.Owner.String())
	}
	if !d.IncomeReceiver.IsValidAddr() {
		return fmt.Errorf("Income receiver address: %s is invalid", d.IncomeReceiver.String())
	}
	return nil

}
