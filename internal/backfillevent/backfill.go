package backfillevent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitasset"
	"github.com/jiajia556/chainkit/models/chainkitdepositrecord"
	"github.com/jiajia556/chainkit/models/chainkitdeposittokens"
	"github.com/jiajia556/chainkit/models/chainkiteventbackfilltask"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddress"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddressassetbalance"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/jiajia556/chainkit/service"
	"github.com/jiajia556/tool-box/log"
	"github.com/shopspring/decimal"
)

const (
	moduleDeposit = "deposit"
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
	if task.Module != moduleDeposit {
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
		if err := cs.ScanBlockRange(scanCtx, task.ContractAddress, task.Module, fromBlock, toBlock, handleDeposit, func(logCtx *service.LogContext) error {
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

func handleDeposit(logCtx *service.LogContext, eventLog types.Log) error {
	log.Debug("handling backfill deposit", "log count", len(eventLog.Data), "chainDbId", logCtx.ChainDbId)
	erc20Abi, _ := erc20.NewErc20(common.Address{}, nil)
	transfer, err := erc20Abi.ParseTransfer(eventLog)
	if err != nil {
		return nil
	}

	to := transfer.To.Hex()
	from := transfer.From.Hex()
	amount := decimal.NewFromBigInt(transfer.Value, 0)
	minDepositAmount, ok := logCtx.Ctx.Value("minDepositAmount").(decimal.Decimal)
	if !ok {
		minDepositAmount = decimal.Zero
	}
	if amount.LessThan(minDepositAmount) {
		return nil
	}
	hash := eventLog.TxHash.Hex()

	if amount.Equal(decimal.Zero) {
		return nil
	}

	userDepAddr := chainkituserdepositaddress.NewRecord().ReadByAddress(to)
	if !userDepAddr.Exists() {
		return nil
	}
	token := chainkittokens.NewRecord(logCtx.Session).ReadByChainAndContractAddress(logCtx.ChainDbId, eventLog.Address.Hex())
	if !token.Exists() {
		return nil
	}
	eventLogId, created, err := logCtx.SaveEventLog(eventLog)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}

	userDepAddrBalance := chainkituserdepositaddressassetbalance.NewRecord(logCtx.Session).
		ReadByChainAndAddressAndToken(logCtx.ChainDbId, userDepAddr.Model.Id, token.Model.Id)
	if !userDepAddrBalance.Exists() {
		userDepAddrBalance.Model.UserId = userDepAddr.Model.UserId
		userDepAddrBalance.Model.ChainDbId = logCtx.ChainDbId
		userDepAddrBalance.Model.TokenId = token.Model.Id
		userDepAddrBalance.Model.UserDepositAddressId = userDepAddr.Model.Id
		userDepAddrBalance.Model.Address = userDepAddr.Model.Address
		userDepAddrBalance.Model.ConfirmedInAmount = decimal.Zero
		userDepAddrBalance.Model.BalanceAmount = decimal.Zero
		userDepAddrBalance.Model.LastInTxHash = ""
		err = userDepAddrBalance.Create()
		if err != nil {
			userDepAddrBalance = chainkituserdepositaddressassetbalance.NewRecord(logCtx.Session).
				ReadByChainAndAddressAndToken(logCtx.ChainDbId, userDepAddr.Model.Id, token.Model.Id)
		}
	}
	err = userDepAddrBalance.Deposit(amount, hash)
	if err != nil {
		log.Debug("failed to update user deposit address asset balance", "chainDbId", logCtx.ChainDbId, "userDepositAddressId", userDepAddr.Model.Id, "tokenId", token.Model.Id, "amount", amount.String(), "hash", hash, "error", err)
		return err
	}

	amountDecimal := amount.Div(decimal.New(1, int32(token.Model.Decimals)))

	depRecord := chainkitdepositrecord.NewRecord(logCtx.Session)
	depRecord.Model.UserId = userDepAddr.Model.UserId
	depRecord.Model.ChainDbId = logCtx.ChainDbId
	depRecord.Model.TokenId = token.Model.Id
	depRecord.Model.EventLogId = eventLogId
	depRecord.Model.UserDepositAddressId = userDepAddr.Model.Id
	depRecord.Model.FromAddress = from
	depRecord.Model.ToAddress = to
	depRecord.Model.Amount = amount
	depRecord.Model.AmountDecimal = amountDecimal
	depRecord.Model.Remark = hash
	_ = depRecord.Create()

	if userDepAddr.Model.UserId == 0 {
		return nil
	}

	userAsset := chainkitasset.NewRecord(logCtx.Session).GetByUserAndTokenGroup(userDepAddr.Model.UserId, token.Model.TokenGroupId)
	err = userAsset.IncreaseBalance(amount, "deposit", depRecord.Model.Id, fmt.Sprintf("deposit_%d", depRecord.Model.Id), "")
	if err != nil {
		log.Debug("failed to increase user asset balance", "error", err)
		return err
	}

	return nil
}

func markTaskFailed(taskId uint64, taskErr error) {
	taskRecord := chainkiteventbackfilltask.NewRecord()
	taskRecord.Model.Id = taskId
	err := taskRecord.SetFailed(taskErr)
	if err != nil {
		log.Error("failed to mark backfill task failed", "taskId", taskId, "error", err)
	}
}
