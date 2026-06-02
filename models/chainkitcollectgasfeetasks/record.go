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
	StatusMaybeSent = 10
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
		err := createTableSession.CreateTableIfNotExists(new(ChainCollectGasFeeTasks))
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
	r.DB().Where("chain_db_id = ? AND status IN (?)", chainDbId, []int{StatusSent, StatusMaybeSent}).Order("sent_at ASC, id ASC").Take(r.Model)
	return r
}

func (r *Record) SinceCreated() time.Duration {
	return time.Since(r.Model.CreatedAt)
}

func (r *Record) SinceSent() time.Duration {
	if r.Model.SentAt.IsZero() {
		return r.SinceCreated()
	}
	return time.Since(r.Model.SentAt)
}

func (r *Record) SetWaiting() {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusWaiting,
		"created_at": time.Now(),
	})
}

func (r *Record) SetSending() {
	r.DB().Model(r.Model).Update("status", StatusSending)
}

func (r *Record) ClaimWaiting() (bool, error) {
	res := r.DB().
		Model(r.Model).
		Where("id = ? AND status = ?", r.Model.Id, StatusWaiting).
		Updates(map[string]interface{}{
			"status":     StatusSending,
			"last_error": "",
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
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
		"last_error":  "",
	})
}

func (r *Record) SetMaybeSent(sendAmount decimal.Decimal, hash string, nonce uint64, gasPrice decimal.Decimal, lastError string) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":      StatusMaybeSent,
		"send_amount": sendAmount,
		"tx_hash":     hash,
		"nonce":       nonce,
		"gas_price":   gasPrice,
		"sent_at":     time.Now(),
		"last_error":  lastError,
	})
}

func (r *Record) SetWaitingWithError(lastError string) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusWaiting,
		"last_error": lastError,
	})
}
