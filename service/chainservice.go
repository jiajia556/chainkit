package service

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
)

type ChainService struct {
	client        *ethclient.Client
	chainDbId     uint64
	chainId       *big.Int
	fromAddressId uint64
	fromAddress   string
	priKey        *ecdsa.PrivateKey
}

type transactionOptions struct {
	nonce        *big.Int
	value        *big.Int
	gasPrice     *big.Int
	gasLimit     uint64
	gasTipCap    *big.Int
	gasFeeCap    *big.Int
	checkBalance bool
}

type Option func(*transactionOptions)

func Nonce(nonce *big.Int) Option {
	return func(o *transactionOptions) {
		o.nonce = nonce
	}
}

func Value(value *big.Int) Option {
	return func(o *transactionOptions) {
		o.value = value
	}
}

func GasPrice(gasPrice *big.Int) Option {
	return func(o *transactionOptions) {
		o.gasPrice = gasPrice
	}
}

func GasLimit(gasLimit uint64) Option {
	return func(o *transactionOptions) {
		o.gasLimit = gasLimit
	}
}

func GasTipCap(tip *big.Int) Option {
	return func(o *transactionOptions) {
		o.gasTipCap = tip
	}
}

func GasFeeCap(fee *big.Int) Option {
	return func(o *transactionOptions) {
		o.gasFeeCap = fee
	}
}

func CheckBalance(checkBalance bool) Option {
	return func(o *transactionOptions) {
		o.checkBalance = checkBalance
	}
}

func NewTransferService(chainDbId, fromAddrId uint64, password string) (*ChainService, error) {
	chain := chainkitchains.NewRecord()
	err := chain.Read(chainDbId)
	if err != nil {
		return nil, err
	}

	address := chainkitmnemonicaddresses.NewRecord()
	var priKey *ecdsa.PrivateKey
	if fromAddrId > 0 {
		err = address.Read(fromAddrId)
		if err != nil {
			return nil, err
		}
		priKey, err = address.GetPriKey(password)
		if err != nil {
			return nil, err
		}
	}

	client, err := ethclient.Dial(chain.Model.Rpc)
	if err != nil {
		return nil, err
	}

	return &ChainService{
		client:        client,
		priKey:        priKey,
		chainId:       big.NewInt(chain.Model.ChainId),
		chainDbId:     chain.Model.Id,
		fromAddressId: address.Model.Id,
		fromAddress:   address.Model.Address,
	}, nil
}
