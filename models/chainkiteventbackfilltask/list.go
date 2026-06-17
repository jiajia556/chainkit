package chainkiteventbackfilltask

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainEventBackfillTask, *Record]
}

func NewList(session ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainEventBackfillTask))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainEventBackfillTask, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainEventBackfillTask, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) FindRunnable(limit int) error {
	if limit <= 0 {
		limit = 10
	}
	return l.DB().
		Where("status IN (?)", []int8{StatusWaiting, StatusRunning}).
		Order("id asc").
		Limit(limit).
		Find(l.Records).Error
}
