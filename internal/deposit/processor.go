package deposit

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitdepositeventinbox"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
)

const (
	processingLease           = 10 * time.Minute
	processingRecoveryPeriod  = time.Minute
	inboxErrorDelay           = 5 * time.Second
	depositTokenCheckInterval = 30 * time.Second
	maxRetryDelay             = 10 * time.Minute
)

var (
	InboxProcessLimit = 200
	InboxIdleInterval = time.Second
)

var transferEventSignature = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

// StartInboxProcessor runs one durable consumer per chain. Full batches are
// drained immediately; an empty or partially filled queue uses the idle delay.
func StartInboxProcessor(ctx context.Context) {
	chains := chainkitchains.NewList()
	if err := chains.FindAll(); err != nil {
		log.Error("failed to load chains for deposit inbox processor", "error", err)
		return
	}

	var wg sync.WaitGroup
	chains.Foreach(func(_ int, chain *chainkitchains.Record) bool {
		wg.Add(1)
		go func(chainDbID, safeConfirmations uint64) {
			defer wg.Done()
			runChainInboxProcessor(ctx, chainDbID, safeConfirmations)
		}(chain.Model.Id, chain.Model.SafeConfirmations)
		return true
	})
	wg.Wait()
}

func runChainInboxProcessor(ctx context.Context, chainDbID, safeConfirmations uint64) {
	for ctx.Err() == nil {
		hasTokens, err := hasAvailableDepositTokens(chainDbID)
		if err != nil {
			log.Error("failed to check enabled deposit tokens", "chainDbId", chainDbID, "error", err)
			if !waitContext(ctx, inboxErrorDelay) {
				return
			}
			continue
		}
		if !hasTokens {
			if !waitContext(ctx, depositTokenCheckInterval) {
				return
			}
			continue
		}

		chainService, err := service.NewChainService(chainDbID)
		if err != nil {
			log.Error("failed to create chain service for deposit inbox", "chainDbId", chainDbID, "error", err)
			if !waitContext(ctx, inboxErrorDelay) {
				return
			}
			continue
		}
		runConnectedInboxProcessor(ctx, chainService, chainDbID, safeConfirmations)
		chainService.CloseClient()
	}
}

func runConnectedInboxProcessor(ctx context.Context, chainService *service.ChainService, chainDbID, safeConfirmations uint64) {
	nextLeaseRecovery := time.Time{}
	nextTokenCheck := time.Now().Add(depositTokenCheckInterval)
	for ctx.Err() == nil {
		now := time.Now()
		if !now.Before(nextTokenCheck) {
			hasTokens, err := hasAvailableDepositTokens(chainDbID)
			if err != nil {
				log.Error("failed to recheck enabled deposit tokens", "chainDbId", chainDbID, "error", err)
			} else if !hasTokens {
				log.Info("closing deposit inbox RPC connection because no deposit tokens are enabled", "chainDbId", chainDbID)
				return
			}
			nextTokenCheck = now.Add(depositTokenCheckInterval)
		}
		if !now.Before(nextLeaseRecovery) {
			inboxList := chainkitdepositeventinbox.NewList()
			if count, err := inboxList.RequeueStaleProcessing(chainDbID, now.Add(-processingLease), now); err != nil {
				log.Error("failed to recover stale deposit inbox events", "chainDbId", chainDbID, "error", err)
			} else if count > 0 {
				log.Info("recovered stale deposit inbox events", "chainDbId", chainDbID, "count", count)
			}
			nextLeaseRecovery = now.Add(processingRecoveryPeriod)
		}

		processed, err := processInboxBatch(ctx, chainService, chainDbID, safeConfirmations)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Error("failed to process deposit inbox batch", "chainDbId", chainDbID, "error", err)
			if !waitContext(ctx, inboxErrorDelay) {
				return
			}
			continue
		}
		if processed < effectiveInboxProcessLimit() && !waitContext(ctx, effectiveInboxIdleInterval()) {
			return
		}
	}
}

func hasAvailableDepositTokens(chainDbID uint64) (bool, error) {
	return chainkitdeposittokens.NewList().HasAvailableByChainDBID(chainDbID)
}

