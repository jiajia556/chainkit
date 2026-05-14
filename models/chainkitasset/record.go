package chainkitasset

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/models/chainkitassetrecord"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Record struct {
	*models.BaseRecord[*ChainAsset]
	// Record is not goroutine-safe. Do not reuse the same Record instance concurrently.
	// Err stores the last error occurred in chain-style methods that historically do not return error.
	// It is optional and does not affect existing callers.
	Err error
}

// NewRecord returns a Record bound to the given session (or a new transaction session).
// Parameters:
//   - session: optional MySQL session to use for queries and writes.
func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	// Never run DDL while in a transaction (MySQL DDL may cause implicit commit).
	if mysqlx.AutoCreateTable() && (dbSession == nil || !dbSession.IsInTransaction()) {
		err := dbSession.CreateTableIfNotExists(new(ChainAsset))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainAsset]{
			Session: dbSession,
			Model:   new(ChainAsset),
		},
	}
	return r
}

var (
	ErrAssetNotLoaded     = fmt.Errorf("asset not loaded; call GetByUserAndToken first")
	ErrDuplicateRequestID = errors.New("duplicate request_id")
	// ErrInsufficientAvailableBalance indicates the available balance is insufficient for the operation.
	// Use errors.Is(err, ErrInsufficientAvailableBalance) to check.
	ErrInsufficientAvailableBalance = errors.New("insufficient available balance")
	ErrTokenNotFound                = errors.New("token not found")
)

const amountDecimalType = "DECIMAL(36,0)"

// GetByUserAndToken loads the asset by user/token, creating it if it does not exist.
// Parameters:
//   - userId: user identifier.
//   - tokenId: token identifier.
func (r *Record) GetByUserAndToken(userId, tokenId uint64) *Record {
	_, err := r.GetByUserAndTokenE(userId, tokenId)
	r.Err = err
	return r
}

// GetByUserAndTokenE is the error-returning variant of GetByUserAndToken.
func (r *Record) GetByUserAndTokenE(userId, tokenId uint64) (*Record, error) {
	if r.Model == nil {
		r.Model = new(ChainAsset)
	}

	// 1) read
	if err := r.DB().Where("user_id = ? AND token_id = ?", userId, tokenId).Take(r.Model).Error; err == nil {
		return r, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return r, err
	}

	// 2) create (safe under concurrency thanks to uk_user_token)
	token := chainkittokens.NewRecord(r.Session)
	_ = token.Read(tokenId)
	if !token.Exists() {
		return r, ErrTokenNotFound
	}
	r.Model.UserId = userId
	r.Model.TokenId = tokenId
	r.Model.Symbol = token.Model.Symbol
	if err := r.Create(); err != nil {
		if isDuplicateKeyError(err) {
			// someone else created it; re-load
			if err2 := r.DB().Where("user_id = ? AND token_id = ?", userId, tokenId).Take(r.Model).Error; err2 != nil {
				return r, err2
			}
			return r, nil
		}
		return r, err
	}
	return r, nil
}

// IncreaseBalance increases available balance and writes a corresponding asset record.
// Parameters:
//   - amount: amount to add (must be > 0).
//   - bizType: business type for ledgering.
//   - bizId: business id for ledgering.
//   - requestId: idempotency key for the asset record.
//   - remark: remark for the asset record.
func (r *Record) IncreaseBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	if err := validatePositiveIntegerAmount(amount); err != nil {
		return err
	}
	return r.changeBalance(amount, decimal.Zero, nil, nil, bizType, bizId, requestId, remark, fmt.Errorf("insufficient balance"))
}

// DecreaseBalance decreases available balance and writes a corresponding asset record.
// Parameters:
//   - amount: amount to subtract (must be > 0).
//   - bizType: business type for ledgering.
//   - bizId: business id for ledgering.
//   - requestId: idempotency key for the asset record.
//   - remark: remark for the asset record.
func (r *Record) DecreaseBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	if err := validatePositiveIntegerAmount(amount); err != nil {
		return err
	}
	minAvailable := amount
	return r.changeBalance(amount.Neg(), decimal.Zero, &minAvailable, nil, bizType, bizId, requestId, remark, ErrInsufficientAvailableBalance)
}

