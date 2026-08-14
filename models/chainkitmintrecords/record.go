package chainkitmintrecords

import (
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/types"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainMintRecords]
}

const (
	StatusPending int8 = 1
	StatusSuccess int8 = 2
	StatusFailed  int8 = -1
	StatusUnknown int8 = -2
)

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		createTableSession := mysqlx.NewTxSession()
		err := createTableSession.CreateTableIfNotExists(new(ChainMintRecords))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainMintRecords]{
			Session: dbSession,
			Model:   new(ChainMintRecords),
		},
	}
	return r
}

func (r *Record) ReadPending(chainDbId uint64, addressType types.ServiceAddressType, addressId uint64) error {
	return r.DB().Where("from_address_id = ? AND from_address_type = ? AND status = ? AND chain_db_id = ?",
		addressId, addressType, StatusPending, chainDbId).Last(r.Model).Error
}

func (r *Record) SetSuccess() error {
	return r.DB().Model(r.Model).Update("status", StatusSuccess).Error
}

func (r *Record) SetFailed() error {
	return r.DB().Model(r.Model).Update("status", StatusFailed).Error
}

func (r *Record) SetUnknown() error {
	return r.DB().Model(r.Model).Update("status", StatusUnknown).Error
}

func (r *Record) SinceCreated() time.Duration {
	return time.Since(r.Model.CreatedAt)
}
