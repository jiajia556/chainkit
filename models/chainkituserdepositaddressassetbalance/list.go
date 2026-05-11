package chainkituserdepositaddressassetbalance

import (
	"fmt"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

type List struct {
	*models.BaseList[*ChainUserDepositAddressAssetBalance, *Record]
}

func NewList(session ...mysqlx.Session) *List {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainUserDepositAddressAssetBalance))
		if err != nil {
			panic(err)
		}
	}
	l := &List{
		BaseList: &models.BaseList[*ChainUserDepositAddressAssetBalance, *Record]{
			Session: dbSession,
			Records: make([]*ChainUserDepositAddressAssetBalance, 0),
		},
	}

	return l
}

func (l *List) GetCanTaskByTokenBalance(tokenId uint64, minCollectAmount decimal.Decimal) *List {
	where := fmt.Sprintf("token_id = %d AND balance_amount >= %s", tokenId, minCollectAmount.String())
	l.DB().Where(where).Find(&l.Records)
	return l
}
