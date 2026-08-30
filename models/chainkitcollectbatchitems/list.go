package chainkitcollectbatchitems

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainCollectBatchItems, *Record]
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
		if err := createTableSession.CreateTableIfNotExists(new(ChainCollectBatchItems)); err != nil {
			panic(err)
		}
	}
	records := make([]*ChainCollectBatchItems, 0)
	return &List{BaseList: &models.BaseList[*ChainCollectBatchItems, *Record]{
		Session: dbSession,
		Records: &records,
		RecordFactory: func() *Record {
			return NewRecord(dbSession)
		},
	}}
}

func (l *List) GetByBatchId(batchId uint64) *List {
	l.DB().Where("batch_id = ?", batchId).Order("item_index ASC").Find(l.Records)
	return l
}

func (l *List) GetWaitingByBatchId(batchId uint64) *List {
	l.DB().Where("batch_id = ? AND status = ?", batchId, StatusWaiting).
		Order("item_index ASC").Find(l.Records)
	return l
}
