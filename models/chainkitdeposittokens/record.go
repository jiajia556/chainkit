package chainkitdeposittokens

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainDepositTokens]
}

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		createTableSession := mysqlx.NewTxSession()
		err := createTableSession.CreateTableIfNotExists(new(ChainDepositTokens))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainDepositTokens]{
			Session: dbSession,
			Model:   new(ChainDepositTokens),
		},
	}
	return r
}

func (r *Record) ReadAvailableByChainAndToken(chainDbId, tokenId uint64) *Record {
	r.DB().Where("chain_db_id = ? AND token_id = ? AND status = 1", chainDbId, tokenId).Take(r.Model)
	return r
}
