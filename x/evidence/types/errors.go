// nolint
package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
)

// Error codes specific to the evidence module
const (
	defaultCodespace sdk.CodespaceType = ModuleName

	CodeNoEvidenceHandlerExists sdk.CodeType = 1
	CodeInvalidEvidence         sdk.CodeType = 2
	CodeNoEvidenceExists        sdk.CodeType = 3
	CodeEvidenceExists          sdk.CodeType = 4
)

// ErrNoEvidenceHandlerExists returns a typed ABCI error for an invalid evidence
// handler route.
func ErrNoEvidenceHandlerExists(route string) sdk.Error {
	return sdk.NewError(
		defaultCodespace,
		CodeNoEvidenceHandlerExists,
		fmt.Sprintf("route '%s' does not have a registered evidence handler", route),
	)
}

// ErrInvalidEvidence returns a typed ABCI error for invalid evidence.
func ErrInvalidEvidence(msg string) sdk.Error {
	return sdk.NewError(
		defaultCodespace,
		CodeInvalidEvidence,
		fmt.Sprintf("invalid evidence: %s", msg),
	)
}

// ErrNoEvidenceExists returns a typed ABCI error for Evidence that does not exist
// for a given hash.
func ErrNoEvidenceExists(hash string) sdk.Error {
	return sdk.NewError(
		defaultCodespace,
		CodeNoEvidenceExists,
		fmt.Sprintf("evidence with hash %s does not exist", hash),
	)
}

// ErrEvidenceExists returns a typed ABCI error for Evidence that already exists
// by hash in state.
func ErrEvidenceExists(hash string) sdk.Error {
	return sdk.NewError(
		defaultCodespace,
		CodeEvidenceExists,
		fmt.Sprintf("evidence with hash %s already exists", hash),
	)
}
