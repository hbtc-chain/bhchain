package token

import (
	"bytes"
	"fmt"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/evidence/exported"
	"github.com/hbtc-chain/bhchain/x/params"
	"github.com/hbtc-chain/bhchain/x/token/internal"
	"github.com/hbtc-chain/bhchain/x/token/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	TokenStoreKeyPrefix = []byte{0x01}
)

type TokenKeeper interface {
	GetSymbol(ctx sdk.Context, symbol sdk.Symbol) sdk.Symbol
	SetTokenInfo(ctx sdk.Context, tokenInfo *sdk.TokenInfo)
	GetTokenInfo(ctx sdk.Context, symbol sdk.Symbol) *sdk.TokenInfo
	GetIssuer(ctx sdk.Context, symbol sdk.Symbol) string
	GetChain(ctx sdk.Context, symbol sdk.Symbol) sdk.Symbol
	GetTokenType(ctx sdk.Context, symbol sdk.Symbol) sdk.TokenType
	IsUtxoBased(ctx sdk.Context, symbol sdk.Symbol) bool
	IsSubToken(ctx sdk.Context, symbol sdk.Symbol) bool
	IsSendEnabled(ctx sdk.Context, symbol sdk.Symbol) bool
	IsDepositEnabled(ctx sdk.Context, symbol sdk.Symbol) bool
	IsWithdrawalEnabled(ctx sdk.Context, symbol sdk.Symbol) bool
	GetDecimals(ctx sdk.Context, symbol sdk.Symbol) uint64
	GetTotalSupply(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetCollectThreshold(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetOpenFee(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetSysOpenFee(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetWithdrawalFeeRate(ctx sdk.Context, symbol sdk.Symbol) sdk.Dec
	EnableSend(ctx sdk.Context, symbol sdk.Symbol)
	DisableSend(ctx sdk.Context, symbol sdk.Symbol)
	EnableDeposit(ctx sdk.Context, symbol sdk.Symbol)
	DisableDeposit(ctx sdk.Context, symbol sdk.Symbol)
	EnableWithdrawal(ctx sdk.Context, symbol sdk.Symbol)
	DisableWithdrawal(ctx sdk.Context, symbol sdk.Symbol)
	GetSymbols(ctx sdk.Context) []string
	IsTokenSupported(ctx sdk.Context, symbol sdk.Symbol) bool
	GetMaxOpCUNumber(ctx sdk.Context, symbol sdk.Symbol) uint64
	GetAllTokenInfo(ctx sdk.Context) []sdk.TokenInfo
	GetDepositThreshold(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetGasLimit(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetGasPrice(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetOpCUSystransferAmount(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	GetSystransferAmount(ctx sdk.Context, symbol sdk.Symbol) sdk.Int
	SynGasPrice(ctx sdk.Context, fromAddr string, height uint64, tokensgasFee []sdk.TokensGasPrice) ([]sdk.TokensGasPrice, sdk.Result)
} //nolint

type Keeper struct {
	storeKey       sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc            *codec.Codec // The wire codec for binary encoding/decoding
	sk             internal.StakingKeeper
	paramSubSpace  params.Subspace
	evidenceKeeper internal.EvidenceKeeper
}

func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, paramSubSpace params.Subspace) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubSpace: paramSubSpace.WithKeyTable(types.ParamKeyTable()),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SetStakingKeeper(sk internal.StakingKeeper) {
	k.sk = sk
}

func (k *Keeper) SetEvidenceKeeper(evidenceKeeper internal.EvidenceKeeper) {
	k.evidenceKeeper = evidenceKeeper
}

func tokenStoreKey(symbol string) []byte {
	return append(TokenStoreKeyPrefix, []byte(symbol)...)
}

var _ TokenKeeper = (*Keeper)(nil)

//Set entire TokenInfo
func (k *Keeper) SetTokenInfo(ctx sdk.Context, tokenInfo *sdk.TokenInfo) {
	store := ctx.KVStore(k.storeKey)
	if !tokenInfo.Symbol.IsValidTokenName() || !tokenInfo.Chain.IsValidTokenName() {
		panic(fmt.Sprintf("invalid token symbol:%v or chain:%v", tokenInfo.Symbol, tokenInfo.Chain))
	}

	store.Set(tokenStoreKey(tokenInfo.Symbol.String()), k.cdc.MustMarshalBinaryBare(tokenInfo))
}

//Delete entire TokenInfo
func (k *Keeper) DeletTokenInfo(ctx sdk.Context, symbol string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(tokenStoreKey(symbol))
}

func (k *Keeper) GetAllTokenInfo(ctx sdk.Context) []sdk.TokenInfo {
	store := ctx.KVStore(k.storeKey)
	var tokens []sdk.TokenInfo
	iter := sdk.KVStorePrefixIterator(store, TokenStoreKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var tokenInfo sdk.TokenInfo
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &tokenInfo)
		tokens = append(tokens, tokenInfo)
	}
	return tokens
}

//TODO(keep), add cache later
func (k *Keeper) GetTokenInfo(ctx sdk.Context, symbol sdk.Symbol) *sdk.TokenInfo {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(tokenStoreKey(symbol.String())) {
		return nil
	}

	bz := store.Get(tokenStoreKey(symbol.String()))
	var tokenInfo sdk.TokenInfo
	k.cdc.MustUnmarshalBinaryBare(bz, &tokenInfo)

	return &tokenInfo
}

func (k *Keeper) GetSymbol(ctx sdk.Context, symbol sdk.Symbol) sdk.Symbol {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return ""
	}
	return token.Symbol
}

func (k *Keeper) GetIssuer(ctx sdk.Context, symbol sdk.Symbol) string {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return ""
	}
	return token.Issuer
}

func (k *Keeper) GetChain(ctx sdk.Context, symbol sdk.Symbol) sdk.Symbol {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return ""
	}
	return token.Chain
}

func (k *Keeper) GetTokenType(ctx sdk.Context, symbol sdk.Symbol) sdk.TokenType {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.TokenType(0)
	}
	return token.TokenType
}

func (k *Keeper) IsUtxoBased(ctx sdk.Context, symbol sdk.Symbol) bool {
	return sdk.UtxoBased == k.GetTokenType(ctx, symbol)
}

func (k *Keeper) IsTokenSupported(ctx sdk.Context, symbol sdk.Symbol) bool {
	store := ctx.KVStore(k.storeKey)
	if !symbol.IsValidTokenName() {
		return false
	}
	if store.Has(tokenStoreKey(symbol.String())) {
		return true
	}
	return false
}

func (k *Keeper) IsSendEnabled(ctx sdk.Context, symbol sdk.Symbol) bool {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return false
	}

	return token.IsSendEnabled
}

func (k *Keeper) IsSubToken(ctx sdk.Context, symbol sdk.Symbol) bool {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return false
	}
	if token.Chain != symbol {
		return true
	}
	return false
}

