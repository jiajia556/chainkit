package chainkitcollectbatchitems

import (
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainCollectBatchItems]
}

const (
	StatusWaiting = iota
	StatusSuccess
	StatusFailed
	StatusNotAttempted
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
		if err := createTableSession.CreateTableIfNotExists(new(ChainCollectBatchItems)); err != nil {
			panic(err)
		}
	}
	return &Record{BaseRecord: &models.BaseRecord[*ChainCollectBatchItems]{
		Session: dbSession,
		Model:   new(ChainCollectBatchItems),
	}}
}

func (r *Record) GetByBatchAndTaskId(batchId, collectTaskId uint64) *Record {
	r.DB().Where("batch_id = ? AND collect_task_id = ?", batchId, collectTaskId).Take(r.Model)
	return r
}

func (r *Record) SetResult(status uint8, resultCode uint8, actualAmount, returnDataHash, lastError string) error {
	return r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":           status,
		"result_code":      resultCode,
		"actual_amount":    actualAmount,
		"return_data_hash": returnDataHash,
		"last_error":       lastError,
	}).Error
}

func (r *Record) MarkNotAttempted(lastError string) error {
	return r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusNotAttempted,
		"last_error": lastError,
	}).Error
}
