package chainkitasset

import (
	"fmt"
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/models/chainkitassetrecord"
	"github.com/jiajia556/tool-box/mysqlx"
	"github.com/shopspring/decimal"
)

type Record struct {
	*models.BaseRecord[*ChainAsset]
}

func NewRecord(session ...mysqlx.Session) *Record {
	var dbSession mysqlx.Session
	if len(session) > 0 {
		dbSession = session[0]
	} else {
		dbSession = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
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

func (r *Record) ReadByUserAndToken(userId, tokenId uint64) *Record {
	r.DB().Where("user_id = ? AND token_id = ?", userId, tokenId).Take(&r.Model)
	if !r.Exists() {
		r.Model.UserId = userId
		r.Model.TokenId = tokenId
		err := r.Create()
		if err != nil {
			r.DB().Where("user_id = ? AND token_id = ?", userId, tokenId).Take(&r.Model)
		}
	}
	return r
}

func (r *Record) IncreaseBalance(amount decimal.Decimal, bizType string, bizId uint64, requestId string, remark string) error {
	sql := fmt.Sprintf(
		"UPDATE `chain_asset` SET `available` = `available` + %s, `updated_at` = %s, `version` = `version` + 1 WHERE `id` = %d AND `version` = %d;",
		amount.String(), time.Now().Format("2006-01-02 15:04:05"), r.Model.Id, r.Model.Version,
	)
	m := r.DB().Exec(sql)
	if m.Error != nil {
		return m.Error
	}
	if m.RowsAffected == 0 {
		return fmt.Errorf("version mismatch")
	}
	record := chainkitassetrecord.NewRecord(r.Session)
	record.Model.UserId = r.Model.UserId
	record.Model.TokenId = r.Model.TokenId
	record.Model.BizType = bizType
	record.Model.BizId = bizId
	record.Model.RequestId = requestId
	record.Model.AvailableChange = amount
	record.Model.AvailableBefore = r.Model.Available
	record.Model.AvailableAfter = r.Model.Available.Add(amount)
	record.Model.Remark = remark
	return record.Create()
}
