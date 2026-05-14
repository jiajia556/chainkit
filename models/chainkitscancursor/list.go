package chainkitscancursor

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainScanCursor, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(ctx) > 0 {
		dbSession = ctx[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainScanCursor))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainScanCursor, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainScanCursor, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}
