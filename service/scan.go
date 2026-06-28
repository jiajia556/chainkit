package service

import (
	"context"
	"errors"
	"math/big"

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
}

const ModuleDeposit = "deposit"

type ScanOption func(*scanOptions)
type LogHandler func(ctx *LogContext, log types.Log) error
type LogContextHandler func(ctx *LogContext) error

func SafeConfirmations(confirmations uint64) ScanOption {
	return func(o *scanOptions) {
		o.safeConfirmations = confirmations
	}
}

func Step(step uint64) ScanOption {
	return func(o *scanOptions) {
		o.step = step
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

func (s *ChainService) ScanBlock(ctx context.Context, contractAddress, module string, handler LogHandler, option ...ScanOption) error {
	log.Debug("starting scan block", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
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
	for _, apply := range option {
		apply(&opts)
	}

	header, err := s.rpcClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	if header.Number.Uint64() <= opts.safeConfirmations {
		return errors.New("latest block number is less than or equal to safe confirmations")
	}
	netSafeLastestBlock := header.Number.Uint64() - opts.safeConfirmations

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

	cursor := chainkitscancursor.NewRecord(session)
	if err := cursor.ReadByContractAndChainForUpdate(contractAddress, module, s.chainDbId); err != nil && err.Error() != "record not found" {
		return rollbackWithErr(err)
	}
	if !cursor.Exists() {
		cursor.Model.ChainDbId = s.chainDbId
		cursor.Model.ContractAddress = contractAddress
		cursor.Model.Module = module
		cursor.Model.LastestBlock = opts.startBlock
		if err := cursor.Create(); err != nil {
			return rollbackWithErr(err)
		}
	}
	if cursor.Model.LastestBlock == 0 {
		return rollbackWithErr(errors.New("start block is not set"))
	}

	toBlock := cursor.Model.LastestBlock + opts.step
	if toBlock > netSafeLastestBlock {
		toBlock = netSafeLastestBlock
	}
	fromBlock := cursor.Model.LastestBlock + 1

	if fromBlock > toBlock {
		return session.Commit()
	}

	if err := s.scanBlockRangeLogs(ctx, contractAddress, module, fromBlock, toBlock, opts.Topics, handler); err != nil {
		return rollbackWithErr(err)
	}

	if toBlock > cursor.Model.LastestBlock {
		if err := cursor.UpdateLastestBlock(toBlock); err != nil {
			return rollbackWithErr(err)
		}
	}

	return session.Commit()
}

func (s *ChainService) ScanBlockRange(ctx context.Context, contractAddress, module string, fromBlock, toBlock uint64, handler LogHandler, afterHandlers ...LogContextHandler) error {
	return s.scanBlockRange(ctx, contractAddress, module, fromBlock, toBlock, nil, handler, afterHandlers...)
}

func (s *ChainService) scanBlockRange(ctx context.Context, contractAddress, module string, fromBlock, toBlock uint64, topics [][]common.Hash, handler LogHandler, afterHandlers ...LogContextHandler) error {
	log.Debug("starting scan block range", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module, "fromBlock", fromBlock, "toBlock", toBlock)
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

	logs, err := s.rpcClient.FilterLogs(ctx, query)
	if err != nil {
		return err
	}

	for _, eventLog := range logs {
		if len(eventLog.Topics) == 0 {
			continue
		}
		if err := s.handleLogInTx(ctx, contractAddress, module, handler, eventLog); err != nil {
			log.Error("failed to handle event log", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module, "txHash", eventLog.TxHash.Hex(), "logIndex", eventLog.Index, "blockNumber", eventLog.BlockNumber, "error", err)
		}
	}

	return nil
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
