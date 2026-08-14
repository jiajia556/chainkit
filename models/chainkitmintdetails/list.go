package chainkitmintdetails

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/types"
	"github.com/jiajia556/tool-box/mysqlx"
)

type List struct {
	*models.BaseList[*ChainMintDetails, *Record]
}

func NewList(session ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		createTableSession := mysqlx.NewTxSession()
		err := createTableSession.CreateTableIfNotExists(new(ChainMintDetails))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainMintDetails, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainMintDetails, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}

func (l *List) FindByFromAddressIdAndStatus(fromAddressId, tokenId uint64, fromAddressType types.ServiceAddressType, status int8, count int) error {
	return l.DB().Where("from_address_id = ? AND status = ? AND from_address_type = ? AND token_id = ?", fromAddressId, status, fromAddressType, tokenId).
		Order("id ASC").
		Limit(count).Find(l.Records).Error
}
