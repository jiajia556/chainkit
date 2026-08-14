package chainkitcollecttasks

import (
	"fmt"
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

type Record struct {
	*models.BaseRecord[*ChainCollectTasks]
}

const (
	StatusWaiting = iota
	StatusNeedGas
	StatusCanSend
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
		err := createTableSession.CreateTableIfNotExists(new(ChainCollectTasks))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainCollectTasks]{
			Session: dbSession,
			Model:   new(ChainCollectTasks),
		},
	}
	return r
}

func (r *Record) GetTaskingByAddressAndToken(userDepositAddressId, tokenId uint64) *Record {
	where := fmt.Sprintf(
		"user_deposit_address_id = %d AND token_id = %d AND status IN (%s)",
		userDepositAddressId,
		tokenId,
		"0,1,2,3,4",
	)
	r.DB().Where(where).Take(r.Model)
	return r
}

func (r *Record) SetStatus(status int8) {
	r.DB().Model(r.Model).Update("status", status)
}

func (r *Record) ClaimCanSend() (bool, error) {
	res := r.DB().
		Model(r.Model).
		Where("id = ? AND status = ?", r.Model.Id, StatusCanSend).
		Updates(map[string]interface{}{
			"status":     StatusSending,
			"last_error": "",
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *Record) SetSent(actualAmount decimal.Decimal, hash string, nonce uint64) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"actual_amount": actualAmount,
		"tx_hash":       hash,
		"nonce":         nonce,
		"status":        StatusSent,
		"sent_at":       time.Now(),
		"last_error":    "",
	})
}

func (r *Record) SetMaybeSent(actualAmount decimal.Decimal, hash string, nonce uint64, lastError string) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"actual_amount": actualAmount,
		"tx_hash":       hash,
		"nonce":         nonce,
		"status":        StatusMaybeSent,
		"sent_at":       time.Now(),
		"last_error":    lastError,
	})
}

func (r *Record) SetUnknown() {
	r.DB().Model(r.Model).Update("status", StatusUnknown)
}

func (r *Record) SetWaiting() {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status": StatusWaiting,
	})
}

func (r *Record) SetCanSendWithError(lastError string) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusCanSend,
		"last_error": lastError,
	})
}

func (r *Record) SetFailed() {
	r.DB().Model(r.Model).Update("status", StatusFailed)
}

func (r *Record) SetConfirmed(gasUsed, txFee decimal.Decimal) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":       StatusConfirmed,
		"gas_used":     gasUsed,
		"tx_fee":       txFee,
		"confirmed_at": time.Now(),
	})
}

func (r *Record) BatchSetNeedGas(ids []uint64, gasTaskId uint64) (int64, error) {
	res := r.DB().Table(r.Model.TableName()).Where("id IN (?) AND status = ?", ids, StatusWaiting).Updates(map[string]interface{}{
		"status":      StatusNeedGas,
		"gas_task_id": gasTaskId,
	})
	return res.RowsAffected, res.Error
}

func (r *Record) BatchSetCanSend(ids []uint64) error {
	return r.DB().Table(r.Model.TableName()).Where("id IN (?) AND status = ?", ids, StatusWaiting).Updates(map[string]interface{}{
		"status": StatusCanSend,
	}).Error
}

func (r *Record) SetWaitingByGasTaskId(gasTaskId uint64) {
	r.DB().Table(r.Model.TableName()).Where("gas_task_id = ?", gasTaskId).Update("status", StatusWaiting)
}

func (r *Record) SetCanSendByGasTaskId(gasTaskId uint64) {
	r.DB().Table(r.Model.TableName()).Where("gas_task_id = ?", gasTaskId).Update("status", StatusCanSend)
}

func (r *Record) GetPendingByAddressAndChain(userDepositAddressId, chainDbId uint64) *Record {
	r.DB().Where("user_deposit_address_id = ? AND chain_db_id = ? AND status IN (?)", userDepositAddressId, chainDbId, []int{StatusSending, StatusSent, StatusMaybeSent}).Take(r.Model)
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