func (k *Keeper) IsDepositEnabled(ctx sdk.Context, symbol sdk.Symbol) bool {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return false
	}
	//return token.IsDepositEnabled && k.bk.GetSendEnabled(ctx)
	return token.IsDepositEnabled
}

func (k *Keeper) IsWithdrawalEnabled(ctx sdk.Context, symbol sdk.Symbol) bool {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return false
	}
	//return token.IsWithdrawalEnabled && k.bk.GetSendEnabled(ctx)
	return token.IsWithdrawalEnabled
}

func (k *Keeper) GetDecimals(ctx sdk.Context, symbol sdk.Symbol) uint64 {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return 0
	}
	return token.Decimals
}

func (k *Keeper) GetTotalSupply(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.TotalSupply
}

func (k *Keeper) GetMaxOpCUNumber(ctx sdk.Context, symbol sdk.Symbol) uint64 {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return 0
	}
	return token.MaxOpCUNumber
}

func (k *Keeper) GetCollectThreshold(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.CollectThreshold
}

func (k *Keeper) GetDepositThreshold(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.DepositThreshold
}

func (k *Keeper) GetOpenFee(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.OpenFee
}

func (k *Keeper) GetSysOpenFee(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.SysOpenFee
}

func (k *Keeper) GetWithdrawalFeeRate(ctx sdk.Context, symbol sdk.Symbol) sdk.Dec {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.ZeroDec()
	}
	return token.WithdrawalFeeRate
}

func (k *Keeper) EnableSend(ctx sdk.Context, symbol sdk.Symbol) {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return
	}
	token.IsSendEnabled = true
	k.SetTokenInfo(ctx, token)
}

func (k *Keeper) DisableSend(ctx sdk.Context, symbol sdk.Symbol) {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return
	}
	token.IsSendEnabled = false
	k.SetTokenInfo(ctx, token)
}

func (k *Keeper) EnableDeposit(ctx sdk.Context, symbol sdk.Symbol) {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return
	}
	token.IsDepositEnabled = true
	k.SetTokenInfo(ctx, token)
}

func (k *Keeper) DisableDeposit(ctx sdk.Context, symbol sdk.Symbol) {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return
	}
	token.IsDepositEnabled = false
	k.SetTokenInfo(ctx, token)
}

