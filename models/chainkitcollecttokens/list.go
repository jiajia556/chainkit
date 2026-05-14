package chainkitcollecttokens

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainCollectTokens, *Record]
}

func NewList(session ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainCollectTokens))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainCollectTokens, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainCollectTokens, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) FindAvailableByChainDBID(chainDBID uint64) *List {
	l.DB().Where("chain_db_id = ?　AND status = 1", chainDBID).Find(l.Records)
	return l
}
