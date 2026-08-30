package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitscancursor"
	"github.com/jiajia556/tool-box/log"
	"github.com/jiajia556/tool-box/mysqlx"
)

type scanOptions struct {
	// Update Clone when adding reference-type fields such as slices, maps, or pointers.
	safeConfirmations uint64
	step              uint64
	startBlock        uint64
	budget            time.Duration
	Topics            [][]common.Hash
}

// Clone returns a deep-copied value of scanOptions.
// Note: Topics is a 2D slice and must be deep-copied to avoid sharing backing arrays
// between callers (e.g., when cloning from defaultScanOptions).
func (o *scanOptions) Clone() scanOptions {
	if o == nil {
		return scanOptions{}
	}

	out := *o // copy primitive fields + slice headers
	if o.Topics != nil {
		out.Topics = make([][]common.Hash, len(o.Topics))
		for i := range o.Topics {
			if o.Topics[i] == nil {
				continue
			}
			inner := make([]common.Hash, len(o.Topics[i]))
			copy(inner, o.Topics[i])
			out.Topics[i] = inner
		}
	}

	return out
}

var defaultScanOptions = &scanOptions{
	safeConfirmations: 20,
	step:              1000,
	budget:            20 * time.Second,
}

var suggestedBlockRangePattern = regexp.MustCompile(`(?i)\[\s*(0x[0-9a-f]+)\s*,\s*(0x[0-9a-f]+)\s*\]`)

const ModuleDeposit = "deposit"

type ScanOption func(*scanOptions)
type LogHandler func(ctx *LogContext, log types.Log) error
type LogContextHandler func(ctx *LogContext) error

// HandleLog runs one log through the same transaction boundary used by block
// scanning. It is used by durable queues that process previously collected logs.
func (s *ChainService) HandleLog(ctx context.Context, contractAddress, module string, handler LogHandler, eventLog types.Log) error {
	if s == nil || s.rpcClient == nil {
		return errors.New("chain service not initialized")
	}
	if handler == nil {
		return errors.New("eventLog handler is nil")
	}
	if !common.IsHexAddress(contractAddress) {
		return errors.New("invalid contract address")
	}
	return s.handleLogInTx(ctx, contractAddress, module, handler, eventLog)
}

func SafeConfirmations(confirmations uint64) ScanOption {
	return func(o *scanOptions) {
		o.safeConfirmations = confirmations
	}
}

func Step(step uint64) ScanOption {
	return func(o *scanOptions) {
		if step > 0 {
			o.step = step
		}
	}
}

func Topics(topics [][]common.Hash) ScanOption {
	return func(o *scanOptions) {
		o.Topics = topics
	}
}

func StartBlock(startBlock uint64) ScanOption {
	return func(o *scanOptions) {
		o.startBlock = startBlock
	}
}

// ScanBudget limits how long one ScanBlock call keeps catching up. A zero
// budget disables the limit and scans until the safe chain head is reached.
func ScanBudget(budget time.Duration) ScanOption {
	return func(o *scanOptions) {
		o.budget = budget
	}
}

// HeaderByNumber performs a rate-limited header lookup. Deposit uses this for
// confirmation processing so it shares the same per-chain limiter as getLogs.
func (s *ChainService) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if s == nil || s.rpcClient == nil {
		return nil, errors.New("chain service not initialized")
	}
	if err := s.waitForRPCRequest(ctx); err != nil {
		return nil, err
	}
	return s.rpcClient.HeaderByNumber(ctx, number)
}

