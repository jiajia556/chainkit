package service

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkiteventlogs"
	"github.com/jiajia556/tool-box/mysqlx"
)

type LogContext struct {
	Session         mysqlx.Session
	ChainDbId       uint64
	ContractAddress string
	Module          string
}

func (c *LogContext) SaveEventLog(eventLog types.Log) (uint64, bool, error) {
	if c == nil || c.Session == nil {
		return 0, false, errors.New("log context session is nil")
	}
	if len(eventLog.Topics) == 0 {
		return 0, false, errors.New("event log topics is empty")
	}

	logRecord := chainkiteventlogs.NewRecord(c.Session)
	logRecord.Model.ChainDbId = c.ChainDbId
	logRecord.Model.ContractAddress = c.ContractAddress
	logRecord.Model.Module = c.Module
	logRecord.Model.TxHash = strings.ToLower(eventLog.TxHash.Hex())
	logRecord.Model.LogIndex = uint32(eventLog.Index)
	logRecord.Model.BlockNumber = eventLog.BlockNumber
	logRecord.Model.BlockHash = eventLog.BlockHash.Hex()
	logRecord.Model.EventSig = eventLog.Topics[0].Bytes()
	logRecord.Model.RawData = eventLog.Data
	if err := logRecord.Create(); err != nil {
		existing := chainkiteventlogs.NewRecord(c.Session)
		if readErr := existing.ReadByChainTxHashAndLogIndex(c.ChainDbId, eventLog.TxHash.Hex(), uint32(eventLog.Index)); readErr == nil && existing.Exists() {
			return existing.Model.Id, false, nil
		}
		return 0, false, err
	}

	return logRecord.Model.Id, true, nil
}
