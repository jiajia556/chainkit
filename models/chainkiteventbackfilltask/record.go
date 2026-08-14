package chainkiteventbackfilltask

import (
	"fmt"
	"strings"
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainEventBackfillTask]
}

const (
	StatusWaiting int8 = 0
	StatusRunning int8 = 1
	StatusDone    int8 = 2
	StatusFailed  int8 = -1
)

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainEventBackfillTask))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainEventBackfillTask]{
			Session: dbSession,
			Model:   new(ChainEventBackfillTask),
		},
	}
	return r
}

func (r *Record) SetRunning() error {
	return r.DB().
		Model(r.Model).
		Where("id = ? AND status IN (?)", r.Model.Id, []int8{StatusWaiting, StatusRunning}).
		Updates(map[string]interface{}{
			"status":     StatusRunning,
			"remark":     "",
			"updated_at": time.Now(),
		}).Error
}

func (r *Record) SetCurrentBlock(currentBlock uint64) error {
	if err := r.Read(r.Model.Id); err != nil {
		return err
	}
	if r.Model.Status != StatusRunning {
		return fmt.Errorf("backfill task status changed: %d", r.Model.Status)
	}
	r.Model.CurrentBlock = currentBlock
	r.Model.UpdatedAt = time.Now()
	return r.Update()
}

func (r *Record) SetDone(endBlock uint64) error {
	return r.DB().
		Model(r.Model).
		Where("id = ?", r.Model.Id).
		Updates(map[string]interface{}{
			"status":        StatusDone,
			"current_block": endBlock,
			"remark":        "",
			"updated_at":    time.Now(),
		}).Error
}

func (r *Record) SetFailed(taskErr error) error {
	remark := ""
	if taskErr != nil {
		remark = taskErr.Error()
	}
	if len(remark) > 255 {
		remark = remark[:255]
	}
	remark = strings.TrimSpace(remark)

	return r.DB().
		Model(r.Model).
		Where("id = ?", r.Model.Id).
		Updates(map[string]interface{}{
			"status":     StatusFailed,
			"remark":     remark,
			"updated_at": time.Now(),
		}).Error
}
