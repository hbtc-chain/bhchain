package types

import (
	"fmt"
	"strings"

	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
)

const (
	// ProposalTypeAddToken defines the type for a AddToken
	ProposalTypeAddToken          = "AddToken"
	ProposalTypeTokenParamsChange = "TokenParamsChange"
	ProposalTypeDisableToken      = "DisableToken"
)

// Assert CommunityPoolSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = AddTokenProposal{}
var _ govtypes.Content = TokenParamsChangeProposal{}
var _ govtypes.Content = DisableTokenProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddToken)
	govtypes.RegisterProposalTypeCodec(AddTokenProposal{}, "hbtcchain/AddTokenProposal")
	govtypes.RegisterProposalType(ProposalTypeTokenParamsChange)
	govtypes.RegisterProposalTypeCodec(TokenParamsChangeProposal{}, "hbtcchain/TokenParamsChangeProposal")
	govtypes.RegisterProposalType(ProposalTypeDisableToken)
	govtypes.RegisterProposalTypeCodec(DisableTokenProposal{}, "hbtcchain/DisableTokenProposal")
}

// AddTokenProposal add a new token
type AddTokenProposal struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	TokenInfo   sdk.TokenInfo `json:"token_info" yaml:"token_info"`
}

// NewAddTokenProposal creates a new add token proposal.
func NewAddTokenProposal(title, description string, tokenInfo sdk.TokenInfo) AddTokenProposal {
	return AddTokenProposal{
		Title:       title,
		Description: description,
		TokenInfo:   tokenInfo,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (atp AddTokenProposal) GetTitle() string { return atp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (atp AddTokenProposal) GetDescription() string { return atp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (atp AddTokenProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (atp AddTokenProposal) ProposalType() string { return ProposalTypeAddToken }

func (atp AddTokenProposal) ProposalToken() string { return sdk.NativeToken }

// ValidateBasic runs basic stateless validity checks
func (atp AddTokenProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, atp)
	if err != nil {
		return err
	}
	if !atp.TokenInfo.IsValid() {
		return ErrInvalidProposalTokenInfo(DefaultCodespace)
	}

	// Suppose IsWithdrawalEnabled, IsSendEnabled,IsDepositEnabled are all disabled
	if !(atp.TokenInfo.IsWithdrawalEnabled || atp.TokenInfo.IsSendEnabled || atp.TokenInfo.IsDepositEnabled) {
		return ErrInvalidProposalTokenInfo(DefaultCodespace)
	}
	return nil
}

// String implements the Stringer interface.
func (atp AddTokenProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Token Proposal:
  Title:       %s
  Description: %s
`, atp.Title, atp.Description))
	b.WriteString(atp.TokenInfo.String())
	return b.String()
}

type ParamChange struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewParamChange(key, value string) ParamChange {
	return ParamChange{key, value}
}

// ModifyTokenProposal modify a token's variable parameter
type TokenParamsChangeProposal struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	Symbol      string        `json:"symbol" yaml:"symbol"`
	Changes     []ParamChange `json:"changes" yaml:"changes"`
}

// NewAddTokenProposal creates a new add token proposal.
func NewTokenParamsChangeProposal(title, description, sybmol string, changes []ParamChange) TokenParamsChangeProposal {
	return TokenParamsChangeProposal{
		Title:       title,
		Description: description,
		Symbol:      sybmol,
		Changes:     changes,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (ctpp TokenParamsChangeProposal) GetTitle() string { return ctpp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (ctpp TokenParamsChangeProposal) GetDescription() string { return ctpp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (ctpp TokenParamsChangeProposal) ProposalRoute() string { return RouterKey }

func (ctpp TokenParamsChangeProposal) ProposalToken() string { return sdk.NativeToken }

// ProposalType returns the type of a community pool spend proposal.
func (ctpp TokenParamsChangeProposal) ProposalType() string { return ProposalTypeTokenParamsChange }

// ValidateBasic runs basic stateless validity checks
func (ctpp TokenParamsChangeProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, ctpp)
	if err != nil {
		return err
	}

	if !sdk.Symbol(ctpp.Symbol).IsValidTokenName() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("%s is invalid", ctpp.Symbol))
	}

	//Remove duplicated keys if any
	keysMap := map[string]interface{}{}

	for _, pc := range ctpp.Changes {
		_, ok := keysMap[pc.Key]
		if !ok {
			keysMap[pc.Key] = nil
		} else {
			return ErrDuplicatedKey(DefaultCodespace)
		}

		if len(pc.Key) == 0 {
			return ErrEmptyKey(DefaultCodespace)
		}
		if len(pc.Value) == 0 {
			return ErrEmptyValue(DefaultCodespace)
		}
	}

	return err
}

// String implements the Stringer interface.
func (ctpp TokenParamsChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Change Token Param Proposal:
 Title:       %s
 Description: %s
 Symbol:      %s
 Changes:
`, ctpp.Title, ctpp.Description, ctpp.Symbol))

	for _, pc := range ctpp.Changes {
		b.WriteString(fmt.Sprintf("%s: %s\t", pc.Key, pc.Value))
	}
	return b.String()
}

// ModifyTokenProposal modify a token's variable parameter
type DisableTokenProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Symbol      string `json:"symbol" yaml:"symbol"`
}

// NewAddTokenProposal creates a new add token proposal.
func NewDisableTokenProposal(title, description, sybmol string) DisableTokenProposal {
	return DisableTokenProposal{
		Title:       title,
		Description: description,
		Symbol:      sybmol,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (dtp DisableTokenProposal) GetTitle() string { return dtp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (dtp DisableTokenProposal) GetDescription() string { return dtp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (dtp DisableTokenProposal) ProposalRoute() string { return RouterKey }

func (dtp DisableTokenProposal) ProposalToken() string { return sdk.NativeToken }

// ProposalType returns the type of a community pool spend proposal.
func (dtp DisableTokenProposal) ProposalType() string { return ProposalTypeDisableToken }

// ValidateBasic runs basic stateless validity checks
func (dtp DisableTokenProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, dtp)
	if err != nil {
		return err
	}

	if !sdk.Symbol(dtp.Symbol).IsValidTokenName() {
		return sdk.ErrInvalidSymbol(fmt.Sprintf("%s is invalid", dtp.Symbol))
	}

	return err
}

// String implements the Stringer interface.
func (dtp DisableTokenProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Disable Token Proposal:
 Title:       %s
 Description: %s
 Symbol:      %s
`, dtp.Title, dtp.Description, dtp.Symbol))
	return b.String()
}
