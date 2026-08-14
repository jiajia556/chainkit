package chainkittransferdetails

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/types"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainTransferDetails, *Record]
}

func NewList(ctx ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(ctx) > 0 {
		dbSession = ctx[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		createTableSession := mysqlx.NewTxSession()
		err := createTableSession.CreateTableIfNotExists(new(ChainTransferDetails))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainTransferDetails, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainTransferDetails, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) FindByFromAddressIdAndStatus(fromAddressId uint64, fromAddressType types.ServiceAddressType, status int8, count int) error {
	return l.DB().Where("from_address_id = ? AND from_address_type = ? AND status = ?", fromAddressId, fromAddressType, status).
		Order("id ASC").
		Limit(count).Find(l.Records).Error
}
