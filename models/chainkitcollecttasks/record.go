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

func (r *Record) SetSent(actualAmount decimal.Decimal, hash string, nonce uint64) {
	r.DB().Model(r.Model).Updates(map[string]interface{}{
		"actual_amount": actualAmount,
		"tx_hash":       hash,
		"nonce":         nonce,
		"status":        StatusSent,
		"sent_at":       time.Now(),
	})
}

func (r *Record) SetUnknown() {
	r.DB().Model(r.Model).Update("status", StatusUnknown)
}

func (r *Record) SetWaiting() {
	r.DB().Model(r.Model).Update("status", StatusWaiting)
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

func (r *Record) BatchSetNeedGas(ids []uint64, gasTaskId uint64) {
	r.DB().Table(r.Model.TableName()).Where("id IN (?)", ids).Updates(map[string]interface{}{
		"status":      StatusNeedGas,
		"gas_task_id": gasTaskId,
	})
}

func (r *Record) BatchSetCanSend(ids []uint64) {
	r.DB().Table(r.Model.TableName()).Where("id IN (?)", ids).Updates(map[string]interface{}{
		"status": StatusCanSend,
	})
}

func (r *Record) SetWaitingByGasTaskId(gasTaskId uint64) {
	r.DB().Table(r.Model.TableName()).Where("gas_task_id = ?", gasTaskId).Update("status", StatusWaiting)
}

func (r *Record) SetCanSendByGasTaskId(gasTaskId uint64) {
	r.DB().Table(r.Model.TableName()).Where("gas_task_id = ?", gasTaskId).Update("status", StatusCanSend)
}

func (r *Record) GetPendingByAddressAndChain(userDepositAddressId, chainDbId uint64) *Record {
	r.DB().Where("user_deposit_address_id = ? AND chain_db_id = ? AND status IN (3,4)", userDepositAddressId, chainDbId).Take(r.Model)
	return r
}

func (r *Record) SinceCreated() time.Duration {
	return time.Since(r.Model.CreatedAt)
}
