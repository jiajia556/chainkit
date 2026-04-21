package service

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jiajia556/chainkit/models/chainkitcontracts"
	"github.com/jiajia556/chainkit/models/chainkittransferdetails"
	"github.com/jiajia556/chainkit/models/chainkittransferrecords"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/jiajia556/chainkit/pkg/contracts/multitransfer"
	"github.com/shopspring/decimal"
)

func (s *ChainService) TransferETH(to string, value decimal.Decimal, opts ...Option) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", errors.New("from address not set")
	}

	if !common.IsHexAddress(to) {
		return "", errors.New("invalid to address")
	}
	toAddr := common.HexToAddress(to)

	// defaults
	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	if opt.checkBalance {
		balance, err := s.BalanceAt(s.fromAddress)
		if err != nil {
			return "", err
		}
		if balance.LessThan(value) {
			return "", errors.New("insufficient balance")
		}
	}

	ctx := context.Background()

	fromAddr := crypto.PubkeyToAddress(s.priKey.PublicKey)
	chainID := s.chainId
	if chainID == nil {
		return "", errors.New("chain id not initialized")
	}

	// nonce
	var nonce uint64
	var err error
	if opt.nonce != nil {
		nonce = opt.nonce.Uint64()
	} else {
		nonce, err = s.client.PendingNonceAt(ctx, fromAddr)
		if err != nil {
			return "", err
		}
	}

	weiValue := value.BigInt()

	// gas price
	gasPrice := opt.gasPrice
	if gasPrice == nil {
		gasPrice, err = s.client.SuggestGasPrice(ctx)
		if err != nil {
			return "", err
		}
	}
	if gasPrice.Cmp(big.NewInt(100000000)) < 0 {
		gasPrice = big.NewInt(100000000)
	}

	// gas limit
	gasLimit := opt.gasLimit
	if gasLimit == 0 {
		callMsg := ethereum.CallMsg{From: fromAddr, To: &toAddr, Value: weiValue}
		estimated, estErr := s.client.EstimateGas(ctx, callMsg)
		if estErr != nil {
			// fallback to standard transfer gas
			gasLimit = 21000
		} else {
			gasLimit = estimated
		}
	}

	// Note: For EIP-1559 transactions, you would need to set GasTipCap and GasFeeCap instead of GasPrice.
	gasTipCap := opt.gasTipCap
	gasFeeCap := opt.gasFeeCap

	tx := &types.Transaction{}
	if gasTipCap != nil && gasFeeCap != nil {
		// EIP-1559 transaction
		tx = types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			To:        &toAddr,
			Value:     weiValue,
			Gas:       gasLimit,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Data:      nil,
		})
	} else {
		// Legacy tx for broad compatibility.
		tx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddr,
			Value:    weiValue,
			Gas:      gasLimit,
			GasPrice: gasPrice,
			Data:     nil,
		})
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.priKey)
	if err != nil {
		return "", err
	}

	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func (s *ChainService) TransferERC20(token, to string, amount decimal.Decimal, opts ...Option) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", errors.New("from address not set")
	}

	if !common.IsHexAddress(token) {
		return "", errors.New("invalid token address")
	}
	tokenAddr := common.HexToAddress(token)

	if !common.IsHexAddress(to) {
		return "", errors.New("invalid to address")
	}
	toAddr := common.HexToAddress(to)

	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	if opt.checkBalance {
		balance, err := s.BalanceOf(token, s.fromAddress)
		if err != nil {
			return "", err
		}
		if balance.LessThan(amount) {
			return "", errors.New("insufficient balance")
		}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.priKey, s.chainId)
	if err != nil {
		return "", err
	}
	if opt.nonce != nil {
		auth.Nonce = opt.nonce
	} else {
		nonce, err := s.client.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			return "", err
		}
		auth.Nonce = big.NewInt(int64(nonce))
	}
	if opt.value != nil {
		auth.Value = opt.value
	} else {
		auth.Value = big.NewInt(0)
	}
	if opt.gasPrice != nil {
		auth.GasPrice = opt.gasPrice
	} else {
		auth.GasPrice, err = s.client.SuggestGasPrice(context.Background())
		if err != nil {
			return "", err
		}
	}
	if auth.GasPrice.Cmp(big.NewInt(100000000)) < 0 {
		auth.GasPrice = big.NewInt(100000000)
	}

	if opt.gasLimit != 0 {
		auth.GasLimit = opt.gasLimit
	}

	if opt.gasTipCap != nil {
		auth.GasTipCap = opt.gasTipCap
	}
	if opt.gasFeeCap != nil {
		auth.GasFeeCap = opt.gasFeeCap
	}

	instance, err := erc20.NewErc20(tokenAddr, s.client)
	if err != nil {
		return "", err
	}

	tx, err := instance.Transfer(auth, toAddr, amount.BigInt())
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (s *ChainService) MultiTransfer(tokensStr, tosStr []string, valuesDec []decimal.Decimal, opts ...Option) (string, *big.Int, error) {
	if s == nil || s.client == nil {
		return "", nil, errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", nil, errors.New("from address not set")
	}
	tokens := make([]common.Address, len(tokensStr))
	tos := make([]common.Address, len(tosStr))
	values := make([]*big.Int, len(valuesDec))
	totalValue := big.NewInt(0)
	tokenAmountMap := make(map[string]decimal.Decimal)
	for i := range tokensStr {
		tokensStr[i] = strings.ToLower(tokensStr[i])
		if !common.IsHexAddress(tokensStr[i]) {
			return "", big.NewInt(0), errors.New("invalid token address: " + tokensStr[i])
		}
		tokens[i] = common.HexToAddress(tokensStr[i])

		if !common.IsHexAddress(tosStr[i]) {
			return "", big.NewInt(0), errors.New("invalid to address: " + tosStr[i])
		}
		tos[i] = common.HexToAddress(tosStr[i])

		values[i] = valuesDec[i].BigInt()
		if tokensStr[i] == "0x0000000000000000000000000000000000000000" {
			totalValue.Add(totalValue, values[i])
		}
		tokenAmountMap[tokensStr[i]] = tokenAmountMap[tokensStr[i]].Add(valuesDec[i])
	}

	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	if opt.checkBalance {
		for tokenAddrStr, amount := range tokenAmountMap {
			var balance decimal.Decimal
			var err error
			if tokenAddrStr == "0x0000000000000000000000000000000000000000" {
				balance, err = s.BalanceAt(s.fromAddress)
			} else {
				balance, err = s.BalanceOf(tokenAddrStr, s.fromAddress)
			}
			if err != nil {
				return "", big.NewInt(0), err
			}
			if balance.LessThan(amount) {
				return "", big.NewInt(0), errors.New("insufficient balance for token: " + tokenAddrStr)
			}
		}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.priKey, s.chainId)
	if err != nil {
		return "", big.NewInt(0), err
	}
	if opt.nonce != nil {
		auth.Nonce = opt.nonce
	} else {
		nonce, err := s.client.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			return "", big.NewInt(0), err
		}
		auth.Nonce = big.NewInt(int64(nonce))
	}
	if opt.value != nil {
		auth.Value = opt.value
	} else {
		auth.Value = big.NewInt(0)
	}
	if opt.gasPrice != nil {
		auth.GasPrice = opt.gasPrice
	} else {
		auth.GasPrice, err = s.client.SuggestGasPrice(context.Background())
		if err != nil {
			return "", big.NewInt(0), err
		}
	}
	if auth.GasPrice.Cmp(big.NewInt(100000000)) < 0 {
		auth.GasPrice = big.NewInt(100000000)
	}

	if opt.gasLimit != 0 {
		auth.GasLimit = opt.gasLimit
	}

	multiTransferContract := chainkitcontracts.NewRecord()
	err = multiTransferContract.ReadByNameAndChainDbId("MultiTransfer", s.chainDbId)
	if err != nil {
		return "", big.NewInt(0), err
	}

	instance, err := multitransfer.NewMultitransfer(common.HexToAddress(multiTransferContract.Model.Address), s.client)
	if err != nil {
		return "", big.NewInt(0), err
	}

	tx, err := instance.MultiTransferToken(auth, tokens, tos, values)
	if err != nil {
		return "", big.NewInt(0), err
	}

	return tx.Hash().Hex(), auth.Nonce, nil
}