func (k *Keeper) EnableWithdrawal(ctx sdk.Context, symbol sdk.Symbol) {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return
	}
	token.IsWithdrawalEnabled = true
	k.SetTokenInfo(ctx, token)
}

func (k *Keeper) DisableWithdrawal(ctx sdk.Context, symbol sdk.Symbol) {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return
	}
	token.IsWithdrawalEnabled = false
	k.SetTokenInfo(ctx, token)
}

func (k *Keeper) GetSymbolIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, TokenStoreKeyPrefix)
}

func (k *Keeper) GetSymbols(ctx sdk.Context) []string {
	var symbols []string
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, TokenStoreKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		symbols = append(symbols, string(bytes.TrimPrefix(iter.Key(), TokenStoreKeyPrefix)))
	}
	return symbols
}

func (k *Keeper) GetSystransferAmount(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.SysTransferAmount()
}

func (k *Keeper) GetOpCUSystransferAmount(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.OpCUSysTransferAmount()
}

func (k *Keeper) GetGasLimit(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.GasLimit
}

func (k *Keeper) GetGasPrice(ctx sdk.Context, symbol sdk.Symbol) sdk.Int {
	token := k.GetTokenInfo(ctx, symbol)
	if token == nil {
		return sdk.NewInt(0)
	}
	return token.GasPrice
}

// SetParams sets the token module's parameters.
func (k *Keeper) SetParams(ctx sdk.Context, params Params) {
	k.paramSubSpace.SetParamSet(ctx, &params)
}

// GetParams gets the token module's parameters.
func (k *Keeper) GetParams(ctx sdk.Context) (params Params) {
	k.paramSubSpace.GetParamSet(ctx, &params)
	return
}

func (k *Keeper) SynGasPrice(ctx sdk.Context, fromAddr string, height uint64, tokensGasPrice []sdk.TokensGasPrice) ([]sdk.TokensGasPrice, sdk.Result) {
	curBlockHeight := uint64(ctx.BlockHeight())
	if height >= curBlockHeight || curBlockHeight-height > sdk.GasPriceBucketWindow {
		return nil, sdk.ErrInvalidTx(fmt.Sprintf("invalid height %d, current block height is %d", height, curBlockHeight)).Result()
	}

	address, err := sdk.CUAddressFromBase58(fromAddr)
	if err != nil {
		return nil, sdk.ErrInvalidAddress(fmt.Sprintf("can't decode addr:%s", fromAddr)).Result()
	}
	bValidator, validatorNum := k.sk.IsActiveKeyNode(ctx, address)
	if validatorNum == 0 {
		return nil, sdk.ErrInsufficientValidatorNum(fmt.Sprintf("validator's number:%v", validatorNum)).Result()
	}
	if !bValidator {
		return nil, sdk.ErrInvalidTx(fmt.Sprintf("FromCu: %v is not a validator", fromAddr)).Result()
	}
	for _, item := range tokensGasPrice {
		if !k.IsTokenSupported(ctx, sdk.Symbol(item.Chain)) {
			return nil, sdk.ErrInvalidTx(fmt.Sprintf("Chain %s not support", item.Chain)).Result()
		}
	}

	validGasPrice := make([]sdk.TokensGasPrice, 0)
	bucket := height / sdk.GasPriceBucketWindow

	for _, item := range tokensGasPrice {
		voteID := fmt.Sprintf("%s-%d", item.Chain, bucket)
		firstConfirmed, _, validVotes := k.evidenceKeeper.VoteWithCustomBox(ctx, voteID, address, item.GasPrice, curBlockHeight, types.NewGasPriceVoteBox)
		if firstConfirmed {
			k.updateGasPrice(ctx, item.Chain, validVotes)
		}
		validGasPrice = append(validGasPrice, item)
	}

	return validGasPrice, sdk.Result{}
}

func (k *Keeper) updateGasPrice(ctx sdk.Context, chain string, validVotes []*exported.VoteItem) {
	totalGasFee := sdk.ZeroInt()
	var count int64
	for _, item := range validVotes {
		price, ok := item.Vote.(sdk.Int)
		if !ok {
			continue
		}
		totalGasFee = totalGasFee.Add(price)
		count++
	}
	if count > 0 {
		averageGasPrice := totalGasFee.QuoRaw(count)
		chainSymbol := sdk.Symbol(chain)
		tokenInfos := k.GetAllTokenInfo(ctx)
		for _, tokenInfo := range tokenInfos {
			if tokenInfo.Chain == chainSymbol {
				tokenInfo.GasPrice = averageGasPrice
				k.SetTokenInfo(ctx, &tokenInfo)
			}
		}
	}
}
