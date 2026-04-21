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

var defaultScanOptions = &scanOptions{
	safeConfirmations: 20,
	step:              1000,
}

type ScanOption func(*scanOptions)
type LogHandler func(mysqlSession *mysqlx.TxSession, log types.Log, eventLogId uint64) error

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

func (s *ChainService) ScanBlock(contractAddress string, handler LogHandler, option ...ScanOption) error {
	opts := *defaultScanOptions
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
	cursor.ReadByContractAndChain(contractAddress, s.chainDbId)
	if !cursor.Exists() {
		cursor.Model.ChainDbId = s.chainDbId
		cursor.Model.ContractAddress = contractAddress
		cursor.Model.LastestBlock = opts.startBlock
		cursor.Create()
	} else {
		lastLogRecord := chainkiteventlogs.NewRecord(session)
		lastLogRecord.ReadLastByContractAndChain(contractAddress, s.chainDbId)
		if !lastLogRecord.Exists() {
			session.Rollback()
			return errors.New("last log record not found")
		}
		if lastLogRecord.Model.BlockNumber != cursor.Model.LastestBlock {
			session.Rollback()
			return errors.New("last log block number is not equal to cursor block number")
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
		logRecord := chainkiteventlogs.NewRecord(session)
		logRecord.Model.ChainDbId = s.chainDbId
		logRecord.Model.ContractAddress = contractAddress
		logRecord.Model.TxHash = log.TxHash.String()
		logRecord.Model.LogIndex = uint32(log.Index)
		logRecord.Model.BlockNumber = log.BlockNumber
		logRecord.Model.BlockHash = log.BlockHash.String()
		logRecord.Model.EventSig = log.Topics[0].Bytes()
		logRecord.Model.RawData = log.Data
		logRecord.Create()

		if err := handler(session, log, logRecord.Model.Id); err != nil {
			session.Rollback()
			return err
		}

		lastLogBlockNumber = log.BlockNumber
		cursor.UpdateLastestBlock(lastLogBlockNumber)
	}
	if toBlock > lastLogBlockNumber || len(logs) == 0 {
		logRecord := chainkiteventlogs.NewRecord(session)
		logRecord.Model.ChainDbId = s.chainDbId
		logRecord.Model.ContractAddress = contractAddress
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
