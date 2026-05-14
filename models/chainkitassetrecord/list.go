package chainkitassetrecord

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainAssetRecord, *Record]
}

func NewList(session ...mysqlx.Session) *List {
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
	records := make([]*ChainAssetRecord, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainAssetRecord, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) GetByUserIDAndTokenGroupID(userId uint64, modules []string, start int, limit int, symbols []string) *List {
	var total int64
	l.DB().Debug().Where("user_id = ? AND symbol IN (?) and biz_type IN (?)", userId, symbols, modules).Offset(start).Limit(limit).Order("id DESC").Find(l.Records)
	l.DB().Model(&ChainAssetRecord{}).Where("user_id = ? AND symbol IN (?) and biz_type IN (?)", userId, symbols, modules).Count(&total)
	l.SetTotal(total)
	return l
}
