package chainkitmnemonics

import (
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/jiajia556/chainkit/models"
	"github.com/jiajia556/chainkit/pkg/mnemonic"
	"github.com/jiajia556/tool-box/cryptox"
	"github.com/jiajia556/tool-box/mysqlx"
)

type Record struct {
	*models.BaseRecord[*ChainMnemonics]

	// seed缓存（短生命周期）
	seedCache []byte
	seedExp   time.Time

	mu sync.Mutex
}

// ================== 构造 ==================

func NewRecord(ctx ...mysqlx.Session) *Record {
	var dbContext mysqlx.Session
	if len(ctx) > 0 {
		dbContext = ctx[0]
	} else {
		dbContext = mysqlx.NewTxSession()
	}

	if mysqlx.AutoCreateTable() {
		if err := dbContext.CreateTableIfNotExists(new(ChainMnemonics)); err != nil {
			panic(err)
		}
	}

	return &Record{
		BaseRecord: &models.BaseRecord[*ChainMnemonics]{
			Session: dbContext,
			Model:   new(ChainMnemonics),
		},
	}
}

// ================== 创建 ==================

// 创建新助记词（加密存储）
func (r *Record) CreateNewMnemonic(password, remark string) error {
	words, err := mnemonic.NewWords()
	if err != nil {
		return err
	}

	cipher, err := cryptox.EncryptWithPassword(1, password, []byte(words))
	if err != nil {
		return err
	}

	r.Model.Words = cipher
	r.Model.Remark = remark

	return r.DB().Create(r.Model).Error
}

// ================== seed 管理（核心） ==================

func (r *Record) getSeed(password string) ([]byte, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	// 命中缓存
	if r.seedCache != nil && time.Now().Before(r.seedExp) {
		return r.seedCache, nil
	}

	// 解密助记词
	words, err := cryptox.DecryptWithPassword(password, r.Model.Words)
	if err != nil {
		return nil, fmt.Errorf("decrypt mnemonic failed: %w", err)
	}
	defer zeroBytes(words)

	// 生成 seed（PBKDF2）
	seed, err := mnemonicSeed(words)
	if err != nil {
		return nil, err
	}

	// 缓存（短期）
	r.seedCache = seed
	r.seedExp = time.Now().Add(10 * time.Second)

	return seed, nil
}

// 从助记词生成 seed（避免外部重复写）
func mnemonicSeed(words []byte) ([]byte, error) {
	return mnemonic.NewSeedFromMnemonic(string(words), "")
}

// ================== 单个派生 ==================

func (r *Record) GetPrivateKey(index uint32, password string) (*ecdsa.PrivateKey, error) {

	seed, err := r.getSeed(password)
	if err != nil {
		return nil, err
	}

	wallet, err := mnemonic.NewWalletFromSeed(seed)
	if err != nil {
		return nil, err
	}

	return wallet.PrivateKey(index)
}

func (r *Record) GetAddress(index uint32, password string) (string, error) {

	seed, err := r.getSeed(password)
	if err != nil {
		return "", err
	}

	wallet, err := mnemonic.NewWalletFromSeed(seed)
	if err != nil {
		return "", err
	}

	return wallet.AddressHex(index)
}

// ================== 批量派生（重点） ==================

func (r *Record) BatchAddresses(start uint32, count uint32, password string) ([]string, error) {

	seed, err := r.getSeed(password)
	if err != nil {
		return nil, err
	}

	wallet, err := mnemonic.NewWalletFromSeed(seed)
	if err != nil {
		return nil, err
	}

	return wallet.BatchAddresses(start, count)
}

// ================== 高级：Session（推荐） ==================

type Session struct {
	wallet *mnemonic.Wallet
}

// 创建短生命周期 session（用于批量操作）
func (r *Record) NewSession(password string) (*Session, error) {

	seed, err := r.getSeed(password)
	if err != nil {
		return nil, err
	}

	wallet, err := mnemonic.NewWalletFromSeed(seed)
	if err != nil {
		return nil, err
	}

	return &Session{
		wallet: wallet,
	}, nil
}

func (s *Session) GetPrivateKey(index uint32) (*ecdsa.PrivateKey, error) {
	return s.wallet.PrivateKey(index)
}

func (s *Session) GetAddress(index uint32) (string, error) {
	return s.wallet.AddressHex(index)
}

func (s *Session) BatchAddresses(start uint32, count uint32) ([]string, error) {
	return s.wallet.BatchAddresses(start, count)
}

func (s *Session) GetAddressAndPrivateKey(index uint32) (string, string, error) {
	return s.wallet.AddressAndPrivateKey(index)
}

// ================== 安全工具 ==================

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
