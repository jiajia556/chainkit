package chainkittransferdetails

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainTransferDetails]
}

const (
	StatusWaiting int8 = 0
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
		err := dbContext.CreateTableIfNotExists(new(ChainTransferDetails))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainTransferDetails]{
			Session: dbContext,
			Model:   new(ChainTransferDetails),
		},
	}
	return r
}

func (r *Record) SetSuccessByTransferRecordId(transferRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).
		Where("transfer_record_id = ?", transferRecordId).
		Update("status", StatusSuccess).Error
}

func (r *Record) SetFailedByTransferRecordId(transferRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).
		Where("transfer_record_id = ?", transferRecordId).
		Update("status", StatusFailed).Error
}

func (r *Record) SetPending(ids []uint64, transferRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).Where("id in ?", ids).
		Updates(map[string]interface{}{
			"status":             StatusPending,
			"transfer_record_id": transferRecordId,
		}).Error
}
