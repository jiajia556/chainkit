package chainkituserdepositaddress

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/utils"
	"github.com/jiajia556/tool-box/cryptox"
	"github.com/jiajia556/tool-box/locker"
	"github.com/jiajia556/tool-box/mysqlx"
	"gorm.io/gorm"
)

type Record struct {
	*models.BaseRecord[*ChainUserDepositAddress]
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
		err := createTableSession.CreateTableIfNotExists(new(ChainUserDepositAddress))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainUserDepositAddress]{
			Session: dbSession,
			Model:   new(ChainUserDepositAddress),
		},
	}
	return r
}

func (r *Record) BatchCreate(count int, password, remark string) (createdCount int, err error) {
	for i := 0; i < count; i++ {
		priKey, addr, genErr := utils.NewKey()
		if genErr != nil {
			err = genErr
			return
		}
		priKeyByte := crypto.FromECDSA(priKey)
		enPriKey, encErr := cryptox.EncryptWithPassword(1, password, priKeyByte)
		if encErr != nil {
			err = encErr
			return
		}
		record := NewRecord()
		record.Model.Address = addr.Hex()
		record.Model.PrivateKeyEncrypted = enPriKey
		record.Model.Remark = remark
		if createErr := record.Create(); createErr != nil {
			err = createErr
			return
		}
		createdCount++
	}
	return
}

func (r *Record) ReadByAddress(address string) *Record {
	r.DB().Where("address = ?", strings.ToLower(address)).Take(r.Model)
	return r
}

func (r *Record) GetPriKey(password string) (*ecdsa.PrivateKey, error) {
	priKeyBytes, err := cryptox.DecryptWithPassword(password, r.Model.PrivateKeyEncrypted)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(priKeyBytes)
}

func (r *Record) GetAddress() string {
	return r.Model.Address
}

func (r *Record) GetByUserId(userId uint64) (res *Record, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("GetByUserId:%d", userId)
	instance, err := locker.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() {
		if unlockErr := instance.Unlock(ctx); unlockErr != nil && err == nil {
			err = unlockErr
		}
	}()

	err = r.DB().Where("user_id = ?", userId).Take(r.Model).Error
	if err == nil && r.Exists() {
		return r, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tx := r.DB().Exec("UPDATE `chain_user_deposit_address` SET `user_id` = ? WHERE `user_id` = 0 LIMIT 1", userId)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, errors.New("no available deposit address")
	}
	if err = r.DB().Where("user_id = ?", userId).Take(r.Model).Error; err != nil {
		return nil, err
	}

	return r, nil
}
