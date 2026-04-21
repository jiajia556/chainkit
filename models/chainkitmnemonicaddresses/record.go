package chainkitmnemonicaddresses

import (
	"crypto/ecdsa"
	"fmt"

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
		if err := dbContext.CreateTableIfNotExists(new(ChainMnemonicAddresses)); err != nil {
			panic(err)
		}
	}

	return &Record{
		BaseRecord: &models.BaseRecord[*ChainMnemonicAddresses]{
			Session: dbContext,
			Model:   new(ChainMnemonicAddresses),
		},
	}
}

//
// ================== 私钥获取 ==================
//

func (r *Record) GetPriKey(password string) (*ecdsa.PrivateKey, error) {
	mn := chainkitmnemonics.NewRecord(r.Session)
	if err := mn.Read(r.Model.MnemonicId); err != nil {
		return nil, err
	}

	session, err := mn.NewSession(password)
	if err != nil {
		return nil, err
	}

	return session.GetPrivateKey(r.Model.Index)
}

func (r *Record) GetPriKeyString(password string) (string, error) {

	mn := chainkitmnemonics.NewRecord(r.Session)
	if err := mn.Read(r.Model.MnemonicId); err != nil {
		return "", err
	}

	session, err := mn.NewSession(password)
	if err != nil {
		return "", err
	}

	_, priKey, err := session.GetAddressAndPrivateKey(r.Model.Index)
	if err != nil {
		return "", err
	}

	return priKey, nil
}

//
// ================== 查询 ==================
//

func (r *Record) ReadLastByMnemonicId(mnemonicId uint64) (*Record, error) {
	err := r.DB().
		Where("mnemonic_id = ?", mnemonicId).
		Order("`index` desc").
		Take(&r.Model).Error
	if err != nil {
		return nil, err
	}
	return r, nil
}

//
// ================== 批量生成 ==================
//

func (r *Record) BatchCreate(
	mnemonicId uint64,
	password string,
	count int,
	remark string,
) (int, error) {

	if count <= 0 {
		return 0, fmt.Errorf("count must > 0")
	}

	mn := chainkitmnemonics.NewRecord()
	if err := mn.Read(mnemonicId); err != nil {
		return 0, err
	}

	session, err := mn.NewSession(password)
	if err != nil {
		return 0, err
	}

	last := NewRecord()
	lastRecord, err := last.ReadLastByMnemonicId(mnemonicId)

	var startIndex uint32 = 0
	if err == nil {
		startIndex = lastRecord.Model.Index + 1
	}

	addrs, err := session.BatchAddresses(startIndex, uint32(count))
	if err != nil {
		return 0, err
	}

	success := 0

	for i, addr := range addrs {

		index := startIndex + uint32(i)

		record := NewRecord()
		record.Model.MnemonicId = mnemonicId
		record.Model.Index = index
		record.Model.Address = addr
		record.Model.Remark = remark

		if err := record.Create(); err != nil {
			continue
		}

		success++
	}

	return success, nil
}

//
// ================== 创建助记词 + 地址 ==================
//

func (r *Record) CreateNewMnemonicAndAddress(
	password string,
	remark string,
) (*Record, error) {
	mn := chainkitmnemonics.NewRecord()
	if err := mn.CreateNewMnemonic(password, remark); err != nil {
		return nil, err
	}

	session, err := mn.NewSession(password)
	if err != nil {
		return nil, err
	}

	addr, err := session.GetAddress(0)
	if err != nil {
		return nil, err
	}

	r.Model.MnemonicId = mn.Model.Id
	r.Model.Index = 0
	r.Model.Address = addr
	r.Model.Remark = remark

	if err := r.Create(); err != nil {
		return nil, err
	}

	return r, nil
}
