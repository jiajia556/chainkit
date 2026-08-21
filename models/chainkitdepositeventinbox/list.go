package chainkitdepositeventinbox

import (
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
	"gorm.io/gorm"
)

type List struct {
	*models.BaseList[*ChainDepositEventInbox, *Record]
}

// FindProcessable returns mature pending events and due retry events. Callers
// must still Claim each record before processing it.
func (l *List) FindProcessable(chainDbID, safeBlock uint64, now time.Time, limit int) error {
	if limit <= 0 {
		limit = 100
	}
	return l.DB().
		Where("chain_db_id = ? AND block_number <= ? AND removed = 0", chainDbID, safeBlock).
		Where("status IN ?", []int8{StatusPending, StatusReady, StatusRetry}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("block_number ASC, log_index ASC, id ASC").
		Limit(limit).
		Find(l.Records).Error
}

// RequeueStaleProcessing makes tasks abandoned by a crashed worker retryable.
func (l *List) RequeueStaleProcessing(chainDbID uint64, updatedBefore, nextRetryAt time.Time) (int64, error) {
	res := l.DB().Model(&ChainDepositEventInbox{}).
		Where("chain_db_id = ? AND status = ? AND updated_at < ?", chainDbID, StatusProcessing, updatedBefore).
		Updates(map[string]interface{}{
			"status":        StatusRetry,
			"retry_count":   gorm.Expr("retry_count + 1"),
			"next_retry_at": nextRetryAt,
			"last_error":    "processing lease expired",
		})
	return res.RowsAffected, res.Error
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
		err := createTableSession.CreateTableIfNotExists(new(ChainDepositEventInbox))
		if err != nil {
			panic(err)
		}
	}
	records := make([]*ChainDepositEventInbox, 0)
	l := &List{
		BaseList: &models.BaseList[*ChainDepositEventInbox, *Record]{
			Session: dbSession,
			Records: &records,
			RecordFactory: func() *Record {
				return NewRecord(dbSession)
			},
		},
	}

	return l
}
