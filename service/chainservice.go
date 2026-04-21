package service

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	nonce          *big.Int
	value          *big.Int
	gasPrice       *big.Int
	gasLimit       uint64
	gasTipCap      *big.Int
	gasFeeCap      *big.Int
	checkBalance   bool
	useMinGasPrice bool
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

func UserMinGasPrice() Option {
	return func(o *transactionOptions) {
		o.useMinGasPrice = true
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

func NewTransferService(chainDbId uint64) (*ChainService, error) {
	chain := chainkitchains.NewRecord()
	err := chain.Read(chainDbId)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(chain.Model.Rpc)
	if err != nil {
		return nil, err
	}

	return &ChainService{
		client:    client,
		chainId:   big.NewInt(chain.Model.ChainId),
		chainDbId: chain.Model.Id,
	}, nil
}

func (s *ChainService) SetFrom(fromAddrId uint64, password string) error {
	address := chainkitmnemonicaddresses.NewRecord()
	err := address.Read(fromAddrId)
	if err != nil {
		return err
	}
	priKey, err := address.GetPriKey(password)
	if err != nil {
		return err
	}

	s.priKey = priKey
	s.fromAddressId = address.Model.Id
	s.fromAddress = address.Model.Address

	return nil
}

func (s *ChainService) GetClient() *ethclient.Client {
	return s.client
}

func (s *ChainService) GetBindTransactOpts(opts ...Option) (*bind.TransactOpts, error) {
	if s.priKey == nil {
		return nil, errors.New("no private key")
	}
	transactOpts, err := bind.NewKeyedTransactorWithChainID(s.priKey, s.chainId)
	if err != nil {
		return nil, err
	}

	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	if opt.nonce != nil {
		transactOpts.Nonce = opt.nonce
	} else {
		nonce, err := s.client.PendingNonceAt(context.Background(), transactOpts.From)
		if err != nil {
			return nil, err
		}
		transactOpts.Nonce = big.NewInt(int64(nonce))
	}
	if opt.value != nil {
		transactOpts.Value = opt.value
	} else {
		transactOpts.Value = big.NewInt(0)
	}
	if opt.gasPrice != nil {
		transactOpts.GasPrice = opt.gasPrice
	} else {
		transactOpts.GasPrice, err = s.client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, err
		}
	}
	if opt.useMinGasPrice {
		if transactOpts.GasPrice.Cmp(big.NewInt(100000000)) < 0 {
			transactOpts.GasPrice = big.NewInt(100000000)
		}
	}

	if opt.gasLimit != 0 {
		transactOpts.GasLimit = opt.gasLimit
	}

	if opt.gasTipCap != nil {
		transactOpts.GasTipCap = opt.gasTipCap
	}
	if opt.gasFeeCap != nil {
		transactOpts.GasFeeCap = opt.gasFeeCap
	}

	return transactOpts, nil
}
