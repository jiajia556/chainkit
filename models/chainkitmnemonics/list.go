package chainkitmnemonics

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainMnemonics, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(ctx) > 0 {
		dbSession = ctx[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainMnemonics))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainMnemonics, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainMnemonics, *Record]{
			Session: dbSession,
			Records: &records,
		},
	}

	return l
}