func processInboxBatch(ctx context.Context, chainService *service.ChainService, chainDbID, safeConfirmations uint64) (int, error) {
	if chainService == nil || chainService.GetClient() == nil {
		return 0, errors.New("chain service is not initialized")
	}

	header, err := chainService.GetClient().HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	latestBlock := header.Number.Uint64()
	if latestBlock <= safeConfirmations {
		return 0, nil
	}
	safeBlock := latestBlock - safeConfirmations
	now := time.Now()

	inboxList := chainkitdepositeventinbox.NewList()
	if err := inboxList.FindProcessable(chainDbID, safeBlock, now, effectiveInboxProcessLimit()); err != nil {
		return 0, err
	}

	processed := 0
	for _, inbox := range *inboxList.Records {
		select {
		case <-ctx.Done():
			return processed, ctx.Err()
		default:
		}

		record := chainkitdepositeventinbox.NewRecord()
		record.Model = inbox
		claimed, err := record.Claim(time.Now())
		if err != nil {
			log.Error("failed to claim deposit inbox event", "chainDbId", chainDbID, "inboxId", inbox.Id, "error", err)
			continue
		}
		if !claimed {
			continue
		}
		processed++

		if err := processInboxRecord(ctx, chainService, record); err != nil {
			nextRetry := time.Now().Add(retryDelay(record.Model.RetryCount + 1))
			if retryErr := record.MarkRetry(err.Error(), nextRetry); retryErr != nil {
				log.Error("failed to mark deposit inbox event for retry", "chainDbId", chainDbID, "inboxId", inbox.Id, "processError", err, "retryError", retryErr)
			}
		}
	}
	return processed, nil
}

func effectiveInboxProcessLimit() int {
	if InboxProcessLimit <= 0 {
		return 200
	}
	return InboxProcessLimit
}

func effectiveInboxIdleInterval() time.Duration {
	if InboxIdleInterval <= 0 {
		return time.Second
	}
	return InboxIdleInterval
}

func processInboxRecord(ctx context.Context, chainService *service.ChainService, record *chainkitdepositeventinbox.Record) error {
	inbox := record.Model
	header, err := chainService.GetClient().HeaderByNumber(ctx, new(big.Int).SetUint64(inbox.BlockNumber))
	if err != nil {
		return err
	}
	if header.Hash() != common.HexToHash(inbox.BlockHash) {
		return record.MarkOrphaned("block hash no longer belongs to the canonical chain", time.Now())
	}
	if err := record.MarkConfirmed(time.Now()); err != nil {
		return err
	}

	eventLog, err := inboxEventLog(inbox)
	if err != nil {
		return record.MarkIgnored(err.Error(), time.Now())
	}
	if err := chainService.HandleLog(ctx, inbox.ContractAddress, service.ModuleDeposit, HandleDeposit, eventLog); err != nil {
		return err
	}
	return record.MarkCompleted(time.Now())
}

func inboxEventLog(inbox *chainkitdepositeventinbox.ChainDepositEventInbox) (types.Log, error) {
	if inbox == nil {
		return types.Log{}, errors.New("deposit inbox event is nil")
	}
	if !common.IsHexAddress(inbox.ContractAddress) || !common.IsHexAddress(inbox.FromAddress) || !common.IsHexAddress(inbox.ToAddress) {
		return types.Log{}, errors.New("deposit inbox contains an invalid address")
	}
	amount := inbox.Amount.BigInt()
	if amount.Sign() <= 0 {
		return types.Log{}, errors.New("deposit inbox amount must be positive")
	}
	return types.Log{
		Address: common.HexToAddress(inbox.ContractAddress),
		Topics: []common.Hash{
			transferEventSignature,
			common.BytesToHash(common.HexToAddress(inbox.FromAddress).Bytes()),
			common.BytesToHash(common.HexToAddress(inbox.ToAddress).Bytes()),
		},
		Data:        common.LeftPadBytes(amount.Bytes(), 32),
		BlockNumber: inbox.BlockNumber,
		TxHash:      common.HexToHash(inbox.TxHash),
		TxIndex:     0,
		BlockHash:   common.HexToHash(inbox.BlockHash),
		Index:       uint(inbox.LogIndex),
		Removed:     false,
	}, nil
}

func retryDelay(retryCount uint32) time.Duration {
	delay := 10 * time.Second
	for i := uint32(1); i < retryCount && delay < maxRetryDelay; i++ {
		delay *= 2
	}
	if delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}
