package chainkitcontracts

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainContracts, *Record]
}

func NewList(session ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainContracts))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainContracts, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainContracts, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}
