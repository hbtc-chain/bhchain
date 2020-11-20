package hrc10

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"golang.org/x/crypto/ripemd160"

	"github.com/hbtc-chain/bhchain/base58"
	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/hrc10/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgNewToken:
			return handleMsgNewToken(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized hrc10 Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgNewToken(ctx sdk.Context, keeper Keeper, msg types.MsgNewToken) sdk.Result {
	ctx.Logger().Info("handleMsgNewToken", "msg", msg)

	symbol := calSymbol(ctx)
	if keeper.tk.HasToken(ctx, symbol) {
		return sdk.ErrAlreadyExitSymbol(fmt.Sprintf("token %s already exist", symbol)).Result()
	}


	issueFee := sdk.NewCoin(sdk.NativeToken, keeper.GetParams(ctx).IssueTokenFee)
	_, flow, err := keeper.trk.SubCoin(ctx, msg.From, issueFee)
	if err != nil {
		return err.Result()
	}

	// transfer openFee to communityPool
	keeper.dk.AddToFeePool(ctx, sdk.NewDecCoins(sdk.NewCoins(issueFee)))

	totalSupply := msg.TotalSupply
	ti := &sdk.BaseToken{
		Name:        msg.Name,
		Symbol:      symbol,
		Chain:       sdk.NativeToken,
		Issuer:      msg.From.String(),
		SendEnabled: true,
		Decimals:    msg.Decimals,
		TotalSupply: totalSupply,
	}
	e := keeper.tk.CreateToken(ctx, ti)
	if e != nil {
		return sdk.ErrInternal(e.Error()).Result()
	}

	flows := make([]sdk.Flow, 0, 2)
	flows = append(flows, flow)

	//set tokeninfo
	//minted newCoins
	mintedCoin := sdk.NewCoin(ti.Symbol.String(), totalSupply)
	_, flow, err = keeper.trk.AddCoin(ctx, msg.To, mintedCoin)
	if err != nil {
		return err.Result()
	}
	flows = append(flows, flow)

	receipt := keeper.rk.NewReceipt(sdk.CategoryTypeHrc10, flows)
	result := sdk.Result{}
	keeper.rk.SaveReceiptToResult(receipt, &result)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewToken,
			sdk.NewAttribute(types.AttributeKeyName, msg.Name),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.From.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, msg.To.String()),
			sdk.NewAttribute(types.AttributeKeySymbol, ti.Symbol.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, totalSupply.String()),
			sdk.NewAttribute(types.AttributeKeyIssueFee, issueFee.String()),
		),
	)
	result.Events = append(result.Events, ctx.EventManager().Events()...)
	return result
}

func calSymbol(ctx sdk.Context) sdk.Symbol {
	hasherSHA256 := sha256.New()
	hasherSHA256.Write(ctx.TxBytes())
	sha := hasherSHA256.Sum(nil)

	hasherRIPEMD160 := ripemd160.New()
	hasherRIPEMD160.Write(sha)
	bz := hasherRIPEMD160.Sum(nil)

	sum := base58.Checksum(bz)
	bz = append(bz, sum[:]...)

	symbol := strings.ToUpper(sdk.NativeToken) + base58.Encode(bz)
	return sdk.Symbol(symbol)
}
