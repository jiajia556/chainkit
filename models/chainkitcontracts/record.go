package chainkitcontracts

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainContracts]
}

func NewRecord(session ...mysqlx.Session) *Record {
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
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainContracts]{
			Session: dbSession,
			Model:   new(ChainContracts),
		},
	}
	return r
}

func (r *Record) ReadByNameAndChainDbId(name string, chainDbId uint64) error {
	return r.DB().Where("name = ? AND chain_db_id = ?", name, chainDbId).Take(r.Model).Error
}
