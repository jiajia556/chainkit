package chainkitcollecttasks

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainCollectTasks, *Record]
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
		err := createTableSession.CreateTableIfNotExists(new(ChainCollectTasks))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainCollectTasks, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainCollectTasks, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) GetWaitingList(chainDbId uint64) *List {
	l.DB().Where("chain_db_id = ? AND status = 0", chainDbId).Find(l.Records)
	return l
}

func (l *List) GetCanSendList(chainDbId uint64) *List {
	l.DB().Where("chain_db_id = ? AND status = 2", chainDbId).Find(l.Records)
	return l
}

func (l *List) GetCentList(chainDbId uint64) *List {
	l.DB().Debug().Where("chain_db_id = ? AND status IN (?)", chainDbId, []int{StatusSent, StatusMaybeSent}).Find(l.Records)
	return l
}
