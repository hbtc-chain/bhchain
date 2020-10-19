package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	cmn "github.com/tendermint/tendermint/libs/common"

	abci "github.com/tendermint/tendermint/abci/types"
)

// CodeType - ABCI code identifier within codespace
type CodeType uint32

// CodespaceType - codespace identifier
type CodespaceType string

// IsOK - is everything okay?
func (code CodeType) IsOK() bool {
	return code == CodeOK
}

// SDK error codes
const (
	// Base error codes
	CodeOK                CodeType = 0
	CodeInternal          CodeType = 1
	CodeTxDecode          CodeType = 2
	CodeInvalidSequence   CodeType = 3
	CodeUnauthorized      CodeType = 4
	CodeInsufficientFunds CodeType = 5
	CodeUnknownRequest    CodeType = 6
	CodeInvalidAddress    CodeType = 7
	CodeInvalidPubKey     CodeType = 8
	CodeUnknownAddress    CodeType = 9
	CodeInsufficientCoins CodeType = 10
	CodeInvalidCoins      CodeType = 11
	CodeOutOfGas          CodeType = 12
	CodeMemoTooLarge      CodeType = 13
	CodeInsufficientFee   CodeType = 14
	CodeTooManySignatures CodeType = 15
	CodeGasOverflow       CodeType = 16
	CodeNoSignatures      CodeType = 17
	CodeTooMuchPrecision  CodeType = 18
	//high level error codes

	CodeDuplicatedUtxo     CodeType = 1000
	CodeUnknownUtxo        CodeType = 1001
	CodeMismatchUtxoAmount CodeType = 1002
	CodeInvalidCollect     CodeType = 1003
	CodeAmountError        CodeType = 1004
	CodeAssetError         CodeType = 1005
	CodeUtxoError          CodeType = 1006
	CodeOnHoldError        CodeType = 1007
	CodeBlkNumberError     CodeType = 1008

	CodeInvalidAccount          CodeType = 1010
	CodeSymbolAlreadyExist      CodeType = 1011
	CodeInvalidSymbol           CodeType = 1012
	CodeUnsupportToken          CodeType = 1013
	CodeInvalidOrder            CodeType = 1014
	CodeInvalidAsset            CodeType = 1015
	CodeInvalidTx               CodeType = 1016
	CodeTransactionIsNotEnabled CodeType = 1017

	CodeUnsupportAddressType  CodeType = 1020
	CodeNotFoundOrder         CodeType = 1021
	CodeNotFoundAsset         CodeType = 1022
	CodeNotFoundCustodianUnit CodeType = 1023
	CodeInvalidCUType         CodeType = 1024

	CodeInsufficientValidtorNumberForKeyGen CodeType = 1030
	CodeInsufficientValidtorNumber          CodeType = 1031
	CodeMigrationInProgress                 CodeType = 1032
	CodeSystemBusy                          CodeType = 1033
	CodePreKeyGenTooMany                    CodeType = 1034
	CodeWaitAssignTooMany                   CodeType = 1035

	CodeEmptyDBGet CodeType = 1040
	// CodespaceRoot is a codespace for error codes in this file only.
	// Notice that 0 is an "unset" codespace, which can be overridden with
	// Error.WithDefaultCodespace().
	CodespaceUndefined CodespaceType = ""
	CodespaceRoot      CodespaceType = "hbtcchain_base"
)

