package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/hbtc-chain/bhchain/client"
	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
	"github.com/hbtc-chain/bhchain/x/evidence/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {

		switch path[0] {
		case types.QueryEvidence:
			res, err = queryEvidence(ctx, req, k)

		case types.QueryAllEvidence:
			res, err = queryAllEvidence(ctx, req, k)

		default:
			err = sdk.ErrUnknownRequest(fmt.Sprintf("unknown %s query endpoint: %s", types.ModuleName, path[0]))
		}

		return res, err
	}
}

func queryEvidence(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryEvidenceParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	hash, err := hex.DecodeString(params.EvidenceHash)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to decode evidence hash string query: %s", err))
	}

	evidence, ok := k.GetEvidence(ctx, hash)
	if !ok {
		return nil, types.ErrNoEvidenceExists(params.EvidenceHash)
	}

	res, err := codec.MarshalJSONIndent(k.cdc, evidence)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func queryAllEvidence(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryAllEvidenceParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	evidence := k.GetAllEvidence(ctx)

	start, end := client.Paginate(len(evidence), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		evidence = []exported.Evidence{}
	} else {
		evidence = evidence[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, evidence)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}
