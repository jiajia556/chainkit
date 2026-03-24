package chainkitmnemonics

import (
	"crypto/ecdsa"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/mnemonic"
	"github.com/jiajia556/tool-box/cryptox"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainMnemonics]
	password string
}

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}
	if mysqlx.AutoCreateTable() {
		err := dbContext.CreateTableIfNotExists(new(ChainMnemonics))
		if err != nil {
			panic(err)
		}
	}
	r := &Record{
		BaseRecord: &models.BaseRecord[*ChainMnemonics]{
			Session: dbContext,
			Model:   new(ChainMnemonics),
		},
	}
	return r
}

func (r *Record) Create(password ...string) (err error) {
	pwd := r.password
	if len(password) > 0 {
		pwd = password[0]
	}
	r.Model.Words, err = cryptox.EncryptWithPassword(1, pwd, r.Model.Words)
	if err != nil {
		return err
	}
	return r.DB().Create(r.Model).Error
}

func (r *Record) SetPassword(password string) {
	r.password = password
}

func (r *Record) GetAddressAndPriKeyStringByIndex(index uint32, password ...string) (address string, priKey string, err error) {
	pwd := r.password
	if len(password) > 0 {
		pwd = password[0]
	}

	r.Model.Words, err = cryptox.DecryptWithPassword(pwd, r.Model.Words)
	if err != nil {
		return "", "", err
	}

	wallet, err := mnemonic.NewWallet(string(r.Model.Words), pwd)
	if err != nil {
		return "", "", err
	}

	return wallet.AddressAndPrivateKey(index)
}

func (r *Record) GetPrivateKey(index uint32, password ...string) (*ecdsa.PrivateKey, error) {
	pwd := r.password
	if len(password) > 0 {
		pwd = password[0]
	}

	var err error

	r.Model.Words, err = cryptox.DecryptWithPassword(pwd, r.Model.Words)
	if err != nil {
		return nil, err
	}

	wallet, err := mnemonic.NewWallet(string(r.Model.Words), pwd)
	if err != nil {
		return nil, err
	}

	return wallet.PrivateKey(index)
}
