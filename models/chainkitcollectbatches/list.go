package chainkitcollectbatches

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainCollectBatches, *Record]
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
		if err := createTableSession.CreateTableIfNotExists(new(ChainCollectBatches)); err != nil {
			panic(err)
		}
	}
	records := make([]*ChainCollectBatches, 0)
	return &List{BaseList: &models.BaseList[*ChainCollectBatches, *Record]{
		Session: dbSession,
		Records: &records,
		RecordFactory: func() *Record {
			return NewRecord(dbSession)
		},
	}}
}

func (l *List) GetWaitingList(chainDbId uint64, limit int) *List {
	query := l.DB().Where("chain_db_id = ? AND status = ?", chainDbId, StatusWaiting).
		Order("id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	query.Find(l.Records)
	return l
}

func (l *List) GetBroadcastedList(chainDbId uint64) *List {
	l.DB().Where(
		"chain_db_id = ? AND status IN (?)",
		chainDbId,
		[]int{StatusSent, StatusMaybeSent},
	).Order("sent_at ASC, id ASC").Find(l.Records)
	return l
}
