package chainkitdepositeventinbox

import (
	"errors"
	"strings"
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Record struct {
	*models.BaseRecord[*ChainDepositEventInbox]
}

// Upsert merges websocket and getLogs observations using Source as a bit mask.
// A getLogs observation is authoritative and can revive an event previously
// marked orphaned when the same transaction is included in the canonical chain.
func (r *Record) Upsert() error {
	if r == nil || r.Model == nil {
		return errors.New("deposit inbox record is nil")
	}
	if r.Model.Source != SourceWebSocket && r.Model.Source != SourceGetLogs {
		return errors.New("invalid deposit inbox source")
	}
	r.Model.normalize()

	updates := map[string]interface{}{
		"source":           gorm.Expr("source | ?", r.Model.Source),
		"contract_address": r.Model.ContractAddress,
		"block_number":     r.Model.BlockNumber,
		"block_hash":       r.Model.BlockHash,
		"from_address":     r.Model.FromAddress,
		"to_address":       r.Model.ToAddress,
		"amount":           r.Model.Amount,
		"updated_at":       time.Now(),
	}
	if r.Model.Source == SourceGetLogs {
		updates["removed"] = false
		updates["status"] = gorm.Expr(
			"CASE WHEN status = ? THEN ? ELSE status END",
			StatusOrphaned,
			StatusPending,
		)
	}

	return r.DB().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "chain_db_id"},
			{Name: "tx_hash"},
			{Name: "log_index"},
		},
		DoUpdates: clause.Assignments(updates),
	}).Create(r.Model).Error
}

func (r *Record) Claim(now time.Time) (bool, error) {
	if r == nil || r.Model == nil || r.Model.Id == 0 {
		return false, errors.New("deposit inbox record is not loaded")
	}
	res := r.DB().Model(r.Model).
		Where("id = ? AND removed = 0 AND status IN ? AND (next_retry_at IS NULL OR next_retry_at <= ?)",
			r.Model.Id,
			[]int8{StatusPending, StatusReady, StatusRetry},
			now,
		).
		Updates(map[string]interface{}{
			"status":     StatusProcessing,
			"last_error": "",
		})
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 1 {
		r.Model.Status = StatusProcessing
		return true, nil
	}
	return false, nil
}

func (r *Record) MarkConfirmed(now time.Time) error {
	return r.updateProcessing(map[string]interface{}{
		"confirmed_at": now,
	})
}

func (r *Record) MarkCompleted(now time.Time) error {
	return r.updateProcessing(map[string]interface{}{
		"status":        StatusCompleted,
		"processed_at":  now,
		"next_retry_at": nil,
		"last_error":    "",
	})
}

func (r *Record) MarkRetry(lastError string, nextRetryAt time.Time) error {
	return r.updateProcessing(map[string]interface{}{
		"status":        StatusRetry,
		"retry_count":   gorm.Expr("retry_count + 1"),
		"next_retry_at": nextRetryAt,
		"last_error":    truncate(lastError, 1024),
	})
}

func (r *Record) MarkIgnored(reason string, now time.Time) error {
	return r.updateProcessing(map[string]interface{}{
		"status":        StatusIgnored,
		"processed_at":  now,
		"next_retry_at": nil,
		"last_error":    truncate(reason, 1024),
	})
}

func (r *Record) MarkOrphaned(reason string, now time.Time) error {
	if r == nil || r.Model == nil || r.Model.Id == 0 {
		return errors.New("deposit inbox record is not loaded")
	}
	return r.DB().Model(r.Model).
		Where("id = ? AND status <> ?", r.Model.Id, StatusCompleted).
		Updates(map[string]interface{}{
			"status":        StatusOrphaned,
			"removed":       true,
			"processed_at":  now,
			"next_retry_at": nil,
			"last_error":    truncate(reason, 1024),
		}).Error
}

func (r *Record) updateProcessing(updates map[string]interface{}) error {
	if r == nil || r.Model == nil || r.Model.Id == 0 {
		return errors.New("deposit inbox record is not loaded")
	}
	res := r.DB().Model(r.Model).
		Where("id = ? AND status = ?", r.Model.Id, StatusProcessing).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return errors.New("deposit inbox record is not in processing state")
	}
	return nil
}

func MarkRemoved(session mysqlx.Session, chainDbID uint64, txHash string, logIndex uint32, reason string, now time.Time) error {
	res := session.DB().Model(&ChainDepositEventInbox{}).
		Where("chain_db_id = ? AND tx_hash = ? AND log_index = ? AND status <> ?",
			chainDbID,
			strings.ToLower(txHash),
			logIndex,
			StatusCompleted,
		).
		Updates(map[string]interface{}{
			"status":        StatusOrphaned,
			"removed":       true,
			"processed_at":  now,
			"next_retry_at": nil,
			"last_error":    truncate(reason, 1024),
		})
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return res.Error
	}
	return nil
}

func truncate(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func NewRecord(session ...mysqlx.Session) *Record {
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
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainDepositEventInbox]{
			Session: dbSession,
			Model:   new(ChainDepositEventInbox),
		},
	}
	return r
}