func (s *ChainService) DBTransfer(count int, opts ...Option) error {
	if s == nil || s.client == nil || s.priKey == nil {
		return errors.New("transfer service not initialized")
	}

	pending := chainkittransferrecords.NewRecord()
	err := pending.ReadPending(s.chainDbId, s.fromAddressId)
	if err == nil && pending.Exists() {
		status, err := s.GetTxStatus(pending.Model.Hash)
		if err != nil {
			return err
		}
		if status == TxStatusPending {
			return nil
		}
		if status == TxStatusConfirmed {
			pending.SetSuccess()
			chainkittransferdetails.NewRecord().SetSuccessByTransferRecordId(pending.Model.Id)
		} else if status == TxStatusFailed {
			pending.SetFailed()
			chainkittransferdetails.NewRecord().SetFailedByTransferRecordId(pending.Model.Id)
		}
	}
	tokensStr := make([]string, 0)
	tosStr := make([]string, 0)
	valuesDec := make([]decimal.Decimal, 0)
	ids := make([]uint64, 0)

	list := chainkittransferdetails.NewList()
	err = list.FindByFromAddressIdAndStatus(s.fromAddressId, chainkittransferdetails.StatusWaiting, count)
	if err != nil {
		return err
	}

	if list.IsEmpty() {
		return nil
	}

	list.Foreach(func(key int, detail chainkittransferdetails.Record) (isBreak bool) {
		tokensStr = append(tokensStr, detail.Model.TokenAddress)
		tosStr = append(tosStr, detail.Model.To)
		valuesDec = append(valuesDec, detail.Model.Amount)
		ids = append(ids, detail.Model.Id)
		return false
	})

	txHash, nonce, err := s.MultiTransfer(tokensStr, tosStr, valuesDec, opts...)
	if err != nil {
		return err
	}

	record := chainkittransferrecords.NewRecord()
	record.Model.ChainDbId = s.chainDbId
	record.Model.AddressId = s.fromAddressId
	record.Model.Hash = txHash
	record.Model.Nonce = nonce.Uint64()
	record.Model.Status = chainkittransferrecords.StatusPending
	err = record.Create()
	if err != nil {
		return err
	}

	chainkittransferdetails.NewRecord().SetPending(ids, record.Model.Id)

	return err
}
