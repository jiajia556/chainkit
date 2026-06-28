package backfillevent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jiajia556/chainkit/internal/deposit"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkiteventbackfilltask"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

var (
	TaskLimit        = 10
	BlockStep uint64 = 1000
)

func Start(ctx context.Context) {
	log.Debug("backfill-event service started")
	tasks, err := findRunnableTasks(TaskLimit)
	if err != nil {
		log.Error("failed to find backfill tasks", "error", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, task := range tasks {
		wg.Add(1)
		go func(task *chainkiteventbackfilltask.ChainEventBackfillTask) {
			defer wg.Done()
			if err := handleTask(ctx, task); err != nil {
				log.Error("failed to handle backfill task", "taskId", task.Id, "error", err)
				markTaskFailed(task.Id, err)
			}
		}(task)
	}
	wg.Wait()
}

func findRunnableTasks(limit int) ([]*chainkiteventbackfilltask.ChainEventBackfillTask, error) {
	taskList := chainkiteventbackfilltask.NewList()
	err := taskList.FindRunnable(limit)
	if err != nil {
		return nil, err
	}
	return *taskList.Records, nil
}

func handleTask(ctx context.Context, task *chainkiteventbackfilltask.ChainEventBackfillTask) error {
	if task == nil {
		return errors.New("backfill task is nil")
	}
	if task.EndBlock < task.StartBlock {
		return errors.New("end block is less than start block")
	}
	if task.Module != service.ModuleDeposit {
		return fmt.Errorf("unsupported backfill module: %s", task.Module)
	}
	if !common.IsHexAddress(task.ContractAddress) {
		return errors.New("invalid contract address")
	}

	taskRecord := chainkiteventbackfilltask.NewRecord()
	taskRecord.Model.Id = task.Id
	if err := taskRecord.SetRunning(); err != nil {
		return err
	}

	cs, err := service.NewChainService(task.ChainDbId)
	if err != nil {
		return err
	}
	defer cs.CloseClient()

	minDepositAmount := readMinDepositAmount(task.ChainDbId, task.ContractAddress)
	nextBlock := task.CurrentBlock + 1
	if task.CurrentBlock < task.StartBlock {
		nextBlock = task.StartBlock
	}

	for nextBlock <= task.EndBlock {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fromBlock := nextBlock
		toBlock := task.EndBlock
		if BlockStep > 0 {
			toBlock = fromBlock + BlockStep - 1
		}
		if toBlock < fromBlock {
			toBlock = task.EndBlock
		}
		if toBlock > task.EndBlock {
			toBlock = task.EndBlock
		}

		scanCtx := context.WithValue(ctx, "minDepositAmount", minDepositAmount)
		if err := cs.ScanBlockRange(scanCtx, task.ContractAddress, task.Module, fromBlock, toBlock, deposit.HandleDeposit, func(logCtx *service.LogContext) error {
			taskRecord := chainkiteventbackfilltask.NewRecord(logCtx.Session)
			taskRecord.Model.Id = task.Id
			return taskRecord.SetCurrentBlock(toBlock)
		}); err != nil {
			return err
		}
		if toBlock == task.EndBlock {
			break
		}
		nextBlock = toBlock + 1
	}

	taskRecord = chainkiteventbackfilltask.NewRecord()
	taskRecord.Model.Id = task.Id
	return taskRecord.SetDone(task.EndBlock)
}

func readMinDepositAmount(chainDbId uint64, contractAddress string) decimal.Decimal {
	token := chainkittokens.NewRecord().ReadByChainAndContractAddress(chainDbId, contractAddress)
	if !token.Exists() {
		log.Error("failed to find token for backfill contract", "chainDbId", chainDbId, "contractAddress", contractAddress)
		return decimal.Zero
	}

	depositToken := chainkitdeposittokens.NewRecord()
	depositToken.ReadAvailableByChainAndToken(chainDbId, token.Model.Id)
	if !depositToken.Exists() {
		return decimal.Zero
	}
	return depositToken.Model.MinDepositAmount
}

func markTaskFailed(taskId uint64, taskErr error) {
	taskRecord := chainkiteventbackfilltask.NewRecord()
	taskRecord.Model.Id = taskId
	err := taskRecord.SetFailed(taskErr)
	if err != nil {
		log.Error("failed to mark backfill task failed", "taskId", taskId, "error", err)
	}
}
