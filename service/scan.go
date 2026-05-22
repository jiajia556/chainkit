package service

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/jiajia556/chainkit/models/chainkiteventlogs"
	"github.com/jiajia556/chainkit/models/chainkitscancursor"
	"github.com/jiajia556/tool-box/log"
	"github.com/jiajia556/tool-box/mysqlx"
)

type scanOptions struct {
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

type ScanOption func(*scanOptions)
type LogHandler func(mysqlSession *mysqlx.TxSession, log types.Log, eventLogId uint64, chainDbId uint64) error

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

func (s *ChainService) ScanBlock(contractAddress, module string, handler LogHandler, option ...ScanOption) error {
	log.Debug("starting scan block", "chainDbId", s.chainDbId, "contractAddress", contractAddress, "module", module)
	if s.client == nil {
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

	// Read cursor outside of the transaction to avoid holding locks during RPC calls.
	cursor := chainkitscancursor.NewRecord()
	if err := cursor.ReadByContractAndChain(contractAddress, module, s.chainDbId); err != nil && err.Error() != "record not found" {
		return err
	}
	if !cursor.Exists() {
		cursor.Model.ChainDbId = s.chainDbId
		cursor.Model.ContractAddress = contractAddress
		cursor.Model.Module = module
		cursor.Model.LastestBlock = opts.startBlock
		if err := cursor.Create(); err != nil {
			return err
		}
	} else {
		lastLogRecord := chainkiteventlogs.NewRecord()
		if err := lastLogRecord.ReadLastByContractAndChain(contractAddress, module, s.chainDbId); err != nil && err.Error() != "record not found" {
			return err
		}
		if lastLogRecord.Exists() {
			if lastLogRecord.Model.BlockNumber != cursor.Model.LastestBlock {
				return errors.New("last eventLog block number is not equal to cursor block number")
			}
		}
	}
	if cursor.Model.LastestBlock == 0 {
		return errors.New("start block is not set")
	}

	expectedCursorBlock := cursor.Model.LastestBlock

	header, err := s.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	if header.Number.Uint64() <= opts.safeConfirmations {
		return errors.New("latest block number is less than or equal to safe confirmations")
	}
	netSafeLastestBlock := header.Number.Uint64() - opts.safeConfirmations
	toBlock := cursor.Model.LastestBlock + opts.step
	if toBlock > netSafeLastestBlock {
		toBlock = netSafeLastestBlock
	}

	contract := common.HexToAddress(contractAddress)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(cursor.Model.LastestBlock + 1)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Topics:    opts.Topics,
		Addresses: []common.Address{contract},
	}

	logs, err := s.client.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}

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

	cursorTx := chainkitscancursor.NewRecord(session)
	if err := cursorTx.ReadByContractAndChain(contractAddress, module, s.chainDbId); err != nil && err.Error() != "record not found" {
		return rollbackWithErr(err)
	}
	if !cursorTx.Exists() {
		return rollbackWithErr(errors.New("scan cursor not found"))
	}
	if cursorTx.Model.LastestBlock != expectedCursorBlock {
		return rollbackWithErr(errors.New("scan cursor moved; retry scan"))
	}

	lastLogBlockNumber := uint64(0)
	for _, eventLog := range logs {
		if len(eventLog.Topics) == 0 {
			continue
		}
		//logRecord := chainkiteventlogs.NewRecord(session)
		//logRecord.Model.ChainDbId = s.chainDbId
		//logRecord.Model.ContractAddress = contractAddress
		//logRecord.Model.Module = module
		//logRecord.Model.TxHash = eventLog.TxHash.String()
		//logRecord.Model.LogIndex = uint32(eventLog.Index)
		//logRecord.Model.BlockNumber = eventLog.BlockNumber
		//logRecord.Model.BlockHash = eventLog.BlockHash.String()
		//logRecord.Model.EventSig = eventLog.Topics[0].Bytes()
		//logRecord.Model.RawData = eventLog.Data
		//if err := logRecord.Create(); err != nil {
		//	return rollbackWithErr(err)
		//}

		if err := handler(session, eventLog, 0, s.chainDbId); err != nil {
			return rollbackWithErr(err)
		}

		//lastLogBlockNumber = eventLog.BlockNumber
	}

	finalCursorBlock := toBlock
	if toBlock > lastLogBlockNumber {
		hash := uuid.NewString()
		logRecord := chainkiteventlogs.NewRecord(session)
		logRecord.Model.ChainDbId = s.chainDbId
		logRecord.Model.ContractAddress = contractAddress
		logRecord.Model.Module = module
		logRecord.Model.TxHash = hash
		logRecord.Model.LogIndex = 0
		logRecord.Model.BlockNumber = toBlock
		logRecord.Model.BlockHash = ""
		logRecord.Model.EventSig = []byte{}
		logRecord.Model.RawData = []byte{}
		if err := logRecord.Create(); err != nil {
			return rollbackWithErr(err)
		}
	}

	if finalCursorBlock > cursorTx.Model.LastestBlock {
		if err := cursorTx.UpdateLastestBlock(finalCursorBlock); err != nil {
			return rollbackWithErr(err)
		}
	}

	if err := session.Commit(); err != nil {
		return err
	}
	return nil
}
