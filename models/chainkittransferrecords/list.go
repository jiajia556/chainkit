package chainkittransferrecords

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainTransferRecords, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(ctx) > 0 {
		dbSession = ctx[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainTransferRecords))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainTransferRecords, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainTransferRecords, *Record]{
			Session: dbSession,
			Records: &records,
		},
	}

	return l
}
