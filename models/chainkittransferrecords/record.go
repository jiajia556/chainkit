package chainkittransferrecords

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/types"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainTransferRecords]
}

const (
	StatusPending int8 = 1
	StatusSuccess int8 = 2
	StatusFailed  int8 = -1
)

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainTransferRecords))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainTransferRecords]{
			Session: dbContext,
			Model:   new(ChainTransferRecords),
		},
	}
	return r
}

func (r *Record) ReadPending(chainDbId uint64, addressType types.ServiceAddressType, addressId uint64) error {
	return r.DB().Where("address_id = ? AND from_address_type = ? AND status = ? AND chain_db_id = ?",
		addressId, addressType, StatusPending, chainDbId).Take(r.Model).Error
}

func (r *Record) SetSuccess() error {
	return r.DB().Model(r.Model).Update("status", StatusSuccess).Error
}

func (r *Record) SetFailed() error {
	return r.DB().Model(r.Model).Update("status", StatusFailed).Error
}
