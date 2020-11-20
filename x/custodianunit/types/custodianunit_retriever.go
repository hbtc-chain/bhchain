package types

import (
	"fmt"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/exported"
)

// NodeQuerier is an interface that is satisfied by types that provide the QueryWithData method
type NodeQuerier interface {
	// QueryWithData performs a query to a Tendermint node with the provided path
	// and a data payload. It returns the result and height of the query upon success
	// or an error if the query fails.
	QueryWithData(path string, data []byte) ([]byte, int64, error)
}

// CURetriever defines the properties of a type that can be used to
// retrieve CUs.
type CURetriever struct {
	querier NodeQuerier
}

// NewCURetriever initialises a new CURetriever instance.
func NewCURetriever(querier NodeQuerier) CURetriever {
	return CURetriever{querier: querier}
}

// GetCU queries for an CustodianUnit given an address and a block height. An
// error is returned if the query or decoding fails.
func (cr CURetriever) GetCU(addr sdk.CUAddress) (exported.CustodianUnit, error) {
	CU, _, err := cr.GetCUWithHeight(addr)
	return CU, err
}

// GetCUWithHeight queries for an CustodianUnit given an address. Returns the
// height of the query with the CustodianUnit. An error is returned if the query
// or decoding fails.
func (cr CURetriever) GetCUWithHeight(addr sdk.CUAddress) (exported.CustodianUnit, int64, error) {
	bs, err := ModuleCdc.MarshalJSON(NewQueryCUParams(addr))
	if err != nil {
		return nil, 0, err
	}

	res, height, err := cr.querier.QueryWithData(fmt.Sprintf("custom/%s/%s", QuerierRoute, QueryCU), bs)
	if err != nil {
		return nil, height, err
	}

	var CU exported.CustodianUnit
	if err := ModuleCdc.UnmarshalJSON(res, &CU); err != nil {
		return nil, height, err
	}

	return CU, height, nil
}

func (cr CURetriever) GetOpCUWithHeight(symbol string) (sdk.OpCUsAstInfo, int64, error) {
	token := sdk.Symbol(symbol)
	if !token.IsValid() {
		return nil, 0, fmt.Errorf("invalid symbol:%v", symbol)
	}

	return nil, 0, nil
}

// EnsureExists returns an error if no CustodianUnit exists for the given address else nil.
func (cr CURetriever) EnsureExists(addr sdk.CUAddress) error {
	if _, err := cr.GetCU(addr); err != nil {
		return err
	}
	return nil
}

// GetSequence returns sequence  for the given address.
// It returns an error if the CustodianUnit couldn't be retrieved from the state.
func (cr CURetriever) GetSequence(addr sdk.CUAddress) (uint64, error) {
	cu, err := cr.GetCU(addr)
	if err != nil {
		return 0, err
	}
	return cu.GetSequence(), nil
}
