package chainkittokens

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainTokens, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(ctx) > 0 {
		dbSession = ctx[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainTokens))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainTokens, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainTokens, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) FindByChain(chainDbId uint64) *List {
	l.DB().Where("chain_db_id = ?", chainDbId).Find(l.Records)
	return l
}
