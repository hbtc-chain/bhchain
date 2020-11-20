package custodianunit

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/hbtc-chain/bhchain/x/custodianunit/internal"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/hbtc-chain/bhchain/codec"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/custodianunit/types"
)

const (
	DefaultGasUsedForRefund = uint64(20000)
)

var (
	// simulation signature values used to estimate gas consumption
	simSecp256k1Pubkey secp256k1.PubKeySecp256k1
	simSecp256k1Sig    [64]byte
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(simSecp256k1Pubkey[:], bz)
}

// SignatureVerificationGasConsumer is the type of function that is used to both consume gas when verifying signatures
// and also to accept or reject different types of PubKey's. This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params Params) sdk.Result

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & CU numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ck CUKeeper, supplyKeeper internal.SupplyKeeper, stakingKeeper internal.StakingKeeper, sigGasConsumer SignatureVerificationGasConsumer) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		if addr := supplyKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
			panic(fmt.Sprintf("%s module CU has not been set", types.FeeCollectorName))
		}

		// all transactions must be of type custodianunit.StdTx
		stdTx, ok := tx.(StdTx)
		if !ok {
			// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
			// during runTx.
			newCtx = SetGasMeter(simulate, ctx, 0)
			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		params := ck.GetParams(ctx)

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only ran on check tx.
		if ctx.IsCheckTx() && !simulate {
			res := EnsureSufficientMempoolFees(ctx, stdTx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		newCtx = SetGasMeter(simulate, ctx, stdTx.Fee.Gas)

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch rType := r.(type) {
				case sdk.ErrorOutOfGas:
					log := fmt.Sprintf(
						"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
						rType.Descriptor, stdTx.Fee.Gas, newCtx.GasMeter().GasConsumed(),
					)
					res = sdk.ErrOutOfGas(log).Result()

					res.GasWanted = stdTx.Fee.Gas
					res.GasUsed = newCtx.GasMeter().GasConsumed()
					abort = true
				default:
					panic(r)
				}
			}
		}()

		if res := ValidateSigCount(stdTx, params); !res.IsOK() {
			return newCtx, res, true
		}

		if err := tx.ValidateBasic(); err != nil {
			return newCtx, err.Result(), true
		}

		newCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(newCtx.TxBytes())), "txSize")

		if res := ValidateMemo(stdTx, params); !res.IsOK() {
			return newCtx, res, true
		}

		// stdSigs contains the sequence number, CU number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()
		signerAccs := make([]CU, len(signerAddrs))
		isGenesis := ctx.BlockHeight() == 0

		// Check CU type
		res = CheckCUType(newCtx, ck, signerAddrs)
		if !res.IsOK() {
			return newCtx, res, true
		}

		// fetch first signer, who's going to pay the fees
		signerAccs[0] = ck.GetOrNewCU(ctx, sdk.CUTypeUser, signerAddrs[0])

		// deduct the fees
		if !stdTx.Fee.Amount.IsZero() {
			shouldDeductFee := true
			if !ctx.IsCheckTx() && stdTx.IsSettleTx() {
				isActiveKeyNode, _ := stakingKeeper.IsActiveKeyNode(newCtx, signerAccs[0].GetAddress())
				if isActiveKeyNode && ctx.SettleQuota().Consume(signerAccs[0].GetAddress().Bytes(), stdTx.Fee.Gas) {
					shouldDeductFee = false
				}
			}
			if shouldDeductFee {
				res = DeductFees(supplyKeeper, newCtx, signerAccs[0], stdTx.Fee.Amount)
				if !res.IsOK() {
					return newCtx, res, true
				}
			}
		}

		// check for settle only msg is sent from settle
		if stdTx.IsSettleOnlyTx() {
			if ok, _ := stakingKeeper.IsActiveKeyNode(newCtx, signerAccs[0].GetAddress()); !ok {
				return newCtx, sdk.ErrInvalidAccount(
					fmt.Sprintf("invalid account for settle only msg"),
				).Result(), true
			}
		}

		// stdSigs contains the sequence number, CU number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := stdTx.GetSignatures()

		for i := 0; i < len(stdSigs); i++ {
			// skip the fee payer, CU is cached and fees were deducted already
			if i != 0 {
				signerAccs[i] = ck.GetOrNewCU(ctx, sdk.CUTypeUser, signerAddrs[i])
			}

			// check signature, return CU with incremented nonce
			signBytes := GetSignBytes(newCtx.ChainID(), stdTx, signerAccs[i], isGenesis)
			signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytes, simulate, params, sigGasConsumer)
			if !res.IsOK() {
				return newCtx, res, true
			}

			ck.SetCU(newCtx, signerAccs[i])
		}

		// TODO: tx tags (?)
		return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false // continue...
	}
}