// FreezeBalance moves funds from available to frozen balance and writes a corresponding asset record.
// Parameters:
//   - amount: amount to freeze (must be > 0).
//   - bizType: business type for ledgering.
//   - bizId: business id for ledgering.
//   - requestId: idempotency key for the asset record.
//   - remark: remark for the asset record.
func (r *Record) FreezeBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	if err := validatePositiveIntegerAmount(amount); err != nil {
		return err
	}
	minAvailable := amount
	return r.changeBalance(amount.Neg(), amount, &minAvailable, nil, bizType, bizId, requestId, remark, ErrInsufficientAvailableBalance)
}

// UnfreezeBalance moves funds from frozen back to available balance and writes a corresponding asset record.
// Parameters:
//   - amount: amount to unfreeze (must be > 0).
//   - bizType: business type for ledgering.
//   - bizId: business id for ledgering.
//   - requestId: idempotency key for the asset record.
//   - remark: remark for the asset record.
func (r *Record) UnfreezeBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	if err := validatePositiveIntegerAmount(amount); err != nil {
		return err
	}
	minFrozen := amount
	return r.changeBalance(amount, amount.Neg(), nil, &minFrozen, bizType, bizId, requestId, remark, fmt.Errorf("insufficient frozen balance"))
}

// IncreaseFrozenBalance increases frozen balance only and writes a corresponding asset record.
// NOTE: This increases total asset. It does NOT move available to frozen.
// Parameters:
//   - amount: amount to add (must be > 0).
//   - bizType: business type for ledgering.
//   - bizId: business id for ledgering.
//   - requestId: idempotency key for the asset record.
//   - remark: remark for the asset record.
func (r *Record) IncreaseFrozenBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	if err := validatePositiveIntegerAmount(amount); err != nil {
		return err
	}
	return r.changeBalance(decimal.Zero, amount, nil, nil, bizType, bizId, requestId, remark, fmt.Errorf("insufficient balance"))
}

// DecreaseFrozenBalance decreases frozen balance only and writes a corresponding asset record.
// Parameters:
//   - amount: amount to subtract (must be > 0).
//   - bizType: business type for ledgering.
//   - bizId: business id for ledgering.
//   - requestId: idempotency key for the asset record.
//   - remark: remark for the asset record.
func (r *Record) DecreaseFrozenBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	if err := validatePositiveIntegerAmount(amount); err != nil {
		return err
	}
	minFrozen := amount
	return r.changeBalance(decimal.Zero, amount.Neg(), nil, &minFrozen, bizType, bizId, requestId, remark, fmt.Errorf("insufficient frozen balance"))
}

// --- auto request_id helpers (NOT retryable) ---

func (r *Record) IncreaseBalanceWithAutoRequestID(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.IncreaseBalance(amount, bizType, bizId, uuid.NewString(), remark)
}

func (r *Record) DecreaseBalanceWithAutoRequestID(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.DecreaseBalance(amount, bizType, bizId, uuid.NewString(), remark)
}

func (r *Record) FreezeBalanceWithAutoRequestID(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.FreezeBalance(amount, bizType, bizId, uuid.NewString(), remark)
}

func (r *Record) UnfreezeBalanceWithAutoRequestID(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.UnfreezeBalance(amount, bizType, bizId, uuid.NewString(), remark)
}

func (r *Record) IncreaseFrozenBalanceWithAutoRequestID(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.IncreaseFrozenBalance(amount, bizType, bizId, uuid.NewString(), remark)
}

func (r *Record) DecreaseFrozenBalanceWithAutoRequestID(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.DecreaseFrozenBalance(amount, bizType, bizId, uuid.NewString(), remark)
}

