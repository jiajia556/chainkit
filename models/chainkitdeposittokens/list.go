package chainkitdeposittokens

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainDepositTokens, *Record]
}

func NewList(session ...mysqlx.Session) *List {
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
	records := make([]*ChainDepositTokens, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainDepositTokens, *Record]{
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
	l.DB().Where("chain_db_id = ? AND status = 1", chainDBID).Find(l.Records)
	return l
}

func (l *List) HasAvailableByChainDBID(chainDBID uint64) (bool, error) {
	var count int64
	err := l.DB().Model(&ChainDepositTokens{}).
		Where("chain_db_id = ? AND status = 1", chainDBID).
		Limit(1).
		Count(&count).Error
	return count > 0, err
}
