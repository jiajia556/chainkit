package service

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkiteventlogs"
	"github.com/jiajia556/chainkit/models/chainkitscancursor"
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
	opts := defaultScanOptions.Clone()
	for _, apply := range option {
		apply(&opts)
	}

	session := mysqlx.NewTxSession()
	session.Begin()
	defer func() {
		e := recover()
		if e != nil {
			session.Rollback()
			panic(e)
		}
	}()

	cursor := chainkitscancursor.NewRecord(session)
	cursor.ReadByContractAndChain(contractAddress, module, s.chainDbId)
	if !cursor.Exists() {
		cursor.Model.ChainDbId = s.chainDbId
		cursor.Model.ContractAddress = contractAddress
		cursor.Model.Module = module
		cursor.Model.LastestBlock = opts.startBlock
		cursor.Create()
	} else {
		lastLogRecord := chainkiteventlogs.NewRecord(session)
		lastLogRecord.ReadLastByContractAndChain(contractAddress, module, s.chainDbId)
		if lastLogRecord.Exists() {
			if lastLogRecord.Model.BlockNumber != cursor.Model.LastestBlock {
				session.Rollback()
				return errors.New("last log block number is not equal to cursor block number")
			}
		}
	}
	if cursor.Model.LastestBlock == 0 {
		session.Rollback()
		return errors.New("start block is not set")
	}

	header, err := s.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		session.Rollback()
		return err
	}

	if header.Number.Uint64() <= opts.safeConfirmations {
		session.Rollback()
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
		Addresses: []common.Address{
			contract,
		},
	}

	logs, err := s.client.FilterLogs(context.Background(), query)
	if err != nil {
		session.Rollback()
		return err
	}

	lastLogBlockNumber := uint64(0)
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		logRecord := chainkiteventlogs.NewRecord(session)
		logRecord.Model.ChainDbId = s.chainDbId
		logRecord.Model.ContractAddress = contractAddress
		logRecord.Model.Module = module
		logRecord.Model.TxHash = log.TxHash.String()
		logRecord.Model.LogIndex = uint32(log.Index)
		logRecord.Model.BlockNumber = log.BlockNumber
		logRecord.Model.BlockHash = log.BlockHash.String()
		logRecord.Model.EventSig = log.Topics[0].Bytes()
		logRecord.Model.RawData = log.Data
		logRecord.Create()

		if err := handler(session, log, logRecord.Model.Id, s.chainDbId); err != nil {
			session.Rollback()
			return err
		}

		lastLogBlockNumber = log.BlockNumber
	}
	if lastLogBlockNumber > cursor.Model.LastestBlock {
		cursor.UpdateLastestBlock(lastLogBlockNumber)
	}
	if toBlock > lastLogBlockNumber || len(logs) == 0 {
		logRecord := chainkiteventlogs.NewRecord(session)
		logRecord.Model.ChainDbId = s.chainDbId
		logRecord.Model.ContractAddress = contractAddress
		logRecord.Model.Module = module
		logRecord.Model.TxHash = ""
		logRecord.Model.LogIndex = 0
		logRecord.Model.BlockNumber = toBlock
		logRecord.Model.BlockHash = ""
		logRecord.Model.EventSig = nil
		logRecord.Model.RawData = nil
		logRecord.Create()
		cursor.UpdateLastestBlock(toBlock)
	}

	session.Commit()
	return nil
}
