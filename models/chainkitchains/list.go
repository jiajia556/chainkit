package chainkitchains

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainChains, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(ctx) > 0 {
		dbSession = ctx[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainChains))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainChains, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainChains, *Record]{
			Session: dbSession,
			Records: &records,
		},
	}

	return l
}
