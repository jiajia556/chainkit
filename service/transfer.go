package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitcontracts"
	"github.com/jiajia556/chainkit/models/chainkittransferdetails"
	"github.com/jiajia556/chainkit/models/chainkittransferrecords"
	"github.com/jiajia556/chainkit/pkg/contracts/erc20"
	"github.com/jiajia556/chainkit/pkg/contracts/multitransfer"
	"github.com/shopspring/decimal"
)

func (s *ChainService) TransferETH(to string, value decimal.Decimal, opts ...Option) (hash string, fakeErr, err error) {
	if s == nil || s.client == nil {
		return "", nil, errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", nil, errors.New("from address not set")
	}

	if !common.IsHexAddress(to) {
		return "", nil, errors.New("invalid to address")
	}
	toAddr := common.HexToAddress(to)

	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	if opt.checkBalance {
		balance, err := s.BalanceAt(s.fromAddress)
		if err != nil {
			return "", nil, err
		}
		if balance.LessThan(value) {
			return "", nil, errors.New("insufficient balance")
		}
	}

	ctx := context.Background()
	chainID := s.chainId
	if chainID == nil {
		return "", nil, errors.New("chain id not initialized")
	}

	txOpts, err := s.GetBindTransactOpts(opts...)
	if err != nil {
		return "", nil, err
	}

	weiValue := value.BigInt()
	txOpts.Value = weiValue

	gasLimit := txOpts.GasLimit
	if gasLimit == 0 {
		callMsg := ethereum.CallMsg{From: txOpts.From, To: &toAddr, Value: weiValue}
		estimated, estErr := s.client.EstimateGas(ctx, callMsg)
		if estErr != nil {
			gasLimit = 21000
		} else {
			gasLimit = estimated
		}
	}

	nonce := uint64(0)
	if txOpts.Nonce != nil {
		nonce = txOpts.Nonce.Uint64()
	}

	tx := &types.Transaction{}
	if txOpts.GasTipCap != nil && txOpts.GasFeeCap != nil {
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &toAddr,
			Value:     weiValue,
			Gas:       gasLimit,
			GasTipCap: txOpts.GasTipCap,
			GasFeeCap: txOpts.GasFeeCap,
			Data:      nil,
		})
	} else {
		if txOpts.GasPrice == nil {
			return "", nil, errors.New("gas price not set")
		}
		tx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddr,
			Value:    weiValue,
			Gas:      gasLimit,
			GasPrice: txOpts.GasPrice,
			Data:     nil,
		})
	}

	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, s.priKey)
	if err != nil {
		return "", nil, err
	}

	if err := s.client.SendTransaction(ctx, signedTx); err != nil {
		hash = signedTx.Hash().Hex()
		fakeErr = err
		return hash, fakeErr, nil
	}

	return signedTx.Hash().Hex(), nil, nil
}

func (s *ChainService) TransferERC20(token, to string, amount decimal.Decimal, opts ...Option) (hash string, fakeErr, err error) {
	if s == nil || s.client == nil {
		return "", nil, errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", nil, errors.New("from address not set")
	}

	if !common.IsHexAddress(token) {
		return "", nil, errors.New("invalid token address")
	}
	tokenAddr := common.HexToAddress(token)

	if !common.IsHexAddress(to) {
		return "", nil, errors.New("invalid to address")
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
			return "", nil, err
		}
		if balance.LessThan(amount) {
			return "", nil, errors.New("insufficient balance")
		}
	}

	txOpts, err := s.GetBindTransactOpts(opts...)
	if err != nil {
		return "", nil, err
	}

	var lastSignedTx *types.Transaction
	originalSigner := txOpts.Signer
	txOpts.Signer = func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		signedTx, signErr := originalSigner(addr, tx)
		if signErr != nil {
			return nil, signErr
		}
		lastSignedTx = signedTx
		return signedTx, nil
	}

	instance, err := erc20.NewErc20(tokenAddr, s.client)
	if err != nil {
		return "", nil, err
	}

	tx, err := instance.Transfer(txOpts, toAddr, amount.BigInt())
	if err != nil {
		if lastSignedTx == nil {
			return "", nil, err
		}
		hash = lastSignedTx.Hash().Hex()
		fakeErr = err
		return hash, fakeErr, nil
	}

	return tx.Hash().Hex(), nil, nil
}

