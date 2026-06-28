package service

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddress"
	"github.com/jiajia556/chainkit/pkg/types"
	"github.com/shopspring/decimal"
)

const (
	Mnemonic    types.ServiceAddressType = "mnemonic"
	UserDeposit types.ServiceAddressType = "user_deposit"
)

type ChainService struct {
	rpcClient         *ethclient.Client
	wsClient          *ethclient.Client
	chainDbId         uint64
	chainId           *big.Int
	fromAddressType   types.ServiceAddressType
	fromAddressId     uint64
	fromAddress       string
	priKey            *ecdsa.PrivateKey
	safeConfirmations uint64
}

type privateKeyAddressRecord interface {
	GetPriKey(password string) (*ecdsa.PrivateKey, error)
	GetAddress() string
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

func Nonce(nonce uint64) Option {
	return func(o *transactionOptions) {
		o.nonce = big.NewInt(int64(nonce))
	}
}

func Value(value decimal.Decimal) Option {
	return func(o *transactionOptions) {
		o.value = value.BigInt()
	}
}

func GasPrice(gasPrice decimal.Decimal) Option {
	return func(o *transactionOptions) {
		o.gasPrice = gasPrice.BigInt()
	}
}

func UserMinGasPrice() Option {
	return func(o *transactionOptions) {
		o.useMinGasPrice = true
	}
}

func GasLimit(gasLimit decimal.Decimal) Option {
	return func(o *transactionOptions) {
		o.gasLimit = gasLimit.BigInt().Uint64()
	}
}

func GasTipCap(tip decimal.Decimal) Option {
	return func(o *transactionOptions) {
		o.gasTipCap = tip.BigInt()
	}
}

func GasFeeCap(fee decimal.Decimal) Option {
	return func(o *transactionOptions) {
		o.gasFeeCap = fee.BigInt()
	}
}

func CheckBalance(checkBalance bool) Option {
	return func(o *transactionOptions) {
		o.checkBalance = checkBalance
	}
}

func NewChainService(chainDbId uint64) (*ChainService, error) {
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
		rpcClient:         client,
		chainId:           big.NewInt(int64(chain.Model.ChainId)),
		chainDbId:         chain.Model.Id,
		safeConfirmations: chain.Model.SafeConfirmations,
	}, nil
}

func (s *ChainService) CloseClient() {
	if s == nil || s.rpcClient == nil {
		return
	}
	s.rpcClient.Close()
	s.rpcClient = nil
}

func (s *ChainService) DialClient() error {
	chain := chainkitchains.NewRecord()
	err := chain.Read(s.chainDbId)
	if err != nil {
		return err
	}

	client, err := ethclient.Dial(chain.Model.Rpc)
	if err != nil {
		return err
	}
	s.rpcClient = client
	return nil
}

func (s *ChainService) SetFromByMnemonicAddress(fromAddrId uint64, password string) error {
	address := chainkitmnemonicaddresses.NewRecord()
	err := address.Read(fromAddrId)
	if err != nil {
		return err
	}
	priKey, fromAddress, err := privateKeyFromAddressRecord(address, password)
	if err != nil {
		return err
	}

	s.priKey = priKey
	s.fromAddressType = Mnemonic
	s.fromAddressId = address.Model.Id
	s.fromAddress = fromAddress

	return nil
}

func (s *ChainService) SetFromByDepositAddress(fromAddrId uint64, password string) error {
	address := chainkituserdepositaddress.NewRecord()
	err := address.Read(fromAddrId)
	if err != nil {
		return err
	}
	priKey, fromAddress, err := privateKeyFromAddressRecord(address, password)
	if err != nil {
		return err
	}

	s.priKey = priKey
	s.fromAddressType = UserDeposit
	s.fromAddressId = address.Model.Id
	s.fromAddress = fromAddress

	return nil
}

func privateKeyFromAddressRecord(record privateKeyAddressRecord, password string) (*ecdsa.PrivateKey, string, error) {
	priKey, err := record.GetPriKey(password)
	if err != nil {
		return nil, "", err
	}
	fromAddress := record.GetAddress()
	if !common.IsHexAddress(fromAddress) {
		return nil, "", errors.New("invalid stored from address")
	}
	derivedAddr := crypto.PubkeyToAddress(priKey.PublicKey)
	storedAddr := common.HexToAddress(fromAddress)
	if derivedAddr != storedAddr {
		return nil, "", errors.New("private key does not match stored address")
	}
	return priKey, fromAddress, nil
}

func (s *ChainService) GetClient() *ethclient.Client {
	return s.rpcClient
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
		nonce, err := s.rpcClient.PendingNonceAt(context.Background(), transactOpts.From)
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
	} else if opt.gasTipCap == nil {
		transactOpts.GasPrice, err = s.rpcClient.SuggestGasPrice(context.Background())
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

func (s *ChainService) GetFromAddress() (string, error) {
	if s.fromAddressType == "" {
		return "", errors.New("from address not set")
	}
	return s.fromAddress, nil
}

func (s *ChainService) GetFromId() (types.ServiceAddressType, uint64, error) {
	if s.fromAddressType == "" {
		return "", 0, errors.New("from address not set")
	}
	return s.fromAddressType, s.fromAddressId, nil
}

func (s *ChainService) GetChainDbId() uint64 {
	return s.chainDbId
}

func (s *ChainService) GetFromETHBalance() (decimal.Decimal, error) {
	if s.fromAddressType == "" {
		return decimal.Zero, errors.New("from address not set")
	}
	balance, err := s.rpcClient.BalanceAt(context.Background(), common.HexToAddress(s.fromAddress), nil)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(balance, 0), nil
}
