package chainkiteventlogs

import (
	"strings"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainEventLogs]
}

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainEventLogs))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainEventLogs]{
			Session: dbContext,
			Model:   new(ChainEventLogs),
		},
	}
	return r
}

func (r *Record) ReadLastByContractAndChain(contractAddress, module string, chainDbId uint64) error {
	return r.DB().Where("contract_address = ? AND module = ? AND chain_db_id = ?", strings.ToLower(contractAddress), module, chainDbId).Order("id desc").Take(r.Model).Error
}

func (r *Record) ReadByChainTxHashAndLogIndex(chainDbId uint64, txHash string, logIndex uint32) error {
	return r.DB().Where("chain_db_id = ? AND tx_hash = ? AND log_index = ?", chainDbId, strings.ToLower(txHash), logIndex).Take(r.Model).Error
}