func (s *ChainService) MultiTransfer(tokensStr, tosStr []string, valuesDec []decimal.Decimal, opts ...Option) (hash string, nonce *big.Int, fakeErr, err error) {
	if s == nil || s.client == nil {
		return "", nil, nil, errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", nil, nil, errors.New("from address not set")
	}
	if len(tokensStr) == 0 || len(tosStr) == 0 || len(valuesDec) == 0 {
		return "", nil, nil, errors.New("transfer lists are empty")
	}
	if len(tokensStr) != len(tosStr) || len(tokensStr) != len(valuesDec) {
		return "", nil, nil, errors.New("transfer lists length mismatch")
	}

	tokens := make([]common.Address, len(tokensStr))
	tos := make([]common.Address, len(tosStr))
	values := make([]*big.Int, len(valuesDec))
	totalValue := big.NewInt(0)
	tokenAmountMap := make(map[string]decimal.Decimal)
	for i := range tokensStr {
		tokensStr[i] = strings.ToLower(tokensStr[i])
		if !common.IsHexAddress(tokensStr[i]) {
			return "", big.NewInt(0), nil, errors.New("invalid token address: " + tokensStr[i])
		}
		tokens[i] = common.HexToAddress(tokensStr[i])

		if !common.IsHexAddress(tosStr[i]) {
			return "", big.NewInt(0), nil, errors.New("invalid to address: " + tosStr[i])
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
				return "", big.NewInt(0), nil, err
			}
			if balance.LessThan(amount) {
				return "", big.NewInt(0), nil, errors.New("insufficient balance for token: " + tokenAddrStr)
			}
		}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.priKey, s.chainId)
	if err != nil {
		return "", big.NewInt(0), nil, err
	}

	var lastSignedTx *types.Transaction
	auth.Signer = func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		signedTx, err := types.SignTx(
			tx,
			types.LatestSignerForChainID(s.chainId),
			s.priKey,
		)
		if err != nil {
			return nil, err
		}

		lastSignedTx = signedTx

		return signedTx, nil
	}

	if opt.nonce != nil {
		auth.Nonce = opt.nonce
	} else {
		nonce, err := s.client.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			return "", big.NewInt(0), nil, err
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
			return "", big.NewInt(0), nil, err
		}
	}
	if opt.useMinGasPrice {
		if auth.GasPrice.Cmp(big.NewInt(100000000)) < 0 {
			auth.GasPrice = big.NewInt(100000000)
		}
	}

	if opt.gasLimit != 0 {
		auth.GasLimit = opt.gasLimit
	}

	multiTransferContract := chainkitcontracts.NewRecord()
	err = multiTransferContract.ReadByNameAndChainDbId("MultiTransfer", s.chainDbId)
	if err != nil {
		return "", big.NewInt(0), nil, err
	}

	instance, err := multitransfer.NewMultitransfer(common.HexToAddress(multiTransferContract.Model.Address), s.client)
	if err != nil {
		return "", big.NewInt(0), nil, err
	}

	nonce = auth.Nonce

	tx, err := instance.MultiTransferToken(auth, tokens, tos, values)
	if err != nil {
		if lastSignedTx == nil {
			return "", big.NewInt(0), nil, err
		}
		hash = lastSignedTx.Hash().Hex()
		fakeErr = err
		err = nil
	} else {
		hash = tx.Hash().Hex()
	}

	return
}

func (s *ChainService) DBTransfer(count int, opts ...Option) error {
	if s == nil || s.client == nil || s.priKey == nil {
		return errors.New("transfer service not initialized")
	}

	pending := chainkittransferrecords.NewRecord()
	err := pending.ReadPending(s.chainDbId, s.fromAddressType, s.fromAddressId)
	if err == nil && pending.Exists() {
		status, err := s.GetTxStatus(pending.Model.Hash)
		if err != nil {
			return err
		}
		if status == TxStatusNotFound {
			if pending.SinceCreated() > time.Minute*15 {
				occupied, err := s.IsNonceOccupied(s.fromAddress, pending.Model.Nonce)
				if err != nil {
					return err
				}
				if occupied {
					// 特殊情况，人工处理
					pending.SetUnknown()
					chainkittransferdetails.NewRecord().SetUnknownByTransferRecordId(pending.Model.Id)
				} else {
					pending.SetFailed()
					chainkittransferdetails.NewRecord().SetWaitingByTransferRecordId(pending.Model.Id)
					opts = append(opts, Nonce(pending.Model.Nonce))
				}
			}
			return nil
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
	err = list.FindByFromAddressIdAndStatus(s.fromAddressId, s.fromAddressType, chainkittransferdetails.StatusWaiting, count)
	if err != nil {
		return err
	}

	if list.IsEmpty() {
		return nil
	}

	list.Foreach(func(key int, detail *chainkittransferdetails.Record) bool {
		tokensStr = append(tokensStr, detail.Model.TokenAddress)
		tosStr = append(tosStr, detail.Model.To)
		valuesDec = append(valuesDec, detail.Model.Amount)
		ids = append(ids, detail.Model.Id)
		return true
	})

	txHash, nonce, _, err := s.MultiTransfer(tokensStr, tosStr, valuesDec, opts...)
	if err != nil {
		return err
	}

	record := chainkittransferrecords.NewRecord()
	record.Model.ChainDbId = s.chainDbId
	record.Model.FromAddressType = string(s.fromAddressType)
	record.Model.FromAddressId = s.fromAddressId
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
