package chainkitcollecttasks

import (
	"fmt"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainCollectTasks]
}

const (
	StatusWaiting = iota
	StatusNeedGas
	StatusCanSend
	StatusSending
	StatusSent
	StatusConfirmed
	StatusFailed
	StatusCancel
	StatusSkip
)

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainCollectTasks))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainCollectTasks]{
			Session: dbSession,
			Model:   new(ChainCollectTasks),
		},
	}
	return r
}

func (r *Record) GetTaskingByAddressAndToken(userDepositAddressId, tokenId uint64) *Record {
	where := fmt.Sprintf(
		"user_deposit_address_id = %d AND token_id = %d AND status IN (%s)",
		userDepositAddressId,
		tokenId,
		"0,1,2,3,4",
	)
	r.DB().Where(where).Take(r.Model)
	return r
}

func (r *Record) SetStatus(status int8) {
	r.DB().Model(r.Model).Update("status", status)
}
