package chainkitmnemonicaddresses

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainMnemonicAddresses, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainMnemonicAddresses))
		if err != nil {
			panic(err)
		}
	}
	l := &List{
		BaseList: &models.BaseList[*ChainMnemonicAddresses, *Record]{
			Session: dbContext,
			Records: make([]*ChainMnemonicAddresses, 0),
		},
	}

	return l
}
