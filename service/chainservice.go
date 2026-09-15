package service

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

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
	rpcRequestLimiter *rpcRequestLimiter
	localRPCLimiter   rpcRequestLimiter
}

type rpcRequestLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

// All ChainService instances for one configured chain share a limiter. This is
// important for processes such as deposit, which use separate clients for the
// scanner, backfill tasks, and inbox processor.
var chainRPCLimiters sync.Map

func sharedRPCLimiter(chainDbID uint64) *rpcRequestLimiter {
	limiter, _ := chainRPCLimiters.LoadOrStore(chainDbID, &rpcRequestLimiter{})
	return limiter.(*rpcRequestLimiter)
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

func NewChainService(chainDbId uint64) (service *ChainService, err error) {
	defer wrapServiceErrors("NewChainService", &err)

	chain := chainkitchains.NewRecord()
	err = chain.Read(chainDbId)
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
		rpcRequestLimiter: sharedRPCLimiter(chain.Model.Id),
	}, nil
}

func NewChainServiceWithRPC(rpc string) (service *ChainService, err error) {
	defer wrapServiceErrors("NewChainServiceWithRPC", &err)

	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &ChainService{
		rpcClient:         client,
		chainId:           chainId,
		chainDbId:         0,
		safeConfirmations: 0,
		rpcRequestLimiter: &rpcRequestLimiter{},
	}, nil
}

func (s *ChainService) CloseClient() {
	if s == nil {
		return
	}
	if s.rpcClient != nil {
		s.rpcClient.Close()
		s.rpcClient = nil
	}
	if s.wsClient != nil {
		s.wsClient.Close()
		s.wsClient = nil
	}
}

func (s *ChainService) DialClient() (err error) {
	defer wrapServiceErrors("DialClient", &err)

	chain := chainkitchains.NewRecord()
	err = chain.Read(s.chainDbId)
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

// DialWSClient connects the websocket client configured for this chain.
// It is explicit rather than part of NewChainService so non-subscription
// services do not depend on websocket availability.
func (s *ChainService) DialWSClient(ctx context.Context) (err error) {
	defer wrapServiceErrors("DialWSClient", &err)

	if s == nil {
		return errors.New("chain service is nil")
	}

	chain := chainkitchains.NewRecord()
	if err := chain.Read(s.chainDbId); err != nil {
		return err
	}
	wsRPC := strings.TrimSpace(chain.Model.WsRpc)
	if wsRPC == "" {
		return errors.New("websocket RPC is not configured")
	}

	client, err := ethclient.DialContext(ctx, wsRPC)
	if err != nil {
		return err
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return err
	}
	if chainID.Uint64() != chain.Model.ChainId {
		client.Close()
		return errors.New("websocket RPC chain ID does not match chain configuration")
	}

	if s.wsClient != nil {
		s.wsClient.Close()
	}
	s.wsClient = client
	return nil
}

func (s *ChainService) SetFromByMnemonicAddress(fromAddrId uint64, password string) (err error) {
	defer wrapServiceErrors("SetFromByMnemonicAddress", &err)

	address := chainkitmnemonicaddresses.NewRecord()
	err = address.Read(fromAddrId)
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

func (s *ChainService) SetFromByDepositAddress(fromAddrId uint64, password string) (err error) {
	defer wrapServiceErrors("SetFromByDepositAddress", &err)

	address := chainkituserdepositaddress.NewRecord()
	err = address.Read(fromAddrId)
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

func (s *ChainService) GetWSClient() *ethclient.Client {
	return s.wsClient
}

// SetRPCRequestInterval sets the minimum time between throttled RPC requests.
// Configured ChainService instances for the same chain share this setting and
// request schedule. A non-positive interval disables throttling.
func (s *ChainService) SetRPCRequestInterval(interval time.Duration) {
	if s == nil {
		return
	}
	limiter := s.getRPCRequestLimiter()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if interval < 0 {
		interval = 0
	}
	limiter.interval = interval
	if interval == 0 {
		limiter.last = time.Time{}
	}
}

func (s *ChainService) getRPCRequestLimiter() *rpcRequestLimiter {
	if s.rpcRequestLimiter != nil {
		return s.rpcRequestLimiter
	}
	return &s.localRPCLimiter
}

func (s *ChainService) waitForRPCRequest(ctx context.Context) error {
	limiter := s.getRPCRequestLimiter()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.interval > 0 && !limiter.last.IsZero() {
		wait := time.Until(limiter.last.Add(limiter.interval))
		if wait > 0 {
			timer := time.NewTimer(wait)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	limiter.last = time.Now()
	return nil
}

func (s *ChainService) GetBindTransactOpts(opts ...Option) (transactOpts *bind.TransactOpts, err error) {
	defer wrapServiceErrors("GetBindTransactOpts", &err)

	if s.priKey == nil {
		return nil, errors.New("no private key")
	}
	transactOpts, err = bind.NewKeyedTransactorWithChainID(s.priKey, s.chainId)
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

func (s *ChainService) GetFromAddress() (address string, err error) {
	defer wrapServiceErrors("GetFromAddress", &err)

	if s.fromAddressType == "" {
		return "", errors.New("from address not set")
	}
	return s.fromAddress, nil
}

func (s *ChainService) GetFromId() (addressType types.ServiceAddressType, addressID uint64, err error) {
	defer wrapServiceErrors("GetFromId", &err)

	if s.fromAddressType == "" {
		return "", 0, errors.New("from address not set")
	}
	return s.fromAddressType, s.fromAddressId, nil
}

func (s *ChainService) GetChainDbId() uint64 {
	return s.chainDbId
}

func (s *ChainService) GetFromETHBalance() (balanceDecimal decimal.Decimal, err error) {
	defer wrapServiceErrors("GetFromETHBalance", &err)

	if s.fromAddressType == "" {
		return decimal.Zero, errors.New("from address not set")
	}
	balance, err := s.rpcClient.BalanceAt(context.Background(), common.HexToAddress(s.fromAddress), nil)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(balance, 0), nil
}