func unknownCodeMsg(code CodeType) string {
	return fmt.Sprintf("unknown code %d", code)
}

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInternal:
		return "internal error"
	case CodeTxDecode:
		return "tx parse error"
	case CodeInvalidSequence:
		return "invalid sequence"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeInsufficientFunds:
		return "insufficient funds"
	case CodeUnknownRequest:
		return "unknown request"
	case CodeInvalidAddress:
		return "invalid address"
	case CodeInvalidPubKey:
		return "invalid pubkey"
	case CodeUnknownAddress:
		return "unknown address"
	case CodeInsufficientCoins:
		return "insufficient coins"
	case CodeInvalidCoins:
		return "invalid coins"
	case CodeOutOfGas:
		return "out of gas"
	case CodeMemoTooLarge:
		return "memo too large"
	case CodeInsufficientFee:
		return "insufficient fee"
	case CodeTooManySignatures:
		return "maximum numer of signatures exceeded"
	case CodeNoSignatures:
		return "no signatures supplied"
	case CodeTooMuchPrecision:
		return "too much precision"

	case CodeDuplicatedUtxo:
		return "Duplicated Utxo"
	case CodeUnknownUtxo:
		return "Unknown Utxo"
	case CodeMismatchUtxoAmount:
		return "Mismatch Utxo Amount"
	case CodeInvalidCollect:
		return "Collect w/o ToAddr Address"
	case CodeAmountError:
		return "Amount set is invalid"
	case CodeAssetError:
		return "Asset is invalid"
	case CodeUtxoError:
		return "Utxo is invalid"
	case CodeOnHoldError:
		return "Onhold is error"
	case CodeBlkNumberError:
		return "Block number is error"
	case CodeInvalidAccount:
		return "Account is invalid"
	case CodeSymbolAlreadyExist:
		return "symbol arleady exist"
	case CodeInvalidSymbol:
		return "Symbol is invalid"
	case CodeUnsupportToken:
		return "Unsupport Token"
	case CodeUnsupportAddressType:
		return "Unsupport address type"
	case CodeNotFoundOrder:
		return "Not Found order"
	case CodeNotFoundAsset:
		return "Not Found asset"
	case CodeNotFoundCustodianUnit:
		return "Not Found CustodianUnit"
	case CodeInvalidOrder:
		return "Order isn invalid"
	case CodeInvalidAsset:
		return "Asset is invalid"
	case CodeInsufficientValidtorNumberForKeyGen:
		return "Insufficient validator number for key gen"
	case CodeInvalidTx:
		return "Invalid tx relative"
	case CodeTransactionIsNotEnabled:
		return "Transaction not enabled temporary"
	case CodeEmptyDBGet:
		return "Data get from db is error or nil"
	case CodeInvalidCUType:
		return "Invalid CU type for op"
	case CodeMigrationInProgress:
		return "validator migration is in progress"
	case CodeSystemBusy:
		return "System is busy"
	case CodePreKeyGenTooMany:
		return "Too many prekeygen orders"
	case CodeWaitAssignTooMany:
		return "Too many wait assign orders"
	default:
		return unknownCodeMsg(code)
	}
}

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

// nolint
func ErrInternal(msg string) Error {
	return newErrorWithRootCodespace(CodeInternal, msg)
}
func ErrTxDecode(msg string) Error {
	return newErrorWithRootCodespace(CodeTxDecode, msg)
}
func ErrInvalidSequence(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidSequence, msg)
}
func ErrUnauthorized(msg string) Error {
	return newErrorWithRootCodespace(CodeUnauthorized, msg)
}
func ErrInsufficientFunds(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientFunds, msg)
}
func ErrUnknownRequest(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownRequest, msg)
}
func ErrInvalidAddress(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidAddress, msg)
}
func ErrUnknownAddress(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownAddress, msg)
}
func ErrInvalidPubKey(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidPubKey, msg)
}
func ErrInsufficientCoins(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientCoins, msg)
}
func ErrInvalidCoins(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidCoins, msg)
}
func ErrOutOfGas(msg string) Error {
	return newErrorWithRootCodespace(CodeOutOfGas, msg)
}
func ErrMemoTooLarge(msg string) Error {
	return newErrorWithRootCodespace(CodeMemoTooLarge, msg)
}
func ErrInsufficientFee(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientFee, msg)
}
func ErrTooManySignatures(msg string) Error {
	return newErrorWithRootCodespace(CodeTooManySignatures, msg)
}
func ErrNoSignatures(msg string) Error {
	return newErrorWithRootCodespace(CodeNoSignatures, msg)
}

