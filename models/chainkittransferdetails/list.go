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
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainTransferDetails))
		if err != nil {
			panic(err)
		}
	}
	l := &List{
		BaseList: &models.BaseList[*ChainTransferDetails, *Record]{
			Session: dbContext,
			Records: make([]*ChainTransferDetails, 0),
		},
	}

	return l
}

func (l *List) FindByFromAddressIdAndStatus(fromAddressId uint64, fromAddressType types.ServiceAddressType, status int8, count int) error {
	return l.DB().Where("from_address_id = ? AND from_address_type = ? AND status = ?", fromAddressId, fromAddressType, status).
		Order("id ASC").
		Limit(count).Find(&l.Records).Error
}
