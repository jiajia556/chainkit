package deposit

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitdepositeventinbox"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddress"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/jiajia556/chainkit/service"
	"github.com/shopspring/decimal"
)

type inboxContextKey uint8

const (
	inboxSourceKey inboxContextKey = iota
	minDepositAmountKey
)

func WithInboxSource(ctx context.Context, source uint8) context.Context {
	return context.WithValue(ctx, inboxSourceKey, source)
}

func WithMinDepositAmount(ctx context.Context, amount decimal.Decimal) context.Context {
	return context.WithValue(ctx, minDepositAmountKey, amount)
}

func minDepositAmountFromContext(ctx context.Context) decimal.Decimal {
	if amount, ok := ctx.Value(minDepositAmountKey).(decimal.Decimal); ok {
		return amount
	}
	// Keep compatibility with callers outside this package during migration.
	if amount, ok := ctx.Value("minDepositAmount").(decimal.Decimal); ok {
		return amount
	}
	return decimal.Zero
}

func inboxSourceFromContext(ctx context.Context) uint8 {
	if source, ok := ctx.Value(inboxSourceKey).(uint8); ok {
		return source
	}
	return chainkitdepositeventinbox.SourceGetLogs
}

// EnqueueDeposit parses and filters a Transfer log, then durably merges it
// into the deposit inbox. Returning nil for an irrelevant event is intentional:
// only transfers to known deposit addresses belong in this queue.
func EnqueueDeposit(logCtx *service.LogContext, eventLog types.Log) error {
	if logCtx == nil || logCtx.Session == nil {
		return errors.New("deposit log context is nil")
	}

	erc20ABI, err := erc20.NewErc20(common.Address{}, nil)
	if err != nil {
		return err
	}
	transfer, err := erc20ABI.ParseTransfer(eventLog)
	if err != nil {
		return nil
	}

	amount := decimal.NewFromBigInt(transfer.Value, 0)
	if amount.IsZero() || amount.LessThan(minDepositAmountFromContext(logCtx.Ctx)) {
		return nil
	}

	userDepositAddress := chainkituserdepositaddress.NewRecord().ReadByAddress(transfer.To.Hex())
	if !userDepositAddress.Exists() {
		return nil
	}
	token := chainkittokens.NewRecord(logCtx.Session).
		ReadByChainAndContractAddress(logCtx.ChainDbId, eventLog.Address.Hex())
	if !token.Exists() {
		return nil
	}

	record := chainkitdepositeventinbox.NewRecord(logCtx.Session)
	record.Model.ChainDbId = logCtx.ChainDbId
	record.Model.ContractAddress = eventLog.Address.Hex()
	record.Model.TxHash = eventLog.TxHash.Hex()
	record.Model.LogIndex = uint32(eventLog.Index)
	record.Model.BlockNumber = eventLog.BlockNumber
	record.Model.BlockHash = eventLog.BlockHash.Hex()
	record.Model.FromAddress = transfer.From.Hex()
	record.Model.ToAddress = transfer.To.Hex()
	record.Model.Amount = amount
	record.Model.Source = inboxSourceFromContext(logCtx.Ctx)
	record.Model.Status = chainkitdepositeventinbox.StatusPending
	if err := record.Upsert(); err != nil {
		return err
	}
	if !eventLog.Removed {
		return nil
	}
	return chainkitdepositeventinbox.MarkRemoved(
		logCtx.Session,
		logCtx.ChainDbId,
		eventLog.TxHash.Hex(),
		uint32(eventLog.Index),
		"event removed by websocket subscription",
		time.Now(),
	)
}