func ErrTooMuchPrecision(msg string) Error {
	return newErrorWithRootCodespace(CodeTooMuchPrecision, msg)
}

func ErrGasOverflow(msg string) Error {
	return newErrorWithRootCodespace(CodeGasOverflow, msg)
}

func ErrNotFoundAsset(msg string) Error {
	return newErrorWithRootCodespace(CodeNotFoundAsset, msg)
}

func ErrDuplicatedUtxo(msg string) Error {
	return newErrorWithRootCodespace(CodeDuplicatedUtxo, msg)
}
func ErrUnknownUtxo(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownUtxo, msg)
}

func ErrMismatchUtxoAmount(msg string) Error {
	return newErrorWithRootCodespace(CodeMismatchUtxoAmount, msg)
}

func ErrInvalidCollect(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidCollect, msg)
}

func ErrInvalidAmount(msg string) Error {
	return newErrorWithRootCodespace(CodeAmountError, msg)
}
func ErrInvalidAsset(msg string) Error {
	return newErrorWithRootCodespace(CodeAssetError, msg)
}

func ErrInvalidUtxo(msg string) Error {
	return newErrorWithRootCodespace(CodeUtxoError, msg)
}

func ErrInvalidAccount(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidAccount, msg)
}

func ErrInvalidOnhold(msg string) Error {
	return newErrorWithRootCodespace(CodeOnHoldError, msg)
}

func ErrInvalidBlkNumber(msg string) Error {
	return newErrorWithRootCodespace(CodeBlkNumberError, msg)

}

func ErrInvalidAddr(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidAddress, msg)
}

func ErrAlreadyExitSymbol(msg string) Error {
	return newErrorWithRootCodespace(CodeSymbolAlreadyExist, msg)
}

func ErrInvalidSymbol(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidSymbol, msg)
}

func ErrUnSupportToken(msg string) Error {
	return newErrorWithRootCodespace(CodeUnsupportToken, msg)
}

func ErrUnSupportAddressType(msg string) Error {
	return newErrorWithRootCodespace(CodeUnsupportAddressType, msg)
}

func ErrNotFoundOrder(msg string) Error {
	return newErrorWithRootCodespace(CodeNotFoundOrder, msg)
}

func ErrInvalidOrder(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidOrder, msg)
}

func ErrInsufficientValidatorNumForKeyGen(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientValidtorNumberForKeyGen, msg)
}

func ErrInsufficientValidatorNum(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientValidtorNumber, msg)
}

func ErrInvalidTx(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidTx, msg)
}

func ErrTransactionIsNotEnabled(msg string) Error {
	return newErrorWithRootCodespace(CodeTransactionIsNotEnabled, msg)
}

func ErrInvalidCUType(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidCUType, msg)
}

func ErrEmptyDBGet(msg string) Error {
	return newErrorWithRootCodespace(CodeEmptyDBGet, msg)
}

func ErrMigrationInProgress(msg string) Error {
	return newErrorWithRootCodespace(CodeMigrationInProgress, msg)
}

func ErrSystemBusy(msg string) Error {
	return newErrorWithRootCodespace(CodeSystemBusy, msg)
}

func ErrPreKeyGenTooMany(msg string) Error {
	return newErrorWithRootCodespace(CodePreKeyGenTooMany, msg)
}

func ErrWaitAssignTooMany(msg string) Error {
	return newErrorWithRootCodespace(CodeWaitAssignTooMany, msg)
}

//----------------------------------------
// Error & bhError

type cmnError = cmn.Error

// sdk Error type
type Error interface {
	// Implements cmn.Error
	// Error() string
	// Stacktrace() cmn.Error
	// Trace(offset int, format string, args ...interface{}) cmn.Error
	// Data() interface{}
	cmnError

	// convenience
	TraceSDK(format string, args ...interface{}) Error

	// set codespace
	WithDefaultCodespace(CodespaceType) Error

	Code() CodeType
	Codespace() CodespaceType
	ABCILog() string
	Result() Result
	QueryResult() abci.ResponseQuery
}

