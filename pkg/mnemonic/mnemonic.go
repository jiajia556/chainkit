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
	masterKey  *bip32.Key
	accountKey *bip32.Key
	mu         sync.Mutex
}

// NewWords generates a new BIP-39 mnemonic (12 words by default).
func NewWords() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	words, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return words, nil
}

func NewWallet(mnemonic string, passphrase string) (*Wallet, error) {

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, passphrase)

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
		masterKey:  masterKey,
		accountKey: key,
	}, nil
}

func (w *Wallet) deriveKey(index uint32) (*bip32.Key, error) {

	w.mu.Lock()
	defer w.mu.Unlock()

	return w.accountKey.NewChildKey(index)
}

func (w *Wallet) PrivateKey(index uint32) (*ecdsa.PrivateKey, error) {

	child, err := w.deriveKey(index)
	if err != nil {
		return nil, err
	}

	return crypto.ToECDSA(child.Key)
}

func (w *Wallet) Address(index uint32) (common.Address, error) {

	priv, err := w.PrivateKey(index)
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

	priv, err := w.PrivateKey(index)
	if err != nil {
		return "", "", err
	}

	addr := crypto.PubkeyToAddress(priv.PublicKey)

	return addr.Hex(), hexutil.Encode(crypto.FromECDSA(priv)), nil
}

func (w *Wallet) BatchAddresses(start uint32, count uint32) ([]string, error) {

	addrs := make([]string, 0, count)

	for i := uint32(0); i < count; i++ {

		addr, err := w.AddressHex(start + i)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, addr)
	}

	return addrs, nil
}
