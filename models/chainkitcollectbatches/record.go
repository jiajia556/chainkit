package chainkitcollectbatches

import (
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainCollectBatches]
}

const (
	StatusWaiting = iota
	StatusBuilding
	StatusSent
	StatusConfirmed
	StatusFailed
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
		if err := createTableSession.CreateTableIfNotExists(new(ChainCollectBatches)); err != nil {
			panic(err)
		}
	}
	return &Record{BaseRecord: &models.BaseRecord[*ChainCollectBatches]{
		Session: dbSession,
		Model:   new(ChainCollectBatches),
	}}
}

func (r *Record) ClaimWaiting() (bool, error) {
	result := r.DB().Model(r.Model).
		Where("id = ? AND status = ?", r.Model.Id, StatusWaiting).
		Updates(map[string]interface{}{"status": StatusBuilding, "last_error": ""})
	return result.RowsAffected == 1, result.Error
}

// RefreshSponsorNonce updates an unbroadcast batch after it has been claimed.
// Once raw_tx exists the nonce is immutable because changing it would create a
// different signed transaction and make recovery ambiguous.
func (r *Record) RefreshSponsorNonce(nonce uint64) (bool, error) {
	result := r.DB().Model(r.Model).
		Where(
			"id = ? AND status = ? AND (raw_tx IS NULL OR LENGTH(raw_tx) = 0)",
			r.Model.Id,
			StatusBuilding,
		).
		Update("sponsor_nonce", nonce)
	if result.Error == nil && result.RowsAffected == 1 {
		r.Model.SponsorNonce = nonce
	}
	return result.RowsAffected == 1, result.Error
}

func (r *Record) GetInFlightBySponsor(chainDbId uint64, sponsorAddress string) *Record {
	r.DB().Where(
		"chain_db_id = ? AND sponsor_address = ? AND status IN (?)",
		chainDbId,
		sponsorAddress,
		[]int{StatusBuilding, StatusSent, StatusMaybeSent},
	).Order("sponsor_nonce ASC, id ASC").Take(r.Model)
	return r
}

func (r *Record) SetSent(hash string, rawTx []byte) (bool, error) {
	result := r.DB().Model(r.Model).
		Where("id = ? AND status IN (?)", r.Model.Id, []int{StatusBuilding, StatusMaybeSent}).
		Updates(map[string]interface{}{
			"status":     StatusSent,
			"tx_hash":    hash,
			"raw_tx":     rawTx,
			"sent_at":    time.Now(),
			"last_error": "",
		})
	return result.RowsAffected == 1, result.Error
}

// SetBuilt persists the exact signed transaction before broadcast. This makes a
// process crash recoverable without signing a different transaction at the same
// sponsor nonce.
func (r *Record) SetBuilt(
	hash string,
	rawTx []byte,
	gasLimit uint64,
	maxFeePerGas string,
	maxPriorityFeePerGas string,
	authorizationCount uint32,
) (bool, error) {
	result := r.DB().Model(r.Model).
		Where("id = ? AND status = ?", r.Model.Id, StatusBuilding).
		Updates(map[string]interface{}{
			"status":                   StatusMaybeSent,
			"tx_hash":                  hash,
			"raw_tx":                   rawTx,
			"gas_limit":                gasLimit,
			"max_fee_per_gas":          maxFeePerGas,
			"max_priority_fee_per_gas": maxPriorityFeePerGas,
			"authorization_count":      authorizationCount,
			"sent_at":                  time.Now(),
			"last_error":               "",
		})
	return result.RowsAffected == 1, result.Error
}

func (r *Record) SetMaybeSent(hash string, rawTx []byte, lastError string) (bool, error) {
	result := r.DB().Model(r.Model).
		Where("id = ? AND status IN (?)", r.Model.Id, []int{StatusBuilding, StatusMaybeSent}).
		Updates(map[string]interface{}{
			"status":     StatusMaybeSent,
			"tx_hash":    hash,
			"raw_tx":     rawTx,
			"sent_at":    time.Now(),
			"last_error": lastError,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *Record) SetWaitingWithError(lastError string) error {
	return r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusWaiting,
		"last_error": lastError,
	}).Error
}

func (r *Record) SetConfirmed(gasUsed uint64, txFee string) error {
	return r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":       StatusConfirmed,
		"gas_used":     gasUsed,
		"tx_fee":       txFee,
		"confirmed_at": time.Now(),
		"last_error":   "",
	}).Error
}

func (r *Record) SetFailed(lastError string) error {
	return r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusFailed,
		"last_error": lastError,
	}).Error
}

func (r *Record) SetUnknown(lastError string) error {
	return r.DB().Model(r.Model).Updates(map[string]interface{}{
		"status":     StatusUnknown,
		"last_error": lastError,
	}).Error
}

func (r *Record) SinceSent() time.Duration {
	if r.Model.SentAt.IsZero() {
		return time.Since(r.Model.CreatedAt)
	}
	return time.Since(r.Model.SentAt)
}
