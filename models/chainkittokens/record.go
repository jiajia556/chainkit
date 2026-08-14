package chainkittokens

import (
	"strings"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainTokens]
}

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainTokens))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainTokens]{
			Session: dbContext,
			Model:   new(ChainTokens),
		},
	}
	return r
}

func (r *Record) ReadByChainAndContractAddress(chainDbId uint64, contractAddress string) *Record {
	r.DB().Where("contract_address = ? AND chain_db_id = ?", strings.ToLower(contractAddress), chainDbId).Take(r.Model)
	return r
}

func (r *Record) ReadByChainAndSymbol(chainDbId uint64, symbol string) *Record {
	r.DB().Where("symbol = ? AND chain_db_id = ?", symbol, chainDbId).Take(r.Model)
	return r
}
