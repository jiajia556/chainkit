package eip7702batch

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatches"
	"github.com/jiajia556/chainkit/models/chainkitcollectbatchitems"
	"github.com/jiajia556/chainkit/models/chainkitcollecttasks"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddressassetbalance"
	"github.com/jiajia556/chainkit/pkg/contracts/batchsweepexecutor"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

const missingTxRebroadcastDelay = 2 * time.Minute

type outcomeKind uint8

const (
	outcomeSuccess outcomeKind = iota
	outcomeFailed
	outcomeNotAttempted
)

type itemOutcome struct {
	item           *chainkitcollectbatchitems.ChainCollectBatchItems
	task           *chainkitcollecttasks.ChainCollectTasks
	kind           outcomeKind
	resultCode     uint8
	actualAmount   decimal.Decimal
	chainBalance   decimal.Decimal
	returnDataHash string
	lastError      string
}

// Check advances one sent/maybe-sent batch. It is safe to call repeatedly.
func Check(ctx context.Context, batchId uint64) error {
	batch := chainkitcollectbatches.NewRecord()
	if err := batch.Read(batchId); err != nil {
		return fmt.Errorf("read collect batch: %w", err)
	}
	if batch.Model.Status != chainkitcollectbatches.StatusSent &&
		batch.Model.Status != chainkitcollectbatches.StatusMaybeSent {
		return nil
	}
	if !common.IsHexHash(batch.Model.TxHash) {
		return fmt.Errorf("batch %d has invalid transaction hash", batchId)
	}

	chainService, err := service.NewChainService(batch.Model.ChainDbId)
	if err != nil {
		return err
	}
	defer chainService.CloseClient()

	status, err := chainService.GetTxStatus(batch.Model.TxHash)
	if err != nil {
		return err
	}
	switch status {
	case service.TxStatusPending, service.TxStatusMined:
		return nil
	case service.TxStatusNotFound:
		return handleMissingTransaction(ctx, chainService, batch)
	case service.TxStatusUnknown:
		if batch.SinceSent() > 10*time.Minute {
			return batch.SetUnknown("transaction status remained unknown for more than 10 minutes")
		}
		return nil
	case service.TxStatusFailed:
		receipt, err := chainService.GetClient().TransactionReceipt(ctx, common.HexToHash(batch.Model.TxHash))
		if err != nil {
			return err
		}
		return finalizeRevertedBatch(batch, receipt)
	case service.TxStatusConfirmed:
		receipt, err := chainService.GetClient().TransactionReceipt(ctx, common.HexToHash(batch.Model.TxHash))
		if err != nil {
			return err
		}
		return finalizeConfirmedBatch(ctx, chainService, batch, receipt)
	default:
		return fmt.Errorf("unsupported transaction status %q", status)
	}
}

func handleMissingTransaction(
	ctx context.Context,
	chainService *service.ChainService,
	batch *chainkitcollectbatches.Record,
) error {
	if batch.SinceSent() < missingTxRebroadcastDelay {
		return nil
	}
	occupied, err := chainService.IsNonceOccupied(batch.Model.SponsorAddress, batch.Model.SponsorNonce)
	if err != nil {
		return err
	}
	if occupied {
		return batch.SetUnknown("transaction hash not found but sponsor nonce is occupied")
	}
	if len(batch.Model.RawTx) == 0 {
		return batch.SetUnknown("transaction hash not found and signed raw transaction is missing")
	}

	var signedTx types.Transaction
	if err := signedTx.UnmarshalBinary(batch.Model.RawTx); err != nil {
		return batch.SetUnknown("stored raw transaction cannot be decoded: " + err.Error())
	}
	if signedTx.Hash() != common.HexToHash(batch.Model.TxHash) || signedTx.Nonce() != batch.Model.SponsorNonce {
		return batch.SetUnknown("stored raw transaction does not match batch hash or nonce")
	}

	hash, fakeErr, sendErr := chainService.SendSignedTransaction(ctx, &signedTx)
	if sendErr != nil {
		_, updateErr := batch.SetMaybeSent(batch.Model.TxHash, batch.Model.RawTx, sendErr.Error())
		if updateErr != nil {
			return fmt.Errorf("rebroadcast failed: %v; persist error: %w", sendErr, updateErr)
		}
		return nil
	}
	if fakeErr != nil {
		_, err = batch.SetMaybeSent(hash, batch.Model.RawTx, fakeErr.Error())
		return err
	}
	_, err = batch.SetSent(hash, batch.Model.RawTx)
	if err != nil {
		return err
	}
	return chainkitcollecttasks.NewRecord().SetBatchSent(batch.Model.Id, hash)
}

