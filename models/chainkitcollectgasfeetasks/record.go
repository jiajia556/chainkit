package chainkitcollectgasfeetasks

import (
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

type Record struct {
	*models.BaseRecord[*ChainCollectGasFeeTasks]
}

const (
	StatusWaiting = iota
	StatusSending
	StatusSent
	StatusConfirmed
	StatusFailed
	StatusCancel
	StatusSkip
	StatusUnknown
)

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbSession.CreateTableIfNotExists(new(ChainCollectGasFeeTasks))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainCollectGasFeeTasks]{
			Session: dbSession,
			Model:   new(ChainCollectGasFeeTasks),
		},
	}
	return r
}

func (r *Record) GetOneWaiting(chainDbId uint64) *Record {
	r.DB().Where("chain_db_id = ? AND status = 0", chainDbId).First(r.Model)
	return r
}

func (r *Record) GetLastSent(chainDbId uint64) *Record {
	r.DB().Where("chain_db_id = ? AND status = 2", chainDbId).Take(r.Model)
	return r
}

func (r *Record) SinceCreated() time.Duration {
	return time.Since(r.Model.CreatedAt)
}

func (r *Record) SetWaiting() {
	r.DB().Model(r.Model).Update("status", StatusWaiting)
}

func (r *Record) SetSending() {
	r.DB().Model(r.Model).Update("status", StatusSending)
}

func (r *Record) SetConfirmed(gasUsed, txFee decimal.Decimal) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":       StatusConfirmed,
		"gas_used":     gasUsed,
		"tx_fee":       txFee,
		"confirmed_at": time.Now(),
	})
}

func (r *Record) SetFailed() {
	r.DB().Model(r.Model).Update("status", StatusFailed)
}

func (r *Record) SetUnknown() {
	r.DB().Model(r.Model).Update("status", StatusUnknown)
}

func (r *Record) SetSent(sendAmount decimal.Decimal, hash string, nonce uint64, gasPrice decimal.Decimal) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":      StatusSent,
		"send_amount": sendAmount,
		"tx_hash":     hash,
		"nonce":       nonce,
		"gas_price":   gasPrice,
		"sent_at":     time.Now(),
	})
}