// Deprecated: use IncreaseBalanceWithAutoRequestID.
func (r *Record) IncreaseBalanceWithoutIdempotency(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.IncreaseBalanceWithAutoRequestID(amount, bizType, bizId, remark)
}

// Deprecated: use DecreaseBalanceWithAutoRequestID.
func (r *Record) DecreaseBalanceWithoutIdempotency(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.DecreaseBalanceWithAutoRequestID(amount, bizType, bizId, remark)
}

// Deprecated: use FreezeBalanceWithAutoRequestID.
func (r *Record) FreezeBalanceWithoutIdempotency(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.FreezeBalanceWithAutoRequestID(amount, bizType, bizId, remark)
}

// Deprecated: use UnfreezeBalanceWithAutoRequestID.
func (r *Record) UnfreezeBalanceWithoutIdempotency(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.UnfreezeBalanceWithAutoRequestID(amount, bizType, bizId, remark)
}

// Deprecated: use IncreaseFrozenBalanceWithAutoRequestID.
func (r *Record) IncreaseFrozenBalanceWithoutIdempotency(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.IncreaseFrozenBalanceWithAutoRequestID(amount, bizType, bizId, remark)
}

// Deprecated: use DecreaseFrozenBalanceWithAutoRequestID.
func (r *Record) DecreaseFrozenBalanceWithoutIdempotency(amount decimal.Decimal, bizType string, bizId uint64, remark string) error {
	return r.DecreaseFrozenBalanceWithAutoRequestID(amount, bizType, bizId, remark)
}

func (r *Record) changeBalance(availableDelta, frozenDelta decimal.Decimal, minAvailable, minFrozen *decimal.Decimal, bizType string, bizId uint64, requestId string, remark string, insufficientErr error) error {
	if err := r.ensureLoaded(); err != nil {
		return err
	}
	var err error
	requestId, err = normalizeRequestIDRequired(requestId)
	if err != nil {
		return err
	}
	return r.withTransaction(func(txRecord *Record) error {
		// Lock the asset row to make the whole operation serializable per asset.
		if err := txRecord.lockForUpdate(); err != nil {
			return err
		}

		availableBefore := txRecord.Model.Available
		frozenBefore := txRecord.Model.Frozen

		// request_id is treated as a strict unique key.
		// If request_id already exists, return duplicate request error.
		assetRecordId, err := txRecord.tryCreateAssetRecordPlaceholder(
			availableDelta,
			frozenDelta,
			availableBefore,
			frozenBefore,
			bizType,
			bizId,
			requestId,
			remark,
		)
		if err != nil {
			return err
		}

		updatedAt := time.Now()
		if err := txRecord.updateBalanceLocked(availableDelta, frozenDelta, minAvailable, minFrozen, updatedAt, insufficientErr); err != nil {
			return err
		}
		if err := txRecord.reload(); err != nil {
			return err
		}
		return txRecord.updateAssetRecordAfter(assetRecordId, txRecord.Model.Available, txRecord.Model.Frozen)
	})
}

func (r *Record) updateBalanceLocked(availableDelta, frozenDelta decimal.Decimal, minAvailable, minFrozen *decimal.Decimal, updatedAt time.Time, noRowsErr error) error {
	// Use CAST to ensure the parameters are treated as numeric DECIMAL in MySQL even if the driver
	// sends them as strings (decimal.Decimal implements driver.Valuer and returns string).
	query := fmt.Sprintf(
		"UPDATE `chain_asset` SET `available` = `available` + CAST(? AS %s), `frozen` = `frozen` + CAST(? AS %s), `updated_at` = ?, `version` = `version` + 1 WHERE `id` = ?",
		amountDecimalType,
		amountDecimalType,
	)
	args := []interface{}{availableDelta, frozenDelta, updatedAt, r.Model.Id}
	if minAvailable != nil {
		query += fmt.Sprintf(" AND `available` >= CAST(? AS %s)", amountDecimalType)
		args = append(args, *minAvailable)
	}
	if minFrozen != nil {
		query += fmt.Sprintf(" AND `frozen` >= CAST(? AS %s)", amountDecimalType)
		args = append(args, *minFrozen)
	}
	m := r.DB().Debug().Exec(query, args...)
	if m.Error != nil {
		return m.Error
	}
	if m.RowsAffected == 0 {
		return noRowsErr
	}
	return nil
}

