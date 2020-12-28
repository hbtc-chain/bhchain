package types

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/hbtc-chain/bhchain/base58"
	sdk "github.com/hbtc-chain/bhchain/types"
	govtypes "github.com/hbtc-chain/bhchain/x/gov/types"
	"golang.org/x/crypto/ripemd160"
)

const (
	// ProposalTypeAddToken defines the type for a AddToken
	ProposalTypeAddToken          = "AddToken"
	ProposalTypeTokenParamsChange = "TokenParamsChange"
)

// Assert CommunityPoolSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = AddTokenProposal{}
var _ govtypes.Content = TokenParamsChangeProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddToken)
	govtypes.RegisterProposalTypeCodec(AddTokenProposal{}, "hbtcchain/AddTokenProposal")
	govtypes.RegisterProposalType(ProposalTypeTokenParamsChange)
	govtypes.RegisterProposalTypeCodec(TokenParamsChangeProposal{}, "hbtcchain/TokenParamsChangeProposal")
}

// AddTokenProposal add a new token
type AddTokenProposal struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	TokenInfo   *sdk.IBCToken `json:"token_info" yaml:"token_info"`
}

// NewAddTokenProposal creates a new add token proposal.
func NewAddTokenProposal(title, description string, tokenInfo *sdk.IBCToken) AddTokenProposal {
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

	if !sdk.Symbol(ctpp.Symbol).IsValid() {
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

func CalSymbol(issuer string, chain sdk.Symbol) sdk.Symbol {
	payload := []byte(fmt.Sprintf("%s-%s", chain, issuer))
	hasherSHA256 := sha256.New()
	hasherSHA256.Write(payload)
	sha := hasherSHA256.Sum(nil)

	hasherRIPEMD160 := ripemd160.New()
	hasherRIPEMD160.Write(sha)
	bz := hasherRIPEMD160.Sum(nil)

	sum := base58.Checksum(bz)
	bz = append(bz, sum[:]...)

	symbol := strings.ToUpper(chain.String()) + base58.Encode(bz)
	return sdk.Symbol(symbol)
}