func NewGasRefundHandler(supplyKeeper internal.SupplyKeeper) sdk.GasRefundHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, gasWanted, gasUsed uint64,
	) bool {

		stdTx := tx.(StdTx)
		if gasWanted < gasUsed+DefaultGasUsedForRefund || stdTx.IsSettleTx() {
			return false
		}
		signerAddrs := stdTx.GetSigners()
		usedFraction := sdk.NewDec(int64(gasUsed + DefaultGasUsedForRefund)).Quo(sdk.NewDec(int64(gasWanted)))
		refundFraction := sdk.OneDec().Sub(usedFraction)
		var refundCoins sdk.Coins
		for _, coin := range stdTx.Fee.Amount {
			amount := sdk.NewDecFromInt(coin.Amount).Mul(refundFraction).TruncateInt()
			refundCoins = refundCoins.Add(sdk.NewCoins(sdk.NewCoin(coin.Denom, amount)))
		}
		if !refundCoins.IsValid() {
			return false
		}
		supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.FeeCollectorName, signerAddrs[0], refundCoins)

		return true
	}
}

func CheckCUType(ctx sdk.Context, ck CUKeeper, signers []sdk.CUAddress) sdk.Result {
	for _, s := range signers {
		if cu := ck.GetCU(ctx, s); cu != nil {
			if cu.GetCUType() == sdk.CUTypeOp {
				return sdk.ErrInvalidAddr(fmt.Sprintf("OP CU %s can not sign any message", s)).Result()
			}
		}
	}

	return sdk.Result{}
}

// ValidateSigCount validates that the transaction has a valid cumulative total
// amount of signatures.
func ValidateSigCount(stdTx StdTx, params Params) sdk.Result {
	stdSigs := stdTx.GetSignatures()

	sigCount := 0
	for i := 0; i < len(stdSigs); i++ {
		sigCount += CountSubKeys(stdSigs[i].PubKey)
		if uint64(sigCount) > params.TxSigLimit {
			return sdk.ErrTooManySignatures(
				fmt.Sprintf("signatures: %d, limit: %d", sigCount, params.TxSigLimit),
			).Result()
		}
	}

	return sdk.Result{}
}

// ValidateMemo validates the memo size.
func ValidateMemo(stdTx StdTx, params Params) sdk.Result {
	memoLength := len(stdTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return sdk.ErrMemoTooLarge(
			fmt.Sprintf(
				"maximum number of characters is %d but received %d characters",
				params.MaxMemoCharacters, memoLength,
			),
		).Result()
	}

	return sdk.Result{}
}

// verify the signature and increment the sequence. If the CU doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, cu CU, sig StdSignature, signBytes []byte, simulate bool, params Params,
	sigGasConsumer SignatureVerificationGasConsumer,
) (updatedAcc CU, res sdk.Result) {
	pubKey, res := ProcessPubKey(cu, sig, simulate)
	if !res.IsOK() {
		return nil, res
	}

	err := cu.SetPubKey(pubKey)
	if err != nil {
		return nil, sdk.ErrInternal("setting PubKey on signer's cu").Result()
	}

	if simulate {
		// Simulated txs should not contain a signature and are not required to
		// contain a pubkey, so we must CU for tx size of including a
		// StdSignature (Amino encoding) and simulate gas consumption
		// (assuming a SECP256k1 simulation key).
		consumeSimSigGas(ctx.GasMeter(), pubKey, sig, params)
	}

	if res := sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params); !res.IsOK() {
		return nil, res
	}

	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed; verify correct CU sequence and chain-id").Result()
	}

	if err := cu.SetSequence(cu.GetSequence() + 1); err != nil {
		panic(err)
	}

	return cu, res
}

