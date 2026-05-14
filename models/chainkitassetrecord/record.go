package chainkitassetrecord

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainAssetRecord]
}

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainAssetRecord))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainAssetRecord]{
			Session: dbSession,
			Model:   new(ChainAssetRecord),
		},
	}
	return r
}

func (l *List) GetByUserIDAndTokenGroupID(userID uint64, modules []string, start, limit int, symbol []string) *List {
	var total int64
	l.DB().Where("user_id = ? AND symbol IN (?) and module IN (?)", userID, symbol, modules).Offset(start).Limit(limit).Order("id DESC").Find(&l.Records)
	l.DB().Model(&ChainAssetRecord{}).Where("user_id = ? AND symbol IN (?) and module IN (?)", userID, symbol, modules).Count(&total)
	l.SetTotal(total)
	return l
}