func (r *Record) reload() error {
	return r.DB().Where("id = ?", r.Model.Id).Take(r.Model).Error
}

func (r *Record) ensureLoaded() error {
	if r.Err != nil {
		return r.Err
	}
	if r.Model == nil || r.Model.Id == 0 {
		return ErrAssetNotLoaded
	}
	return nil
}

func (r *Record) lockForUpdate() error {
	// Requires to be called in an open transaction.
	return r.DB().Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", r.Model.Id).Take(r.Model).Error
}

func (r *Record) withTransaction(fn func(txRecord *Record) error) error {
	if r.Session != nil && r.Session.IsInTransaction() {
		txRecord := NewRecord(r.Session)
		orig := r.Model
		if orig != nil {
			m := *orig
			txRecord.Model = &m
		}
		err := fn(txRecord)
		if err == nil {
			if orig != nil && txRecord.Model != nil {
				*orig = *txRecord.Model
			} else if orig == nil {
				r.Model = txRecord.Model
			}
		}
		return err
	}

	tx := mysqlx.NewTxSession()
	if err := tx.Begin(); err != nil {
		return err
	}
	defer func() {
		recovered := recover()
		if recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
	}()

	txRecord := NewRecord(tx)
	orig := r.Model
	if orig != nil {
		m := *orig
		txRecord.Model = &m
	}
	if err := fn(txRecord); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback failed: %v", err, rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if orig != nil && txRecord.Model != nil {
		*orig = *txRecord.Model
	} else if orig == nil {
		r.Model = txRecord.Model
	}
	return nil
}

func (r *Record) tryCreateAssetRecordPlaceholder(availableChange, frozenChange, availableBefore, frozenBefore decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) (recordId uint64, err error) {
	record := chainkitassetrecord.NewRecord(r.Session)
	record.Model.UserId = r.Model.UserId
	record.Model.TokenId = r.Model.TokenId
	record.Model.Symbol = r.Model.Symbol
	record.Model.BizType = bizType
	record.Model.BizId = bizId
	record.Model.RequestId = requestId
	record.Model.AvailableChange = availableChange
	record.Model.FrozenChange = frozenChange
	// placeholder values, will be updated after asset balance update
	record.Model.AvailableBefore = availableBefore
	record.Model.AvailableAfter = availableBefore
	record.Model.FrozenBefore = frozenBefore
	record.Model.FrozenAfter = frozenBefore
	record.Model.Remark = remark
	if err := record.Create(); err != nil {
		if isDuplicateKeyError(err) {
			// request_id is treated as a strict unique key: any duplicate is an error.
			return 0, fmt.Errorf("%w: %s", ErrDuplicateRequestID, requestId)
		}
		return 0, err
	}
	return record.Model.Id, nil
}

func (r *Record) updateAssetRecordAfter(recordId uint64, availableAfter, frozenAfter decimal.Decimal) error {
	if recordId == 0 {
		return nil
	}
	updates := map[string]interface{}{
		"available_after": availableAfter,
		"frozen_after":    frozenAfter,
	}
	return r.DB().Table("chain_asset_record").Where("id = ?", recordId).Updates(updates).Error
}

func validatePositiveIntegerAmount(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than 0")
	}
	// chainkit uses decimal(36,0) (integer in smallest unit). Reject fractional input early.
	if !amount.Equal(amount.Truncate(0)) {
		return fmt.Errorf("amount must be an integer in the smallest unit")
	}
	return nil
}

func normalizeRequestIDRequired(requestId string) (string, error) {
	requestId = strings.TrimSpace(requestId)
	if requestId == "" {
		return "", fmt.Errorf("request_id is required")
	}
	return requestId, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	// last resort
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "duplicate")
}
