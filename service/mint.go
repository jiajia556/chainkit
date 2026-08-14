package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jiajia556/chainkit/models/chainkitmintdetails"
	"github.com/jiajia556/chainkit/models/chainkitmintrecords"
	"github.com/jiajia556/chainkit/models/chainkittokens"
	"github.com/jiajia556/chainkit/pkg/contracts/mintburnerc20"
	"github.com/shopspring/decimal"
)

func (s *ChainService) MintERC20(token, to string, amount decimal.Decimal, opts ...Option) (hash string, nonce uint64, fakeErr, err error) {
	if s == nil || s.rpcClient == nil {
		return "", 0, nil, errors.New("MintERC20: transfer service not initialized")
	}
	if s.priKey == nil {
		return "", 0, nil, errors.New("MintERC20: from address not set")
	}

	if !common.IsHexAddress(token) {
		return "", 0, nil, errors.New("MintERC20: invalid token address")
	}
	tokenAddr := common.HexToAddress(token)

	if !common.IsHexAddress(to) {
		return "", 0, nil, errors.New("MintERC20: invalid to address")
	}
	toAddr := common.HexToAddress(to)

	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	txOpts, err := s.GetBindTransactOpts(opts...)
	if err != nil {
		return "", 0, nil, fmt.Errorf("MintERC20: get transact opts: %w", err)
	}

	nonce = 0
	if txOpts.Nonce != nil {
		nonce = txOpts.Nonce.Uint64()
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

	instance, err := mintburnerc20.NewMintburnerc20(tokenAddr, s.rpcClient)
	if err != nil {
		return "", 0, nil, fmt.Errorf("MintERC20: new contract instance: %w", err)
	}

	tx, err := instance.Mint(txOpts, toAddr, amount.BigInt())
	if err != nil {
		if lastSignedTx == nil {
			return "", 0, nil, fmt.Errorf("MintERC20: mint send failed: %w", err)
		}
		hash = lastSignedTx.Hash().Hex()
		fakeErr = err
		return hash, nonce, fakeErr, nil
	}

	return tx.Hash().Hex(), nonce, nil, nil
}

func (s *ChainService) BatchMintERC20(token string, tosStr []string, valuesDec []decimal.Decimal, opts ...Option) (hash string, nonce uint64, fakeErr, err error) {
	if s == nil || s.rpcClient == nil {
		return "", 0, nil, errors.New("transfer service not initialized")
	}
	if s.priKey == nil {
		return "", 0, nil, errors.New("from address not set")
	}
	if !common.IsHexAddress(token) {
		return "", 0, nil, errors.New("invalid token address")
	}
	if len(tosStr) == 0 || len(valuesDec) == 0 {
		return "", 0, nil, errors.New("transfer lists are empty")
	}
	if len(valuesDec) != len(tosStr) {
		return "", 0, nil, errors.New("transfer lists length mismatch")
	}

	tokenAddr := common.HexToAddress(token)
	tos := make([]common.Address, len(tosStr))
	values := make([]*big.Int, len(valuesDec))
	for i := range tosStr {
		if !common.IsHexAddress(tosStr[i]) {
			return "", 0, nil, errors.New("invalid to address: " + tosStr[i])
		}
		tos[i] = common.HexToAddress(tosStr[i])

		values[i] = valuesDec[i].BigInt()
	}

	opt := &transactionOptions{}
	for _, apply := range opts {
		if apply != nil {
			apply(opt)
		}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.priKey, s.chainId)
	if err != nil {
		return "", 0, nil, err
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
		nonceValue, err := s.rpcClient.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			return "", 0, nil, err
		}
		auth.Nonce = big.NewInt(int64(nonceValue))
	}
	if opt.value != nil {
		auth.Value = opt.value
	} else {
		auth.Value = big.NewInt(0)
	}
	if opt.gasPrice != nil {
		auth.GasPrice = opt.gasPrice
	} else {
		auth.GasPrice, err = s.rpcClient.SuggestGasPrice(context.Background())
		if err != nil {
			return "", 0, nil, err
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

	instance, err := mintburnerc20.NewMintburnerc20(tokenAddr, s.rpcClient)
	if err != nil {
		return "", 0, nil, err
	}

	nonce = auth.Nonce.Uint64()

	tx, err := instance.BatchMint(auth, tos, values)
	if err != nil {
		if lastSignedTx == nil {
			return "", 0, nil, err
		}
		hash = lastSignedTx.Hash().Hex()
		fakeErr = err
		err = nil
	} else {
		hash = tx.Hash().Hex()
	}

	return
}

func (s *ChainService) DBMint(count int, tokenId uint64, opts ...Option) error {
	if s == nil || s.rpcClient == nil || s.priKey == nil {
		return errors.New("DBMint: service not initialized")
	}

	retryPending := false
	pending := chainkitmintrecords.NewRecord()
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
					chainkitmintdetails.NewRecord().SetUnknownByTransferRecordId(pending.Model.Id)
				} else {
					pending.SetFailed()
					chainkitmintdetails.NewRecord().SetWaitingByTransferRecordId(pending.Model.Id)
					opts = append(opts, Nonce(pending.Model.Nonce))
					retryPending = true
				}
			}
			if !retryPending {
				return nil
			}
		}
		if status == TxStatusPending {
			return nil
		}
		if status == TxStatusConfirmed {
			pending.SetSuccess()
			chainkitmintdetails.NewRecord().SetSuccessByTransferRecordId(pending.Model.Id)
		} else if status == TxStatusFailed {
			pending.SetFailed()
			chainkitmintdetails.NewRecord().SetFailedByTransferRecordId(pending.Model.Id)
		}
	}
	token := chainkittokens.NewRecord()
	_ = token.Read(tokenId)
	if !token.Exists() {
		return errors.New("DBMint: token not found")
	}
	tosStr := make([]string, 0)
	valuesDec := make([]decimal.Decimal, 0)
	ids := make([]uint64, 0)

	list := chainkitmintdetails.NewList()
	err = list.FindByFromAddressIdAndStatus(s.fromAddressId, tokenId, s.fromAddressType, chainkitmintdetails.StatusWaiting, count)
	if err != nil {
		return fmt.Errorf("DBMint: find details: %w", err)
	}

	if list.IsEmpty() {
		return nil
	}

	list.Foreach(func(key int, detail *chainkitmintdetails.Record) bool {
		tosStr = append(tosStr, detail.Model.To)
		valuesDec = append(valuesDec, detail.Model.Amount)
		ids = append(ids, detail.Model.Id)
		return true
	})

	txHash, nonce, fakeErr, err := s.BatchMintERC20(token.Model.ContractAddress, tosStr, valuesDec, opts...)
	if err != nil {
		return fmt.Errorf("DBMint: batch mint: %w", err)
	}
	if fakeErr != nil {
		return fmt.Errorf("DBMint: batch mint uncertain: %w", fakeErr)
	}

	record := chainkitmintrecords.NewRecord()
	record.Model.ChainDbId = s.chainDbId
	record.Model.FromAddressType = string(s.fromAddressType)
	record.Model.FromAddressId = s.fromAddressId
	record.Model.Hash = txHash
	record.Model.Nonce = nonce
	record.Model.Status = chainkitmintrecords.StatusPending
	err = record.Create()
	if err != nil {
		return fmt.Errorf("DBMint: create record: %w", err)
	}

	chainkitmintdetails.NewRecord().SetPending(ids, record.Model.Id)

	return err
}
