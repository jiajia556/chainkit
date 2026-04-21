package chainkitscancursor

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainScanCursor]
}

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainScanCursor))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainScanCursor]{
			Session: dbContext,
			Model:   new(ChainScanCursor),
		},
	}
	return r
}

func (r *Record) ReadByContractAndChain(contractAddress string, chainDbId uint64) error {
	return r.DB().Where("contract_address = ? AND chain_db_id = ?", contractAddress, chainDbId).Take(&r.Model).Error
}

func (r *Record) UpdateLastestBlock(lastestBlock uint64) error {
	return r.DB().Model(r.Model).Update("lastest_block", lastestBlock).Error
}
