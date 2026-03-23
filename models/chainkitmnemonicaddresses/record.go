package chainkitmnemonicaddresses

import (
	"crypto/ecdsa"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/models/chainkitmnemonics"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainMnemonicAddresses]
}

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainMnemonicAddresses))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainMnemonicAddresses]{
			Session: dbContext,
			Model:   new(ChainMnemonicAddresses),
		},
	}
	return r
}

func (r *Record) GetPriKeyString(password string) (string, error) {
	wallet := chainkitmnemonics.NewRecord()
	err := wallet.Read(r.Model.MnemonicId)
	if err != nil {
		return "", err
	}
	_, priKey, err := wallet.GetAddressAndPriKeyStringByIndex(r.Model.Index, password)
	if err != nil {
		return "", err
	}
	return priKey, nil
}

func (r *Record) GetPriKey(password string) (*ecdsa.PrivateKey, error) {
	wallet := chainkitmnemonics.NewRecord()
	err := wallet.Read(r.Model.MnemonicId)
	if err != nil {
		return nil, err
	}

	return wallet.GetPrivateKey(r.Model.Index, password)
}

func (r *Record) ReadLastByMnemonicId(mnemonicId uint64) *Record {
	r.DB().Where("mnemonic_id = ?", mnemonicId).Order("index desc").Take(&r.Model)
	return r
}

func (r *Record) BatchCreate(mnemonicId uint64, password string, count int) (int, error) {
	mn := chainkitmnemonics.NewRecord()
	err := mn.Read(mnemonicId)
	if err != nil {
		return 0, err
	}
	mn.SetPassword(password)
	last := r.ReadLastByMnemonicId(mnemonicId)
	lastIndex := last.Model.Index
	successCount := 0
	for index := lastIndex + 1; index <= lastIndex+uint32(count); index++ {
		address, _, err := mn.GetAddressAndPriKeyStringByIndex(index)
		if err != nil {
			continue
		}
		record := NewRecord(r.Session)
		record.Model.Address = address
		record.Model.Index = index
		record.Model.MnemonicId = mnemonicId
		err = record.Create()
		if err != nil {
			continue
		}
		successCount++
	}
	return successCount, nil
}
