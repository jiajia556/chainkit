package chainkittokens

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainTokens, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainTokens))
		if err != nil {
			panic(err)
		}
	}
	l := &List{
		BaseList: &models.BaseList[*ChainTokens, *Record]{
			Session: dbContext,
			Records: make([]*ChainTokens, 0),
		},
	}

	return l
}

func (l *List) FindByChain(chainDbId uint64) *List {
	l.DB().Where("chain_db_id = ?", chainDbId).Find(&l.Records)
	return l
}