// NewError - create an error.
func NewError(codespace CodespaceType, code CodeType, format string, args ...interface{}) Error {
	return newError(codespace, code, format, args...)
}

func newErrorWithRootCodespace(code CodeType, format string, args ...interface{}) *bhError {
	return newError(CodespaceRoot, code, format, args...)
}

func newError(codespace CodespaceType, code CodeType, format string, args ...interface{}) *bhError {
	if format == "" {
		format = CodeToDefaultMsg(code)
	}
	return &bhError{
		codespace: codespace,
		code:      code,
		cmnError:  cmn.NewError(format, args...),
	}
}

type bhError struct {
	codespace CodespaceType
	code      CodeType
	cmnError
}

// Implements Error.
func (err *bhError) WithDefaultCodespace(cs CodespaceType) Error {
	codespace := err.codespace
	if codespace == CodespaceUndefined {
		codespace = cs
	}
	return &bhError{
		codespace: cs,
		code:      err.code,
		cmnError:  err.cmnError,
	}
}

// Implements ABCIError.
// nolint: errcheck
func (err *bhError) TraceSDK(format string, args ...interface{}) Error {
	err.Trace(1, format, args...)
	return err
}

// Implements ABCIError.
func (err *bhError) Error() string {
	return fmt.Sprintf(`ERROR:
Codespace: %s
Code: %d
Message: %#v
`, err.codespace, err.code, err.cmnError.Error())
}

// Implements Error.
func (err *bhError) Codespace() CodespaceType {
	return err.codespace
}

// Implements Error.
func (err *bhError) Code() CodeType {
	return err.code
}

// Implements ABCIError.
func (err *bhError) ABCILog() string {
	errMsg := err.cmnError.Error()
	jsonErr := humanReadableError{
		Codespace: err.codespace,
		Code:      err.code,
		Message:   errMsg,
	}

	var buff bytes.Buffer
	enc := json.NewEncoder(&buff)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(jsonErr); err != nil {
		panic(errors.Wrap(err, "failed to encode ABCI error log"))
	}

	return strings.TrimSpace(buff.String())
}

func (err *bhError) Result() Result {
	return Result{
		Code:      err.Code(),
		Codespace: err.Codespace(),
		Log:       err.ABCILog(),
	}
}

// QueryResult allows us to return Error.QueryResult() in query responses
func (err *bhError) QueryResult() abci.ResponseQuery {
	return abci.ResponseQuery{
		Code:      uint32(err.Code()),
		Codespace: string(err.Codespace()),
		Log:       err.ABCILog(),
	}
}

//----------------------------------------
// REST error utilities

// appends a message to the head of the given error
func AppendMsgToErr(msg string, err string) string {
	msgIdx := strings.Index(err, "message\":\"")
	if msgIdx != -1 {
		errMsg := err[msgIdx+len("message\":\"") : len(err)-2]
		errMsg = fmt.Sprintf("%s; %s", msg, errMsg)
		return fmt.Sprintf("%s%s%s",
			err[:msgIdx+len("message\":\"")],
			errMsg,
			err[len(err)-2:],
		)
	}
	return fmt.Sprintf("%s; %s", msg, err)
}

// returns the index of the message in the ABCI Log
// nolint: deadcode unused
func mustGetMsgIndex(abciLog string) int {
	msgIdx := strings.Index(abciLog, "message\":\"")
	if msgIdx == -1 {
		panic(fmt.Sprintf("invalid error format: %s", abciLog))
	}
	return msgIdx + len("message\":\"")
}

// parses the error into an object-like struct for exporting
type humanReadableError struct {
	Codespace CodespaceType `json:"codespace"`
	Code      CodeType      `json:"code"`
	Message   string        `json:"message"`
}
