package mnemonic

import (
	"crypto/ecdsa"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

const HardenedOffset = 0x80000000

// m/44'/60'/0'/0
var defaultPath = []uint32{
	44 + HardenedOffset,
	60 + HardenedOffset,
	0 + HardenedOffset,
	0,
}

type Wallet struct {
	accountKey *bip32.Key

	// child key cache（只缓存 bip32.Key，不缓存私钥）
	cache map[uint32]*bip32.Key
	mu    sync.RWMutex
}

// ================== 创建 ==================

// NewWords 生成助记词
func NewWords() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	return bip39.NewMnemonic(entropy)
}

// 从 mnemonic 创建
func NewWallet(mnemonic string, passphrase string) (*Wallet, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, passphrase)
	return NewWalletFromSeed(seed)
}

// 推荐：直接用 seed（避免重复 PBKDF2）
func NewWalletFromSeed(seed []byte) (*Wallet, error) {

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	key := masterKey

	for _, p := range defaultPath {
		key, err = key.NewChildKey(p)
		if err != nil {
			return nil, err
		}
	}

	return &Wallet{
		accountKey: key,
		cache:      make(map[uint32]*bip32.Key),
	}, nil
}

func NewSeedFromMnemonic(words string, passphrase string) ([]byte, error) {
	if !bip39.IsMnemonicValid(words) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	return bip39.NewSeed(words, passphrase), nil
}

// ================== 核心派生 ==================

func (w *Wallet) deriveKey(index uint32) (*bip32.Key, error) {

	// 先读缓存
	w.mu.RLock()
	if k, ok := w.cache[index]; ok {
		w.mu.RUnlock()
		return k, nil
	}
	w.mu.RUnlock()

	// 写缓存
	w.mu.Lock()
	defer w.mu.Unlock()

	// double check
	if k, ok := w.cache[index]; ok {
		return k, nil
	}

	child, err := w.accountKey.NewChildKey(index)
	if err != nil {
		return nil, err
	}

	w.cache[index] = child
	return child, nil
}

// ================== 对外接口 ==================

func (w *Wallet) PrivateKey(index uint32) (*ecdsa.PrivateKey, error) {

	child, err := w.deriveKey(index)
	if err != nil {
		return nil, err
	}

	return crypto.ToECDSA(child.Key)
}

func (w *Wallet) Address(index uint32) (common.Address, error) {

	child, err := w.deriveKey(index)
	if err != nil {
		return common.Address{}, err
	}

	priv, err := crypto.ToECDSA(child.Key)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(priv.PublicKey), nil
}

func (w *Wallet) AddressHex(index uint32) (string, error) {

	addr, err := w.Address(index)
	if err != nil {
		return "", err
	}

	return addr.Hex(), nil
}

func (w *Wallet) PrivateKeyHex(index uint32) (string, error) {

	priv, err := w.PrivateKey(index)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(crypto.FromECDSA(priv)), nil
}

func (w *Wallet) AddressAndPrivateKey(index uint32) (string, string, error) {

	child, err := w.deriveKey(index)
	if err != nil {
		return "", "", err
	}

	priv, err := crypto.ToECDSA(child.Key)
	if err != nil {
		return "", "", err
	}

	addr := crypto.PubkeyToAddress(priv.PublicKey)

	return addr.Hex(), hexutil.Encode(crypto.FromECDSA(priv)), nil
}

// ================== 批量接口（重点优化） ==================

func (w *Wallet) BatchAddresses(start uint32, count uint32) ([]string, error) {

	addrs := make([]string, 0, count)

	for i := uint32(0); i < count; i++ {

		child, err := w.deriveKey(start + i)
		if err != nil {
			return nil, err
		}

		priv, err := crypto.ToECDSA(child.Key)
		if err != nil {
			return nil, err
		}

		addr := crypto.PubkeyToAddress(priv.PublicKey)

		addrs = append(addrs, addr.Hex())
	}

	return addrs, nil
}