func consumeSimSigGas(gasmeter sdk.GasMeter, pubkey crypto.PubKey, sig StdSignature, params Params) {
	simSig := StdSignature{PubKey: pubkey}
	if len(sig.Signature) == 0 {
		simSig.Signature = simSecp256k1Sig[:]
	}

	sigBz := ModuleCdc.MustMarshalBinaryLengthPrefixed(simSig)
	cost := sdk.Gas(len(sigBz) + 6)

	// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
	// number of signers.
	if _, ok := pubkey.(multisig.PubKeyMultisigThreshold); ok {
		cost *= params.TxSigLimit
	}

	gasmeter.ConsumeGas(params.TxSizeCostPerByte*cost, "txSize")
}

// ProcessPubKey verifies that the given CU address matches that of the
// StdSignature. In addition, it will set the public key of the CU if it
// has not been set.
func ProcessPubKey(cu CU, sig StdSignature, simulate bool) (crypto.PubKey, sdk.Result) {
	// If pubkey is not known for CU, set it from the StdSignature.
	pubKey := cu.GetPubKey()
	if simulate {
		// In simulate mode the transaction comes with no signatures, thus if the
		// CU's pubkey is nil, both signature verification and gasKVStore.Set()
		// shall consume the largest amount, i.e. it takes more gas to verify
		// secp256k1 keys than ed25519 ones.
		if pubKey == nil {
			return simSecp256k1Pubkey, sdk.Result{}
		}

		return pubKey, sdk.Result{}
	}

	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, sdk.ErrInvalidPubKey("PubKey not found").Result()
		}

		if !bytes.Equal(pubKey.Address(), cu.GetAddress()) {
			return nil, sdk.ErrInvalidPubKey(
				fmt.Sprintf("PubKey does not match Signer address %s", cu.GetAddress())).Result()
		}
	}

	return pubKey, sdk.Result{}
}

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params Params,
) sdk.Result {
	switch pubkey := pubkey.(type) {
	case ed25519.PubKeyEd25519:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return sdk.ErrInvalidPubKey("ED25519 public keys are unsupported").Result()

	case secp256k1.PubKeySecp256k1:
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return sdk.Result{}

	case multisig.PubKeyMultisigThreshold:
		var multisignature multisig.Multisignature
		codec.Cdc.MustUnmarshalBinaryBare(sig, &multisignature)

		consumeMultisignatureVerificationGas(meter, multisignature, pubkey, params)
		return sdk.Result{}

	default:
		return sdk.ErrInvalidPubKey(fmt.Sprintf("unrecognized public key type: %T", pubkey)).Result()
	}
}

func consumeMultisignatureVerificationGas(meter sdk.GasMeter,
	sig multisig.Multisignature, pubkey multisig.PubKeyMultisigThreshold,
	params Params) {

	size := sig.BitArray.Size()
	sigIndex := 0
	for i := 0; i < size; i++ {
		if sig.BitArray.GetIndex(i) {
			DefaultSigVerificationGasConsumer(meter, sig.Sigs[sigIndex], pubkey.PubKeys[i], params)
			sigIndex++
		}
	}
}

// DeductFees deducts fees from the given CU.
//
// NOTE: We could use the CoinKeeper (in addition to the CUKeeper, because
// the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func DeductFees(supplyKeeper internal.SupplyKeeper, ctx sdk.Context, cu CU, fees sdk.Coins) sdk.Result {
	_, err := supplyKeeper.SendCoinsFromAccountToModule(ctx, cu.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, stdFee StdFee) sdk.Result {
	minGasPrices := ctx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(stdFee.Gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !stdFee.Amount.IsAnyGTE(requiredFees) {
			return sdk.ErrInsufficientFee(
				fmt.Sprintf(
					"insufficient fees; got: %q required: %q", stdFee.Amount, requiredFees,
				),
			).Result()
		}
	}

	return sdk.Result{}
}

// SetGasMeter returns a new context with a gas meter set from a given context.
func SetGasMeter(simulate bool, ctx sdk.Context, gasLimit uint64) sdk.Context {
	// In various cases such as simulation and during the genesis block, we do not
	// meter any gas utilization.
	if simulate || ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))
}

// GetSignBytes returns a slice of bytes to sign over for a given transaction
// and an CU.
func GetSignBytes(chainID string, stdTx StdTx, cu CU, genesis bool) []byte {
	return StdSignBytes(
		chainID, cu.GetSequence(), stdTx.Fee, stdTx.Msgs, stdTx.Memo,
	)
}
