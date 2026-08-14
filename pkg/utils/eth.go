package utils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewKey() (*ecdsa.PrivateKey, common.Address, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, common.Address{}, err
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr, nil
}