func (s *ChainService) ScanBlock(ctx context.Context, contractAddress, module string, handler LogHandler, option ...ScanOption) error {
	//log.Debug("starting scan block", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
	if s.rpcClient == nil {
		return errors.New("chain service not initialized")
	}
	if handler == nil {
		return errors.New("eventLog handler is nil")
	}
	if !common.IsHexAddress(contractAddress) {
		return errors.New("invalid contract address")
	}

	opts := defaultScanOptions.Clone()
	if s.safeConfirmations > 0 {
		opts.safeConfirmations = s.safeConfirmations
	}
	for _, apply := range option {
		apply(&opts)
	}
	if opts.step == 0 {
		return errors.New("scan step must be greater than zero")
	}

	startedAt := time.Now()
	currentStep := opts.step
	for {
		advanced, caughtUp, err := s.scanBlockOnce(ctx, contractAddress, module, handler, &opts, currentStep)
		if err != nil {
			if currentStep > 1 && isFilterLogsRangeLimit(err) {
				previousStep := currentStep
				providerSuggested := false
				if suggestedStep, ok := suggestedFilterLogsStep(err); ok && suggestedStep < currentStep {
					currentStep = suggestedStep
					providerSuggested = true
				} else {
					currentStep /= 2
					if currentStep == 0 {
						currentStep = 1
					}
				}
				log.Info("reducing getLogs block range", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "previousStep", previousStep, "nextStep", currentStep, "providerSuggested", providerSuggested, "error", err)
				continue
			}
			return err
		}
		if caughtUp || !advanced {
			return nil
		}
		if currentStep < opts.step {
			currentStep *= 2
			if currentStep > opts.step {
				currentStep = opts.step
			}
		}
		if opts.budget > 0 && time.Since(startedAt) >= opts.budget {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

func (s *ChainService) scanBlockOnce(ctx context.Context, contractAddress, module string, handler LogHandler, opts *scanOptions, step uint64) (bool, bool, error) {

	//log.Debug("retrieved header", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
	header, err := s.HeaderByNumber(ctx, nil)
	if err != nil {
		return false, false, err
	}
	if header.Number.Uint64() <= opts.safeConfirmations {
		return false, true, nil
	}
	netSafeLastestBlock := header.Number.Uint64() - opts.safeConfirmations

	session := mysqlx.NewTxSession()
	if err := session.Begin(); err != nil {
		return false, false, err
	}
	defer func() {
		e := recover()
		if e != nil {
			_ = session.Rollback()
			panic(e)
		}
	}()

	//log.Debug("rolling back transaction", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
	rollbackWithErr := func(err error) error {
		if rbErr := session.Rollback(); rbErr != nil {
			return errors.New(err.Error() + "; rollback failed: " + rbErr.Error())
		}
		return err
	}

	//log.Debug("reading cursor", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
	cursor := chainkitscancursor.NewRecord(session)
	if err := cursor.ReadByContractAndChainForUpdate(contractAddress, module, s.chainDbId); err != nil && err.Error() != "record not found" {
		return false, false, rollbackWithErr(err)
	}
	if !cursor.Exists() {
		cursor.Model.ChainDbId = s.chainDbId
		cursor.Model.ContractAddress = contractAddress
		cursor.Model.Module = module
		cursor.Model.LastestBlock = opts.startBlock
		if err := cursor.Create(); err != nil {
			return false, false, rollbackWithErr(err)
		}
	}
	if cursor.Model.LastestBlock == 0 {
		return false, false, rollbackWithErr(errors.New("start block is not set"))
	}

	//log.Debug("to block", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
	fromBlock := cursor.Model.LastestBlock + 1
	toBlock := fromBlock + step - 1
	if toBlock < fromBlock { // uint64 overflow
		toBlock = netSafeLastestBlock
	}
	if toBlock > netSafeLastestBlock {
		toBlock = netSafeLastestBlock
	}

	//log.Debug("block", "fromBlock", fromBlock, "toBlock", toBlock)
	if fromBlock > toBlock {
		return false, true, session.Commit()
	}

	//log.Debug("scanning block range", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module, "fromBlock", fromBlock, "toBlock", toBlock)
	if err := s.scanBlockRangeLogs(ctx, contractAddress, module, fromBlock, toBlock, opts.Topics, handler); err != nil {
		return false, false, rollbackWithErr(err)
	}

	if toBlock > cursor.Model.LastestBlock {
		if err := cursor.UpdateLastestBlock(toBlock); err != nil {
			return false, false, rollbackWithErr(err)
		}
	}

	if err := session.Commit(); err != nil {
		return false, false, err
	}
	return true, toBlock >= netSafeLastestBlock, nil
}

func (s *ChainService) ScanBlockRange(ctx context.Context, contractAddress, module string, fromBlock, toBlock uint64, handler LogHandler, afterHandlers ...LogContextHandler) error {
	return s.scanBlockRange(ctx, contractAddress, module, fromBlock, toBlock, nil, handler, afterHandlers...)
}

func (s *ChainService) scanBlockRange(ctx context.Context, contractAddress, module string, fromBlock, toBlock uint64, topics [][]common.Hash, handler LogHandler, afterHandlers ...LogContextHandler) error {
	//log.Debug("starting scan block range", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module, "fromBlock", fromBlock, "toBlock", toBlock)
	if s.rpcClient == nil {
		return errors.New("chain service not initialized")
	}
	if handler == nil {
		return errors.New("eventLog handler is nil")
	}
	if !common.IsHexAddress(contractAddress) {
		return errors.New("invalid contract address")
	}
	if fromBlock > toBlock {
		return errors.New("from block is greater than to block")
	}

	if err := s.scanBlockRangeLogs(ctx, contractAddress, module, fromBlock, toBlock, topics, handler); err != nil {
		return err
	}

	if len(afterHandlers) == 0 {
		return nil
	}

	return s.withLogContextTx(ctx, contractAddress, module, func(logCtx *LogContext) error {
		for _, afterHandler := range afterHandlers {
			if afterHandler == nil {
				continue
			}
			if err := afterHandler(logCtx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ChainService) scanBlockRangeLogs(ctx context.Context, contractAddress, module string, fromBlock, toBlock uint64, topics [][]common.Hash, handler LogHandler) error {
	contract := common.HexToAddress(contractAddress)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Topics:    topics,
		Addresses: []common.Address{contract},
	}

	if err := s.waitForRPCRequest(ctx); err != nil {
		return err
	}
	logs, err := s.rpcClient.FilterLogs(ctx, query)
	if err != nil {
		return &filterLogsQueryError{fromBlock: fromBlock, toBlock: toBlock, err: err}
	}

	for _, eventLog := range logs {
		if len(eventLog.Topics) == 0 {
			continue
		}
		if err := s.handleLogInTx(ctx, contractAddress, module, handler, eventLog); err != nil {
			log.Error("failed to handle event log", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module, "txHash", eventLog.TxHash.Hex(), "logIndex", eventLog.Index, "blockNumber", eventLog.BlockNumber, "error", err)
			return err
		}
	}

	return nil
}

type filterLogsQueryError struct {
	fromBlock uint64
	toBlock   uint64
	err       error
}

func (e *filterLogsQueryError) Error() string {
	return fmt.Sprintf("eth_getLogs failed for blocks [%d,%d]: %v", e.fromBlock, e.toBlock, e.err)
}

func (e *filterLogsQueryError) Unwrap() error {
	return e.err
}

func isFilterLogsRangeLimit(err error) bool {
	var queryErr *filterLogsQueryError
	if !errors.As(err, &queryErr) {
		return false
	}
	message := strings.ToLower(queryErr.err.Error())
	patterns := []string{
		"too many results",
		"query returned more than",
		"response size exceeded",
		"response too large",
		"result set too large",
		"log response size exceeded",
		"block range is too wide",
		"block range too large",
		"limit exceeded",
		"please limit the query",
	}
	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// suggestedFilterLogsStep extracts a provider-recommended inclusive block
// range such as "[0x6E57751, 0x6E57785]". The recommendation is accepted only
// when it starts at the attempted fromBlock and strictly reduces that request,
// so a malformed provider response can never skip blocks.
func suggestedFilterLogsStep(err error) (uint64, bool) {
	var queryErr *filterLogsQueryError
	if !errors.As(err, &queryErr) {
		return 0, false
	}
	match := suggestedBlockRangePattern.FindStringSubmatch(queryErr.err.Error())
	if len(match) != 3 {
		return 0, false
	}
	fromBlock, fromErr := strconv.ParseUint(match[1][2:], 16, 64)
	toBlock, toErr := strconv.ParseUint(match[2][2:], 16, 64)
	if fromErr != nil || toErr != nil || fromBlock != queryErr.fromBlock || toBlock < fromBlock || toBlock >= queryErr.toBlock {
		return 0, false
	}
	return toBlock - fromBlock + 1, true
}

func (s *ChainService) handleLogInTx(ctx context.Context, contractAddress, module string, handler LogHandler, eventLog types.Log) error {
	return s.withLogContextTx(ctx, contractAddress, module, func(logCtx *LogContext) error {
		return handler(logCtx, eventLog)
	})
}

func (s *ChainService) withLogContextTx(ctx context.Context, contractAddress, module string, fn LogContextHandler) error {
	session := mysqlx.NewTxSession()
	if err := session.Begin(); err != nil {
		return err
	}
	defer func() {
		e := recover()
		if e != nil {
			_ = session.Rollback()
			panic(e)
		}
	}()

	rollbackWithErr := func(err error) error {
		if rbErr := session.Rollback(); rbErr != nil {
			return errors.New(err.Error() + "; rollback failed: " + rbErr.Error())
		}
		return err
	}

	logCtx := &LogContext{
		Ctx:             ctx,
		Session:         session,
		ChainDbId:       s.chainDbId,
		ContractAddress: contractAddress,
		Module:          module,
	}
	if fn != nil {
		if err := fn(logCtx); err != nil {
			return rollbackWithErr(err)
		}
	}

	if err := session.Commit(); err != nil {
		return err
	}
	return nil
}