func finalizeRevertedBatch(
	batch *chainkitcollectbatches.Record,
	receipt *types.Receipt,
) error {
	items := chainkitcollectbatchitems.NewList()
	items.GetByBatchId(batch.Model.Id)
	outcomes, err := buildNotAttemptedOutcomes(items, "outer transaction reverted")
	if err != nil {
		return err
	}
	return finalizeBatch(batch, receipt, outcomes, false)
}

func finalizeConfirmedBatch(
	ctx context.Context,
	chainService *service.ChainService,
	batch *chainkitcollectbatches.Record,
	receipt *types.Receipt,
) error {
	items := chainkitcollectbatchitems.NewList()
	items.GetByBatchId(batch.Model.Id)
	if items.IsEmpty() {
		return batch.SetUnknown("confirmed batch has no items")
	}

	contract, err := batchsweepexecutor.NewBatchSweepExecutor(
		common.HexToAddress(batch.Model.ExecutorAddress),
		chainService.GetClient(),
	)
	if err != nil {
		return err
	}
	contractABI, err := batchsweepexecutor.BatchSweepExecutorMetaData.GetAbi()
	if err != nil {
		return err
	}
	collectResultID := contractABI.Events["CollectResult"].ID
	batchStoppedID := contractABI.Events["BatchStopped"].ID

	itemByTask := make(map[uint64]*chainkitcollectbatchitems.ChainCollectBatchItems, len(*items.Records))
	for _, item := range *items.Records {
		itemByTask[item.CollectTaskId] = item
	}
	outcomes := make(map[uint64]*itemOutcome, len(itemByTask))
	stoppedAt := uint64(len(itemByTask))

	for _, receiptLog := range receipt.Logs {
		if receiptLog.Address != common.HexToAddress(batch.Model.ExecutorAddress) || len(receiptLog.Topics) == 0 {
			continue
		}
		switch receiptLog.Topics[0] {
		case collectResultID:
			event, err := contract.ParseCollectResult(*receiptLog)
			if err != nil {
				return fmt.Errorf("parse collect result: %w", err)
			}
			if !event.TaskId.IsUint64() {
				return batch.SetUnknown("collect result task id overflows uint64")
			}
			taskId := event.TaskId.Uint64()
			item, exists := itemByTask[taskId]
			if !exists {
				return batch.SetUnknown(fmt.Sprintf("collect result contains unknown task %d", taskId))
			}
			if _, duplicate := outcomes[taskId]; duplicate {
				return batch.SetUnknown(fmt.Sprintf("collect result duplicated task %d", taskId))
			}
			if event.Account != common.HexToAddress(item.AuthorityAddress) ||
				event.Token != common.HexToAddress(item.TokenAddress) ||
				event.Recipient != common.HexToAddress(item.RecipientAddress) ||
				event.RequestedAmount.String() != item.PlannedAmount {
				return batch.SetUnknown(fmt.Sprintf("collect result does not match batch item for task %d", taskId))
			}
			outcome := &itemOutcome{
				item:           item,
				resultCode:     event.Result,
				actualAmount:   decimal.NewFromBigInt(event.RequestedAmount, 0),
				returnDataHash: common.BytesToHash(event.ReturnDataHash[:]).Hex(),
			}
			if event.Result == 0 {
				outcome.kind = outcomeSuccess
			} else {
				outcome.kind = outcomeFailed
				outcome.lastError = collectResultError(event.Result)
			}
			outcomes[taskId] = outcome
		case batchStoppedID:
			event, err := contract.ParseBatchStopped(*receiptLog)
			if err != nil {
				return fmt.Errorf("parse batch stopped: %w", err)
			}
			if !event.NextItemIndex.IsUint64() {
				return batch.SetUnknown("batch stopped index overflows uint64")
			}
			stoppedAt = event.NextItemIndex.Uint64()
		}
	}

	orderedOutcomes := make([]*itemOutcome, 0, len(itemByTask))
	for _, item := range *items.Records {
		outcome, exists := outcomes[item.CollectTaskId]
		if !exists {
			if uint64(item.ItemIndex) < stoppedAt {
				return batch.SetUnknown(fmt.Sprintf("confirmed batch is missing result for task %d", item.CollectTaskId))
			}
			outcome = &itemOutcome{
				item:      item,
				kind:      outcomeNotAttempted,
				lastError: "batch stopped before this item was attempted",
			}
		}
		task := chainkitcollecttasks.NewRecord()
		if err := task.Read(item.CollectTaskId); err != nil {
			return err
		}
		if task.Model.BatchId != batch.Model.Id {
			return batch.SetUnknown(fmt.Sprintf("task %d no longer belongs to batch", item.CollectTaskId))
		}
		outcome.task = task.Model
		if outcome.kind == outcomeSuccess {
			if task.Model.TokenId == 0 {
				outcome.chainBalance, err = chainService.BalanceAt(task.Model.FromAddress)
			} else {
				outcome.chainBalance, err = chainService.BalanceOf(item.TokenAddress, task.Model.FromAddress)
			}
			if err != nil {
				return fmt.Errorf("read post-collect balance for task %d: %w", item.CollectTaskId, err)
			}
		}
		orderedOutcomes = append(orderedOutcomes, outcome)
	}

	return finalizeBatch(batch, receipt, orderedOutcomes, true)
}

