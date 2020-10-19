package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/hbtc-chain/bhchain/types"
)

type Plan struct {
	Name   string    `json:"name"`
	Time   time.Time `json:"time"`
	Height int64     `json:"height"`
	Info   string    `json:"info"`
}

func (p Plan) String() string {
	due := p.DueAt()
	dueUp := strings.ToUpper(due[0:1]) + due[1:]
	return fmt.Sprintf(`Upgrade Plan
  Name: %s
  %s
  Info: %s`, p.Name, dueUp, p.Info)
}

// ValidateBasic does basic validation of a Plan
func (p Plan) ValidateBasic() sdk.Error {
	if len(p.Name) == 0 {
		return sdk.ErrInvalidTx("name cannot be empty")
	}
	if p.Height < 0 {
		return sdk.ErrInvalidTx("height cannot be negative")
	}
	if p.Time.IsZero() && p.Height == 0 {
		return sdk.ErrInvalidTx("must set either time or height")
	}
	if !p.Time.IsZero() && p.Height != 0 {
		return sdk.ErrInvalidTx("cannot set both time and height")
	}

	return nil
}

// ShouldExecute returns true if the Plan is ready to execute given the current context
func (p Plan) ShouldExecute(ctx sdk.Context) bool {
	if !p.Time.IsZero() {
		return !ctx.BlockTime().Before(p.Time)
	}
	if p.Height > 0 {
		return p.Height <= ctx.BlockHeight()
	}
	return false
}

// DueAt is a string representation of when this plan is due to be executed
func (p Plan) DueAt() string {
	if !p.Time.IsZero() {
		return fmt.Sprintf("time: %s", p.Time.UTC().Format(time.RFC3339))
	}
	return fmt.Sprintf("height: %d", p.Height)
}
