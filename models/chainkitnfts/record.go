package chainkitnfts

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainNfts]
}

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		createTableSession := mysqlx.NewTxSession()
		err := createTableSession.CreateTableIfNotExists(new(ChainNfts))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainNfts]{
			Session: dbSession,
			Model:   new(ChainNfts),
		},
	}
	return r
}