func buildNotAttemptedOutcomes(
	items *chainkitcollectbatchitems.List,
	reason string,
) ([]*itemOutcome, error) {
	outcomes := make([]*itemOutcome, 0, len(*items.Records))
	for _, item := range *items.Records {
		task := chainkitcollecttasks.NewRecord()
		if err := task.Read(item.CollectTaskId); err != nil {
			return nil, fmt.Errorf("read collect task %d: %w", item.CollectTaskId, err)
		}
		if task.Model.BatchId != item.BatchId {
			return nil, fmt.Errorf("collect task %d no longer belongs to batch %d", item.CollectTaskId, item.BatchId)
		}
		outcomes = append(outcomes, &itemOutcome{
			item:      item,
			task:      task.Model,
			kind:      outcomeNotAttempted,
			lastError: reason,
		})
	}
	return outcomes, nil
}

func finalizeBatch(
	batch *chainkitcollectbatches.Record,
	receipt *types.Receipt,
	outcomes []*itemOutcome,
	outerSuccess bool,
) error {
	if receipt == nil || receipt.EffectiveGasPrice == nil {
		return errors.New("transaction receipt is missing gas information")
	}
	txFee := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)

	session := mysqlx.NewTxSession()
	if err := session.Begin(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = session.Rollback()
		}
	}()

	for _, outcome := range outcomes {
		item := chainkitcollectbatchitems.NewRecord(session)
		item.Model = outcome.item
		task := chainkitcollecttasks.NewRecord(session)
		task.Model = outcome.task

		switch outcome.kind {
		case outcomeSuccess:
			if err := item.SetResult(
				chainkitcollectbatchitems.StatusSuccess,
				outcome.resultCode,
				outcome.actualAmount.String(),
				outcome.returnDataHash,
				"",
			); err != nil {
				return err
			}
			if err := task.DB().Model(task.Model).Updates(map[string]interface{}{
				"status":        chainkitcollecttasks.StatusConfirmed,
				"actual_amount": outcome.actualAmount,
				"tx_hash":       batch.Model.TxHash,
				"confirmed_at":  time.Now(),
				"last_error":    "",
			}).Error; err != nil {
				return err
			}
			balance := chainkituserdepositaddressassetbalance.NewRecord(session).
				ReadByChainAndAddressAndToken(task.Model.ChainDbId, task.Model.UserDepositAddressId, task.Model.TokenId)
			if !balance.Exists() {
				return fmt.Errorf("asset balance not found for task %d", task.Model.Id)
			}
			if err := balance.CollectedWithBalance(outcome.actualAmount, outcome.chainBalance, batch.Model.TxHash); err != nil {
				return err
			}
		case outcomeFailed:
			if err := item.SetResult(
				chainkitcollectbatchitems.StatusFailed,
				outcome.resultCode,
				"0",
				outcome.returnDataHash,
				outcome.lastError,
			); err != nil {
				return err
			}
			if err := task.DB().Model(task.Model).Updates(map[string]interface{}{
				"status":     chainkitcollecttasks.StatusFailed,
				"tx_hash":    batch.Model.TxHash,
				"last_error": outcome.lastError,
			}).Error; err != nil {
				return err
			}
		case outcomeNotAttempted:
			if err := item.MarkNotAttempted(outcome.lastError); err != nil {
				return err
			}
			if err := task.DB().Model(task.Model).Updates(map[string]interface{}{
				"status":         chainkitcollecttasks.StatusWaiting,
				"collect_method": chainkitcollecttasks.CollectMethodUndecided,
				"batch_id":       0,
				"tx_hash":        "",
				"sent_at":        nil,
				"last_error":     outcome.lastError,
			}).Error; err != nil {
				return err
			}
		}
	}

	batchRecord := chainkitcollectbatches.NewRecord(session)
	batchRecord.Model = batch.Model
	if outerSuccess {
		if err := batchRecord.SetConfirmed(receipt.GasUsed, txFee.String()); err != nil {
			return err
		}
	} else {
		if err := batchRecord.SetFailed("outer transaction reverted"); err != nil {
			return err
		}
	}
	if err := session.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func collectResultError(result uint8) string {
	switch result {
	case 1:
		return "invalid account"
	case 2:
		return "invalid token"
	case 3:
		return "recipient is not allowed"
	case 4:
		return "wrong delegation"
	case 5:
		return "delegate call failed"
	case 6:
		return "invalid delegate return value"
	case 7:
		return "insufficient execution gas"
	default:
		return fmt.Sprintf("unknown collect result code %d", result)
	}
}
