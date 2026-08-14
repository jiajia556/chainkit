package chainkitmintdetails

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainMintDetails]
}

const (
	StatusWaiting int8 = 0
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
		err := createTableSession.CreateTableIfNotExists(new(ChainMintDetails))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainMintDetails]{
			Session: dbSession,
			Model:   new(ChainMintDetails),
		},
	}
	return r
}

func (r *Record) SetSuccessByTransferRecordId(mintRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).
		Where("mint_record_id = ?", mintRecordId).
		Update("status", StatusSuccess).Error
}

func (r *Record) SetFailedByTransferRecordId(mintRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).
		Where("mint_record_id = ?", mintRecordId).
		Update("status", StatusFailed).Error
}

func (r *Record) SetWaitingByTransferRecordId(mintRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).
		Where("mint_record_id = ?", mintRecordId).
		Update("status", StatusWaiting).Error
}

func (r *Record) SetUnknownByTransferRecordId(mintRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).
		Where("mint_record_id = ?", mintRecordId).
		Update("status", StatusUnknown).Error
}

func (r *Record) SetPending(ids []uint64, mintRecordId uint64) error {
	return r.DB().Table(r.Model.TableName()).Where("id in ?", ids).
		Updates(map[string]interface{}{
			"status":         StatusPending,
			"mint_record_id": mintRecordId,
		}).Error
}
