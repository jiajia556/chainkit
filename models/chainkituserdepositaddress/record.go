package chainkituserdepositaddress

import (
	"crypto/ecdsa"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/tool-box/cryptox"
	"github.com/jiajia556/tool-box/mysqlx"
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
		err := dbSession.CreateTableIfNotExists(new(ChainUserDepositAddress))
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

func (r *Record) ReadByAddress(address string) *Record {
	r.DB().Where("address = ?", strings.ToLower(address)).Take(&r.Model)
	return r
}

func (r *Record) GetPriKey(password string) (*ecdsa.PrivateKey, error) {
	priKeyBytes, err := cryptox.DecryptWithPassword(password, r.Model.PrivateKeyEncrypted)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(priKeyBytes)
}
