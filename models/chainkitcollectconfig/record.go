package chainkitcollectconfig

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainCollectConfig]
}

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainCollectConfig))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainCollectConfig]{
			Session: dbSession,
			Model:   new(ChainCollectConfig),
		},
	}
	return r
}
